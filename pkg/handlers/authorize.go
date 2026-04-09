package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/session"
)

const (
	// ResponseTypeCode is the authorization code response type
	ResponseTypeCode = "code"
	// ResponseTypeIDToken is the ID token response type (implicit flow)
	ResponseTypeIDToken = "id_token"
	// ResponseTypeToken is the access token response type (implicit flow - not standard OIDC)
	ResponseTypeToken = "token"
	// ResponseTypeTokenIDToken is the access token + ID token response type
	ResponseTypeTokenIDToken = "token id_token"
	// Hybrid flow response types (for future implementation)
	// ResponseTypeCodeIDToken is the authorization code + ID token response type
	ResponseTypeCodeIDToken = "code id_token"
	// ResponseTypeCodeToken is the authorization code + access token response type
	ResponseTypeCodeToken = "code token"
	// ResponseTypeCodeTokenIDToken is the authorization code + access token + ID token response type
	ResponseTypeCodeTokenIDToken = "code id_token token"
)

// validateAuthorizationRequest validates required OAuth/OIDC parameters
func (h *Handlers) validateAuthorizationRequest(c echo.Context, clientID, redirectURI, responseType, scope, state string) (*models.Client, error) {
	// Validate parameters
	if clientID == "" {
		return nil, jsonError(c, http.StatusBadRequest, ErrorInvalidRequest, "client_id is required")
	}

	if redirectURI == "" {
		return nil, jsonError(c, http.StatusBadRequest, ErrorInvalidRequest, "redirect_uri is required")
	}

	// Support both authorization code flow and implicit flow
	if responseType != ResponseTypeCode && responseType != ResponseTypeIDToken && responseType != ResponseTypeTokenIDToken {
		return nil, authorizationError(c, redirectURI, responseType, ErrorUnsupportedResponseType, "Only 'code', 'id_token', and 'token id_token' response types are supported", state)
	}

	if !strings.Contains(scope, "openid") {
		return nil, authorizationError(c, redirectURI, responseType, ErrorInvalidScope, "scope must contain 'openid'", state)
	}

	// Nonce is REQUIRED for implicit flow (OIDC Core Section 3.2.2.1)
	if responseType == ResponseTypeIDToken || responseType == ResponseTypeTokenIDToken {
		nonce := c.QueryParam("nonce")
		if nonce == "" {
			return nil, authorizationError(c, redirectURI, responseType, ErrorInvalidRequest, "nonce parameter is required for implicit flow", state)
		}
	}

	// Validate client
	client, err := h.storage.GetClientByID(clientID)
	if err != nil {
		return nil, authorizationError(c, redirectURI, responseType, ErrorUnauthorizedClient, "Client not found", state)
	}

	// Validate redirect URI
	if !contains(client.RedirectURIs, redirectURI) {
		return nil, jsonError(c, http.StatusBadRequest, ErrorInvalidRequest, "Invalid redirect_uri")
	}

	return client, nil
}

// handlePromptParameter handles the prompt parameter logic
// Returns true and an error/response if prompt was handled and flow should stop
// Returns false and nil if normal flow should continue
func (h *Handlers) handlePromptParameter(c echo.Context, authSession *models.AuthSession, userSession *models.UserSession, redirectURI, state string) (bool, error) {
	prompt := c.QueryParam("prompt")
	if prompt == "" {
		return false, nil // No prompt parameter, continue normal flow
	}

	switch prompt {
	case "none":
		// Must not display any UI - check if consent already given
		if !authSession.ConsentGiven {
			return true, authorizationError(c, redirectURI, authSession.ResponseType, ErrorConsentRequired, "User consent required but prompt=none", state)
		}
		// Proceed to generate code/tokens
		return true, h.completeAuthorization(c, authSession, userSession)
	case "login":
		// Force re-authentication
		return true, c.Redirect(http.StatusFound, "/login?auth_session="+authSession.ID)
	case "consent":
		// Force consent screen
		return true, c.Redirect(http.StatusFound, "/consent?auth_session="+authSession.ID)
	case "select_account":
		// Show account selection (simplified: redirect to login)
		return true, c.Redirect(http.StatusFound, "/login?auth_session="+authSession.ID)
	}

	return false, nil // Unknown prompt value, ignore
}

