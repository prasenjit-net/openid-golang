package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prasenjit-net/openid-golang/internal/crypto"
	"github.com/prasenjit-net/openid-golang/internal/storage"
)

// AdminHandler handles admin API endpoints
type AdminHandler struct {
	store storage.Storage
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(store storage.Storage) *AdminHandler {
	return &AdminHandler{store: store}
}

// GetStats returns dashboard statistics
func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := map[string]interface{}{
		"users":   5,  // TODO: Get real count from database
		"clients": 2,  // TODO: Get real count from database
		"tokens":  15, // TODO: Get real count from database
		"logins":  42, // TODO: Get real count from database
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

// ListUsers returns all users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get users from database
	users := []map[string]interface{}{
		{
			"id":         "1",
			"username":   "testuser",
			"email":      "test@example.com",
			"name":       "Test User",
			"created_at": time.Now().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}

// CreateUser creates a new user
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Validate and create user in database
	user := map[string]interface{}{
		"id":         "new-id",
		"username":   req.Username,
		"email":      req.Email,
		"name":       req.Name,
		"created_at": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Extract user ID from URL and delete from database
	w.WriteHeader(http.StatusNoContent)
}

// ListClients returns all OAuth clients
func (h *AdminHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Get clients from database
	clients := []map[string]interface{}{
		{
			"id":            "1",
			"client_id":     "test-client",
			"client_secret": "***hidden***",
			"name":          "Test Application",
			"redirect_uris": []string{"http://localhost:3000/callback"},
			"created_at":    time.Now().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(clients)
}

// CreateClient creates a new OAuth client
func (h *AdminHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name         string   `json:"name"`
		RedirectURIs []string `json:"redirect_uris"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Generate client ID and secret, save to database
	client := map[string]interface{}{
		"id":            "new-id",
		"client_id":     "generated-client-id",
		"client_secret": "generated-client-secret",
		"name":          req.Name,
		"redirect_uris": req.RedirectURIs,
		"created_at":    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(client)
}

// DeleteClient deletes an OAuth client
func (h *AdminHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Extract client ID from URL and delete from database
	w.WriteHeader(http.StatusNoContent)
}

// GetSettings returns server settings
func (h *AdminHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	settings := map[string]interface{}{
		"issuer":             "http://localhost:8080",
		"token_ttl":          3600,
		"refresh_token_ttl":  2592000,
		"jwks_rotation_days": 90,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(settings)
}

// UpdateSettings updates server settings
func (h *AdminHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Validate and save settings to database
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Settings updated successfully"})
}

// GetKeys returns signing keys
func (h *AdminHandler) GetKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keys := []map[string]interface{}{
		{
			"kid":        "key-1",
			"alg":        "RS256",
			"use":        "sig",
			"created_at": time.Now().Format(time.RFC3339),
			"is_active":  true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(keys)
}

// RotateKeys rotates signing keys
func (h *AdminHandler) RotateKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Generate new keys and mark old keys as inactive
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Keys rotated successfully"})
}

// GetSetupStatus returns whether initial setup is complete
func (h *AdminHandler) GetSetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Check if setup is complete (e.g., admin user exists)
	status := map[string]interface{}{
		"setupComplete": true,
		"authenticated": false,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

// CompleteSetup performs initial setup
func (h *AdminHandler) CompleteSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Issuer        string `json:"issuer"`
		AdminUsername string `json:"adminUsername"`
		AdminPassword string `json:"adminPassword"`
		AdminEmail    string `json:"adminEmail"`
		AdminName     string `json:"adminName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Create admin user, save settings, generate initial keys
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Setup completed successfully"})
}

// Login handles admin authentication
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user from storage
	user, err := h.store.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if user exists
	if user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Validate password using bcrypt
	if !crypto.ValidatePassword(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check if user has admin role
	if !user.IsAdmin() {
		http.Error(w, "Access denied: Admin privileges required", http.StatusForbidden)
		return
	}

	// TODO: Generate proper session token with expiration
	response := map[string]string{
		"token": "dummy-session-token",
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
