package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// RevokeRequest represents a token revocation request per RFC 7009
type RevokeRequest struct {
	Token         string
	TokenTypeHint string // "access_token" or "refresh_token"
}

// Revoke handles token revocation (POST /revoke) per RFC 7009
// This endpoint allows clients to notify the authorization server that a token is no longer needed
func (h *Handlers) Revoke(c echo.Context) error {
	req := &RevokeRequest{
		Token:         c.FormValue("token"),
		TokenTypeHint: c.FormValue("token_type_hint"),
	}

	// Token parameter is REQUIRED
	if req.Token == "" {
		return jsonError(c, http.StatusBadRequest, ErrorInvalidRequest, "token parameter is required")
	}

	// Get client credentials
	clientID := c.FormValue("client_id")
	clientSecret := c.FormValue("client_secret")

	// Try Basic Auth if not in form parameters
	if clientID == "" || clientSecret == "" {
		clientID, clientSecret, _ = parseBasicAuth(c.Request().Header.Get("Authorization"))
	}

	// Validate client credentials
	client, err := h.storage.ValidateClient(clientID, clientSecret)
	if err != nil || client == nil {
		// RFC 7009 §2.2.1: The authorization server validates the client credentials
		// If invalid, return invalid_client error
		return ErrorInvalidClientAuth(c, "Invalid client credentials")
	}

	// Attempt to revoke the token
	// Note: RFC 7009 §2.2 states that if the token doesn't exist or belongs to another client,
	// the request should still succeed (return 200) to prevent token scanning attacks
	_ = h.revokeToken(req.Token, req.TokenTypeHint, client.ID)

	// RFC 7009 §2.2: The authorization server responds with HTTP status code 200
	// if the token has been revoked successfully or if the client submitted an invalid token
	return c.NoContent(http.StatusOK)
}

// revokeToken attempts to revoke a token
func (h *Handlers) revokeToken(token, tokenTypeHint, clientID string) error {
	// Try as refresh token first if hint provided or no hint given
	if tokenTypeHint == TokenTypeHintRefreshToken || tokenTypeHint == "" {
		if err := h.revokeRefreshToken(token, clientID); err == nil {
			return nil
		}
	}

	// Try as access token
	if tokenTypeHint == TokenTypeHintAccessToken || tokenTypeHint == "" {
		if err := h.revokeAccessToken(token, clientID); err == nil {
			return nil
		}
	}

	// Token not found or already revoked - this is fine per RFC 7009
	return nil
}

// revokeRefreshToken revokes a refresh token and all associated access tokens
func (h *Handlers) revokeRefreshToken(refreshToken, clientID string) error {
	// Get token by refresh token
	token, err := h.storage.GetTokenByRefreshToken(refreshToken)
	if err != nil || token == nil {
		return fmt.Errorf("token not found")
	}

	// Verify token belongs to the requesting client
	if token.ClientID != clientID {
		return fmt.Errorf("token does not belong to client")
	}

	// RFC 7009 §2: If the particular token is a refresh token and the authorization server
	// supports the revocation of access tokens, then the authorization server SHOULD also
	// invalidate all access tokens based on the same authorization grant
	if token.AuthorizationCodeID != "" {
		_ = h.storage.RevokeTokensByAuthCode(token.AuthorizationCodeID)
	}

	// Delete the token using its ID
	return h.storage.DeleteToken(token.ID)
}

// revokeAccessToken revokes an access token
func (h *Handlers) revokeAccessToken(accessToken, clientID string) error {
	// Get token by access token
	token, err := h.storage.GetTokenByAccessToken(accessToken)
	if err != nil || token == nil {
		return fmt.Errorf("token not found")
	}

	// Verify token belongs to the requesting client
	if token.ClientID != clientID {
		return fmt.Errorf("token does not belong to client")
	}

	// Delete the token using its ID
	return h.storage.DeleteToken(token.ID)
}