// checkAndApplyConsent checks for existing consent and applies it if valid
func (h *Handlers) checkAndApplyConsent(authSession *models.AuthSession, userSession *models.UserSession, clientID, scope string) error {
	existingConsent, err := h.storage.GetConsent(userSession.UserID, clientID)
	if err == nil && existingConsent != nil {
		// Check if existing consent covers all requested scopes
		requestedScopes := strings.Split(scope, " ")
		if existingConsent.HasAllScopes(requestedScopes) {
			authSession.ConsentGiven = true
			authSession.ConsentedScopes = existingConsent.Scopes
		}
	}
	return nil
}

// handleAuthenticatedUser processes authorization when user is already authenticated
func (h *Handlers) handleAuthenticatedUser(c echo.Context, authSession *models.AuthSession, userSession *models.UserSession, clientID, scope, redirectURI, state string) error {
	// Handle prompt parameter - if it was handled, return immediately
	handled, err := h.handlePromptParameter(c, authSession, userSession, redirectURI, state)
	if handled {
		return err
	}

	// Check max_age parameter
	if authSession.MaxAge > 0 {
		if !userSession.IsAuthTimeFresh(authSession.MaxAge) {
			// Re-authentication required
			return c.Redirect(http.StatusFound, "/login?auth_session="+authSession.ID)
		}
	}

	// Note: Don't check consent if prompt=consent was handled above
	// Check if user has previously consented to this client
	_ = h.checkAndApplyConsent(authSession, userSession, clientID, scope)

	if !authSession.ConsentGiven {
		// Redirect to consent screen
		return c.Redirect(http.StatusFound, "/consent?auth_session="+authSession.ID)
	}

	// All checks passed, complete authorization
	return h.completeAuthorization(c, authSession, userSession)
}

// Authorize handles the authorization endpoint (GET /authorize)
func (h *Handlers) Authorize(c echo.Context) error {
	query := c.QueryParams()

	// Parse required parameters
	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	responseType := query.Get("response_type")
	scope := query.Get("scope")
	state := query.Get("state")

	// Validate parameters
	client, err := h.validateAuthorizationRequest(c, clientID, redirectURI, responseType, scope, state)
	if err != nil {
		return err
	}
	_ = client // Client validated but not used directly in this function

	// Create authorization session to store request parameters
	authSession, err := h.sessionManager.CreateAuthSession(c, clientID, redirectURI, responseType, scope, state)
	if err != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to create authorization session")
	}

	// Check if user is already authenticated
	userSession := session.GetUserSession(c)
	if userSession != nil && userSession.IsAuthenticated() {
		return h.handleAuthenticatedUser(c, authSession, userSession, clientID, scope, redirectURI, state)
	}

	// User not authenticated, redirect to login
	return c.Redirect(http.StatusFound, "/login?auth_session="+authSession.ID)
}

// Login handles the login page (GET/POST /login)
func (h *Handlers) Login(c echo.Context) error {
	authSessionID := c.QueryParam("auth_session")

	if c.Request().Method == "GET" {
		// Render login page
		return h.renderLoginPage(c, authSessionID)
	}

	// POST - handle login
	username := c.FormValue("username")
	password := c.FormValue("password")

	// Authenticate user
	user, err := h.storage.GetUserByUsername(username)
	if err != nil || user == nil {
		return h.renderLoginPageWithError(c, authSessionID, "Invalid username or password")
	}

	// Validate password
	if !crypto.ValidatePassword(password, user.PasswordHash) {
		return h.renderLoginPageWithError(c, authSessionID, "Invalid username or password")
	}

	// Get authorization session if exists
	var authSession *models.AuthSession
	if authSessionID != "" {
		var authErr error
		authSession, authErr = h.storage.GetAuthSession(authSessionID)
		if authErr != nil || authSession == nil {
			return jsonError(c, http.StatusBadRequest, ErrorInvalidRequest, "Invalid or expired authorization session")
		}

		// For admin UI, verify user has admin role
		if authSession.ClientID == "admin-ui" && user.Role != "admin" {
			return h.renderLoginPageWithError(c, authSessionID, "Access denied: Admin privileges required")
		}
	}

	// Create user session with authentication details
	authMethod := "password"
	acr := "urn:mace:incommon:iap:silver" // Authentication Context Class Reference
	amr := []string{"pwd"}                // Authentication Methods References

	userSession, sessionErr := h.sessionManager.CreateUserSession(c, user.ID, authMethod, acr, amr)
	if sessionErr != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to create user session")
	}

	// Update auth session with user info
	if authSession != nil {
		authSession.UserID = user.ID
		authTime := userSession.AuthTime
		authSession.AuthTime = &authTime
		authSession.ACR = acr
		authSession.AMR = amr
		authSession.AuthenticationMethod = authMethod

		if updateErr := h.storage.UpdateAuthSession(authSession); updateErr != nil {
			return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to update authorization session")
		}

		// Redirect to consent screen
		return c.Redirect(http.StatusFound, "/consent?auth_session="+authSession.ID)
	}

	// No auth session, just logged in (e.g., admin UI direct access)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Login successful",
		"user_id": user.ID,
	})
}

