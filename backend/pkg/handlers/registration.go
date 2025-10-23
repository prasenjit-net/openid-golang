package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
)

// Constants for client registration
const (
	applicationTypeWeb          = "web"
	applicationTypeNative       = "native"
	grantTypeAuthorizationCode  = "authorization_code"
	grantTypeImplicit           = "implicit"
	tokenEndpointAuthMethodNone = "none"
)

// Register handles dynamic client registration (POST /register)
// Implements RFC 7591 (OAuth 2.0 Dynamic Client Registration Protocol)
// and OpenID Connect Dynamic Client Registration 1.0
func (h *Handlers) Register(c echo.Context) error {
	// 1. Check if registration is enabled in config
	if !h.config.Registration.Enabled {
		return c.JSON(http.StatusForbidden, models.ClientRegistrationError{
			Error:            "registration_not_supported",
			ErrorDescription: "Dynamic client registration is not enabled on this server",
		})
	}

	// 2. Validate initial access token if required
	if h.config.Registration.RequireInitialAccessToken {
		token := extractBearerToken(c)
		if token == "" {
			return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
				Error:            "invalid_token",
				ErrorDescription: "Initial access token required for registration",
			})
		}

		// Validate and consume the initial access token
		initialToken, err := h.storage.GetInitialAccessToken(token)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ClientRegistrationError{
				Error:            "server_error",
				ErrorDescription: "Failed to validate initial access token",
			})
		}

		if initialToken == nil || initialToken.Used {
			return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
				Error:            "invalid_token",
				ErrorDescription: "Invalid or already used initial access token",
			})
		}

		// Check expiration
		if time.Now().After(initialToken.ExpiresAt) {
			return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
				Error:            "invalid_token",
				ErrorDescription: "Initial access token has expired",
			})
		}

		// Mark token as used (we'll update it after successful client creation)
		defer func() {
			now := time.Now()
			initialToken.Used = true
			initialToken.UsedAt = &now
			// UsedBy will be set after client creation
		}()

		// Store the initial token for later update
		c.Set("initial_access_token", initialToken)
	}

	// 3. Parse registration request
	var req models.ClientRegistrationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ClientRegistrationError{
			Error:            "invalid_request",
			ErrorDescription: "Invalid request body: " + err.Error(),
		})
	}

	// 4. Validate the registration request
	if err := h.validateRegistrationRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, *err)
	}

	// 4. Create the client from the request
	client, err := h.createClientFromRequest(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ClientRegistrationError{
			Error:            models.ErrInvalidClientMetadata,
			ErrorDescription: err.Error(),
		})
	}

	// 5. Store the client
	if err := h.storage.CreateClient(client); err != nil {
		return c.JSON(http.StatusInternalServerError, models.ClientRegistrationError{
			Error:            "server_error",
			ErrorDescription: "Failed to register client",
		})
	}

	// 6. Mark initial access token as used (if applicable)
	if initialToken, ok := c.Get("initial_access_token").(*models.InitialAccessToken); ok {
		initialToken.UsedBy = client.ID
		if err := h.storage.UpdateInitialAccessToken(initialToken); err != nil {
			// Log error but don't fail the registration
			// The client was already created successfully
		}
	}

	// 7. Build and return the registration response
	response := h.buildRegistrationResponse(client)
	return c.JSON(http.StatusCreated, response)
}

