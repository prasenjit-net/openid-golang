# Front-Channel Logout Implementation Plan

**Document Version:** 1.0  
**Date:** October 24, 2025  
**Status:** Planning Phase  
**Specification:** [OpenID Connect Front-Channel Logout 1.0](https://openid.net/specs/openid-connect-frontchannel-1_0.html)

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Specification Overview](#specification-overview)
- [Current Implementation Status](#current-implementation-status)
- [Implementation Tasks](#implementation-tasks)
- [API Specifications](#api-specifications)
- [Security Considerations](#security-considerations)
- [Testing Strategy](#testing-strategy)

---

## Executive Summary

This document outlines the plan to implement **OpenID Connect Front-Channel Logout** in the openid-golang project. Front-Channel Logout enables an OpenID Provider (OP) to notify Relying Parties (RPs) when a user logs out, allowing them to clean up their sessions.

**Key Features:**
- Propagate logout events to all logged-in RPs
- Use browser redirects (front-channel) for logout notifications
- Maintain user privacy and security
- Support both iframe and redirect-based logout methods

**Estimated Effort:** 5-7 days

---

## Specification Overview

### What is Front-Channel Logout?

Front-Channel Logout allows the OP to notify RPs about logout events using the user's browser (front-channel) rather than back-channel server-to-server communication.

### How It Works:

1. User initiates logout at the OP
2. OP renders a logout page with hidden iframes
3. Each iframe loads the RP's `frontchannel_logout_uri`
4. RP receives logout notification and clears local session
5. User is redirected to `post_logout_redirect_uri` (optional)

### Key Parameters:

**Client Registration:**
- `frontchannel_logout_uri` - URI where RP receives logout notifications
- `frontchannel_logout_session_required` - Whether to include session parameters

**Logout Request:**
- `iss` (Issuer) - OP's issuer identifier
- `sid` (Session ID) - Session identifier (if `frontchannel_logout_session_required=true`)

---

## Current Implementation Status

### ✅ Already Implemented

| Feature | Status | Location |
|---------|--------|----------|
| User Sessions | ✅ Complete | `pkg/models/models.go` - UserSession |
| Session Storage | ✅ Complete | `pkg/storage/*.go` |
| Session Middleware | ✅ Complete | `pkg/session/middleware.go` |
| Login Handler | ✅ Complete | `pkg/handlers/handlers.go` |
| Client Model | ✅ Complete | `pkg/models/models.go` - Client |

### ❌ Missing Features

| Feature | Priority | Impact |
|---------|----------|--------|
| Front-Channel Logout URIs in Client | 🔴 High | Required for logout notifications |
| Logout Endpoint | 🔴 High | Entry point for logout |
| Logout Page with iframes | 🔴 High | Notifies all RPs |
| Session ID (sid) claim | 🟡 Medium | For session-specific logout |
| Discovery metadata | 🟡 Medium | Advertise logout support |

---

## Implementation Tasks

### Phase 1: Enhance Client Model (1 day)

#### Task 1.1: Add Front-Channel Logout Fields to Client Model
**File:** `backend/pkg/models/models.go`

```go
type Client struct {
    // ... existing fields ...
    
    // Front-Channel Logout
    FrontChannelLogoutURI            string `json:"frontchannel_logout_uri,omitempty" bson:"frontchannel_logout_uri,omitempty"`
    FrontChannelLogoutSessionRequired bool   `json:"frontchannel_logout_session_required,omitempty" bson:"frontchannel_logout_session_required,omitempty"`
}
```

#### Task 1.2: Update Client Registration
**File:** `backend/pkg/handlers/registration.go`

Add validation for `frontchannel_logout_uri`:
- Must be HTTPS (except localhost)
- Must be absolute URI
- Must not contain fragment
- Should match client's redirect URI domain (security)

```go
func validateFrontChannelLogoutURI(uri string) error {
    if uri == "" {
        return nil // Optional field
    }
    
    parsedURI, err := url.Parse(uri)
    if err != nil {
        return errors.New("invalid frontchannel_logout_uri")
    }
    
    if !parsedURI.IsAbs() {
        return errors.New("frontchannel_logout_uri must be absolute")
    }
    
    if parsedURI.Fragment != "" {
        return errors.New("frontchannel_logout_uri must not contain fragment")
    }
    
    if parsedURI.Scheme != "https" && !isLocalhost(parsedURI.Host) {
        return errors.New("frontchannel_logout_uri must use HTTPS")
    }
    
    return nil
}
```

---

### Phase 2: Enhance Session Tracking (1-2 days)

#### Task 2.1: Add Session ID (sid) to Tokens
**File:** `backend/pkg/models/models.go`

```go
type Token struct {
    // ... existing fields ...
    SessionID string `json:"session_id,omitempty" bson:"session_id,omitempty"`
}
```

#### Task 2.2: Add sid Claim to ID Tokens
**File:** `backend/pkg/handlers/token.go`

```go
// When generating ID token
claims := map[string]interface{}{
    "iss": config.Issuer,
    "sub": user.ID,
    "aud": client.ID,
    "exp": time.Now().Add(time.Hour).Unix(),
    "iat": time.Now().Unix(),
    "nonce": authCode.Nonce,
    "sid": authCode.SessionID, // Add session ID
    // ... other claims
}
```

#### Task 2.3: Track Client Sessions
**File:** `backend/pkg/models/models.go`

```go
// ClientSession tracks which clients a user has logged into
type ClientSession struct {
    UserID    string    `json:"user_id" bson:"user_id"`
    ClientID  string    `json:"client_id" bson:"client_id"`
    SessionID string    `json:"session_id" bson:"session_id"`
    CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
```

Add storage methods:
```go
CreateClientSession(session *ClientSession) error
GetClientSessionsByUserSession(sessionID string) ([]*ClientSession, error)
DeleteClientSession(sessionID, clientID string) error
DeleteClientSessionsByUserSession(sessionID string) error
```

---

### Phase 3: Implement Logout Endpoint (2-3 days)

#### Task 3.1: Create Logout Handler
**File:** `backend/pkg/handlers/logout.go`

```go
// Logout handles the logout request
// Implements OpenID Connect RP-Initiated Logout 1.0 and Front-Channel Logout 1.0
func (h *Handlers) Logout(c echo.Context) error {
    // 1. Parse logout request
    idTokenHint := c.QueryParam("id_token_hint")
    postLogoutRedirectURI := c.QueryParam("post_logout_redirect_uri")
    state := c.QueryParam("state")
    
    // 2. Validate id_token_hint (optional but recommended)
    var userID string
    var sessionID string
    if idTokenHint != "" {
        claims, err := h.validateIDToken(idTokenHint)
        if err == nil {
            userID = claims["sub"].(string)
            if sid, ok := claims["sid"].(string); ok {
                sessionID = sid
            }
        }
    }
    
    // 3. Get user session from cookie
    if userID == "" {
        sessionCookie, err := c.Cookie("session_id")
        if err == nil {
            userSession, _ := h.storage.GetUserSession(sessionCookie.Value)
            if userSession != nil {
                userID = userSession.UserID
                sessionID = sessionCookie.Value
            }
        }
    }
    
    // 4. Get all client sessions for this user session
    clientSessions, err := h.storage.GetClientSessionsByUserSession(sessionID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to retrieve sessions",
        })
    }
    
    // 5. Build list of logout URIs
    logoutURIs := []LogoutTarget{}
    for _, cs := range clientSessions {
        client, err := h.storage.GetClientByID(cs.ClientID)
        if err != nil || client == nil || client.FrontChannelLogoutURI == "" {
            continue
        }
        
        target := LogoutTarget{
            URI:                      client.FrontChannelLogoutURI,
            SessionRequired:          client.FrontChannelLogoutSessionRequired,
            SessionID:                cs.SessionID,
            ClientID:                 cs.ClientID,
        }
        logoutURIs = append(logoutURIs, target)
    }
    
    // 6. Delete all sessions
    if sessionID != "" {
        _ = h.storage.DeleteUserSession(sessionID)
        _ = h.storage.DeleteClientSessionsByUserSession(sessionID)
        
        // Clear session cookie
        c.SetCookie(&http.Cookie{
            Name:     "session_id",
            Value:    "",
            Path:     "/",
            MaxAge:   -1,
            HttpOnly: true,
            Secure:   true,
            SameSite: http.SameSiteLaxMode,
        })
    }
    
    // 7. Render logout page with iframes
    return c.Render(http.StatusOK, "logout.html", map[string]interface{}{
        "LogoutTargets":           logoutURIs,
        "PostLogoutRedirectURI":   postLogoutRedirectURI,
        "State":                   state,
        "Issuer":                  h.config.Issuer,
    })
}

type LogoutTarget struct {
    URI             string
    SessionRequired bool
    SessionID       string
    ClientID        string
}
```

#### Task 3.2: Create Logout Page Template
**File:** `backend/pkg/ui/logout.html`

```html
<!DOCTYPE html>
<html>
<head>
    <title>Logging out...</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            text-align: center;
            padding: 50px;
        }
        .spinner {
            border: 4px solid #f3f3f3;
            border-top: 4px solid #3498db;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 20px auto;
        }
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <h1>Logging out...</h1>
    <div class="spinner"></div>
    <p>Please wait while we log you out from all applications.</p>
    
    <!-- Hidden iframes for front-channel logout -->
    {{range .LogoutTargets}}
    <iframe 
        style="display:none;" 
        src="{{.URI}}?iss={{$.Issuer}}{{if .SessionRequired}}&sid={{.SessionID}}{{end}}"
        onload="logoutComplete('{{.ClientID}}')"
    ></iframe>
    {{end}}
    
    <script>
        let completedLogouts = 0;
        const totalLogouts = {{len .LogoutTargets}};
        const postLogoutURI = "{{.PostLogoutRedirectURI}}";
        const state = "{{.State}}";
        
        function logoutComplete(clientId) {
            completedLogouts++;
            console.log('Logout complete for client:', clientId);
            
            if (completedLogouts >= totalLogouts) {
                // All logouts complete
                setTimeout(function() {
                    if (postLogoutURI) {
                        let redirectURL = postLogoutURI;
                        if (state) {
                            redirectURL += (postLogoutURI.includes('?') ? '&' : '?') + 'state=' + encodeURIComponent(state);
                        }
                        window.location.href = redirectURL;
                    } else {
                        // Default logout success page
                        document.body.innerHTML = '<h1>You have been logged out</h1><p>You can now close this window.</p>';
                    }
                }, 1000);
            }
        }
        
        // Timeout after 5 seconds
        setTimeout(function() {
            if (completedLogouts < totalLogouts) {
                console.warn('Some logouts did not complete');
                if (postLogoutURI) {
                    let redirectURL = postLogoutURI;
                    if (state) {
                        redirectURL += (postLogoutURI.includes('?') ? '&' : '?') + 'state=' + encodeURIComponent(state);
                    }
                    window.location.href = redirectURL;
                } else {
                    document.body.innerHTML = '<h1>Logout complete</h1><p>Some applications may still have you logged in.</p>';
                }
            }
        }, 5000);
    </script>
</body>
</html>
```

---

### Phase 4: Update Discovery Document (0.5 days)

#### Task 4.1: Add Front-Channel Logout Metadata
**File:** `backend/pkg/handlers/discovery.go`

```go
response := map[string]interface{}{
    // ... existing fields ...
    
    // Front-Channel Logout
    "frontchannel_logout_supported":          true,
    "frontchannel_logout_session_supported":  true,
    
    // RP-Initiated Logout
    "end_session_endpoint": h.config.Issuer + "/logout",
}
```

---

### Phase 5: Testing & Documentation (1-2 days)

#### Task 5.1: Unit Tests
**File:** `backend/pkg/handlers/logout_test.go`

Test cases:
- Logout with valid id_token_hint
- Logout with session cookie
- Logout with multiple client sessions
- Logout with post_logout_redirect_uri
- Logout without any sessions
- Invalid id_token_hint
- Invalid post_logout_redirect_uri

#### Task 5.2: Integration Tests

Create test client that:
1. Registers with `frontchannel_logout_uri`
2. Logs in user
3. Initiates logout
4. Verifies logout notification received
5. Verifies session cleared

#### Task 5.3: Documentation
**File:** `docs/FRONT_CHANNEL_LOGOUT.md`

Document:
- How to configure front-channel logout
- Client registration parameters
- Logout endpoint usage
- Security considerations
- Example implementations

---

## API Specifications

### Client Registration

**Request:**
```json
POST /register
Content-Type: application/json

{
  "client_name": "Example App",
  "redirect_uris": ["https://example.com/callback"],
  "frontchannel_logout_uri": "https://example.com/logout",
  "frontchannel_logout_session_required": true
}
```

**Response:**
```json
{
  "client_id": "client_123",
  "client_secret": "secret_abc",
  "frontchannel_logout_uri": "https://example.com/logout",
  "frontchannel_logout_session_required": true
}
```

### Logout Endpoint

**Request:**
```http
GET /logout?id_token_hint=<token>&post_logout_redirect_uri=https://example.com&state=xyz
```

**Parameters:**
- `id_token_hint` (RECOMMENDED) - Previously issued ID Token
- `post_logout_redirect_uri` (OPTIONAL) - Where to redirect after logout
- `state` (OPTIONAL) - Opaque value to maintain state

**Response:**
HTML page with iframes that trigger logout at all RPs.

### Front-Channel Logout Notification

**Request to RP:**
```http
GET https://example.com/logout?iss=https://op.example.com&sid=session_123
```

**Parameters:**
- `iss` (REQUIRED) - Issuer identifier
- `sid` (OPTIONAL) - Session ID (if `frontchannel_logout_session_required=true`)

---

## Security Considerations

### 1. URI Validation
- Logout URIs MUST use HTTPS (except localhost)
- Validate URI matches client's domain
- Prevent open redirects

### 2. Session Security
- Use HttpOnly, Secure cookies
- Implement CSRF protection
- Validate id_token_hint signature

### 3. Privacy
- Don't leak user information in logout URIs
- Limit session ID visibility
- Use same-origin iframes when possible

### 4. Timeout Protection
- Implement JavaScript timeout (5-10 seconds)
- Don't block logout on RP failures
- Log failed logout notifications

### 5. Post-Logout Redirect Validation
- Validate `post_logout_redirect_uri` against registered URIs
- Use whitelist approach
- Prevent XSS attacks

---

## Testing Strategy

### Unit Tests
- ✅ Client model validation
- ✅ Logout URI validation
- ✅ Session tracking
- ✅ ID token validation
- ✅ Logout handler logic

### Integration Tests
- ✅ Full logout flow with multiple RPs
- ✅ Session cleanup verification
- ✅ iframe loading and callbacks
- ✅ Post-logout redirect

### Manual Testing
- ✅ Test with real browser
- ✅ Multiple tabs/windows
- ✅ Network failure scenarios
- ✅ Cross-browser compatibility

### Security Testing
- ✅ CSRF protection
- ✅ XSS prevention
- ✅ Session fixation
- ✅ Open redirect prevention

---

## Timeline

| Phase | Tasks | Duration | Dependencies |
|-------|-------|----------|--------------|
| Phase 1 | Client model enhancement | 1 day | None |
| Phase 2 | Session tracking | 1-2 days | Phase 1 |
| Phase 3 | Logout endpoint | 2-3 days | Phase 2 |
| Phase 4 | Discovery update | 0.5 day | Phase 3 |
| Phase 5 | Testing & docs | 1-2 days | Phase 4 |
| **Total** | | **5.5-8.5 days** | |

---

## Success Criteria

1. ✅ Clients can register with `frontchannel_logout_uri`
2. ✅ Logout endpoint notifies all logged-in RPs
3. ✅ Sessions are properly cleaned up
4. ✅ Discovery document advertises logout support
5. ✅ All tests passing
6. ✅ Security best practices followed
7. ✅ Documentation complete

---

## References

- [OpenID Connect Front-Channel Logout 1.0](https://openid.net/specs/openid-connect-frontchannel-1_0.html)
- [OpenID Connect RP-Initiated Logout 1.0](https://openid.net/specs/openid-connect-rpinitiated-1_0.html)
- [OpenID Connect Session Management 1.0](https://openid.net/specs/openid-connect-session-1_0.html)
