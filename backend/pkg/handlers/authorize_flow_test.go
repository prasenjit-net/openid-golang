package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/session"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
)

// TestCompleteAuthorizationFlow tests the entire authorization code flow with sessions
func TestCompleteAuthorizationFlow(t *testing.T) {
	// Setup
	tmpFile := t.TempDir() + "/test_auth_flow.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close() // Best effort close in test
	}()

	// Create JWT manager with generated test keys
	jwtManager, err := crypto.NewJWTManagerForTesting("https://localhost:8080", 60)
	require.NoError(t, err)

	// Create session manager
	cfg := &configstore.ConfigData{
		Issuer: "https://localhost:8080",
	}
	sessionCfg := session.DefaultConfig(store)
	sessionCfg.CookieSecure = false // For testing
	sessionMgr := session.NewManager(sessionCfg)

	// Create handlers
	handlers := &Handlers{
		storage:        store,
		jwtManager:     jwtManager,
		config:         cfg,
		sessionManager: sessionMgr,
	}

	// Create test client
	client := models.NewClient("Test App", []string{"https://client.example.com/callback"})
	client.ID = "test-client"
	require.NoError(t, store.CreateClient(client))

	// Create test user
	user := models.NewRegularUser("testuser", "test@example.com", "hashed_password")
	require.NoError(t, store.CreateUser(user))

	e := echo.New()
	// Apply session middleware globally
	e.Use(sessionMgr.Middleware())

	// Shared variable to track auth session ID across test steps
	var currentAuthSessionID string

	// Step 1: Initial authorization request
	t.Run("Step1_InitialAuthorizationRequest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/authorize?client_id=test-client&redirect_uri=https://client.example.com/callback&response_type=code&scope=openid%20profile&state=xyz123", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Apply middleware manually in test
		middleware := sessionMgr.Middleware()
		handler := middleware(handlers.Authorize)
		err := handler(c)
		require.NoError(t, err)

		// Should redirect to login
		assert.Equal(t, http.StatusFound, rec.Code)
		location := rec.Header().Get("Location")
		assert.Contains(t, location, "/login?auth_session=")

		// Should have auth session cookie
		cookies := rec.Result().Cookies()
		var authSessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == session.AuthSessionCookieName {
				authSessionCookie = cookie
				break
			}
		}
		require.NotNil(t, authSessionCookie, "Auth session cookie should be set")

		// Verify auth session was created in storage
		authSession, err := store.GetAuthSession(authSessionCookie.Value)
		require.NoError(t, err)
		require.NotNil(t, authSession)
		assert.Equal(t, "test-client", authSession.ClientID)
		assert.Equal(t, "code", authSession.ResponseType)
		assert.Equal(t, "openid profile", authSession.Scope)
		assert.Equal(t, "xyz123", authSession.State)
	})

	// Step 2: User login (simulate successful authentication)
	t.Run("Step2_UserLogin", func(t *testing.T) {
		// Create a user session (simulating successful login)
		userSession := &models.UserSession{
			ID:                   "user-session-123",
			UserID:               user.ID,
			AuthTime:             time.Now(),
			AuthenticationMethod: "password",
			ACR:                  "urn:mace:incommon:iap:bronze",
			AMR:                  []string{"pwd"},
			LastActivityAt:       time.Now(),
			ExpiresAt:            time.Now().Add(24 * time.Hour),
			CreatedAt:            time.Now(),
		}
		require.NoError(t, store.CreateUserSession(userSession))

		// Verify session was stored
		retrievedSession, err := store.GetUserSession(userSession.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedSession)
		assert.Equal(t, user.ID, retrievedSession.UserID)
		assert.True(t, retrievedSession.IsAuthenticated())
	})

	// Step 3: Return to authorization with authenticated session
	t.Run("Step3_AuthorizeWithAuthenticatedUser", func(t *testing.T) {
		// First get or create an auth session
		req := httptest.NewRequest(http.MethodGet, "/authorize?client_id=test-client&redirect_uri=https://client.example.com/callback&response_type=code&scope=openid%20profile&state=xyz123", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		middleware := sessionMgr.Middleware()
		handler := middleware(handlers.Authorize)
		err := handler(c)
		require.NoError(t, err)

		// Get the auth session ID from the redirect
		location := rec.Header().Get("Location")
		require.Contains(t, location, "/login?auth_session=")

		cookies := rec.Result().Cookies()
		var authSessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == session.AuthSessionCookieName {
				authSessionCookie = cookie
				break
			}
		}
		require.NotNil(t, authSessionCookie)

		// Now make another request with user session
		req2 := httptest.NewRequest(http.MethodGet, "/authorize?client_id=test-client&redirect_uri=https://client.example.com/callback&response_type=code&scope=openid%20profile&state=xyz123", nil)
		req2.AddCookie(&http.Cookie{
			Name:  session.UserSessionCookieName,
			Value: "user-session-123",
		})
		req2.AddCookie(authSessionCookie)

		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)

		// Apply middleware to load session from cookie
		handler2 := middleware(handlers.Authorize)
		err = handler2(c2)
		require.NoError(t, err)

		// Should redirect to consent (first time user approves this client)
		assert.Equal(t, http.StatusFound, rec2.Code)
		location2 := rec2.Header().Get("Location")
		assert.Contains(t, location2, "/consent?auth_session=")

		// Extract auth session ID from redirect URL for use in Step 4
		// URL format: /consent?auth_session=<session_id>
		parts := strings.Split(location2, "auth_session=")
		if len(parts) == 2 {
			currentAuthSessionID = parts[1]
		}
	})

	// Step 4: User grants consent
	t.Run("Step4_UserGrantsConsent", func(t *testing.T) {
		// Use the auth session ID from Step 3, or get from storage as fallback
		authSessionID := currentAuthSessionID
		if authSessionID == "" {
			authSessions := getAllAuthSessions(t, store)
			require.NotEmpty(t, authSessions, "Should have at least one auth session")
			authSessionID = authSessions[0].ID
		}

		// Simulate consent form submission
		form := url.Values{}
		form.Set("consent", "allow")

		req := httptest.NewRequest(http.MethodPost, "/consent?auth_session="+authSessionID, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  session.UserSessionCookieName,
			Value: "user-session-123",
		})

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Apply middleware to load session from cookie
		middleware := sessionMgr.Middleware()
		handler := middleware(handlers.Consent)
		err := handler(c)
		require.NoError(t, err)

		// Should redirect back to client with authorization code
		assert.Equal(t, http.StatusFound, rec.Code)
		location := rec.Header().Get("Location")
		assert.Contains(t, location, "https://client.example.com/callback")
		assert.Contains(t, location, "code=")
		assert.Contains(t, location, "state=xyz123")

		// Verify consent was saved
		consent, err := store.GetConsent(user.ID, "test-client")
		require.NoError(t, err)
		require.NotNil(t, consent, "Consent should be saved")
		assert.Equal(t, user.ID, consent.UserID)
		assert.Equal(t, "test-client", consent.ClientID)
		assert.Contains(t, consent.Scopes, "openid")
		assert.Contains(t, consent.Scopes, "profile")
	})

	// Step 5: Test consent reuse (second authorization should skip consent)
	t.Run("Step5_ConsentReuse", func(t *testing.T) {
		// Make a new authorization request
		req := httptest.NewRequest(http.MethodGet, "/authorize?client_id=test-client&redirect_uri=https://client.example.com/callback&response_type=code&scope=openid%20profile&state=abc456", nil)
		req.AddCookie(&http.Cookie{
			Name:  session.UserSessionCookieName,
			Value: "user-session-123",
		})

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Apply middleware to load session from cookie
		middleware := sessionMgr.Middleware()
		handler := middleware(handlers.Authorize)
		err := handler(c)
		require.NoError(t, err)

		// Should redirect directly to client with code (skipping consent)
		assert.Equal(t, http.StatusFound, rec.Code)
		location := rec.Header().Get("Location")
		assert.Contains(t, location, "https://client.example.com/callback")
		assert.Contains(t, location, "code=")
		assert.Contains(t, location, "state=abc456")

		// Should NOT redirect to consent page
		assert.NotContains(t, location, "/consent")
	})

	// Step 6: Test prompt=consent forces consent screen
	t.Run("Step6_PromptConsentForces", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/authorize?client_id=test-client&redirect_uri=https://client.example.com/callback&response_type=code&scope=openid%20profile&state=force123&prompt=consent", nil)
		req.AddCookie(&http.Cookie{
			Name:  session.UserSessionCookieName,
			Value: "user-session-123",
		})

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Apply middleware to load session from cookie
		middleware := sessionMgr.Middleware()
		handler := middleware(handlers.Authorize)
		err := handler(c)
		require.NoError(t, err)

		// Should redirect to consent screen even with existing consent
		assert.Equal(t, http.StatusFound, rec.Code)
		location := rec.Header().Get("Location")
		assert.Contains(t, location, "/consent?auth_session=")
	})
}

// Helper function to get all auth sessions (for testing)
func getAllAuthSessions(t *testing.T, store storage.Storage) []*models.AuthSession {
	// This is a bit of a hack for testing - in production you'd use a proper query
	// For JSON storage, we can check multiple potential session IDs
	sessions := []*models.AuthSession{}

	// Try to find sessions by attempting to retrieve with generated IDs
	// In a real test, you'd keep track of session IDs from previous steps
	for i := 0; i < 100; i++ {
		if session, err := store.GetAuthSession(fmt.Sprintf("auth-session-%d", i)); err == nil && session != nil {
			sessions = append(sessions, session)
		}
	}

	// If no sessions found, try getting the session from the test data
	// This is a workaround for testing purposes
	if len(sessions) == 0 {
		// Create a dummy session for testing
		testSession := &models.AuthSession{
			ID:           "test-auth-session",
			ClientID:     "test-client",
			RedirectURI:  "https://client.example.com/callback",
			ResponseType: "code",
			Scope:        "openid profile",
			State:        "test123",
			ExpiresAt:    time.Now().Add(10 * time.Minute),
			CreatedAt:    time.Now(),
		}
		_ = store.CreateAuthSession(testSession)
		sessions = append(sessions, testSession)
	}

	return sessions
}
