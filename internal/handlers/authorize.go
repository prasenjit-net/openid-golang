package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/prasenjit-net/openid-golang/internal/crypto"
	"github.com/prasenjit-net/openid-golang/internal/models"
)

const (
	// ResponseTypeCode is the authorization code response type
	ResponseTypeCode = "code"
	// ResponseTypeIDToken is the ID token response type (implicit flow)
	ResponseTypeIDToken = "id_token"
	// ResponseTypeTokenIDToken is the access token + ID token response type
	ResponseTypeTokenIDToken = "token id_token"
)

// Authorize handles the authorization endpoint (GET /authorize)
func (h *Handlers) Authorize(c echo.Context) error {
	query := c.QueryParams()

	// Parse required parameters
	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	responseType := query.Get("response_type")
	scope := query.Get("scope")
	state := query.Get("state")
	nonce := query.Get("nonce")
	codeChallenge := query.Get("code_challenge")
	codeChallengeMethod := query.Get("code_challenge_method")

	// Validate parameters
	if clientID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "client_id is required",
		})
	}

	if redirectURI == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "redirect_uri is required",
		})
	}

	// Support both authorization code flow and implicit flow
	if responseType != ResponseTypeCode && responseType != ResponseTypeIDToken && responseType != ResponseTypeTokenIDToken {
		return redirectWithError(c, redirectURI, "unsupported_response_type", "Only 'code', 'id_token', and 'token id_token' response types are supported", state)
	}

	if !strings.Contains(scope, "openid") {
		return redirectWithError(c, redirectURI, "invalid_scope", "scope must contain 'openid'", state)
	}

	// Validate client
	client, err := h.storage.GetClientByID(clientID)
	if err != nil {
		return redirectWithError(c, redirectURI, "invalid_client", "Client not found", state)
	}

	// Validate redirect URI
	if !contains(client.RedirectURIs, redirectURI) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "Invalid redirect_uri",
		})
	}

	// For this example, we'll create a session and redirect to login
	// In production, check if user is already authenticated
	sessionID := uuid.New().String()

	// Store authorization request in session (simplified - should use proper session storage)
	// For now, redirect to login with parameters
	loginURL := fmt.Sprintf("/login?session_id=%s&client_id=%s&redirect_uri=%s&response_type=%s&scope=%s&state=%s&nonce=%s&code_challenge=%s&code_challenge_method=%s",
		sessionID, clientID, url.QueryEscape(redirectURI), url.QueryEscape(responseType), url.QueryEscape(scope), state, nonce, codeChallenge, codeChallengeMethod)

	return c.Redirect(http.StatusFound, loginURL)
}

