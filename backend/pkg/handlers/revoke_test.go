package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

func setupRevokeTest(t *testing.T) (*Handlers, storage.Storage, *models.Client, *models.Token) {
	// Create storage
	store, err := storage.NewJSONStorage(":memory:")
	require.NoError(t, err)

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

	// Create test client
	client := &models.Client{
		ID:            "test-client",
		Secret:        "test-secret",
		Name:          "Test Client",
		GrantTypes:    []string{"authorization_code", "refresh_token", "client_credentials"},
		ResponseTypes: []string{"code"},
		RedirectURIs:  []string{"https://example.com/callback"},
	}
	err = store.CreateClient(client)
	require.NoError(t, err)

	// Create test token
	token := &models.Token{
		ID:                  "test-token-id",
		AccessToken:         "test-access-token",
		RefreshToken:        "test-refresh-token",
		TokenType:           "Bearer",
		Scope:               "openid profile",
		UserID:              "test-user",
		ClientID:            client.ID,
		AuthorizationCodeID: "test-auth-code",
		CreatedAt:           time.Now(),
		ExpiresAt:           time.Now().Add(time.Hour),
	}
	err = store.CreateToken(token)
	require.NoError(t, err)

	return h, store, client, token
}

func TestRevoke_WithRefreshToken_Success(t *testing.T) {
	h, store, client, token := setupRevokeTest(t)

	// Create request
	e := echo.New()
	form := url.Values{}
	form.Set("token", token.RefreshToken)
	form.Set("token_type_hint", "refresh_token")
	form.Set("client_id", client.ID)
	form.Set("client_secret", client.Secret)

	req := httptest.NewRequest(http.MethodPost, "/revoke", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := h.Revoke(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify token was revoked
	retrievedToken, err := store.GetTokenByRefreshToken(token.RefreshToken)
	assert.NoError(t, err)
	assert.Nil(t, retrievedToken, "Token should be revoked")
}

func TestRevoke_WithAccessToken_Success(t *testing.T) {
	h, store, client, token := setupRevokeTest(t)

	// Create request
	e := echo.New()
	form := url.Values{}
	form.Set("token", token.AccessToken)
	form.Set("token_type_hint", "access_token")
	form.Set("client_id", client.ID)
	form.Set("client_secret", client.Secret)

	req := httptest.NewRequest(http.MethodPost, "/revoke", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := h.Revoke(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify token was revoked
	retrievedToken, err := store.GetTokenByAccessToken(token.AccessToken)
	assert.NoError(t, err)
	assert.Nil(t, retrievedToken, "Token should be revoked")
}

func TestRevoke_WithBasicAuth_Success(t *testing.T) {
	h, store, client, token := setupRevokeTest(t)

	// Create request with Basic Auth
	e := echo.New()
	form := url.Values{}
	form.Set("token", token.RefreshToken)
	form.Set("token_type_hint", "refresh_token")

	req := httptest.NewRequest(http.MethodPost, "/revoke", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.SetBasicAuth(client.ID, client.Secret)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := h.Revoke(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify token was revoked
	retrievedToken, err := store.GetTokenByRefreshToken(token.RefreshToken)
	assert.NoError(t, err)
	assert.Nil(t, retrievedToken, "Token should be revoked")
}

func TestRevoke_WithoutTypeHint_Success(t *testing.T) {
	h, store, client, token := setupRevokeTest(t)

	// Create request without token_type_hint
	e := echo.New()
	form := url.Values{}
	form.Set("token", token.RefreshToken)
	form.Set("client_id", client.ID)
	form.Set("client_secret", client.Secret)

	req := httptest.NewRequest(http.MethodPost, "/revoke", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := h.Revoke(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify token was revoked
	retrievedToken, err := store.GetTokenByRefreshToken(token.RefreshToken)
	assert.NoError(t, err)
	assert.Nil(t, retrievedToken, "Token should be revoked")
}

func TestRevoke_MissingToken_Error(t *testing.T) {
	h, _, client, _ := setupRevokeTest(t)

	// Create request without token parameter
	e := echo.New()
	form := url.Values{}
	form.Set("client_id", client.ID)
	form.Set("client_secret", client.Secret)

	req := httptest.NewRequest(http.MethodPost, "/revoke", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := h.Revoke(c)

	// Assert - should return bad request status
	// Note: jsonError() writes response and returns nil, so we check rec.Code
	assert.NoError(t, err, "Handler shouldn't return error - it writes response")
	assert.Equal(t, http.StatusBadRequest, rec.Code, "Should return 400 Bad Request")
}

func TestRevoke_InvalidClient_Error(t *testing.T) {
	h, _, _, token := setupRevokeTest(t)

	// Create request with invalid client credentials
	e := echo.New()
	form := url.Values{}
	form.Set("token", token.RefreshToken)
	form.Set("client_id", "invalid-client")
	form.Set("client_secret", "invalid-secret")

	req := httptest.NewRequest(http.MethodPost, "/revoke", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := h.Revoke(c)

	// Assert - should return unauthorized status
	// Note: jsonError() writes response and returns nil, so we check rec.Code
	assert.NoError(t, err, "Handler shouldn't return error - it writes response")
	assert.Equal(t, http.StatusUnauthorized, rec.Code, "Should return 401 Unauthorized")
}

func TestRevoke_NonExistentToken_Success(t *testing.T) {
	h, _, client, _ := setupRevokeTest(t)

	// Create request with non-existent token
	// Per RFC 7009 §2.2, this should still return 200
	e := echo.New()
	form := url.Values{}
	form.Set("token", "non-existent-token")
	form.Set("client_id", client.ID)
	form.Set("client_secret", client.Secret)

	req := httptest.NewRequest(http.MethodPost, "/revoke", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := h.Revoke(c)

	// Assert - should succeed per RFC 7009 §2.2
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRevoke_TokenBelongsToOtherClient_Success(t *testing.T) {
	h, store, _, token := setupRevokeTest(t)

	// Create another client
	otherClient := &models.Client{
		ID:            "other-client",
		Secret:        "other-secret",
		Name:          "Other Client",
		GrantTypes:    []string{"authorization_code"},
		ResponseTypes: []string{"code"},
		RedirectURIs:  []string{"https://example.com/callback"},
	}
	err := store.CreateClient(otherClient)
	require.NoError(t, err)

	// Create request from other client trying to revoke first client's token
	// Per RFC 7009 §2.2, this should still return 200 to prevent token scanning
	e := echo.New()
	form := url.Values{}
	form.Set("token", token.RefreshToken)
	form.Set("client_id", otherClient.ID)
	form.Set("client_secret", otherClient.Secret)

	req := httptest.NewRequest(http.MethodPost, "/revoke", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = h.Revoke(c)

	// Assert - should succeed per RFC 7009 §2.2
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify token was NOT revoked (belongs to different client)
	retrievedToken, err := store.GetTokenByRefreshToken(token.RefreshToken)
	assert.NoError(t, err, "Token should still exist")
	assert.NotNil(t, retrievedToken)
}

func TestRevoke_CascadeRevocation_RefreshToken(t *testing.T) {
	h, store, client, _ := setupRevokeTest(t)

	// Create multiple tokens with same auth code
	authCodeID := "test-auth-code-cascade"

	token1 := &models.Token{
		ID:                  "token-1-id",
		AccessToken:         "access-token-1",
		RefreshToken:        "refresh-token-1",
		TokenType:           "Bearer",
		Scope:               "openid profile",
		UserID:              "test-user",
		ClientID:            client.ID,
		AuthorizationCodeID: authCodeID,
		CreatedAt:           time.Now(),
		ExpiresAt:           time.Now().Add(time.Hour),
	}
	err := store.CreateToken(token1)
	require.NoError(t, err)

	token2 := &models.Token{
		ID:                  "token-2-id",
		AccessToken:         "access-token-2",
		RefreshToken:        "refresh-token-2",
		TokenType:           "Bearer",
		Scope:               "openid profile",
		UserID:              "test-user",
		ClientID:            client.ID,
		AuthorizationCodeID: authCodeID,
		CreatedAt:           time.Now(),
		ExpiresAt:           time.Now().Add(time.Hour),
	}
	err = store.CreateToken(token2)
	require.NoError(t, err)

	// Revoke token1's refresh token
	e := echo.New()
	form := url.Values{}
	form.Set("token", token1.RefreshToken)
	form.Set("token_type_hint", "refresh_token")
	form.Set("client_id", client.ID)
	form.Set("client_secret", client.Secret)

	req := httptest.NewRequest(http.MethodPost, "/revoke", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = h.Revoke(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify all tokens with same auth code were revoked
	retrievedToken1, err := store.GetTokenByRefreshToken(token1.RefreshToken)
	assert.NoError(t, err)
	assert.Nil(t, retrievedToken1, "Token 1 should be revoked")

	retrievedToken2, err := store.GetTokenByRefreshToken(token2.RefreshToken)
	assert.NoError(t, err)
	assert.Nil(t, retrievedToken2, "Token 2 should be revoked (cascade)")
}
