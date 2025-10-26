package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/crypto"
	"github.com/prasenjit-net/openid-golang/pkg/models"
)

// TestIDToken_ProfileScope tests that ID token only includes profile claims when profile scope is requested
func TestIDToken_ProfileScope(t *testing.T) {
	// Setup JWT manager
	jwtManager, err := crypto.NewJWTManagerForTesting("https://test-issuer.example.com", 60)
	assert.NoError(t, err)

	// Create test user with all data
	user := &models.User{
		ID:            "user123",
		Username:      "testuser",
		Email:         "test@example.com",
		EmailVerified: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Picture:       "https://example.com/pic.jpg",
		Address: &models.Address{
			Formatted:     "123 Main St, City, ST 12345, Country",
			StreetAddress: "123 Main St",
			Locality:      "City",
			Region:        "ST",
			PostalCode:    "12345",
			Country:       "Country",
		},
	}

	// Test with only profile scope
	token, err := jwtManager.GenerateIDToken(user, "client123", "nonce123", "openid profile")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := jwtManager.ValidateToken(token)
	assert.NoError(t, err)

	// Should have profile claims
	assert.Equal(t, "Test User", claims.Name)
	assert.Equal(t, "Test", claims.GivenName)
	assert.Equal(t, "User", claims.FamilyName)
	assert.Equal(t, "https://example.com/pic.jpg", claims.Picture)

	// Should NOT have email claims
	assert.Empty(t, claims.Email)
	assert.False(t, claims.EmailVerified)

	// Should NOT have address claims
	assert.Nil(t, claims.Address)
}

// TestIDToken_EmailScope tests that ID token only includes email claims when email scope is requested
func TestIDToken_EmailScope(t *testing.T) {
	// Setup JWT manager
	jwtManager, err := crypto.NewJWTManagerForTesting("https://test-issuer.example.com", 60)
	assert.NoError(t, err)

	// Create test user with all data
	user := &models.User{
		ID:            "user123",
		Email:         "test@example.com",
		EmailVerified: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Picture:       "https://example.com/pic.jpg",
		Address: &models.Address{
			Formatted: "123 Main St, City, ST 12345, Country",
		},
	}

	// Test with only email scope
	token, err := jwtManager.GenerateIDToken(user, "client123", "nonce123", "openid email")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := jwtManager.ValidateToken(token)
	assert.NoError(t, err)

	// Should have email claims
	assert.Equal(t, "test@example.com", claims.Email)
	assert.True(t, claims.EmailVerified)

	// Should NOT have profile claims
	assert.Empty(t, claims.Name)
	assert.Empty(t, claims.GivenName)
	assert.Empty(t, claims.FamilyName)
	assert.Empty(t, claims.Picture)

	// Should NOT have address claims
	assert.Nil(t, claims.Address)
}

// TestIDToken_AddressScope tests that ID token only includes address claims when address scope is requested
func TestIDToken_AddressScope(t *testing.T) {
	// Setup JWT manager
	jwtManager, err := crypto.NewJWTManagerForTesting("https://test-issuer.example.com", 60)
	assert.NoError(t, err)

	// Create test user with all data
	user := &models.User{
		ID:            "user123",
		Email:         "test@example.com",
		EmailVerified: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Picture:       "https://example.com/pic.jpg",
		Address: &models.Address{
			Formatted:     "123 Main St, City, ST 12345, Country",
			StreetAddress: "123 Main St",
			Locality:      "City",
			Region:        "ST",
			PostalCode:    "12345",
			Country:       "Country",
		},
	}

	// Test with only address scope
	token, err := jwtManager.GenerateIDToken(user, "client123", "nonce123", "openid address")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := jwtManager.ValidateToken(token)
	assert.NoError(t, err)

	// Should have address claims
	assert.NotNil(t, claims.Address)
	assert.Equal(t, "123 Main St, City, ST 12345, Country", claims.Address.Formatted)
	assert.Equal(t, "123 Main St", claims.Address.StreetAddress)
	assert.Equal(t, "City", claims.Address.Locality)
	assert.Equal(t, "ST", claims.Address.Region)
	assert.Equal(t, "12345", claims.Address.PostalCode)
	assert.Equal(t, "Country", claims.Address.Country)

	// Should NOT have profile claims
	assert.Empty(t, claims.Name)
	assert.Empty(t, claims.GivenName)
	assert.Empty(t, claims.FamilyName)
	assert.Empty(t, claims.Picture)

	// Should NOT have email claims
	assert.Empty(t, claims.Email)
	assert.False(t, claims.EmailVerified)
}

