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

	// Setup router
	router := setupRouter(h)

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

func setupRouter(h *handlers.Handlers) *mux.Router {
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

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	return router
}
