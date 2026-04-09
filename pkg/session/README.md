# Session Management Implementation

## Overview

This document describes the session management implementation for OpenID Connect compliance in the openid-golang project.

## Architecture

### Session Types

#### 1. **AuthSession** - Authorization Flow Sessions
Stores authorization request parameters during the OAuth2/OIDC authentication flow.

**Fields:**
- `ID`: Unique session identifier
- `ClientID`: OAuth client making the request
- `RedirectURI`: Where to redirect after authorization
- `ResponseType`: OAuth response type (code, token, id_token, etc.)
- `Scope`: Requested OAuth scopes
- `State`: OAuth state parameter for CSRF protection
- `Nonce`: OIDC nonce for replay protection
- `CodeChallenge`: PKCE code challenge
- `CodeChallengeMethod`: PKCE challenge method (S256 or plain)
- `Prompt`: OIDC prompt parameter (none, login, consent, select_account)
- `MaxAge`: Maximum authentication age in seconds
- `ACRValues`: Requested Authentication Context Class References
- `Display`: UI display mode (page, popup, touch, wap)
- `UILocales`: Preferred languages for UI
- `ClaimsLocales`: Preferred languages for claims
- `Claims`: Requested claims (JSON object)
- `UserID`: Associated user after authentication
- `AuthTime`: Time of authentication
- `ConsentGiven`: Whether user gave consent
- `ConsentedScopes`: Scopes user consented to
- `AuthenticationMethod`: Method used to authenticate
- `ACR`: Achieved Authentication Context Class Reference
- `AMR`: Authentication Methods References used
- `ExpiresAt`: Session expiration time
- `CreatedAt`: Session creation time

**Lifetime:** Short-lived (default 10 minutes)

**Purpose:**
- Preserve authorization request parameters across redirects
- Track user authentication and consent
- Support OIDC mandatory parameters (prompt, max_age, etc.)

#### 2. **UserSession** - Authenticated User Sessions
Represents an authenticated user session with persistent cookie.

**Fields:**
- `ID`: Unique session identifier
- `UserID`: Authenticated user ID
- `AuthTime`: Time when user authenticated
- `AuthenticationMethod`: Method used (password, mfa, etc.)
- `ACR`: Authentication Context Class Reference
- `AMR`: Authentication Methods References
- `LastActivityAt`: Last activity timestamp
- `ExpiresAt`: Session expiration time
- `CreatedAt`: Session creation time

**Lifetime:** Long-lived (default 24 hours)

**Purpose:**
- Enable Single Sign-On (SSO) across multiple authorization requests
- Track authentication time for max_age validation
- Store authentication context for ACR/AMR claims

#### 3. **Session** (Legacy)
Simple session model maintained for backward compatibility.

## Storage Backends

### JSON Storage
- Stores sessions in JSON file alongside users, clients, tokens
- Added `auth_sessions` and `user_sessions` maps to JSONData
- Automatic cleanup of expired sessions on save
- Thread-safe with mutex locking

### MongoDB Storage
- Separate collections: `auth_sessions`, `user_sessions`
- TTL indexes on `expires_at` for automatic expiration
- Indexes on `client_id` (auth sessions) and `user_id` (user sessions)
- Efficient bulk cleanup with time-based queries

## Session Middleware

### Configuration

```go
import "github.com/prasenjit-net/openid-golang/pkg/session"

config := session.Config{
    Storage:            storage,                    // Storage backend
    UserSessionTimeout: 24 * time.Hour,            // User session lifetime
    AuthSessionTimeout: 10 * time.Minute,          // Auth session lifetime
    CookieSecure:       true,                      // HTTPS only
    CookieHTTPOnly:     true,                      // No JavaScript access
    CookieSameSite:     http.SameSiteLaxMode,      // CSRF protection
    CookieDomain:       "example.com",             // Cookie domain
    CookiePath:         "/",                       // Cookie path
    CleanupInterval:    1 * time.Hour,             // Cleanup frequency
}

sessionMgr := session.NewManager(config)
```

### Usage in Echo

