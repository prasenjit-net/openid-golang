package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/prasenjit-net/openid-golang/pkg/configstore"
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
	// Storage backend configuration
	StorageType   string `json:"storage_type"`   // "json" or "mongodb"
	JSONFilePath  string `json:"json_file_path,omitempty"`
	MongoURI      string `json:"mongo_uri,omitempty"`
	MongoDatabase string `json:"mongo_database,omitempty"`

	// Server configuration
	Issuer string `json:"issuer"`
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`

	// JWT configuration
	JWTExpiryMinutes int  `json:"jwt_expiry_minutes,omitempty"`
	RefreshEnabled   bool `json:"refresh_enabled,omitempty"`
}

// SetupResponse represents the setup response
type SetupResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CheckInitialized checks if the application is already initialized
func (h *BootstrapHandler) CheckInitialized(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	initialized, err := h.configStore.IsInitialized(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"initialized": initialized,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Initialize performs the initial setup
func (h *BootstrapHandler) Initialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Check if already initialized
	initialized, err := h.configStore.IsInitialized(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if initialized {
		http.Error(w, "Already initialized", http.StatusBadRequest)
		return
	}

	// Parse request
	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateSetupRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate JWT key pair
	privateKey, publicKey, err := configstore.GenerateJWTKeyPair(4096)
	if err != nil {
		http.Error(w, "Failed to generate JWT keys: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create config
	config := buildConfigFromRequest(&req, privateKey, publicKey)

	// Initialize config store
	if err := h.configStore.Initialize(ctx); err != nil {
		http.Error(w, "Failed to initialize config store: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save config
	if err := h.configStore.SaveConfig(ctx, config); err != nil {
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SetupResponse{
		Success: true,
		Message: "Setup completed successfully",
	})
}

// validateSetupRequest validates the setup request
func validateSetupRequest(req *SetupRequest) error {
	if req.StorageType == "" {
		req.StorageType = "json"
	}

	if req.StorageType != "json" && req.StorageType != "mongodb" {
		return http.ErrNotSupported
	}

	if req.StorageType == "mongodb" && req.MongoURI == "" {
		return http.ErrNotSupported
	}

	if req.Issuer == "" {
		return http.ErrNotSupported
	}

	return nil
}

// buildConfigFromRequest builds a ConfigData from the setup request
func buildConfigFromRequest(req *SetupRequest, privateKey, publicKey string) *configstore.ConfigData {
	config := configstore.DefaultConfig()

	// Set issuer
	config.Issuer = req.Issuer

	// Set server config
	if req.Host != "" {
		config.Server.Host = req.Host
	}
	if req.Port > 0 {
		config.Server.Port = req.Port
	}

	// Set JWT config
	config.JWT.PrivateKey = privateKey
	config.JWT.PublicKey = publicKey
	if req.JWTExpiryMinutes > 0 {
		config.JWT.ExpiryMinutes = req.JWTExpiryMinutes
	}
	config.JWT.RefreshEnabled = req.RefreshEnabled

	// Set storage config
	config.Storage.Type = req.StorageType
	if req.StorageType == "json" {
		if req.JSONFilePath != "" {
			config.Storage.JSONFilePath = req.JSONFilePath
		} else {
			config.Storage.JSONFilePath = "data/openid.json"
		}
	} else if req.StorageType == "mongodb" {
		config.Storage.MongoURI = req.MongoURI
		if req.MongoDatabase != "" {
			config.Storage.MongoDatabase = req.MongoDatabase
		} else {
			config.Storage.MongoDatabase = "openid"
		}
	}

	return config
}
