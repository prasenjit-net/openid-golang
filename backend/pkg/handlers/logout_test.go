package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/session"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLogoutTest(t *testing.T) (*Handlers, storage.Storage, *models.UserSession, *models.Client) {
	// Create temporary file for test storage
	tmpFile := filepath.Join(t.TempDir(), "test-storage.json")

	// Create storage
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)

	// Clean up after test
	t.Cleanup(func() {
		_ = os.Remove(tmpFile)
	})

	// Create config
	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		JWT: configstore.JWTConfig{
			ExpiryMinutes: 60,
		},
	}

	// Create JWT manager for testing
	jwtManager, err := crypto.NewJWTManagerForTesting(cfg.Issuer, cfg.JWT.ExpiryMinutes)
	require.NoError(t, err)

	// Create session manager
	sessionCfg := session.DefaultConfig(store)
	sessionMgr := session.NewManager(sessionCfg)

	// Create handlers
	h := NewHandlers(store, jwtManager, cfg, sessionMgr)

	// Create test client with front-channel logout URI
	client := &models.Client{
		ID:                                 "test-client",
		Secret:                             "test-secret",
		Name:                               "Test Client",
		GrantTypes:                         []string{"authorization_code"},
		ResponseTypes:                      []string{"code"},
		RedirectURIs:                       []string{"https://example.com/callback"},
		FrontchannelLogoutURI:              "https://example.com/logout",
		FrontchannelLogoutSessionRequired: true,
	}
	err = store.CreateClient(client)
	require.NoError(t, err)

	// Create test user session
	userSession := &models.UserSession{
		ID:                   "test-session",
		UserID:               "test-user",
		AuthTime:             time.Now(),
		AuthenticationMethod: "password",
		ACR:                  "urn:mace:incommon:iap:silver",
		AMR:                  []string{"pwd"},
		LastActivityAt:       time.Now(),
		ExpiresAt:            time.Now().Add(24 * time.Hour),
		CreatedAt:            time.Now(),
	}
	err = store.CreateUserSession(userSession)
	require.NoError(t, err)

	return h, store, userSession, client
}

func TestLogout_NoActiveSession_Success(t *testing.T) {
	h, _, _, _ := setupLogoutTest(t)

	// Create request without session cookie
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute logout
	err := h.Logout(c)
	require.NoError(t, err)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "No active session")
}

func TestLogout_WithSession_NoClients_Success(t *testing.T) {
	h, store, userSession, _ := setupLogoutTest(t)

	// Create request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set session in context
	c.Set(session.UserSessionKey, userSession)

	// Execute logout
	err := h.Logout(c)
	require.NoError(t, err)

	// Verify session was deleted
	session, err := store.GetUserSession(userSession.ID)
	assert.True(t, err != nil || session == nil, "Session should be deleted")

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Logout successful")
}

func TestLogout_WithSession_WithClients_RendersLogoutPage(t *testing.T) {
	h, store, userSession, client := setupLogoutTest(t)

	// Create session-client association
	sessionClient := models.NewSessionClient(userSession.ID, client.ID, "test-sid-123")
	err := store.CreateSessionClient(sessionClient)
	require.NoError(t, err)

	// Create request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set session in context
	c.Set(session.UserSessionKey, userSession)

	// Execute logout
	err = h.Logout(c)
	require.NoError(t, err)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	responseBody := rec.Body.String()
	
	// Should render HTML logout page
	assert.Contains(t, responseBody, "<!DOCTYPE html>")
	assert.Contains(t, responseBody, "Logging out...")
	
	// Should include iframe with logout URL
	assert.Contains(t, responseBody, "<iframe")
	assert.Contains(t, responseBody, "https://example.com/logout")
	assert.Contains(t, responseBody, "sid=test-sid-123")
	assert.Contains(t, responseBody, "iss=https%3A%2F%2Fexample.com")

	// Verify session was deleted
	session, err := store.GetUserSession(userSession.ID)
	assert.True(t, err != nil || session == nil, "Session should be deleted")

	// Verify session-client was deleted
	sessionClients, err := store.GetSessionClientsBySessionID(userSession.ID)
	assert.NoError(t, err)
	assert.Empty(t, sessionClients)
}

