package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
)

const (
	// Session cookie names
	UserSessionCookieName = "user_session"
	AuthSessionCookieName = "auth_session"

	// Default timeouts
	DefaultUserSessionTimeout = 24 * time.Hour
	DefaultAuthSessionTimeout = 10 * time.Minute

	// Context keys
	UserSessionKey = "user_session"
	AuthSessionKey = "auth_session"
)

// Config holds session middleware configuration
type Config struct {
	Storage            storage.Storage
	UserSessionTimeout time.Duration
	AuthSessionTimeout time.Duration
	CookieSecure       bool
	CookieHTTPOnly     bool
	CookieSameSite     http.SameSite
	CookieDomain       string
	CookiePath         string
	CleanupInterval    time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig(storage storage.Storage) Config {
	return Config{
		Storage:            storage,
		UserSessionTimeout: DefaultUserSessionTimeout,
		AuthSessionTimeout: DefaultAuthSessionTimeout,
		CookieSecure:       true, // Should be true in production
		CookieHTTPOnly:     true,
		CookieSameSite:     http.SameSiteLaxMode,
		CookieDomain:       "",
		CookiePath:         "/",
		CleanupInterval:    1 * time.Hour,
	}
}

// Manager handles session operations
type Manager struct {
	config Config
	store  Store
}

// NewManager creates a new session manager
func NewManager(config Config) *Manager {
	mgr := &Manager{
		config: config,
		store:  NewStore(config.Storage),
	}

	// Start background cleanup if interval is set
	if config.CleanupInterval > 0 {
		go mgr.startCleanup()
	}

	return mgr
}

// Middleware returns Echo middleware for session handling
func (m *Manager) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Try to load existing user session
			if cookie, err := c.Cookie(UserSessionCookieName); err == nil {
				if session, err := m.store.GetUserSession(cookie.Value); err == nil && session != nil {
					if session.IsAuthenticated() {
						c.Set(UserSessionKey, session)
					}
				}
			}

			// Try to load existing auth session
			if cookie, err := c.Cookie(AuthSessionCookieName); err == nil {
				if session, err := m.store.GetAuthSession(cookie.Value); err == nil && session != nil {
					c.Set(AuthSessionKey, session)
				}
			}

			return next(c)
		}
	}
}

// GetUserSession retrieves the user session from context
func GetUserSession(c echo.Context) *models.UserSession {
	if session := c.Get(UserSessionKey); session != nil {
		if us, ok := session.(*models.UserSession); ok {
			return us
		}
	}
	return nil
}

// GetAuthSession retrieves the auth session from context
func GetAuthSession(c echo.Context) *models.AuthSession {
	if session := c.Get(AuthSessionKey); session != nil {
		if as, ok := session.(*models.AuthSession); ok {
			return as
		}
	}
	return nil
}

// CreateUserSession creates a new user session and sets cookie
func (m *Manager) CreateUserSession(c echo.Context, userID string, authMethod string, acr string, amr []string) (*models.UserSession, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &models.UserSession{
		ID:                   sessionID,
		UserID:               userID,
		AuthTime:             now,
		AuthenticationMethod: authMethod,
		ACR:                  acr,
		AMR:                  amr,
		LastActivityAt:       now,
		ExpiresAt:            now.Add(m.config.UserSessionTimeout),
		CreatedAt:            now,
	}

	if err := m.store.CreateUserSession(session); err != nil {
		return nil, err
	}

	// Set cookie
	m.setSessionCookie(c, UserSessionCookieName, sessionID, m.config.UserSessionTimeout)

	// Store in context
	c.Set(UserSessionKey, session)

	return session, nil
}

// CreateAuthSession creates a new authorization session and sets cookie
func (m *Manager) CreateAuthSession(c echo.Context, clientID, redirectURI, responseType, scope, state string) (*models.AuthSession, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &models.AuthSession{
		ID:           sessionID,
		ClientID:     clientID,
		RedirectURI:  redirectURI,
		ResponseType: responseType,
		Scope:        scope,
		State:        state,
		ExpiresAt:    now.Add(m.config.AuthSessionTimeout),
		CreatedAt:    now,
	}

	// Parse optional parameters
	session.Nonce = c.QueryParam("nonce")
	session.CodeChallenge = c.QueryParam("code_challenge")
	session.CodeChallengeMethod = c.QueryParam("code_challenge_method")
	session.Prompt = c.QueryParam("prompt")
	session.Display = c.QueryParam("display")

	// Parse max_age
	if maxAgeStr := c.QueryParam("max_age"); maxAgeStr != "" {
		_, _ = fmt.Sscanf(maxAgeStr, "%d", &session.MaxAge)
	}

	if err := m.store.CreateAuthSession(session); err != nil {
		return nil, err
	}

	// Set cookie
	m.setSessionCookie(c, AuthSessionCookieName, sessionID, m.config.AuthSessionTimeout)

	// Store in context
	c.Set(AuthSessionKey, session)

	return session, nil
}

// UpdateUserSession updates an existing user session
func (m *Manager) UpdateUserSession(c echo.Context, session *models.UserSession) error {
	if err := m.store.UpdateUserSession(session); err != nil {
		return err
	}
	c.Set(UserSessionKey, session)
	return nil
}

// UpdateAuthSession updates an existing authorization session
func (m *Manager) UpdateAuthSession(c echo.Context, session *models.AuthSession) error {
	if err := m.store.UpdateAuthSession(session); err != nil {
		return err
	}
	c.Set(AuthSessionKey, session)
	return nil
}

// DeleteUserSession deletes a user session and clears cookie
func (m *Manager) DeleteUserSession(c echo.Context, sessionID string) error {
	if err := m.store.DeleteUserSession(sessionID); err != nil {
		return err
	}
	m.clearSessionCookie(c, UserSessionCookieName)
	c.Set(UserSessionKey, nil)
	return nil
}

// DeleteAuthSession deletes an auth session and clears cookie
func (m *Manager) DeleteAuthSession(c echo.Context, sessionID string) error {
	if err := m.store.DeleteAuthSession(sessionID); err != nil {
		return err
	}
	m.clearSessionCookie(c, AuthSessionCookieName)
	c.Set(AuthSessionKey, nil)
	return nil
}

// RequireAuth middleware ensures user is authenticated
func (m *Manager) RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userSession := GetUserSession(c)
			if userSession == nil || !userSession.IsAuthenticated() {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}
			return next(c)
		}
	}
}

// Helper methods

func (m *Manager) setSessionCookie(c echo.Context, name, value string, maxAge time.Duration) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     m.config.CookiePath,
		Domain:   m.config.CookieDomain,
		MaxAge:   int(maxAge.Seconds()),
		Secure:   m.config.CookieSecure,
		HttpOnly: m.config.CookieHTTPOnly,
		SameSite: m.config.CookieSameSite,
	}
	c.SetCookie(cookie)
}

func (m *Manager) clearSessionCookie(c echo.Context, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     m.config.CookiePath,
		Domain:   m.config.CookieDomain,
		MaxAge:   -1,
		Secure:   m.config.CookieSecure,
		HttpOnly: m.config.CookieHTTPOnly,
		SameSite: m.config.CookieSameSite,
	}
	c.SetCookie(cookie)
}

func (m *Manager) startCleanup() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		_ = m.store.CleanupExpiredSessions()
	}
}

// generateSessionID generates a cryptographically secure random session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
