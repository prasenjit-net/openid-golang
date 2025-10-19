package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/prasenjit-net/openid-golang/internal/crypto"
)

// DiscoveryResponse represents OpenID Connect Discovery response
type DiscoveryResponse struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JWKSUri                           string   `json:"jwks_uri"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ResponseModesSupported            []string `json:"response_modes_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
}

// Discovery handles the OpenID Connect Discovery endpoint
func (h *Handlers) Discovery(c echo.Context) error {
	baseURL := h.config.Issuer

	response := DiscoveryResponse{
		Issuer:                baseURL,
		AuthorizationEndpoint: baseURL + "/authorize",
		TokenEndpoint:         baseURL + "/token",
		UserInfoEndpoint:      baseURL + "/userinfo",
		JWKSUri:               baseURL + "/.well-known/jwks.json",
		ScopesSupported: []string{
			"openid",
			"profile",
			"email",
		},
		ResponseTypesSupported: []string{
			ResponseTypeCode,
			ResponseTypeIDToken,
			ResponseTypeTokenIDToken,
		},
		ResponseModesSupported: []string{
			"query",
			"fragment",
		},
		GrantTypesSupported: []string{
			"authorization_code",
			"refresh_token",
		},
		SubjectTypesSupported: []string{
			"public",
		},
		IDTokenSigningAlgValuesSupported: []string{
			"RS256",
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
