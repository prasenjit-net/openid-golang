# API Reference

## Base URL

All endpoints are relative to the server base URL (e.g. `http://localhost:8080`).

---

## OpenID Connect Endpoints

### `GET /.well-known/openid-configuration`

Returns the OpenID Connect Discovery document.

**Response (200)**
```json
{
  "issuer": "http://localhost:8080",
  "authorization_endpoint": "http://localhost:8080/authorize",
  "token_endpoint": "http://localhost:8080/token",
  "userinfo_endpoint": "http://localhost:8080/userinfo",
  "jwks_uri": "http://localhost:8080/.well-known/jwks.json",
  "revocation_endpoint": "http://localhost:8080/revoke",
  "introspection_endpoint": "http://localhost:8080/introspect",
  "response_types_supported": ["code", "id_token", "token id_token"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "scopes_supported": ["openid", "profile", "email", "address", "phone", "offline_access"],
  "grant_types_supported": ["authorization_code", "implicit", "refresh_token"],
  "token_endpoint_auth_methods_supported": ["client_secret_basic", "client_secret_post"],
  "claims_supported": ["sub", "iss", "aud", "exp", "iat", "name", "email", ...]
}
```

---

### `GET /.well-known/jwks.json`

Returns the JSON Web Key Set used to verify ID tokens.  
All non-expired signing keys are included. Each key carries `x5c` (certificate chain) and `x5t#S256` (SHA-256 cert thumbprint) fields per RFC 7517.

**Response (200)**
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "base64url(SHA-256(DER cert))",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB",
      "x5c": ["base64(DER)"],
      "x5t#S256": "base64url(SHA-256(DER))"
    }
  ]
}
```

---

### `GET /authorize`

Initiates an authorization flow. Renders the login page if the user is not authenticated, then the consent page.

**Query parameters**

| Parameter | Required | Description |
|---|---|---|
| `client_id` | ✅ | Registered client identifier |
| `redirect_uri` | ✅ | Must match a registered redirect URI |
| `response_type` | ✅ | `code`, `id_token`, or `token id_token` |
| `scope` | ✅ | Space-separated; must include `openid` |
| `state` | recommended | Opaque value returned in redirect |
| `nonce` | recommended | Included in ID token; used for replay protection |
| `code_challenge` | optional | PKCE challenge (Base64URL-encoded SHA-256) |
| `code_challenge_method` | optional | `S256` (recommended) or `plain` |
| `prompt` | optional | `login` forces re-authentication |
| `max_age` | optional | Maximum age of authentication in seconds |

**Success redirect (authorization_code)**
```
GET {redirect_uri}?code={code}&state={state}
```

**Error redirect**
```
GET {redirect_uri}?error=access_denied&error_description=...&state={state}
```

---

### `POST /token`

Exchanges an authorization code or refresh token for access/ID/refresh tokens.

**Content-Type:** `application/x-www-form-urlencoded`

**Client authentication** (pick one):
- `Authorization: Basic base64(client_id:client_secret)`
- Form fields: `client_id` + `client_secret`

#### Authorization Code Grant

| Parameter | Required | Description |
|---|---|---|
| `grant_type` | ✅ | `authorization_code` |
| `code` | ✅ | Code from `/authorize` redirect |
| `redirect_uri` | ✅ | Must match the one used in `/authorize` |
| `code_verifier` | if PKCE used | Plain PKCE verifier string |

#### Refresh Token Grant

| Parameter | Required | Description |
|---|---|---|
| `grant_type` | ✅ | `refresh_token` |
| `refresh_token` | ✅ | A valid refresh token |
| `scope` | optional | Subset of original scopes |

**Response (200)**
```json
{
  "access_token": "eyJ...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "id_token": "eyJ...",
  "refresh_token": "eyJ..."
}
```

---

### `GET /userinfo` · `POST /userinfo`

Returns claims about the authenticated user. Requires `Authorization: Bearer {access_token}`.

**Response (200)** — claims depend on granted scopes:
```json
{
  "sub": "user-uuid",
  "name": "Jane Doe",
  "given_name": "Jane",
  "family_name": "Doe",
  "email": "jane@example.com",
  "email_verified": true
}
```

---

### `POST /revoke`

Revokes an access or refresh token (RFC 7009).

| Parameter | Required | Description |
|---|---|---|
| `token` | ✅ | The token to revoke |
| `token_type_hint` | optional | `access_token` or `refresh_token` |

Returns `200 OK` even if the token was already revoked or not found.

---

### `POST /introspect`

Returns metadata about a token (RFC 7662).

| Parameter | Required | Description |
|---|---|---|
| `token` | ✅ | The token to inspect |

**Response (200)**
```json
{
  "active": true,
  "sub": "user-uuid",
  "client_id": "my-app",
  "scope": "openid profile email",
  "exp": 1712345678
}
```

---

## Dynamic Client Registration (RFC 7591 / 7592)

### `POST /register`

Register a new OAuth client without authentication (or with initial access token if configured).

**Request body (JSON)**
```json
{
  "client_name": "My App",
  "redirect_uris": ["http://localhost:3000/callback"],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "token_endpoint_auth_method": "client_secret_basic",
  "scope": "openid profile email"
}
```

**Response (201)**
```json
{
  "client_id": "abc123",
  "client_secret": "s3cr3t",
  "client_id_issued_at": 1712345678,
  "client_secret_expires_at": 0,
  ...
}
```

### `GET /register/:client_id`
Returns the current registration metadata. Requires `Authorization: Bearer {registration_access_token}`.

### `PUT /register/:client_id`
Updates the registration. Requires `Authorization: Bearer {registration_access_token}`.

### `DELETE /register/:client_id`
Deletes the registration. Requires `Authorization: Bearer {registration_access_token}`.

---

## Admin API

### Authentication

```http
POST /api/auth/login
Content-Type: application/json

