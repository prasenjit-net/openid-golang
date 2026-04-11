# Architecture

## Overview

The server is a single Go binary that embeds:
- The React admin UI (`frontend/dist` → `cmd/adminUIFS`)
- The server-rendered HTML pages for login, consent, and setup (`public/` → `publicFS`)

```
┌──────────────────────────────────────────────────────────────┐
│                        HTTP Layer (Echo v4)                   │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐  │
│  │  OIDC Routes   │  │  Admin Routes  │  │  Static Files  │  │
│  │  /authorize    │  │  /api/*        │  │  / (React SPA) │  │
│  │  /token        │  │  (JWT authed)  │  │  /setup        │  │
│  │  /userinfo     │  └────────────────┘  └────────────────┘  │
│  │  /.well-known  │                                           │
│  └────────────────┘                                           │
├──────────────────────────────────────────────────────────────┤
│                      Middleware Stack                         │
│   Logger │ CORS │ Recover │ Session middleware                │
├──────────────────────────────────────────────────────────────┤
│                      Handler Layer                            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────────┐  │
│  │Authorize │ │  Token   │ │ UserInfo │ │  Admin Handler │  │
│  │ Handler  │ │ Handler  │ │ Handler  │ │  (users,clients│  │
│  └──────────┘ └──────────┘ └──────────┘ │  keys, tokens, │  │
│  ┌──────────┐ ┌──────────┐              │  audit, settings│ │
│  │Discovery │ │Bootstrap │              └────────────────┘  │
│  │ /jwks    │ │  /setup  │                                   │
│  └──────────┘ └──────────┘                                   │
├──────────────────────────────────────────────────────────────┤
│                     Business Logic / Packages                 │
│  ┌──────────────┐  ┌───────────────┐  ┌───────────────────┐  │
│  │  pkg/crypto  │  │ pkg/session   │  │ pkg/configstore   │  │
│  │  RSA keygen  │  │ Server-side   │  │ TOML/env/MongoDB  │  │
│  │  Cert/CSR    │  │ session store │  │ config loading    │  │
│  │  JWT sign/   │  │ + cleanup     │  └───────────────────┘  │
│  │  verify      │  └───────────────┘                         │
│  │  PKCE        │                                             │
│  │  bcrypt      │                                             │
│  └──────────────┘                                             │
├──────────────────────────────────────────────────────────────┤
│                       Storage Layer                           │
│              storage.Storage (interface)                      │
│   ┌────────────────────┐     ┌────────────────────────────┐  │
│   │  JSON File Storage │     │    MongoDB Storage         │  │
│   │  data/openid.json  │     │  (TTL indexes, multi-inst) │  │
│   └────────────────────┘     └────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

---

## Authorization Code Flow

```
Browser                    OIDC Server                    Client App
  │                             │                             │
  │  GET /authorize?...         │                             │
  │────────────────────────────►│                             │
  │                             │ (validate params, create    │
  │                             │  auth session in session    │
  │                             │  store)                     │
  │  302 → /login               │                             │
  │◄────────────────────────────│                             │
  │  GET /login                 │                             │
  │────────────────────────────►│                             │
  │  login page HTML            │                             │
  │◄────────────────────────────│                             │
  │  POST /login (user+pass)    │                             │
  │────────────────────────────►│                             │
  │                             │ (verify bcrypt, mark        │
  │                             │  session authenticated)     │
  │  302 → /consent             │                             │
  │◄────────────────────────────│                             │
  │  POST /consent (allow)      │                             │
  │────────────────────────────►│                             │
  │                             │ (generate auth code,        │
  │                             │  store code → auth session) │
  │  302 → redirect_uri?code=   │                             │
  │◄────────────────────────────│                             │
  │                             │       POST /token           │
  │                             │◄────────────────────────────│
  │                             │  (exchange code for         │
  │                             │   access_token, id_token,   │
  │                             │   refresh_token)            │
  │                             │  200 {tokens}               │
  │                             │────────────────────────────►│
```

---

## Signing Key Lifecycle

```
GenerateSigningKeyWithCert(validityDays)
  │
  ├── rsa.GenerateKey(2048)
  ├── x509.CreateCertificate(self-signed, NotAfter = now + validityDays)
  ├── KID = base64url(SHA-256(DER cert))   ← x5t#S256 (RFC 7517 §4.9)
  └── persist SigningKey{PrivateKey, PublicKey, CertPEM, KID, ExpiresAt}