// validateRegistrationRequest validates the client registration request
func (h *Handlers) validateRegistrationRequest(req *models.ClientRegistrationRequest) *models.ClientRegistrationError {
	// redirect_uris is REQUIRED
	if len(req.RedirectURIs) == 0 {
		return &models.ClientRegistrationError{
			Error:            models.ErrInvalidRedirectURI,
			ErrorDescription: "redirect_uris is required and must contain at least one URI",
		}
	}

	// Validate all redirect URIs
	for _, uri := range req.RedirectURIs {
		if errMsg := validateRedirectURI(uri, req.ApplicationType); errMsg != "" {
			return &models.ClientRegistrationError{
				Error:            models.ErrInvalidRedirectURI,
				ErrorDescription: errMsg,
			}
		}
	}

	// Validate application_type if provided
	if req.ApplicationType != "" && req.ApplicationType != applicationTypeWeb && req.ApplicationType != applicationTypeNative {
		return &models.ClientRegistrationError{
			Error:            models.ErrInvalidClientMetadata,
			ErrorDescription: "application_type must be 'web' or 'native'",
		}
	}

	// Validate subject_type if provided
	if req.SubjectType != "" && req.SubjectType != "public" && req.SubjectType != "pairwise" {
		return &models.ClientRegistrationError{
			Error:            models.ErrInvalidClientMetadata,
			ErrorDescription: "subject_type must be 'public' or 'pairwise'",
		}
	}

	// Validate grant_types and response_types consistency
	if err := h.validateGrantTypesAndResponseTypes(req); err != nil {
		return err
	}

	// Validate token_endpoint_auth_method
	if err := h.validateTokenEndpointAuthMethod(req); err != nil {
		return err
	}

	// Validate URIs format
	if err := h.validateURIs(req); err != nil {
		return err
	}

	// Validate JWKS - can't have both jwks and jwks_uri
	if req.JWKS != "" && req.JWKSURI != "" {
		return &models.ClientRegistrationError{
			Error:            models.ErrInvalidClientMetadata,
			ErrorDescription: "Cannot specify both jwks and jwks_uri",
		}
	}

	// Validate JWKS is valid JSON if provided
	if req.JWKS != "" {
		var jwks map[string]interface{}
		if err := json.Unmarshal([]byte(req.JWKS), &jwks); err != nil {
			return &models.ClientRegistrationError{
				Error:            models.ErrInvalidClientMetadata,
				ErrorDescription: "jwks must be valid JSON",
			}
		}
	}

	return nil
}

// validateRedirectURI validates a redirect URI per OAuth 2.0 spec
func validateRedirectURI(uri string, applicationType string) string {
	// Check if empty
	if uri == "" {
		return "redirect_uri cannot be empty"
	}

	// Parse the URI
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return "redirect_uri must be a valid URI: " + uri
	}

	// Must be absolute URI (has scheme)
	if !parsedURI.IsAbs() {
		return "redirect_uri must be an absolute URI: " + uri
	}

	// Must not contain fragment
	if parsedURI.Fragment != "" {
		return "redirect_uri must not contain a fragment: " + uri
	}

	// For web applications (default), must be HTTPS unless localhost
	if applicationType == "" || applicationType == applicationTypeWeb {
		if parsedURI.Scheme != "https" {
			// Allow localhost for development
			if !isLocalhost(parsedURI.Host) {
				return "redirect_uri must use HTTPS scheme for web applications (except localhost): " + uri
			}
		}
	}

	// For native applications, custom schemes are allowed
	// No additional validation needed

	return ""
}

// isLocalhost checks if a host is localhost
func isLocalhost(host string) bool {
	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host == "localhost" || host == "127.0.0.1" || host == "[::1]"
}