```go
// Apply middleware to all routes
e.Use(sessionMgr.Middleware())

// Protected routes
admin := e.Group("/admin")
admin.Use(sessionMgr.RequireAuth())

// In handlers
func loginHandler(c echo.Context) error {
    // ... authenticate user ...
    
    // Create user session
    userSession, err := sessionMgr.CreateUserSession(
        c,
        userID,
        "password",              // authentication method
        "urn:mace:incommon:iap:silver", // ACR
        []string{"pwd"},         // AMR
    )
    
    return c.JSON(http.StatusOK, userSession)
}

func authorizeHandler(c echo.Context) error {
    // Create auth session
    authSession, err := sessionMgr.CreateAuthSession(
        c,
        clientID,
        redirectURI,
        responseType,
        scope,
        state,
    )
    
    // Check if user already authenticated
    userSession := session.GetUserSession(c)
    if userSession != nil {
        // User is authenticated, check max_age
        if authSession.MaxAge > 0 {
            if !userSession.IsAuthTimeFresh(authSession.MaxAge) {
                // Re-authentication required
                return c.Redirect(http.StatusFound, "/login")
            }
        }
        
        // Proceed to consent
        return c.Redirect(http.StatusFound, "/consent")
    }
    
    // Redirect to login
    return c.Redirect(http.StatusFound, "/login")
}
```

## OIDC Compliance Features

### ✅ Implemented

1. **Session Management** (Spec 3.1.2.3, 3.1.2.4)
   - Authorization request parameter preservation
   - User authentication tracking
   - Consent tracking and storage

2. **Authentication Time** (Spec 3.1.3.6, 15.1)
   - `auth_time` stored in UserSession
   - Available for ID token claims
   - Used for max_age validation

3. **Max Age Parameter** (Spec 3.1.2.1, 15.1)
   - Stored in AuthSession
   - `IsAuthTimeFresh()` method for validation
   - Forces re-authentication when exceeded

4. **Nonce Parameter** (Spec 3.1.3.7, 15.5.2)
   - Stored in AuthSession
   - Preserved for ID token validation
   - Ready for replay protection implementation

5. **PKCE Parameters** (RFC 7636)
   - code_challenge stored in AuthSession
   - code_challenge_method stored
   - Available for token endpoint validation

6. **Prompt Parameter** (Spec 3.1.2.1, 15.1)
   - Stored in AuthSession
   - Ready for flow control (none, login, consent, select_account)

7. **Display Parameter** (Spec 3.1.2.1, 15.1)
   - Stored in AuthSession
   - Available for UI rendering decisions

8. **Locales Parameters** (Spec 3.1.2.1, 15.1)
   - ui_locales stored in AuthSession
   - claims_locales stored
   - Ready for internationalization

9. **ACR/AMR Support** (Spec 3.1.2.1, 5.5.1.1)
   - acr_values stored in AuthSession (requested)
   - acr stored in UserSession (achieved)
   - amr array stored for authentication methods

10. **Claims Parameter** (Spec 5.5)
    - Stored as JSON in AuthSession
    - Ready for selective claim requests

## Cookie Security

### Settings
- **Secure**: Set to `true` for HTTPS-only transmission
- **HttpOnly**: Set to `true` to prevent JavaScript access
- **SameSite**: Set to `Lax` for CSRF protection
- **Path**: Configurable, default "/"
- **Domain**: Configurable for subdomain support
- **MaxAge**: Matches session timeout

### Cookie Names
- `user_session`: Authenticated user session ID
- `auth_session`: Authorization flow session ID

## Session Lifecycle

### Authorization Flow

```
1. Client → /authorize
   ↓
2. Create AuthSession (store params)
   Set auth_session cookie
   ↓
3. Check UserSession cookie
   ↓
4a. Not authenticated → /login
    ↓
    Authenticate user
    ↓
    Create UserSession
    Set user_session cookie
    ↓
4b. Already authenticated → Continue
    ↓
5. Check consent in AuthSession
   ↓
6a. No consent → /consent
    ↓
    User grants consent
    ↓
    Update AuthSession
    ↓
6b. Consent given → Continue
    ↓
7. Generate authorization code
   Delete AuthSession
   ↓
8. Redirect to client with code
```

### Token Exchange

