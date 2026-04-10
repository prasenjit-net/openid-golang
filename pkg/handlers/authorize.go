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
	// ResponseTypeTokenIDToken is the access token + ID token response type (implicit flow)
	// OIDC Core 1.0 §3.2.2.1 specifies this order: id_token first, then token
	ResponseTypeTokenIDToken = "id_token token"
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
	if err != nil || client == nil {
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
		h.logAudit(models.AuditActionLoginFailed, models.AuditActorUser, username,
			"user", "", models.AuditStatusFailure,
			c.RealIP(), c.Request().UserAgent(),
			map[string]interface{}{"reason": "user not found"})
		return h.renderLoginPageWithError(c, authSessionID, "Invalid username or password")
	}

	// Validate password
	if !crypto.ValidatePassword(password, user.PasswordHash) {
		h.logAudit(models.AuditActionLoginFailed, models.AuditActorUser, username,
			"user", user.ID, models.AuditStatusFailure,
			c.RealIP(), c.Request().UserAgent(),
			map[string]interface{}{"reason": "invalid password"})
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
			h.logAudit(models.AuditActionLoginFailed, models.AuditActorUser, username,
				"user", user.ID, models.AuditStatusFailure,
				c.RealIP(), c.Request().UserAgent(),
				map[string]interface{}{"reason": "not admin", "client_id": authSession.ClientID})
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

	// Audit successful login
	h.logAudit(models.AuditActionLogin, models.AuditActorUser, username,
		"user", user.ID, models.AuditStatusSuccess,
		c.RealIP(), c.Request().UserAgent(), nil)

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
		h.logAudit(models.AuditActionConsentDeny, models.AuditActorUser, h.userIDToUsername(userSession.UserID),
			"client", authSession.ClientID, models.AuditStatusSuccess,
			c.RealIP(), c.Request().UserAgent(),
			map[string]interface{}{"scope": authSession.Scope})
		return authorizationError(c, authSession.RedirectURI, authSession.ResponseType, ErrorAccessDenied, "User denied consent", authSession.State)
	}

	// Update auth session with consent
	authSession.ConsentGiven = true
	authSession.ConsentedScopes = strings.Split(authSession.Scope, " ")

	if err := h.storage.UpdateAuthSession(authSession); err != nil {
		return jsonError(c, http.StatusInternalServerError, ErrorServerError, "Failed to update authorization session")
	}

	// Audit consent granted
	h.logAudit(models.AuditActionConsentGrant, models.AuditActorUser, h.userIDToUsername(userSession.UserID),
		"client", authSession.ClientID, models.AuditStatusSuccess,
		c.RealIP(), c.Request().UserAgent(),
		map[string]interface{}{"scope": authSession.Scope})

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
		errorHTML = fmt.Sprintf(`
		<div class="error-banner">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/>
				<line x1="12" y1="16" x2="12.01" y2="16"/>
			</svg>
			<span>%s</span>
		</div>`, errorMsg)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sign In — OpenID Connect</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

        body {
            font-family: 'Inter', system-ui, sans-serif;
            min-height: 100vh;
            background: #0B1120;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 24px;
            position: relative;
            overflow: hidden;
        }

        body::before {
            content: '';
            position: absolute;
            inset: 0;
            background:
                radial-gradient(ellipse 80%% 60%% at 20%% 20%%, rgba(13,148,136,0.18) 0%%, transparent 60%%),
                radial-gradient(ellipse 60%% 80%% at 80%% 80%%, rgba(245,158,11,0.10) 0%%, transparent 60%%);
            pointer-events: none;
        }

        .card {
            position: relative;
            background: #1E293B;
            border: 1px solid rgba(255,255,255,0.08);
            border-radius: 16px;
            padding: 40px 36px;
            width: 100%%;
            max-width: 400px;
            box-shadow: 0 25px 60px rgba(0,0,0,0.5), 0 0 0 1px rgba(13,148,136,0.12);
        }

        .logo {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 10px;
            margin-bottom: 28px;
        }

        .logo-icon {
            width: 40px;
            height: 40px;
            background: linear-gradient(135deg, #0D9488 0%%, #0F766E 100%%);
            border-radius: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            box-shadow: 0 4px 12px rgba(13,148,136,0.35);
        }

        .logo-text {
            font-size: 18px;
            font-weight: 700;
            color: #F1F5F9;
            letter-spacing: -0.3px;
        }

        .logo-text span {
            color: #0D9488;
        }

        h2 {
            font-size: 22px;
            font-weight: 700;
            color: #F1F5F9;
            text-align: center;
            letter-spacing: -0.3px;
            margin-bottom: 6px;
        }

        .subtitle {
            font-size: 13px;
            color: #94A3B8;
            text-align: center;
            margin-bottom: 28px;
        }

        .error-banner {
            display: flex;
            align-items: center;
            gap: 8px;
            background: rgba(239,68,68,0.12);
            border: 1px solid rgba(239,68,68,0.3);
            color: #FCA5A5;
            border-radius: 8px;
            padding: 10px 14px;
            font-size: 13px;
            margin-bottom: 20px;
        }

        .field {
            margin-bottom: 16px;
        }

        label {
            display: block;
            font-size: 12px;
            font-weight: 600;
            color: #94A3B8;
            text-transform: uppercase;
            letter-spacing: 0.06em;
            margin-bottom: 6px;
        }

        input {
            width: 100%%;
            padding: 11px 14px;
            background: #0F172A;
            border: 1px solid rgba(255,255,255,0.1);
            border-radius: 8px;
            color: #F1F5F9;
            font-family: 'Inter', sans-serif;
            font-size: 14px;
            outline: none;
            transition: border-color 0.15s, box-shadow 0.15s;
        }

        input::placeholder { color: #475569; }

        input:focus {
            border-color: #0D9488;
            box-shadow: 0 0 0 3px rgba(13,148,136,0.2);
        }

        button[type="submit"] {
            width: 100%%;
            padding: 12px;
            margin-top: 8px;
            background: linear-gradient(135deg, #0D9488, #0F766E);
            color: #fff;
            border: none;
            border-radius: 8px;
            font-family: 'Inter', sans-serif;
            font-size: 15px;
            font-weight: 600;
            cursor: pointer;
            letter-spacing: 0.01em;
            transition: opacity 0.15s, transform 0.1s, box-shadow 0.15s;
            box-shadow: 0 4px 14px rgba(13,148,136,0.35);
        }

        button[type="submit"]:hover {
            opacity: 0.92;
            transform: translateY(-1px);
            box-shadow: 0 6px 20px rgba(13,148,136,0.45);
        }

        button[type="submit"]:active { transform: translateY(0); }

        .footer {
            text-align: center;
            margin-top: 24px;
            font-size: 12px;
            color: #475569;
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="logo">
            <div class="logo-icon">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
                    <path d="M12 2L4 6v6c0 5.25 3.5 10.15 8 11.35C16.5 22.15 20 17.25 20 12V6L12 2z" fill="rgba(255,255,255,0.9)"/>
                    <circle cx="12" cy="11" r="2" fill="#0D9488"/>
                    <path d="M12 13v3" stroke="#0D9488" stroke-width="2" stroke-linecap="round"/>
                </svg>
            </div>
            <span class="logo-text">Secure<span>ID</span></span>
        </div>

        <h2>Welcome back</h2>
        <p class="subtitle">Sign in to your account to continue</p>

        %s

        <form method="POST" action="/login?auth_session=%s">
            <div class="field">
                <label for="username">Username</label>
                <input type="text" id="username" name="username" placeholder="Enter your username" required autofocus autocomplete="username">
            </div>
            <div class="field">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" placeholder="Enter your password" required autocomplete="current-password">
            </div>
            <button type="submit">
                Sign In
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="vertical-align:middle;margin-left:6px;">
                    <path d="M5 12h14M12 5l7 7-7 7"/>
                </svg>
            </button>
        </form>

        <p class="footer">Protected by OpenID Connect</p>
    </div>
</body>
</html>`, errorHTML, authSessionID)

	return c.HTML(http.StatusOK, html)
}

// scopeInfo maps a scope name to a human-readable description and an SVG icon path.
var scopeInfo = map[string][2]string{
	"openid":  {"Verify your identity", `<path d="M12 2a5 5 0 1 1 0 10A5 5 0 0 1 12 2zm0 12c-5.33 0-8 2.67-8 4v2h16v-2c0-1.33-2.67-4-8-4z"/>`},
	"profile": {"Access your name and profile info", `<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>`},
	"email":   {"Read your email address", `<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/>`},
}

func (h *Handlers) renderConsentPage(c echo.Context, authSession *models.AuthSession, client *models.Client) error {
	scopes := strings.Split(authSession.Scope, " ")
	scopeItems := ""
	for _, scope := range scopes {
		info, ok := scopeInfo[scope]
		label := scope
		iconPath := `<circle cx="12" cy="12" r="9"/><path d="M12 8v4l3 3"/>`
		if ok {
			label = info[1]
			iconPath = info[0]
		}
		scopeItems += fmt.Sprintf(`
		<li class="scope-item">
			<div class="scope-icon">
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">%s</svg>
			</div>
			<div class="scope-text">
				<span class="scope-name">%s</span>
				<span class="scope-desc">%s</span>
			</div>
		</li>`, iconPath, scope, label)
	}

	// Generate client initials for avatar
	clientName := client.Name
	if clientName == "" {
		clientName = "App"
	}
	initials := string([]rune(clientName)[0:1])
	if len([]rune(clientName)) > 1 {
		words := strings.Fields(clientName)
		if len(words) >= 2 {
			initials = string([]rune(words[0])[0:1]) + string([]rune(words[1])[0:1])
		}
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Authorize Access — OpenID Connect</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

        body {
            font-family: 'Inter', system-ui, sans-serif;
            min-height: 100vh;
            background: #0B1120;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 24px;
            position: relative;
            overflow: hidden;
        }

        body::before {
            content: '';
            position: absolute;
            inset: 0;
            background:
                radial-gradient(ellipse 80%% 60%% at 20%% 20%%, rgba(13,148,136,0.18) 0%%, transparent 60%%),
                radial-gradient(ellipse 60%% 80%% at 80%% 80%%, rgba(245,158,11,0.10) 0%%, transparent 60%%);
            pointer-events: none;
        }

        .card {
            position: relative;
            background: #1E293B;
            border: 1px solid rgba(255,255,255,0.08);
            border-radius: 16px;
            padding: 36px 32px;
            width: 100%%;
            max-width: 440px;
            box-shadow: 0 25px 60px rgba(0,0,0,0.5), 0 0 0 1px rgba(13,148,136,0.12);
        }

        .logo-bar {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 10px;
            margin-bottom: 24px;
        }

        .logo-icon {
            width: 32px;
            height: 32px;
            background: linear-gradient(135deg, #0D9488 0%%, #0F766E 100%%);
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            box-shadow: 0 4px 12px rgba(13,148,136,0.35);
            flex-shrink: 0;
        }

        .logo-text {
            font-size: 16px;
            font-weight: 700;
            color: #F1F5F9;
        }

        .logo-text span { color: #0D9488; }

        .app-header {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 12px;
            margin-bottom: 24px;
        }

        .app-avatar {
            width: 56px;
            height: 56px;
            background: linear-gradient(135deg, #0D9488 0%%, #0F766E 100%%);
            border-radius: 14px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 22px;
            font-weight: 700;
            color: #fff;
            box-shadow: 0 4px 16px rgba(13,148,136,0.35);
            text-transform: uppercase;
        }

        .app-name {
            font-size: 18px;
            font-weight: 700;
            color: #F1F5F9;
            letter-spacing: -0.3px;
        }

        .app-sub {
            font-size: 13px;
            color: #94A3B8;
            margin-top: 2px;
            text-align: center;
        }

        .divider {
            height: 1px;
            background: rgba(255,255,255,0.07);
            margin: 20px 0;
        }

        .permissions-label {
            font-size: 11px;
            font-weight: 600;
            color: #64748B;
            text-transform: uppercase;
            letter-spacing: 0.07em;
            margin-bottom: 12px;
        }

        .scope-list {
            list-style: none;
            display: flex;
            flex-direction: column;
            gap: 8px;
            margin-bottom: 24px;
        }

        .scope-item {
            display: flex;
            align-items: center;
            gap: 12px;
            background: rgba(255,255,255,0.04);
            border: 1px solid rgba(255,255,255,0.07);
            border-radius: 10px;
            padding: 12px 14px;
        }

        .scope-icon {
            width: 32px;
            height: 32px;
            background: rgba(13,148,136,0.15);
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #0D9488;
            flex-shrink: 0;
        }

        .scope-text {
            display: flex;
            flex-direction: column;
            gap: 2px;
        }

        .scope-name {
            font-size: 12px;
            font-weight: 600;
            color: #94A3B8;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .scope-desc {
            font-size: 13px;
            color: #CBD5E1;
        }

        .buttons {
            display: flex;
            gap: 10px;
        }

        button {
            flex: 1;
            padding: 12px 16px;
            border: none;
            border-radius: 8px;
            font-family: 'Inter', sans-serif;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            transition: opacity 0.15s, transform 0.1s, box-shadow 0.15s;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 6px;
        }

        button:hover { opacity: 0.88; transform: translateY(-1px); }
        button:active { transform: translateY(0); }

        .btn-deny {
            background: rgba(239,68,68,0.1);
            border: 1px solid rgba(239,68,68,0.25);
            color: #FCA5A5;
        }

        .btn-deny:hover {
            background: rgba(239,68,68,0.18);
            box-shadow: 0 4px 12px rgba(239,68,68,0.2);
        }

        .btn-allow {
            background: linear-gradient(135deg, #0D9488, #0F766E);
            color: #fff;
            box-shadow: 0 4px 14px rgba(13,148,136,0.35);
        }

        .btn-allow:hover {
            box-shadow: 0 6px 20px rgba(13,148,136,0.45);
        }

        .footer {
            text-align: center;
            margin-top: 20px;
            font-size: 11px;
            color: #475569;
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="logo-bar">
            <div class="logo-icon">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                    <path d="M12 2L4 6v6c0 5.25 3.5 10.15 8 11.35C16.5 22.15 20 17.25 20 12V6L12 2z" fill="rgba(255,255,255,0.9)"/>
                    <circle cx="12" cy="11" r="2" fill="#0D9488"/>
                    <path d="M12 13v3" stroke="#0D9488" stroke-width="2" stroke-linecap="round"/>
                </svg>
            </div>
            <span class="logo-text">Secure<span>ID</span></span>
        </div>

        <div class="app-header">
            <div class="app-avatar">%s</div>
            <div>
                <div class="app-name">%s</div>
                <div class="app-sub">is requesting access to your account</div>
            </div>
        </div>

        <div class="divider"></div>

        <p class="permissions-label">Requested permissions</p>
        <ul class="scope-list">%s</ul>

        <form method="POST" action="/consent?auth_session=%s">
            <div class="buttons">
                <button type="submit" name="consent" value="deny" class="btn-deny">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                    Deny
                </button>
                <button type="submit" name="consent" value="allow" class="btn-allow">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
                    Allow Access
                </button>
            </div>
        </form>

        <p class="footer">Your data is protected · Powered by OpenID Connect</p>
    </div>
</body>
</html>`, initials, clientName, scopeItems, authSession.ID)

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
