package handlers

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
)

const (
	// GrantTypeAuthorizationCode is the authorization code grant type
	GrantTypeAuthorizationCode = "authorization_code"
	// GrantTypeRefreshToken is the refresh token grant type
	GrantTypeRefreshToken = "refresh_token"
	// GrantTypeClientCredentials is the client credentials grant type
	GrantTypeClientCredentials = "client_credentials"
	// GrantTypePassword is the resource owner password credentials grant type
	GrantTypePassword = "password"

	// TokenTypeHintAccessToken is the access token type hint
	TokenTypeHintAccessToken = "access_token"
	// TokenTypeHintRefreshToken is the refresh token type hint
	TokenTypeHintRefreshToken = "refresh_token"
)

// TokenRequest represents a token request
type TokenRequest struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	CodeVerifier string
	RefreshToken string
	Scope        string // For client_credentials and password grants
	Username     string // For password grant
	Password     string // For password grant
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"` // For client_credentials grant
}

// Token handles the token endpoint (POST /token)
func (h *Handlers) Token(c echo.Context) error {
	req := &TokenRequest{
		GrantType:    c.FormValue("grant_type"),
		Code:         c.FormValue("code"),
		RedirectURI:  c.FormValue("redirect_uri"),
		ClientID:     c.FormValue("client_id"),
		ClientSecret: c.FormValue("client_secret"),
		CodeVerifier: c.FormValue("code_verifier"),
		RefreshToken: c.FormValue("refresh_token"),
		Scope:        c.FormValue("scope"),
		Username:     c.FormValue("username"),
		Password:     c.FormValue("password"),
	}

	// Try to get client credentials from Authorization header
	if req.ClientID == "" || req.ClientSecret == "" {
		clientID, clientSecret, ok := parseBasicAuth(c.Request().Header.Get("Authorization"))
		if ok {
			req.ClientID = clientID
			req.ClientSecret = clientSecret
		}
	}

	// Validate client
	client, err := h.storage.ValidateClient(req.ClientID, req.ClientSecret)
	if err != nil {
		return ErrorInvalidClientAuth(c, "Invalid client credentials")
	}

	switch req.GrantType {
	case GrantTypeAuthorizationCode:
		return h.handleAuthorizationCodeGrant(c, req, client)
	case GrantTypeRefreshToken:
		return h.handleRefreshTokenGrant(c, req, client)
	case GrantTypeClientCredentials:
		return h.handleClientCredentialsGrant(c, req, client)
	case GrantTypePassword:
		return h.handlePasswordGrant(c, req, client)
	default:
		return jsonError(c, http.StatusBadRequest, ErrorUnsupportedGrantType, "Grant type not supported")
	}
}

// validateAndMarkAuthCode validates authorization code and marks it as used
func (h *Handlers) validateAndMarkAuthCode(c echo.Context, req *TokenRequest) (*models.AuthorizationCode, error) {
	authCode, err := h.storage.GetAuthorizationCode(req.Code)
	if err != nil || authCode == nil {
		return nil, ErrorInvalidAuthorizationCode(c, "Invalid authorization code")
	}

	// Check if code has already been used (replay attack prevention)
	if authCode.Used {
		_ = h.storage.RevokeTokensByAuthCode(authCode.Code)
		_ = h.storage.DeleteAuthorizationCode(req.Code)
		return nil, ErrorInvalidAuthorizationCode(c, "Authorization code has already been used")
	}

	// Mark code as used immediately to prevent concurrent replay
	now := time.Now()
	authCode.Used = true
	authCode.UsedAt = &now
	_ = h.storage.UpdateAuthorizationCode(authCode)

	return authCode, nil
}

// validateAuthCodeConstraints validates expiry, client, redirect URI, and PKCE
func (h *Handlers) validateAuthCodeConstraints(c echo.Context, authCode *models.AuthorizationCode, req *TokenRequest) error {
	// Check if code is expired
	if authCode.IsExpired() {
		_ = h.storage.DeleteAuthorizationCode(req.Code)
		return ErrorInvalidAuthorizationCode(c, "Authorization code expired")
	}

	// Validate client ID
	if authCode.ClientID != req.ClientID {
		_ = h.storage.DeleteAuthorizationCode(req.Code)
		return ErrorInvalidAuthorizationCode(c, "Client ID mismatch")
	}

	// Validate redirect URI
	if authCode.RedirectURI != req.RedirectURI {
		_ = h.storage.DeleteAuthorizationCode(req.Code)
		return ErrorInvalidAuthorizationCode(c, "Redirect URI mismatch")
	}

	// Verify PKCE if used
	if authCode.CodeChallenge != "" {
		if req.CodeVerifier == "" {
			_ = h.storage.DeleteAuthorizationCode(req.Code)
			return jsonError(c, http.StatusBadRequest, ErrorInvalidRequest, "code_verifier required")
		}
		if !crypto.VerifyCodeChallenge(req.CodeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
			_ = h.storage.DeleteAuthorizationCode(req.Code)
			return ErrorInvalidAuthorizationCode(c, "Invalid code_verifier")
		}
	}

	return nil
}

