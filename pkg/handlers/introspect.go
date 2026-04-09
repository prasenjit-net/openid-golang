package handlers

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/prasenjit-net/openid-golang/pkg/models"
)

// IntrospectRequest represents a token introspection request per RFC 7662
type IntrospectRequest struct {
	Token         string
	TokenTypeHint string // "access_token" or "refresh_token"
}

// IntrospectResponse represents a token introspection response per RFC 7662
type IntrospectResponse struct {
	// REQUIRED - Token validity
	Active bool `json:"active"`

	// OPTIONAL - Token metadata (only present if active=true)
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Username  string `json:"username,omitempty"` // Subject's human-readable identifier
	TokenType string `json:"token_type,omitempty"`
	Exp       int64  `json:"exp,omitempty"` // Expiration time (seconds since epoch)
	Iat       int64  `json:"iat,omitempty"` // Issued at time (seconds since epoch)
	Nbf       int64  `json:"nbf,omitempty"` // Not before time (seconds since epoch)
	Sub       string `json:"sub,omitempty"` // Subject identifier
	Aud       string `json:"aud,omitempty"` // Audience
	Iss       string `json:"iss,omitempty"` // Issuer
	Jti       string `json:"jti,omitempty"` // JWT ID
}

// Introspect handles token introspection (POST /introspect) per RFC 7662
// This endpoint allows resource servers to query the authorization server
// about the active state and metadata of a token
func (h *Handlers) Introspect(c echo.Context) error {
	req := &IntrospectRequest{
		Token:         c.FormValue("token"),
		TokenTypeHint: c.FormValue("token_type_hint"),
	}

	// Token parameter is REQUIRED
	if req.Token == "" {
		return jsonError(c, http.StatusBadRequest, ErrorInvalidRequest, "token parameter is required")
	}

	// Get client credentials - RFC 7662 §2.1: The protected resource calls the introspection
	// endpoint using an HTTP POST request with parameters in "application/x-www-form-urlencoded" format.
	// Client authentication is REQUIRED per RFC 7662 §2.1
	clientID := c.FormValue("client_id")
	clientSecret := c.FormValue("client_secret")

	// Try Basic Auth if not in form parameters
	if clientID == "" || clientSecret == "" {
		clientID, clientSecret, _ = parseBasicAuth(c.Request().Header.Get("Authorization"))
	}

	// Validate client credentials - REQUIRED per RFC 7662 §2.1
	client, err := h.storage.ValidateClient(clientID, clientSecret)
	if err != nil || client == nil {
		return ErrorInvalidClientAuth(c, "Invalid client credentials")
	}

	// Introspect the token
	response := h.introspectToken(req.Token, req.TokenTypeHint)

	// RFC 7662 §2.2: The authorization server responds with a JSON object
	return c.JSON(http.StatusOK, response)
}

// introspectToken performs the actual token introspection
func (h *Handlers) introspectToken(tokenString, tokenTypeHint string) *IntrospectResponse {
	// Try to find token in storage first
	// Check as access token or refresh token based on hint
	var token *models.Token
	var err error

	if tokenTypeHint == TokenTypeHintAccessToken || tokenTypeHint == "" {
		token, err = h.storage.GetTokenByAccessToken(tokenString)
		if err == nil && token != nil {
			return h.buildIntrospectResponse(token, tokenString)
		}
	}

	if tokenTypeHint == TokenTypeHintRefreshToken || tokenTypeHint == "" {
		token, err = h.storage.GetTokenByRefreshToken(tokenString)
		if err == nil && token != nil {
			return h.buildIntrospectResponse(token, tokenString)
		}
	}

	// Token not found in storage, try to validate as JWT
	// This is useful for clients checking JWTs they received
	return h.introspectJWT(tokenString)
}

// buildIntrospectResponse builds an introspection response from a stored token
func (h *Handlers) buildIntrospectResponse(token *models.Token, tokenString string) *IntrospectResponse {
	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return &IntrospectResponse{Active: false}
	}

	// Get user info for username
	username := ""
	if token.UserID != "" {
		user, err := h.storage.GetUserByID(token.UserID)
		if err == nil && user != nil {
			username = user.Username
		}
	}

	// Build response
	return &IntrospectResponse{
		Active:    true,
		Scope:     token.Scope,
		ClientID:  token.ClientID,
		Username:  username,
		TokenType: token.TokenType,
		Exp:       token.ExpiresAt.Unix(),
		Iat:       token.CreatedAt.Unix(),
		Sub:       token.UserID,
		Iss:       h.config.Issuer,
	}
}

// introspectJWT attempts to introspect a JWT token
func (h *Handlers) introspectJWT(tokenString string) *IntrospectResponse {
	// Parse and validate JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if token.Method.Alg() != "RS256" {
			return nil, jwt.ErrSignatureInvalid
		}
		return h.jwtManager.GetPublicKey(), nil
	})

	if err != nil || !token.Valid {
		// Invalid or expired JWT
		return &IntrospectResponse{Active: false}
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &IntrospectResponse{Active: false}
	}

	// Build response from JWT claims
	response := &IntrospectResponse{
		Active:    true,
		TokenType: "Bearer",
	}

	// Extract standard claims
	if exp, ok := claims["exp"].(float64); ok {
		response.Exp = int64(exp)
		// Check if expired
		if time.Now().Unix() > int64(exp) {
			return &IntrospectResponse{Active: false}
		}
	}

	if iat, ok := claims["iat"].(float64); ok {
		response.Iat = int64(iat)
	}

	if nbf, ok := claims["nbf"].(float64); ok {
		response.Nbf = int64(nbf)
		// Check if not yet valid
		if time.Now().Unix() < int64(nbf) {
			return &IntrospectResponse{Active: false}
		}
	}

	if sub, ok := claims["sub"].(string); ok {
		response.Sub = sub
		// Try to get username from user ID
		if user, err := h.storage.GetUserByID(sub); err == nil && user != nil {
			response.Username = user.Username
		}
	}

	if iss, ok := claims["iss"].(string); ok {
		response.Iss = iss
	}

	if aud, ok := claims["aud"].(string); ok {
		response.Aud = aud
	} else if audList, ok := claims["aud"].([]interface{}); ok && len(audList) > 0 {
		if audStr, ok := audList[0].(string); ok {
			response.Aud = audStr
		}
	}

	if jti, ok := claims["jti"].(string); ok {
		response.Jti = jti
	}

	if scope, ok := claims["scope"].(string); ok {
		response.Scope = scope
	}

	if clientID, ok := claims["client_id"].(string); ok {
		response.ClientID = clientID
	} else if azp, ok := claims["azp"].(string); ok {
		// Authorized party (azp) can be used as client_id
		response.ClientID = azp
	}

	return response
}
