package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	// ResponseTypeIDToken is the ID token response type (implicit flow)
	ResponseTypeIDToken = "id_token"
	// ResponseTypeTokenIDToken is the access token + ID token response type
	ResponseTypeTokenIDToken = "token id_token"
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
	Code                string     `json:"code" bson:"code"`
	ClientID            string     `json:"client_id" bson:"client_id"`
	UserID              string     `json:"user_id" bson:"user_id"`
	RedirectURI         string     `json:"redirect_uri" bson:"redirect_uri"`
	Scope               string     `json:"scope" bson:"scope"`
	Nonce               string     `json:"nonce,omitempty" bson:"nonce,omitempty"`
	CodeChallenge       string     `json:"code_challenge,omitempty" bson:"code_challenge,omitempty"`
	CodeChallengeMethod string     `json:"code_challenge_method,omitempty" bson:"code_challenge_method,omitempty"`
	Used                bool       `json:"used" bson:"used"`
	UsedAt              *time.Time `json:"used_at,omitempty" bson:"used_at,omitempty"`
	ExpiresAt           time.Time  `json:"expires_at" bson:"expires_at"`
	CreatedAt           time.Time  `json:"created_at" bson:"created_at"`
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

// AuthSession represents an OpenID Connect authorization session
// Stores authorization request parameters during the authentication flow
type AuthSession struct {
	ID                   string                 `json:"id" bson:"_id"`
	ClientID             string                 `json:"client_id" bson:"client_id"`
	RedirectURI          string                 `json:"redirect_uri" bson:"redirect_uri"`
	ResponseType         string                 `json:"response_type" bson:"response_type"`
	Scope                string                 `json:"scope" bson:"scope"`
	State                string                 `json:"state" bson:"state"`
	Nonce                string                 `json:"nonce,omitempty" bson:"nonce,omitempty"`
	CodeChallenge        string                 `json:"code_challenge,omitempty" bson:"code_challenge,omitempty"`
	CodeChallengeMethod  string                 `json:"code_challenge_method,omitempty" bson:"code_challenge_method,omitempty"`
	Prompt               string                 `json:"prompt,omitempty" bson:"prompt,omitempty"`
	MaxAge               int                    `json:"max_age,omitempty" bson:"max_age,omitempty"`
	ACRValues            []string               `json:"acr_values,omitempty" bson:"acr_values,omitempty"`
	Display              string                 `json:"display,omitempty" bson:"display,omitempty"`
	UILocales            []string               `json:"ui_locales,omitempty" bson:"ui_locales,omitempty"`
	ClaimsLocales        []string               `json:"claims_locales,omitempty" bson:"claims_locales,omitempty"`
	Claims               map[string]interface{} `json:"claims,omitempty" bson:"claims,omitempty"`
	UserID               string                 `json:"user_id,omitempty" bson:"user_id,omitempty"`
	AuthTime             *time.Time             `json:"auth_time,omitempty" bson:"auth_time,omitempty"`
	ConsentGiven         bool                   `json:"consent_given" bson:"consent_given"`
	ConsentedScopes      []string               `json:"consented_scopes,omitempty" bson:"consented_scopes,omitempty"`
	AuthenticationMethod string                 `json:"authentication_method,omitempty" bson:"authentication_method,omitempty"`
	ACR                  string                 `json:"acr,omitempty" bson:"acr,omitempty"`
	AMR                  []string               `json:"amr,omitempty" bson:"amr,omitempty"`
	ExpiresAt            time.Time              `json:"expires_at" bson:"expires_at"`
	CreatedAt            time.Time              `json:"created_at" bson:"created_at"`
}

// UserSession represents an authenticated user session with cookies
type UserSession struct {
	ID                   string    `json:"id" bson:"_id"`
	UserID               string    `json:"user_id" bson:"user_id"`
	AuthTime             time.Time `json:"auth_time" bson:"auth_time"`
	AuthenticationMethod string    `json:"authentication_method" bson:"authentication_method"`
	ACR                  string    `json:"acr,omitempty" bson:"acr,omitempty"`
	AMR                  []string  `json:"amr,omitempty" bson:"amr,omitempty"`
	LastActivityAt       time.Time `json:"last_activity_at" bson:"last_activity_at"`
	ExpiresAt            time.Time `json:"expires_at" bson:"expires_at"`
	CreatedAt            time.Time `json:"created_at" bson:"created_at"`
}

// IsAuthenticated checks if the user session is authenticated
func (us *UserSession) IsAuthenticated() bool {
	return us.UserID != "" && time.Now().Before(us.ExpiresAt)
}

// IsAuthTimeFresh checks if authentication time is within max_age seconds
func (us *UserSession) IsAuthTimeFresh(maxAge int) bool {
	if maxAge == 0 {
		return false
	}
	elapsed := time.Since(us.AuthTime).Seconds()
	return elapsed <= float64(maxAge)
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
		ResponseTypes: []string{ResponseTypeIDToken, ResponseTypeTokenIDToken},
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
