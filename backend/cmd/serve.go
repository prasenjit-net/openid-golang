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

	"github.com/prasenjit-net/openid-golang/pkg/config"
	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/handlers"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
	"github.com/prasenjit-net/openid-golang/pkg/ui"
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
	ctx := context.Background()

	// Try to auto-load config from config store (MongoDB env or JSON file)
	loaderCfg := configstore.LoaderConfig{
		MongoURIEnv:      "MONGODB_URI",
		MongoDatabaseEnv: "MONGODB_DATABASE",
		JSONFilePath:     "data/config.json",
	}

	configStoreInstance, initialized, err := configstore.AutoLoadConfigStore(ctx, loaderCfg)
	if err != nil {
		log.Fatalf("Failed to load config store: %v", err)
	}
	defer configStoreInstance.Close()

	// If not initialized, start in setup mode
	if !initialized {
		log.Println("Configuration not found. Starting in setup mode...")
		log.Println("Please visit http://localhost:8080/setup to configure the server")
		runSetupMode(configStoreInstance)
		return
	}

	// Load configuration from config store
	configData, err := configStoreInstance.GetConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Convert configstore.ConfigData to config.Config
	cfg := convertConfigData(configData)

	// Validate configuration
	if validateErr := cfg.Validate(); validateErr != nil {
		log.Fatalf("Invalid configuration: %v", validateErr)
	}

	// Start normal server with full OpenID functionality
	runNormalMode(cfg, configData)
}

// runSetupMode starts the server in setup mode with only /setup endpoint
func runSetupMode(configStoreInstance configstore.ConfigStore) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Setup wizard handler
	bootstrapHandler := handlers.NewBootstrapHandler(configStoreInstance)

	// Setup routes - wrap http.HandlerFunc for Echo
	e.GET("/setup", echo.WrapHandler(http.HandlerFunc(bootstrapHandler.ServeSetupWizard)))
	e.GET("/api/setup/status", echo.WrapHandler(http.HandlerFunc(bootstrapHandler.CheckInitialized)))
	e.POST("/api/setup/initialize", echo.WrapHandler(http.HandlerFunc(bootstrapHandler.Initialize))) // Redirect root to setup
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/setup")
	})

	// Start server
	addr := "0.0.0.0:8080"
	log.Printf("Starting OpenID Connect Server in SETUP mode")
	log.Printf("Setup wizard available at: http://localhost:8080/setup")

	// Start server with graceful shutdown
	go func() {
		if startErr := e.Start(addr); startErr != nil && startErr != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", startErr)
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

// runNormalMode starts the server in normal mode with full OpenID functionality
func runNormalMode(cfg *config.Config, configData *configstore.ConfigData) {
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
		if createErr := store.CreateClient(adminClient); createErr != nil {
			log.Printf("Warning: Failed to create admin-ui client: %v", createErr)
		} else {
			log.Println("Created admin-ui client")
		}
	}

	// Initialize JWT manager from PEM strings stored in config
	jwtManager, err := crypto.NewJWTManagerFromPEM(
		configData.JWT.PrivateKey,
		configData.JWT.PublicKey,
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

	// Register routes (without /setup - it's disabled in normal mode)
	registerRoutes(e, h, cfg)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting OpenID Connect Server v%s", getVersion())
	log.Printf("Using %s storage", cfg.Storage.Type)
	log.Printf("Starting OpenID Connect server on %s", addr)
	log.Printf("Issuer: %s", cfg.Issuer)

	// Start server with graceful shutdown
	go func() {
		if startErr := e.Start(addr); startErr != nil && startErr != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", startErr)
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

// convertConfigData converts configstore.ConfigData to config.Config
func convertConfigData(data *configstore.ConfigData) *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Host: data.Server.Host,
			Port: data.Server.Port,
		},
		Storage: config.StorageConfig{
			Type:         data.Storage.Type,
			JSONFilePath: data.Storage.JSONFilePath,
			MongoURI:     data.Storage.MongoURI,
		},
		JWT: config.JWTConfig{
			// Keys are stored as PEM strings in configstore, handled separately in runNormalMode
			PrivateKeyPath: "",
			PublicKeyPath:  "",
			ExpiryMinutes:  data.JWT.ExpiryMinutes,
		},
		Issuer: data.Issuer,
	}
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
	// Note: /setup is NOT served here - it's only available in setup mode
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
