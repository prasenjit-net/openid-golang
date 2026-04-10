package handlers

import (
	"crypto/sha256"
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

const (
	bearerPrefix = "Bearer"
	unknownAdmin = "admin"
)

// AdminHandler handles admin API endpoints
type AdminHandler struct {
	store       storage.Storage
	config      *configstore.ConfigData
	adminSecret []byte // HMAC secret for admin JWT tokens
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(store storage.Storage, cfg *configstore.ConfigData) *AdminHandler {
	// Derive a stable HMAC secret from the JWT private key PEM.
	// Falls back to a constant salt when no key is configured yet.
	seed := []byte(cfg.JWT.PrivateKey)
	if len(seed) == 0 {
		seed = []byte("openid-admin-default-secret-seed")
	}
	sum := sha256.Sum256(seed)
	return &AdminHandler{
		store:       store,
		config:      cfg,
		adminSecret: sum[:],
	}
}

// getAdminActor extracts the authenticated admin's username from the Bearer
// token.  Returns unknownAdmin when the token is absent or cannot be parsed so
// that the audit actor field is never left blank.
func (h *AdminHandler) getAdminActor(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return unknownAdmin
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return unknownAdmin
	}
	claims, err := crypto.ValidateAdminToken(parts[1], h.adminSecret)
	if err != nil {
		return unknownAdmin
	}
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sub
	}
	return unknownAdmin
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

	h.logAdminAudit(models.AuditActionAdminUserCreated, models.AuditActorAdmin, h.getAdminActor(c),
		"user", user.ID, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(),
		map[string]interface{}{"created_username": req.Username})

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
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	var req struct {
		models.User
		Password string `json:"password,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Get existing user
	existingUser, err := h.store.GetUserByID(id)
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

	h.logAdminAudit(models.AuditActionAdminUserUpdated, models.AuditActorAdmin, h.getAdminActor(c),
		"user", existingUser.ID, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(), nil)

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

	h.logAdminAudit(models.AuditActionAdminUserDeleted, models.AuditActorAdmin, h.getAdminActor(c),
		"user", id, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(), nil)

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

	h.logAdminAudit(models.AuditActionAdminClientCreated, models.AuditActorAdmin, h.getAdminActor(c),
		"client", client.ID, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(),
		map[string]interface{}{"name": client.Name})

	// Return client with secret (only shown once)
	response := map[string]interface{}{
		"id":               client.ID,
		"client_id":        client.ID,
		"client_secret":    client.Secret,
		"name":             client.Name,
		"redirect_uris":    client.RedirectURIs,
		"grant_types":      client.GrantTypes,
		"response_types":   client.ResponseTypes,
		"scope":            client.Scope,
		"application_type": client.ApplicationType,
		"created_at":       client.CreatedAt,
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

	h.logAdminAudit(models.AuditActionAdminClientUpdated, models.AuditActorAdmin, h.getAdminActor(c),
		"client", id, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(), nil)

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

	h.logAdminAudit(models.AuditActionAdminClientDeleted, models.AuditActorAdmin, h.getAdminActor(c),
		"client", id, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(), nil)

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
		"id":                         client.ID,
		"client_id":                  client.ID,
		"name":                       client.Name,
		"redirect_uris":              client.RedirectURIs,
		"grant_types":                client.GrantTypes,
		"response_types":             client.ResponseTypes,
		"scope":                      client.Scope,
		"application_type":           client.ApplicationType,
		"contacts":                   client.Contacts,
		"client_uri":                 client.ClientURI,
		"logo_uri":                   client.LogoURI,
		"policy_uri":                 client.PolicyURI,
		"tos_uri":                    client.TosURI,
		"jwks_uri":                   client.JWKSURI,
		"token_endpoint_auth_method": client.TokenEndpointAuthMethod,
		"created_at":                 client.CreatedAt,
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
	h.logAdminAudit(models.AuditActionAdminSettingsUpdated, models.AuditActorAdmin, h.getAdminActor(c),
		"settings", "server", models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(), nil)

	return c.JSON(http.StatusOK, map[string]string{"message": "Settings updated in memory. Note: Changes are not persisted. Restart may revert changes."})
}

// GetKeys returns signing keys
func (h *AdminHandler) GetKeys(c echo.Context) error {
	keys, err := h.store.GetAllSigningKeys()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get signing keys"})
	}

	type CertInfo struct {
		Subject     string    `json:"subject"`
		Issuer      string    `json:"issuer"`
		Serial      string    `json:"serial"`
		NotBefore   time.Time `json:"not_before"`
		NotAfter    time.Time `json:"not_after"`
		Fingerprint string    `json:"fingerprint"` // x5t#S256
		SelfSigned  bool      `json:"self_signed"`
	}

	type KeyResponse struct {
		ID        string    `json:"id"`
		KID       string    `json:"kid"`
		Algorithm string    `json:"algorithm"`
		IsActive  bool      `json:"is_active"`
		CreatedAt time.Time `json:"created_at"`
		ExpiresAt time.Time `json:"expires_at,omitempty"`
		Status    string    `json:"status"` // "active", "expired", "inactive"
		Cert      *CertInfo `json:"cert,omitempty"`
		HasCSR    bool      `json:"has_csr"`
	}

	response := make([]KeyResponse, len(keys))
	for i, key := range keys {
		status := "inactive"
		if key.IsActive {
			status = "active"
		} else if key.IsExpired() {
			status = "expired"
		}

		kr := KeyResponse{
			ID:        key.ID,
			KID:       key.KID,
			Algorithm: key.Algorithm,
			IsActive:  key.IsActive,
			CreatedAt: key.CreatedAt,
			ExpiresAt: key.ExpiresAt,
			Status:    status,
			HasCSR:    key.CSRPEM != "",
		}

		if key.CertPEM != "" {
			if cert, err := crypto.ParseCertFromPEM(key.CertPEM); err == nil {
				fp, _ := crypto.CertThumbprintS256(key.CertPEM)
				kr.Cert = &CertInfo{
					Subject:     cert.Subject.CommonName,
					Issuer:      cert.Issuer.CommonName,
					Serial:      cert.SerialNumber.Text(16),
					NotBefore:   cert.NotBefore,
					NotAfter:    cert.NotAfter,
					Fingerprint: fp,
					SelfSigned:  cert.Subject.String() == cert.Issuer.String(),
				}
			}
		}

		response[i] = kr
	}

	return c.JSON(http.StatusOK, response)
}

// RotateKeys generates a new RSA key pair with a self-signed certificate and
// deactivates the current active key. Old keys remain in storage for JWT validation
// until their certificate expires. Accepts JSON body: {"validity_days": 90}
func (h *AdminHandler) RotateKeys(c echo.Context) error {
	// Parse optional body for validity_days (default 90 = ~3 months)
	var req struct {
		ValidityDays int `json:"validity_days"`
	}
	req.ValidityDays = 90
	_ = c.Bind(&req) // not fatal if body is empty
	if req.ValidityDays <= 0 {
		req.ValidityDays = 90
	}

	// Generate new key pair + self-signed certificate
	km, err := crypto.GenerateSigningKeyWithCert(req.ValidityDays)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate key: " + err.Error()})
	}

	// Deactivate existing active keys. Keep their ExpiresAt as-is so they remain
	// available for JWT verification until their own cert validity runs out.
	existingKeys, err := h.store.GetAllSigningKeys()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get existing keys: " + err.Error()})
	}
	for _, key := range existingKeys {
		if key.IsActive {
			key.IsActive = false
			// If the old key has no cert-based expiry, set a 90-day grace period
			if key.ExpiresAt.IsZero() {
				key.ExpiresAt = time.Now().Add(90 * 24 * time.Hour)
			}
			if err := h.store.UpdateSigningKey(key); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update old key: " + err.Error()})
			}
		}
	}

	// Persist new active key
	newKey := &models.SigningKey{
		ID:         uuid.New().String(),
		KID:        km.KID,
		Algorithm:  "RS256",
		PrivateKey: km.PrivateKeyPEM,
		PublicKey:  km.PublicKeyPEM,
		CertPEM:    km.CertPEM,
		IsActive:   true,
		CreatedAt:  time.Now(),
		ExpiresAt:  km.NotAfter, // cert validity drives key lifecycle
	}
	if err := h.store.CreateSigningKey(newKey); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create new key: " + err.Error()})
	}

	// Keep JWTManager in sync for tokens issued in this process lifetime
	h.config.JWT.PrivateKey = km.PrivateKeyPEM
	h.config.JWT.PublicKey = km.PublicKeyPEM

	h.logAdminAudit(models.AuditActionAdminKeysRotated, models.AuditActorAdmin, h.getAdminActor(c),
		"key", newKey.KID, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(),
		map[string]interface{}{"new_kid": newKey.KID, "validity_days": req.ValidityDays, "expires_at": newKey.ExpiresAt})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "RSA key rotated successfully",
		"info":          fmt.Sprintf("New key valid for %d days until %s", req.ValidityDays, km.NotAfter.Format("2006-01-02")),
		"new_key_id":    newKey.KID,
		"validity_days": req.ValidityDays,
		"not_before":    km.NotBefore,
		"not_after":     km.NotAfter,
		"active_key": map[string]interface{}{
			"id":         newKey.ID,
			"kid":        newKey.KID,
			"algorithm":  newKey.Algorithm,
			"created_at": newKey.CreatedAt,
			"expires_at": newKey.ExpiresAt,
		},
	})
}

// GenerateKeyCSR generates a PKCS#10 Certificate Signing Request for the signing key
// identified by :id and persists it on the key record. Returns the CSR as PEM text.
// GET /api/keys/:id/csr
func (h *AdminHandler) GenerateKeyCSR(c echo.Context) error {
	id := c.Param("id")
	key, err := h.store.GetSigningKey(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Key not found"})
	}

	csrPEM, err := crypto.GenerateCSR(key.PrivateKey, key.CertPEM)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate CSR: " + err.Error()})
	}

	// Persist CSR so it can be retrieved again without regenerating
	key.CSRPEM = csrPEM
	if err := h.store.UpdateSigningKey(key); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save CSR: " + err.Error()})
	}

	h.logAdminAudit(models.AuditActionAdminKeysRotated, models.AuditActorAdmin, h.getAdminActor(c),
		"key", key.KID, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(),
		map[string]interface{}{"action": "generate_csr", "kid": key.KID})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"kid":     key.KID,
		"csr_pem": csrPEM,
	})
}

// ImportKeyCert imports a CA-signed certificate for the signing key identified by :id.
// The certificate's public key must match the key's private key.
// Updating the certificate re-derives the KID from the new cert and resets ExpiresAt.
// POST /api/keys/:id/import-cert   Body: {"cert_pem": "-----BEGIN CERTIFICATE-----\n..."}
func (h *AdminHandler) ImportKeyCert(c echo.Context) error {
	id := c.Param("id")
	key, err := h.store.GetSigningKey(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Key not found"})
	}

	var req struct {
		CertPEM string `json:"cert_pem"`
	}
	if err := c.Bind(&req); err != nil || req.CertPEM == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "cert_pem is required"})
	}

	// Validate that the cert matches this key's private key
	if err := crypto.ValidateCertMatchesPrivateKey(req.CertPEM, key.PrivateKey); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "Certificate validation failed: " + err.Error()})
	}

	cert, err := crypto.ParseCertFromPEM(req.CertPEM)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid certificate PEM: " + err.Error()})
	}

	// Re-derive KID from the new cert thumbprint
	newKID, err := crypto.CertThumbprintS256(req.CertPEM)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to compute cert thumbprint: " + err.Error()})
	}

	oldKID := key.KID
	key.CertPEM = req.CertPEM
	key.KID = newKID
	key.ExpiresAt = cert.NotAfter

	if err := h.store.UpdateSigningKey(key); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update key: " + err.Error()})
	}

	h.logAdminAudit(models.AuditActionAdminKeysRotated, models.AuditActorAdmin, h.getAdminActor(c),
		"key", newKID, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(),
		map[string]interface{}{
			"action":       "import_cert",
			"old_kid":      oldKID,
			"new_kid":      newKID,
			"cert_subject": cert.Subject.CommonName,
			"cert_issuer":  cert.Issuer.CommonName,
			"not_before":   cert.NotBefore,
			"not_after":    cert.NotAfter,
		})

	fp, _ := crypto.CertThumbprintS256(req.CertPEM)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Certificate imported successfully",
		"kid":     newKID,
		"cert": map[string]interface{}{
			"subject":     cert.Subject.CommonName,
			"issuer":      cert.Issuer.CommonName,
			"serial":      cert.SerialNumber.Text(16),
			"not_before":  cert.NotBefore,
			"not_after":   cert.NotAfter,
			"fingerprint": fp,
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
		h.logAdminAudit(models.AuditActionAdminLogin, models.AuditActorAdmin, req.Username,
			"user", user.ID, models.AuditStatusFailure, c.RealIP(), c.Request().UserAgent(),
			map[string]interface{}{"reason": "not_admin"})
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied: Admin privileges required"})
	}

	h.logAdminAudit(models.AuditActionAdminLogin, models.AuditActorAdmin, req.Username,
		"user", user.ID, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(), nil)

	token, err := crypto.GenerateAdminToken(req.Username, h.adminSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}
	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

// GetProfile returns the current user's profile
func (h *AdminHandler) GetProfile(c echo.Context) error {
	// Extract user ID from JWT token in Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header required"})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != bearerPrefix {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization header format"})
	}

	tokenString := parts[1]

	// Parse and validate JWT token
	claims, err := crypto.ValidateAdminToken(tokenString, h.adminSecret)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	// Get user by username from claims
	username, ok := claims["sub"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token claims"})
	}

	// Find user by username
	users, err := h.store.GetAllUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user"})
	}

	var user *models.User
	for _, u := range users {
		if u.Username == username {
			user = u
			break
		}
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	// Return user profile without sensitive data
	profile := map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"name":     user.Name,
		"role":     user.Role,
	}

	return c.JSON(http.StatusOK, profile)
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UpdateProfile updates the current user's profile
func (h *AdminHandler) UpdateProfile(c echo.Context) error {
	// Extract user ID from JWT token
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header required"})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != bearerPrefix {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization header format"})
	}

	tokenString := parts[1]

	// Parse and validate JWT token
	claims, err := crypto.ValidateAdminToken(tokenString, h.adminSecret)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	username, ok := claims["sub"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token claims"})
	}

	// Parse request
	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Find user by username
	users, err := h.store.GetAllUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user"})
	}

	var user *models.User
	for _, u := range users {
		if u.Username == username {
			user = u
			break
		}
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	// Update fields
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}

	// Save updates
	if err := h.store.UpdateUser(user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update profile"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"name":     user.Name,
		"role":     user.Role,
	})
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ChangePassword changes the current user's password
func (h *AdminHandler) ChangePassword(c echo.Context) error {
	// Extract user ID from JWT token
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header required"})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != bearerPrefix {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization header format"})
	}

	tokenString := parts[1]

	// Parse and validate JWT token
	claims, err := crypto.ValidateAdminToken(tokenString, h.adminSecret)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	username, ok := claims["sub"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token claims"})
	}

	// Parse request
	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate passwords
	if req.CurrentPassword == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Current password is required"})
	}
	if req.NewPassword == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "New password is required"})
	}
	if len(req.NewPassword) < 6 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "New password must be at least 6 characters"})
	}

	// Find user by username
	users, err := h.store.GetAllUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user"})
	}

	var user *models.User
	for _, u := range users {
		if u.Username == username {
			user = u
			break
		}
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Current password is incorrect"})
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	if err := h.store.UpdateUser(user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update password"})
	}

	h.logAdminAudit(models.AuditActionAdminPasswordReset, models.AuditActorAdmin, username,
		"user", user.ID, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(), nil)

	return c.JSON(http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

// AdminTokenInfo is the admin view of a token, with sensitive values masked and username resolved.
type AdminTokenInfo struct {
	ID                  string    `json:"id"`
	AccessTokenPrefix   string    `json:"access_token_prefix"`   // first 12 chars + "..."
	RefreshTokenPresent bool      `json:"refresh_token_present"` // true if a refresh token exists
	TokenType           string    `json:"token_type"`
	ClientID            string    `json:"client_id"`
	UserID              string    `json:"user_id"`
	Username            string    `json:"username"`
	Scope               string    `json:"scope"`
	ExpiresAt           time.Time `json:"expires_at"`
	CreatedAt           time.Time `json:"created_at"`
	IsActive            bool      `json:"is_active"`
}

// ListTokens returns a paginated, filtered list of tokens for the admin UI.
// Query params: active (bool, default true), client_id, user_id
func (h *AdminHandler) ListTokens(c echo.Context) error {
	activeOnly := true
	if v := c.QueryParam("active"); v == "false" {
		activeOnly = false
	}
	clientID := c.QueryParam("client_id")
	userID := c.QueryParam("user_id")

	tokens, err := h.store.ListTokens(clientID, userID, activeOnly)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list tokens"})
	}

	// Build username cache to avoid repeated lookups
	usernameCache := map[string]string{}
	resolveUsername := func(uid string) string {
		if uid == "" {
			return ""
		}
		if name, ok := usernameCache[uid]; ok {
			return name
		}
		if user, err := h.store.GetUserByID(uid); err == nil && user != nil {
			usernameCache[uid] = user.Username
			return user.Username
		}
		usernameCache[uid] = uid
		return uid
	}

	now := time.Now()
	result := make([]AdminTokenInfo, 0, len(tokens))
	for _, t := range tokens {
		prefix := t.AccessToken
		if len(prefix) > 12 {
			prefix = prefix[:12] + "..."
		}
		result = append(result, AdminTokenInfo{
			ID:                  t.ID,
			AccessTokenPrefix:   prefix,
			RefreshTokenPresent: t.RefreshToken != "",
			TokenType:           t.TokenType,
			ClientID:            t.ClientID,
			UserID:              t.UserID,
			Username:            resolveUsername(t.UserID),
			Scope:               t.Scope,
			ExpiresAt:           t.ExpiresAt,
			CreatedAt:           t.CreatedAt,
			IsActive:            t.ExpiresAt.After(now),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tokens": result,
		"total":  len(result),
	})
}

// RevokeToken deletes a token by ID (admin revocation).
func (h *AdminHandler) RevokeToken(c echo.Context) error {
	tokenID := c.Param("id")
	if tokenID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Token ID is required"})
	}

	actor := h.getAdminActor(c)

	if err := h.store.DeleteToken(tokenID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to revoke token"})
	}

	h.logAdminAudit(models.AuditActionTokenRevoked, models.AuditActorAdmin, actor,
		"token", tokenID, models.AuditStatusSuccess, c.RealIP(), c.Request().UserAgent(), nil)

	return c.JSON(http.StatusOK, map[string]string{"message": "Token revoked"})
}
