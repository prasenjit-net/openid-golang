package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/prasenjit-net/openid-golang/pkg/models"
	"github.com/prasenjit-net/openid-golang/pkg/storage"
)

const (
	testRedirectURIJSON = `{"redirect_uris": ["https://app.example.com/callback"]}`
)

func TestRegister_Success(t *testing.T) {
	// Setup
	tmpFile := t.TempDir() + "/test_register.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	// Create registration request
	reqBody := `{
		"redirect_uris": ["https://client.example.com/callback"],
		"client_name": "Test Client",
		"grant_types": ["authorization_code", "refresh_token"],
		"response_types": ["code"],
		"scope": "openid profile email"
	}`

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err = handlers.Register(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response fields
	assert.NotEmpty(t, response.ID, "client_id should be generated")
	assert.NotEmpty(t, response.Secret, "client_secret should be generated")
	assert.Equal(t, int64(0), response.SecretExpiresAt, "secret should never expire")
	assert.Equal(t, "Test Client", response.ClientName)
	assert.Equal(t, []string{"https://client.example.com/callback"}, response.RedirectURIs)
	assert.Equal(t, []string{"authorization_code", "refresh_token"}, response.GrantTypes)
	assert.Equal(t, []string{"code"}, response.ResponseTypes)
	assert.Equal(t, "openid profile email", response.Scope)
	assert.Equal(t, "web", response.ApplicationType, "default application_type should be 'web'")
	assert.Equal(t, "public", response.SubjectType, "default subject_type should be 'public'")
	assert.Equal(t, "client_secret_basic", response.TokenEndpointAuthMethod, "default auth method")
	assert.NotEmpty(t, response.RegistrationAccessToken, "registration access token should be generated")
	assert.NotEmpty(t, response.RegistrationClientURI, "registration client URI should be generated")
	assert.Contains(t, response.RegistrationClientURI, response.ID, "URI should contain client_id")
	assert.NotZero(t, response.ClientIDIssuedAt, "client_id_issued_at should be set")

	// Verify client was stored
	storedClient, err := store.GetClientByID(response.ID)
	require.NoError(t, err)
	assert.NotNil(t, storedClient)
	assert.Equal(t, response.ClientName, storedClient.ClientName)
}

func TestRegister_DisabledRegistration(t *testing.T) {
	// Setup
	tmpFile := t.TempDir() + "/test_disabled.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  false, // Registration disabled
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	reqBody := `{"redirect_uris": ["https://client.example.com/callback"]}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err = handlers.Register(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var errResp models.ClientRegistrationError
	err = json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "registration_not_supported", errResp.Error)
}

func TestRegister_MissingRedirectURIs(t *testing.T) {
	// Setup
	tmpFile := t.TempDir() + "/test_missing_uri.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	reqBody := `{"client_name": "Test Client"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err = handlers.Register(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp models.ClientRegistrationError
	err = json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrInvalidRedirectURI, errResp.Error)
	assert.Contains(t, errResp.ErrorDescription, "redirect_uris is required")
}

func TestRegister_InvalidRedirectURI(t *testing.T) {
	testCases := []struct {
		name        string
		redirectURI string
		errorDesc   string
	}{
		{
			name:        "http not allowed for web apps",
			redirectURI: "http://example.com/callback",
			errorDesc:   "must use HTTPS",
		},
		{
			name:        "relative URI not allowed",
			redirectURI: "/callback",
			errorDesc:   "must be an absolute URI",
		},
		{
			name:        "fragment not allowed",
			redirectURI: "https://example.com/callback#fragment",
			errorDesc:   "must not contain a fragment",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile := t.TempDir() + "/test_invalid_uri.json"
			store, err := storage.NewJSONStorage(tmpFile)
			require.NoError(t, err)
			defer func() {
				_ = store.Close()
			}()

			cfg := &configstore.ConfigData{
				Issuer: "https://example.com",
				Registration: configstore.RegistrationConfig{
					Enabled:  true,
					Endpoint: "/register",
				},
			}

			handlers := &Handlers{
				storage: store,
				config:  cfg,
			}

			reqBody := `{"redirect_uris": ["` + tc.redirectURI + `"]}`
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			err = handlers.Register(c)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var errResp models.ClientRegistrationError
			err = json.Unmarshal(rec.Body.Bytes(), &errResp)
			require.NoError(t, err)
			assert.Equal(t, models.ErrInvalidRedirectURI, errResp.Error)
			assert.Contains(t, errResp.ErrorDescription, tc.errorDesc)
		})
	}
}

func TestRegister_LocalhostAllowed(t *testing.T) {
	tmpFile := t.TempDir() + "/test_localhost.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	reqBody := `{"redirect_uris": ["http://localhost:3000/callback"]}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err = handlers.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestRegister_PublicClient(t *testing.T) {
	tmpFile := t.TempDir() + "/test_public_client.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	// Public client using implicit flow
	reqBody := `{
		"redirect_uris": ["https://app.example.com/callback"],
		"grant_types": ["implicit"],
		"response_types": ["id_token token"],
		"token_endpoint_auth_method": "none"
	}`

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err = handlers.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Empty(t, response.Secret, "public client should not have a secret")
	assert.Equal(t, "none", response.TokenEndpointAuthMethod)
}

func TestRegister_GrantTypeResponseTypeConsistency(t *testing.T) {
	testCases := []struct {
		name          string
		grantTypes    []string
		responseTypes []string
		shouldError   bool
		errorDesc     string
	}{
		{
			name:          "code response requires authorization_code grant",
			grantTypes:    []string{"implicit"},
			responseTypes: []string{"code"},
			shouldError:   true,
			errorDesc:     "does not include 'authorization_code'",
		},
		{
			name:          "implicit response requires implicit grant",
			grantTypes:    []string{"authorization_code"},
			responseTypes: []string{"id_token token"},
			shouldError:   true,
			errorDesc:     "does not include 'implicit'",
		},
		{
			name:          "valid authorization code flow",
			grantTypes:    []string{"authorization_code"},
			responseTypes: []string{"code"},
			shouldError:   false,
		},
		{
			name:          "valid implicit flow",
			grantTypes:    []string{"implicit"},
			responseTypes: []string{"id_token token"},
			shouldError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile := t.TempDir() + "/test_consistency.json"
			store, err := storage.NewJSONStorage(tmpFile)
			require.NoError(t, err)
			defer func() {
				_ = store.Close()
			}()

			cfg := &configstore.ConfigData{
				Issuer: "https://example.com",
				Registration: configstore.RegistrationConfig{
					Enabled:  true,
					Endpoint: "/register",
				},
			}

			handlers := &Handlers{
				storage: store,
				config:  cfg,
			}

			grantTypesJSON, _ := json.Marshal(tc.grantTypes)
			responseTypesJSON, _ := json.Marshal(tc.responseTypes)

			reqBody := `{
				"redirect_uris": ["https://app.example.com/callback"],
				"grant_types": ` + string(grantTypesJSON) + `,
				"response_types": ` + string(responseTypesJSON) + `
			}`

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			err = handlers.Register(c)
			assert.NoError(t, err)

			if tc.shouldError {
				assert.Equal(t, http.StatusBadRequest, rec.Code)
				var errResp models.ClientRegistrationError
				err = json.Unmarshal(rec.Body.Bytes(), &errResp)
				require.NoError(t, err)
				assert.Contains(t, errResp.ErrorDescription, tc.errorDesc)
			} else {
				assert.Equal(t, http.StatusCreated, rec.Code)
			}
		})
	}
}