// validateGrantTypesAndResponseTypes ensures consistency between grant_types and response_types
func (h *Handlers) validateGrantTypesAndResponseTypes(req *models.ClientRegistrationRequest) *models.ClientRegistrationError {
	grantTypes := req.GrantTypes
	responseTypes := req.ResponseTypes

	// Set defaults if not provided
	if len(grantTypes) == 0 {
		grantTypes = []string{"authorization_code"}
	}
	if len(responseTypes) == 0 {
		responseTypes = []string{"code"}
	}

	// Validate grant_types values
	validGrantTypes := map[string]bool{
		"authorization_code": true,
		"implicit":           true,
		"refresh_token":      true,
		"client_credentials": true,
		"password":           true, // Resource Owner Password Credentials
	}

	for _, gt := range grantTypes {
		if !validGrantTypes[gt] {
			return &models.ClientRegistrationError{
				Error:            models.ErrInvalidClientMetadata,
				ErrorDescription: "invalid grant_type: " + gt,
			}
		}
	}

	// Validate response_types values
	validResponseTypes := map[string]bool{
		"code":                true,
		"token":               true,
		"id_token":            true,
		"code token":          true,
		"code id_token":       true,
		"id_token token":      true,
		"code id_token token": true,
	}

	for _, rt := range responseTypes {
		if !validResponseTypes[rt] {
			return &models.ClientRegistrationError{
				Error:            models.ErrInvalidClientMetadata,
				ErrorDescription: "invalid response_type: " + rt,
			}
		}
	}

	// Validate consistency: if response_type includes "code", must have authorization_code grant
	hasCode := false
	for _, rt := range responseTypes {
		if strings.Contains(rt, "code") {
			hasCode = true
			break
		}
	}

	hasAuthCodeGrant := false
	for _, gt := range grantTypes {
		if gt == grantTypeAuthorizationCode {
			hasAuthCodeGrant = true
			break
		}
	}

	if hasCode && !hasAuthCodeGrant {
		return &models.ClientRegistrationError{
			Error:            models.ErrInvalidClientMetadata,
			ErrorDescription: "response_type includes 'code' but grant_types does not include 'authorization_code'",
		}
	}

	// Validate consistency: if response_type includes "token" or "id_token" (implicit), must have implicit grant
	hasImplicit := false
	for _, rt := range responseTypes {
		if rt == "token" || strings.Contains(rt, "id_token") {
			hasImplicit = true
			break
		}
	}

	hasImplicitGrant := false
	for _, gt := range grantTypes {
		if gt == "implicit" {
			hasImplicitGrant = true
			break
		}
	}

	if hasImplicit && !hasImplicitGrant {
		return &models.ClientRegistrationError{
			Error:            models.ErrInvalidClientMetadata,
			ErrorDescription: "response_type includes implicit flow types but grant_types does not include 'implicit'",
		}
	}

	return nil
}

// validateTokenEndpointAuthMethod validates the token endpoint authentication method
func (h *Handlers) validateTokenEndpointAuthMethod(req *models.ClientRegistrationRequest) *models.ClientRegistrationError {
	if req.TokenEndpointAuthMethod == "" {
		return nil // Will use default
	}

	validMethods := map[string]bool{
		"none":                true, // Public clients
		"client_secret_post":  true, // Client secret in POST body
		"client_secret_basic": true, // Client secret in Basic auth header
		"client_secret_jwt":   true, // JWT signed with client secret
		"private_key_jwt":     true, // JWT signed with private key
	}

	if !validMethods[req.TokenEndpointAuthMethod] {
		return &models.ClientRegistrationError{
			Error:            models.ErrInvalidClientMetadata,
			ErrorDescription: "invalid token_endpoint_auth_method: " + req.TokenEndpointAuthMethod,
		}
	}

	// If using "none", client must be using implicit or have no client_secret
	if req.TokenEndpointAuthMethod == tokenEndpointAuthMethodNone {
		// Check if all grant types are implicit or don't require token endpoint
		hasTokenEndpointGrant := false
		for _, gt := range req.GrantTypes {
			if gt == grantTypeAuthorizationCode || gt == "refresh_token" || gt == "client_credentials" || gt == "password" {
				hasTokenEndpointGrant = true
				break
			}
		}
		if hasTokenEndpointGrant {
			return &models.ClientRegistrationError{
				Error:            models.ErrInvalidClientMetadata,
				ErrorDescription: "token_endpoint_auth_method 'none' requires implicit grant only",
			}
		}
	}

	return nil
}

// validateURIs validates the format of various URI fields
func (h *Handlers) validateURIs(req *models.ClientRegistrationRequest) *models.ClientRegistrationError {
	uriFields := map[string]string{
		"client_uri":            req.ClientURI,
		"logo_uri":              req.LogoURI,
		"tos_uri":               req.TosURI,
		"policy_uri":            req.PolicyURI,
		"jwks_uri":              req.JWKSURI,
		"sector_identifier_uri": req.SectorIdentifierURI,
		"initiate_login_uri":    req.InitiateLoginURI,
	}

	for fieldName, uri := range uriFields {
		if uri == "" {
			continue
		}

		parsed, err := url.Parse(uri)
		if err != nil || !parsed.IsAbs() {
			return &models.ClientRegistrationError{
				Error:            models.ErrInvalidClientMetadata,
				ErrorDescription: fieldName + " must be a valid absolute URI",
			}
		}

		// Most URIs should use HTTPS
		if fieldName != "logo_uri" && parsed.Scheme != "https" && !isLocalhost(parsed.Host) {
			return &models.ClientRegistrationError{
				Error:            models.ErrInvalidClientMetadata,
				ErrorDescription: fieldName + " should use https",
			}
		}
	}

	// Validate request_uris
	for _, uri := range req.RequestURIs {
		parsed, err := url.Parse(uri)
		if err != nil || !parsed.IsAbs() {
			return &models.ClientRegistrationError{
				Error:            models.ErrInvalidClientMetadata,
				ErrorDescription: "request_uris must contain valid absolute URIs",
			}
		}
	}

	return nil
}