JWKS()
  └── GetAllSigningKeys()
        └── for each key where ExpiresAt > now:
              PublicKeyToJWKWithCert(pubKey, kid, certPEM)
                └── JWK{kty,use,kid,alg,n,e, x5c, x5t#S256}

RotateKeys(validityDays)
  ├── GenerateSigningKeyWithCert(validityDays)
  ├── mark old active key IsActive=false (ExpiresAt unchanged)
  └── persist new key as IsActive=true

GenerateKeyCSR(keyID)
  ├── load SigningKey
  ├── x509.CreateCertificateRequest(privKey, subject from existing cert)
  ├── persist csr_pem on SigningKey
  └── return csr_pem

ImportKeyCert(keyID, certPEM)
  ├── ValidateCertMatchesPrivateKey(certPEM, privKey)
  ├── new KID = CertThumbprintS256(certPEM)
  ├── new ExpiresAt = cert.NotAfter
  └── persist updated SigningKey
```

---

## Session Store

Authentication state during the OIDC flow (`/authorize` → `/login` → `/consent` → redirect) is held in a **server-side session store** (`pkg/session`).

- Sessions are stored in memory (keyed by a random cookie value).
- A background goroutine cleans up expired sessions.
- Cookies are `HttpOnly`, `SameSite=Lax`.

Admin authentication uses the same session store independently; no OIDC token is issued for admin logins.

---

## Storage Interface

```go
type Storage interface {
    // Users
    CreateUser(*models.User) error
    GetUser(id string) (*models.User, error)
    GetUserByUsername(username string) (*models.User, error)
    ListUsers(filter UserFilter) ([]*models.User, error)
    UpdateUser(*models.User) error
    DeleteUser(id string) error

    // OAuth Clients
    CreateClient(*models.Client) error
    GetClient(id string) (*models.Client, error)
    ListClients() ([]*models.Client, error)
    UpdateClient(*models.Client) error
    DeleteClient(id string) error

    // Tokens
    CreateToken(*models.Token) error
    GetToken(id string) (*models.Token, error)
    GetTokenByValue(value string) (*models.Token, error)
    ListTokens(filter TokenFilter) ([]*models.Token, error)
    UpdateToken(*models.Token) error
    DeleteToken(id string) error

    // Auth Sessions (code ↔ session binding)
    CreateAuthSession(*models.AuthSession) error
    GetAuthSession(id string) (*models.AuthSession, error)
    UpdateAuthSession(*models.AuthSession) error
    DeleteAuthSession(id string) error

    // Signing Keys
    CreateSigningKey(*models.SigningKey) error
    GetSigningKey(id string) (*models.SigningKey, error)
    GetSigningKeyByKID(kid string) (*models.SigningKey, error)
    GetAllSigningKeys() ([]*models.SigningKey, error)
    GetActiveSigningKey() (*models.SigningKey, error)
    UpdateSigningKey(*models.SigningKey) error
    DeleteSigningKey(id string) error

    // Audit Log
    CreateAuditLog(*models.AuditLog) error
    ListAuditLogs(filter AuditFilter) ([]*models.AuditLog, error)
}
```

Both `JSONStorage` and `MongoDBStorage` implement this interface. Swapping backends requires only a config change.

---

## Embedded Assets

```go
// embed.go
//go:embed frontend/dist
var adminUIFS embed.FS  // React SPA served at /

//go:embed public
var publicFS embed.FS   // login.html, consent.html, setup.html
```

The `public/` directory contains Go `html/template` files rendered server-side with per-request data (error messages, client name, scopes, etc.).

---

## Audit Logging

Every security-relevant operation calls `logAdminAudit` or the equivalent OIDC-path logger, which writes a structured `AuditLog` record to the storage backend.

Fields: `id`, `timestamp`, `action`, `actor_id`, `actor_type`, `resource_type`, `resource_id`, `status`, `ip_address`, `user_agent`, `metadata`.

Actors are resolved as follows:
- OIDC login / consent → user's username
- Admin API → admin username from session
- Dynamic registration → client ID
- System bootstrap → `"system"`