// Login handles the login page (GET/POST /login)
func (h *Handlers) Login(c echo.Context) error {
	if c.Request().Method == "GET" {
		// Render login page (simplified HTML)
		return h.renderLoginPage(c)
	}

	// POST - handle login
	username := c.FormValue("username")
	sessionID := c.FormValue("session_id")
	clientID := c.FormValue("client_id")
	redirectURI := c.FormValue("redirect_uri")
	responseType := c.FormValue("response_type")
	scope := c.FormValue("scope")
	state := c.FormValue("state")
	nonce := c.FormValue("nonce")
	codeChallenge := c.FormValue("code_challenge")
	codeChallengeMethod := c.FormValue("code_challenge_method")

	// Authenticate user
	user, err := h.storage.GetUserByUsername(username)
	if err != nil {
		return h.renderLoginPageWithError(c, "Invalid username or password")
	}

	// Validate password
	password := c.FormValue("password")
	if !crypto.ValidatePassword(password, user.PasswordHash) {
		return h.renderLoginPageWithError(c, "Invalid username or password")
	}

	// For admin UI, verify user has admin role
	if clientID == "admin-ui" && user.Role != "admin" {
		return h.renderLoginPageWithError(c, "Access denied: Admin privileges required")
	}

	// Get client for token generation
	client, err := h.storage.GetClientByID(clientID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to get client",
		})
	}

	// Handle implicit flow (id_token or token id_token)
	if responseType == ResponseTypeIDToken || responseType == ResponseTypeTokenIDToken {
		// Generate ID token
		idToken, err := h.jwtManager.GenerateIDToken(user, clientID, nonce)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":             "server_error",
				"error_description": "Failed to generate ID token",
			})
		}

		// Build redirect URL with fragment
		fragment := fmt.Sprintf("id_token=%s&state=%s", idToken, state)

		// If response_type includes 'token', also generate access token
		if responseType == ResponseTypeTokenIDToken {
			accessToken, err := h.jwtManager.GenerateAccessToken(user, client.ID, scope)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error":             "server_error",
					"error_description": "Failed to generate access token",
				})
			}
			fragment += fmt.Sprintf("&access_token=%s&token_type=Bearer&expires_in=3600", accessToken)
		}

		redirectURL := fmt.Sprintf("%s#%s", redirectURI, fragment)
		return c.Redirect(http.StatusFound, redirectURL)
	}

	// Handle authorization code flow
	// Create authorization code
	authCode := models.NewAuthorizationCode(clientID, user.ID, redirectURI, scope)
	authCode.Nonce = nonce
	authCode.CodeChallenge = codeChallenge
	authCode.CodeChallengeMethod = codeChallengeMethod

	if err := h.storage.CreateAuthorizationCode(authCode); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to create authorization code",
		})
	}

	// Create session
	session := &models.Session{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: authCode.ExpiresAt,
		CreatedAt: authCode.CreatedAt,
	}
	_ = h.storage.CreateSession(session)

	// Redirect back to client with authorization code
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, authCode.Code, state)
	return c.Redirect(http.StatusFound, redirectURL)
}

// Consent handles the consent page (GET/POST /consent)
func (h *Handlers) Consent(c echo.Context) error {
	// Simplified consent - in production, show consent screen
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Consent endpoint - to be implemented",
	})
}

func (h *Handlers) renderLoginPage(c echo.Context) error {
	return h.renderLoginPageWithError(c, "")
}

func (h *Handlers) renderLoginPageWithError(c echo.Context, errorMsg string) error {
	query := c.QueryParams()
	errorHTML := ""
	if errorMsg != "" {
		errorHTML = fmt.Sprintf(`<div style="color: red; text-align: center; margin: 10px 0;">%s</div>`, errorMsg)
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Login - OpenID Connect</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 100px auto; padding: 20px; }
        input { width: 100%%; padding: 10px; margin: 10px 0; box-sizing: border-box; }
        button { width: 100%%; padding: 10px; background: #007bff; color: white; border: none; cursor: pointer; }
        button:hover { background: #0056b3; }
        h2 { text-align: center; }
    </style>
</head>
<body>
    <h2>Sign In</h2>
    %s
    <form method="POST" action="/login">
        <input type="hidden" name="session_id" value="%s">
        <input type="hidden" name="client_id" value="%s">
        <input type="hidden" name="redirect_uri" value="%s">
        <input type="hidden" name="response_type" value="%s">
        <input type="hidden" name="scope" value="%s">
        <input type="hidden" name="state" value="%s">
        <input type="hidden" name="nonce" value="%s">
        <input type="hidden" name="code_challenge" value="%s">
        <input type="hidden" name="code_challenge_method" value="%s">
        <input type="text" name="username" placeholder="Username" required autofocus>
        <input type="password" name="password" placeholder="Password" required>
        <button type="submit">Sign In</button>
    </form>
</body>
</html>
	`, errorHTML, query.Get("session_id"), query.Get("client_id"), query.Get("redirect_uri"),
		query.Get("response_type"), query.Get("scope"), query.Get("state"), query.Get("nonce"),
		query.Get("code_challenge"), query.Get("code_challenge_method"))

	return c.HTML(http.StatusOK, html)
}

func redirectWithError(c echo.Context, redirectURI, errorCode, description, state string) error {
	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("error", errorCode)
	q.Set("error_description", description)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	return c.Redirect(http.StatusFound, u.String())
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