// createClientFromRequest creates a Client entity from the registration request
func (h *Handlers) createClientFromRequest(req *models.ClientRegistrationRequest) (*models.Client, error) {
	now := time.Now()
	clientID := uuid.New().String()

	// Determine if client is confidential (needs a secret)
	isConfidential := h.isConfidentialClient(req)

	var clientSecret string
	var secretExpiresAt int64 = 0

	if isConfidential {
		// Generate client secret
		secret, err := crypto.GenerateRandomString(64)
		if err != nil {
			return nil, err
		}
		clientSecret = secret
		// Secret never expires (0 means never)
		secretExpiresAt = 0
	}

	// Generate registration access token
	registrationAccessToken, err := crypto.GenerateRandomString(32)
	if err != nil {
		return nil, err
	}

	// Apply defaults
	applicationType := req.ApplicationType
	if applicationType == "" {
		applicationType = "web"
	}

	subjectType := req.SubjectType
	if subjectType == "" {
		subjectType = "public"
	}

	tokenEndpointAuthMethod := req.TokenEndpointAuthMethod
	if tokenEndpointAuthMethod == "" {
		if isConfidential {
			tokenEndpointAuthMethod = "client_secret_basic"
		} else {
			tokenEndpointAuthMethod = "none"
		}
	}

	grantTypes := req.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{"authorization_code"}
	}

	responseTypes := req.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []string{"code"}
	}

	idTokenSignedAlg := req.IDTokenSignedResponseAlg
	if idTokenSignedAlg == "" {
		idTokenSignedAlg = "RS256"
	}

	scope := req.Scope
	if scope == "" {
		scope = "openid"
	}

	// Create the client
	client := &models.Client{
		// Core fields
		ID:              clientID,
		Secret:          clientSecret,
		SecretExpiresAt: secretExpiresAt,
		RedirectURIs:    req.RedirectURIs,
		GrantTypes:      grantTypes,
		ResponseTypes:   responseTypes,
		Scope:           scope,

		// OIDC fields
		ApplicationType: applicationType,
		Contacts:        req.Contacts,
		ClientName:      req.ClientName,
		Name:            req.ClientName, // Legacy compatibility

		// URIs
		LogoURI:   req.LogoURI,
		ClientURI: req.ClientURI,
		PolicyURI: req.PolicyURI,
		TosURI:    req.TosURI,

		// JWK fields
		JWKSURI:             req.JWKSURI,
		JWKS:                req.JWKS,
		SectorIdentifierURI: req.SectorIdentifierURI,
		SubjectType:         subjectType,

		// ID Token preferences
		IDTokenSignedResponseAlg:    idTokenSignedAlg,
		IDTokenEncryptedResponseAlg: req.IDTokenEncryptedResponseAlg,
		IDTokenEncryptedResponseEnc: req.IDTokenEncryptedResponseEnc,

		// UserInfo preferences
		UserInfoSignedResponseAlg:    req.UserInfoSignedResponseAlg,
		UserInfoEncryptedResponseAlg: req.UserInfoEncryptedResponseAlg,
		UserInfoEncryptedResponseEnc: req.UserInfoEncryptedResponseEnc,

		// Request Object preferences
		RequestObjectSigningAlg:    req.RequestObjectSigningAlg,
		RequestObjectEncryptionAlg: req.RequestObjectEncryptionAlg,
		RequestObjectEncryptionEnc: req.RequestObjectEncryptionEnc,

		// Token endpoint auth
		TokenEndpointAuthMethod:     tokenEndpointAuthMethod,
		TokenEndpointAuthSigningAlg: req.TokenEndpointAuthSigningAlg,

		// Authentication requirements
		DefaultMaxAge:    req.DefaultMaxAge,
		RequireAuthTime:  req.RequireAuthTime,
		DefaultACRValues: req.DefaultACRValues,

		// Advanced features
		InitiateLoginURI: req.InitiateLoginURI,
		RequestURIs:      req.RequestURIs,

		// Software statement
		SoftwareID:        req.SoftwareID,
		SoftwareVersion:   req.SoftwareVersion,
		SoftwareStatement: req.SoftwareStatement,

		// Registration metadata
		RegistrationAccessToken: registrationAccessToken,
		ClientIDIssuedAt:        now.Unix(),
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	return client, nil
}

