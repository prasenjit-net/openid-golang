# OpenID Connect Core 1.0 Compliance Plan

**Document Version:** 1.0  
**Date:** October 20, 2025  
**Status:** Planning Phase  
**Specification:** [OpenID Connect Core 1.0 incorporating errata set 2](https://openid.net/specs/openid-connect-core-1_0.html)

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Current Implementation Status](#current-implementation-status)
- [Compliance Gap Analysis](#compliance-gap-analysis)
- [Priority 1: Critical Security & Mandatory Features](#priority-1-critical-security--mandatory-features)
- [Priority 2: Important Features](#priority-2-important-features)
- [Priority 3: Optional Enhancements](#priority-3-optional-enhancements)
- [Implementation Roadmap](#implementation-roadmap)
- [Testing & Validation Strategy](#testing--validation-strategy)
- [References](#references)

---

## Executive Summary

This document outlines the compliance plan for implementing OpenID Connect Core 1.0 specification in the openid-golang project. The current implementation provides basic Authorization Code Flow support but lacks several mandatory features and security controls required by the specification.

**Key Findings:**
- ✅ **7 of 15 mandatory features** implemented
- ⚠️ **8 mandatory features** missing or incomplete
- 🔒 **3 critical security gaps** identified
- 📋 **20 tasks** defined across 3 priority levels

**Recommended Approach:**
Start with Priority 1 (7 critical tasks) to address security vulnerabilities and mandatory requirements, then proceed to Priority 2 (13 tasks) for important features, and finally Priority 3 (optional enhancements).

---

## Current Implementation Status

### ✅ Implemented Features

| Feature | Status | Spec Reference |
|---------|--------|----------------|
| Authorization Code Flow | ✅ Working | 3.1 |
| ID Token Generation (RS256) | ✅ Working | 2, 10.1 |
| Token Endpoint | ✅ Working | 3.1.3 |
| UserInfo Endpoint | ✅ Working | 5.3 |
| Discovery Endpoint | ✅ Working | OpenID.Discovery |
| JWKS Endpoint | ✅ Working | 3.1.3.7 |
| Refresh Token Flow | ✅ Working | 12 |
| Basic PKCE Validation | ✅ Working | RFC 7636 |
| Implicit Flow (partial) | ⚠️ Partial | 3.2 |

### ⚠️ Missing or Incomplete Features

| Feature | Priority | Impact |
|---------|----------|--------|
| Session Management | P1 | Critical - No consent tracking |
| Nonce Validation | P1 | Security - Replay attacks possible |
| Token Hash Validation | P1 | Security - Token substitution risk |
| Prompt Parameter | P1 | Mandatory - Spec violation |
| Max Age Enforcement | P1 | Mandatory - Spec violation |
| Authentication Context (ACR) | P2 | Mandatory - Spec violation |
| Hybrid Flows | P2 | Required for Dynamic OPs |
| Request Objects | P2 | Mandatory for Dynamic OPs |

---

## Compliance Gap Analysis

### Mandatory Features for All OpenID Providers (Section 15.1)

| Requirement | Status | Task |
|-------------|--------|------|
| RS256 Signing | ✅ Implemented | - |
| Prompt Parameter | ❌ Missing | P1-5 |
| Display Parameter | ❌ Missing | P2-10 |
| UI/Claims Locales | ❌ Missing | P2-10 |
| Authentication Time | ⚠️ Partial | P1-2, P1-6 |
| Max Age Support | ❌ Missing | P1-6 |
| ACR Values | ❌ Missing | P2-7 |

### Mandatory Features for Dynamic OpenID Providers (Section 15.2)

| Requirement | Status | Task |
|-------------|--------|------|
| Authorization Code Flow | ✅ Implemented | - |
| Implicit Flow | ⚠️ Partial | P1-4 |
| Hybrid Flows | ❌ Missing | P2-8 |
| Discovery | ✅ Implemented | - |
| Dynamic Registration | ❌ Missing | Future |
| UserInfo Endpoint | ✅ Implemented | - |
| Public Keys (Bare JWK) | ✅ Implemented | - |
| Request URI Support | ❌ Missing | P2-12 |

### Critical Security Gaps (Chapter 16)

| Vulnerability | Risk Level | Task |
|---------------|------------|------|
| Nonce Replay Attack | 🔴 High | P1-3 |
| Token Substitution | 🔴 High | P1-4 |
| Incomplete Error Handling | 🟡 Medium | P1-17 |
| TLS Enforcement | 🟡 Medium | P2-18 |

---

## Priority 1: Critical Security & Mandatory Features

### Task 1: Implement Proper Session Management and Consent Flow

**Status:** ⚠️ Critical  
**Spec References:** 3.1.2.3, 3.1.2.4, 11  
**Estimated Effort:** 5-7 days

#### Problem
- Authorization endpoint redirects to `/login` but doesn't store authorization request parameters
- No consent screen after authentication
- No session cookies for authenticated users
- Cannot track consent decisions for offline_access

#### Required Implementation

1. **Session Storage Layer**
   ```go
   type AuthSession struct {
       SessionID        string
       ClientID         string
       RedirectURI      string
       ResponseType     string
       Scope            string
       State            string
       Nonce            string
       CodeChallenge    string
       CodeChallengeMethod string
       Prompt           string
       MaxAge           int
       ACRValues        []string
       Claims           map[string]interface{}
       AuthTime         time.Time
       ConsentGiven     bool
       UserID           string
       CreatedAt        time.Time
       ExpiresAt        time.Time
   }
   ```

2. **Session Store Interface**
   - In-memory implementation (default)
   - Redis implementation (optional)
   - Session cleanup/expiry

3. **Consent Screen**
   - Display requested scopes and claims
   - Show client information
   - Allow/deny decision
   - Remember consent option

4. **Authenticated User Session**
   - HTTP-only secure session cookies
   - Session validation middleware
   - Single sign-on support

#### Acceptance Criteria
- [ ] Authorization requests stored in sessions
- [ ] Consent screen displays after successful authentication
- [ ] User can approve/deny authorization
- [ ] Session cookies enable SSO across multiple auth requests
- [ ] Sessions expire after configurable timeout

---

### Task 2: Fix ID Token Validation and Claims

**Status:** ⚠️ Critical  
**Spec References:** 2, 3.1.3.6, 15.1  
**Estimated Effort:** 2-3 days

#### Problem
- `auth_time` claim missing (mandatory when max_age or acr requested)
- `acr` (Authentication Context Class) claim missing
- `amr` (Authentication Methods References) claim missing
- No tracking of when user authenticated

#### Required Implementation

1. **Track Authentication Time**
   - Store auth_time in session
   - Include in authorization code metadata
   - Return in ID token when required

2. **Authentication Context (ACR)**
   - Define authentication levels (e.g., password, mfa, etc.)
   - Map authentication method to ACR values
   - Include acr claim based on auth method used

3. **Authentication Methods (AMR)**
   - Track authentication methods used (pwd, otp, bio, etc.)
   - Support multiple methods per authentication
   - Include as array in ID token

4. **Update JWT Manager**
   ```go
   type IDTokenClaims struct {
       // ... existing claims
       AuthTime  *int64   `json:"auth_time,omitempty"`
       ACR       string   `json:"acr,omitempty"`
       AMR       []string `json:"amr,omitempty"`
   }
   ```

#### Acceptance Criteria
- [ ] auth_time included when max_age parameter used
- [ ] auth_time included when requested in claims parameter
- [ ] acr claim reflects authentication strength
- [ ] amr array lists authentication methods used
- [ ] ID token validation checks auth_time freshness

---

### Task 3: Implement Proper Nonce Handling and Replay Protection

**Status:** 🔴 Critical Security  
**Spec References:** 3.1.3.7 (step 11), 15.5.2, 16.11  
**Estimated Effort:** 2-3 days

#### Problem
- Nonce passed through parameters but not validated
- No protection against replay attacks
- Authorization codes can be reused
- ID tokens can be replayed

#### Required Implementation

1. **Nonce Storage**
   - Store nonce in authorization code
   - Associate nonce with client session
   - Validate nonce is not reused

2. **Nonce Validation**
   ```go
   // In token endpoint handler
   func validateNonce(authCode *models.AuthorizationCode, idToken *IDToken) error {
       if authCode.Nonce != "" {
           if idToken.Nonce != authCode.Nonce {
               return errors.New("nonce mismatch")
           }
       }
       return nil
   }
   ```

3. **Authorization Code Single-Use**
   - Mark code as used immediately upon exchange
   - Reject reuse attempts
   - Implement code expiry (short-lived: 10 minutes max)

4. **Nonce Generation Guidelines**
   - Minimum 128-bit entropy
   - Cryptographically random
   - Tie to client session

#### Acceptance Criteria
- [ ] Nonce stored in authorization code
- [ ] Nonce validated during token exchange
- [ ] Authorization codes single-use only
- [ ] Proper error responses for nonce mismatch
- [ ] Replay attacks prevented

---

### Task 4: Implement at_hash and c_hash Validation

**Status:** 🔴 Critical Security  
**Spec References:** 3.2.2.9, 3.2.2.10, 3.3.2.10, 3.3.2.11, 16.11  
**Estimated Effort:** 2-3 days

#### Problem
- Missing `at_hash` (access token hash) claim
- Missing `c_hash` (code hash) claim
- Token substitution attacks possible in implicit/hybrid flows

#### Required Implementation

1. **Hash Calculation Function**
   ```go
   func calculateTokenHash(token string, alg string) (string, error) {
       var hash hash.Hash
       switch alg {
       case "RS256", "HS256":
           hash = sha256.New()
       case "RS384", "HS384":
           hash = sha384.New()
       case "RS512", "HS512":
           hash = sha512.New()
       default:
           return "", fmt.Errorf("unsupported algorithm: %s", alg)
       }
       
       hash.Write([]byte(token))
       fullHash := hash.Sum(nil)
       
       // Take left-most half
       halfLength := len(fullHash) / 2
       leftHalf := fullHash[:halfLength]
       
       return base64.RawURLEncoding.EncodeToString(leftHalf), nil
   }
   ```

2. **Include Hashes in ID Tokens**
   - Add `at_hash` when access token returned from authorization endpoint
   - Add `c_hash` when code returned from authorization endpoint
   - Hash based on JWT signing algorithm

3. **Validation in Clients**
   - Verify at_hash matches access token
   - Verify c_hash matches authorization code
   - Reject tokens with mismatched hashes

4. **Flow-Specific Requirements**
   - Implicit Flow (id_token token): REQUIRED at_hash
   - Hybrid Flow (code id_token): REQUIRED c_hash
   - Hybrid Flow (code id_token token): REQUIRED both

#### Acceptance Criteria
- [ ] at_hash included in implicit flow ID tokens
- [ ] c_hash included in hybrid flow ID tokens
- [ ] Hash calculation follows spec (SHA-256, left 128 bits)
- [ ] Validation prevents token substitution
- [ ] Tests cover all flow combinations

---

### Task 5: Add Mandatory Prompt Parameter Support

**Status:** ⚠️ Mandatory  
**Spec References:** 3.1.2.1, 11, 15.1  
**Estimated Effort:** 3-4 days

#### Problem
- `prompt` parameter not handled
- Cannot request re-authentication
- Cannot enforce consent
- Cannot support silent authentication

#### Required Implementation

1. **Prompt Values**
   - `none`: Must not show UI, error if auth needed
   - `login`: Force re-authentication
   - `consent`: Force consent screen
   - `select_account`: Show account selection

2. **Implementation Logic**
   ```go
   func (h *Handlers) handlePrompt(c echo.Context, session *AuthSession) error {
       prompts := strings.Split(c.QueryParam("prompt"), " ")
       
       for _, prompt := range prompts {
           switch prompt {
           case "none":
               if !isUserAuthenticated(session) {
                   return redirectWithError(c, session.RedirectURI, 
                       "login_required", "User not authenticated", session.State)
               }
               if !hasConsent(session) {
                   return redirectWithError(c, session.RedirectURI,
                       "consent_required", "Consent required", session.State)
               }
           case "login":
               // Force re-authentication even if session exists
               clearUserSession(session)
               return redirectToLogin(c, session)
           case "consent":
               // Force consent screen
               session.ConsentGiven = false
               return redirectToConsent(c, session)
           case "select_account":
               return redirectToAccountSelection(c, session)
           }
       }
       return nil
   }
   ```

3. **Error Responses**
   - `login_required`: User must authenticate
   - `consent_required`: Consent must be given
   - `interaction_required`: UI interaction needed
   - `account_selection_required`: Account selection needed

4. **Prompt=none Validation**
   - Check for existing authentication
   - Check for existing consent
   - Return error if interaction required
   - Critical for offline_access requests

#### Acceptance Criteria
- [ ] All prompt values supported
- [ ] prompt=none returns errors without UI
- [ ] prompt=login forces re-authentication
- [ ] prompt=consent shows consent screen
- [ ] Multiple prompt values handled (space-separated)
- [ ] Proper error responses per spec

---

### Task 6: Implement max_age Parameter and auth_time Enforcement

**Status:** ⚠️ Mandatory  
**Spec References:** 3.1.2.1, 15.1  
**Estimated Effort:** 2-3 days

#### Problem
- `max_age` parameter not enforced
- No re-authentication based on authentication age
- `auth_time` not returned in ID tokens

#### Required Implementation

1. **Max Age Validation**
   ```go
   func validateMaxAge(authTime time.Time, maxAge int) bool {
       if maxAge == 0 {
           return false // Equivalent to prompt=login
       }
       
       elapsed := time.Since(authTime).Seconds()
       return elapsed <= float64(maxAge)
   }
   ```

2. **Authorization Endpoint Logic**
   ```go
   maxAge := c.QueryParam("max_age")
   if maxAge != "" {
       maxAgeInt, _ := strconv.Atoi(maxAge)
       if session.AuthTime.IsZero() || !validateMaxAge(session.AuthTime, maxAgeInt) {
           // Force re-authentication
           return redirectToLogin(c, session)
       }
   }
   ```

3. **ID Token Claims**
   - Always include `auth_time` when max_age used
   - Include `auth_time` when requested in claims parameter
   - Use Unix timestamp format

4. **Special Cases**
   - max_age=0 equivalent to prompt=login
   - Combine with prompt parameter validation
   - Client validates auth_time in ID token

#### Acceptance Criteria
- [ ] max_age parameter parsed and validated
- [ ] Re-authentication enforced when max_age exceeded
- [ ] auth_time included in ID tokens when required
- [ ] max_age=0 treated as prompt=login
- [ ] Client-side validation documented

---

### Task 17: Implement Proper Error Responses per Spec

**Status:** ⚠️ Important  
**Spec References:** 3.1.2.6  
**Estimated Effort:** 1-2 days

#### Problem
- Error responses may not match spec format
- Missing error codes: interaction_required, login_required, etc.
- Error descriptions may be inconsistent

#### Required Implementation

1. **Standard Error Codes**
   ```go
   const (
       ErrorInvalidRequest          = "invalid_request"
       ErrorUnauthorizedClient      = "unauthorized_client"
       ErrorAccessDenied            = "access_denied"
       ErrorUnsupportedResponseType = "unsupported_response_type"
       ErrorInvalidScope            = "invalid_scope"
       ErrorServerError             = "server_error"
       ErrorTemporarilyUnavailable  = "temporarily_unavailable"
       
       // OpenID Connect specific
       ErrorInteractionRequired         = "interaction_required"
       ErrorLoginRequired              = "login_required"
       ErrorAccountSelectionRequired   = "account_selection_required"
       ErrorConsentRequired            = "consent_required"
       ErrorInvalidRequestURI          = "invalid_request_uri"
       ErrorInvalidRequestObject       = "invalid_request_object"
       ErrorRequestNotSupported        = "request_not_supported"
       ErrorRequestURINotSupported     = "request_uri_not_supported"
       ErrorRegistrationNotSupported   = "registration_not_supported"
   )
   ```

2. **Error Response Function**
   ```go
   func redirectWithError(c echo.Context, redirectURI, error, errorDescription, state string) error {
       u, _ := url.Parse(redirectURI)
       q := u.Query()
       q.Set("error", error)
       if errorDescription != "" {
           q.Set("error_description", errorDescription)
       }
       if state != "" {
           q.Set("state", state)
       }
       u.RawQuery = q.Encode()
       return c.Redirect(http.StatusFound, u.String())
   }
   ```

3. **Update Error Handling**
   - Authorization endpoint errors
   - Token endpoint errors
   - UserInfo endpoint errors
   - Consistent error format

#### Acceptance Criteria
- [ ] All spec error codes defined
- [ ] Error responses use correct format
- [ ] state parameter preserved in errors
- [ ] Error descriptions are helpful
- [ ] Tests cover all error scenarios

---

## Priority 2: Important Features

### Task 7: Implement acr_values Parameter and Authentication Context

**Status:** ⚠️ Mandatory  
**Spec References:** 3.1.2.1, 5.5.1.1, 15.1  
**Estimated Effort:** 3-4 days

#### Requirements
- Parse acr_values request parameter
- Define authentication levels (e.g., urn:mace:incommon:iap:silver, urn:mace:incommon:iap:bronze)
- Map authentication methods to ACR levels
- Return acr claim in ID token
- Support essential acr claims requests

---

### Task 8: Add Hybrid Flow Support

**Status:** Required for Dynamic OPs  
**Spec References:** 3.3, 15.2  
**Estimated Effort:** 5-7 days

#### Requirements
- Implement `code id_token` response type
- Implement `code token` response type
- Implement `code id_token token` response type
- Return tokens from authorization endpoint (fragment)
- Include c_hash and at_hash in ID tokens
- Update discovery document

---

### Task 9: Implement Claims Parameter

**Status:** Optional but Valuable  
**Spec References:** 5.5, 5.5.1  
**Estimated Effort:** 4-5 days

#### Requirements
- Parse claims request parameter (JSON)
- Support userinfo and id_token member objects
- Handle essential vs voluntary claims
- Support value and values qualifiers
- Filter claims based on request

---

### Task 10: Add Display and UI Locales Support

**Status:** ⚠️ Mandatory  
**Spec References:** 3.1.2.1, 15.1  
**Estimated Effort:** 1-2 days

#### Requirements
- Accept display parameter (page, popup, touch, wap)
- Accept ui_locales parameter
- Accept claims_locales parameter
- Minimum: Don't error when used
- Optional: Actually use for UI rendering

---

### Task 11: Enhance PKCE Validation

**Status:** Current but needs review  
**Spec References:** RFC 7636  
**Estimated Effort:** 1-2 days

#### Requirements
- Make PKCE mandatory for public clients
- Support both S256 and plain methods
- Proper validation in token endpoint
- Clear error messages
- Update discovery to advertise support

---

### Task 12: Add Request and Request_URI Support

**Status:** ⚠️ Mandatory for Dynamic OPs  
**Spec References:** 6, 6.1, 6.2, 15.2  
**Estimated Effort:** 5-7 days

#### Requirements
- Parse request parameter (signed/encrypted JWT)
- Fetch and validate request_uri
- Validate JWT signatures
- Decrypt encrypted requests
- Merge with OAuth parameters
- Update discovery document

---

### Task 18: TLS/HTTPS Enforcement

**Status:** ⚠️ Mandatory  
**Spec References:** 16.17  
**Estimated Effort:** 2-3 days

#### Requirements
- Enforce TLS 1.2+ minimum
- Proper certificate validation
- Strong ciphersuite configuration
- HTTPS redirect middleware
- Configuration options

---

### Task 19: Token Endpoint Client Authentication

**Status:** Current but needs review  
**Spec References:** 9, 3.1.3.1  
**Estimated Effort:** 2-3 days

#### Requirements
- Support client_secret_basic (Authorization header)
- Support client_secret_post (form parameters)
- Proper secret validation
- Clear error messages
- Rate limiting

---

## Priority 3: Optional Enhancements

### Task 13: ID Token Encryption (JWE)

**Spec References:** 10.2, 3.1.3.7  
**Estimated Effort:** 5-7 days

Advanced feature for encrypting ID tokens with client public keys.

---

### Task 14: Pairwise Subject Identifiers

**Spec References:** 8, 8.1, 17.3  
**Estimated Effort:** 3-4 days

Privacy feature to prevent user correlation across clients.

---

### Task 15: Aggregated and Distributed Claims

**Spec References:** 5.6.2  
**Estimated Effort:** 5-7 days

Support for claims from external sources.

---

### Task 16: Advanced Client Authentication

**Spec References:** 9  
**Estimated Effort:** 4-5 days

Add private_key_jwt and client_secret_jwt methods.

---

### Task 20: Comprehensive Audit Logging

**Spec References:** 17.2  
**Estimated Effort:** 3-4 days

Detailed logging for compliance and monitoring.

---

## Implementation Roadmap

### Phase 1: Critical Security (Weeks 1-3)
- ✅ Week 1-2: Tasks 1, 2, 3 (Sessions, Claims, Nonce)
- ✅ Week 2-3: Tasks 4, 5, 6 (Hashes, Prompt, MaxAge)

### Phase 2: Mandatory Features (Weeks 4-6)
- ✅ Week 4: Tasks 17, 19 (Errors, Client Auth)
- ✅ Week 5: Tasks 7, 10, 11 (ACR, Display, PKCE)
- ✅ Week 6: Task 18 (TLS Enforcement)

### Phase 3: Advanced Features (Weeks 7-10)
- ✅ Week 7-8: Task 8 (Hybrid Flows)
- ✅ Week 9: Task 9 (Claims Parameter)
- ✅ Week 10: Task 12 (Request Objects)

### Phase 4: Optional Enhancements (Weeks 11+)
- Tasks 13-16, 20 as needed

---

## Testing & Validation Strategy

### Unit Tests
- [ ] Test each parameter validation function
- [ ] Test session storage operations
- [ ] Test hash calculations
- [ ] Test error response formats
- [ ] Test claim generation logic

### Integration Tests
- [ ] Full authorization code flow
- [ ] Implicit flow with token validation
- [ ] Hybrid flows (all variants)
- [ ] Token refresh flow
- [ ] UserInfo endpoint access

### Compliance Tests
- [ ] Use OpenID Connect test suite
- [ ] Test against official conformance tests
- [ ] Validate with RP test clients
- [ ] Security penetration testing

### Performance Tests
- [ ] Session storage under load
- [ ] Token generation throughput
- [ ] Concurrent authorization requests
- [ ] Database query optimization

---

## References

### Specifications
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
- [OpenID Connect Discovery 1.0](https://openid.net/specs/openid-connect-discovery-1_0.html)
- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [OAuth 2.0 Threat Model RFC 6819](https://tools.ietf.org/html/rfc6819)
- [PKCE RFC 7636](https://tools.ietf.org/html/rfc7636)

### Testing Resources
- [OpenID Connect Conformance Tests](https://openid.net/certification/testing/)
- [OAuth 2.0 Security Best Practices](https://tools.ietf.org/html/draft-ietf-oauth-security-topics)

### Implementation Guides
- [OpenID Connect Basic Client Implementer's Guide](https://openid.net/specs/openid-connect-basic-1_0.html)
- [OpenID Connect Implicit Client Implementer's Guide](https://openid.net/specs/openid-connect-implicit-1_0.html)

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-20 | System | Initial compliance plan created |

---

## Next Steps

1. **Review and Approve Plan**
   - Stakeholder review
   - Timeline validation
   - Resource allocation

2. **Begin Phase 1 Implementation**
   - Start with Task 1 (Session Management)
   - Set up testing framework
   - Create development branch

3. **Establish Metrics**
   - Track compliance percentage
   - Monitor test coverage
   - Measure performance impact

4. **Regular Reviews**
   - Weekly progress updates
   - Bi-weekly compliance checks
   - Monthly security audits
