package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
)

// AdminHandler handles admin API endpoints
type AdminHandler struct {
	store  storage.Storage
	config *configstore.ConfigData
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(store storage.Storage, cfg *configstore.ConfigData) *AdminHandler {
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

	// Count active tokens (non-expired)
	activeTokens := h.store.GetActiveTokensCount()

	// Count recent user sessions (last 24 hours)
	recentLogins := h.store.GetRecentUserSessionsCount()

	// Get signing keys statistics
	allKeys, err := h.store.GetAllSigningKeys()
	totalKeys := 0
	activeKeys := 0
	if err == nil {
		totalKeys = len(allKeys)
		for _, key := range allKeys {
			if key.IsActive && !key.IsExpired() {
				activeKeys++
			}
		}
	}

	stats := map[string]interface{}{
		"users":       len(users),
		"clients":     len(clients),
		"tokens":      activeTokens,
		"logins":      recentLogins,
		"total_keys":  totalKeys,
		"active_keys": activeKeys,
	}

	return c.JSON(http.StatusOK, stats)
}

// ListUsers returns all users with optional filtering
func (h *AdminHandler) ListUsers(c echo.Context) error {
	users, err := h.store.GetAllUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get users"})
	}

	// Get filter parameters from query string
	usernameFilter := c.QueryParam("username")
	emailFilter := c.QueryParam("email")
	nameFilter := c.QueryParam("name")
	roleFilter := c.QueryParam("role")

	// Filter users based on query parameters
	var filteredUsers []*models.User
	for _, user := range users {
		// Apply filters (case-insensitive partial match)
		if usernameFilter != "" && !containsIgnoreCase(user.Username, usernameFilter) {
			continue
		}
		if emailFilter != "" && !containsIgnoreCase(user.Email, emailFilter) {
			continue
		}
		if nameFilter != "" && !containsIgnoreCase(user.Name, nameFilter) {
			continue
		}
		if roleFilter != "" && string(user.Role) != roleFilter {
			continue
		}
		filteredUsers = append(filteredUsers, user)
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

	safeUsers := make([]SafeUser, len(filteredUsers))
	for i, user := range filteredUsers {
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

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// GetUser returns a single user by ID
func (h *AdminHandler) GetUser(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	user, err := h.store.GetUserByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user"})
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	// Don't send password hash to client, but include all profile fields
	response := map[string]interface{}{
		"id":                    user.ID,
		"username":              user.Username,
		"email":                 user.Email,
		"email_verified":        user.EmailVerified,
		"name":                  user.Name,
		"given_name":            user.GivenName,
		"family_name":           user.FamilyName,
		"middle_name":           user.MiddleName,
		"nickname":              user.Nickname,
		"preferred_username":    user.PreferredUsername,
		"profile":               user.Profile,
		"picture":               user.Picture,
		"website":               user.Website,
		"gender":                user.Gender,
		"birthdate":             user.Birthdate,
		"zoneinfo":              user.Zoneinfo,
		"locale":                user.Locale,
		"phone_number":          user.PhoneNumber,
		"phone_number_verified": user.PhoneNumberVerified,
		"address":               user.Address,
		"role":                  user.Role,
		"created_at":            user.CreatedAt,
		"updated_at":            user.UpdatedAt,
	}

	return c.JSON(http.StatusOK, response)
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
		models.User
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

	// Update basic fields
	existingUser.Username = req.Username
	existingUser.Email = req.Email
	existingUser.EmailVerified = req.EmailVerified
	existingUser.Name = req.Name
	if req.Role != "" {
		existingUser.Role = req.Role
	}

	// Update profile fields
	existingUser.GivenName = req.GivenName
	existingUser.FamilyName = req.FamilyName
	existingUser.MiddleName = req.MiddleName
	existingUser.Nickname = req.Nickname
	existingUser.PreferredUsername = req.PreferredUsername
	existingUser.Profile = req.Profile
	existingUser.Picture = req.Picture
	existingUser.Website = req.Website
	existingUser.Gender = req.Gender
	existingUser.Birthdate = req.Birthdate
	existingUser.Zoneinfo = req.Zoneinfo
	existingUser.Locale = req.Locale

	// Update contact fields
	existingUser.PhoneNumber = req.PhoneNumber
	existingUser.PhoneNumberVerified = req.PhoneNumberVerified

	// Update address
	existingUser.Address = req.Address

	// Update password if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		}
		existingUser.PasswordHash = string(hashedPassword)
	}

	existingUser.UpdatedAt = time.Now()

	if err := h.store.UpdateUser(existingUser); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
	}

	// Return user without password hash
	response := map[string]interface{}{
		"id":                    existingUser.ID,
		"username":              existingUser.Username,
		"email":                 existingUser.Email,
		"email_verified":        existingUser.EmailVerified,
		"name":                  existingUser.Name,
		"given_name":            existingUser.GivenName,
		"family_name":           existingUser.FamilyName,
		"middle_name":           existingUser.MiddleName,
		"nickname":              existingUser.Nickname,
		"preferred_username":    existingUser.PreferredUsername,
		"profile":               existingUser.Profile,
		"picture":               existingUser.Picture,
		"website":               existingUser.Website,
		"gender":                existingUser.Gender,
		"birthdate":             existingUser.Birthdate,
		"zoneinfo":              existingUser.Zoneinfo,
		"locale":                existingUser.Locale,
		"phone_number":          existingUser.PhoneNumber,
		"phone_number_verified": existingUser.PhoneNumberVerified,
		"address":               existingUser.Address,
		"role":                  existingUser.Role,
		"created_at":            existingUser.CreatedAt,
		"updated_at":            existingUser.UpdatedAt,
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

	// Get query parameters for filtering
	filterClientID := c.QueryParam("client_id")
	filterName := c.QueryParam("name")

	// Apply filters if provided
	var filteredClients []*models.Client
	for _, client := range clients {
		// Skip if doesn't match filters
		if filterClientID != "" && !containsIgnoreCase(client.ID, filterClientID) {
			continue
		}
		if filterName != "" && !containsIgnoreCase(client.Name, filterName) {
			continue
		}
		filteredClients = append(filteredClients, client)
	}

	// Convert to response format
	type ClientResponse struct {
		ID                      string    `json:"id"`
		ClientID                string    `json:"client_id"`
		ClientSecret            string    `json:"client_secret,omitempty"`
		Name                    string    `json:"name"`
		RedirectURIs            []string  `json:"redirect_uris"`
		GrantTypes              []string  `json:"grant_types"`
		ResponseTypes           []string  `json:"response_types"`
		Scope                   string    `json:"scope"`
		ApplicationType         string    `json:"application_type"`
		Contacts                []string  `json:"contacts,omitempty"`
		ClientURI               string    `json:"client_uri,omitempty"`
		LogoURI                 string    `json:"logo_uri,omitempty"`
		PolicyURI               string    `json:"policy_uri,omitempty"`
		TosURI                  string    `json:"tos_uri,omitempty"`
		JwksURI                 string    `json:"jwks_uri,omitempty"`
		TokenEndpointAuthMethod string    `json:"token_endpoint_auth_method"`
		CreatedAt               time.Time `json:"created_at"`
	}

	response := make([]ClientResponse, len(filteredClients))
	for i, client := range filteredClients {
		response[i] = ClientResponse{
			ID:                      client.ID,
			ClientID:                client.ID, // In our model, ID is the client_id
			ClientSecret:            "",        // Don't expose secret in list view
			Name:                    client.Name,
			RedirectURIs:            client.RedirectURIs,
			GrantTypes:              client.GrantTypes,
			ResponseTypes:           client.ResponseTypes,
			Scope:                   client.Scope,
			ApplicationType:         client.ApplicationType,
		Contacts:                client.Contacts,
		ClientURI:               client.ClientURI,
		LogoURI:                 client.LogoURI,
		PolicyURI:               client.PolicyURI,
		TosURI:                  client.TosURI,
		JwksURI:                 client.JWKSURI,
		TokenEndpointAuthMethod: client.TokenEndpointAuthMethod,
			CreatedAt:               client.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// CreateClient creates a new OAuth client
func (h *AdminHandler) CreateClient(c echo.Context) error {
	var req struct {
		Name            string   `json:"name"`
		RedirectURIs    []string `json:"redirect_uris"`
		GrantTypes      []string `json:"grant_types"`
		ResponseTypes   []string `json:"response_types"`
		Scope           string   `json:"scope"`
		ApplicationType string   `json:"application_type"`
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

	// Set defaults for optional fields
	grantTypes := req.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{"authorization_code", "implicit"}
	}
	responseTypes := req.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []string{"code", "token", "id_token", "id_token token"}
	}
	scope := req.Scope
	if scope == "" {
		scope = "openid profile email"
	}
	applicationType := req.ApplicationType
	if applicationType == "" {
		applicationType = "web"
	}

	// Generate client credentials
	clientID := uuid.New().String()
	clientSecret, err := crypto.GenerateRandomString(32)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate client secret"})
	}

	client := &models.Client{
		ID:              clientID,
		Secret:          clientSecret,
		Name:            req.Name,
		RedirectURIs:    req.RedirectURIs,
		GrantTypes:      grantTypes,
		ResponseTypes:   responseTypes,
		Scope:           scope,
		ApplicationType: applicationType,
		CreatedAt:       time.Now(),
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
		"grant_types":   client.GrantTypes,
		"response_types": client.ResponseTypes,
		"scope":         client.Scope,
		"application_type": client.ApplicationType,
		"created_at":    client.CreatedAt,
	}

	return c.JSON(http.StatusCreated, response)
}

// UpdateClient updates an existing OAuth client
func (h *AdminHandler) UpdateClient(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Client ID is required"})
	}

	var req struct {
		Name            string   `json:"name"`
		RedirectURIs    []string `json:"redirect_uris"`
		GrantTypes      []string `json:"grant_types"`
		ResponseTypes   []string `json:"response_types"`
		Scope           string   `json:"scope"`
		ApplicationType string   `json:"application_type"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Get existing client
	existingClient, err := h.store.GetClientByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get client"})
	}
	if existingClient == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Client not found"})
	}

	// Update fields
	if req.Name != "" {
		existingClient.Name = req.Name
	}
	if len(req.RedirectURIs) > 0 {
		existingClient.RedirectURIs = req.RedirectURIs
	}
	if len(req.GrantTypes) > 0 {
		existingClient.GrantTypes = req.GrantTypes
	}
	if len(req.ResponseTypes) > 0 {
		existingClient.ResponseTypes = req.ResponseTypes
	}
	if req.Scope != "" {
		existingClient.Scope = req.Scope
	}
	if req.ApplicationType != "" {
		existingClient.ApplicationType = req.ApplicationType
	}

	if err := h.store.UpdateClient(existingClient); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update client: " + err.Error()})
	}

	// Return updated client without secret
	response := map[string]interface{}{
		"id":               existingClient.ID,
		"client_id":        existingClient.ID,
		"name":             existingClient.Name,
		"redirect_uris":    existingClient.RedirectURIs,
		"grant_types":      existingClient.GrantTypes,
		"response_types":   existingClient.ResponseTypes,
		"scope":            existingClient.Scope,
		"application_type": existingClient.ApplicationType,
		"created_at":       existingClient.CreatedAt,
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

// GetClient returns a single OAuth client by ID
func (h *AdminHandler) GetClient(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Client ID is required"})
	}

	client, err := h.store.GetClientByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get client"})
	}
	if client == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Client not found"})
	}

	// Return client with all fields except secret
	response := map[string]interface{}{
		"id":                        client.ID,
		"client_id":                 client.ID,
		"name":                      client.Name,
		"redirect_uris":             client.RedirectURIs,
		"grant_types":               client.GrantTypes,
		"response_types":            client.ResponseTypes,
		"scope":                     client.Scope,
		"application_type":          client.ApplicationType,
		"contacts":                  client.Contacts,
		"client_uri":                client.ClientURI,
		"logo_uri":                  client.LogoURI,
		"policy_uri":                client.PolicyURI,
		"tos_uri":                   client.TosURI,
		"jwks_uri":                  client.JWKSURI,
		"token_endpoint_auth_method": client.TokenEndpointAuthMethod,
		"created_at":                client.CreatedAt,
	}

	return c.JSON(http.StatusOK, response)
}

// RegenerateClientSecret generates a new secret for an OAuth client
func (h *AdminHandler) RegenerateClientSecret(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Client ID is required"})
	}

	// Get existing client
	client, err := h.store.GetClientByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get client"})
	}
	if client == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Client not found"})
	}

	// Generate new secret
	newSecret, err := crypto.GenerateRandomString(32)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate client secret"})
	}

	// Update client with new secret
	client.Secret = newSecret
	if err := h.store.UpdateClient(client); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update client: " + err.Error()})
	}

	// Return new secret (only shown once)
	response := map[string]interface{}{
		"client_id":     client.ID,
		"client_secret": newSecret,
		"message":       "Client secret regenerated successfully",
	}

	return c.JSON(http.StatusOK, response)
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
		"jwt_private_key":    h.config.JWT.PrivateKey, // PEM string
		"jwt_public_key":     h.config.JWT.PublicKey,  // PEM string
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
		h.config.JWT.PrivateKey = req.JWTPrivateKey // PEM string
	}
	if req.JWTPublicKey != "" {
		h.config.JWT.PublicKey = req.JWTPublicKey // PEM string
	}

	// Note: ConfigData doesn't have Validate or SaveToTOML methods
	// These would need to be implemented if runtime config updates are required
	// For now, return success - config is in memory only
	return c.JSON(http.StatusOK, map[string]string{"message": "Settings updated in memory. Note: Changes are not persisted. Restart may revert changes."})
}

// GetKeys returns signing keys
func (h *AdminHandler) GetKeys(c echo.Context) error {
	keys, err := h.store.GetAllSigningKeys()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get signing keys"})
	}

	// Convert to response format (don't expose private keys in list)
	type KeyResponse struct {
		ID        string    `json:"id"`
		KID       string    `json:"kid"`
		Algorithm string    `json:"algorithm"`
		IsActive  bool      `json:"is_active"`
		CreatedAt time.Time `json:"created_at"`
		ExpiresAt time.Time `json:"expires_at,omitempty"`
		Status    string    `json:"status"` // "active", "expired", "inactive"
	}

	response := make([]KeyResponse, len(keys))
	for i, key := range keys {
		status := "inactive"
		if key.IsActive {
			status = "active"
		} else if key.IsExpired() {
			status = "expired"
		}

		response[i] = KeyResponse{
			ID:        key.ID,
			KID:       key.KID,
			Algorithm: key.Algorithm,
			IsActive:  key.IsActive,
			CreatedAt: key.CreatedAt,
			ExpiresAt: key.ExpiresAt,
			Status:    status,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// RotateKeys rotates signing keys
func (h *AdminHandler) RotateKeys(c echo.Context) error {
	// Generate new RSA key pair
	privateKey, publicKey, err := crypto.GenerateRSAKeyPair()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate new keys: " + err.Error()})
	}

	// Convert keys to PEM format
	privateKeyPEM, err := crypto.EncodePrivateKeyToPEM(privateKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to encode private key: " + err.Error()})
	}

	publicKeyPEM, err := crypto.EncodePublicKeyToPEM(publicKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to encode public key: " + err.Error()})
	}

	// Get all existing keys and mark them as inactive with expiration
	existingKeys, err := h.store.GetAllSigningKeys()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get existing keys: " + err.Error()})
	}

	// Set expiration for old keys (30 days from now to allow token validation)
	expirationTime := time.Now().Add(30 * 24 * time.Hour)
	for _, key := range existingKeys {
		if key.IsActive {
			key.IsActive = false
			key.ExpiresAt = expirationTime
			if err := h.store.UpdateSigningKey(key); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update old key: " + err.Error()})
			}
		}
	}

	// Create new active key
	newKey := &models.SigningKey{
		ID:         uuid.New().String(),
		KID:        fmt.Sprintf("key-%d", time.Now().Unix()),
		Algorithm:  "RS256",
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
		IsActive:   true,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Time{}, // No expiration for active key
	}

	if err := h.store.CreateSigningKey(newKey); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new key: " + err.Error()})
	}

	// Update config with new active key (for backward compatibility)
	h.config.JWT.PrivateKey = privateKeyPEM
	h.config.JWT.PublicKey = publicKeyPEM

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "RSA keys rotated successfully",
		"info":       "Old keys marked as inactive and will expire in 30 days",
		"new_key_id": newKey.KID,
		"active_key": map[string]interface{}{
			"id":         newKey.ID,
			"kid":        newKey.KID,
			"algorithm":  newKey.Algorithm,
			"created_at": newKey.CreatedAt,
		},
	})
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
