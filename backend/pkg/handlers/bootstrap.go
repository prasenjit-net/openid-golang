package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/ui"
)

// BootstrapHandler handles the initial setup wizard
type BootstrapHandler struct {
	configStore configstore.ConfigStore
}

// NewBootstrapHandler creates a new bootstrap handler
func NewBootstrapHandler(configStore configstore.ConfigStore) *BootstrapHandler {
	return &BootstrapHandler{
		configStore: configStore,
	}
}

// SetupRequest represents the initial setup request
type SetupRequest struct {
	Issuer string `json:"issuer"`
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

	// Initialize with minimal config
	if err := configstore.InitializeMinimalConfig(ctx, h.configStore, req.Issuer); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to initialize: " + err.Error(),
		})
	}

	// Return success
	return c.JSON(http.StatusOK, SetupResponse{
		Success: true,
		Message: "Setup completed successfully. Reloading...",
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
