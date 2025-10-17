package handlers

import (
	"net/http"
	"strings"
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
func (h *Handlers) UserInfo(w http.ResponseWriter, r *http.Request) {
	// Extract access token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, http.StatusUnauthorized, "invalid_token", "Missing authorization header")
		return
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		writeError(w, http.StatusUnauthorized, "invalid_token", "Invalid authorization header format")
		return
	}

	accessToken := parts[1]

	// Validate access token
	token, err := h.storage.GetTokenByAccessToken(accessToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_token", "Invalid access token")
		return
	}

	// Check if token is expired
	if token.IsExpired() {
		writeError(w, http.StatusUnauthorized, "invalid_token", "Token expired")
		return
	}

	// Get user information
	user, err := h.storage.GetUserByID(token.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get user information")
		return
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

	writeJSON(w, http.StatusOK, response)
}