// generateIDTokenForAuthCode generates ID token with session claims if available
func (h *Handlers) generateIDTokenForAuthCode(user *models.User, client *models.Client, authCode *models.AuthorizationCode) (string, error) {
	// Try to get user session for auth_time, acr, amr claims
	userSession, _ := h.storage.GetUserSessionByUserID(authCode.UserID)

	var idToken string
	var err error

	if userSession != nil && userSession.IsAuthenticated() {
		// Include auth_time, acr, amr from user session
		// No at_hash/c_hash needed for authorization code flow
		idToken, err = h.jwtManager.GenerateIDTokenWithClaims(
			user,
			client.ID,
			authCode.Nonce,
			authCode.Scope, // Pass scope for claim filtering
			userSession.AuthTime,
			userSession.ACR,
			userSession.AMR,
			"", // accessToken - not included in code flow
			"", // authCode - not included in code flow
		)
	} else {
		// Fallback to basic ID token without session-specific claims
		idToken, err = h.jwtManager.GenerateIDToken(user, client.ID, authCode.Nonce, authCode.Scope)
	}

	return idToken, err
}

func (h *Handlers) handleAuthorizationCodeGrant(c echo.Context, req *TokenRequest, client *models.Client) error {
	// Validate and mark authorization code as used
	authCode, err := h.validateAndMarkAuthCode(c, req)
	if err != nil {
		return err
	}

	// Validate constraints (expiry, client, redirect URI, PKCE)
	if validationErr := h.validateAuthCodeConstraints(c, authCode, req); validationErr != nil {
		return validationErr
	}

	// Get user
	user, err := h.storage.GetUserByID(authCode.UserID)
	if err != nil {
		_ = h.storage.DeleteAuthorizationCode(req.Code)
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to get user")
	}

	// Create tokens
	token := models.NewToken(client.ID, user.ID, authCode.Scope, h.config.JWT.ExpiryMinutes)
	token.AuthorizationCodeID = authCode.Code
	if createErr := h.storage.CreateToken(token); createErr != nil {
		_ = h.storage.DeleteAuthorizationCode(req.Code)
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to create token")
	}

	// Generate ID token with enhanced claims if user session exists
	idToken, err := h.generateIDTokenForAuthCode(user, client, authCode)
	if err != nil {
		_ = h.storage.DeleteAuthorizationCode(req.Code)
		_ = h.storage.DeleteToken(token.AccessToken)
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to generate ID token")
	}

	// Delete authorization code (one-time use completed)
	_ = h.storage.DeleteAuthorizationCode(req.Code)

	// Return token response
	response := TokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.config.JWT.ExpiryMinutes * 60,
		RefreshToken: token.RefreshToken,
		IDToken:      idToken,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handlers) handleRefreshTokenGrant(c echo.Context, req *TokenRequest, client *models.Client) error {
	// Get token by refresh token
	oldToken, err := h.storage.GetTokenByRefreshToken(req.RefreshToken)
	if err != nil {
		return jsonError(c, http.StatusBadRequest, ErrorInvalidGrant, "Invalid refresh token")
	}

	// Validate client ID
	if oldToken.ClientID != req.ClientID {
		return jsonError(c, http.StatusBadRequest, ErrorInvalidGrant, "Client ID mismatch")
	}

	// Get user
	user, userErr := h.storage.GetUserByID(oldToken.UserID)
	if userErr != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to get user")
	}

	// Create new tokens
	newToken := models.NewToken(client.ID, user.ID, oldToken.Scope, h.config.JWT.ExpiryMinutes)
	if createErr := h.storage.CreateToken(newToken); createErr != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to create token")
	}

	// Generate new ID token with scope filtering
	idToken, tokenErr := h.jwtManager.GenerateIDToken(user, client.ID, "", oldToken.Scope)
	if tokenErr != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to generate ID token")
	}

	// Delete old token
	_ = h.storage.DeleteToken(oldToken.AccessToken)

	// Return token response
	response := TokenResponse{
		AccessToken:  newToken.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.config.JWT.ExpiryMinutes * 60,
		RefreshToken: newToken.RefreshToken,
		IDToken:      idToken,
	}

	return c.JSON(http.StatusOK, response)
}

// parseBasicAuth parses HTTP Basic Authentication credentials
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

