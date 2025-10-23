# Dynamic Client Registration & Discovery Compliance Plan

**Document Version:** 1.0  
**Date:** October 22, 2025  
**Status:** Planning Phase  
**Specifications:**
- [OpenID Connect Discovery 1.0](https://openid.net/specs/openid-connect-discovery-1_0.html)
- [OpenID Connect Dynamic Client Registration 1.0 incorporating errata set 2](https://openid.net/specs/openid-connect-registration-1_0.html)
- [OAuth 2.0 Dynamic Client Registration (RFC 7591)](https://tools.ietf.org/html/rfc7591)

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Current Implementation Status](#current-implementation-status)
- [Compliance Requirements](#compliance-requirements)
- [Implementation Tasks](#implementation-tasks)
- [API Specifications](#api-specifications)
- [Security Considerations](#security-considerations)
- [Testing Strategy](#testing-strategy)
- [Deployment Options](#deployment-options)

---

## Executive Summary

This document outlines the plan to implement **OpenID Connect Discovery** and **Dynamic Client Registration** compliance in the openid-golang project. These features enable automated client onboarding and service discovery, making the server a fully compliant **Dynamic OpenID Provider**.

**Key Findings:**
- ✅ **Discovery endpoint partially implemented** - Basic metadata present
- ❌ **Dynamic Registration not implemented** - Missing POST /register endpoint
- ❌ **Client Configuration endpoint not implemented** - Missing GET /register/{client_id}
- ❌ **Client model incomplete** - Missing many OIDC-specific metadata fields

**Recommended Approach:**
- **Phase 1:** Enhance Discovery document (1-2 days)
- **Phase 2:** Implement core Dynamic Registration (3-4 days)
- **Phase 3:** Add Client Configuration endpoint (2-3 days)
- **Phase 4:** Optional features and security hardening (3-5 days)

**Total Estimated Effort:** 9-14 days

---

## Current Implementation Status

### ✅ Partially Implemented

| Feature | Status | Location |
|---------|--------|----------|
| Discovery Endpoint | ⚠️ Partial | `GET /.well-known/openid-configuration` |
| JWKS Endpoint | ✅ Complete | `GET /.well-known/jwks.json` |
| Basic Client Model | ⚠️ Partial | `backend/pkg/models/models.go` |

### ❌ Not Implemented

| Feature | Priority | Impact |
|---------|----------|--------|
| Dynamic Registration Endpoint | 🔴 High | Required for Dynamic OP |
| Client Configuration Endpoint | 🟡 Medium | Read/Update client metadata |
| Registration Access Tokens | 🟡 Medium | Secure client management |
| Initial Access Tokens | 🟢 Low | Optional rate limiting |
| Enhanced Client Metadata | 🔴 High | OIDC compliance |

---

## Compliance Requirements

### Discovery Document (OpenID.Discovery Section 3)

**Current Discovery Response:**
```json
{
  "issuer": "https://example.com",
  "authorization_endpoint": "https://example.com/authorize",
  "token_endpoint": "https://example.com/token",
  "userinfo_endpoint": "https://example.com/userinfo",
  "jwks_uri": "https://example.com/.well-known/jwks.json",
  "scopes_supported": ["openid", "profile", "email"],
  "response_types_supported": ["code", "id_token", "id_token token"],
  "response_modes_supported": ["query", "fragment"],
  "grant_types_supported": ["authorization_code", "refresh_token"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "token_endpoint_auth_methods_supported": ["client_secret_basic", "client_secret_post"],
  "claims_supported": ["sub", "iss", "aud", "exp", "iat", "name", "given_name", "family_name", "email", "picture"],
  "code_challenge_methods_supported": ["plain", "S256"]
}
```

**Missing REQUIRED Fields:**
- ✅ All mandatory fields present

**Missing RECOMMENDED Fields:**
- ❌ `registration_endpoint` - URL of dynamic registration endpoint
- ❌ `service_documentation` - URL to human-readable service documentation
- ❌ `ui_locales_supported` - Supported UI locales
- ❌ `claims_locales_supported` - Supported claims locales
- ❌ `op_policy_uri` - URL to OP's policy
- ❌ `op_tos_uri` - URL to OP's terms of service

### Dynamic Registration (OpenID.Registration Section 2)

**Required Client Metadata Fields:**
- ✅ `redirect_uris` - Registered redirect URIs (array)
- ⚠️ `response_types` - Supported response types (partially)
- ⚠️ `grant_types` - Supported grant types (partially)
- ❌ `application_type` - "web" or "native"
- ❌ `contacts` - Contact emails
- ❌ `client_name` - Human-readable client name
- ❌ `logo_uri` - Client logo URL
- ❌ `client_uri` - Client homepage URL
- ❌ `policy_uri` - Privacy policy URL
- ❌ `tos_uri` - Terms of service URL
- ❌ `jwks_uri` - Client's JWK Set URL
- ❌ `jwks` - Client's JWK Set by value
- ❌ `sector_identifier_uri` - Sector identifier for pairwise subjects
- ❌ `subject_type` - "public" or "pairwise"
- ❌ `id_token_signed_response_alg` - ID token signing algorithm
- ❌ `id_token_encrypted_response_alg` - ID token encryption algorithm
- ❌ `id_token_encrypted_response_enc` - ID token encryption encoding
- ❌ `userinfo_signed_response_alg` - UserInfo signing algorithm
- ❌ `userinfo_encrypted_response_alg` - UserInfo encryption algorithm
- ❌ `userinfo_encrypted_response_enc` - UserInfo encryption encoding
- ❌ `request_object_signing_alg` - Request object signing algorithm
- ❌ `request_object_encryption_alg` - Request object encryption algorithm
- ❌ `request_object_encryption_enc` - Request object encryption encoding
- ❌ `token_endpoint_auth_method` - Token endpoint auth method
- ❌ `token_endpoint_auth_signing_alg` - JWT auth signing algorithm
- ❌ `default_max_age` - Default max authentication age
- ❌ `require_auth_time` - Whether auth_time is required
- ❌ `default_acr_values` - Default ACR values
- ❌ `initiate_login_uri` - Third-party initiated login URI
- ❌ `request_uris` - Pre-registered request URIs

**Registration Response Fields:**
- ❌ `client_id` - Unique client identifier (REQUIRED)
- ❌ `client_secret` - Client secret (OPTIONAL)
- ❌ `registration_access_token` - Token for client management (OPTIONAL)
- ❌ `registration_client_uri` - Client configuration endpoint URL (OPTIONAL)
- ❌ `client_id_issued_at` - Timestamp of client_id issuance (OPTIONAL)
- ❌ `client_secret_expires_at` - Expiration of client_secret or 0 (REQUIRED if client_secret issued)

---

## Implementation Tasks

### Phase 1: Enhance Discovery Document (1-2 days)

#### Task 1.1: Add Registration Endpoint to Discovery
**Priority:** 🔴 Critical  
**Effort:** 0.5 days

**Changes:**
1. Update `DiscoveryResponse` struct in `discovery.go`
2. Add `registration_endpoint` field
3. Add optional documentation fields

```go
type DiscoveryResponse struct {
    // ... existing fields ...
    RegistrationEndpoint              *string  `json:"registration_endpoint,omitempty"`
    ServiceDocumentation              *string  `json:"service_documentation,omitempty"`
    UILocalesSupported                []string `json:"ui_locales_supported,omitempty"`
    ClaimsLocalesSupported            []string `json:"claims_locales_supported,omitempty"`
    OPPolicyURI                       *string  `json:"op_policy_uri,omitempty"`
    OPTosURI                          *string  `json:"op_tos_uri,omitempty"`
}
```

#### Task 1.2: Add Registration Capabilities to Discovery
**Priority:** 🔴 Critical  
**Effort:** 0.5 days

**Add to discovery response:**
- `token_endpoint_auth_methods_supported` - Already present ✅
- `registration_endpoint` - New field
- Enhanced `response_types_supported` for hybrid flows (future)

**Acceptance Criteria:**
- [ ] `registration_endpoint` included in discovery when enabled
- [ ] Discovery document validates against OIDC spec
- [ ] Tests verify new fields

---

### Phase 2: Implement Dynamic Registration (3-4 days)

#### Task 2.1: Enhance Client Model
**Priority:** 🔴 Critical  
**Effort:** 1 day

**Current Client struct:**
```go
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
```

**New Enhanced Client struct:**
```go
type Client struct {
    // Core OAuth 2.0 fields
    ID            string    `json:"client_id" bson:"_id"`
    Secret        string    `json:"client_secret,omitempty" bson:"client_secret,omitempty"`
    SecretExpiresAt int64   `json:"client_secret_expires_at" bson:"client_secret_expires_at"` // 0 = never expires
    RedirectURIs  []string  `json:"redirect_uris" bson:"redirect_uris"`
    GrantTypes    []string  `json:"grant_types,omitempty" bson:"grant_types,omitempty"`
    ResponseTypes []string  `json:"response_types,omitempty" bson:"response_types,omitempty"`
    Scope         string    `json:"scope,omitempty" bson:"scope,omitempty"`
    
    // OIDC-specific fields
    ApplicationType              string   `json:"application_type,omitempty" bson:"application_type,omitempty"` // "web" or "native"
    Contacts                     []string `json:"contacts,omitempty" bson:"contacts,omitempty"`
    ClientName                   string   `json:"client_name,omitempty" bson:"client_name,omitempty"`
    ClientNameLocalized          map[string]string `json:"-" bson:"client_name_localized,omitempty"` // e.g., "client_name#ja-JP"
    LogoURI                      string   `json:"logo_uri,omitempty" bson:"logo_uri,omitempty"`
    LogoURILocalized             map[string]string `json:"-" bson:"logo_uri_localized,omitempty"`
    ClientURI                    string   `json:"client_uri,omitempty" bson:"client_uri,omitempty"`
    ClientURILocalized           map[string]string `json:"-" bson:"client_uri_localized,omitempty"`
    PolicyURI                    string   `json:"policy_uri,omitempty" bson:"policy_uri,omitempty"`
    PolicyURILocalized           map[string]string `json:"-" bson:"policy_uri_localized,omitempty"`
    TosURI                       string   `json:"tos_uri,omitempty" bson:"tos_uri,omitempty"`
    TosURILocalized              map[string]string `json:"-" bson:"tos_uri_localized,omitempty"`
    
    // JWK/Signing fields
    JWKSURI                      string   `json:"jwks_uri,omitempty" bson:"jwks_uri,omitempty"`
    JWKS                         string   `json:"jwks,omitempty" bson:"jwks,omitempty"` // JWK Set as JSON string
    SectorIdentifierURI          string   `json:"sector_identifier_uri,omitempty" bson:"sector_identifier_uri,omitempty"`
    SubjectType                  string   `json:"subject_type,omitempty" bson:"subject_type,omitempty"` // "public" or "pairwise"
    
    // ID Token preferences
    IDTokenSignedResponseAlg     string   `json:"id_token_signed_response_alg,omitempty" bson:"id_token_signed_response_alg,omitempty"`
    IDTokenEncryptedResponseAlg  string   `json:"id_token_encrypted_response_alg,omitempty" bson:"id_token_encrypted_response_alg,omitempty"`
    IDTokenEncryptedResponseEnc  string   `json:"id_token_encrypted_response_enc,omitempty" bson:"id_token_encrypted_response_enc,omitempty"`
    
    // UserInfo preferences
    UserInfoSignedResponseAlg    string   `json:"userinfo_signed_response_alg,omitempty" bson:"userinfo_signed_response_alg,omitempty"`
    UserInfoEncryptedResponseAlg string   `json:"userinfo_encrypted_response_alg,omitempty" bson:"userinfo_encrypted_response_alg,omitempty"`
    UserInfoEncryptedResponseEnc string   `json:"userinfo_encrypted_response_enc,omitempty" bson:"userinfo_encrypted_response_enc,omitempty"`
    
    // Request Object preferences
    RequestObjectSigningAlg      string   `json:"request_object_signing_alg,omitempty" bson:"request_object_signing_alg,omitempty"`
    RequestObjectEncryptionAlg   string   `json:"request_object_encryption_alg,omitempty" bson:"request_object_encryption_alg,omitempty"`
    RequestObjectEncryptionEnc   string   `json:"request_object_encryption_enc,omitempty" bson:"request_object_encryption_enc,omitempty"`
    
    // Token Endpoint Auth
    TokenEndpointAuthMethod      string   `json:"token_endpoint_auth_method,omitempty" bson:"token_endpoint_auth_method,omitempty"`
    TokenEndpointAuthSigningAlg  string   `json:"token_endpoint_auth_signing_alg,omitempty" bson:"token_endpoint_auth_signing_alg,omitempty"`
    
    // Authentication requirements
    DefaultMaxAge                int      `json:"default_max_age,omitempty" bson:"default_max_age,omitempty"`
    RequireAuthTime              bool     `json:"require_auth_time,omitempty" bson:"require_auth_time,omitempty"`
    DefaultACRValues             []string `json:"default_acr_values,omitempty" bson:"default_acr_values,omitempty"`
    
    // Advanced features
    InitiateLoginURI             string   `json:"initiate_login_uri,omitempty" bson:"initiate_login_uri,omitempty"`
    RequestURIs                  []string `json:"request_uris,omitempty" bson:"request_uris,omitempty"`
    
    // Registration metadata
    RegistrationAccessToken      string    `json:"-" bson:"registration_access_token,omitempty"` // Never send in client responses
    ClientIDIssuedAt             int64     `json:"client_id_issued_at,omitempty" bson:"client_id_issued_at,omitempty"`
    CreatedAt                    time.Time `json:"-" bson:"created_at"`
    UpdatedAt                    time.Time `json:"-" bson:"updated_at"`
}
```

**Acceptance Criteria:**
- [ ] Client model includes all OIDC Registration fields
- [ ] Localized fields supported (e.g., `client_name#ja-JP`)
- [ ] Backward compatibility maintained
- [ ] MongoDB and JSON storage updated

#### Task 2.2: Create Registration Request/Response Types
**Priority:** 🔴 Critical  
**Effort:** 0.5 days

**New types:**
```go
// ClientRegistrationRequest represents a dynamic client registration request
type ClientRegistrationRequest struct {
    RedirectURIs                 []string  `json:"redirect_uris"` // REQUIRED
    ResponseTypes                []string  `json:"response_types,omitempty"`
    GrantTypes                   []string  `json:"grant_types,omitempty"`
    ApplicationType              string    `json:"application_type,omitempty"`
    Contacts                     []string  `json:"contacts,omitempty"`
    ClientName                   string    `json:"client_name,omitempty"`
    LogoURI                      string    `json:"logo_uri,omitempty"`
    ClientURI                    string    `json:"client_uri,omitempty"`
    PolicyURI                    string    `json:"policy_uri,omitempty"`
    TosURI                       string    `json:"tos_uri,omitempty"`
    JWKSURI                      string    `json:"jwks_uri,omitempty"`
    JWKS                         *json.RawMessage `json:"jwks,omitempty"`
    SubjectType                  string    `json:"subject_type,omitempty"`
    TokenEndpointAuthMethod      string    `json:"token_endpoint_auth_method,omitempty"`
    // ... other fields
}

// ClientRegistrationResponse represents the registration response
type ClientRegistrationResponse struct {
    Client // Embedded client with all metadata
    RegistrationAccessToken string  `json:"registration_access_token,omitempty"`
    RegistrationClientURI   string  `json:"registration_client_uri,omitempty"`
}

// ClientRegistrationError represents registration error response
type ClientRegistrationError struct {
    Error            string `json:"error"`
    ErrorDescription string `json:"error_description,omitempty"`
}
```

**Error codes:**
- `invalid_redirect_uri` - Invalid redirect URI
- `invalid_client_metadata` - Invalid metadata field value
- `invalid_software_statement` - Invalid software statement (if used)
- `unapproved_software_statement` - Software statement not approved

#### Task 2.3: Implement Registration Endpoint Handler
**Priority:** 🔴 Critical  
**Effort:** 2 days

**File:** `backend/pkg/handlers/registration.go`

**Implementation:**
```go
// Register handles dynamic client registration (POST /register)
func (h *Handlers) Register(c echo.Context) error {
    // 1. Check if registration is enabled in config
    if !h.config.EnableDynamicRegistration {
        return c.JSON(http.StatusForbidden, ClientRegistrationError{
            Error: "registration_not_supported",
            ErrorDescription: "Dynamic client registration is not enabled",
        })
    }
    
    // 2. Optionally validate Initial Access Token
    if h.config.RequireInitialAccessToken {
        token := extractBearerToken(c)
        if !h.validateInitialAccessToken(token) {
            return c.JSON(http.StatusUnauthorized, map[string]string{
                "error": "invalid_token",
                "error_description": "Invalid or missing initial access token",
            })
        }
    }
    
    // 3. Parse request body
    var req ClientRegistrationRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, ClientRegistrationError{
            Error: "invalid_client_metadata",
            ErrorDescription: "Invalid JSON in request body",
        })
    }
    
    // 4. Validate request
    if err := h.validateRegistrationRequest(&req); err != nil {
        return err // Returns appropriate JSON error
    }
    
    // 5. Generate client_id and optional client_secret
    client := h.buildClientFromRequest(&req)
    client.ID = generateClientID()
    
    if needsClientSecret(req.TokenEndpointAuthMethod) {
        client.Secret = generateClientSecret()
        client.SecretExpiresAt = 0 // Never expires by default
    }
    
    client.ClientIDIssuedAt = time.Now().Unix()
    
    // 6. Generate registration access token
    client.RegistrationAccessToken = generateRegistrationAccessToken()
    
    // 7. Store client
    if err := h.storage.CreateClient(client); err != nil {
        return c.JSON(http.StatusInternalServerError, ClientRegistrationError{
            Error: "server_error",
            ErrorDescription: "Failed to register client",
        })
    }
    
    // 8. Build response
    response := ClientRegistrationResponse{
        Client: *client,
        RegistrationAccessToken: client.RegistrationAccessToken,
        RegistrationClientURI: fmt.Sprintf("%s/register/%s", h.config.Issuer, client.ID),
    }
    
    // Remove internal fields
    response.Client.RegistrationAccessToken = ""
    
    return c.JSON(http.StatusCreated, response)
}
```

**Validation logic:**
```go
func (h *Handlers) validateRegistrationRequest(req *ClientRegistrationRequest) error {
    // 1. redirect_uris is REQUIRED
    if len(req.RedirectURIs) == 0 {
        return c.JSON(http.StatusBadRequest, ClientRegistrationError{
            Error: "invalid_redirect_uri",
            ErrorDescription: "redirect_uris is required and must not be empty",
        })
    }
    
    // 2. Validate redirect URIs based on application_type
    appType := req.ApplicationType
    if appType == "" {
        appType = "web" // Default
    }
    
    for _, uri := range req.RedirectURIs {
        if err := validateRedirectURI(uri, appType); err != nil {
            return c.JSON(http.StatusBadRequest, ClientRegistrationError{
                Error: "invalid_redirect_uri",
                ErrorDescription: err.Error(),
            })
        }
    }
    
    // 3. Validate grant_types and response_types alignment
    if err := validateGrantAndResponseTypes(req.GrantTypes, req.ResponseTypes); err != nil {
        return c.JSON(http.StatusBadRequest, ClientRegistrationError{
            Error: "invalid_client_metadata",
            ErrorDescription: err.Error(),
        })
    }
    
    // 4. Validate JWKS URI and JWKS are not both present
    if req.JWKSURI != "" && req.JWKS != nil {
        return c.JSON(http.StatusBadRequest, ClientRegistrationError{
            Error: "invalid_client_metadata",
            ErrorDescription: "jwks_uri and jwks cannot both be present",
        })
    }
    
    // 5. Validate sector_identifier_uri if present
    if req.SubjectType == "pairwise" && req.SectorIdentifierURI != "" {
        if err := h.validateSectorIdentifierURI(req.SectorIdentifierURI, req.RedirectURIs); err != nil {
            return c.JSON(http.StatusBadRequest, ClientRegistrationError{
                Error: "invalid_client_metadata",
                ErrorDescription: err.Error(),
            })
        }
    }
    
    return nil
}
```

**Acceptance Criteria:**
- [ ] POST /register endpoint implemented
- [ ] Generates unique client_id and client_secret
- [ ] Validates all required fields
- [ ] Returns proper error codes per spec
- [ ] Stores client in database
- [ ] Returns 201 Created with client metadata
- [ ] Optional Initial Access Token validation

---

### Phase 3: Client Configuration Endpoint (2-3 days)

#### Task 3.1: Implement Client Read Endpoint
**Priority:** 🟡 Medium  
**Effort:** 1.5 days

**Endpoint:** `GET /register/{client_id}`

**Implementation:**
```go
// GetClientConfiguration handles reading client configuration (GET /register/:client_id)
func (h *Handlers) GetClientConfiguration(c echo.Context) error {
    // 1. Extract client_id from URL
    clientID := c.Param("client_id")
    
    // 2. Extract and validate Registration Access Token
    token := extractBearerToken(c)
    if token == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{
            "error": "invalid_token",
            "error_description": "Registration access token required",
        })
    }
    
    // 3. Get client from storage
    client, err := h.storage.GetClientByID(clientID)
    if err != nil || client == nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{
            "error": "invalid_token",
            "error_description": "Invalid client or token",
        })
    }
    
    // 4. Validate registration access token matches
    if client.RegistrationAccessToken != token {
        return c.JSON(http.StatusUnauthorized, map[string]string{
            "error": "invalid_token",
            "error_description": "Invalid registration access token",
        })
    }
    
    // 5. Build response with all client metadata
    response := *client
    response.RegistrationAccessToken = "" // Don't include in response
    
    return c.JSON(http.StatusOK, response)
}
```

**Acceptance Criteria:**
- [ ] GET /register/:client_id implemented
- [ ] Requires valid Registration Access Token
- [ ] Returns all client metadata
- [ ] Returns 200 OK on success
- [ ] Returns 401 Unauthorized on invalid token
- [ ] Never returns 404 (security requirement)

#### Task 3.2: Implement Client Update Endpoint (Optional)
**Priority:** 🟢 Low  
**Effort:** 1 day

**Endpoint:** `PUT /register/{client_id}`

This is defined in [RFC 7592](https://tools.ietf.org/html/rfc7592) as an experimental feature.

**Implementation:** Similar to registration but updates existing client.

#### Task 3.3: Implement Client Delete Endpoint (Optional)
**Priority:** 🟢 Low  
**Effort:** 0.5 days

**Endpoint:** `DELETE /register/{client_id}`

---

### Phase 4: Security & Optional Features (3-5 days)

#### Task 4.1: Implement Initial Access Token Support
**Priority:** 🟡 Medium  
**Effort:** 1 day

**Purpose:** Control who can register clients

**Options:**
1. **Simple pre-shared tokens** - Store allowed tokens in config/database
2. **JWT-based tokens** - Issue time-limited registration JWTs
3. **Disable requirement** - Allow open registration (risky)

**Implementation:**
```go
type InitialAccessToken struct {
    Token     string    `json:"token" bson:"_id"`
    IssuedBy  string    `json:"issued_by" bson:"issued_by"`
    ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`
    Used      bool      `json:"used" bson:"used"`
    UsedAt    *time.Time `json:"used_at,omitempty" bson:"used_at,omitempty"`
    CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
```

#### Task 4.2: Add Rate Limiting
**Priority:** 🟡 Medium  
**Effort:** 1 day

**Purpose:** Prevent DoS attacks on registration endpoint

**Options:**
- Use middleware like `golang.org/x/time/rate`
- Implement IP-based throttling
- Redis-based distributed rate limiting

#### Task 4.3: Implement Sector Identifier URI Validation
**Priority:** 🟡 Medium  
**Effort:** 1 day

**Purpose:** Validate sector_identifier_uri for pairwise subject identifiers

**Implementation:**
```go
func (h *Handlers) validateSectorIdentifierURI(sectorURI string, redirectURIs []string) error {
    // 1. Must use HTTPS
    if !strings.HasPrefix(sectorURI, "https://") {
        return errors.New("sector_identifier_uri must use https scheme")
    }
    
    // 2. Fetch JSON array from URI
    resp, err := http.Get(sectorURI)
    if err != nil {
        return fmt.Errorf("failed to fetch sector_identifier_uri: %w", err)
    }
    defer resp.Body.Close()
    
    // 3. Parse JSON array
    var sectorRedirectURIs []string
    if err := json.NewDecoder(resp.Body).Decode(&sectorRedirectURIs); err != nil {
        return errors.New("sector_identifier_uri must return JSON array of redirect URIs")
    }
    
    // 4. Verify all redirect_uris are in the sector array
    for _, uri := range redirectURIs {
        if !contains(sectorRedirectURIs, uri) {
            return fmt.Errorf("redirect_uri %s not found in sector_identifier_uri", uri)
        }
    }
    
    return nil
}
```

#### Task 4.4: Add Software Statement Support (Optional)
**Priority:** 🟢 Low  
**Effort:** 2 days

**Purpose:** Support signed assertions about client software

Defined in [RFC 7591 Section 2.3](https://tools.ietf.org/html/rfc7591#section-2.3)

---

## API Specifications

### POST /register

**Request:**
```http
POST /register HTTP/1.1
Host: server.example.com
Content-Type: application/json
Authorization: Bearer <initial_access_token> (optional)

{
  "redirect_uris": [
    "https://client.example.org/callback",
    "https://client.example.org/callback2"
  ],
  "response_types": ["code", "code id_token"],
  "grant_types": ["authorization_code", "implicit"],
  "application_type": "web",
  "client_name": "My Example Client",
  "client_name#ja-JP": "クライアント名",
  "logo_uri": "https://client.example.org/logo.png",
  "subject_type": "pairwise",
  "sector_identifier_uri": "https://client.example.org/sector.json",
  "token_endpoint_auth_method": "client_secret_basic",
  "jwks_uri": "https://client.example.org/jwks.json",
  "contacts": ["admin@client.example.org"]
}
```

**Success Response (201 Created):**
```json
{
  "client_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_secret": "cf136dc3c1fc93f31185e5885805d",
  "client_secret_expires_at": 0,
  "registration_access_token": "this.is.a.registration.access.token",
  "registration_client_uri": "https://server.example.com/register/550e8400-e29b-41d4-a716-446655440000",
  "client_id_issued_at": 1729670000,
  "redirect_uris": [...],
  "response_types": [...],
  "grant_types": [...],
  "application_type": "web",
  "client_name": "My Example Client",
  ...
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "invalid_redirect_uri",
  "error_description": "One or more redirect_uri values are invalid"
}
```

### GET /register/:client_id

**Request:**
```http
GET /register/550e8400-e29b-41d4-a716-446655440000 HTTP/1.1
Host: server.example.com
Authorization: Bearer this.is.a.registration.access.token
```

**Success Response (200 OK):**
```json
{
  "client_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_secret": "cf136dc3c1fc93f31185e5885805d",
  "client_secret_expires_at": 0,
  "registration_client_uri": "https://server.example.com/register/550e8400-e29b-41d4-a716-446655440000",
  "client_id_issued_at": 1729670000,
  ...all client metadata...
}
```

**Error Response (401 Unauthorized):**
```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer error="invalid_token",
  error_description="The registration access token is invalid"
```

---

## Security Considerations

### 1. TLS Requirements (Section 9.3)
- **MUST** use TLS for all registration endpoints
- **MUST** validate TLS certificates
- **SHOULD** use TLS 1.2+ with strong cipher suites

### 2. Impersonation Prevention (Section 9.1)
- Validate `logo_uri` and `policy_uri` match redirect URI domains
- Warn users about untrusted dynamically registered clients
- Display warnings for clients not previously trusted

### 3. Redirect URI Validation
**Web clients:**
- **MUST** use HTTPS scheme (except localhost for development)
- **MUST NOT** use localhost in production

**Native clients:**
- **MUST** use custom URI schemes or http://localhost
- **MAY** use http://127.0.0.1 or http://[::1]

### 4. Rate Limiting
- Implement rate limiting on POST /register
- Prevent denial-of-service attacks
- Consider IP-based throttling

### 5. Initial Access Tokens
- Optional but **RECOMMENDED** for production
- Restricts who can register clients
- Can be time-limited or single-use

### 6. Registration Access Tokens
- Cryptographically random (minimum 128 bits)
- Associated with specific client
- Required for reading/updating client

---

## Testing Strategy

### Unit Tests
- [ ] Client model validation
- [ ] Redirect URI validation (web vs native)
- [ ] Grant type and response type alignment
- [ ] JWKS validation
- [ ] Sector identifier URI validation
- [ ] Token generation functions

### Integration Tests
- [ ] Complete registration flow
- [ ] Client read operations
- [ ] Error handling for invalid requests
- [ ] Initial access token validation
- [ ] Registration access token validation

### Compliance Tests
- [ ] Use OpenID Certification test suite
- [ ] Test against conformance test clients
- [ ] Validate discovery document structure
- [ ] Test localized metadata handling

### Security Tests
- [ ] Invalid redirect URI attempts
- [ ] Token substitution attempts
- [ ] Rate limiting effectiveness
- [ ] TLS certificate validation

---

## Deployment Options

### Option 1: Fully Open Registration
- **Pros:** Easy onboarding, developer-friendly
- **Cons:** Risk of abuse, spam clients
- **Use Case:** Internal development, trusted networks

### Option 2: Initial Access Token Required
- **Pros:** Controlled registration, prevents abuse
- **Cons:** Requires token distribution mechanism
- **Use Case:** Production environments, B2B

### Option 3: Admin-Approved Only
- **Pros:** Maximum control and security
- **Cons:** Manual process, slower onboarding
- **Use Case:** High-security environments, financial services

### Option 4: Hybrid Approach
- Allow open registration with limitations (read-only scopes)
- Require approval for privileged scopes
- Rate limit open registrations aggressively

---

## Configuration

**New config options needed:**
```go
type RegistrationConfig struct {
    Enabled                     bool     `json:"enabled" yaml:"enabled"`
    RequireInitialAccessToken   bool     `json:"require_initial_access_token" yaml:"require_initial_access_token"`
    AllowedRedirectURISchemes   []string `json:"allowed_redirect_uri_schemes" yaml:"allowed_redirect_uri_schemes"`
    DefaultTokenEndpointAuthMethod string `json:"default_token_endpoint_auth_method" yaml:"default_token_endpoint_auth_method"`
    ClientSecretExpiryDays      int      `json:"client_secret_expiry_days" yaml:"client_secret_expiry_days"` // 0 = never
    RateLimitPerIP              int      `json:"rate_limit_per_ip" yaml:"rate_limit_per_ip"` // requests per hour
    ValidateSectorIdentifierURI bool     `json:"validate_sector_identifier_uri" yaml:"validate_sector_identifier_uri"`
}
```

---

## Migration Strategy

### Existing Clients
1. Add migration script to populate new fields with defaults
2. Set `application_type` to "web" for existing clients
3. Set `token_endpoint_auth_method` based on whether client has secret
4. Populate `client_id_issued_at` with creation timestamp
5. Generate `registration_access_token` for existing clients (optional)

### Database Schema
**MongoDB:**
- Add indexes on `client_id`, `registration_access_token`
- Add compound index on `application_type` + `subject_type`

**JSON:**
- Ensure backward compatibility with existing client files
- Add default values for new fields when reading

---

## Implementation Checklist

### Phase 1: Discovery Enhancement
- [ ] Add `registration_endpoint` to discovery
- [ ] Add optional documentation fields
- [ ] Update tests
- [ ] Update API documentation

### Phase 2: Core Registration
- [ ] Enhance Client model with all OIDC fields
- [ ] Update storage interfaces
- [ ] Implement registration request validation
- [ ] Implement POST /register endpoint
- [ ] Generate client credentials
- [ ] Store clients in database
- [ ] Add comprehensive error handling
- [ ] Write unit tests
- [ ] Write integration tests

### Phase 3: Client Configuration
- [ ] Implement GET /register/:client_id
- [ ] Implement registration access token validation
- [ ] Add security headers
- [ ] Write tests

### Phase 4: Security & Polish
- [ ] Implement Initial Access Token support
- [ ] Add rate limiting
- [ ] Implement sector identifier validation
- [ ] Add admin API for token management
- [ ] Security audit
- [ ] Performance testing
- [ ] Documentation

---

## Timeline

| Phase | Tasks | Duration | Dependencies |
|-------|-------|----------|--------------|
| Phase 1 | Discovery enhancement | 1-2 days | None |
| Phase 2 | Core registration | 3-4 days | Phase 1 |
| Phase 3 | Client configuration | 2-3 days | Phase 2 |
| Phase 4 | Security features | 3-5 days | Phase 3 |
| **Total** | | **9-14 days** | |

---

## Success Criteria

1. ✅ POST /register endpoint accepts valid registration requests
2. ✅ Clients can be registered with full OIDC metadata
3. ✅ Discovery document advertises registration endpoint
4. ✅ Registration access tokens secure client management
5. ✅ Proper error responses for all failure cases
6. ✅ All OIDC Registration compliance tests pass
7. ✅ Rate limiting prevents abuse
8. ✅ TLS enforced on all endpoints
9. ✅ Comprehensive test coverage (>80%)
10. ✅ Documentation complete and accurate

---

## References

- [OpenID Connect Discovery 1.0](https://openid.net/specs/openid-connect-discovery-1_0.html)
- [OpenID Connect Dynamic Client Registration 1.0](https://openid.net/specs/openid-connect-registration-1_0.html)
- [OAuth 2.0 Dynamic Client Registration Protocol (RFC 7591)](https://tools.ietf.org/html/rfc7591)
- [OAuth 2.0 Dynamic Client Registration Management Protocol (RFC 7592)](https://tools.ietf.org/html/rfc7592)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-22 | System | Initial plan created |

---

## Next Steps

1. **Review and approve this plan** - Get stakeholder buy-in
2. **Start with Phase 1** - Quick win with discovery enhancement
3. **Implement Phase 2** - Core registration functionality
4. **Test thoroughly** - Ensure compliance before Phase 3
5. **Deploy incrementally** - Roll out feature flags for gradual adoption