func TestLogout_WithPostLogoutRedirectURI_Redirects(t *testing.T) {
	h, store, userSession, client := setupLogoutTest(t)

	// Create session-client association
	sessionClient := models.NewSessionClient(userSession.ID, client.ID, "test-sid-123")
	err := store.CreateSessionClient(sessionClient)
	require.NoError(t, err)

	// Create request with post_logout_redirect_uri
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/logout?post_logout_redirect_uri=https://example.com/logged-out&state=xyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set session in context
	c.Set(session.UserSessionKey, userSession)

	// Execute logout
	err = h.Logout(c)
	require.NoError(t, err)

	// Verify response includes redirect URL in HTML
	assert.Equal(t, http.StatusOK, rec.Code)
	responseBody := rec.Body.String()
	// The URL is escaped in JavaScript, so check for the escaped version
	assert.Contains(t, responseBody, "example.com")
	assert.Contains(t, responseBody, "logged-out")
	assert.Contains(t, responseBody, "state=xyz")

	// Verify session was deleted
	session, err := store.GetUserSession(userSession.ID)
	assert.True(t, err != nil || session == nil, "Session should be deleted")
}

func TestLogout_NoSession_WithPostLogoutRedirectURI_Redirects(t *testing.T) {
	h, _, _, _ := setupLogoutTest(t)

	// Create request with post_logout_redirect_uri but no session
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/logout?post_logout_redirect_uri=https://example.com/logged-out&state=abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute logout
	err := h.Logout(c)
	require.NoError(t, err)

	// Verify redirect
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "https://example.com/logged-out?state=abc", rec.Header().Get("Location"))
}

func TestLogout_ClientWithoutSessionRequired_NoSidParameter(t *testing.T) {
	h, store, userSession, _ := setupLogoutTest(t)

	// Create client without session required
	client := &models.Client{
		ID:                                 "test-client-2",
		Secret:                             "test-secret-2",
		Name:                               "Test Client 2",
		GrantTypes:                         []string{"authorization_code"},
		ResponseTypes:                      []string{"code"},
		RedirectURIs:                       []string{"https://example.com/callback"},
		FrontchannelLogoutURI:              "https://example.com/logout2",
		FrontchannelLogoutSessionRequired: false, // No session ID required
	}
	err := store.CreateClient(client)
	require.NoError(t, err)

	// Create session-client association
	sessionClient := models.NewSessionClient(userSession.ID, client.ID, "test-sid-456")
	err = store.CreateSessionClient(sessionClient)
	require.NoError(t, err)

	// Create request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set session in context
	c.Set(session.UserSessionKey, userSession)

	// Execute logout
	err = h.Logout(c)
	require.NoError(t, err)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
	responseBody := rec.Body.String()
	
	// Should include iframe with logout URL (no sid parameter)
	assert.Contains(t, responseBody, "https://example.com/logout2")
	// Should NOT include sid or iss parameters when session not required
	assert.NotContains(t, responseBody, "sid=")
	assert.NotContains(t, responseBody, "iss=")
}

func TestGenerateSid_ReturnsUniqueValues(t *testing.T) {
	// Generate multiple session IDs
	sid1 := generateSid()
	sid2 := generateSid()
	sid3 := generateSid()

	// Verify they are not empty
	assert.NotEmpty(t, sid1)
	assert.NotEmpty(t, sid2)
	assert.NotEmpty(t, sid3)

	// Verify they are unique
	assert.NotEqual(t, sid1, sid2)
	assert.NotEqual(t, sid2, sid3)
	assert.NotEqual(t, sid1, sid3)

	// Verify they are base64 URL-safe encoded
	assert.NotContains(t, sid1, "+")
	assert.NotContains(t, sid1, "/")
	assert.NotContains(t, sid1, "=")
}
