# OAuth 2.0 Compliance Gap Analysis & Implementation Plan

**Document Version:** 1.0  
**Date:** October 24, 2025  
**Status:** Gap Analysis  
**Specifications:** [RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749), [RFC 6750](https://datatracker.ietf.org/doc/html/rfc6750), [RFC 7009](https://datatracker.ietf.org/doc/html/rfc7009), [RFC 7662](https://datatracker.ietf.org/doc/html/rfc7662)

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Current Implementation Status](#current-implementation-status)
- [Gap Analysis](#gap-analysis)
- [Implementation Priorities](#implementation-priorities)
- [Detailed Implementation Plans](#detailed-implementation-plans)
- [Timeline & Resources](#timeline--resources)

---

## Executive Summary

This document analyzes the current OpenID Connect implementation against **OAuth 2.0 specifications** to identify missing features and create a comprehensive implementation plan.

### Current OAuth 2.0 Coverage:

✅ **Implemented (70%):**
- Authorization Code Grant with PKCE
- Refresh Token Grant
- Implicit Grant (partial)
- Client authentication (client_secret_basic, client_secret_post)
- Token endpoint
- Bearer token authentication

❌ **Missing (30%):**
- **Client Credentials Grant** (RFC 6749 §4.4)
- **Resource Owner Password Credentials Grant** (RFC 6749 §4.3)
- **Token Revocation** (RFC 7009)
- **Token Introspection** (RFC 7662)
- **Device Authorization Grant** (RFC 8628)
- **Token Exchange** (RFC 8693)
- Additional client authentication methods (private_key_jwt, client_secret_jwt)

---

## Current Implementation Status

### ✅ Fully Implemented OAuth 2.0 Features

#### 1. Authorization Code Grant (RFC 6749 §4.1)
**Status:** ✅ Complete with PKCE

**Location:** `backend/pkg/handlers/authorize.go`, `backend/pkg/handlers/token.go`

**Features:**
- Authorization endpoint (`/authorize`)
- Token endpoint (`/token`)
- PKCE support (S256 and plain)
- State parameter for CSRF protection
- Redirect URI validation
- Authorization code expiration (10 minutes)
- One-time use enforcement

**Compliance Level:** 100%

#### 2. Refresh Token Grant (RFC 6749 §6)
**Status:** ✅ Complete

**Location:** `backend/pkg/handlers/token.go`

**Features:**
- Refresh token issuance
- Refresh token validation
- Token rotation
- Scope restriction

**Compliance Level:** 100%

#### 3. Implicit Grant (RFC 6749 §4.2)
**Status:** 🟡 Partial (ID Token only)

**Location:** `backend/pkg/handlers/authorize.go`

**Features:**
- ✅ ID token response (`response_type=id_token`)
- ✅ Combined response (`response_type=token id_token`)
- ❌ Pure access token response (`response_type=token`)
- ❌ Access token in implicit flow has limited scope

**Compliance Level:** 70%

#### 4. Client Authentication
**Status:** ✅ Complete for basic methods

**Location:** `backend/pkg/handlers/token.go`

**Implemented:**
- ✅ `client_secret_basic` (HTTP Basic Auth)
- ✅ `client_secret_post` (Form parameters)
- ✅ `none` (Public clients)

**Missing:**
- ❌ `client_secret_jwt` (JWT with shared secret)
- ❌ `private_key_jwt` (JWT with private key)

**Compliance Level:** 60%

#### 5. Token Response (RFC 6749 §5.1)
**Status:** ✅ Complete

**Features:**
- Access token
- Token type (Bearer)
- Expires in
- Refresh token
- Scope (if different from requested)

**Compliance Level:** 100%

#### 6. Error Responses (RFC 6749 §5.2)
**Status:** ✅ Complete

**Location:** `backend/pkg/handlers/errors.go`

**Error Codes:**
- `invalid_request`
- `invalid_client`
- `invalid_grant`
- `unauthorized_client`
- `unsupported_grant_type`
- `invalid_scope`

**Compliance Level:** 100%

---

## Gap Analysis

### Priority 1: Critical Missing Features (High Impact)

#### Gap 1: Client Credentials Grant (RFC 6749 §4.4)
**Impact:** 🔴 High - Required for machine-to-machine (M2M) authentication

**Use Cases:**
- Backend services authenticating to APIs
- Microservices communication
- Scheduled jobs accessing resources
- Server-to-server integrations

**Current Workaround:** None - This flow is completely missing

**Specification Requirements:**
```http
POST /token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials
&scope=api:read api:write
&client_id=service_123
&client_secret=secret_abc
```

**Response:**
```json
{
  "access_token": "eyJhbGci...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "api:read api:write"
}
```

**Key Differences from Authorization Code:**
- No user involved
- Client is the resource owner
- No refresh token issued
- No authorization code step

---

#### Gap 2: Token Revocation (RFC 7009)
**Impact:** 🔴 High - Required for security and logout

**Use Cases:**
- User logout (revoke all tokens)
- Token compromise (immediate revocation)
- Client uninstallation
- Permission changes
- Security incidents

**Current Workaround:** Tokens expire naturally, but can't be revoked early

**Specification Requirements:**
```http
POST /revoke
Content-Type: application/x-www-form-urlencoded

token=eyJhbGci...
&token_type_hint=access_token
&client_id=client_123
&client_secret=secret_abc
```

**Response:**
```http
HTTP/1.1 200 OK
```

**Features Required:**
- Revoke access tokens
- Revoke refresh tokens
- Cascade revocation (refresh token revokes all derived tokens)
- Token type hint parameter
- Client authentication

---

#### Gap 3: Token Introspection (RFC 7662)
**Impact:** 🟡 Medium - Useful for resource servers

**Use Cases:**
- Resource servers validating tokens
- Checking token status (active/inactive)
- Getting token metadata (scope, exp, client)
- Centralized token validation

**Current Workaround:** Resource servers must decode JWT locally

**Specification Requirements:**
```http
POST /introspect
Content-Type: application/x-www-form-urlencoded

token=eyJhbGci...
&token_type_hint=access_token
&client_id=resource_server
&client_secret=secret_xyz
```

**Response:**
```json
{
  "active": true,
  "scope": "read write",
  "client_id": "client_123",
  "username": "john.doe",
  "token_type": "Bearer",
  "exp": 1698163200,
  "iat": 1698159600,
  "sub": "user_456"
}
```

---

### Priority 2: Optional Features (Medium Impact)

#### Gap 4: Resource Owner Password Credentials (RFC 6749 §4.3)
**Impact:** 🟡 Medium - Useful for legacy apps, but discouraged

**Use Cases:**
- Legacy applications migration
- First-party mobile apps (owned by OP)
- Command-line tools
- Trusted applications

**Security Concerns:**
- ⚠️ User credentials exposed to client
- ⚠️ Deprecated in OAuth 2.1
- ⚠️ Should only be used for trusted first-party clients
- ⚠️ Authorization Code + PKCE is preferred

**Specification Requirements:**
```http
POST /token
Content-Type: application/x-www-form-urlencoded

grant_type=password
&username=john.doe
&password=secret123
&scope=openid profile email
&client_id=mobile_app
&client_secret=secret_abc
```

**Response:**
```json
{
  "access_token": "eyJhbGci...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "def...",
  "id_token": "ghi..."
}
```

**Implementation Notes:**
- Should be disabled by default
- Require explicit configuration to enable
- Rate limiting essential (brute force prevention)
- Support MFA before issuing tokens

---

#### Gap 5: Advanced Client Authentication
**Impact:** 🟢 Low - Enhanced security for high-value clients

**Missing Methods:**

##### a) client_secret_jwt (RFC 7523)
Client authenticates with JWT signed using shared secret (HS256)

```http
POST /token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=abc123
&client_assertion_type=urn:ietf:params:oauth:client-assertion-type:jwt-bearer
&client_assertion=eyJhbGci...
```

##### b) private_key_jwt (RFC 7523)
Client authenticates with JWT signed using private key (RS256)

**Benefits:**
- No shared secret transmission
- Asymmetric cryptography
- Better for distributed systems
- Required for some banking APIs (Open Banking, FAPI)

---

#### Gap 6: Device Authorization Grant (RFC 8628)
**Impact:** 🟢 Low - For input-constrained devices

**Use Cases:**
- Smart TVs
- IoT devices
- Gaming consoles
- CLI tools without browser

**Flow:**
1. Device requests code
2. User visits URL on phone/computer
3. User enters code
4. Device polls for token

**Specification Requirements:**
```http
POST /device_authorization
Content-Type: application/x-www-form-urlencoded

client_id=tv_app
&scope=openid profile
```

**Response:**
```json
{
  "device_code": "GmRhmhcxhwAzkoEqiMEg_DnyEysNkuNhszIySk9eS",
  "user_code": "WDJB-MJHT",
  "verification_uri": "https://example.com/device",
  "verification_uri_complete": "https://example.com/device?user_code=WDJB-MJHT",
  "expires_in": 1800,
  "interval": 5
}
```

---

#### Gap 7: Token Exchange (RFC 8693)
**Impact:** 🟢 Low - For advanced delegation scenarios

**Use Cases:**
- Token delegation (exchange user token for service token)
- Token impersonation (acting on behalf of)
- Multi-tier architectures
- Gateway patterns

**Specification:**
```http
POST /token
Content-Type: application/x-www-form-urlencoded

grant_type=urn:ietf:params:oauth:grant-type:token-exchange
&subject_token=eyJhbGci...
&subject_token_type=urn:ietf:params:oauth:token-type:access_token
&requested_token_type=urn:ietf:params:oauth:token-type:access_token
&audience=https://api.example.com
```

---

## Implementation Priorities

### Recommended Implementation Order:

| Priority | Feature | Effort | Business Value | Security Impact |
|----------|---------|--------|----------------|-----------------|
| 🔴 **P1** | Client Credentials Grant | 2 days | High (M2M) | Medium |
| 🔴 **P1** | Token Revocation | 2-3 days | High (Security) | High |
| 🟡 **P2** | Token Introspection | 2 days | Medium (RS validation) | Medium |
| 🟡 **P2** | Resource Owner Password | 1-2 days | Medium (Legacy) | Low (if disabled) |
| 🟡 **P2** | Complete Implicit Flow | 1 day | Low (Deprecated) | Low |
| 🟢 **P3** | client_secret_jwt | 2 days | Low | High |
| 🟢 **P3** | private_key_jwt | 3 days | Low | High |
| 🟢 **P3** | Device Authorization | 4-5 days | Low | Medium |
| 🟢 **P3** | Token Exchange | 3-4 days | Low | Medium |

**Total Estimated Effort:** 22-30 days for all features

---

## Detailed Implementation Plans

### Plan 1: Client Credentials Grant (2 days)

#### Phase 1.1: Update Token Endpoint (1 day)

**File:** `backend/pkg/handlers/token.go`

```go
func (h *Handlers) Token(c echo.Context) error {
    req := &TokenRequest{
        GrantType:    c.FormValue("grant_type"),
        Code:         c.FormValue("code"),
        RedirectURI:  c.FormValue("redirect_uri"),
        ClientID:     c.FormValue("client_id"),
        ClientSecret: c.FormValue("client_secret"),
        CodeVerifier: c.FormValue("code_verifier"),
        RefreshToken: c.FormValue("refresh_token"),
        Scope:        c.FormValue("scope"), // NEW
        Username:     c.FormValue("username"), // For ROPC
        Password:     c.FormValue("password"), // For ROPC
    }

    // ... existing client authentication ...

    switch req.GrantType {
    case "authorization_code":
        return h.handleAuthorizationCodeGrant(c, req, client)
    case "refresh_token":
        return h.handleRefreshTokenGrant(c, req, client)
    case "client_credentials": // NEW
        return h.handleClientCredentialsGrant(c, req, client)
    case "password": // NEW (Optional)
        return h.handlePasswordGrant(c, req, client)
    default:
        return jsonError(c, http.StatusBadRequest, ErrorUnsupportedGrantType, "Grant type not supported")
    }
}

// handleClientCredentialsGrant implements RFC 6749 §4.4
func (h *Handlers) handleClientCredentialsGrant(c echo.Context, req *TokenRequest, client *models.Client) error {
    // 1. Validate client is authorized for this grant
    if !client.HasGrantType("client_credentials") {
        return jsonError(c, http.StatusBadRequest, ErrorUnauthorizedClient, 
            "Client not authorized for client_credentials grant")
    }
    
    // 2. Validate requested scope
    requestedScope := req.Scope
    if requestedScope == "" {
        requestedScope = client.Scope // Use default client scope
    }
    
    // Validate scope is subset of client's allowed scopes
    if !h.validateScope(requestedScope, client.Scope) {
        return jsonError(c, http.StatusBadRequest, ErrorInvalidScope, 
            "Requested scope exceeds client scope")
    }
    
    // 3. Generate access token (NO USER - client is the resource owner)
    token := models.NewToken(client.ID, "", requestedScope, h.config.JWT.ExpiryMinutes)
    token.TokenType = "client_credentials" // Mark token type
    
    if err := h.storage.CreateToken(token); err != nil {
        return jsonError(c, http.StatusInternalServerError, ErrorServerError, 
            "Failed to create token")
    }
    
    // 4. Return response (NO refresh token, NO ID token)
    response := TokenResponse{
        AccessToken: token.AccessToken,
        TokenType:   "Bearer",
        ExpiresIn:   h.config.JWT.ExpiryMinutes * 60,
        Scope:       token.Scope,
    }
    
    return c.JSON(http.StatusOK, response)
}

func (h *Handlers) validateScope(requested, allowed string) bool {
    requestedScopes := strings.Split(requested, " ")
    allowedScopes := strings.Split(allowed, " ")
    
    allowedMap := make(map[string]bool)
    for _, scope := range allowedScopes {
        allowedMap[scope] = true
    }
    
    for _, scope := range requestedScopes {
        if !allowedMap[scope] {
            return false
        }
    }
    return true
}
```

#### Phase 1.2: Update Token Model (0.5 days)

**File:** `backend/pkg/models/models.go`

```go
type Token struct {
    // ... existing fields ...
    
    UserID     string `json:"user_id,omitempty" bson:"user_id,omitempty"` // Empty for client_credentials
    TokenType  string `json:"token_type,omitempty" bson:"token_type,omitempty"` // "authorization_code", "client_credentials", "password"
}
```

#### Phase 1.3: Testing (0.5 days)

**File:** `backend/pkg/handlers/token_test.go`

```go
func TestClientCredentialsGrant(t *testing.T)
func TestClientCredentialsGrant_InvalidScope(t *testing.T)
func TestClientCredentialsGrant_UnauthorizedClient(t *testing.T)
func TestClientCredentialsGrant_NoRefreshToken(t *testing.T)
```

---

### Plan 2: Token Revocation (2-3 days)

#### Phase 2.1: Create Revocation Endpoint (1 day)

**File:** `backend/pkg/handlers/revoke.go`

```go
package handlers

// RevokeRequest represents a token revocation request per RFC 7009
type RevokeRequest struct {
    Token         string
    TokenTypeHint string // "access_token" or "refresh_token"
}

// Revoke handles token revocation (POST /revoke) per RFC 7009
func (h *Handlers) Revoke(c echo.Context) error {
    req := &RevokeRequest{
        Token:         c.FormValue("token"),
        TokenTypeHint: c.FormValue("token_type_hint"),
    }
    
    // Get client credentials
    clientID := c.FormValue("client_id")
    clientSecret := c.FormValue("client_secret")
    
    // Try Basic Auth if not in form
    if clientID == "" || clientSecret == "" {
        clientID, clientSecret, _ = parseBasicAuth(c.Request().Header.Get("Authorization"))
    }
    
    // Validate client
    client, err := h.storage.ValidateClient(clientID, clientSecret)
    if err != nil {
        // RFC 7009 §2.2: Return 200 even for invalid client (prevent token scanning)
        return c.NoContent(http.StatusOK)
    }
    
    // Revoke token
    if err := h.revokeToken(req.Token, req.TokenTypeHint, client.ID); err != nil {
        // RFC 7009 §2.2: Still return 200 (idempotent operation)
        // Log error for debugging but don't expose to client
    }
    
    return c.NoContent(http.StatusOK)
}

func (h *Handlers) revokeToken(token, tokenTypeHint, clientID string) error {
    // Try as refresh token first if hint provided
    if tokenTypeHint == "refresh_token" || tokenTypeHint == "" {
        if err := h.revokeRefreshToken(token, clientID); err == nil {
            return nil
        }
    }
    
    // Try as access token
    if tokenTypeHint == "access_token" || tokenTypeHint == "" {
        if err := h.revokeAccessToken(token, clientID); err == nil {
            return nil
        }
    }
    
    return nil // Return success even if token not found (idempotent)
}

func (h *Handlers) revokeRefreshToken(refreshToken, clientID string) error {
    token, err := h.storage.GetTokenByRefreshToken(refreshToken)
    if err != nil || token == nil {
        return err
    }
    
    // Verify token belongs to client
    if token.ClientID != clientID {
        return fmt.Errorf("token does not belong to client")
    }
    
    // Revoke all tokens issued from this refresh token
    if token.AuthorizationCodeID != "" {
        _ = h.storage.RevokeTokensByAuthCode(token.AuthorizationCodeID)
    }
    
    // Delete the token
    return h.storage.DeleteToken(token.AccessToken)
}

func (h *Handlers) revokeAccessToken(accessToken, clientID string) error {
    token, err := h.storage.GetTokenByAccessToken(accessToken)
    if err != nil || token == nil {
        return err
    }
    
    // Verify token belongs to client
    if token.ClientID != clientID {
        return fmt.Errorf("token does not belong to client")
    }
    
    // Delete the token
    return h.storage.DeleteToken(accessToken)
}
```

#### Phase 2.2: Add Route (0.5 days)

**File:** `backend/cmd/serve.go`

```go
// Token revocation endpoint (RFC 7009)
e.POST("/revoke", handlers.Revoke)
```

#### Phase 2.3: Update Discovery (0.5 days)

**File:** `backend/pkg/handlers/discovery.go`

```go
response := map[string]interface{}{
    // ... existing fields ...
    
    "revocation_endpoint": h.config.Issuer + "/revoke",
    "revocation_endpoint_auth_methods_supported": []string{
        "client_secret_basic",
        "client_secret_post",
    },
}
```

#### Phase 2.4: Testing (1 day)

Tests for:
- Revoke access token
- Revoke refresh token
- Cascade revocation
- Invalid token (still returns 200)
- Wrong client (still returns 200)
- Token type hint

---

### Plan 3: Token Introspection (2 days)

#### Phase 3.1: Create Introspection Endpoint (1 day)

**File:** `backend/pkg/handlers/introspect.go`

```go
package handlers

// IntrospectionResponse represents token introspection response per RFC 7662
type IntrospectionResponse struct {
    Active    bool   `json:"active"`
    Scope     string `json:"scope,omitempty"`
    ClientID  string `json:"client_id,omitempty"`
    Username  string `json:"username,omitempty"`
    TokenType string `json:"token_type,omitempty"`
    Exp       int64  `json:"exp,omitempty"`
    Iat       int64  `json:"iat,omitempty"`
    Sub       string `json:"sub,omitempty"`
    Aud       string `json:"aud,omitempty"`
    Iss       string `json:"iss,omitempty"`
    Jti       string `json:"jti,omitempty"`
}

// Introspect handles token introspection (POST /introspect) per RFC 7662
func (h *Handlers) Introspect(c echo.Context) error {
    token := c.FormValue("token")
    tokenTypeHint := c.FormValue("token_type_hint")
    
    // Authenticate client (resource server)
    clientID := c.FormValue("client_id")
    clientSecret := c.FormValue("client_secret")
    
    if clientID == "" || clientSecret == "" {
        clientID, clientSecret, _ = parseBasicAuth(c.Request().Header.Get("Authorization"))
    }
    
    // Validate client
    client, err := h.storage.ValidateClient(clientID, clientSecret)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, IntrospectionResponse{Active: false})
    }
    
    // Introspect token
    response := h.introspectToken(token, tokenTypeHint)
    
    return c.JSON(http.StatusOK, response)
}

func (h *Handlers) introspectToken(tokenString, tokenTypeHint string) IntrospectionResponse {
    // Try to find token in database
    token, err := h.storage.GetTokenByAccessToken(tokenString)
    if err != nil || token == nil {
        return IntrospectionResponse{Active: false}
    }
    
    // Check if token is expired
    if token.IsExpired() {
        return IntrospectionResponse{Active: false}
    }
    
    // Get user information if token has user
    var username string
    if token.UserID != "" {
        user, _ := h.storage.GetUserByID(token.UserID)
        if user != nil {
            username = user.Username
        }
    }
    
    // Build response
    return IntrospectionResponse{
        Active:    true,
        Scope:     token.Scope,
        ClientID:  token.ClientID,
        Username:  username,
        TokenType: "Bearer",
        Exp:       token.ExpiresAt.Unix(),
        Iat:       token.CreatedAt.Unix(),
        Sub:       token.UserID,
        Aud:       token.ClientID,
        Iss:       h.config.Issuer,
        Jti:       token.AccessToken, // Or generate unique ID
    }
}
```

#### Phase 3.2: Add Route & Discovery (0.5 days)

#### Phase 3.3: Testing (0.5 days)

---

### Plan 4: Resource Owner Password Credentials (1-2 days)

⚠️ **Security Note:** This grant type is discouraged. Implement only if required for legacy systems.

#### Phase 4.1: Implement Password Grant (1 day)

**File:** `backend/pkg/handlers/token.go`

```go
func (h *Handlers) handlePasswordGrant(c echo.Context, req *TokenRequest, client *models.Client) error {
    // 1. Check if grant is enabled (disabled by default)
    if !h.config.PasswordGrantEnabled {
        return jsonError(c, http.StatusBadRequest, ErrorUnsupportedGrantType, 
            "Password grant is not enabled")
    }
    
    // 2. Validate client is authorized for this grant
    if !client.HasGrantType("password") {
        return jsonError(c, http.StatusBadRequest, ErrorUnauthorizedClient, 
            "Client not authorized for password grant")
    }
    
    // 3. Rate limiting (prevent brute force)
    if h.isRateLimited(c.RealIP(), req.Username) {
        return jsonError(c, http.StatusTooManyRequests, ErrorInvalidGrant, 
            "Too many failed attempts")
    }
    
    // 4. Validate username and password
    user, err := h.storage.GetUserByUsername(req.Username)
    if err != nil || user == nil {
        h.recordFailedAttempt(c.RealIP(), req.Username)
        return jsonError(c, http.StatusBadRequest, ErrorInvalidGrant, 
            "Invalid username or password")
    }
    
    if !crypto.ValidatePassword(req.Password, user.PasswordHash) {
        h.recordFailedAttempt(c.RealIP(), req.Username)
        return jsonError(c, http.StatusBadRequest, ErrorInvalidGrant, 
            "Invalid username or password")
    }
    
    // 5. Validate scope
    requestedScope := req.Scope
    if requestedScope == "" {
        requestedScope = client.Scope
    }
    
    // 6. Create token
    token := models.NewToken(client.ID, user.ID, requestedScope, h.config.JWT.ExpiryMinutes)
    token.TokenType = "password"
    
    if err := h.storage.CreateToken(token); err != nil {
        return jsonError(c, http.StatusInternalServerError, ErrorServerError, 
            "Failed to create token")
    }
    
    // 7. Generate ID token if openid scope requested
    var idToken string
    if strings.Contains(requestedScope, "openid") {
        idToken, _ = h.jwtManager.GenerateIDToken(user, client.ID, "", "")
    }
    
    // 8. Return response
    response := TokenResponse{
        AccessToken:  token.AccessToken,
        TokenType:    "Bearer",
        ExpiresIn:    h.config.JWT.ExpiryMinutes * 60,
        RefreshToken: token.RefreshToken,
        IDToken:      idToken,
        Scope:        token.Scope,
    }
    
    // Clear failed attempts on success
    h.clearFailedAttempts(c.RealIP(), req.Username)
    
    return c.JSON(http.StatusOK, response)
}
```

#### Phase 4.2: Rate Limiting (0.5 days)

Implement rate limiting to prevent brute force attacks.

#### Phase 4.3: Configuration (0.5 days)

Add config flag to enable/disable password grant.

---

## Timeline & Resources

### Phase-by-Phase Timeline

| Phase | Features | Duration | Dependencies |
|-------|----------|----------|--------------|
| **Phase 1** | Client Credentials Grant | 2 days | None |
| **Phase 2** | Token Revocation | 2-3 days | None |
| **Phase 3** | Token Introspection | 2 days | Phase 2 (optional) |
| **Phase 4** | Password Grant (optional) | 1-2 days | None |
| **Phase 5** | Advanced Auth Methods | 5 days | None |
| **Phase 6** | Device Authorization | 4-5 days | Phase 1 |
| **Phase 7** | Token Exchange | 3-4 days | Phase 1 |
| **Total** | All features | **19-23 days** | |

### Recommended MVP (Essential Features)

**Estimated Time:** 7-10 days

1. Client Credentials Grant (2 days)
2. Token Revocation (2-3 days)
3. Token Introspection (2 days)
4. Testing & Documentation (1-2 days)

---

## Success Criteria

### Phase 1: Client Credentials
- ✅ Clients can authenticate with client_credentials grant
- ✅ Tokens issued without user context
- ✅ Scope validation working
- ✅ No refresh tokens issued
- ✅ All tests passing

### Phase 2: Token Revocation
- ✅ Tokens can be revoked via /revoke endpoint
- ✅ Cascade revocation working (refresh token revokes all)
- ✅ Idempotent operation (always returns 200)
- ✅ Client authorization validated
- ✅ Discovery document updated

### Phase 3: Token Introspection
- ✅ Resource servers can validate tokens
- ✅ Active/inactive status returned correctly
- ✅ Token metadata exposed
- ✅ Client authentication required
- ✅ Discovery document updated

---

## References

- [RFC 6749 - OAuth 2.0 Authorization Framework](https://datatracker.ietf.org/doc/html/rfc6749)
- [RFC 6750 - OAuth 2.0 Bearer Token Usage](https://datatracker.ietf.org/doc/html/rfc6750)
- [RFC 7009 - OAuth 2.0 Token Revocation](https://datatracker.ietf.org/doc/html/rfc7009)
- [RFC 7662 - OAuth 2.0 Token Introspection](https://datatracker.ietf.org/doc/html/rfc7662)
- [RFC 8628 - OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628)
- [RFC 8693 - OAuth 2.0 Token Exchange](https://datatracker.ietf.org/doc/html/rfc8693)
- [OAuth 2.1 (Draft)](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1-07)