{"username": "admin", "password": "..."}
```

**Response (200)**
```json
{"token": "eyJ..."}
```

All subsequent admin API calls require:
```
Authorization: Bearer {token}
```

---

### Users

| Method | Path | Body / Query | Description |
|---|---|---|---|
| GET | `/api/users` | `?search=` | List users |
| POST | `/api/users` | `{username, password, email, ...}` | Create user |
| GET | `/api/users/:id` | — | Get user |
| PUT | `/api/users/:id` | `{email, ...}` | Update user |
| DELETE | `/api/users/:id` | — | Delete user |

---

### OAuth Clients

| Method | Path | Description |
|---|---|---|
| GET | `/api/clients` | List clients |
| POST | `/api/clients` | Create client |
| GET | `/api/clients/:id` | Get client |
| PUT | `/api/clients/:id` | Update client |
| DELETE | `/api/clients/:id` | Delete client |
| POST | `/api/clients/:id/regenerate-secret` | Rotate client secret |

---

### Signing Keys

#### `GET /api/keys`

Lists all signing keys with certificate details.

**Response**
```json
[
  {
    "id": "uuid",
    "kid": "base64url(SHA-256(cert))",
    "algorithm": "RS256",
    "is_active": true,
    "created_at": "2026-01-01T00:00:00Z",
    "expires_at": "2026-04-01T00:00:00Z",
    "status": "active",
    "has_csr": false,
    "cert": {
      "subject": "openid-server",
      "issuer": "openid-server",
      "serial": "hex",
      "not_before": "2026-01-01T00:00:00Z",
      "not_after": "2026-04-01T00:00:00Z",
      "fingerprint": "base64url(SHA-256(DER))",
      "self_signed": true
    }
  }
]
```

#### `POST /api/settings/rotate-keys`

Generates a new RSA key + certificate, deactivates the current active key.

**Request body (JSON)**
```json
{"validity_days": 90}
```
`validity_days` defaults to 90 if omitted (range: 30–3650).

#### `GET /api/keys/:id/csr`

Generates and returns a PKCS#10 Certificate Signing Request for the specified key.  
The CSR PEM is also persisted on the key record (`has_csr = true`).

**Response**
```json
{
  "kid": "...",
  "csr_pem": "-----BEGIN CERTIFICATE REQUEST-----\n...\n-----END CERTIFICATE REQUEST-----\n"
}
```

#### `POST /api/keys/:id/import-cert`

Imports a CA-signed certificate for the key. Validates that the certificate's public key matches the stored private key. Re-derives `kid` from the new cert's SHA-256 thumbprint.

**Request body (JSON)**
```json
{
  "cert_pem": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n"
}
```

**Response (200)**
```json
{
  "message": "Certificate imported successfully",
  "kid": "new-kid-from-new-cert",
  "cert": {
    "subject": "my-service",
    "issuer": "My Internal CA",
    "serial": "abc123",
    "not_before": "...",
    "not_after": "...",
    "fingerprint": "..."
  }
}
```

**Error (422)** if cert does not match private key.

---

### Tokens

#### `GET /api/tokens`

Search and list tokens. Results are **not returned** unless at least one filter is provided.

**Query parameters**

| Parameter | Description |
|---|---|
| `subject` | Filter by user ID |
| `client_id` | Filter by client |
| `token_type` | `access_token` or `refresh_token` |
| `active_only` | `true` (default) — only non-expired, non-revoked tokens |

#### `DELETE /api/tokens/:id`

Revokes a token by its storage ID.

---

### Audit Log

#### `GET /api/audit`

**Query parameters**

| Parameter | Description |
|---|---|
| `action` | Filter by audit action (e.g. `user.login`) |
| `actor_type` | `user`, `client`, `admin`, `system` |
| `from` | ISO 8601 start time |
| `to` | ISO 8601 end time |
| `limit` | Max results (default 100) |

---

### Settings

#### `GET /api/settings`

Returns current server configuration (issuer, token lifetimes, registration settings, etc.).

#### `PUT /api/settings`

Updates server configuration. Changes take effect immediately.
