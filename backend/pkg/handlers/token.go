package handlers

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
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
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
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
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "invalid_client",
			"error_description": "Invalid client credentials",
		})
	}

	switch req.GrantType {
	case "authorization_code":
		return h.handleAuthorizationCodeGrant(c, req, client)
	case "refresh_token":
		return h.handleRefreshTokenGrant(c, req, client)
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "unsupported_grant_type",
			"error_description": "Grant type not supported",
		})
	}
}

func (h *Handlers) handleAuthorizationCodeGrant(c echo.Context, req *TokenRequest, client *models.Client) error {
	// Validate authorization code
	authCode, err := h.storage.GetAuthorizationCode(req.Code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "Invalid authorization code",
		})
	}

	// Check if code is expired
	if authCode.IsExpired() {
		_ = h.storage.DeleteAuthorizationCode(req.Code)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "Authorization code expired",
		})
	}

	// Validate client ID
	if authCode.ClientID != req.ClientID {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "Client ID mismatch",
		})
	}

	// Validate redirect URI
	if authCode.RedirectURI != req.RedirectURI {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "Redirect URI mismatch",
		})
	}

	// Verify PKCE if used
	if authCode.CodeChallenge != "" {
		if req.CodeVerifier == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":             "invalid_request",
				"error_description": "code_verifier required",
			})
		}
		if !crypto.VerifyCodeChallenge(req.CodeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":             "invalid_grant",
				"error_description": "Invalid code_verifier",
			})
		}
	}

	// Get user
	user, err := h.storage.GetUserByID(authCode.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to get user",
		})
	}

	// Create tokens
	token := models.NewToken(client.ID, user.ID, authCode.Scope, h.config.JWT.ExpiryMinutes)
	if createErr := h.storage.CreateToken(token); createErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to create token",
		})
	}

	// Generate ID token
	idToken, err := h.jwtManager.GenerateIDToken(user, client.ID, authCode.Nonce)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to generate ID token",
		})
	}

	// Delete authorization code (one-time use)
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
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "Invalid refresh token",
		})
	}

	// Validate client ID
	if oldToken.ClientID != req.ClientID {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "Client ID mismatch",
		})
	}

	// Get user
	user, err := h.storage.GetUserByID(oldToken.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to get user",
		})
	}

	// Create new tokens
	newToken := models.NewToken(client.ID, user.ID, oldToken.Scope, h.config.JWT.ExpiryMinutes)
	if createErr := h.storage.CreateToken(newToken); createErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to create token",
		})
	}

	// Generate new ID token
	idToken, err := h.jwtManager.GenerateIDToken(user, client.ID, "")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to generate ID token",
		})
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