```
1. Client → /token (with code)
   ↓
2. Validate code
   ↓
3. Load AuthSession from code
   ↓
4. Generate tokens
   Include auth_time from UserSession
   Include acr, amr from UserSession
   ↓
5. Delete authorization code
   ↓
6. Return tokens
```

## Cleanup Strategy

### Automatic Cleanup
- Background goroutine runs at configured interval (default 1 hour)
- Removes expired sessions from storage
- Reduces storage overhead

### MongoDB TTL Indexes
- MongoDB automatically expires documents
- `expires_at` field with TTL index
- No manual cleanup needed (redundant with background)

### Manual Cleanup
```go
err := sessionMgr.store.CleanupExpiredSessions()
```

## Testing

### Unit Tests
```bash
cd backend
go test ./pkg/session/... -v
```

### Test Coverage
- ✅ UserSession.IsAuthenticated()
- ✅ UserSession.IsAuthTimeFresh()
- ✅ generateSessionID() uniqueness
- ⏳ Full integration tests (TODO)

## Migration Guide

### From Simple Sessions

**Before:**
```go
// Old session model
type Session struct {
    ID        string
    UserID    string
    ExpiresAt time.Time
    CreatedAt time.Time
}
```

**After:**
```go
// New models
userSession := &models.UserSession{
    ID:             sessionID,
    UserID:         userID,
    AuthTime:       time.Now(),
    ExpiresAt:      time.Now().Add(24*time.Hour),
    // ... additional OIDC fields
}

authSession := &models.AuthSession{
    ID:           sessionID,
    ClientID:     clientID,
    RedirectURI:  redirectURI,
    // ... OAuth/OIDC parameters
}
```

### Storage Backend Changes

**JSON Storage:**
- Added `auth_sessions` map
- Added `user_sessions` map
- Backward compatible with old `sessions` map

**MongoDB:**
- New collections: `auth_sessions`, `user_sessions`
- Old `sessions` collection still supported
- Run migration or use both in parallel

## Performance Considerations

### Memory Usage
- Each AuthSession: ~500 bytes
- Each UserSession: ~300 bytes
- JSON storage loads all in memory
- MongoDB storage: document-based, scalable

### Recommended Limits
- JSON storage: < 10,000 concurrent sessions
- MongoDB: Unlimited (production-ready)

### Optimization Tips
1. Set appropriate session timeouts
2. Enable automatic cleanup
3. Use MongoDB for high traffic
4. Consider Redis for ultra-high performance (future)

## Security Best Practices

1. **Always use HTTPS in production** (`CookieSecure: true`)
2. **Set HttpOnly on all session cookies**
3. **Use SameSite=Lax or Strict for CSRF protection**
4. **Implement short auth session timeouts** (10 minutes)
5. **Rotate session IDs after authentication**
6. **Clean up sessions on logout**
7. **Monitor for suspicious session patterns**
8. **Use secure random for session ID generation**

## Future Enhancements

### Planned
- [ ] Redis storage backend for high performance
- [ ] Session fixation protection (ID rotation)
- [ ] Remember me functionality
- [ ] Device tracking and management
- [ ] Suspicious activity detection
- [ ] Rate limiting per session
- [ ] Session analytics and metrics

### Under Consideration
- [ ] Distributed session management
- [ ] Session encryption at rest
- [ ] Multi-factor session validation
- [ ] Session delegation/impersonation (admin)

## References

- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
- Section 3.1.2.1: Authentication Request
- Section 3.1.2.3: Authentication Response
- Section 3.1.2.4: Consent
- Section 11: Offline Access (consent)
- Section 15.1: Mandatory Features
- [RFC 6749: OAuth 2.0](https://tools.ietf.org/html/rfc6749)
- [RFC 7636: PKCE](https://tools.ietf.org/html/rfc7636)

## Related Documentation

- [OIDC_COMPLIANCE_PLAN.md](../../docs/OIDC_COMPLIANCE_PLAN.md) - Full compliance roadmap
- [ARCHITECTURE.md](../../docs/ARCHITECTURE.md) - System architecture
- [STORAGE.md](../../docs/STORAGE.md) - Storage backend details
