package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/prasenjit-net/openid-golang/pkg/config"
	"github.com/prasenjit-net/openid-golang/pkg/models"
)

// MockStorage is a mock implementation of the storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetTokenByAccessToken(accessToken string) (*models.Token, error) {
	args := m.Called(accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Token), args.Error(1)
}

func (m *MockStorage) GetUserByID(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// Add stubs for other required storage methods
func (m *MockStorage) CreateUser(user *models.User) error                      { return nil }
func (m *MockStorage) GetUserByUsername(username string) (*models.User, error) { return nil, nil }
func (m *MockStorage) GetUserByEmail(email string) (*models.User, error)       { return nil, nil }
func (m *MockStorage) GetAllUsers() ([]*models.User, error)                    { return nil, nil }
func (m *MockStorage) UpdateUser(user *models.User) error                      { return nil }
func (m *MockStorage) DeleteUser(id string) error                              { return nil }
func (m *MockStorage) CreateClient(client *models.Client) error                { return nil }
func (m *MockStorage) GetClientByID(id string) (*models.Client, error)         { return nil, nil }
func (m *MockStorage) GetAllClients() ([]*models.Client, error)                { return nil, nil }
func (m *MockStorage) UpdateClient(client *models.Client) error                { return nil }
func (m *MockStorage) DeleteClient(id string) error                            { return nil }
func (m *MockStorage) ValidateClient(clientID, clientSecret string) (*models.Client, error) {
	return nil, nil
}
func (m *MockStorage) CreateAuthorizationCode(code *models.AuthorizationCode) error { return nil }
func (m *MockStorage) GetAuthorizationCode(code string) (*models.AuthorizationCode, error) {
	return nil, nil
}
func (m *MockStorage) UpdateAuthorizationCode(code *models.AuthorizationCode) error { return nil }
func (m *MockStorage) DeleteAuthorizationCode(code string) error                    { return nil }
func (m *MockStorage) CreateToken(token *models.Token) error                        { return nil }
func (m *MockStorage) GetTokenByRefreshToken(refreshToken string) (*models.Token, error) {
	return nil, nil
}
func (m *MockStorage) DeleteToken(accessToken string) error                  { return nil }
func (m *MockStorage) CreateSession(session *models.Session) error           { return nil }
func (m *MockStorage) GetSession(id string) (*models.Session, error)         { return nil, nil }
func (m *MockStorage) DeleteSession(id string) error                         { return nil }
func (m *MockStorage) CreateAuthSession(session *models.AuthSession) error   { return nil }
func (m *MockStorage) GetAuthSession(id string) (*models.AuthSession, error) { return nil, nil }
func (m *MockStorage) UpdateAuthSession(session *models.AuthSession) error   { return nil }
func (m *MockStorage) DeleteAuthSession(id string) error                     { return nil }
func (m *MockStorage) CreateUserSession(session *models.UserSession) error   { return nil }
func (m *MockStorage) GetUserSession(id string) (*models.UserSession, error) { return nil, nil }
func (m *MockStorage) GetUserSessionByUserID(userID string) (*models.UserSession, error) {
	return nil, nil
}
func (m *MockStorage) UpdateUserSession(session *models.UserSession) error { return nil }
func (m *MockStorage) DeleteUserSession(id string) error                   { return nil }
func (m *MockStorage) CleanupExpiredSessions() error                       { return nil }
func (m *MockStorage) CreateConsent(consent *models.Consent) error         { return nil }
func (m *MockStorage) GetConsent(userID, clientID string) (*models.Consent, error) {
	return nil, nil
}
func (m *MockStorage) UpdateConsent(consent *models.Consent) error { return nil }
func (m *MockStorage) DeleteConsent(userID, clientID string) error { return nil }
func (m *MockStorage) DeleteConsentsForUser(userID string) error   { return nil }
func (m *MockStorage) GetTokensByAuthCode(authCodeID string) ([]*models.Token, error) {
	return nil, nil
}
func (m *MockStorage) RevokeTokensByAuthCode(authCodeID string) error { return nil }
func (m *MockStorage) Close() error                                   { return nil }

func TestUserInfo_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockStorage := new(MockStorage)

	handlers := &Handlers{
		storage: mockStorage,
		config: &config.Config{
			JWT: config.JWTConfig{
				ExpiryMinutes: 60,
			},
		},
	}

	// Create test user
	user := &models.User{
		ID:         "user123",
		Username:   "testuser",
		Email:      "test@example.com",
		Name:       "Test User",
		GivenName:  "Test",
		FamilyName: "User",
		Picture:    "https://example.com/pic.jpg",
		UpdatedAt:  time.Now(),
	}

	// Create test token with openid and profile scopes
	token := &models.Token{
		ID:          "token123",
		AccessToken: "valid_access_token",
		UserID:      "user123",
		ClientID:    "client123",
		Scope:       "openid profile email",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		CreatedAt:   time.Now(),
	}

	// Setup mock expectations
	mockStorage.On("GetTokenByAccessToken", "valid_access_token").Return(token, nil)
	mockStorage.On("GetUserByID", "user123").Return(user, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/userinfo", nil)
	req.Header.Set("Authorization", "Bearer valid_access_token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handlers.UserInfo(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response UserInfoResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "user123", response.Sub)
	assert.Equal(t, "Test User", response.Name)
	assert.Equal(t, "Test", response.GivenName)
	assert.Equal(t, "User", response.FamilyName)
	assert.Equal(t, "test@example.com", response.Email)
	assert.True(t, response.EmailVerified)
	assert.Equal(t, "https://example.com/pic.jpg", response.Picture)
	assert.NotZero(t, response.UpdatedAt)

	mockStorage.AssertExpectations(t)
}

func TestUserInfo_ProfileScopeOnly(t *testing.T) {
	// Setup
	e := echo.New()
	mockStorage := new(MockStorage)

	handlers := &Handlers{
		storage: mockStorage,
		config: &config.Config{
			JWT: config.JWTConfig{
				ExpiryMinutes: 60,
			},
		},
	}

	user := &models.User{
		ID:         "user123",
		Email:      "test@example.com",
		Name:       "Test User",
		GivenName:  "Test",
		FamilyName: "User",
	}

	// Token with only profile scope (no email scope)
	token := &models.Token{
		ID:          "token123",
		AccessToken: "profile_only_token",
		UserID:      "user123",
		Scope:       "openid profile",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	mockStorage.On("GetTokenByAccessToken", "profile_only_token").Return(token, nil)
	mockStorage.On("GetUserByID", "user123").Return(user, nil)

	req := httptest.NewRequest(http.MethodGet, "/userinfo", nil)
	req.Header.Set("Authorization", "Bearer profile_only_token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handlers.UserInfo(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response UserInfoResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Should include profile claims
	assert.Equal(t, "Test User", response.Name)
	assert.Equal(t, "Test", response.GivenName)
	// Should NOT include email (no email scope)
	assert.Empty(t, response.Email)

	mockStorage.AssertExpectations(t)
}

func TestUserInfo_MissingAuthHeader(t *testing.T) {
	e := echo.New()
	handlers := &Handlers{}

	req := httptest.NewRequest(http.MethodGet, "/userinfo", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handlers.UserInfo(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_token", response["error"])
}

func TestUserInfo_InvalidToken(t *testing.T) {
	e := echo.New()
	mockStorage := new(MockStorage)
	handlers := &Handlers{
		storage: mockStorage,
	}

	mockStorage.On("GetTokenByAccessToken", "invalid_token").Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/userinfo", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handlers.UserInfo(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_token", response["error"])
}

func TestUserInfo_ExpiredToken(t *testing.T) {
	e := echo.New()
	mockStorage := new(MockStorage)
	handlers := &Handlers{
		storage: mockStorage,
	}

	// Expired token
	token := &models.Token{
		AccessToken: "expired_token",
		UserID:      "user123",
		Scope:       "openid",
		ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}

	mockStorage.On("GetTokenByAccessToken", "expired_token").Return(token, nil)

	req := httptest.NewRequest(http.MethodGet, "/userinfo", nil)
	req.Header.Set("Authorization", "Bearer expired_token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handlers.UserInfo(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_token", response["error"])
}

func TestUserInfo_MissingOpenIDScope(t *testing.T) {
	e := echo.New()
	mockStorage := new(MockStorage)
	handlers := &Handlers{
		storage: mockStorage,
	}

	// Token without openid scope
	token := &models.Token{
		AccessToken: "no_openid_token",
		UserID:      "user123",
		Scope:       "profile email", // Missing openid scope
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	mockStorage.On("GetTokenByAccessToken", "no_openid_token").Return(token, nil)

	req := httptest.NewRequest(http.MethodGet, "/userinfo", nil)
	req.Header.Set("Authorization", "Bearer no_openid_token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handlers.UserInfo(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "insufficient_scope", response["error"])
}
