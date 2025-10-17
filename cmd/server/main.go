package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prasenjit/openid-golang/internal/config"
	"github.com/prasenjit/openid-golang/internal/handlers"
	"github.com/prasenjit/openid-golang/internal/middleware"
	"github.com/prasenjit/openid-golang/internal/storage"
	"github.com/prasenjit/openid-golang/internal/ui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize handlers
	h := handlers.NewHandlers(cfg, store)
	adminHandler := handlers.NewAdminHandler(store)

	// Setup router
	router := setupRouter(h, adminHandler)

	// Apply middleware
	handler := middleware.Logging(
		middleware.CORS(
			middleware.Recovery(router),
		),
	)

	// Create server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting OpenID Connect server on %s", addr)
		log.Printf("Issuer: %s", cfg.Issuer)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func setupRouter(h *handlers.Handlers, adminHandler *handlers.AdminHandler) *mux.Router {
	router := mux.NewRouter()

	// OpenID Connect Discovery
	router.HandleFunc("/.well-known/openid-configuration", h.Discovery).Methods("GET")
	router.HandleFunc("/.well-known/jwks.json", h.JWKS).Methods("GET")

	// OpenID Connect endpoints
	router.HandleFunc("/authorize", h.Authorize).Methods("GET")
	router.HandleFunc("/token", h.Token).Methods("POST")
	router.HandleFunc("/userinfo", h.UserInfo).Methods("GET", "POST")

	// Authentication endpoints
	router.HandleFunc("/login", h.Login).Methods("GET", "POST")
	router.HandleFunc("/consent", h.Consent).Methods("GET", "POST")

	// Admin API endpoints
	router.HandleFunc("/api/admin/stats", adminHandler.GetStats).Methods("GET")
	router.HandleFunc("/api/admin/users", adminHandler.ListUsers).Methods("GET")
	router.HandleFunc("/api/admin/users", adminHandler.CreateUser).Methods("POST")
	router.HandleFunc("/api/admin/users/{id}", adminHandler.DeleteUser).Methods("DELETE")
	router.HandleFunc("/api/admin/clients", adminHandler.ListClients).Methods("GET")
	router.HandleFunc("/api/admin/clients", adminHandler.CreateClient).Methods("POST")
	router.HandleFunc("/api/admin/clients/{id}", adminHandler.DeleteClient).Methods("DELETE")
	router.HandleFunc("/api/admin/settings", adminHandler.GetSettings).Methods("GET")
	router.HandleFunc("/api/admin/settings", adminHandler.UpdateSettings).Methods("PUT")
	router.HandleFunc("/api/admin/keys", adminHandler.GetKeys).Methods("GET")
	router.HandleFunc("/api/admin/keys/rotate", adminHandler.RotateKeys).Methods("POST")
	router.HandleFunc("/api/admin/setup/status", adminHandler.GetSetupStatus).Methods("GET")
	router.HandleFunc("/api/admin/setup", adminHandler.CompleteSetup).Methods("POST")
	router.HandleFunc("/api/admin/login", adminHandler.Login).Methods("POST")

	// TODO: Serve embedded admin UI static files
	router.PathPrefix("/").Handler(http.FileServer(ui.GetAdminUI()))

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	return router
}
