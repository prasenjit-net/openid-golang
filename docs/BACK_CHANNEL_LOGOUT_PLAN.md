# Back-Channel Logout Implementation Plan

**Document Version:** 1.0  
**Date:** October 24, 2025  
**Status:** Planning Phase  
**Specification:** [OpenID Connect Back-Channel Logout 1.0](https://openid.net/specs/openid-connect-backchannel-1_0.html)

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Specification Overview](#specification-overview)
- [Front-Channel vs Back-Channel](#front-channel-vs-back-channel)
- [Current Implementation Status](#current-implementation-status)
- [Implementation Tasks](#implementation-tasks)
- [API Specifications](#api-specifications)
- [Security Considerations](#security-considerations)
- [Testing Strategy](#testing-strategy)

---

## Executive Summary

This document outlines the plan to implement **OpenID Connect Back-Channel Logout** in the openid-golang project. Back-Channel Logout enables an OpenID Provider (OP) to directly notify Relying Parties (RPs) when a user logs out using secure server-to-server communication.

**Key Features:**
- Server-to-server logout notifications
- Logout tokens (JWT) instead of browser redirects
- Asynchronous notification delivery
- Support for both session-specific and user-wide logout
- More reliable than front-channel (no browser dependencies)

**Estimated Effort:** 6-8 days

---

## Specification Overview

### What is Back-Channel Logout?

Back-Channel Logout allows the OP to notify RPs about logout events using direct HTTPS requests (back-channel) rather than browser-based communication. The OP sends a signed JWT (Logout Token) to each RP's registered logout endpoint.

### How It Works:

1. User initiates logout at the OP
2. OP creates a Logout Token (JWT) for each RP
3. OP sends POST request to each RP's `backchannel_logout_uri`
4. RP validates the Logout Token
5. RP clears the specified session(s)
6. RP responds with HTTP 200 OK
7. User is redirected to `post_logout_redirect_uri` (optional)

### Key Differences from Front-Channel:

| Aspect | Front-Channel | Back-Channel |
|--------|--------------|--------------|
| Communication | Browser (iframes) | Server-to-server (HTTPS POST) |
| Reliability | Depends on browser | More reliable |
| Timing | Synchronous | Asynchronous |
| Privacy | Browser sees all RPs | Server-to-server only |
| Token | Query parameters | Signed JWT |
| User must wait | Yes | No (can redirect immediately) |

---

## Front-Channel vs Back-Channel

### When to Use Back-Channel:

✅ **More reliable** - No dependency on browser, cookies, or JavaScript  
✅ **Better privacy** - User's browser doesn't see all RPs  
✅ **Asynchronous** - Don't block user during logout  
✅ **Corporate environments** - Works behind firewalls/proxies  
✅ **Mobile apps** - Better for native/mobile applications  

### When to Use Front-Channel:

✅ **Simple setup** - No server endpoint required at RP  
✅ **Cookie-based sessions** - Can directly access browser cookies  
✅ **Legacy systems** - Easier for older applications  
✅ **Immediate feedback** - User sees when logout completes  

### Best Practice:

Support **both** mechanisms and let RPs choose which to implement. Many implementations support both for maximum compatibility.

---

## Current Implementation Status

### ✅ Already Implemented (from Front-Channel plan)

| Feature | Status | Location |
|---------|--------|----------|
| User Sessions | ✅ Complete | `pkg/models/models.go` - UserSession |
| Session Storage | ✅ Complete | `pkg/storage/*.go` |
| Client Model | ✅ Complete | `pkg/models/models.go` - Client |
| Session ID (sid) | ✅ Complete | ID token claims |
| ClientSession tracking | ✅ Complete | `pkg/models/models.go` |
| Logout endpoint | ✅ Complete | `/logout` handler |

### ❌ New Features Required

| Feature | Priority | Impact |
|---------|----------|--------|
| Back-Channel Logout URIs in Client | 🔴 High | Required for logout notifications |
| Logout Token generation | 🔴 High | JWT with specific claims |
| Asynchronous HTTP client | 🔴 High | Send logout requests to RPs |
| Retry mechanism | 🟡 Medium | Handle RP failures |
| Logout notification queue | 🟡 Medium | For reliability |
| Audit logging | 🟢 Low | Track logout events |

---

## Implementation Tasks

### Phase 1: Enhance Client Model (1 day)

#### Task 1.1: Add Back-Channel Logout Fields to Client Model
**File:** `backend/pkg/models/models.go`

```go
type Client struct {
    // ... existing fields ...
    
    // Front-Channel Logout (already implemented)
    FrontChannelLogoutURI            string `json:"frontchannel_logout_uri,omitempty" bson:"frontchannel_logout_uri,omitempty"`
    FrontChannelLogoutSessionRequired bool   `json:"frontchannel_logout_session_required,omitempty" bson:"frontchannel_logout_session_required,omitempty"`
    
    // Back-Channel Logout (new)
    BackChannelLogoutURI              string `json:"backchannel_logout_uri,omitempty" bson:"backchannel_logout_uri,omitempty"`
    BackChannelLogoutSessionRequired  bool   `json:"backchannel_logout_session_required,omitempty" bson:"backchannel_logout_session_required,omitempty"`
}
```

#### Task 1.2: Update Client Registration Validation
**File:** `backend/pkg/handlers/registration.go`

Add validation for `backchannel_logout_uri`:
- Must be HTTPS (no exceptions - more strict than front-channel)
- Must be absolute URI
- Must not contain fragment
- Should be different from other endpoints
- Validate endpoint is reachable (optional health check)

```go
func validateBackChannelLogoutURI(uri string) error {
    if uri == "" {
        return nil // Optional field
    }
    
    parsedURI, err := url.Parse(uri)
    if err != nil {
        return errors.New("invalid backchannel_logout_uri")
    }
    
    if !parsedURI.IsAbs() {
        return errors.New("backchannel_logout_uri must be absolute")
    }
    
    if parsedURI.Fragment != "" {
        return errors.New("backchannel_logout_uri must not contain fragment")
    }
    
    // MUST use HTTPS (no localhost exception for back-channel)
    if parsedURI.Scheme != "https" {
        return errors.New("backchannel_logout_uri must use HTTPS")
    }
    
    return nil
}
```

---

### Phase 2: Implement Logout Token Generation (1-2 days)

#### Task 2.1: Create Logout Token Model
**File:** `backend/pkg/models/models.go`

```go
// LogoutToken represents an OpenID Connect Logout Token (JWT)
// Spec: https://openid.net/specs/openid-connect-backchannel-1_0.html#LogoutToken
type LogoutTokenClaims struct {
    Issuer    string                 `json:"iss"`           // REQUIRED: Issuer Identifier
    Subject   string                 `json:"sub,omitempty"` // Subject Identifier (REQUIRED unless sid)
    Audience  []string               `json:"aud"`           // REQUIRED: Client ID(s)
    IssuedAt  int64                  `json:"iat"`           // REQUIRED: Issue time
    JTI       string                 `json:"jti"`           // REQUIRED: Unique identifier
    Events    map[string]interface{} `json:"events"`        // REQUIRED: Logout event
    SessionID string                 `json:"sid,omitempty"` // Session ID (REQUIRED unless sub)
    
    // MUST NOT contain: nonce, azp, auth_time, acr, amr
}
```

#### Task 2.2: Create Logout Token Generator
**File:** `backend/pkg/crypto/logout_token.go`

```go
package crypto

import (
    "crypto/rand"
    "encoding/base64"
    "errors"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
    "github.com/prasenjit-net/openid-golang/backend/pkg/models"
)

// GenerateLogoutToken creates a JWT logout token per OpenID Connect Back-Channel Logout spec
func GenerateLogoutToken(
    issuer string,
    clientID string,
    userID string,
    sessionID string,
    signingKey interface{},
) (string, error) {
    // At least one of sub or sid MUST be included
    if userID == "" && sessionID == "" {
        return "", errors.New("either userID or sessionID must be provided")
    }
    
    // Generate unique JTI
    jti, err := generateJTI()
    if err != nil {
        return "", err
    }
    
    // Create logout event claim
    // The member name is: http://schemas.openid.net/event/backchannel-logout
    events := map[string]interface{}{
        "http://schemas.openid.net/event/backchannel-logout": map[string]interface{}{},
    }
    
    // Build claims
    claims := models.LogoutTokenClaims{
        Issuer:    issuer,
        Subject:   userID,
        Audience:  []string{clientID},
        IssuedAt:  time.Now().Unix(),
        JTI:       jti,
        Events:    events,
        SessionID: sessionID,
    }
    
    // Create token
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
        "iss":    claims.Issuer,
        "aud":    claims.Audience,
        "iat":    claims.IssuedAt,
        "jti":    claims.JTI,
        "events": claims.Events,
    })
    
    // Add optional claims
    if claims.Subject != "" {
        token.Claims.(jwt.MapClaims)["sub"] = claims.Subject
    }
    if claims.SessionID != "" {
        token.Claims.(jwt.MapClaims)["sid"] = claims.SessionID
    }
    
    // Sign token
    tokenString, err := token.SignedString(signingKey)
    if err != nil {
        return "", err
    }
    
    return tokenString, nil
}

// generateJTI creates a unique identifier for the logout token
func generateJTI() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateLogoutToken validates a logout token (for testing or RP implementation)
func ValidateLogoutToken(tokenString string, publicKey interface{}) (*models.LogoutTokenClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Verify signing method
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return publicKey, nil
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
    
    // Validate required claims
    if _, ok := claims["iss"]; !ok {
        return nil, errors.New("missing iss claim")
    }
    if _, ok := claims["aud"]; !ok {
        return nil, errors.New("missing aud claim")
    }
    if _, ok := claims["iat"]; !ok {
        return nil, errors.New("missing iat claim")
    }
    if _, ok := claims["jti"]; !ok {
        return nil, errors.New("missing jti claim")
    }
    if _, ok := claims["events"]; !ok {
        return nil, errors.New("missing events claim")
    }
    
    // At least one of sub or sid must be present
    hasSub := false
    hasSid := false
    if _, ok := claims["sub"]; ok {
        hasSub = true
    }
    if _, ok := claims["sid"]; ok {
        hasSid = true
    }
    if !hasSub && !hasSid {
        return nil, errors.New("logout token must contain sub or sid")
    }
    
    // Verify events claim structure
    events, ok := claims["events"].(map[string]interface{})
    if !ok {
        return nil, errors.New("invalid events claim")
    }
    if _, ok := events["http://schemas.openid.net/event/backchannel-logout"]; !ok {
        return nil, errors.New("missing backchannel-logout event")
    }
    
    // MUST NOT contain these claims
    forbiddenClaims := []string{"nonce", "azp", "auth_time", "acr", "amr"}
    for _, claim := range forbiddenClaims {
        if _, ok := claims[claim]; ok {
            return nil, errors.New("logout token must not contain " + claim)
        }
    }
    
    // Build LogoutTokenClaims
    logoutClaims := &models.LogoutTokenClaims{
        Issuer:   claims["iss"].(string),
        IssuedAt: int64(claims["iat"].(float64)),
        JTI:      claims["jti"].(string),
        Events:   events,
    }
    
    if sub, ok := claims["sub"].(string); ok {
        logoutClaims.Subject = sub
    }
    if sid, ok := claims["sid"].(string); ok {
        logoutClaims.SessionID = sid
    }
    
    // Handle audience (can be string or array)
    switch aud := claims["aud"].(type) {
    case string:
        logoutClaims.Audience = []string{aud}
    case []interface{}:
        for _, a := range aud {
            if audStr, ok := a.(string); ok {
                logoutClaims.Audience = append(logoutClaims.Audience, audStr)
            }
        }
    }
    
    return logoutClaims, nil
}
```

---

### Phase 3: Implement Logout Notification Service (2-3 days)

#### Task 3.1: Create Logout Notification Queue
**File:** `backend/pkg/models/models.go`

```go
// LogoutNotification represents a pending logout notification
type LogoutNotification struct {
    ID               string    `json:"id" bson:"_id"`
    ClientID         string    `json:"client_id" bson:"client_id"`
    BackChannelURI   string    `json:"backchannel_uri" bson:"backchannel_uri"`
    LogoutToken      string    `json:"logout_token" bson:"logout_token"`
    CreatedAt        time.Time `json:"created_at" bson:"created_at"`
    Attempts         int       `json:"attempts" bson:"attempts"`
    LastAttemptAt    time.Time `json:"last_attempt_at" bson:"last_attempt_at"`
    Status           string    `json:"status" bson:"status"` // pending, sent, failed
    ErrorMessage     string    `json:"error_message,omitempty" bson:"error_message,omitempty"`
}
```

#### Task 3.2: Add Storage Methods
**File:** `backend/pkg/storage/storage.go`

```go
// Logout notification management
CreateLogoutNotification(notification *LogoutNotification) error
GetPendingLogoutNotifications(limit int) ([]*LogoutNotification, error)
UpdateLogoutNotification(notification *LogoutNotification) error
DeleteLogoutNotification(id string) error
```

#### Task 3.3: Create Logout Notification Service
**File:** `backend/pkg/services/logout_notifier.go`

```go
package services

import (
    "bytes"
    "context"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"
    
    "github.com/prasenjit-net/openid-golang/backend/pkg/models"
    "github.com/prasenjit-net/openid-golang/backend/pkg/storage"
)

// LogoutNotifier handles sending back-channel logout notifications to RPs
type LogoutNotifier struct {
    storage    storage.Storage
    httpClient *http.Client
    maxRetries int
    retryDelay time.Duration
}

// NewLogoutNotifier creates a new logout notifier service
func NewLogoutNotifier(storage storage.Storage) *LogoutNotifier {
    return &LogoutNotifier{
        storage: storage,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    MinVersion: tls.VersionTLS12,
                },
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
            },
        },
        maxRetries: 3,
        retryDelay: 5 * time.Second,
    }
}

// SendLogoutNotification sends a logout token to an RP's backchannel_logout_uri
func (ln *LogoutNotifier) SendLogoutNotification(
    ctx context.Context,
    backChannelURI string,
    logoutToken string,
) error {
    // Prepare form data
    formData := url.Values{}
    formData.Set("logout_token", logoutToken)
    
    // Create request
    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        backChannelURI,
        bytes.NewBufferString(formData.Encode()),
    )
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("User-Agent", "OpenID-Provider/1.0")
    
    // Send request
    resp, err := ln.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()
    
    // Read response body for logging
    body, _ := io.ReadAll(resp.Body)
    
    // Check response status
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("RP returned error status %d: %s", resp.StatusCode, string(body))
    }
    
    return nil
}

// QueueLogoutNotification adds a logout notification to the queue
func (ln *LogoutNotifier) QueueLogoutNotification(
    clientID string,
    backChannelURI string,
    logoutToken string,
) error {
    notification := &models.LogoutNotification{
        ID:             generateNotificationID(),
        ClientID:       clientID,
        BackChannelURI: backChannelURI,
        LogoutToken:    logoutToken,
        CreatedAt:      time.Now(),
        Attempts:       0,
        Status:         "pending",
    }
    
    return ln.storage.CreateLogoutNotification(notification)
}

// ProcessQueue processes pending logout notifications
func (ln *LogoutNotifier) ProcessQueue(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            ln.processPendingNotifications(ctx)
        }
    }
}

func (ln *LogoutNotifier) processPendingNotifications(ctx context.Context) {
    // Get pending notifications
    notifications, err := ln.storage.GetPendingLogoutNotifications(10)
    if err != nil {
        return
    }
    
    for _, notification := range notifications {
        // Skip if too many attempts
        if notification.Attempts >= ln.maxRetries {
            notification.Status = "failed"
            _ = ln.storage.UpdateLogoutNotification(notification)
            continue
        }
        
        // Skip if recently attempted
        if time.Since(notification.LastAttemptAt) < ln.retryDelay {
            continue
        }
        
        // Attempt to send
        notification.Attempts++
        notification.LastAttemptAt = time.Now()
        
        err := ln.SendLogoutNotification(
            ctx,
            notification.BackChannelURI,
            notification.LogoutToken,
        )
        
        if err != nil {
            notification.ErrorMessage = err.Error()
            _ = ln.storage.UpdateLogoutNotification(notification)
        } else {
            notification.Status = "sent"
            _ = ln.storage.UpdateLogoutNotification(notification)
        }
    }
}

func generateNotificationID() string {
    return fmt.Sprintf("notification_%d", time.Now().UnixNano())
}
```

---

### Phase 4: Update Logout Handler (1-2 days)

#### Task 4.1: Enhance Logout Handler with Back-Channel Support
**File:** `backend/pkg/handlers/logout.go`

```go
func (h *Handlers) Logout(c echo.Context) error {
    // ... existing front-channel logout code ...
    
    // Get all client sessions for this user session
    clientSessions, err := h.storage.GetClientSessionsByUserSession(sessionID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to retrieve sessions",
        })
    }
    
    // Build list of logout targets (front-channel) and send back-channel notifications
    var frontChannelTargets []LogoutTarget
    var backChannelCount int
    
    for _, cs := range clientSessions {
        client, err := h.storage.GetClientByID(cs.ClientID)
        if err != nil || client == nil {
            continue
        }
        
        // Back-Channel Logout
        if client.BackChannelLogoutURI != "" {
            err := h.sendBackChannelLogout(client, cs, sessionID)
            if err != nil {
                // Log error but continue (best effort)
                // TODO: Add proper logging
            } else {
                backChannelCount++
            }
        }
        
        // Front-Channel Logout (if also configured)
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
    
    // Delete all sessions immediately (don't wait for back-channel confirmations)
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
    
    // If only back-channel, redirect immediately
    if len(frontChannelTargets) == 0 && backChannelCount > 0 {
        if postLogoutRedirectURI != "" {
            redirectURL := postLogoutRedirectURI
            if state != "" {
                redirectURL += "?state=" + url.QueryEscape(state)
            }
            return c.Redirect(http.StatusFound, redirectURL)
        }
        return c.HTML(http.StatusOK, "<h1>You have been logged out</h1>")
    }
    
    // Render logout page with iframes (for front-channel)
    return c.Render(http.StatusOK, "logout.html", map[string]interface{}{
        "LogoutTargets":         frontChannelTargets,
        "PostLogoutRedirectURI": postLogoutRedirectURI,
        "State":                 state,
        "Issuer":                h.config.Issuer,
    })
}

// sendBackChannelLogout sends a back-channel logout notification to an RP
func (h *Handlers) sendBackChannelLogout(
    client *models.Client,
    clientSession *models.ClientSession,
    userSessionID string,
) error {
    // Get user for sub claim
    user, err := h.storage.GetUserByID(clientSession.UserID)
    if err != nil {
        return err
    }
    
    // Determine which claims to include based on configuration
    var subClaim string
    var sidClaim string
    
    if client.BackChannelLogoutSessionRequired {
        // Include session ID
        sidClaim = clientSession.SessionID
    } else {
        // Include subject (user ID)
        subClaim = user.ID
    }
    
    // Generate logout token
    logoutToken, err := h.crypto.GenerateLogoutToken(
        h.config.Issuer,
        client.ID,
        subClaim,
        sidClaim,
        h.signingKey,
    )
    if err != nil {
        return err
    }
    
    // Queue notification for async delivery
    err = h.logoutNotifier.QueueLogoutNotification(
        client.ID,
        client.BackChannelLogoutURI,
        logoutToken,
    )
    
    return err
}
```

#### Task 4.2: Initialize Logout Notifier Service
**File:** `backend/cmd/serve.go`

```go
import (
    "github.com/prasenjit-net/openid-golang/backend/pkg/services"
)

func serve(cmd *cobra.Command, args []string) error {
    // ... existing code ...
    
    // Initialize logout notifier service
    logoutNotifier := services.NewLogoutNotifier(store)
    
    // Start background worker for logout notifications
    ctx := context.Background()
    go logoutNotifier.ProcessQueue(ctx)
    
    // Pass to handlers
    handlers := &handlers.Handlers{
        // ... existing fields ...
        logoutNotifier: logoutNotifier,
    }
    
    // ... rest of code ...
}
```

---

### Phase 5: Update Discovery Document (0.5 days)

#### Task 5.1: Add Back-Channel Logout Metadata
**File:** `backend/pkg/handlers/discovery.go`

```go
response := map[string]interface{}{
    // ... existing fields ...
    
    // Front-Channel Logout
    "frontchannel_logout_supported":          true,
    "frontchannel_logout_session_supported":  true,
    
    // Back-Channel Logout
    "backchannel_logout_supported":           true,
    "backchannel_logout_session_supported":   true,
    
    // RP-Initiated Logout
    "end_session_endpoint": h.config.Issuer + "/logout",
}
```

---

### Phase 6: Testing & Documentation (1-2 days)

#### Task 6.1: Unit Tests
**File:** `backend/pkg/crypto/logout_token_test.go`

Test cases:
- Generate logout token with sub
- Generate logout token with sid
- Generate logout token with both sub and sid
- Validate logout token structure
- Verify forbidden claims are not present
- Verify events claim structure
- Validate logout token signature

#### Task 6.2: Service Tests
**File:** `backend/pkg/services/logout_notifier_test.go`

Test cases:
- Send successful logout notification
- Handle RP returning error status
- Handle network timeout
- Retry failed notifications
- Queue management
- Concurrent notification sending

#### Task 6.3: Integration Tests
**File:** `backend/pkg/handlers/logout_backchannel_test.go`

Test cases:
- Full back-channel logout flow
- Mixed front-channel and back-channel
- Session cleanup verification
- Logout token validation at RP
- Async notification delivery

#### Task 6.4: Create Mock RP for Testing
**File:** `examples/backchannel-rp/main.go`

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    
    "github.com/golang-jwt/jwt/v5"
)

func main() {
    http.HandleFunc("/backchannel-logout", handleBackChannelLogout)
    log.Println("Mock RP listening on :9090")
    log.Fatal(http.ListenAndServe(":9090", nil))
}

func handleBackChannelLogout(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Parse form data
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form data", http.StatusBadRequest)
        return
    }
    
    logoutToken := r.FormValue("logout_token")
    if logoutToken == "" {
        http.Error(w, "Missing logout_token", http.StatusBadRequest)
        return
    }
    
    // Parse (but don't validate) the token for demo purposes
    token, _, err := new(jwt.Parser).ParseUnverified(logoutToken, jwt.MapClaims{})
    if err != nil {
        http.Error(w, "Invalid token", http.StatusBadRequest)
        return
    }
    
    claims := token.Claims.(jwt.MapClaims)
    
    // Log the logout event
    log.Printf("Received logout notification:")
    log.Printf("  Issuer: %v", claims["iss"])
    log.Printf("  Audience: %v", claims["aud"])
    log.Printf("  Subject: %v", claims["sub"])
    log.Printf("  Session ID: %v", claims["sid"])
    log.Printf("  JTI: %v", claims["jti"])
    
    // In a real RP, you would:
    // 1. Validate the token signature
    // 2. Verify issuer and audience
    // 3. Clear the session(s) for the user/session
    // 4. Return 200 OK
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
    })
}
```

#### Task 6.5: Documentation
**File:** `docs/BACK_CHANNEL_LOGOUT.md`

Document:
- How to configure back-channel logout
- Client registration parameters
- Logout token structure
- RP implementation guide
- Security considerations
- Troubleshooting
- Comparison with front-channel

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
  "backchannel_logout_uri": "https://example.com/backchannel-logout",
  "backchannel_logout_session_required": true
}
```

**Response:**
```json
{
  "client_id": "client_123",
  "client_secret": "secret_abc",
  "backchannel_logout_uri": "https://example.com/backchannel-logout",
  "backchannel_logout_session_required": true
}
```

### Logout Endpoint (OP)

**Request:**
```http
GET /logout?id_token_hint=<token>&post_logout_redirect_uri=https://example.com&state=xyz
```

**Behavior:**
- Generates logout tokens for all RPs with back-channel URIs
- Sends POST requests asynchronously to each RP
- Deletes user sessions immediately
- Redirects user (doesn't wait for RP responses)

### Back-Channel Logout Notification (OP → RP)

**Request:**
```http
POST https://example.com/backchannel-logout HTTP/1.1
Content-Type: application/x-www-form-urlencoded

logout_token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Logout Token (decoded):**
```json
{
  "iss": "https://op.example.com",
  "aud": "client_123",
  "iat": 1698163200,
  "jti": "unique-identifier-xyz",
  "events": {
    "http://schemas.openid.net/event/backchannel-logout": {}
  },
  "sid": "session_abc"
}
```

**RP Response:**
```http
HTTP/1.1 200 OK
```

---

## Security Considerations

### 1. HTTPS Required
- Back-channel URIs **MUST** use HTTPS (no exceptions)
- No localhost exception (unlike front-channel)
- Verify TLS certificates

### 2. Token Validation (RP Side)
- Verify token signature using OP's public key
- Validate issuer (iss) matches expected OP
- Validate audience (aud) matches client ID
- Verify events claim structure
- Check for forbidden claims (nonce, azp, etc.)
- Validate JTI is unique (prevent replay)

### 3. Asynchronous Processing
- Don't block user logout on RP responses
- Implement retry mechanism with backoff
- Set reasonable timeouts (10 seconds)
- Log failed notifications for audit

### 4. Privacy
- Use sid (session-specific) instead of sub when possible
- Don't log sensitive user information
- Limit retention of notification queue

### 5. Denial of Service Protection
- Rate limit logout notifications
- Implement circuit breaker for failing RPs
- Set maximum retry attempts (3-5)
- Timeout long-running requests

### 6. Token Expiration
- Logout tokens should be short-lived (< 5 minutes)
- RPs should reject old tokens
- Include iat claim for validation

---

## Testing Strategy

### Unit Tests
- ✅ Logout token generation
- ✅ Token validation
- ✅ Claims validation
- ✅ Forbidden claims detection
- ✅ JTI uniqueness

### Service Tests
- ✅ HTTP client functionality
- ✅ Retry mechanism
- ✅ Queue processing
- ✅ Concurrent requests
- ✅ Error handling

### Integration Tests
- ✅ Full logout flow
- ✅ Multiple RP notifications
- ✅ Mixed front/back channel
- ✅ Session cleanup
- ✅ Async delivery

### Manual Testing
- ✅ Test with mock RP
- ✅ Network failure scenarios
- ✅ RP timeout scenarios
- ✅ Token validation at RP
- ✅ Concurrent logouts

### Load Testing
- ✅ Multiple simultaneous logouts
- ✅ Many RPs per user
- ✅ Queue performance
- ✅ Memory usage

---

## Timeline

| Phase | Tasks | Duration | Dependencies |
|-------|-------|----------|--------------|
| Phase 1 | Client model enhancement | 1 day | None |
| Phase 2 | Logout token generation | 1-2 days | Phase 1 |
| Phase 3 | Notification service | 2-3 days | Phase 2 |
| Phase 4 | Update logout handler | 1-2 days | Phase 3 |
| Phase 5 | Discovery update | 0.5 day | Phase 4 |
| Phase 6 | Testing & docs | 1-2 days | Phase 5 |
| **Total** | | **6.5-10.5 days** | |

---

## Success Criteria

1. ✅ Clients can register with `backchannel_logout_uri`
2. ✅ Logout tokens generated with correct claims structure
3. ✅ Asynchronous notification delivery works
4. ✅ Retry mechanism handles failures
5. ✅ Sessions cleaned up immediately
6. ✅ Discovery document advertises back-channel support
7. ✅ All tests passing
8. ✅ Security best practices followed
9. ✅ Documentation complete
10. ✅ Mock RP for testing provided

---

## Implementation Order

### Recommended Approach:

1. **Implement Front-Channel First** (if not done)
   - Simpler to implement and test
   - Provides immediate value
   - Foundation for back-channel

2. **Then Implement Back-Channel**
   - Builds on front-channel infrastructure
   - Can support both methods simultaneously
   - RPs can choose which to use

3. **Support Both Methods**
   - Best compatibility
   - RPs can use what works for their architecture
   - Enterprise and mobile clients prefer back-channel
   - Browser-based apps may prefer front-channel

---

## Comparison Matrix

| Feature | Front-Channel | Back-Channel |
|---------|---------------|--------------|
| **Communication** | Browser (iframe) | Server-to-server (HTTPS) |
| **Reliability** | ★★☆☆☆ | ★★★★★ |
| **Privacy** | ★★☆☆☆ | ★★★★★ |
| **Complexity** | ★★☆☆☆ | ★★★★☆ |
| **User Wait Time** | Yes | No |
| **Cookie Access** | Yes | No |
| **Firewall Friendly** | ★★★☆☆ | ★★★★★ |
| **Mobile Support** | ★★☆☆☆ | ★★★★★ |
| **HTTPS Required** | Some exceptions | Always |
| **Best For** | Browser apps | Enterprise, mobile |

---

## References

- [OpenID Connect Back-Channel Logout 1.0](https://openid.net/specs/openid-connect-backchannel-1_0.html)
- [OpenID Connect Front-Channel Logout 1.0](https://openid.net/specs/openid-connect-frontchannel-1_0.html)
- [OpenID Connect RP-Initiated Logout 1.0](https://openid.net/specs/openid-connect-rpinitiated-1_0.html)
- [RFC 8417 - Security Event Token (SET)](https://tools.ietf.org/html/rfc8417)
- [JWT Best Practices](https://datatracker.ietf.org/doc/html/rfc8725)
