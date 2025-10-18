package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/prasenjit-net/openid-golang/internal/config"
	"github.com/prasenjit-net/openid-golang/internal/handlers"
	"github.com/prasenjit-net/openid-golang/internal/setup"
	"github.com/prasenjit-net/openid-golang/internal/storage"
	"github.com/prasenjit-net/openid-golang/internal/ui"
)

// Version is set by the build process
var Version = "dev"

func main() {
	// Parse command line flags
	versionFlag := flag.Bool("version", false, "Print version and exit")
	setupFlag := flag.Bool("setup", false, "Run interactive setup wizard")
	jsonStoreFlag := flag.Bool("json-store", false, "Use JSON file storage instead of MongoDB")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("OpenID Connect Server v%s\n", Version)
		os.Exit(0)
	}

	if *setupFlag {
		if err := setup.Run(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
		os.Exit(0)
	}

	log.Printf("Starting OpenID Connect Server v%s", Version)

	// Check if config.toml exists
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		log.Println("❌ Configuration file not found!")
		log.Println("")
		log.Println("Please run the setup wizard first:")
		log.Println("  ./openid-server --setup")
		log.Println("")
		log.Println("This will:")
		log.Println("  - Generate RSA keys for JWT signing")
		log.Println("  - Create config.toml with your preferences")
		log.Println("  - Choose storage backend (MongoDB or JSON)")
		log.Println("  - Create admin user and OAuth clients")
		log.Println("")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override storage type if --json-store flag is provided
	if *jsonStoreFlag {
		cfg.Storage.Type = "json"
		if cfg.Storage.JSONFilePath == "" {
			cfg.Storage.JSONFilePath = "data.json"
		}
		log.Printf("Using JSON file storage: %s", cfg.Storage.JSONFilePath)
	} else if cfg.Storage.Type == "mongodb" {
		log.Printf("Using MongoDB storage: %s", cfg.Storage.MongoURI)
	} else {
		log.Printf("Using JSON file storage: %s", cfg.Storage.JSONFilePath)
	}

	// Initialize storage
	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize handlers
	h := handlers.NewHandlers(cfg, store)
	adminHandler := handlers.NewAdminHandler(store, cfg)

	// Setup Echo server
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Setup routes
	setupRoutes(e, h, adminHandler)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	go func() {
		log.Printf("Starting OpenID Connect server on %s", addr)
		log.Printf("Issuer: %s", cfg.Issuer)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
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

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func setupRoutes(e *echo.Echo, h *handlers.Handlers, adminHandler *handlers.AdminHandler) {
	// OpenID Connect Discovery
	e.GET("/.well-known/openid-configuration", echo.WrapHandler(http.HandlerFunc(h.Discovery)))
	e.GET("/.well-known/jwks.json", echo.WrapHandler(http.HandlerFunc(h.JWKS)))

	// OpenID Connect endpoints
	e.GET("/authorize", echo.WrapHandler(http.HandlerFunc(h.Authorize)))
	e.POST("/token", echo.WrapHandler(http.HandlerFunc(h.Token)))
	e.GET("/userinfo", echo.WrapHandler(http.HandlerFunc(h.UserInfo)))
	e.POST("/userinfo", echo.WrapHandler(http.HandlerFunc(h.UserInfo)))

	// Authentication endpoints
	e.GET("/login", echo.WrapHandler(http.HandlerFunc(h.Login)))
	e.POST("/login", echo.WrapHandler(http.HandlerFunc(h.Login)))
	e.GET("/consent", echo.WrapHandler(http.HandlerFunc(h.Consent)))
	e.POST("/consent", echo.WrapHandler(http.HandlerFunc(h.Consent)))

	// Admin API endpoints
	e.GET("/api/admin/stats", echo.WrapHandler(http.HandlerFunc(adminHandler.GetStats)))
	e.GET("/api/admin/users", echo.WrapHandler(http.HandlerFunc(adminHandler.ListUsers)))
	e.POST("/api/admin/users", echo.WrapHandler(http.HandlerFunc(adminHandler.CreateUser)))
	e.PUT("/api/admin/users/:id", echo.WrapHandler(http.HandlerFunc(adminHandler.UpdateUser)))
	e.DELETE("/api/admin/users/:id", echo.WrapHandler(http.HandlerFunc(adminHandler.DeleteUser)))
	e.GET("/api/admin/clients", echo.WrapHandler(http.HandlerFunc(adminHandler.ListClients)))
	e.POST("/api/admin/clients", echo.WrapHandler(http.HandlerFunc(adminHandler.CreateClient)))
	e.PUT("/api/admin/clients/:id", echo.WrapHandler(http.HandlerFunc(adminHandler.UpdateClient)))
	e.DELETE("/api/admin/clients/:id", echo.WrapHandler(http.HandlerFunc(adminHandler.DeleteClient)))
	e.GET("/api/admin/settings", echo.WrapHandler(http.HandlerFunc(adminHandler.GetSettings)))
	e.PUT("/api/admin/settings", echo.WrapHandler(http.HandlerFunc(adminHandler.UpdateSettings)))
	e.GET("/api/admin/keys", echo.WrapHandler(http.HandlerFunc(adminHandler.GetKeys)))
	e.POST("/api/admin/keys/rotate", echo.WrapHandler(http.HandlerFunc(adminHandler.RotateKeys)))
	e.GET("/api/admin/setup/status", echo.WrapHandler(http.HandlerFunc(adminHandler.GetSetupStatus)))
	e.POST("/api/admin/setup", echo.WrapHandler(http.HandlerFunc(adminHandler.CompleteSetup)))
	e.POST("/api/admin/login", echo.WrapHandler(http.HandlerFunc(adminHandler.Login)))

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Serve embedded admin UI with SPA routing support
	// Convert http.FileSystem to fs.FS
	adminFS := ui.GetAdminUI()
	assetHandler := http.FileServer(adminFS)

	// Serve static assets directly
	e.GET("/assets/*", echo.WrapHandler(assetHandler))
	e.GET("/vite.svg", echo.WrapHandler(assetHandler))

	// For all other routes, serve index.html (SPA fallback)
	e.GET("/*", func(c echo.Context) error {
		// Skip API routes and already handled routes
		path := c.Request().URL.Path
		if len(path) > 4 && path[:4] == "/api" {
			return echo.ErrNotFound
		}
		if len(path) > 12 && path[:12] == "/.well-known" {
			return echo.ErrNotFound
		}
		if path == "/authorize" || path == "/login" || path == "/consent" ||
			path == "/token" || path == "/userinfo" || path == "/health" {
			return echo.ErrNotFound
		}

		// Try to open the file
		file, err := adminFS.Open(path)
		if err == nil {
			file.Close()
			// File exists, serve it
			assetHandler.ServeHTTP(c.Response(), c.Request())
			return nil
		}

		// File doesn't exist, serve index.html for SPA routing
		c.Request().URL.Path = "/"
		assetHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}