// isConfidentialClient determines if a client should be confidential (have a secret)
func (h *Handlers) isConfidentialClient(req *models.ClientRegistrationRequest) bool {
	// Public clients use implicit grant only or have token_endpoint_auth_method = "none"
	if req.TokenEndpointAuthMethod == tokenEndpointAuthMethodNone {
		return false
	}

	// Check if all grant types are implicit
	if len(req.GrantTypes) == 0 {
		return true // Default to confidential
	}

	for _, gt := range req.GrantTypes {
		if gt != grantTypeImplicit {
			return true // Any non-implicit grant needs a secret
		}
	}

	// All grants are implicit - public client
	return false
}

// buildRegistrationResponse creates the registration response
func (h *Handlers) buildRegistrationResponse(client *models.Client) models.ClientRegistrationResponse {
	// Build the registration_client_uri
	registrationClientURI := h.config.Issuer + h.config.Registration.Endpoint + "/" + client.ID

	response := models.ClientRegistrationResponse{
		Client:                  *client,
		RegistrationAccessToken: client.RegistrationAccessToken,
		RegistrationClientURI:   registrationClientURI,
	}

	// Clear the internal registration access token from the embedded client
	// (it's returned separately in the response)
	response.Client.RegistrationAccessToken = ""

	return response
}

// GetClientConfiguration handles reading client configuration (GET /register/:client_id)
// Implements RFC 7592 Section 2 (Client Read Request)
// and OpenID Connect Dynamic Client Registration 1.0 Section 4 (Client Read Request)
func (h *Handlers) GetClientConfiguration(c echo.Context) error {
	// 1. Check if registration is enabled
	if !h.config.Registration.Enabled {
		return c.JSON(http.StatusForbidden, models.ClientRegistrationError{
			Error:            "registration_not_supported",
			ErrorDescription: "Dynamic client registration is not enabled on this server",
		})
	}

	// 2. Extract client_id from URL parameter
	clientID := c.Param("client_id")
	if clientID == "" {
		return c.JSON(http.StatusBadRequest, models.ClientRegistrationError{
			Error:            "invalid_request",
			ErrorDescription: "client_id is required",
		})
	}

	// 3. Extract and validate Registration Access Token from Authorization header
	token := extractBearerToken(c)
	if token == "" {
		return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
			Error:            "invalid_token",
			ErrorDescription: "Registration access token required",
		})
	}

	// 4. Get client from storage
	client, err := h.storage.GetClientByID(clientID)
	if err != nil || client == nil {
		// Per spec: Never return 404, always return 401 for security reasons
		// (don't leak information about which clients exist)
		return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
			Error:            "invalid_token",
			ErrorDescription: "Invalid client or token",
		})
	}

	// 5. Validate registration access token matches
	if client.RegistrationAccessToken != token {
		return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
			Error:            "invalid_token",
			ErrorDescription: "Invalid registration access token",
		})
	}

	// 6. Build response with all client metadata
	response := h.buildRegistrationResponse(client)

	return c.JSON(http.StatusOK, response)
}

