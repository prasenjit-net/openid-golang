package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prasenjit-net/openid-golang/internal/config"
	"github.com/prasenjit-net/openid-golang/internal/crypto"
	"github.com/prasenjit-net/openid-golang/internal/models"
	"github.com/prasenjit-net/openid-golang/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

// AdminHandler handles admin API endpoints
type AdminHandler struct {
	store  storage.Storage
	config *config.Config
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(store storage.Storage, cfg *config.Config) *AdminHandler {
	return &AdminHandler{
		store:  store,
		config: cfg,
	}
}

// GetStats returns dashboard statistics
func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := h.store.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	clients, err := h.store.GetAllClients()
	if err != nil {
		http.Error(w, "Failed to get clients", http.StatusInternalServerError)
		return
	}

	stats := map[string]interface{}{
		"users":   len(users),
		"clients": len(clients),
		"tokens":  0, // TODO: Count active tokens
		"logins":  0, // TODO: Count recent logins
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

	users, err := h.store.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	// Don't send password hashes to client
	type SafeUser struct {
		ID        string    `json:"id"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		Name      string    `json:"name"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"created_at"`
	}

	safeUsers := make([]SafeUser, len(users))
	for i, user := range users {
		safeUsers[i] = SafeUser{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Name:      user.Name,
			Role:      string(user.Role),
			CreatedAt: user.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(safeUsers)
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
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "username, email, and password are required", http.StatusBadRequest)
		return
	}

	// Set default role if not provided
	role := models.RoleUser
	if req.Role != "" {
		role = models.UserRole(req.Role)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		Name:         req.Name,
		Role:         role,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.store.CreateUser(user); err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return user without password hash
	response := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"name":       user.Name,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

// UpdateUser updates an existing user
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Role     string `json:"role"`
		Password string `json:"password,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get existing user
	existingUser, err := h.store.GetUserByID(req.ID)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	if existingUser == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update fields
	existingUser.Username = req.Username
	existingUser.Email = req.Email
	existingUser.Name = req.Name
	if req.Role != "" {
		existingUser.Role = models.UserRole(req.Role)
	}

	// Update password if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		existingUser.PasswordHash = string(hashedPassword)
	}

	if err := h.store.UpdateUser(existingUser); err != nil {
		http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return user without password hash
	response := map[string]interface{}{
		"id":         existingUser.ID,
		"username":   existingUser.Username,
		"email":      existingUser.Email,
		"name":       existingUser.Name,
		"role":       existingUser.Role,
		"created_at": existingUser.CreatedAt,
		"updated_at": existingUser.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := r.URL.Path[len("/api/admin/users/"):]
	if id == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteUser(id); err != nil {
		http.Error(w, "Failed to delete user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListClients returns all OAuth clients
func (h *AdminHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clients, err := h.store.GetAllClients()
	if err != nil {
		http.Error(w, "Failed to get clients", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	type ClientResponse struct {
		ID           string    `json:"id"`
		ClientID     string    `json:"client_id"`
		ClientSecret string    `json:"client_secret,omitempty"`
		Name         string    `json:"name"`
		RedirectURIs []string  `json:"redirect_uris"`
		CreatedAt    time.Time `json:"created_at"`
	}

	response := make([]ClientResponse, len(clients))
	for i, client := range clients {
		response[i] = ClientResponse{
			ID:           client.ID,
			ClientID:     client.ID, // In our model, ID is the client_id
			ClientSecret: "", // Don't expose secret in list view
			Name:         client.Name,
			RedirectURIs: client.RedirectURIs,
			CreatedAt:    client.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
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

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if len(req.RedirectURIs) == 0 {
		http.Error(w, "redirect_uris is required", http.StatusBadRequest)
		return
	}

	// Generate client credentials
	clientID := uuid.New().String()
	clientSecret, err := crypto.GenerateRandomString(32)
	if err != nil {
		http.Error(w, "Failed to generate client secret", http.StatusInternalServerError)
		return
	}

	client := &models.Client{
		ID:           clientID,
		Secret:       clientSecret,
		Name:         req.Name,
		RedirectURIs: req.RedirectURIs,
		GrantTypes:   []string{"authorization_code", "implicit"},
		ResponseTypes: []string{"code", "token", "id_token", "id_token token"},
		Scope:        "openid profile email",
		CreatedAt:    time.Now(),
	}

	if err := h.store.CreateClient(client); err != nil {
		http.Error(w, "Failed to create client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return client with secret (only shown once)
	response := map[string]interface{}{
		"id":            client.ID,
		"client_id":     client.ID,
		"client_secret": client.Secret,
		"name":          client.Name,
		"redirect_uris": client.RedirectURIs,
		"created_at":    client.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

// UpdateClient updates an existing OAuth client
func (h *AdminHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		RedirectURIs []string `json:"redirect_uris"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get existing client
	existingClient, err := h.store.GetClientByID(req.ID)
	if err != nil {
		http.Error(w, "Failed to get client", http.StatusInternalServerError)
		return
	}
	if existingClient == nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	// Update fields
	existingClient.Name = req.Name
	existingClient.RedirectURIs = req.RedirectURIs

	if err := h.store.UpdateClient(existingClient); err != nil {
		http.Error(w, "Failed to update client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated client without secret
	response := map[string]interface{}{
		"id":            existingClient.ID,
		"client_id":     existingClient.ID,
		"name":          existingClient.Name,
		"redirect_uris": existingClient.RedirectURIs,
		"created_at":    existingClient.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// DeleteClient deletes an OAuth client
func (h *AdminHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := r.URL.Path[len("/api/admin/clients/"):]
	if id == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteClient(id); err != nil {
		http.Error(w, "Failed to delete client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetSettings returns server settings
func (h *AdminHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	settings := map[string]interface{}{
		"issuer":          h.config.Issuer,
		"server_host":     h.config.Server.Host,
		"server_port":     h.config.Server.Port,
		"storage_type":    h.config.Storage.Type,
		"json_file_path":  h.config.Storage.JSONFilePath,
		"mongo_uri":       h.config.Storage.MongoURI,
		"jwt_expiry_minutes": h.config.JWT.ExpiryMinutes,
		"jwt_private_key": h.config.JWT.PrivateKeyPath,
		"jwt_public_key":  h.config.JWT.PublicKeyPath,
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

	var req struct {
		Issuer          string `json:"issuer"`
		ServerHost      string `json:"server_host"`
		ServerPort      int    `json:"server_port"`
		StorageType     string `json:"storage_type"`
		JSONFilePath    string `json:"json_file_path"`
		MongoURI        string `json:"mongo_uri"`
		JWTExpiryMinutes int   `json:"jwt_expiry_minutes"`
		JWTPrivateKey   string `json:"jwt_private_key"`
		JWTPublicKey    string `json:"jwt_public_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update config values
	if req.Issuer != "" {
		h.config.Issuer = req.Issuer
	}
	if req.ServerHost != "" {
		h.config.Server.Host = req.ServerHost
	}
	if req.ServerPort > 0 {
		h.config.Server.Port = req.ServerPort
	}
	if req.StorageType != "" {
		h.config.Storage.Type = req.StorageType
	}
	if req.JSONFilePath != "" {
		h.config.Storage.JSONFilePath = req.JSONFilePath
	}
	if req.MongoURI != "" {
		h.config.Storage.MongoURI = req.MongoURI
	}
	if req.JWTExpiryMinutes > 0 {
		h.config.JWT.ExpiryMinutes = req.JWTExpiryMinutes
	}
	if req.JWTPrivateKey != "" {
		h.config.JWT.PrivateKeyPath = req.JWTPrivateKey
	}
	if req.JWTPublicKey != "" {
		h.config.JWT.PublicKeyPath = req.JWTPublicKey
	}

	// Validate config
	if err := h.config.Validate(); err != nil {
		http.Error(w, "Invalid configuration: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Save to config.toml
	if err := h.config.SaveToTOML("config.toml"); err != nil {
		http.Error(w, "Failed to save configuration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Settings updated successfully. Restart server for changes to take effect."})
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
