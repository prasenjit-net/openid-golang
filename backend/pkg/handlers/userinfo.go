package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// UserInfoResponse represents the UserInfo response
type UserInfoResponse struct {
	Sub        string `json:"sub"`
	Name       string `json:"name,omitempty"`
	GivenName  string `json:"given_name,omitempty"`
	FamilyName string `json:"family_name,omitempty"`
	Email      string `json:"email,omitempty"`
	Picture    string `json:"picture,omitempty"`
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
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "invalid_token",
			"error_description": "Invalid access token",
		})
	}

	// Check if token is expired
	if token.IsExpired() {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "invalid_token",
			"error_description": "Token expired",
		})
	}

	// Get user information
	user, err := h.storage.GetUserByID(token.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to get user information",
		})
	}

	// Build response based on scopes
	response := UserInfoResponse{
		Sub: user.ID,
	}

	scopes := strings.Split(token.Scope, " ")
	for _, scope := range scopes {
		switch scope {
		case "profile":
			response.Name = user.Name
			response.GivenName = user.GivenName
			response.FamilyName = user.FamilyName
			response.Picture = user.Picture
		case "email":
			response.Email = user.Email
		}
	}

	return c.JSON(http.StatusOK, response)
}
