package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	Name         string    `json:"name"`
	GivenName    string    `json:"given_name,omitempty"`
	FamilyName   string    `json:"family_name,omitempty"`
	Picture      string    `json:"picture,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// HasRole checks if the user has the specified role
func (u *User) HasRole(role UserRole) bool {
	return u.Role == role
}

// Client represents an OAuth2/OIDC client
type Client struct {
	ID            string    `json:"client_id"`
	Secret        string    `json:"client_secret,omitempty"`
	Name          string    `json:"name"`
	RedirectURIs  []string  `json:"redirect_uris"`
	GrantTypes    []string  `json:"grant_types"`
	ResponseTypes []string  `json:"response_types"`
	Scope         string    `json:"scope"`
	CreatedAt     time.Time `json:"created_at"`
}

// AuthorizationCode represents an authorization code
type AuthorizationCode struct {
	Code                string    `json:"code"`
	ClientID            string    `json:"client_id"`
	UserID              string    `json:"user_id"`
	RedirectURI         string    `json:"redirect_uri"`
	Scope               string    `json:"scope"`
	Nonce               string    `json:"nonce,omitempty"`
	CodeChallenge       string    `json:"code_challenge,omitempty"`
	CodeChallengeMethod string    `json:"code_challenge_method,omitempty"`
	ExpiresAt           time.Time `json:"expires_at"`
	CreatedAt           time.Time `json:"created_at"`
}

// Token represents an access or refresh token
type Token struct {
	ID           string    `json:"id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ClientID     string    `json:"client_id"`
	UserID       string    `json:"user_id"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// NewUser creates a new user with generated ID
func NewUser(username, email, passwordHash string, role UserRole) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewAdminUser creates a new user with admin role
func NewAdminUser(username, email, passwordHash string) *User {
	return NewUser(username, email, passwordHash, RoleAdmin)
}

// NewRegularUser creates a new user with regular user role
func NewRegularUser(username, email, passwordHash string) *User {
	return NewUser(username, email, passwordHash, RoleUser)
}

// NewClient creates a new client with generated ID and secret
func NewClient(name string, redirectURIs []string) *Client {
	return &Client{
		ID:            uuid.New().String(),
		Secret:        uuid.New().String(),
		Name:          name,
		RedirectURIs:  redirectURIs,
		GrantTypes:    []string{"authorization_code", "refresh_token"},
		ResponseTypes: []string{"code"},
		Scope:         "openid profile email",
		CreatedAt:     time.Now(),
	}
}

// NewAdminUIClient creates a client for the admin UI using implicit flow
func NewAdminUIClient(issuerURL string) *Client {
	return &Client{
		ID:            "admin-ui",
		Secret:        "", // No secret needed for implicit flow
		Name:          "Admin UI",
		RedirectURIs:  []string{issuerURL + "/admin/callback"},
		GrantTypes:    []string{"implicit"},
		ResponseTypes: []string{"id_token", "token id_token"},
		Scope:         "openid profile email",
		CreatedAt:     time.Now(),
	}
}

// NewAuthorizationCode creates a new authorization code
func NewAuthorizationCode(clientID, userID, redirectURI, scope string) *AuthorizationCode {
	return &AuthorizationCode{
		Code:        uuid.New().String(),
		ClientID:    clientID,
		UserID:      userID,
		RedirectURI: redirectURI,
		Scope:       scope,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}
}

// NewToken creates a new token
func NewToken(clientID, userID, scope string, expiryMinutes int) *Token {
	return &Token{
		ID:           uuid.New().String(),
		AccessToken:  uuid.New().String(),
		RefreshToken: uuid.New().String(),
		TokenType:    "Bearer",
		ClientID:     clientID,
		UserID:       userID,
		Scope:        scope,
		ExpiresAt:    time.Now().Add(time.Duration(expiryMinutes) * time.Minute),
		CreatedAt:    time.Now(),
	}
}

// IsExpired checks if the authorization code is expired
func (ac *AuthorizationCode) IsExpired() bool {
	return time.Now().After(ac.ExpiresAt)
}

// IsExpired checks if the token is expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
