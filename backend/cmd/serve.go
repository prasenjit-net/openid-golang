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

	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/handlers"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/session"
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
	defer func() {
		if err := configStoreInstance.Close(); err != nil {
			log.Printf("Error closing config store: %v", err)
		}
	}()

	// If not initialized, start in setup mode with hot-reload capability
	if !initialized {
		log.Println("Configuration not found. Starting in setup mode...")
		log.Println("Please visit http://localhost:8080/setup to configure the server")
		runSetupModeWithReload(configStoreInstance, loaderCfg)
		return
	}

	// Load configuration from config store
	configData, err := configStoreInstance.GetConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Start normal server with full OpenID functionality
	runNormalMode(configData)
}

// runSetupModeWithReload starts the server in setup mode and transitions to normal mode after initialization
func runSetupModeWithReload(configStoreInstance configstore.ConfigStore, loaderCfg configstore.LoaderConfig) {
	addr := "0.0.0.0:8080"

	// Channel to signal when initialization is complete
	reloadChan := make(chan bool, 1)

	// Start setup mode server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Setup wizard handler with reload callback
	bootstrapHandler := handlers.NewBootstrapHandlerWithCallback(configStoreInstance, func() {
		// Signal that initialization is complete
		select {
		case reloadChan <- true:
		default:
		}
	})

	// Setup routes
	e.GET("/setup", bootstrapHandler.ServeSetupWizard)
	e.GET("/api/setup/status", bootstrapHandler.CheckInitialized)
	e.POST("/api/setup/initialize", bootstrapHandler.Initialize)

	// Redirect root to setup
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/setup")
	})

	log.Printf("Starting OpenID Connect Server in SETUP mode")
	log.Printf("Setup wizard available at: http://localhost:8080/setup")

	// Start server with graceful shutdown
	go func() {
		if startErr := e.Start(addr); startErr != nil && startErr != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", startErr)
		}
	}()

	// Wait for either initialization complete or interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	defer signal.Stop(quit) // Cleanup signal notification

	select {
	case <-reloadChan:
		log.Println("Configuration initialized! Transitioning to normal mode...")

		// Gracefully shutdown setup mode server
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := e.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error shutting down setup server: %v", err)
		}

		// Small delay to ensure clean shutdown
		time.Sleep(500 * time.Millisecond)

		// Load config and start normal mode
		loadCtx, loadCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer loadCancel()

		configData, err := configStoreInstance.GetConfig(loadCtx)
		if err != nil {
			log.Fatalf("Failed to load configuration after initialization: %v", err)
		}

		log.Println("Restarting in NORMAL mode with full functionality...")
		runNormalMode(configData)

	case <-quit:
		log.Println("Shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := e.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
		log.Println("Server stopped")
	}
}

// runNormalMode starts the server in normal mode with full OpenID functionality
func runNormalMode(configData *configstore.ConfigData) {
	// Initialize storage
	store, err := storage.NewStorage(configData)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("Error closing storage: %v", err)
		}
	}()

	// Ensure admin-ui client exists
	adminClient, err := store.GetClientByID("admin-ui")
	if err != nil || adminClient == nil {
		adminClient = models.NewAdminUIClient(configData.Issuer)
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
		configData.Issuer,
		configData.JWT.ExpiryMinutes,
	)
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	// Create session manager
	sessionConfig := session.DefaultConfig(store)
	sessionConfig.CookieSecure = configData.Server.Port == 443 // Secure cookies for HTTPS
	sessionManager := session.NewManager(sessionConfig)

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(sessionManager.Middleware()) // Add session middleware

	// Initialize handlers
	h := handlers.NewHandlers(store, jwtManager, configData, sessionManager)

	// Register routes (without /setup - it's disabled in normal mode)
	registerRoutes(e, h, configData)

	// Start server
	addr := fmt.Sprintf("%s:%d", configData.Server.Host, configData.Server.Port)
	log.Printf("Starting OpenID Connect Server v%s", getVersion())
	log.Printf("Using %s storage", configData.Storage.Type)
	log.Printf("Starting OpenID Connect server on %s", addr)
	log.Printf("Issuer: %s", configData.Issuer)

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

func registerRoutes(e *echo.Echo, h *handlers.Handlers, cfg *configstore.ConfigData) {
	// OpenID Connect Discovery
	e.GET("/.well-known/openid-configuration", h.Discovery)
	e.GET("/.well-known/jwks.json", h.JWKS)

	// OAuth/OpenID endpoints
	e.GET("/authorize", h.Authorize)
	e.POST("/token", h.Token)
	e.POST("/revoke", h.Revoke)
	e.POST("/introspect", h.Introspect)
	e.GET("/userinfo", h.UserInfo)
	e.POST("/userinfo", h.UserInfo)

	// Dynamic Client Registration (if enabled)
	if cfg.Registration.Enabled {
		e.POST(cfg.Registration.Endpoint, h.Register)
		e.GET(cfg.Registration.Endpoint+"/:client_id", h.GetClientConfiguration)
		e.PUT(cfg.Registration.Endpoint+"/:client_id", h.UpdateClientConfiguration)
		e.DELETE(cfg.Registration.Endpoint+"/:client_id", h.DeleteClientConfiguration)
	}

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
	api.GET("/users/:id", adminAPIHandler.GetUser)
	api.POST("/users", adminAPIHandler.CreateUser)
	api.PUT("/users/:id", adminAPIHandler.UpdateUser)
	api.DELETE("/users/:id", adminAPIHandler.DeleteUser)
	api.GET("/clients", adminAPIHandler.ListClients)
	api.GET("/clients/:id", adminAPIHandler.GetClient)
	api.POST("/clients", adminAPIHandler.CreateClient)
	api.POST("/clients/:id/regenerate-secret", adminAPIHandler.RegenerateClientSecret)
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