// TestIDToken_MultipleScopesAtOnce tests that ID token includes claims for all requested scopes
func TestIDToken_MultipleScopesAtOnce(t *testing.T) {
	// Setup JWT manager
	jwtManager, err := crypto.NewJWTManagerForTesting("https://test-issuer.example.com", 60)
	assert.NoError(t, err)

	// Create test user with all data
	user := &models.User{
		ID:            "user123",
		Email:         "test@example.com",
		EmailVerified: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Picture:       "https://example.com/pic.jpg",
		Address: &models.Address{
			Formatted: "123 Main St, City, ST 12345, Country",
			Locality:  "City",
		},
	}

	// Test with all scopes
	token, err := jwtManager.GenerateIDToken(user, "client123", "nonce123", "openid profile email address")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := jwtManager.ValidateToken(token)
	assert.NoError(t, err)

	// Should have profile claims
	assert.Equal(t, "Test User", claims.Name)
	assert.Equal(t, "Test", claims.GivenName)
	assert.Equal(t, "User", claims.FamilyName)
	assert.Equal(t, "https://example.com/pic.jpg", claims.Picture)

	// Should have email claims
	assert.Equal(t, "test@example.com", claims.Email)
	assert.True(t, claims.EmailVerified)

	// Should have address claims
	assert.NotNil(t, claims.Address)
	assert.Equal(t, "123 Main St, City, ST 12345, Country", claims.Address.Formatted)
	assert.Equal(t, "City", claims.Address.Locality)
}

// TestIDToken_NoScopes tests that ID token only includes required claims when no optional scopes are requested
func TestIDToken_NoScopes(t *testing.T) {
	// Setup JWT manager
	jwtManager, err := crypto.NewJWTManagerForTesting("https://test-issuer.example.com", 60)
	assert.NoError(t, err)

	// Create test user with all data
	user := &models.User{
		ID:            "user123",
		Email:         "test@example.com",
		EmailVerified: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Picture:       "https://example.com/pic.jpg",
		Address: &models.Address{
			Formatted: "123 Main St",
		},
	}

	// Test with only openid scope (no profile, email, or address)
	token, err := jwtManager.GenerateIDToken(user, "client123", "nonce123", "openid")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := jwtManager.ValidateToken(token)
	assert.NoError(t, err)

	// Should have required claims
	assert.Equal(t, "user123", claims.Subject)
	assert.Equal(t, "client123", claims.Audience[0])
	assert.Equal(t, "nonce123", claims.Nonce)

	// Should NOT have any optional claims
	assert.Empty(t, claims.Name)
	assert.Empty(t, claims.GivenName)
	assert.Empty(t, claims.FamilyName)
	assert.Empty(t, claims.Picture)
	assert.Empty(t, claims.Email)
	assert.False(t, claims.EmailVerified)
	assert.Nil(t, claims.Address)
}

// TestIDToken_AddressScopeWithoutAddressData tests that address claim is nil when user has no address
func TestIDToken_AddressScopeWithoutAddressData(t *testing.T) {
	// Setup JWT manager
	jwtManager, err := crypto.NewJWTManagerForTesting("https://test-issuer.example.com", 60)
	assert.NoError(t, err)

	// Create test user without address
	user := &models.User{
		ID:    "user123",
		Name:  "Test User",
		Email: "test@example.com",
		// No address field set
	}

	// Test with address scope
	token, err := jwtManager.GenerateIDToken(user, "client123", "nonce123", "openid address")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := jwtManager.ValidateToken(token)
	assert.NoError(t, err)

	// Address should be nil even though address scope was requested
	assert.Nil(t, claims.Address)
}

