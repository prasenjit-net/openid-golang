package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/prasenjit-net/openid-golang/pkg/crypto"
)

// DiscoveryResponse represents OpenID Connect Discovery response
type DiscoveryResponse struct {
	// REQUIRED - Core endpoints
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSUri               string `json:"jwks_uri"`

	// RECOMMENDED - Additional endpoints
	UserInfoEndpoint     string `json:"userinfo_endpoint,omitempty"`
	RegistrationEndpoint string `json:"registration_endpoint,omitempty"`

	// OPTIONAL - Documentation and policies
	ServiceDocumentation string `json:"service_documentation,omitempty"`
	OPPolicyURI          string `json:"op_policy_uri,omitempty"`
	OPTosURI             string `json:"op_tos_uri,omitempty"`

	// REQUIRED - Supported features
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`

	// RECOMMENDED - Additional capabilities
	ScopesSupported                   []string `json:"scopes_supported,omitempty"`
	ResponseModesSupported            []string `json:"response_modes_supported,omitempty"`
	GrantTypesSupported               []string `json:"grant_types_supported,omitempty"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`
	ClaimsSupported                   []string `json:"claims_supported,omitempty"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported,omitempty"`

	// OPTIONAL - Localization support
	UILocalesSupported     []string `json:"ui_locales_supported,omitempty"`
	ClaimsLocalesSupported []string `json:"claims_locales_supported,omitempty"`

	// OPTIONAL - Advanced features
	ClaimsParameterSupported      bool `json:"claims_parameter_supported,omitempty"`
	RequestParameterSupported     bool `json:"request_parameter_supported,omitempty"`
	RequestURIParameterSupported  bool `json:"request_uri_parameter_supported,omitempty"`
	RequireRequestURIRegistration bool `json:"require_request_uri_registration,omitempty"`
}

// Discovery handles the OpenID Connect Discovery endpoint
func (h *Handlers) Discovery(c echo.Context) error {
	baseURL := h.config.Issuer

	response := DiscoveryResponse{
		// REQUIRED - Core endpoints
		Issuer:                baseURL,
		AuthorizationEndpoint: baseURL + "/authorize",
		TokenEndpoint:         baseURL + "/token",
		JWKSUri:               baseURL + "/.well-known/jwks.json",

		// RECOMMENDED - Additional endpoints
		UserInfoEndpoint: baseURL + "/userinfo",

		// REQUIRED - Supported features
		ResponseTypesSupported: []string{
			ResponseTypeCode,
			ResponseTypeIDToken,
			ResponseTypeTokenIDToken,
		},
		SubjectTypesSupported: []string{
			"public",
		},
		IDTokenSigningAlgValuesSupported: []string{
			"RS256",
		},

		// RECOMMENDED - Additional capabilities
		ScopesSupported: []string{
			"openid",
			"profile",
			"email",
		},
		ResponseModesSupported: []string{
			"query",
			"fragment",
		},
		GrantTypesSupported: []string{
			"authorization_code",
			"refresh_token",
		},
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_basic",
			"client_secret_post",
		},
		ClaimsSupported: []string{
			"sub",
			"iss",
			"aud",
			"exp",
			"iat",
			"auth_time",
			"acr",
			"amr",
			"nonce",
			"name",
			"given_name",
			"family_name",
			"email",
			"picture",
		},
		CodeChallengeMethodsSupported: []string{
			"plain",
			"S256",
		},

		// OPTIONAL - Currently not supported
		ClaimsParameterSupported:      false,
		RequestParameterSupported:     false,
		RequestURIParameterSupported:  false,
		RequireRequestURIRegistration: false,
	}

	// Add dynamic registration endpoint if enabled
	if h.config.Registration.Enabled {
		response.RegistrationEndpoint = baseURL + h.config.Registration.Endpoint
	}

	// Add documentation URIs if configured
	if h.config.Registration.ServiceDocumentation != "" {
		response.ServiceDocumentation = h.config.Registration.ServiceDocumentation
	}
	if h.config.Registration.PolicyURI != "" {
		response.OPPolicyURI = h.config.Registration.PolicyURI
	}
	if h.config.Registration.TosURI != "" {
		response.OPTosURI = h.config.Registration.TosURI
	}

	return c.JSON(http.StatusOK, response)
}

// JWKS handles the JWKS endpoint
func (h *Handlers) JWKS(c echo.Context) error {
	publicKey := h.jwtManager.GetPublicKey()
	jwks, err := crypto.PublicKeyToJWKS(publicKey, "default")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to generate JWKS",
		})
	}

	c.Response().Header().Set("Content-Type", "application/json")
	jwksJSON, _ := crypto.MarshalJWKS(jwks)
	return c.Blob(http.StatusOK, "application/json", jwksJSON)
}
