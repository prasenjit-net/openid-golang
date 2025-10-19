package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/prasenjit-net/openid-golang/pkg/config"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
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
func (h *AdminHandler) GetStats(c echo.Context) error {
	users, err := h.store.GetAllUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get users"})
	}

	var clients []*models.Client
	clients, err = h.store.GetAllClients()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get clients"})
	}

	stats := map[string]interface{}{
		"users":   len(users),
		"clients": len(clients),
		"tokens":  0, // TODO: Count active tokens
		"logins":  0, // TODO: Count recent logins
	}

	return c.JSON(http.StatusOK, stats)
}

// ListUsers returns all users
func (h *AdminHandler) ListUsers(c echo.Context) error {
	users, err := h.store.GetAllUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get users"})
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

	return c.JSON(http.StatusOK, safeUsers)
}

// CreateUser creates a new user
func (h *AdminHandler) CreateUser(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Role     string `json:"role"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username, email, and password are required"})
	}

	// Set default role if not provided
	role := models.RoleUser
	if req.Role != "" {
		role = models.UserRole(req.Role)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
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
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user: " + err.Error()})
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

	return c.JSON(http.StatusCreated, response)
}

// UpdateUser updates an existing user
func (h *AdminHandler) UpdateUser(c echo.Context) error {
	var req struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		Role     string `json:"role"`
		Password string `json:"password,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Get existing user
	existingUser, err := h.store.GetUserByID(req.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user"})
	}
	if existingUser == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
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
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		}
		existingUser.PasswordHash = string(hashedPassword)
	}

	if err := h.store.UpdateUser(existingUser); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
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

	return c.JSON(http.StatusOK, response)
}

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser(c echo.Context) error {
	// Extract ID from URL parameter
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	if err := h.store.DeleteUser(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user: " + err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// ListClients returns all OAuth clients
func (h *AdminHandler) ListClients(c echo.Context) error {
	clients, err := h.store.GetAllClients()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get clients"})
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
			ClientSecret: "",        // Don't expose secret in list view
			Name:         client.Name,
			RedirectURIs: client.RedirectURIs,
			CreatedAt:    client.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// CreateClient creates a new OAuth client
func (h *AdminHandler) CreateClient(c echo.Context) error {
	var req struct {
		Name         string   `json:"name"`
		RedirectURIs []string `json:"redirect_uris"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Validate required fields
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}
	if len(req.RedirectURIs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "redirect_uris is required"})
	}

	// Generate client credentials
	clientID := uuid.New().String()
	clientSecret, err := crypto.GenerateRandomString(32)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate client secret"})
	}

	client := &models.Client{
		ID:            clientID,
		Secret:        clientSecret,
		Name:          req.Name,
		RedirectURIs:  req.RedirectURIs,
		GrantTypes:    []string{"authorization_code", "implicit"},
		ResponseTypes: []string{"code", "token", "id_token", "id_token token"},
		Scope:         "openid profile email",
		CreatedAt:     time.Now(),
	}

	if err := h.store.CreateClient(client); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create client: " + err.Error()})
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

	return c.JSON(http.StatusCreated, response)
}

// UpdateClient updates an existing OAuth client
func (h *AdminHandler) UpdateClient(c echo.Context) error {
	var req struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		RedirectURIs []string `json:"redirect_uris"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Get existing client
	existingClient, err := h.store.GetClientByID(req.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get client"})
	}
	if existingClient == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Client not found"})
	}

	// Update fields
	existingClient.Name = req.Name
	existingClient.RedirectURIs = req.RedirectURIs

	if err := h.store.UpdateClient(existingClient); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update client: " + err.Error()})
	}

	// Return updated client without secret
	response := map[string]interface{}{
		"id":            existingClient.ID,
		"client_id":     existingClient.ID,
		"name":          existingClient.Name,
		"redirect_uris": existingClient.RedirectURIs,
		"created_at":    existingClient.CreatedAt,
	}

	return c.JSON(http.StatusOK, response)
}

// DeleteClient deletes an OAuth client
func (h *AdminHandler) DeleteClient(c echo.Context) error {
	// Extract ID from URL parameter
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Client ID is required"})
	}

	if err := h.store.DeleteClient(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete client: " + err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetSettings returns server settings
func (h *AdminHandler) GetSettings(c echo.Context) error {
	settings := map[string]interface{}{
		"issuer":             h.config.Issuer,
		"server_host":        h.config.Server.Host,
		"server_port":        h.config.Server.Port,
		"storage_type":       h.config.Storage.Type,
		"json_file_path":     h.config.Storage.JSONFilePath,
		"mongo_uri":          h.config.Storage.MongoURI,
		"jwt_expiry_minutes": h.config.JWT.ExpiryMinutes,
		"jwt_private_key":    h.config.JWT.PrivateKeyPath,
		"jwt_public_key":     h.config.JWT.PublicKeyPath,
	}

	return c.JSON(http.StatusOK, settings)
}

// UpdateSettings updates server settings
func (h *AdminHandler) UpdateSettings(c echo.Context) error {
	var req struct {
		Issuer           string `json:"issuer"`
		ServerHost       string `json:"server_host"`
		ServerPort       int    `json:"server_port"`
		StorageType      string `json:"storage_type"`
		JSONFilePath     string `json:"json_file_path"`
		MongoURI         string `json:"mongo_uri"`
		JWTExpiryMinutes int    `json:"jwt_expiry_minutes"`
		JWTPrivateKey    string `json:"jwt_private_key"`
		JWTPublicKey     string `json:"jwt_public_key"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid configuration: " + err.Error()})
	}

	// Save to config.toml
	if err := h.config.SaveToTOML("config.toml"); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save configuration: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Settings updated successfully. Restart server for changes to take effect."})
}

// GetKeys returns signing keys
func (h *AdminHandler) GetKeys(c echo.Context) error {
	keys := []map[string]interface{}{
		{
			"kid":        "key-1",
			"alg":        "RS256",
			"use":        "sig",
			"created_at": time.Now().Format(time.RFC3339),
			"is_active":  true,
		},
	}

	return c.JSON(http.StatusOK, keys)
}

// RotateKeys rotates signing keys
func (h *AdminHandler) RotateKeys(c echo.Context) error {
	// TODO: Generate new keys and mark old keys as inactive
	return c.JSON(http.StatusOK, map[string]string{"message": "Keys rotated successfully"})
}

// GetSetupStatus returns whether initial setup is complete
func (h *AdminHandler) GetSetupStatus(c echo.Context) error {
	// TODO: Check if setup is complete (e.g., admin user exists)
	status := map[string]interface{}{
		"setupComplete": true,
		"authenticated": false,
	}

	return c.JSON(http.StatusOK, status)
}

// CompleteSetup performs initial setup
func (h *AdminHandler) CompleteSetup(c echo.Context) error {
	var req struct {
		Issuer        string `json:"issuer"`
		AdminUsername string `json:"adminUsername"`
		AdminPassword string `json:"adminPassword"`
		AdminEmail    string `json:"adminEmail"`
		AdminName     string `json:"adminName"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// TODO: Create admin user, save settings, generate initial keys
	return c.JSON(http.StatusOK, map[string]string{"message": "Setup completed successfully"})
}

// Login handles admin authentication
func (h *AdminHandler) Login(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Get user from storage
	user, err := h.store.GetUserByUsername(req.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	// Check if user exists
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	// Validate password using bcrypt
	if !crypto.ValidatePassword(req.Password, user.PasswordHash) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	// Check if user has admin role
	if !user.IsAdmin() {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied: Admin privileges required"})
	}

	// TODO: Generate proper session token with expiration
	response := map[string]string{
		"token": "dummy-session-token",
	}
	return c.JSON(http.StatusOK, response)
}