// TestUserInfo_AddressScope tests that UserInfo endpoint only includes address when address scope is granted
func TestUserInfo_AddressScope(t *testing.T) {
	// Setup
	e := echo.New()
	mockStorage := new(MockStorage)

	// Setup JWT manager for handlers
	jwtManager, err := crypto.NewJWTManagerForTesting("https://test-issuer.example.com", 60)
	assert.NoError(t, err)

	handlers := &Handlers{
		storage:    mockStorage,
		jwtManager: jwtManager,
		config: &configstore.ConfigData{
			JWT: configstore.JWTConfig{
				ExpiryMinutes: 60,
			},
		},
	}

	// Create test user with address
	user := &models.User{
		ID:            "user123",
		Email:         "test@example.com",
		EmailVerified: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Address: &models.Address{
			Formatted:     "123 Main St, City, ST 12345, Country",
			StreetAddress: "123 Main St",
			Locality:      "City",
			Region:        "ST",
			PostalCode:    "12345",
			Country:       "Country",
		},
		UpdatedAt: time.Now(),
	}

	// Create test token with openid, profile, email, and address scopes
	token := &models.Token{
		ID:          "token123",
		AccessToken: "valid_access_token",
		UserID:      "user123",
		ClientID:    "client123",
		Scope:       "openid profile email address",
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
	err = handlers.UserInfo(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response UserInfoResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Should have all claims including address
	assert.Equal(t, "user123", response.Sub)
	assert.Equal(t, "Test User", response.Name)
	assert.Equal(t, "test@example.com", response.Email)
	assert.True(t, response.EmailVerified)
	assert.NotNil(t, response.Address)
	assert.Equal(t, "123 Main St, City, ST 12345, Country", response.Address.Formatted)
	assert.Equal(t, "123 Main St", response.Address.StreetAddress)
	assert.Equal(t, "City", response.Address.Locality)
	assert.Equal(t, "ST", response.Address.Region)
	assert.Equal(t, "12345", response.Address.PostalCode)
	assert.Equal(t, "Country", response.Address.Country)

	mockStorage.AssertExpectations(t)
}

// TestUserInfo_WithoutAddressScope tests that UserInfo endpoint excludes address when address scope is not granted
func TestUserInfo_WithoutAddressScope(t *testing.T) {
	// Setup
	e := echo.New()
	mockStorage := new(MockStorage)

	// Setup JWT manager for handlers
	jwtManager, err := crypto.NewJWTManagerForTesting("https://test-issuer.example.com", 60)
	assert.NoError(t, err)

	handlers := &Handlers{
		storage:    mockStorage,
		jwtManager: jwtManager,
		config: &configstore.ConfigData{
			JWT: configstore.JWTConfig{
				ExpiryMinutes: 60,
			},
		},
	}

	// Create test user with address
	user := &models.User{
		ID:            "user123",
		Email:         "test@example.com",
		EmailVerified: true,
		Name:          "Test User",
		Address: &models.Address{
			Formatted: "123 Main St, City, ST 12345, Country",
		},
	}

	// Create test token WITHOUT address scope
	token := &models.Token{
		ID:          "token123",
		AccessToken: "valid_access_token",
		UserID:      "user123",
		ClientID:    "client123",
		Scope:       "openid profile email", // No address scope
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
	err = handlers.UserInfo(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response UserInfoResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Should have profile and email claims
	assert.Equal(t, "user123", response.Sub)
	assert.Equal(t, "Test User", response.Name)
	assert.Equal(t, "test@example.com", response.Email)

	// Should NOT have address even though user has one
	assert.Nil(t, response.Address)

	mockStorage.AssertExpectations(t)
}

// TestUserInfo_WithoutEmailScope tests that UserInfo endpoint excludes email when email scope is not granted
func TestUserInfo_WithoutEmailScope(t *testing.T) {
	// Setup
	e := echo.New()
	mockStorage := new(MockStorage)

	handlers := &Handlers{
		storage: mockStorage,
		config: &configstore.ConfigData{
			JWT: configstore.JWTConfig{
				ExpiryMinutes: 60,
			},
		},
	}

	// Create test user with email
	user := &models.User{
		ID:            "user123",
		Email:         "test@example.com",
		EmailVerified: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
	}

	// Create test token WITHOUT email scope
	token := &models.Token{
		ID:          "token123",
		AccessToken: "valid_access_token",
		UserID:      "user123",
		ClientID:    "client123",
		Scope:       "openid profile", // No email scope
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

	// Should have profile claims
	assert.Equal(t, "user123", response.Sub)
	assert.Equal(t, "Test User", response.Name)
	assert.Equal(t, "Test", response.GivenName)
	assert.Equal(t, "User", response.FamilyName)

	// Should NOT have email claims
	assert.Empty(t, response.Email)
	assert.False(t, response.EmailVerified)

	mockStorage.AssertExpectations(t)
}