// Consent handles the consent page (GET/POST /consent)
func (h *Handlers) Consent(c echo.Context) error {
	authSessionID := c.QueryParam("auth_session")

	if authSessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "auth_session is required",
		})
	}

	// Get authorization session
	authSession, authErr := h.storage.GetAuthSession(authSessionID)
	if authErr != nil || authSession == nil {
		return jsonError(c, http.StatusBadRequest, ErrorInvalidRequest, "Invalid or expired authorization session")
	}

	// Get user session
	userSession := session.GetUserSession(c)
	if userSession == nil || !userSession.IsAuthenticated() {
		return c.Redirect(http.StatusFound, "/login?auth_session="+authSessionID)
	}

	// Get client info
	client, clientErr := h.storage.GetClientByID(authSession.ClientID)
	if clientErr != nil || client == nil {
		return jsonError(c, http.StatusBadRequest, ErrorInvalidRequest, "Invalid client")
	}

	if c.Request().Method == "GET" {
		// Render consent page
		return h.renderConsentPage(c, authSession, client)
	}

	// POST - handle consent decision
	consentDecision := c.FormValue("consent")
	if consentDecision != "allow" {
		// User denied consent
		return authorizationError(c, authSession.RedirectURI, authSession.ResponseType, ErrorAccessDenied, "User denied consent", authSession.State)
	}

	// Update auth session with consent
	authSession.ConsentGiven = true
	authSession.ConsentedScopes = strings.Split(authSession.Scope, " ")

	if err := h.storage.UpdateAuthSession(authSession); err != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to update authorization session")
	}

	// Save consent for future authorization requests
	// Check if consent already exists
	existingConsent, err := h.storage.GetConsent(userSession.UserID, authSession.ClientID)
	if err == nil && existingConsent != nil {
		// Update existing consent with new scopes
		existingConsent.Scopes = authSession.ConsentedScopes
		_ = h.storage.UpdateConsent(existingConsent)
	} else {
		// Create new consent record
		newConsent := models.NewConsent(userSession.UserID, authSession.ClientID, authSession.ConsentedScopes)
		_ = h.storage.CreateConsent(newConsent)
	}

	// Complete authorization
	return h.completeAuthorization(c, authSession, userSession)
}

