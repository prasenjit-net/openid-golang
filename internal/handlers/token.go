package handlers

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/prasenjit-net/openid-golang/internal/crypto"
	"github.com/prasenjit-net/openid-golang/internal/models"
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
func (h *Handlers) Token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Failed to parse form")
		return
	}

	req := &TokenRequest{
		GrantType:    r.FormValue("grant_type"),
		Code:         r.FormValue("code"),
		RedirectURI:  r.FormValue("redirect_uri"),
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
		CodeVerifier: r.FormValue("code_verifier"),
		RefreshToken: r.FormValue("refresh_token"),
	}

	// Try to get client credentials from Authorization header
	if req.ClientID == "" || req.ClientSecret == "" {
		clientID, clientSecret, ok := parseBasicAuth(r.Header.Get("Authorization"))
		if ok {
			req.ClientID = clientID
			req.ClientSecret = clientSecret
		}
	}

	// Validate client
	client, err := h.storage.ValidateClient(req.ClientID, req.ClientSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_client", "Invalid client credentials")
		return
	}

	switch req.GrantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(w, r, req, client)
	case "refresh_token":
		h.handleRefreshTokenGrant(w, r, req, client)
	default:
		writeError(w, http.StatusBadRequest, "unsupported_grant_type", "Grant type not supported")
	}
}

func (h *Handlers) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request, req *TokenRequest, client *models.Client) {
	// Validate authorization code
	authCode, err := h.storage.GetAuthorizationCode(req.Code)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_grant", "Invalid authorization code")
		return
	}

	// Check if code is expired
	if authCode.IsExpired() {
		_ = h.storage.DeleteAuthorizationCode(req.Code)
		writeError(w, http.StatusBadRequest, "invalid_grant", "Authorization code expired")
		return
	}

	// Validate client ID
	if authCode.ClientID != req.ClientID {
		writeError(w, http.StatusBadRequest, "invalid_grant", "Client ID mismatch")
		return
	}

	// Validate redirect URI
	if authCode.RedirectURI != req.RedirectURI {
		writeError(w, http.StatusBadRequest, "invalid_grant", "Redirect URI mismatch")
		return
	}

	// Verify PKCE if used
	if authCode.CodeChallenge != "" {
		if req.CodeVerifier == "" {
			writeError(w, http.StatusBadRequest, "invalid_request", "code_verifier required")
			return
		}
		if !crypto.VerifyCodeChallenge(req.CodeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
			writeError(w, http.StatusBadRequest, "invalid_grant", "Invalid code_verifier")
			return
		}
	}

	// Get user
	user, err := h.storage.GetUserByID(authCode.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get user")
		return
	}

	// Create tokens
	token := models.NewToken(client.ID, user.ID, authCode.Scope, h.config.JWT.ExpiryMinutes)
	if createErr := h.storage.CreateToken(token); createErr != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create token")
		return
	}

	// Generate ID token
	idToken, err := h.jwtManager.GenerateIDToken(user, client.ID, authCode.Nonce)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to generate ID token")
		return
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

	writeJSON(w, http.StatusOK, response)
}

func (h *Handlers) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request, req *TokenRequest, client *models.Client) {
	// Get token by refresh token
	oldToken, err := h.storage.GetTokenByRefreshToken(req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_grant", "Invalid refresh token")
		return
	}

	// Validate client ID
	if oldToken.ClientID != req.ClientID {
		writeError(w, http.StatusBadRequest, "invalid_grant", "Client ID mismatch")
		return
	}

	// Get user
	user, err := h.storage.GetUserByID(oldToken.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get user")
		return
	}

	// Create new tokens
	newToken := models.NewToken(client.ID, user.ID, oldToken.Scope, h.config.JWT.ExpiryMinutes)
	if createErr := h.storage.CreateToken(newToken); createErr != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create token")
		return
	}

	// Generate new ID token
	idToken, err := h.jwtManager.GenerateIDToken(user, client.ID, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to generate ID token")
		return
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

	writeJSON(w, http.StatusOK, response)
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