// handleClientCredentialsGrant implements OAuth 2.0 Client Credentials Grant (RFC 6749 §4.4)
// This grant type is used for machine-to-machine (M2M) authentication where the client is the resource owner
func (h *Handlers) handleClientCredentialsGrant(c echo.Context, req *TokenRequest, client *models.Client) error {
	// 1. Validate client is authorized for this grant type
	if !client.HasGrantType("client_credentials") {
		return jsonError(c, http.StatusBadRequest, ErrorUnauthorizedClient,
			"Client not authorized for client_credentials grant")
	}

	// 2. Determine scope (use requested scope or default to client's scope)
	requestedScope := req.Scope
	if requestedScope == "" {
		requestedScope = client.Scope
	}

	// 3. Validate requested scope is subset of client's allowed scopes
	if !h.validateScope(requestedScope, client.Scope) {
		return jsonError(c, http.StatusBadRequest, ErrorInvalidScope,
			"Requested scope exceeds client allowed scope")
	}

	// 4. Generate access token (NO user - client is the resource owner)
	token := models.NewToken(client.ID, "", requestedScope, h.config.JWT.ExpiryMinutes)
	if err := h.storage.CreateToken(token); err != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError,
			"Failed to create token")
	}

	// 5. Return token response
	// Note: No refresh token per RFC 6749 §4.4.3
	// Note: No ID token (this is not an OpenID Connect flow)
	response := TokenResponse{
		AccessToken: token.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   h.config.JWT.ExpiryMinutes * 60,
		Scope:       token.Scope,
	}

	return c.JSON(http.StatusOK, response)
}

// validateScope checks if requested scope is a subset of allowed scope
func (h *Handlers) validateScope(requested, allowed string) bool {
	if requested == "" {
		return true
	}

	requestedScopes := strings.Split(requested, " ")
	allowedScopes := strings.Split(allowed, " ")

	// Create map of allowed scopes for quick lookup
	allowedMap := make(map[string]bool)
	for _, scope := range allowedScopes {
		if scope != "" {
			allowedMap[scope] = true
		}
	}

	// Check each requested scope is in allowed scopes
	for _, scope := range requestedScopes {
		if scope != "" && !allowedMap[scope] {
			return false
		}
	}

	return true
}

// handlePasswordGrant implements RFC 6749 §4.3 Resource Owner Password Credentials Grant
// WARNING: This grant type is deprecated in OAuth 2.1 and should only be used when
// other more secure flows are not viable (e.g., legacy applications)
func (h *Handlers) handlePasswordGrant(c echo.Context, req *TokenRequest, client *models.Client) error {
	// RFC 6749 §4.3.2: Validate that client is authorized for this grant type
	hasPasswordGrant := false
	for _, gt := range client.GrantTypes {
		if gt == GrantTypePassword {
			hasPasswordGrant = true
			break
		}
	}
	if !hasPasswordGrant {
		return jsonError(c, http.StatusUnauthorized, ErrorUnauthorizedClient,
			"Client not authorized for password grant")
	}

	// RFC 6749 §4.3.2: username and password are REQUIRED
	if req.Username == "" || req.Password == "" {
		return jsonError(c, http.StatusBadRequest, ErrorInvalidRequest,
			"username and password parameters are required")
	}

	// Authenticate the user
	user, err := h.storage.GetUserByUsername(req.Username)
	if err != nil || user == nil {
		return jsonError(c, http.StatusUnauthorized, ErrorInvalidGrant,
			"Invalid username or password")
	}

	// Verify password
	if !crypto.ValidatePassword(req.Password, user.PasswordHash) {
		return jsonError(c, http.StatusUnauthorized, ErrorInvalidGrant,
			"Invalid username or password")
	}

	// Determine scope
	// If scope is requested, validate it against client's allowed scope
	scope := req.Scope
	if scope == "" {
		// Default to client's scope if none specified
		scope = client.Scope
	} else {
		// Validate requested scope is subset of client's allowed scope
		if !h.validateScope(scope, client.Scope) {
			return jsonError(c, http.StatusBadRequest, ErrorInvalidScope,
				"Requested scope exceeds client's allowed scope")
		}
	}

	// Generate tokens
	token := models.NewToken(client.ID, user.ID, scope, h.config.JWT.ExpiryMinutes)

	err = h.storage.CreateToken(token)
	if err != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to save token")
	}

	// Generate ID token if openid scope is requested
	var idToken string
	if strings.Contains(scope, "openid") {
		idToken, err = h.jwtManager.GenerateIDToken(user, client.ID, "", scope)
		if err != nil {
			return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to generate ID token")
		}
	}

	// Return token response
	return c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    h.config.JWT.ExpiryMinutes * 60,
		RefreshToken: token.RefreshToken,
		IDToken:      idToken,
		Scope:        scope,
	})
}