// completeAuthorization completes the authorization flow
func (h *Handlers) completeAuthorization(c echo.Context, authSession *models.AuthSession, userSession *models.UserSession) error {
	// Get user
	user, err := h.storage.GetUserByID(userSession.UserID)
	if err != nil || user == nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to get user")
	}

	// Get client
	client, err := h.storage.GetClientByID(authSession.ClientID)
	if err != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to get client")
	}

	// Handle implicit flow (id_token or token id_token)
	if authSession.ResponseType == ResponseTypeIDToken || authSession.ResponseType == ResponseTypeTokenIDToken {
		var accessToken string
		var err error

		// If response_type includes 'token', generate access token first
		if authSession.ResponseType == ResponseTypeTokenIDToken {
			accessToken, err = h.jwtManager.GenerateAccessToken(user, client.ID, authSession.Scope)
			if err != nil {
				return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to generate access token")
			}
		}

		// Generate ID token with auth_time, acr, amr, and at_hash (if access token present)
		idToken, err := h.jwtManager.GenerateIDTokenWithClaims(
			user,
			authSession.ClientID,
			authSession.Nonce,
			authSession.Scope, // Pass scope for claim filtering
			userSession.AuthTime,
			userSession.ACR,
			userSession.AMR,
			accessToken, // at_hash will be included if this is not empty
			"",          // c_hash not used in implicit flow
		)
		if err != nil {
			return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to generate ID token")
		}

		// Build redirect URL with fragment
		fragment := fmt.Sprintf("id_token=%s&state=%s", idToken, authSession.State)

		// Add access token to fragment if present
		if accessToken != "" {
			fragment += fmt.Sprintf("&access_token=%s&token_type=Bearer&expires_in=3600", accessToken)
		}

		// Clean up auth session
		_ = h.storage.DeleteAuthSession(authSession.ID)

		redirectURL := fmt.Sprintf("%s#%s", authSession.RedirectURI, fragment)
		return c.Redirect(http.StatusFound, redirectURL)
	}

	// Handle authorization code flow
	// Create authorization code
	authCode := models.NewAuthorizationCode(authSession.ClientID, user.ID, authSession.RedirectURI, authSession.Scope)
	authCode.Nonce = authSession.Nonce
	authCode.CodeChallenge = authSession.CodeChallenge
	authCode.CodeChallengeMethod = authSession.CodeChallengeMethod

	if err := h.storage.CreateAuthorizationCode(authCode); err != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to create authorization code")
	}

	// Clean up auth session
	_ = h.storage.DeleteAuthSession(authSession.ID)

	// Redirect back to client with authorization code
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", authSession.RedirectURI, authCode.Code, authSession.State)
	return c.Redirect(http.StatusFound, redirectURL)
}

func (h *Handlers) renderLoginPage(c echo.Context, authSessionID string) error {
	return h.renderLoginPageWithError(c, authSessionID, "")
}

func (h *Handlers) renderLoginPageWithError(c echo.Context, authSessionID, errorMsg string) error {
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
    <form method="POST" action="/login?auth_session=%s">
        <input type="text" name="username" placeholder="Username" required autofocus>
        <input type="password" name="password" placeholder="Password" required>
        <button type="submit">Sign In</button>
    </form>
</body>
</html>
	`, errorHTML, authSessionID)

	return c.HTML(http.StatusOK, html)
}

func (h *Handlers) renderConsentPage(c echo.Context, authSession *models.AuthSession, client *models.Client) error {
	scopes := strings.Split(authSession.Scope, " ")
	scopeList := ""
	for _, scope := range scopes {
		scopeList += fmt.Sprintf("<li>%s</li>", scope)
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Grant Permission - OpenID Connect</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 500px; margin: 100px auto; padding: 20px; }
        .client { font-weight: bold; color: #007bff; }
        .scopes { background: #f5f5f5; padding: 15px; margin: 20px 0; border-radius: 5px; }
        ul { list-style: none; padding: 0; }
        li { padding: 5px 0; }
        .buttons { display: flex; gap: 10px; }
        button { flex: 1; padding: 12px; border: none; cursor: pointer; font-size: 16px; }
        .allow { background: #28a745; color: white; }
        .allow:hover { background: #218838; }
        .deny { background: #dc3545; color: white; }
        .deny:hover { background: #c82333; }
        h2 { text-align: center; }
    </style>
</head>
<body>
    <h2>Grant Permission</h2>
    <p>The application <span class="client">%s</span> is requesting access to your account.</p>
    
    <div class="scopes">
        <strong>Requested permissions:</strong>
        <ul>%s</ul>
    </div>

    <form method="POST" action="/consent?auth_session=%s">
        <div class="buttons">
            <button type="submit" name="consent" value="deny" class="deny">Deny</button>
            <button type="submit" name="consent" value="allow" class="allow">Allow</button>
        </div>
    </form>
</body>
</html>
	`, client.Name, scopeList, authSession.ID)

	return c.HTML(http.StatusOK, html)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
