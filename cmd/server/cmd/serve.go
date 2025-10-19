package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prasenjit-net/openid-golang/internal/config"
	"github.com/prasenjit-net/openid-golang/internal/crypto"
	"github.com/prasenjit-net/openid-golang/internal/handlers"
	"github.com/prasenjit-net/openid-golang/internal/models"
	"github.com/prasenjit-net/openid-golang/internal/storage"
	"github.com/prasenjit-net/openid-golang/internal/ui"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the OpenID Connect server",
	Long:  `Start the OpenID Connect server and begin accepting requests.`,
	Run:   runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) {
	// Load configuration from Viper
	cfg, err := loadConfigFromViper()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize storage
	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Ensure admin-ui client exists
	adminClient, err := store.GetClientByID("admin-ui")
	if err != nil || adminClient == nil {
		adminClient = models.NewAdminUIClient(cfg.Issuer)
		if err := store.CreateClient(adminClient); err != nil {
			log.Printf("Warning: Failed to create admin-ui client: %v", err)
		} else {
			log.Println("Created admin-ui client")
		}
	}

	// Initialize JWT manager
	jwtManager, err := crypto.NewJWTManager(
		cfg.JWT.PrivateKeyPath,
		cfg.JWT.PublicKeyPath,
		cfg.Issuer,
		cfg.JWT.ExpiryMinutes,
	)
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize handlers
	h := handlers.NewHandlers(store, jwtManager, cfg)

	// Register routes
	registerRoutes(e, h, cfg)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting OpenID Connect Server v%s", getVersion())
	log.Printf("Using %s storage", cfg.Storage.Type)
	log.Printf("Starting OpenID Connect server on %s", addr)
	log.Printf("Issuer: %s", cfg.Issuer)

	// Start server with graceful shutdown
	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("Server stopped")
}

func loadConfigFromViper() (*config.Config, error) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: viper.GetString("server.host"),
			Port: viper.GetInt("server.port"),
		},
		Storage: config.StorageConfig{
			Type:         viper.GetString("storage.type"),
			JSONFilePath: viper.GetString("storage.json_file_path"),
			MongoURI:     viper.GetString("storage.mongo_uri"),
		},
		JWT: config.JWTConfig{
			PrivateKeyPath: viper.GetString("jwt.private_key_path"),
			PublicKeyPath:  viper.GetString("jwt.public_key_path"),
			ExpiryMinutes:  viper.GetInt("jwt.expiry_minutes"),
		},
		Issuer: viper.GetString("issuer"),
	}

	// Set defaults if not provided
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Storage.Type == "" {
		cfg.Storage.Type = config.StorageTypeJSON
	}
	if cfg.Storage.JSONFilePath == "" {
		cfg.Storage.JSONFilePath = config.DefaultJSONFile
	}
	if cfg.JWT.ExpiryMinutes == 0 {
		cfg.JWT.ExpiryMinutes = 60
	}
	if cfg.JWT.PrivateKeyPath == "" {
		cfg.JWT.PrivateKeyPath = "config/keys/private.key"
	}
	if cfg.JWT.PublicKeyPath == "" {
		cfg.JWT.PublicKeyPath = "config/keys/public.key"
	}
	if cfg.Issuer == "" {
		cfg.Issuer = fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		if cfg.Server.Host == "0.0.0.0" {
			cfg.Issuer = fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
		}
	}

	return cfg, nil
}

func registerRoutes(e *echo.Echo, h *handlers.Handlers, cfg *config.Config) {
	// OpenID Connect Discovery
	e.GET("/.well-known/openid-configuration", h.Discovery)
	e.GET("/.well-known/jwks.json", h.JWKS)

	// OAuth/OpenID endpoints
	e.GET("/authorize", h.Authorize)
	e.POST("/token", h.Token)
	e.GET("/userinfo", h.UserInfo)
	e.POST("/userinfo", h.UserInfo)

	// Login and consent pages
	e.GET("/login", h.Login)
	e.POST("/login", h.Login)
	e.GET("/consent", h.Consent)
	e.POST("/consent", h.Consent)

	// Admin API
	adminAPIHandler := handlers.NewAdminHandler(h.GetStorage(), cfg)
	api := e.Group("/api/admin")

	// Setup endpoints (no auth required)
	api.GET("/setup/status", adminAPIHandler.GetSetupStatus)

	// Stats and management (should be authenticated in production)
	api.GET("/stats", adminAPIHandler.GetStats)
	api.GET("/users", adminAPIHandler.ListUsers)
	api.POST("/users", adminAPIHandler.CreateUser)
	api.PUT("/users/:id", adminAPIHandler.UpdateUser)
	api.DELETE("/users/:id", adminAPIHandler.DeleteUser)
	api.GET("/clients", adminAPIHandler.ListClients)
	api.POST("/clients", adminAPIHandler.CreateClient)
	api.PUT("/clients/:id", adminAPIHandler.UpdateClient)
	api.DELETE("/clients/:id", adminAPIHandler.DeleteClient)
	api.GET("/settings", adminAPIHandler.GetSettings)
	api.PUT("/settings", adminAPIHandler.UpdateSettings)

	// Serve Admin UI at root with HTML5 routing (must be last)
	adminFS := ui.GetAdminFS()
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       "/",
		Index:      "index.html",
		HTML5:      true,
		Browse:     false,
		Filesystem: http.FS(adminFS),
		Skipper: func(c echo.Context) bool {
			// Skip static serving for API routes and OpenID endpoints
			path := c.Request().URL.Path
			return path == "/authorize" ||
				path == "/token" ||
				path == "/userinfo" ||
				path == "/login" ||
				path == "/consent" ||
				len(path) >= 4 && path[:4] == "/api" ||
				len(path) >= 12 && path[:12] == "/.well-known"
		},
	}))
}

func getVersion() string {
	version := os.Getenv("VERSION")
	if version == "" {
		version = "dev"
	}
	return version
}
