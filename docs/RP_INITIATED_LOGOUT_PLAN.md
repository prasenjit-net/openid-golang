# RP-Initiated Logout Implementation Plan

**Document Version:** 1.0  
**Date:** October 24, 2025  
**Status:** Planning Phase  
**Specification:** [OpenID Connect RP-Initiated Logout 1.0](https://openid.net/specs/openid-connect-rpinitiated-1_0.html)

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Specification Overview](#specification-overview)
- [Relationship to Other Logout Mechanisms](#relationship-to-other-logout-mechanisms)
- [Current Implementation Status](#current-implementation-status)
- [Implementation Tasks](#implementation-tasks)
- [API Specifications](#api-specifications)
- [Security Considerations](#security-considerations)
- [Testing Strategy](#testing-strategy)

---

## Executive Summary

This document outlines the plan to implement **OpenID Connect RP-Initiated Logout** in the openid-golang project. RP-Initiated Logout allows a Relying Party (RP) to request that an OpenID Provider (OP) log out an End-User, optionally redirecting the user back to the RP after logout.

**Key Features:**
- RP can initiate logout at the OP
- Optional `id_token_hint` for user identification
- Optional `post_logout_redirect_uri` for return to RP
- State parameter for CSRF protection
- Coordinated logout with all RPs via front-channel or back-channel
- Clean, user-friendly logout experience

**Estimated Effort:** 3-5 days

---

## Specification Overview

### What is RP-Initiated Logout?

RP-Initiated Logout enables a Relying Party to notify the OpenID Provider when an End-User has logged out of the RP. The OP then:
1. Ends the user's session at the OP
2. Notifies other RPs where the user is logged in (via front-channel or back-channel)
3. Optionally redirects the user back to the RP

### How It Works:

```
┌────────┐                  ┌─────────────┐                    ┌────────┐
│   RP   │                  │     OP      │                    │ Other  │
│        │                  │             │                    │  RPs   │
└───┬────┘                  └──────┬──────┘                    └───┬────┘
    │                              │                               │
    │ 1. User clicks logout        │                               │
    │                              │                               │
    │ 2. Redirect to               │                               │
    │    /logout?id_token_hint=... │                               │
    ├─────────────────────────────>│                               │
    │                              │                               │
    │                              │ 3. Validate id_token_hint     │
    │                              │                               │
    │                              │ 4. End OP session             │
    │                              │                               │
    │                              │ 5. Notify other RPs           │
    │                              ├──────────────────────────────>│
    │                              │   (front-channel/back-channel)│
    │                              │                               │
    │                              │<──────────────────────────────┤
    │                              │ 6. RPs clear sessions         │
    │                              │                               │
    │ 7. Redirect back to RP       │                               │
    │<─────────────────────────────┤                               │
    │    (if post_logout_redirect) │                               │
    │                              │                               │
    │ 8. Display "Logged out"      │                               │
    │                              │                               │
```

### Request Parameters:

| Parameter | Required | Description |
|-----------|----------|-------------|
| `id_token_hint` | RECOMMENDED | Previously issued ID Token |
| `post_logout_redirect_uri` | OPTIONAL | Where to redirect after logout |
| `state` | OPTIONAL | Opaque value for CSRF protection |
| `ui_locales` | OPTIONAL | Preferred languages for logout UI |

### Key Benefits:

✅ **Single Sign-Out (SSO)** - Logout from one RP logs out from all  
✅ **User Control** - User initiates logout from familiar RP interface  
✅ **Clean Experience** - Seamless redirect back to RP  
✅ **CSRF Protection** - State parameter prevents attacks  
✅ **Multi-RP Support** - Works with front-channel and back-channel logout  

---

## Relationship to Other Logout Mechanisms

### The Complete Logout Picture

```
┌─────────────────────────────────────────────────────────────┐
│                    RP-Initiated Logout                      │
│              (This Implementation Plan)                     │
│                                                             │
│  User clicks logout at RP → RP redirects to OP /logout     │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│              OP Logout Endpoint (/logout)                   │
│                                                             │
│  • Validates id_token_hint                                  │
│  • Ends user's OP session                                   │
│  • Decides how to notify RPs                                │
└──────┬───────────────────────────────┬──────────────────────┘
       │                               │
       ↓                               ↓
┌──────────────────┐          ┌──────────────────────┐
│  Front-Channel   │          │   Back-Channel       │
│     Logout       │          │      Logout          │
│                  │          │                      │
│ Browser iframes  │          │  Server-to-server    │
│ to RP logout URIs│          │  POST with JWT       │
└──────────────────┘          └──────────────────────┘
```

### Integration Points:

| Component | Status | Integration |
|-----------|--------|-------------|
| **RP-Initiated Logout** | 🔴 NEW | Entry point - this plan |
| **Front-Channel Logout** | ✅ Planned | Called by RP-Initiated Logout |
| **Back-Channel Logout** | ✅ Planned | Called by RP-Initiated Logout |
| **Logout Endpoint** | 🟡 Partial | Needs enhancement for RP-initiated flow |

### Comparison Matrix:

| Aspect | RP-Initiated | Front-Channel | Back-Channel |
|--------|--------------|---------------|--------------|
| **Initiator** | Relying Party | OP directly | OP directly |
| **Entry Point** | RP's logout button | OP's logout button | OP's logout button |
| **User Flow** | RP → OP → back to RP | Stays at OP | Stays at OP |
| **Use Case** | User logs out from app | User logs out from OP | User logs out from OP |
| **Redirect** | Yes (back to RP) | Optional | Optional |
| **CSRF Protection** | State parameter | Not needed | Not needed |

---

## Current Implementation Status

### ✅ Already Implemented

| Feature | Status | Location | Notes |
|---------|--------|----------|-------|
| User Sessions | ✅ Complete | `pkg/models/models.go` | UserSession model |
| Session Storage | ✅ Complete | `pkg/storage/*.go` | CRUD operations |
| Client Model | ✅ Complete | `pkg/models/models.go` | Full OIDC metadata |
| ID Token Generation | ✅ Complete | `pkg/crypto/jwt.go` | With all required claims |
| Session Middleware | ✅ Complete | `pkg/session/middleware.go` | Session management |

### 🟡 Partially Implemented

| Feature | Status | Location | What's Missing |
|---------|--------|----------|----------------|
| Logout Endpoint | 🟡 Partial | `pkg/handlers/logout.go` | Needs RP-initiated flow |
| Discovery Document | 🟡 Partial | `pkg/handlers/discovery.go` | Needs `end_session_endpoint` |
| Client Sessions | 🟡 Partial | Need to track | Which RPs user logged into |

### ❌ Missing Features

| Feature | Priority | Impact |
|---------|----------|--------|
| RP-Initiated Logout Flow | 🔴 Critical | Core feature |
| Post-Logout Redirect Validation | 🔴 Critical | Security requirement |
| State Parameter Support | 🔴 Critical | CSRF protection |
| Client Session Tracking | 🟡 High | For multi-RP logout |
| Logout Consent Screen | 🟢 Low | Optional feature |

---

## Implementation Tasks

### Phase 1: Enhance Client Model (0.5 days)

#### Task 1.1: Add Post-Logout Redirect URIs to Client
**File:** `backend/pkg/models/models.go`

```go
type Client struct {
    // ... existing fields ...
    
    // RP-Initiated Logout
    PostLogoutRedirectURIs []string `json:"post_logout_redirect_uris,omitempty" bson:"post_logout_redirect_uris,omitempty"`
}
```

#### Task 1.2: Update Client Registration Validation
**File:** `backend/pkg/handlers/registration.go`

```go
func validatePostLogoutRedirectURIs(uris []string) error {
    for _, uri := range uris {
        parsedURI, err := url.Parse(uri)
        if err != nil {
            return fmt.Errorf("invalid post_logout_redirect_uri: %s", uri)
        }
        
        if !parsedURI.IsAbs() {
            return errors.New("post_logout_redirect_uri must be absolute")
        }
        
        if parsedURI.Fragment != "" {
            return errors.New("post_logout_redirect_uri must not contain fragment")
        }
        
        // Should use HTTPS except for localhost
        if parsedURI.Scheme != "https" && !isLocalhost(parsedURI.Host) {
            return errors.New("post_logout_redirect_uri must use HTTPS")
        }
    }
    return nil
}

// Add to Register() function
if len(req.PostLogoutRedirectURIs) > 0 {
    if err := validatePostLogoutRedirectURIs(req.PostLogoutRedirectURIs); err != nil {
        return errorResponse(c, ErrInvalidClientMetadata, err.Error())
    }
}
```

---

### Phase 2: Implement Logout Endpoint (2-3 days)

#### Task 2.1: Enhance Logout Handler
**File:** `backend/pkg/handlers/logout.go`

```go
// Logout handles RP-Initiated Logout per OpenID Connect RP-Initiated Logout 1.0
func (h *Handlers) Logout(c echo.Context) error {
    // 1. Parse logout request parameters
    idTokenHint := c.QueryParam("id_token_hint")
    postLogoutRedirectURI := c.QueryParam("post_logout_redirect_uri")
    state := c.QueryParam("state")
    uiLocales := c.QueryParam("ui_locales")
    
    var userID string
    var sessionID string
    var clientID string
    
    // 2. Validate and extract information from id_token_hint (RECOMMENDED)
    if idTokenHint != "" {
        claims, err := h.validateIDTokenHint(idTokenHint)
        if err != nil {
            // Log error but continue - id_token_hint is optional
            log.Printf("Invalid id_token_hint: %v", err)
        } else {
            userID = claims["sub"].(string)
            if aud, ok := claims["aud"].(string); ok {
                clientID = aud
            }
            if sid, ok := claims["sid"].(string); ok {
                sessionID = sid
            }
        }
    }
    
    // 3. Get user session from cookie if not identified via token
    if userID == "" {
        sessionCookie, err := c.Cookie("session_id")
        if err == nil {
            userSession, err := h.storage.GetUserSession(sessionCookie.Value)
            if err == nil && userSession != nil {
                userID = userSession.UserID
                sessionID = sessionCookie.Value
            }
        }
    }
    
    // 4. Validate post_logout_redirect_uri if provided
    var validatedRedirectURI string
    if postLogoutRedirectURI != "" && clientID != "" {
        client, err := h.storage.GetClientByID(clientID)
        if err != nil || client == nil {
            // Invalid client - ignore redirect URI
            log.Printf("Cannot validate redirect URI: client not found")
        } else {
            // Validate that redirect URI is registered for this client
            if isValidPostLogoutRedirectURI(postLogoutRedirectURI, client.PostLogoutRedirectURIs) {
                validatedRedirectURI = postLogoutRedirectURI
            } else {
                log.Printf("Invalid post_logout_redirect_uri: not registered for client %s", clientID)
            }
        }
    }
    
    // 5. Get all client sessions for this user
    var clientSessions []*models.ClientSession
    if sessionID != "" {
        clientSessions, _ = h.storage.GetClientSessionsByUserSession(sessionID)
    }
    
    // 6. Prepare logout notifications
    var frontChannelTargets []LogoutTarget
    var backChannelCount int
    
    for _, cs := range clientSessions {
        client, err := h.storage.GetClientByID(cs.ClientID)
        if err != nil || client == nil {
            continue
        }
        
        // Back-Channel Logout (async)
        if client.BackChannelLogoutURI != "" {
            err := h.sendBackChannelLogout(client, cs, sessionID)
            if err != nil {
                log.Printf("Back-channel logout failed for client %s: %v", client.ID, err)
            } else {
                backChannelCount++
            }
        }
        
        // Front-Channel Logout (via iframes)
        if client.FrontChannelLogoutURI != "" {
            target := LogoutTarget{
                URI:             client.FrontChannelLogoutURI,
                SessionRequired: client.FrontChannelLogoutSessionRequired,
                SessionID:       cs.SessionID,
                ClientID:        cs.ClientID,
            }
            frontChannelTargets = append(frontChannelTargets, target)
        }
    }
    
    // 7. Delete all sessions immediately
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
    
    // 8. Determine logout flow based on redirect URI and logout targets
    if validatedRedirectURI != "" && len(frontChannelTargets) == 0 {
        // Direct redirect - no front-channel notifications needed
        redirectURL := buildRedirectURL(validatedRedirectURI, state)
        return c.Redirect(http.StatusFound, redirectURL)
    }
    
    if len(frontChannelTargets) > 0 {
        // Render logout page with iframes (front-channel notifications)
        return c.Render(http.StatusOK, "logout.html", map[string]interface{}{
            "LogoutTargets":         frontChannelTargets,
            "PostLogoutRedirectURI": validatedRedirectURI,
            "State":                 state,
            "Issuer":                h.config.Issuer,
            "UILocales":             uiLocales,
        })
    }
    
    // 9. No redirect URI and no front-channel targets
    if validatedRedirectURI != "" {
        redirectURL := buildRedirectURL(validatedRedirectURI, state)
        return c.Redirect(http.StatusFound, redirectURL)
    }
    
    // Default: Show logout confirmation page
    return c.Render(http.StatusOK, "logout_complete.html", map[string]interface{}{
        "Issuer": h.config.Issuer,
    })
}

// validateIDTokenHint validates and parses the ID token hint
func (h *Handlers) validateIDTokenHint(tokenString string) (jwt.MapClaims, error) {
    // Parse token
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Verify signing method
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        
        // Return public key for verification
        return h.signingKey.Public(), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if !token.Valid {
        return nil, errors.New("invalid token")
    }
    
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid claims")
    }
    
    // Verify issuer
    if iss, ok := claims["iss"].(string); !ok || iss != h.config.Issuer {
        return nil, errors.New("invalid issuer")
    }
    
    // Check expiration (allow some grace period for logout)
    if exp, ok := claims["exp"].(float64); ok {
        if time.Now().Unix() > int64(exp)+300 { // 5 minute grace period
            return nil, errors.New("token expired")
        }
    }
    
    return claims, nil
}

// isValidPostLogoutRedirectURI checks if URI is registered for the client
func isValidPostLogoutRedirectURI(uri string, registeredURIs []string) bool {
    for _, registered := range registeredURIs {
        if uri == registered {
            return true
        }
    }
    return false
}

// buildRedirectURL constructs the final redirect URL with state parameter
func buildRedirectURL(baseURI, state string) string {
    if state == "" {
        return baseURI
    }
    
    separator := "?"
    if strings.Contains(baseURI, "?") {
        separator = "&"
    }
    
    return fmt.Sprintf("%s%sstate=%s", baseURI, separator, url.QueryEscape(state))
}
```

#### Task 2.2: Create Logout Complete Template
**File:** `backend/pkg/ui/logout_complete.html`

```html
<!DOCTYPE html>
<html>
<head>
    <title>Logged Out - {{.Issuer}}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
            padding: 40px;
            max-width: 500px;
            text-align: center;
        }
        .icon {
            width: 80px;
            height: 80px;
            background: #10B981;
            border-radius: 50%;
            display: flex;
            justify-content: center;
            align-items: center;
            margin: 0 auto 20px;
        }
        .icon svg {
            width: 50px;
            height: 50px;
            fill: white;
        }
        h1 {
            color: #1F2937;
            margin: 0 0 10px;
            font-size: 28px;
        }
        p {
            color: #6B7280;
            margin: 0 0 30px;
            line-height: 1.6;
        }
        .info {
            background: #F3F4F6;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
        }
        .info strong {
            color: #1F2937;
            display: block;
            margin-bottom: 10px;
        }
        .info ul {
            text-align: left;
            margin: 10px 0;
            padding-left: 20px;
        }
        .info li {
            color: #6B7280;
            margin-bottom: 5px;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #E5E7EB;
            color: #9CA3AF;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">
            <svg viewBox="0 0 24 24">
                <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
        </div>
        
        <h1>You've been logged out</h1>
        <p>You have been successfully logged out from {{.Issuer}} and all connected applications.</p>
        
        <div class="info">
            <strong>What happened?</strong>
            <ul>
                <li>Your session at the identity provider has been terminated</li>
                <li>You've been logged out from all connected applications</li>
                <li>Your tokens have been invalidated</li>
            </ul>
        </div>
        
        <p>You can now safely close this window or tab.</p>
        
        <div class="footer">
            Powered by {{.Issuer}}
        </div>
    </div>
</body>
</html>
```

#### Task 2.3: Update Existing Logout Template
**File:** `backend/pkg/ui/logout.html`

Enhance to support post_logout_redirect_uri:

```html
<!-- Add to existing logout.html -->
<script>
    const postLogoutURI = "{{.PostLogoutRedirectURI}}";
    const state = "{{.State}}";
    
    function redirectAfterLogout() {
        if (postLogoutURI) {
            let redirectURL = postLogoutURI;
            if (state) {
                redirectURL += (postLogoutURI.includes('?') ? '&' : '?') + 'state=' + encodeURIComponent(state);
            }
            window.location.href = redirectURL;
        } else {
            // Show default logout complete message
            document.body.innerHTML = '<h1>You have been logged out</h1><p>You can now close this window.</p>';
        }
    }
    
    // Call after all logouts complete
    if (completedLogouts >= totalLogouts) {
        setTimeout(redirectAfterLogout, 1000);
    }
</script>
```

---

### Phase 3: Client Session Tracking (1 day)

#### Task 3.1: Enhance Client Session Model
**File:** `backend/pkg/models/models.go`

```go
// ClientSession tracks which clients a user has logged into
type ClientSession struct {
    ID         string    `json:"id" bson:"_id"`
    UserID     string    `json:"user_id" bson:"user_id"`
    ClientID   string    `json:"client_id" bson:"client_id"`
    SessionID  string    `json:"session_id" bson:"session_id"`      // OP session ID
    ClientSID  string    `json:"client_sid" bson:"client_sid"`      // Client-specific session ID
    CreatedAt  time.Time `json:"created_at" bson:"created_at"`
    LastUsedAt time.Time `json:"last_used_at" bson:"last_used_at"`
    IPAddress  string    `json:"ip_address,omitempty" bson:"ip_address,omitempty"`
    UserAgent  string    `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
}
```

#### Task 3.2: Add Storage Methods
**File:** `backend/pkg/storage/storage.go`

```go
// Client session tracking for logout
CreateClientSession(session *ClientSession) error
GetClientSession(id string) (*ClientSession, error)
GetClientSessionsByUserSession(sessionID string) ([]*ClientSession, error)
GetClientSessionsByUser(userID string) ([]*ClientSession, error)
UpdateClientSession(session *ClientSession) error
DeleteClientSession(id string) error
DeleteClientSessionsByUserSession(sessionID string) error
DeleteClientSessionsByClient(clientID string) error
```

#### Task 3.3: Implement in JSON Storage
**File:** `backend/pkg/storage/json.go`

```go
type JSONData struct {
    // ... existing fields ...
    ClientSessions map[string]*models.ClientSession `json:"client_sessions"`
}

// Implement all ClientSession methods
```

#### Task 3.4: Implement in MongoDB Storage
**File:** `backend/pkg/storage/mongodb.go`

```go
type MongoDBStorage struct {
    // ... existing fields ...
    clientSessions *mongo.Collection
}

// Implement all ClientSession methods with proper indexes
```

#### Task 3.5: Track Sessions in Token Handler
**File:** `backend/pkg/handlers/token.go`

```go
// When issuing tokens, create/update client session
func (h *Handlers) Token(c echo.Context) error {
    // ... existing token generation ...
    
    // Track client session
    clientSession := &models.ClientSession{
        ID:         generateSessionID(),
        UserID:     user.ID,
        ClientID:   client.ID,
        SessionID:  authCode.SessionID, // OP session
        ClientSID:  generateClientSID(), // Unique per client
        CreatedAt:  time.Now(),
        LastUsedAt: time.Now(),
        IPAddress:  c.RealIP(),
        UserAgent:  c.Request().UserAgent(),
    }
    
    _ = h.storage.CreateClientSession(clientSession)
    
    // ... rest of token issuance ...
}
```

---

### Phase 4: Update Discovery Document (0.5 days)

#### Task 4.1: Add End Session Endpoint
**File:** `backend/pkg/handlers/discovery.go`

```go
response := map[string]interface{}{
    // ... existing fields ...
    
    // RP-Initiated Logout
    "end_session_endpoint": h.config.Issuer + "/logout",
    
    // Indicate support for front-channel and back-channel logout
    "frontchannel_logout_supported":          true,
    "frontchannel_logout_session_supported":  true,
    "backchannel_logout_supported":           true,
    "backchannel_logout_session_supported":   true,
}
```

---

### Phase 5: Testing & Documentation (1 day)

#### Task 5.1: Unit Tests
**File:** `backend/pkg/handlers/logout_test.go`

```go
func TestLogout_WithIDTokenHint(t *testing.T)
func TestLogout_WithPostLogoutRedirectURI(t *testing.T)
func TestLogout_WithStateParameter(t *testing.T)
func TestLogout_InvalidIDTokenHint(t *testing.T)
func TestLogout_InvalidPostLogoutRedirectURI(t *testing.T)
func TestLogout_WithoutRedirectURI(t *testing.T)
func TestLogout_MultipleClientSessions(t *testing.T)
func TestLogout_ExpiredIDToken(t *testing.T)
```

#### Task 5.2: Integration Tests
**File:** `backend/integration/logout_flow_test.go`

Test complete flow:
1. User logs into multiple RPs
2. RP initiates logout
3. Verify all sessions cleared
4. Verify redirect back to RP
5. Verify state parameter preserved

#### Task 5.3: Create Example RP
**File:** `examples/rp-initiated-logout/main.go`

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "net/url"
)

const (
    issuerURL    = "http://localhost:8080"
    clientID     = "example-client"
    redirectURI  = "http://localhost:9090/callback"
    postLogoutURI = "http://localhost:9090/post-logout"
)

func main() {
    http.HandleFunc("/", handleHome)
    http.HandleFunc("/login", handleLogin)
    http.HandleFunc("/callback", handleCallback)
    http.HandleFunc("/logout", handleLogout)
    http.HandleFunc("/post-logout", handlePostLogout)
    
    log.Println("Example RP running on http://localhost:9090")
    log.Fatal(http.ListenAndServe(":9090", nil))
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
    // Get ID token from session
    cookie, err := r.Cookie("id_token")
    if err != nil {
        http.Error(w, "Not logged in", http.StatusUnauthorized)
        return
    }
    
    idToken := cookie.Value
    
    // Build logout URL
    logoutURL, _ := url.Parse(issuerURL + "/logout")
    q := logoutURL.Query()
    q.Set("id_token_hint", idToken)
    q.Set("post_logout_redirect_uri", postLogoutURI)
    q.Set("state", "xyz123") // Generate random state in production
    logoutURL.RawQuery = q.Encode()
    
    // Clear local session
    http.SetCookie(w, &http.Cookie{
        Name:   "id_token",
        Value:  "",
        MaxAge: -1,
    })
    
    // Redirect to OP logout
    http.Redirect(w, r, logoutURL.String(), http.StatusFound)
}

func handlePostLogout(w http.ResponseWriter, r *http.Request) {
    state := r.URL.Query().Get("state")
    
    // Verify state parameter
    if state != "xyz123" {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }
    
    // Show logout success page
    fmt.Fprintf(w, `
        <html>
        <head><title>Logged Out</title></head>
        <body>
            <h1>Successfully Logged Out</h1>
            <p>You have been logged out from the OpenID Provider and all applications.</p>
            <a href="/">Return to Home</a>
        </body>
        </html>
    `)
}
```

#### Task 5.4: Documentation
**File:** `docs/RP_INITIATED_LOGOUT.md`

Document:
- How to implement RP-initiated logout in client apps
- Request parameters and validation
- Post-logout redirect URI registration
- State parameter usage
- Error handling
- Example code for various frameworks

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
  "post_logout_redirect_uris": ["https://example.com/logged-out"]
}
```

**Response:**
```json
{
  "client_id": "client_123",
  "client_secret": "secret_abc",
  "redirect_uris": ["https://example.com/callback"],
  "post_logout_redirect_uris": ["https://example.com/logged-out"]
}
```

### Logout Endpoint

**Request (from RP):**
```http
GET /logout?id_token_hint=eyJhbGci...&post_logout_redirect_uri=https://example.com/logged-out&state=xyz123 HTTP/1.1
Host: op.example.com
```

**Parameters:**

| Parameter | Required | Description | Example |
|-----------|----------|-------------|---------|
| `id_token_hint` | RECOMMENDED | Previously issued ID Token | `eyJhbGci...` |
| `post_logout_redirect_uri` | OPTIONAL | Where to redirect after logout | `https://rp.example.com/logged-out` |
| `state` | OPTIONAL | Opaque value for CSRF protection | `xyz123` |
| `ui_locales` | OPTIONAL | Preferred languages | `en-US` |

**Response (redirect back to RP):**
```http
HTTP/1.1 302 Found
Location: https://example.com/logged-out?state=xyz123
```

### Discovery Metadata

```json
{
  "issuer": "https://op.example.com",
  "authorization_endpoint": "https://op.example.com/authorize",
  "token_endpoint": "https://op.example.com/token",
  "end_session_endpoint": "https://op.example.com/logout",
  "frontchannel_logout_supported": true,
  "frontchannel_logout_session_supported": true,
  "backchannel_logout_supported": true,
  "backchannel_logout_session_supported": true
}
```

---

## Security Considerations

### 1. ID Token Hint Validation

**Critical Security:**
- MUST verify token signature
- MUST verify issuer matches
- SHOULD allow expired tokens (with grace period)
- MUST NOT reject logout if token invalid (fail open)

```go
// Graceful handling
claims, err := h.validateIDTokenHint(idTokenHint)
if err != nil {
    // Log error but continue logout
    log.Printf("Invalid id_token_hint: %v", err)
    // Try to identify user from session cookie
}
```

### 2. Post-Logout Redirect URI Validation

**REQUIRED Security:**
- MUST validate against registered URIs
- MUST use exact match (no wildcards)
- MUST reject unregistered URIs
- MUST NOT redirect to arbitrary URLs (open redirect vulnerability)

```go
// Strict validation
if !isValidPostLogoutRedirectURI(postLogoutRedirectURI, client.PostLogoutRedirectURIs) {
    // Reject - do not redirect
    log.Printf("Invalid post_logout_redirect_uri")
    postLogoutRedirectURI = "" // Clear invalid URI
}
```

### 3. State Parameter

**CSRF Protection:**
- RP SHOULD generate random state
- RP MUST validate state on return
- State parameter MUST be preserved by OP
- Protects against CSRF attacks

```go
// RP side
state := generateRandomState() // e.g., cryptographically random 32 bytes
storeInSession(state)

// On return
receivedState := r.URL.Query().Get("state")
if receivedState != getFromSession() {
    // CSRF attack detected
    http.Error(w, "Invalid state", http.StatusBadRequest)
    return
}
```

### 4. Session Security

**Best Practices:**
- Clear ALL sessions (OP + all RPs)
- Invalidate tokens immediately
- Clear cookies with proper flags
- Log logout events for audit

### 5. Privacy Considerations

**User Privacy:**
- Don't leak user information in error messages
- Don't expose which RPs user was logged into
- Limit session ID exposure
- Respect DNT headers

---

## Testing Strategy

### Unit Tests

**Logout Handler:**
- ✅ Valid id_token_hint
- ✅ Invalid id_token_hint
- ✅ Expired id_token_hint
- ✅ Missing id_token_hint
- ✅ Valid post_logout_redirect_uri
- ✅ Invalid post_logout_redirect_uri
- ✅ State parameter preserved
- ✅ Multiple client sessions
- ✅ Front-channel + back-channel logout
- ✅ Session cleanup

**Validation Functions:**
- ✅ URI validation (absolute, no fragment, HTTPS)
- ✅ Registered URI matching
- ✅ State parameter handling

### Integration Tests

**Full Logout Flow:**
1. User logs into RP1 and RP2
2. RP1 initiates logout
3. Verify OP session ended
4. Verify RP1 and RP2 sessions ended
5. Verify redirect to RP1 post-logout URI
6. Verify state parameter preserved

**Error Scenarios:**
- Invalid id_token_hint (continue logout)
- Unregistered post_logout_redirect_uri (show default page)
- Missing state (logout succeeds, no state returned)
- Network failures during RP notification

### Manual Testing

**Browser Testing:**
- Test with Chrome, Firefox, Safari
- Test with multiple tabs
- Test with third-party cookies blocked
- Test with JavaScript disabled

**Security Testing:**
- CSRF protection
- Open redirect prevention
- Token validation
- Session fixation
- XSS prevention

---

## Timeline

| Phase | Tasks | Duration | Dependencies |
|-------|-------|----------|--------------|
| Phase 1 | Client model enhancement | 0.5 day | None |
| Phase 2 | Logout endpoint | 2-3 days | Phase 1 |
| Phase 3 | Client session tracking | 1 day | Phase 2 |
| Phase 4 | Discovery update | 0.5 day | Phase 2 |
| Phase 5 | Testing & docs | 1 day | All phases |
| **Total** | | **5-6 days** | |

---

## Success Criteria

1. ✅ RPs can initiate logout at OP
2. ✅ ID token hint properly validated
3. ✅ Post-logout redirect URIs validated
4. ✅ State parameter preserved
5. ✅ All sessions cleaned up (OP + RPs)
6. ✅ Front-channel and back-channel logout triggered
7. ✅ Discovery document updated
8. ✅ Security requirements met
9. ✅ All tests passing
10. ✅ Documentation complete
11. ✅ Example RP provided

---

## User Experience Flow

### Happy Path:

```
1. User is logged into App A and App B
2. User clicks "Logout" in App A
3. App A redirects to: OP/logout?id_token_hint=...&post_logout_redirect_uri=...&state=...
4. OP validates request
5. OP ends user's session
6. OP notifies App B via front-channel or back-channel
7. App B clears its session
8. OP redirects back to App A's post-logout URI with state
9. App A validates state
10. App A shows "Logged out successfully"
```

### Alternative Flows:

**Without Post-Logout Redirect:**
- User clicks logout in App A
- App A redirects to OP/logout with id_token_hint
- OP shows "You've been logged out" page
- User can close browser

**Direct OP Logout:**
- User goes directly to OP/logout
- OP shows logout confirmation
- User confirms
- All sessions cleared
- "Logged out" page shown

---

## Integration with Existing Plans

### Front-Channel Logout Plan
- RP-Initiated Logout USES Front-Channel Logout
- When logout endpoint receives request, it triggers front-channel notifications
- iframe-based logout to all registered RPs

### Back-Channel Logout Plan
- RP-Initiated Logout USES Back-Channel Logout
- Async server-to-server notifications
- More reliable than front-channel

### Implementation Order:
1. ✅ Front-Channel Logout (planned)
2. ✅ Back-Channel Logout (planned)
3. 🔴 RP-Initiated Logout (this plan) ← **Depends on above**

---

## References

- [OpenID Connect RP-Initiated Logout 1.0](https://openid.net/specs/openid-connect-rpinitiated-1_0.html)
- [OpenID Connect Front-Channel Logout 1.0](https://openid.net/specs/openid-connect-frontchannel-1_0.html)
- [OpenID Connect Back-Channel Logout 1.0](https://openid.net/specs/openid-connect-backchannel-1_0.html)
- [OpenID Connect Session Management 1.0](https://openid.net/specs/openid-connect-session-1_0.html)
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
