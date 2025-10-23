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

// Client represents an OAuth2/OIDC client with full OIDC Dynamic Registration support
type Client struct {
	// Core OAuth 2.0 fields
	ID              string   `json:"client_id" bson:"_id"`
	Secret          string   `json:"client_secret,omitempty" bson:"client_secret,omitempty"`
	SecretExpiresAt int64    `json:"client_secret_expires_at" bson:"client_secret_expires_at"` // Unix timestamp, 0 = never expires
	RedirectURIs    []string `json:"redirect_uris" bson:"redirect_uris"`
	GrantTypes      []string `json:"grant_types,omitempty" bson:"grant_types,omitempty"`
	ResponseTypes   []string `json:"response_types,omitempty" bson:"response_types,omitempty"`
	Scope           string   `json:"scope,omitempty" bson:"scope,omitempty"`

	// OIDC Dynamic Registration fields
	ApplicationType string   `json:"application_type,omitempty" bson:"application_type,omitempty"` // "web" or "native"
	Contacts        []string `json:"contacts,omitempty" bson:"contacts,omitempty"`                 // Email addresses
	ClientName      string   `json:"client_name,omitempty" bson:"client_name,omitempty"`           // Human-readable name

	// Localized metadata (stored as JSON internally, exposed via special handling)
	ClientNameLocalized map[string]string `json:"-" bson:"client_name_localized,omitempty"` // e.g., "en" -> "My App"

	// Client URIs
	LogoURI            string            `json:"logo_uri,omitempty" bson:"logo_uri,omitempty"`
	LogoURILocalized   map[string]string `json:"-" bson:"logo_uri_localized,omitempty"`
	ClientURI          string            `json:"client_uri,omitempty" bson:"client_uri,omitempty"`
	ClientURILocalized map[string]string `json:"-" bson:"client_uri_localized,omitempty"`
	PolicyURI          string            `json:"policy_uri,omitempty" bson:"policy_uri,omitempty"`
	PolicyURILocalized map[string]string `json:"-" bson:"policy_uri_localized,omitempty"`
	TosURI             string            `json:"tos_uri,omitempty" bson:"tos_uri,omitempty"`
	TosURILocalized    map[string]string `json:"-" bson:"tos_uri_localized,omitempty"`

	// JWK/Signing fields
	JWKSURI             string `json:"jwks_uri,omitempty" bson:"jwks_uri,omitempty"`
	JWKS                string `json:"jwks,omitempty" bson:"jwks,omitempty"` // JWK Set as JSON string
	SectorIdentifierURI string `json:"sector_identifier_uri,omitempty" bson:"sector_identifier_uri,omitempty"`
	SubjectType         string `json:"subject_type,omitempty" bson:"subject_type,omitempty"` // "public" or "pairwise"

	// ID Token signing/encryption preferences
	IDTokenSignedResponseAlg    string `json:"id_token_signed_response_alg,omitempty" bson:"id_token_signed_response_alg,omitempty"`
	IDTokenEncryptedResponseAlg string `json:"id_token_encrypted_response_alg,omitempty" bson:"id_token_encrypted_response_alg,omitempty"`
	IDTokenEncryptedResponseEnc string `json:"id_token_encrypted_response_enc,omitempty" bson:"id_token_encrypted_response_enc,omitempty"`

	// UserInfo signing/encryption preferences
	UserInfoSignedResponseAlg    string `json:"userinfo_signed_response_alg,omitempty" bson:"userinfo_signed_response_alg,omitempty"`
	UserInfoEncryptedResponseAlg string `json:"userinfo_encrypted_response_alg,omitempty" bson:"userinfo_encrypted_response_alg,omitempty"`
	UserInfoEncryptedResponseEnc string `json:"userinfo_encrypted_response_enc,omitempty" bson:"userinfo_encrypted_response_enc,omitempty"`

	// Request Object signing/encryption preferences
	RequestObjectSigningAlg    string `json:"request_object_signing_alg,omitempty" bson:"request_object_signing_alg,omitempty"`
	RequestObjectEncryptionAlg string `json:"request_object_encryption_alg,omitempty" bson:"request_object_encryption_alg,omitempty"`
	RequestObjectEncryptionEnc string `json:"request_object_encryption_enc,omitempty" bson:"request_object_encryption_enc,omitempty"`

	// Token Endpoint Authentication
	TokenEndpointAuthMethod     string `json:"token_endpoint_auth_method,omitempty" bson:"token_endpoint_auth_method,omitempty"`
	TokenEndpointAuthSigningAlg string `json:"token_endpoint_auth_signing_alg,omitempty" bson:"token_endpoint_auth_signing_alg,omitempty"`

	// Authentication requirements
	DefaultMaxAge    int      `json:"default_max_age,omitempty" bson:"default_max_age,omitempty"`
	RequireAuthTime  bool     `json:"require_auth_time,omitempty" bson:"require_auth_time,omitempty"`
	DefaultACRValues []string `json:"default_acr_values,omitempty" bson:"default_acr_values,omitempty"`

	// Advanced features
	InitiateLoginURI string   `json:"initiate_login_uri,omitempty" bson:"initiate_login_uri,omitempty"`
	RequestURIs      []string `json:"request_uris,omitempty" bson:"request_uris,omitempty"`

	// Software Statement (JWT containing client metadata claims)
	SoftwareID        string `json:"software_id,omitempty" bson:"software_id,omitempty"`
	SoftwareVersion   string `json:"software_version,omitempty" bson:"software_version,omitempty"`
	SoftwareStatement string `json:"software_statement,omitempty" bson:"software_statement,omitempty"` // JWT

	// Registration metadata (internal use)
	RegistrationAccessToken string    `json:"-" bson:"registration_access_token,omitempty"`                       // Never exposed in responses (except registration response)
	ClientIDIssuedAt        int64     `json:"client_id_issued_at,omitempty" bson:"client_id_issued_at,omitempty"` // Unix timestamp
	CreatedAt               time.Time `json:"-" bson:"created_at"`
	UpdatedAt               time.Time `json:"-" bson:"updated_at"`

	// Legacy compatibility field (maps to ClientName)
	Name string `json:"name,omitempty" bson:"-"` // Deprecated: use ClientName
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
	ID                  string    `json:"id"`
	AccessToken         string    `json:"access_token"`
	RefreshToken        string    `json:"refresh_token,omitempty"`
	TokenType           string    `json:"token_type"`
	ClientID            string    `json:"client_id"`
	UserID              string    `json:"user_id"`
	Scope               string    `json:"scope"`
	AuthorizationCodeID string    `json:"authorization_code_id,omitempty" bson:"authorization_code_id,omitempty"`
	ExpiresAt           time.Time `json:"expires_at"`
	CreatedAt           time.Time `json:"created_at"`
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

// Consent represents a user's consent to a client accessing specific scopes
type Consent struct {
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"user_id" bson:"user_id"`
	ClientID  string    `json:"client_id" bson:"client_id"`
	Scopes    []string  `json:"scopes" bson:"scopes"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// HasScope checks if the consent includes a specific scope
func (c *Consent) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// HasAllScopes checks if the consent includes all requested scopes
func (c *Consent) HasAllScopes(requestedScopes []string) bool {
	for _, requested := range requestedScopes {
		if !c.HasScope(requested) {
			return false
		}
	}
	return true
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
	now := time.Now()
	return &Client{
		ID:                       uuid.New().String(),
		Secret:                   uuid.New().String(),
		SecretExpiresAt:          0, // Never expires
		ClientName:               name,
		Name:                     name, // Legacy compatibility
		RedirectURIs:             redirectURIs,
		GrantTypes:               []string{"authorization_code", "refresh_token"},
		ResponseTypes:            []string{"code"},
		Scope:                    "openid profile email",
		ApplicationType:          "web",
		SubjectType:              "public",
		TokenEndpointAuthMethod:  "client_secret_basic",
		IDTokenSignedResponseAlg: "RS256",
		ClientIDIssuedAt:         now.Unix(),
		CreatedAt:                now,
		UpdatedAt:                now,
	}
}

// NewAdminUIClient creates a client for the admin UI using implicit flow
func NewAdminUIClient(issuerURL string) *Client {
	now := time.Now()
	return &Client{
		ID:                       "admin-ui",
		Secret:                   "", // No secret needed for implicit flow
		SecretExpiresAt:          0,  // N/A for public clients
		ClientName:               "Admin UI",
		Name:                     "Admin UI", // Legacy compatibility
		RedirectURIs:             []string{issuerURL + "/admin/callback"},
		GrantTypes:               []string{"implicit"},
		ResponseTypes:            []string{ResponseTypeIDToken, ResponseTypeTokenIDToken},
		Scope:                    "openid profile email",
		ApplicationType:          "web",
		SubjectType:              "public",
		TokenEndpointAuthMethod:  "none", // Public client
		IDTokenSignedResponseAlg: "RS256",
		ClientIDIssuedAt:         now.Unix(),
		CreatedAt:                now,
		UpdatedAt:                now,
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

// NewConsent creates a new consent record
func NewConsent(userID, clientID string, scopes []string) *Consent {
	now := time.Now()
	return &Consent{
		ID:        uuid.New().String(),
		UserID:    userID,
		ClientID:  clientID,
		Scopes:    scopes,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Client helper methods

// IsPublicClient returns true if the client doesn't have a secret (public client)
func (c *Client) IsPublicClient() bool {
	return c.Secret == ""
}

// IsConfidentialClient returns true if the client has a secret (confidential client)
func (c *Client) IsConfidentialClient() bool {
	return c.Secret != ""
}

// HasGrantType checks if the client supports a specific grant type
func (c *Client) HasGrantType(grantType string) bool {
	for _, gt := range c.GrantTypes {
		if gt == grantType {
			return true
		}
	}
	return false
}

// HasResponseType checks if the client supports a specific response type
func (c *Client) HasResponseType(responseType string) bool {
	for _, rt := range c.ResponseTypes {
		if rt == responseType {
			return true
		}
	}
	return false
}

// ValidateRedirectURI checks if the provided redirect URI is registered
func (c *Client) ValidateRedirectURI(redirectURI string) bool {
	for _, uri := range c.RedirectURIs {
		if uri == redirectURI {
			return true
		}
	}
	return false
}

// GetDisplayName returns the client name or falls back to client ID
func (c *Client) GetDisplayName() string {
	if c.ClientName != "" {
		return c.ClientName
	}
	if c.Name != "" {
		return c.Name
	}
	return c.ID
}

// ClientRegistrationRequest represents a dynamic client registration request
// As defined in RFC 7591 (OAuth 2.0 Dynamic Client Registration Protocol)
// and OpenID Connect Dynamic Client Registration 1.0
type ClientRegistrationRequest struct {
	// REQUIRED fields
	RedirectURIs []string `json:"redirect_uris" validate:"required,min=1"`

	// OPTIONAL OAuth 2.0 fields
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	ClientName              string   `json:"client_name,omitempty"`
	ClientURI               string   `json:"client_uri,omitempty"`
	LogoURI                 string   `json:"logo_uri,omitempty"`
	Scope                   string   `json:"scope,omitempty"`
	Contacts                []string `json:"contacts,omitempty"`
	TosURI                  string   `json:"tos_uri,omitempty"`
	PolicyURI               string   `json:"policy_uri,omitempty"`
	JWKSURI                 string   `json:"jwks_uri,omitempty"`
	JWKS                    string   `json:"jwks,omitempty"` // JSON string
	SoftwareID              string   `json:"software_id,omitempty"`
	SoftwareVersion         string   `json:"software_version,omitempty"`
	SoftwareStatement       string   `json:"software_statement,omitempty"` // JWT

	// OPTIONAL OIDC-specific fields
	ApplicationType              string   `json:"application_type,omitempty"`
	SectorIdentifierURI          string   `json:"sector_identifier_uri,omitempty"`
	SubjectType                  string   `json:"subject_type,omitempty"`
	RequestObjectSigningAlg      string   `json:"request_object_signing_alg,omitempty"`
	RequestObjectEncryptionAlg   string   `json:"request_object_encryption_alg,omitempty"`
	RequestObjectEncryptionEnc   string   `json:"request_object_encryption_enc,omitempty"`
	UserInfoSignedResponseAlg    string   `json:"userinfo_signed_response_alg,omitempty"`
	UserInfoEncryptedResponseAlg string   `json:"userinfo_encrypted_response_alg,omitempty"`
	UserInfoEncryptedResponseEnc string   `json:"userinfo_encrypted_response_enc,omitempty"`
	IDTokenSignedResponseAlg     string   `json:"id_token_signed_response_alg,omitempty"`
	IDTokenEncryptedResponseAlg  string   `json:"id_token_encrypted_response_alg,omitempty"`
	IDTokenEncryptedResponseEnc  string   `json:"id_token_encrypted_response_enc,omitempty"`
	TokenEndpointAuthSigningAlg  string   `json:"token_endpoint_auth_signing_alg,omitempty"`
	DefaultMaxAge                int      `json:"default_max_age,omitempty"`
	RequireAuthTime              bool     `json:"require_auth_time,omitempty"`
	DefaultACRValues             []string `json:"default_acr_values,omitempty"`
	InitiateLoginURI             string   `json:"initiate_login_uri,omitempty"`
	RequestURIs                  []string `json:"request_uris,omitempty"`
}

// ClientRegistrationResponse represents the successful registration response
type ClientRegistrationResponse struct {
	Client                         // Embedded client with all metadata
	RegistrationAccessToken string `json:"registration_access_token,omitempty"`
	RegistrationClientURI   string `json:"registration_client_uri,omitempty"`
}

// ClientRegistrationError represents an error response from registration endpoint
type ClientRegistrationError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// OIDC Dynamic Registration error codes
const (
	ErrInvalidRedirectURI          = "invalid_redirect_uri"
	ErrInvalidClientMetadata       = "invalid_client_metadata"
	ErrInvalidSoftwareStatement    = "invalid_software_statement"
	ErrUnapprovedSoftwareStatement = "unapproved_software_statement"
)

// InitialAccessToken represents a token used to authenticate client registration requests
// This provides access control for who can register new OAuth clients
type InitialAccessToken struct {
	Token     string     `json:"token" bson:"_id"`
	IssuedBy  string     `json:"issued_by" bson:"issued_by"`
	ExpiresAt time.Time  `json:"expires_at" bson:"expires_at"`
	Used      bool       `json:"used" bson:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty" bson:"used_at,omitempty"`
	UsedBy    string     `json:"used_by,omitempty" bson:"used_by,omitempty"` // Client ID
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
}
