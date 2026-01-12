# Front-Channel Logout Implementation

**Status:** ✅ Implemented  
**Specification:** [OpenID Connect Front-Channel Logout 1.0](https://openid.net/specs/openid-connect-frontchannel-1_0.html)  
**Date:** January 12, 2026

---

## Overview

The openid-golang server now supports **OpenID Connect Front-Channel Logout**, enabling Relying Parties (RPs) to be notified when a user logs out at the OpenID Provider (OP). This allows RPs to clean up their local sessions and ensure a complete logout across all applications.

### How It Works

1. User initiates logout at the OP by visiting `/logout`
2. OP identifies all clients associated with the user's session
3. OP renders an HTML page with hidden iframes, one for each client
4. Each iframe loads the client's `frontchannel_logout_uri` with session parameters
5. The RP receives the logout notification and clears its local session
6. After a timeout, the user is redirected to the `post_logout_redirect_uri` (if provided)

---

## Client Registration

To enable front-channel logout for a client, register it with the following parameters:

### Required Parameters

- `frontchannel_logout_uri` - The URI where your RP will receive logout notifications

### Optional Parameters

- `frontchannel_logout_session_required` - Boolean indicating whether to include `sid` and `iss` parameters in the logout request (default: false)

### Example: Dynamic Client Registration

```json
POST /register
Content-Type: application/json

{
  "client_name": "My Application",
  "redirect_uris": ["https://myapp.example.com/callback"],
  "frontchannel_logout_uri": "https://myapp.example.com/logout",
  "frontchannel_logout_session_required": true
}
```

### Example: Manual Client Creation

```json
{
  "client_id": "my-app",
  "client_secret": "secret-123",
  "client_name": "My Application",
  "redirect_uris": ["https://myapp.example.com/callback"],
  "frontchannel_logout_uri": "https://myapp.example.com/logout",
  "frontchannel_logout_session_required": true
}
```

---

## Logout Endpoint

### Endpoint: `/logout`

**Methods:** GET, POST

### Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `id_token_hint` | No | ID token previously issued to the client (not currently validated) |
| `post_logout_redirect_uri` | No | URI to redirect the user after logout |
| `state` | No | Opaque value to maintain state between request and callback |

### Example Requests

#### Simple Logout
```
GET /logout
```

#### Logout with Redirect
```
GET /logout?post_logout_redirect_uri=https://myapp.example.com/logged-out&state=xyz123
```

---

## RP Implementation Guide

As a Relying Party, you need to implement a logout notification handler.

### 1. Register Your Logout URI

When registering your client, provide a `frontchannel_logout_uri`:

```json
{
  "frontchannel_logout_uri": "https://myapp.example.com/logout",
  "frontchannel_logout_session_required": true
}
```

### 2. Implement the Logout Handler

Your RP must implement an endpoint that:
- Accepts GET requests (loaded in an iframe)
- Receives `iss` and `sid` parameters (if `frontchannel_logout_session_required=true`)
- Clears the local session for that user
- Returns a simple success response (usually an empty page)

#### Example: Node.js/Express

```javascript
app.get('/logout', (req, res) => {
  const { iss, sid } = req.query;
  
  // Validate issuer (optional but recommended)
  if (iss !== 'https://auth.example.com') {
    return res.status(400).send('Invalid issuer');
  }
  
  // Find and delete session by sid
  if (sid) {
    sessions.deleteBySessionId(sid);
  }
  
  // Return empty page (loaded in iframe)
  res.send('<html><body>Logged out</body></html>');
});
```

#### Example: Python/Flask

```python
@app.route('/logout')
def logout():
    iss = request.args.get('iss')
    sid = request.args.get('sid')
    
    # Validate issuer
    if iss != 'https://auth.example.com':
        return 'Invalid issuer', 400
    
    # Find and delete session by sid
    if sid:
        session_store.delete_by_session_id(sid)
    
    # Return empty page
    return '<html><body>Logged out</body></html>'
```

#### Example: Go

```go
func logoutHandler(w http.ResponseWriter, r *http.Request) {
    iss := r.URL.Query().Get("iss")
    sid := r.URL.Query().Get("sid")
    
    // Validate issuer
    if iss != "https://auth.example.com" {
        http.Error(w, "Invalid issuer", http.StatusBadRequest)
        return
    }
    
    // Find and delete session by sid
    if sid != "" {
        sessionStore.DeleteBySessionID(sid)
    }
    
    // Return empty page
    w.Write([]byte("<html><body>Logged out</body></html>"))
}
```

### 3. Handle the Session ID

The `sid` parameter is a session identifier that was included in the ID token when the user logged in. You need to:

1. **Store the `sid` claim** when you receive the ID token during login
2. **Associate it with the user's session** in your application
3. **Use it to identify which session to terminate** when you receive a logout notification

#### Storing the sid

When you receive an ID token during login:

```javascript
// Decode ID token
const idToken = jwt.decode(tokenResponse.id_token);

// Store sid with the session
req.session.sid = idToken.sid;
req.session.userId = idToken.sub;
```

#### Using the sid for logout

When you receive a logout notification:

```javascript
app.get('/logout', (req, res) => {
  const { sid } = req.query;
  
  // Find all sessions with this sid and terminate them
  sessionStore.findBySid(sid).forEach(session => {
    session.destroy();
  });
  
  res.send('<html><body>Logged out</body></html>');
});
```

---

## Discovery Support

The front-channel logout feature is advertised in the OpenID Provider Configuration endpoint:

```
GET /.well-known/openid-configuration
```

Response includes:
```json
{
  "issuer": "https://auth.example.com",
  "end_session_endpoint": "https://auth.example.com/logout",
  "frontchannel_logout_supported": true,
  "frontchannel_logout_session_supported": true,
  ...
}
```

---

## Security Considerations

### 1. Validate the Issuer

Always validate that the `iss` parameter matches your expected OP issuer:

```javascript
if (iss !== expectedIssuer) {
  return res.status(400).send('Invalid issuer');
}
```

### 2. Use HTTPS

Front-channel logout URIs MUST use HTTPS in production to prevent session hijacking.

### 3. CORS Configuration

The logout endpoint is loaded in an iframe, so configure CORS appropriately:

```javascript
app.get('/logout', (req, res) => {
  // Set appropriate CORS headers
  res.header('Access-Control-Allow-Origin', 'https://auth.example.com');
  res.header('Access-Control-Allow-Credentials', 'true');
  
  // ... logout logic
});
```

### 4. X-Frame-Options

Ensure your logout endpoint allows framing from the OP:

```javascript
// Do NOT set X-Frame-Options: DENY on logout endpoint
// Or use Content-Security-Policy frame-ancestors
res.header('Content-Security-Policy', "frame-ancestors 'self' https://auth.example.com");
```

### 5. Timeout Handling

The OP waits up to 5 seconds for all logout iframes to load. Ensure your logout handler responds quickly.

---

## Testing

### Testing the Logout Flow

1. **Set up a test RP** with a `frontchannel_logout_uri`
2. **Login to the test RP** using the authorization code flow
3. **Verify the ID token contains a `sid` claim**
4. **Initiate logout** at the OP: `GET /logout?post_logout_redirect_uri=...`
5. **Verify the RP's logout endpoint is called** with `iss` and `sid` parameters
6. **Verify the RP's session is terminated**
7. **Verify the user is redirected** to the `post_logout_redirect_uri`

### Example Test Scenario

```bash
# 1. Register a test client
curl -X POST https://auth.example.com/register \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "Test App",
    "redirect_uris": ["http://localhost:3000/callback"],
    "frontchannel_logout_uri": "http://localhost:3000/logout",
    "frontchannel_logout_session_required": true
  }'

# 2. Start authorization flow
open "https://auth.example.com/authorize?client_id=...&redirect_uri=...&response_type=code&scope=openid"

# 3. After login, decode ID token to get sid
echo $ID_TOKEN | jwt decode -

# 4. Initiate logout
open "https://auth.example.com/logout?post_logout_redirect_uri=http://localhost:3000/logged-out"

# 5. Check RP received logout notification
curl http://localhost:3000/logout?iss=https://auth.example.com&sid=SESSION_ID
```

---

## Troubleshooting

### Logout notifications not received

**Problem:** The RP's logout endpoint is not being called during logout.

**Solutions:**
1. Verify the client has a valid `frontchannel_logout_uri` registered
2. Check that the RP's logout endpoint is accessible from the browser
3. Ensure CORS is properly configured on the RP's logout endpoint
4. Check browser console for errors (X-Frame-Options, CSP violations)

### Session ID (sid) not included

**Problem:** The `sid` parameter is not included in logout notifications.

**Solutions:**
1. Ensure `frontchannel_logout_session_required` is set to `true` in client registration
2. Verify the client is properly registered with this setting

### Logout page shows indefinitely

**Problem:** The logout page with iframes keeps showing without redirecting.

**Solutions:**
1. Check that all RP logout endpoints respond within 5 seconds
2. Verify network connectivity to all registered RPs
3. Check browser console for JavaScript errors

### Sessions not terminating at RP

**Problem:** The RP receives logout notifications but sessions remain active.

**Solutions:**
1. Verify the RP is correctly mapping `sid` to local sessions
2. Ensure the RP stores the `sid` claim from the ID token during login
3. Check that the RP's session termination logic is working correctly

---

## Example: Complete Integration

Here's a complete example of integrating front-channel logout in a Node.js application:

### 1. Client Registration

```javascript
// Register client with front-channel logout support
const registration = {
  client_name: 'My App',
  redirect_uris: ['http://localhost:3000/callback'],
  frontchannel_logout_uri: 'http://localhost:3000/logout',
  frontchannel_logout_session_required: true
};

const response = await fetch('https://auth.example.com/register', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(registration)
});

const client = await response.json();
```

### 2. Login Handler

```javascript
app.get('/callback', async (req, res) => {
  const { code } = req.query;
  
  // Exchange code for tokens
  const tokenResponse = await getTokens(code);
  const idToken = jwt.decode(tokenResponse.id_token);
  
  // Store session with sid
  req.session.userId = idToken.sub;
  req.session.sid = idToken.sid;  // Store sid for logout
  
  res.redirect('/dashboard');
});
```

### 3. Logout Handler

```javascript
app.get('/logout', (req, res) => {
  const { iss, sid } = req.query;
  
  // Validate issuer
  if (iss && iss !== 'https://auth.example.com') {
    return res.status(400).send('Invalid issuer');
  }
  
  // Terminate session by sid
  if (sid) {
    sessionStore.all((err, sessions) => {
      Object.keys(sessions).forEach(sessionId => {
        if (sessions[sessionId].sid === sid) {
          sessionStore.destroy(sessionId);
        }
      });
    });
  }
  
  // Return empty page for iframe
  res.send('<html><body>Logged out</body></html>');
});
```

### 4. Initiate Logout

```javascript
app.get('/initiate-logout', (req, res) => {
  const logoutUrl = 'https://auth.example.com/logout';
  const params = new URLSearchParams({
    post_logout_redirect_uri: 'http://localhost:3000/logged-out',
    state: generateState()
  });
  
  res.redirect(`${logoutUrl}?${params}`);
});
```

---

## API Reference

### Models

#### Client
```go
type Client struct {
    // ... other fields ...
    
    // Front-Channel Logout
    FrontchannelLogoutURI              string `json:"frontchannel_logout_uri,omitempty"`
    FrontchannelLogoutSessionRequired bool   `json:"frontchannel_logout_session_required,omitempty"`
}
```

#### SessionClient
```go
type SessionClient struct {
    ID        string    `json:"id"`
    SessionID string    `json:"session_id"` // UserSession ID
    ClientID  string    `json:"client_id"`  // Client ID
    Sid       string    `json:"sid"`        // Session ID in ID token
    CreatedAt time.Time `json:"created_at"`
}
```

### Storage Interface

```go
// SessionClient operations (for front-channel logout)
CreateSessionClient(sc *SessionClient) error
GetSessionClientsBySessionID(sessionID string) ([]*SessionClient, error)
DeleteSessionClient(id string) error
DeleteSessionClientsBySessionID(sessionID string) error
```

### ID Token Claims

```go
type IDTokenClaims struct {
    jwt.RegisteredClaims
    // ... other claims ...
    
    Sid string `json:"sid,omitempty"` // Session ID for front-channel logout
}
```

---

## Resources

- [OpenID Connect Front-Channel Logout 1.0](https://openid.net/specs/openid-connect-frontchannel-1_0.html)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
- [Implementation Plan](FRONT_CHANNEL_LOGOUT_PLAN.md)