func TestRegister_NativeApp(t *testing.T) {
	tmpFile := t.TempDir() + "/test_native.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	// Native app with custom scheme
	reqBody := `{
		"redirect_uris": ["com.example.app:/callback", "http://localhost:8080/callback"],
		"application_type": "native",
		"client_name": "Native App"
	}`

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err = handlers.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "native", response.ApplicationType)
	assert.Contains(t, response.RedirectURIs, "com.example.app:/callback")
	assert.Contains(t, response.RedirectURIs, "http://localhost:8080/callback")
}

func TestRegister_InvalidJWKS(t *testing.T) {
	tmpFile := t.TempDir() + "/test_jwks.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	// Invalid: both jwks and jwks_uri
	reqBody := `{
		"redirect_uris": ["https://app.example.com/callback"],
		"jwks": {
			"keys": [
				{
					"kty": "RSA",
					"use": "sig",
					"kid": "test-key",
					"n": "test-n",
					"e": "AQAB"
				}
			]
		},
		"jwks_uri": "https://app.example.com/jwks"
	}`

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err = handlers.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp models.ClientRegistrationError
	err = json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrInvalidClientMetadata, errResp.Error)
	assert.Contains(t, errResp.ErrorDescription, "Cannot specify both")
}

func TestRegister_FullMetadata(t *testing.T) {
	tmpFile := t.TempDir() + "/test_full_metadata.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	reqBody := `{
		"redirect_uris": ["https://app.example.com/callback"],
		"client_name": "Full Metadata Client",
		"logo_uri": "https://app.example.com/logo.png",
		"client_uri": "https://app.example.com",
		"policy_uri": "https://app.example.com/policy",
		"tos_uri": "https://app.example.com/tos",
		"contacts": ["admin@example.com", "support@example.com"],
		"grant_types": ["authorization_code", "refresh_token"],
		"response_types": ["code"],
		"scope": "openid profile email",
		"subject_type": "public",
		"id_token_signed_response_alg": "RS256",
		"token_endpoint_auth_method": "client_secret_post",
		"default_max_age": 3600,
		"require_auth_time": true
	}`

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err = handlers.Register(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Full Metadata Client", response.ClientName)
	assert.Equal(t, "https://app.example.com/logo.png", response.LogoURI)
	assert.Equal(t, "https://app.example.com", response.ClientURI)
	assert.Equal(t, "https://app.example.com/policy", response.PolicyURI)
	assert.Equal(t, "https://app.example.com/tos", response.TosURI)
	assert.Equal(t, []string{"admin@example.com", "support@example.com"}, response.Contacts)
	assert.Equal(t, "client_secret_post", response.TokenEndpointAuthMethod)
	assert.Equal(t, 3600, response.DefaultMaxAge)
	assert.True(t, response.RequireAuthTime)
}

// ============================================================================
// GET /register/:client_id Tests
// ============================================================================

func TestGetClientConfiguration_Success(t *testing.T) {
	tmpFile := t.TempDir() + "/test_get_config.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	// First, register a client
	reqBody := testRedirectURIJSON
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	err = handlers.Register(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var regResponse models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &regResponse)
	require.NoError(t, err)

	// Now, get the client configuration
	req = httptest.NewRequest(http.MethodGet, "/register/"+regResponse.ID, nil)
	req.Header.Set("Authorization", "Bearer "+regResponse.RegistrationAccessToken)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("client_id")
	c.SetParamValues(regResponse.ID)

	err = handlers.GetClientConfiguration(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var getResponse models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &getResponse)
	require.NoError(t, err)

	// Verify response matches original registration
	assert.Equal(t, regResponse.ID, getResponse.ID)
	assert.Equal(t, regResponse.RedirectURIs, getResponse.RedirectURIs)
	assert.Equal(t, regResponse.RegistrationAccessToken, getResponse.RegistrationAccessToken)
	assert.NotEmpty(t, getResponse.RegistrationClientURI)
}

func TestGetClientConfiguration_DisabledRegistration(t *testing.T) {
	tmpFile := t.TempDir() + "/test_get_disabled.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  false,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	req := httptest.NewRequest(http.MethodGet, "/register/some-client-id", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetParamNames("client_id")
	c.SetParamValues("some-client-id")

	err = handlers.GetClientConfiguration(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var errResp models.ClientRegistrationError
	err = json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "registration_not_supported", errResp.Error)
}

func TestGetClientConfiguration_MissingToken(t *testing.T) {
	tmpFile := t.TempDir() + "/test_get_no_token.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	req := httptest.NewRequest(http.MethodGet, "/register/some-client-id", nil)
	// No Authorization header
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetParamNames("client_id")
	c.SetParamValues("some-client-id")

	err = handlers.GetClientConfiguration(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errResp models.ClientRegistrationError
	err = json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_token", errResp.Error)
	assert.Contains(t, errResp.ErrorDescription, "required")
}

func TestGetClientConfiguration_InvalidToken(t *testing.T) {
	tmpFile := t.TempDir() + "/test_get_invalid_token.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	// First, register a client
	reqBody := testRedirectURIJSON
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	err = handlers.Register(c)
	require.NoError(t, err)

	var regResponse models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &regResponse)
	require.NoError(t, err)

	// Try to get config with wrong token
	req = httptest.NewRequest(http.MethodGet, "/register/"+regResponse.ID, nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("client_id")
	c.SetParamValues(regResponse.ID)

	err = handlers.GetClientConfiguration(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errResp models.ClientRegistrationError
	err = json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_token", errResp.Error)
}

func TestGetClientConfiguration_NonExistentClient(t *testing.T) {
	tmpFile := t.TempDir() + "/test_get_nonexistent.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	req := httptest.NewRequest(http.MethodGet, "/register/nonexistent-client", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetParamNames("client_id")
	c.SetParamValues("nonexistent-client")

	err = handlers.GetClientConfiguration(c)
	require.NoError(t, err)
	// Per spec: Always return 401, never 404 (security requirement)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errResp models.ClientRegistrationError
	err = json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_token", errResp.Error)
}

// ============================================================================
// PUT /register/:client_id Tests
// ============================================================================

func TestUpdateClientConfiguration_Success(t *testing.T) {
	tmpFile := t.TempDir() + "/test_update_config.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	// First, register a client
	reqBody := `{"redirect_uris": ["https://app.example.com/callback"], "client_name": "Original Name"}`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	err = handlers.Register(c)
	require.NoError(t, err)

	var regResponse models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &regResponse)
	require.NoError(t, err)

	// Now, update the client
	updateBody := `{
		"redirect_uris": ["https://app.example.com/callback", "https://app.example.com/callback2"],
		"client_name": "Updated Name",
		"logo_uri": "https://app.example.com/logo.png"
	}`
	req = httptest.NewRequest(http.MethodPut, "/register/"+regResponse.ID, strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+regResponse.RegistrationAccessToken)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("client_id")
	c.SetParamValues(regResponse.ID)

	err = handlers.UpdateClientConfiguration(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var updateResponse models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &updateResponse)
	require.NoError(t, err)

	// Verify updates were applied
	assert.Equal(t, regResponse.ID, updateResponse.ID) // ID should not change
	assert.Equal(t, "Updated Name", updateResponse.ClientName)
	assert.Equal(t, "https://app.example.com/logo.png", updateResponse.LogoURI)
	assert.Equal(t, 2, len(updateResponse.RedirectURIs))
	assert.Contains(t, updateResponse.RedirectURIs, "https://app.example.com/callback2")

	// Secret should be preserved
	assert.Equal(t, regResponse.Secret, updateResponse.Secret)
}

func TestUpdateClientConfiguration_InvalidToken(t *testing.T) {
	tmpFile := t.TempDir() + "/test_update_invalid_token.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	// Register a client
	reqBody := testRedirectURIJSON
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	err = handlers.Register(c)
	require.NoError(t, err)

	var regResponse models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &regResponse)
	require.NoError(t, err)

	// Try to update with wrong token
	updateBody := `{"redirect_uris": ["https://app.example.com/callback"], "client_name": "Hacked"}`
	req = httptest.NewRequest(http.MethodPut, "/register/"+regResponse.ID, strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("client_id")
	c.SetParamValues(regResponse.ID)

	err = handlers.UpdateClientConfiguration(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errResp models.ClientRegistrationError
	err = json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid_token", errResp.Error)
}

func TestUpdateClientConfiguration_ValidationError(t *testing.T) {
	tmpFile := t.TempDir() + "/test_update_validation.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	// Register a client
	reqBody := testRedirectURIJSON
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	err = handlers.Register(c)
	require.NoError(t, err)

	var regResponse models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &regResponse)
	require.NoError(t, err)

	// Try to update with invalid redirect URI (http not allowed)
	updateBody := `{"redirect_uris": ["http://example.com/callback"]}`
	req = httptest.NewRequest(http.MethodPut, "/register/"+regResponse.ID, strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+regResponse.RegistrationAccessToken)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("client_id")
	c.SetParamValues(regResponse.ID)

	err = handlers.UpdateClientConfiguration(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp models.ClientRegistrationError
	err = json.Unmarshal(rec.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, models.ErrInvalidRedirectURI, errResp.Error)
	assert.Contains(t, errResp.ErrorDescription, "HTTPS")
}

// ============================================================================
// DELETE /register/:client_id Tests
// ============================================================================

func TestDeleteClientConfiguration_Success(t *testing.T) {
	tmpFile := t.TempDir() + "/test_delete_config.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	// First, register a client
	reqBody := testRedirectURIJSON
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	err = handlers.Register(c)
	require.NoError(t, err)

	var regResponse models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &regResponse)
	require.NoError(t, err)

	// Now, delete the client
	req = httptest.NewRequest(http.MethodDelete, "/register/"+regResponse.ID, nil)
	req.Header.Set("Authorization", "Bearer "+regResponse.RegistrationAccessToken)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("client_id")
	c.SetParamValues(regResponse.ID)

	err = handlers.DeleteClientConfiguration(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify client is actually deleted
	deletedClient, err := store.GetClientByID(regResponse.ID)
	assert.Nil(t, deletedClient)
	assert.NoError(t, err) // JSON storage returns (nil, nil) for non-existent clients
}

func TestDeleteClientConfiguration_InvalidToken(t *testing.T) {
	tmpFile := t.TempDir() + "/test_delete_invalid_token.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	// Register a client
	reqBody := testRedirectURIJSON
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)

	err = handlers.Register(c)
	require.NoError(t, err)

	var regResponse models.ClientRegistrationResponse
	err = json.Unmarshal(rec.Body.Bytes(), &regResponse)
	require.NoError(t, err)

	// Try to delete with wrong token
	req = httptest.NewRequest(http.MethodDelete, "/register/"+regResponse.ID, nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("client_id")
	c.SetParamValues(regResponse.ID)

	err = handlers.DeleteClientConfiguration(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Verify client was NOT deleted
	existingClient, err := store.GetClientByID(regResponse.ID)
	assert.NotNil(t, existingClient)
	assert.NoError(t, err)
}

func TestDeleteClientConfiguration_NonExistentClient(t *testing.T) {
	tmpFile := t.TempDir() + "/test_delete_nonexistent.json"
	store, err := storage.NewJSONStorage(tmpFile)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/register",
		},
	}

	handlers := &Handlers{
		storage: store,
		config:  cfg,
	}

	req := httptest.NewRequest(http.MethodDelete, "/register/nonexistent-client", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rec := httptest.NewRecorder()
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetParamNames("client_id")
	c.SetParamValues("nonexistent-client")

	err = handlers.DeleteClientConfiguration(c)
	require.NoError(t, err)
	// Per spec: Always return 401, never 404
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
