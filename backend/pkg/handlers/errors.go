package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
)

// OAuth 2.0 Error Codes (RFC 6749 Section 4.1.2.1)
const (
	ErrorInvalidRequest          = "invalid_request"
	ErrorUnauthorizedClient      = "unauthorized_client"
	ErrorAccessDenied            = "access_denied"
	ErrorUnsupportedResponseType = "unsupported_response_type"
	ErrorInvalidScope            = "invalid_scope"
	ErrorServerError             = "server_error"
	ErrorTemporarilyUnavailable  = "temporarily_unavailable"
)

// OpenID Connect Specific Error Codes (OIDC Core Section 3.1.2.6)
const (
	ErrorInteractionRequired      = "interaction_required"
	ErrorLoginRequired            = "login_required"
	ErrorAccountSelectionRequired = "account_selection_required"
	ErrorConsentRequired          = "consent_required"
	ErrorInvalidRequestURI        = "invalid_request_uri"
	ErrorInvalidRequestObject     = "invalid_request_object"
	ErrorRequestNotSupported      = "request_not_supported"
	ErrorRequestURINotSupported   = "request_uri_not_supported"
	ErrorRegistrationNotSupported = "registration_not_supported"
)

// Token Endpoint Error Codes (RFC 6749 Section 5.2)
const (
	ErrorInvalidClient        = "invalid_client"
	ErrorInvalidGrant         = "invalid_grant"
	ErrorUnsupportedGrantType = "unsupported_grant_type"
)

// UserInfo Endpoint Error Codes (OIDC Core Section 5.3.3)
const (
	ErrorInvalidToken      = "invalid_token"
	ErrorInsufficientScope = "insufficient_scope"
)

// ErrorResponse represents an OAuth 2.0 error response
type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

// redirectWithError redirects to the client's redirect_uri with error parameters
// Used for authorization endpoint errors
func redirectWithError(c echo.Context, redirectURI, errorCode, errorDescription, state string, useFragment bool) error {
	if redirectURI == "" {
		// If no valid redirect_uri, return JSON error instead
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:            errorCode,
			ErrorDescription: errorDescription,
		})
	}

	u, err := url.Parse(redirectURI)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:            ErrorInvalidRequest,
			ErrorDescription: "Invalid redirect_uri",
		})
	}

	// Build error parameters
	params := url.Values{}
	params.Set("error", errorCode)
	if errorDescription != "" {
		params.Set("error_description", errorDescription)
	}
	if state != "" {
		params.Set("state", state)
	}

	if useFragment {
		// For implicit/hybrid flows - use fragment
		u.Fragment = params.Encode()
	} else {
		// For authorization code flow - use query
		q := u.Query()
		for k, v := range params {
			q.Set(k, v[0])
		}
		u.RawQuery = q.Encode()
	}

	return c.Redirect(http.StatusFound, u.String())
}

// jsonError returns a JSON error response
// Used for token endpoint and userinfo endpoint errors
func jsonError(c echo.Context, statusCode int, errorCode, errorDescription string) error {
	return c.JSON(statusCode, ErrorResponse{
		Error:            errorCode,
		ErrorDescription: errorDescription,
	})
}

// determineErrorRedirectMethod determines whether to use query or fragment for error redirect
// based on response_type
func determineErrorRedirectMethod(responseType string) bool {
	// Use fragment for implicit and hybrid flows
	// These response types return tokens/id_tokens in fragment
	switch responseType {
	case ResponseTypeIDToken,
		ResponseTypeToken,
		ResponseTypeTokenIDToken,
		ResponseTypeCodeIDToken,
		ResponseTypeCodeToken,
		ResponseTypeCodeTokenIDToken:
		return true // use fragment
	default:
		return false // use query (code flow)
	}
}

// authorizationError is a convenience function for authorization endpoint errors
// Automatically determines whether to use query or fragment
func authorizationError(c echo.Context, redirectURI, responseType, errorCode, errorDescription, state string) error {
	useFragment := determineErrorRedirectMethod(responseType)
	return redirectWithError(c, redirectURI, errorCode, errorDescription, state, useFragment)
}

// Helper functions for common error scenarios

// ErrorInvalidClientAuth returns an unauthorized client error for token endpoint
func ErrorInvalidClientAuth(c echo.Context, description string) error {
	if description == "" {
		description = "Client authentication failed"
	}
	return jsonError(c, http.StatusUnauthorized, ErrorInvalidClient, description)
}

// ErrorInvalidAuthorizationCode returns an invalid grant error
func ErrorInvalidAuthorizationCode(c echo.Context, description string) error {
	if description == "" {
		description = "Invalid or expired authorization code"
	}
	return jsonError(c, http.StatusBadRequest, ErrorInvalidGrant, description)
}

// ErrorInvalidTokenRequest returns an invalid request error for token endpoint
func ErrorInvalidTokenRequest(c echo.Context, description string) error {
	if description == "" {
		description = "Invalid token request"
	}
	return jsonError(c, http.StatusBadRequest, ErrorInvalidRequest, description)
}

// ErrorServerErrorJSON returns a server error as JSON
func ErrorServerErrorJSON(c echo.Context, description string) error {
	if description == "" {
		description = "Internal server error"
	}
	return jsonError(c, http.StatusInternalServerError, ErrorServerError, description)
}

// ErrorInvalidAccessToken returns an invalid token error for userinfo endpoint
func ErrorInvalidAccessToken(c echo.Context, description string) error {
	if description == "" {
		description = "Invalid or expired access token"
	}
	c.Response().Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer error="%s", error_description="%s"`, ErrorInvalidToken, description))
	return jsonError(c, http.StatusUnauthorized, ErrorInvalidToken, description)
}
