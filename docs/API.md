# OpenID Connect Identity Server - API Documentation

## Endpoints

### OpenID Connect Discovery

#### GET /.well-known/openid-configuration

Returns OpenID Connect discovery information.

**Response:**
```json
{
  "issuer": "http://localhost:8080",
  "authorization_endpoint": "http://localhost:8080/authorize",
  "token_endpoint": "http://localhost:8080/token",
  "userinfo_endpoint": "http://localhost:8080/userinfo",
  "jwks_uri": "http://localhost:8080/.well-known/jwks.json",
  "scopes_supported": ["openid", "profile", "email"],
  "response_types_supported": ["code"],
  "grant_types_supported": ["authorization_code", "refresh_token"]
}
```

### Authorization Endpoint

#### GET /authorize

Initiates the authorization flow.

**Parameters:**
- `client_id` (required) - The client identifier
- `redirect_uri` (required) - Redirect URI after authorization
- `response_type` (required) - Must be "code"
- `scope` (required) - Space-separated scopes, must include "openid"
- `state` (recommended) - State parameter for CSRF protection
- `nonce` (optional) - Nonce for ID token replay protection
- `code_challenge` (optional) - PKCE code challenge
- `code_challenge_method` (optional) - PKCE method (plain or S256)

**Example:**
```
GET /authorize?client_id=abc123&redirect_uri=http://localhost:3000/callback&response_type=code&scope=openid%20profile%20email&state=xyz
```

### Token Endpoint

#### POST /token

Exchanges authorization code for tokens.

**Content-Type:** application/x-www-form-urlencoded

**Parameters:**

For authorization code grant:
- `grant_type=authorization_code`
- `code` - The authorization code
- `redirect_uri` - Must match the redirect URI from authorization request
- `client_id` - The client identifier
- `client_secret` - The client secret
- `code_verifier` - PKCE code verifier (if code_challenge was used)

For refresh token grant:
- `grant_type=refresh_token`
- `refresh_token` - The refresh token
- `client_id` - The client identifier
- `client_secret` - The client secret

**Authentication:**
Client credentials can be sent via:
1. Form parameters (`client_id` and `client_secret`)
2. HTTP Basic Auth (recommended)

**Response:**
```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "...",
  "id_token": "..."
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "client_id:client_secret" \
  -d "grant_type=authorization_code&code=abc123&redirect_uri=http://localhost:3000/callback"
```

### UserInfo Endpoint

#### GET /userinfo

Returns user information for the authenticated user.

**Headers:**
- `Authorization: Bearer <access_token>`

**Response:**
```json
{
  "sub": "user-id",
  "name": "Test User",
  "given_name": "Test",
  "family_name": "User",
  "email": "test@example.com",
  "picture": "https://example.com/photo.jpg"
}
```

**Example:**
```bash
curl http://localhost:8080/userinfo \
  -H "Authorization: Bearer <access_token>"
```

### JWKS Endpoint

#### GET /.well-known/jwks.json

Returns the JSON Web Key Set for token verification.

**Response:**
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "default",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

## Authorization Code Flow

1. **Client initiates authorization:**
   ```
   GET /authorize?client_id=...&redirect_uri=...&response_type=code&scope=openid profile email&state=...
   ```

2. **User authenticates and consents**

3. **Server redirects back to client with code:**
   ```
   http://localhost:3000/callback?code=abc123&state=...
   ```

4. **Client exchanges code for tokens:**
   ```
   POST /token
   grant_type=authorization_code&code=abc123&redirect_uri=...
   ```

5. **Server returns tokens:**
   ```json
   {
     "access_token": "...",
     "id_token": "...",
     "refresh_token": "..."
   }
   ```

6. **Client uses access token to get user info:**
   ```
   GET /userinfo
   Authorization: Bearer <access_token>
   ```

## PKCE Flow

For public clients (mobile apps, SPAs), use PKCE:

1. Generate code verifier and challenge:
   ```javascript
   const verifier = generateRandomString(128);
   const challenge = base64url(sha256(verifier));
   ```

2. Include in authorization request:
   ```
   GET /authorize?...&code_challenge=<challenge>&code_challenge_method=S256
   ```

3. Include verifier in token request:
   ```
   POST /token
   ...&code_verifier=<verifier>
   ```

## Error Responses

All errors follow the OAuth 2.0 error response format:

```json
{
  "error": "invalid_request",
  "error_description": "Missing required parameter: client_id"
}
```

Common error codes:
- `invalid_request` - Malformed request
- `invalid_client` - Invalid client credentials
- `invalid_grant` - Invalid authorization code or refresh token
- `unauthorized_client` - Client not authorized
- `unsupported_grant_type` - Grant type not supported
- `invalid_scope` - Invalid scope requested