// extractBearerToken extracts the Bearer token from the Authorization header
func extractBearerToken(c echo.Context) string {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// UpdateClientConfiguration handles updating client configuration (PUT /register/:client_id)
// Implements RFC 7592 Section 3 (Client Update Request)
func (h *Handlers) UpdateClientConfiguration(c echo.Context) error {
	// 1. Check if registration is enabled
	if !h.config.Registration.Enabled {
		return c.JSON(http.StatusForbidden, models.ClientRegistrationError{
			Error:            "registration_not_supported",
			ErrorDescription: "Dynamic client registration is not enabled on this server",
		})
	}

	// 2. Extract client_id from URL parameter
	clientID := c.Param("client_id")
	if clientID == "" {
		return c.JSON(http.StatusBadRequest, models.ClientRegistrationError{
			Error:            "invalid_request",
			ErrorDescription: "client_id is required",
		})
	}

	// 3. Extract and validate Registration Access Token
	token := extractBearerToken(c)
	if token == "" {
		return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
			Error:            "invalid_token",
			ErrorDescription: "Registration access token required",
		})
	}

	// 4. Get existing client from storage
	existingClient, err := h.storage.GetClientByID(clientID)
	if err != nil || existingClient == nil {
		return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
			Error:            "invalid_token",
			ErrorDescription: "Invalid client or token",
		})
	}

	// 5. Validate registration access token matches
	if existingClient.RegistrationAccessToken != token {
		return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
			Error:            "invalid_token",
			ErrorDescription: "Invalid registration access token",
		})
	}

	// 6. Parse update request
	var req models.ClientRegistrationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.ClientRegistrationError{
			Error:            "invalid_request",
			ErrorDescription: "Invalid JSON in request body",
		})
	}

	// 7. Validate the update request (same validations as registration)
	if validationErr := h.validateRegistrationRequest(&req); validationErr != nil {
		return c.JSON(http.StatusBadRequest, *validationErr)
	}

	// 8. Update client while preserving certain fields
	updatedClient, err := h.createClientFromRequest(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ClientRegistrationError{
			Error:            models.ErrInvalidClientMetadata,
			ErrorDescription: err.Error(),
		})
	}

	// Preserve immutable fields from existing client
	updatedClient.ID = existingClient.ID
	updatedClient.Secret = existingClient.Secret
	updatedClient.RegistrationAccessToken = existingClient.RegistrationAccessToken
	updatedClient.CreatedAt = existingClient.CreatedAt
	updatedClient.UpdatedAt = time.Now()

	// 9. Save updated client
	if err := h.storage.UpdateClient(updatedClient); err != nil {
		return c.JSON(http.StatusInternalServerError, models.ClientRegistrationError{
			Error:            "server_error",
			ErrorDescription: "Failed to update client",
		})
	}

	// 10. Build and return response
	response := h.buildRegistrationResponse(updatedClient)
	return c.JSON(http.StatusOK, response)
}

// DeleteClientConfiguration handles deleting a client (DELETE /register/:client_id)
// Implements RFC 7592 Section 4 (Client Delete Request)
func (h *Handlers) DeleteClientConfiguration(c echo.Context) error {
	// 1. Check if registration is enabled
	if !h.config.Registration.Enabled {
		return c.JSON(http.StatusForbidden, models.ClientRegistrationError{
			Error:            "registration_not_supported",
			ErrorDescription: "Dynamic client registration is not enabled on this server",
		})
	}

	// 2. Extract client_id from URL parameter
	clientID := c.Param("client_id")
	if clientID == "" {
		return c.JSON(http.StatusBadRequest, models.ClientRegistrationError{
			Error:            "invalid_request",
			ErrorDescription: "client_id is required",
		})
	}

	// 3. Extract and validate Registration Access Token
	token := extractBearerToken(c)
	if token == "" {
		return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
			Error:            "invalid_token",
			ErrorDescription: "Registration access token required",
		})
	}

	// 4. Get client from storage
	client, err := h.storage.GetClientByID(clientID)
	if err != nil || client == nil {
		return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
			Error:            "invalid_token",
			ErrorDescription: "Invalid client or token",
		})
	}

	// 5. Validate registration access token matches
	if client.RegistrationAccessToken != token {
		return c.JSON(http.StatusUnauthorized, models.ClientRegistrationError{
			Error:            "invalid_token",
			ErrorDescription: "Invalid registration access token",
		})
	}

	// 6. Delete the client
	if err := h.storage.DeleteClient(clientID); err != nil {
		return c.JSON(http.StatusInternalServerError, models.ClientRegistrationError{
			Error:            "server_error",
			ErrorDescription: "Failed to delete client",
		})
	}

	// 7. Return 204 No Content on successful deletion
	return c.NoContent(http.StatusNoContent)
}
