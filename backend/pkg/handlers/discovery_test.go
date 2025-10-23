package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/prasenjit-net/openid-golang/pkg/configstore"
	"github.com/stretchr/testify/assert"
)

func TestDiscovery_WithoutRegistration(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled: false,
		},
	}

	handlers := &Handlers{
		config: cfg,
	}

	// Execute
	err := handlers.Discovery(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response DiscoveryResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify required fields
	assert.Equal(t, "https://example.com", response.Issuer)
	assert.Equal(t, "https://example.com/authorize", response.AuthorizationEndpoint)
	assert.Equal(t, "https://example.com/token", response.TokenEndpoint)
	assert.Equal(t, "https://example.com/.well-known/jwks.json", response.JWKSUri)
	assert.Equal(t, "https://example.com/userinfo", response.UserInfoEndpoint)

	// Verify registration endpoint is NOT included when disabled
	assert.Empty(t, response.RegistrationEndpoint)

	// Verify supported features
	assert.Contains(t, response.ResponseTypesSupported, "code")
	assert.Contains(t, response.ResponseTypesSupported, "id_token")
	assert.Contains(t, response.ResponseTypesSupported, "token id_token")
	assert.Contains(t, response.SubjectTypesSupported, "public")
	assert.Contains(t, response.IDTokenSigningAlgValuesSupported, "RS256")

	// Verify claims include new ones
	assert.Contains(t, response.ClaimsSupported, "auth_time")
	assert.Contains(t, response.ClaimsSupported, "acr")
	assert.Contains(t, response.ClaimsSupported, "amr")
	assert.Contains(t, response.ClaimsSupported, "nonce")

	// Verify PKCE support
	assert.Contains(t, response.CodeChallengeMethodsSupported, "S256")
	assert.Contains(t, response.CodeChallengeMethodsSupported, "plain")

	// Verify advanced features flags
	assert.False(t, response.ClaimsParameterSupported)
	assert.False(t, response.RequestParameterSupported)
	assert.False(t, response.RequestURIParameterSupported)
}

func TestDiscovery_WithRegistrationEnabled(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:              true,
			Endpoint:             "/register",
			ServiceDocumentation: "https://example.com/docs",
			PolicyURI:            "https://example.com/policy",
			TosURI:               "https://example.com/tos",
		},
	}

	handlers := &Handlers{
		config: cfg,
	}

	// Execute
	err := handlers.Discovery(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response DiscoveryResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify registration endpoint IS included when enabled
	assert.Equal(t, "https://example.com/register", response.RegistrationEndpoint)

	// Verify documentation URIs
	assert.Equal(t, "https://example.com/docs", response.ServiceDocumentation)
	assert.Equal(t, "https://example.com/policy", response.OPPolicyURI)
	assert.Equal(t, "https://example.com/tos", response.OPTosURI)
}

func TestDiscovery_CustomRegistrationEndpoint(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled:  true,
			Endpoint: "/oauth2/register",
		},
	}

	handlers := &Handlers{
		config: cfg,
	}

	// Execute
	err := handlers.Discovery(c)

	// Assert
	assert.NoError(t, err)

	var response DiscoveryResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify custom endpoint is used
	assert.Equal(t, "https://example.com/oauth2/register", response.RegistrationEndpoint)
}

func TestDiscovery_JSONStructure(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	cfg := &configstore.ConfigData{
		Issuer: "https://example.com",
		Registration: configstore.RegistrationConfig{
			Enabled: false,
		},
	}

	handlers := &Handlers{
		config: cfg,
	}

	// Execute
	err := handlers.Discovery(c)

	// Assert
	assert.NoError(t, err)

	// Verify it's valid JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &jsonMap)
	assert.NoError(t, err)

	// Verify required fields exist
	assert.NotNil(t, jsonMap["issuer"])
	assert.NotNil(t, jsonMap["authorization_endpoint"])
	assert.NotNil(t, jsonMap["token_endpoint"])
	assert.NotNil(t, jsonMap["jwks_uri"])
	assert.NotNil(t, jsonMap["response_types_supported"])
	assert.NotNil(t, jsonMap["subject_types_supported"])
	assert.NotNil(t, jsonMap["id_token_signing_alg_values_supported"])
}
