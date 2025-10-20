package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/prasenjit-net/openid-golang/pkg/config"
	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
	"github.com/prasenjit-net/openid-golang/pkg/ui"
)

// BootstrapHandler handles the initial setup wizard
type BootstrapHandler struct {
	configStore    configstore.ConfigStore
	onInitComplete func() // Callback function when initialization is complete
}

// NewBootstrapHandler creates a new bootstrap handler
func NewBootstrapHandler(configStore configstore.ConfigStore) *BootstrapHandler {
	return &BootstrapHandler{
		configStore:    configStore,
		onInitComplete: nil,
	}
}

// NewBootstrapHandlerWithCallback creates a new bootstrap handler with initialization callback
func NewBootstrapHandlerWithCallback(configStore configstore.ConfigStore, onInitComplete func()) *BootstrapHandler {
	return &BootstrapHandler{
		configStore:    configStore,
		onInitComplete: onInitComplete,
	}
}

// SetupRequest represents the initial setup request
type SetupRequest struct {
	Issuer        string `json:"issuer"`
	AdminUsername string `json:"adminUsername,omitempty"`
	AdminPassword string `json:"adminPassword,omitempty"`
}

// SetupResponse represents the setup response
type SetupResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CheckInitialized checks if the application is already initialized
func (h *BootstrapHandler) CheckInitialized(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	initialized, err := h.configStore.IsInitialized(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]bool{
		"initialized": initialized,
	})
}

// Initialize performs the initial setup with minimal config (just issuer)
func (h *BootstrapHandler) Initialize(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()

	// Check if already initialized
	initialized, err := h.configStore.IsInitialized(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	if initialized {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Already initialized",
		})
	}

	// Parse request
	var req SetupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Issuer == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Issuer URL is required",
		})
	}

	// Validate admin user if provided
	if req.AdminUsername != "" && req.AdminPassword == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Admin password is required when username is provided",
		})
	}

	if req.AdminPassword != "" && req.AdminUsername == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Admin username is required when password is provided",
		})
	}

	if req.AdminPassword != "" && len(req.AdminPassword) < 6 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Admin password must be at least 6 characters long",
		})
	}

	// Initialize with minimal config
	if err := configstore.InitializeMinimalConfig(ctx, h.configStore, req.Issuer); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to initialize: " + err.Error(),
		})
	}

	// Create admin user if provided
	if req.AdminUsername != "" && req.AdminPassword != "" {
		if err := h.createAdminUser(ctx, req); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Config initialized but failed to create admin user: " + err.Error(),
			})
		}
	}

	// Trigger reload callback if configured (for hot-reload)
	if h.onInitComplete != nil {
		// Trigger the callback in a goroutine to avoid blocking the response
		go func() {
			time.Sleep(500 * time.Millisecond) // Small delay to ensure response is sent
			h.onInitComplete()
		}()
	}

	// Return success
	message := "Setup completed successfully. Server will reload automatically..."
	if req.AdminUsername != "" {
		message = "Setup completed successfully with admin user. Server will reload automatically..."
	}

	return c.JSON(http.StatusOK, SetupResponse{
		Success: true,
		Message: message,
	})
}

// ServeSetupWizard serves a simple HTML setup wizard
func (h *BootstrapHandler) ServeSetupWizard(c echo.Context) error {
	storageInfo := getStorageIndicator()
	
	html, err := ui.GetSetupHTML(storageInfo)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load setup wizard: "+err.Error())
	}
	
	return c.HTML(http.StatusOK, html)
}

func getStorageIndicator() string {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI != "" {
		return "<strong>MongoDB</strong> (from MONGODB_URI environment variable)"
	}
	return "<strong>JSON file</strong> (data/config.json)"
}

// createAdminUser creates an admin user in the storage
func (h *BootstrapHandler) createAdminUser(ctx context.Context, req SetupRequest) error {
	// Load the config that was just created
	configData, err := h.configStore.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Convert to config.Config for storage initialization
	cfg := &config.Config{
		Storage: config.StorageConfig{
			Type:         configData.Storage.Type,
			JSONFilePath: configData.Storage.JSONFilePath,
			MongoURI:     configData.Storage.MongoURI,
		},
	}

	// Initialize storage
	store, err := storage.NewStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Hash the password
	hashedPassword, err := crypto.HashPassword(req.AdminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	adminUser := &models.User{
		ID:           uuid.New().String(),
		Username:     req.AdminUsername,
		PasswordHash: hashedPassword,
		Email:        req.AdminUsername + "@local",
		Role:         "admin",
		Name:         "Administrator",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save user to storage
	if err := store.CreateUser(adminUser); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	return nil
}
