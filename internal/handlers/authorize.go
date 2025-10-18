package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/prasenjit-net/openid-golang/internal/crypto"
	"github.com/prasenjit-net/openid-golang/internal/models"
)

// Authorize handles the authorization endpoint (GET /authorize)
func (h *Handlers) Authorize(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

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
		writeError(w, http.StatusBadRequest, "invalid_request", "client_id is required")
		return
	}

	if redirectURI == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "redirect_uri is required")
		return
	}

	// Support both authorization code flow and implicit flow
	if responseType != "code" && responseType != "id_token" && responseType != "token id_token" {
		redirectWithError(w, redirectURI, "unsupported_response_type", "Only 'code', 'id_token', and 'token id_token' response types are supported", state)
		return
	}

	if !strings.Contains(scope, "openid") {
		redirectWithError(w, redirectURI, "invalid_scope", "scope must contain 'openid'", state)
		return
	}

	// Validate client
	client, err := h.storage.GetClientByID(clientID)
	if err != nil {
		redirectWithError(w, redirectURI, "invalid_client", "Client not found", state)
		return
	}

	// Validate redirect URI
	if !contains(client.RedirectURIs, redirectURI) {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid redirect_uri")
		return
	}

	// For this example, we'll create a session and redirect to login
	// In production, check if user is already authenticated
	sessionID := uuid.New().String()

	// Store authorization request in session (simplified - should use proper session storage)
	// For now, redirect to login with parameters
	loginURL := fmt.Sprintf("/login?session_id=%s&client_id=%s&redirect_uri=%s&response_type=%s&scope=%s&state=%s&nonce=%s&code_challenge=%s&code_challenge_method=%s",
		sessionID, clientID, url.QueryEscape(redirectURI), url.QueryEscape(responseType), url.QueryEscape(scope), state, nonce, codeChallenge, codeChallengeMethod)

	http.Redirect(w, r, loginURL, http.StatusFound)
}

// Login handles the login page (GET/POST /login)
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Render login page (simplified HTML)
		h.renderLoginPage(w, r)
		return
	}

	// POST - handle login
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Failed to parse form")
		return
	}

	username := r.FormValue("username")
	sessionID := r.FormValue("session_id")
	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")
	responseType := r.FormValue("response_type")
	scope := r.FormValue("scope")
	state := r.FormValue("state")
	nonce := r.FormValue("nonce")
	codeChallenge := r.FormValue("code_challenge")
	codeChallengeMethod := r.FormValue("code_challenge_method")

	// Authenticate user
	user, err := h.storage.GetUserByUsername(username)
	if err != nil {
		h.renderLoginPageWithError(w, r, "Invalid username or password")
		return
	}

	// Validate password
	password := r.FormValue("password")
	if !crypto.ValidatePassword(password, user.PasswordHash) {
		h.renderLoginPageWithError(w, r, "Invalid username or password")
		return
	}

	// For admin UI, verify user has admin role
	if clientID == "admin-ui" && user.Role != "admin" {
		h.renderLoginPageWithError(w, r, "Access denied: Admin privileges required")
		return
	}

	// Get client for token generation
	client, err := h.storage.GetClientByID(clientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get client")
		return
	}

	// Handle implicit flow (id_token or token id_token)
	if responseType == "id_token" || responseType == "token id_token" {
		// Generate ID token
		idToken, err := h.jwtManager.GenerateIDToken(user, clientID, nonce)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", "Failed to generate ID token")
			return
		}

		// Build redirect URL with fragment
		fragment := fmt.Sprintf("id_token=%s&state=%s", idToken, state)

		// If response_type includes 'token', also generate access token
		if responseType == "token id_token" {
			accessToken, err := h.jwtManager.GenerateAccessToken(user, client.ID, scope)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server_error", "Failed to generate access token")
				return
			}
			fragment += fmt.Sprintf("&access_token=%s&token_type=Bearer&expires_in=3600", accessToken)
		}

		redirectURL := fmt.Sprintf("%s#%s", redirectURI, fragment)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	// Handle authorization code flow
	// Create authorization code
	authCode := models.NewAuthorizationCode(clientID, user.ID, redirectURI, scope)
	authCode.Nonce = nonce
	authCode.CodeChallenge = codeChallenge
	authCode.CodeChallengeMethod = codeChallengeMethod

	if err := h.storage.CreateAuthorizationCode(authCode); err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create authorization code")
		return
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
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// Consent handles the consent page (GET/POST /consent)
func (h *Handlers) Consent(w http.ResponseWriter, r *http.Request) {
	// Simplified consent - in production, show consent screen
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Consent endpoint - to be implemented",
	})
}

func (h *Handlers) renderLoginPage(w http.ResponseWriter, r *http.Request) {
	h.renderLoginPageWithError(w, r, "")
}

func (h *Handlers) renderLoginPageWithError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	query := r.URL.Query()
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

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

func redirectWithError(w http.ResponseWriter, redirectURI, errorCode, description, state string) {
	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("error", errorCode)
	q.Set("error_description", description)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	http.Redirect(w, nil, u.String(), http.StatusFound)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
