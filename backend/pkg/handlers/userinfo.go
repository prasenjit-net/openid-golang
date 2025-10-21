package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/prasenjit-net/openid-golang/pkg/models"
)

// UserInfoResponse represents the UserInfo response (OIDC Core 1.0 Section 5.1)
type UserInfoResponse struct {
	Sub string `json:"sub"` // Subject - required

	// Profile scope claims (OIDC Core 1.0 Section 5.4)
	Name       string `json:"name,omitempty"`
	GivenName  string `json:"given_name,omitempty"`
	FamilyName string `json:"family_name,omitempty"`
	Picture    string `json:"picture,omitempty"`
	UpdatedAt  int64  `json:"updated_at,omitempty"` // Unix timestamp

	// Email scope claims
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`

	// TODO: Add address scope claims when User model is extended
	// TODO: Add phone scope claims when User model is extended
}

// UserInfo handles the UserInfo endpoint (GET/POST /userinfo)
func (h *Handlers) UserInfo(c echo.Context) error {
	// Extract access token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "invalid_token",
			"error_description": "Missing authorization header",
		})
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "invalid_token",
			"error_description": "Invalid authorization header format",
		})
	}

	accessToken := parts[1]

	// Validate access token
	token, err := h.storage.GetTokenByAccessToken(accessToken)
	if err != nil || token == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "invalid_token",
			"error_description": "Invalid or expired access token",
		})
	}

	// Check if token is expired
	if token.IsExpired() {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "invalid_token",
			"error_description": "Access token has expired",
		})
	}

	// Verify token has openid scope (required for UserInfo endpoint)
	if !h.hasScope(token.Scope, "openid") {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error":             "insufficient_scope",
			"error_description": "Access token does not have openid scope",
		})
	}

	// Get user information
	user, err := h.storage.GetUserByID(token.UserID)
	if err != nil || user == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to retrieve user information",
		})
	}

	// Build response based on requested scopes
	response := h.buildUserInfoResponse(user, token.Scope)

	return c.JSON(http.StatusOK, response)
}

// buildUserInfoResponse constructs the UserInfo response based on granted scopes
func (h *Handlers) buildUserInfoResponse(user *models.User, scopeString string) UserInfoResponse {
	response := UserInfoResponse{
		Sub: user.ID, // Subject identifier is always included
	}

	// Parse scopes into a map for quick lookup
	scopeMap := make(map[string]bool)
	if scopeString != "" {
		scopes := strings.Split(scopeString, " ")
		for _, scope := range scopes {
			scopeMap[strings.TrimSpace(scope)] = true
		}
	}

	// Profile scope: name, family_name, given_name, picture, updated_at, etc.
	// (OIDC Core 1.0 Section 5.4)
	if scopeMap["profile"] {
		if user.Name != "" {
			response.Name = user.Name
		}
		if user.GivenName != "" {
			response.GivenName = user.GivenName
		}
		if user.FamilyName != "" {
			response.FamilyName = user.FamilyName
		}
		if user.Picture != "" {
			response.Picture = user.Picture
		}
		// Include updated_at timestamp
		if !user.UpdatedAt.IsZero() {
			response.UpdatedAt = user.UpdatedAt.Unix()
		}
	}

	// Email scope: email, email_verified
	if scopeMap["email"] {
		if user.Email != "" {
			response.Email = user.Email
			// TODO: Implement proper email verification system
			// For now, mark as verified since we don't have a verification system
			response.EmailVerified = true
		}
	}

	// TODO: Address scope - requires extending User model with address fields
	// if scopeMap["address"] {
	//     response.Address = user.Address
	// }

	// TODO: Phone scope - requires extending User model with phone fields
	// if scopeMap["phone"] {
	//     response.PhoneNumber = user.PhoneNumber
	//     response.PhoneNumberVerified = user.PhoneNumberVerified
	// }

	return response
}

// hasScope checks if a scope string contains a specific scope
func (h *Handlers) hasScope(scopeString, targetScope string) bool {
	if scopeString == "" {
		return false
	}
	scopes := strings.Split(scopeString, " ")
	for _, scope := range scopes {
		if strings.TrimSpace(scope) == targetScope {
			return true
		}
	}
	return false
}
