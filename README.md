# OpenID Connect Identity Server

[![CI](https://github.com/prasenjit-net/openid-golang/actions/workflows/ci.yml/badge.svg)](https://github.com/prasenjit-net/openid-golang/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/prasenjit-net/openid-golang)](https://github.com/prasenjit-net/openid-golang/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/prasenjit-net/openid-golang)](go.mod)
[![License](https://img.shields.io/github/license/prasenjit-net/openid-golang)](LICENSE)

A lightweight, production-ready **OpenID Connect (OIDC) identity provider** written in Go with an embedded React admin UI.  
No external key-management tooling needed — RSA key pairs and X.509 certificates are generated in pure Go.

---

## ✨ Features

### OpenID Connect / OAuth 2.0
| Feature | Status |
|---|---|
| Authorization Code Flow | ✅ |
| Implicit Flow (id_token, token id_token) | ✅ |
| PKCE (S256 + plain) | ✅ |
| Refresh Token Grant | ✅ |
| Token Revocation (`/revoke`) | ✅ |
| Token Introspection (`/introspect`) | ✅ |
| Dynamic Client Registration (RFC 7591/7592) | ✅ |
| OpenID Connect Discovery (`/.well-known/openid-configuration`) | ✅ |
| JWKS Endpoint (`/.well-known/jwks.json`) with `x5c` / `x5t#S256` | ✅ |
| Scopes: `openid`, `profile`, `email`, `address`, `phone`, `offline_access` | ✅ |
| Nonce replay protection | ✅ |
| `auth_time` claim | ✅ |

### Signing Keys
| Feature | Status |
|---|---|
| RSA 2048-bit key generation (pure Go) | ✅ |
| Self-signed X.509 certificate per key | ✅ |
| KID derived from cert SHA-256 fingerprint (`x5t#S256`, RFC 7517) | ✅ |
| Configurable certificate validity (default 90 days) | ✅ |
| Key rotation — old keys kept until cert expires | ✅ |
| Generate PKCS#10 CSR for CA submission | ✅ |
| Import CA-signed certificate (replaces self-signed) | ✅ |
| JWKS serves all valid keys with `x5c` + `x5t#S256` | ✅ |

### Storage
| Backend | Use case |
|---|---|
| JSON file | Development, small deployments |
| MongoDB | Production, high-traffic |

### Admin UI
- Modern **"Secure Slate"** design with light / dark theme
- Dashboard with live statistics
- User management (CRUD, password reset)
- OAuth client management + secret rotation
- Token management — search, filter, revoke
- Signing key management — rotate, generate CSR, import cert
- Audit log viewer
- Server settings
- First-run setup wizard

---

## 🚀 Quick Start

### Docker (recommended)

```bash
# JSON-file storage (no external dependencies)
docker-compose --profile json-storage up -d

# MongoDB storage
MONGO_USER=admin MONGO_PASSWORD=secret docker-compose --profile with-mongodb up -d
```

The server starts at **http://localhost:8080**.  
On first launch, the setup wizard runs automatically at `http://localhost:8080/setup`.

### Binary

```bash
# Download from GitHub Releases and make executable
chmod +x openid-server-*

# Interactive first-run wizard (creates config + admin user)
./openid-server-* --setup

# Start
./openid-server-*
```

### Development

```bash
# Install Go + Node dependencies, build frontend, run server
make deps
make run
```

---

## 📁 Project Structure

```
openid-golang/
├── cmd/
│   ├── root.go          # CLI entry point (cobra)
│   └── serve.go         # Server startup, route registration
├── pkg/
│   ├── configstore/     # Configuration loading (TOML / env / MongoDB)
│   ├── crypto/          # RSA keygen, cert gen, CSR, JWT, PKCE, bcrypt
│   ├── handlers/
│   │   ├── admin.go     # Admin REST API (users, clients, keys, tokens, audit)
│   │   ├── authorize.go # /authorize endpoint
│   │   ├── bootstrap.go # Setup wizard handler
│   │   ├── discovery.go # /.well-known/* + JWKS
│   │   ├── handlers.go  # Shared handler wiring
│   │   ├── token.go     # /token, /revoke, /introspect
│   │   └── userinfo.go  # /userinfo
│   ├── middleware/      # JWT auth middleware for OIDC endpoints
│   ├── models/          # Data models (User, Client, Token, SigningKey, AuditLog…)
│   ├── session/         # Server-side session store + cookie middleware
│   └── storage/
│       ├── storage.go   # Storage interface
│       ├── json.go      # JSON file implementation
│       └── mongodb.go   # MongoDB implementation
├── frontend/            # React + TypeScript + Ant Design admin UI
│   └── src/
│       ├── pages/       # Dashboard, Users, Clients, Tokens, KeyManagement, AuditLog…
│       └── hooks/       # useApi.ts — all React Query hooks
├── public/              # Embedded HTML templates (login, consent, setup wizard)
├── embed.go             # go:embed declarations
├── main.go
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

---

## 🔧 Configuration

The server reads configuration from (in priority order):
1. Environment variables
2. `data/config.json` (written by setup wizard)
3. Built-in defaults

### Key environment variables

| Variable | Default | Description |
|---|---|---|
| `SERVER_HOST` | `0.0.0.0` | Listen address |
| `SERVER_PORT` | `8080` | Listen port |
| `MONGODB_URI` | — | MongoDB connection string (enables MongoDB storage) |
| `MONGODB_DATABASE` | `openid` | MongoDB database name |

When `MONGODB_URI` is set, all data (config + storage) is persisted to MongoDB.  
Without it, the JSON file backend (`data/openid.json`) is used.

### First-run Setup Wizard

Visit **`http://localhost:8080/setup`** (or pass `--setup` to the binary) to:
- Set the issuer URL
- Choose storage backend
- Create the first admin user
- Optionally pre-create an OAuth client

---

## 🔐 OpenID Connect Endpoints

| Endpoint | Method | Description |
|---|---|---|
| `/.well-known/openid-configuration` | GET | OIDC Discovery document |
| `/.well-known/jwks.json` | GET | JSON Web Key Set (all valid keys) |
| `/authorize` | GET | Authorization endpoint |
| `/token` | POST | Token endpoint |
| `/userinfo` | GET / POST | UserInfo endpoint |
| `/revoke` | POST | Token revocation (RFC 7009) |
| `/introspect` | POST | Token introspection (RFC 7662) |
| `/login` | GET / POST | Login page (rendered server-side) |
| `/consent` | GET / POST | Consent page (rendered server-side) |

### Dynamic Client Registration

Enabled by default at `/register`:

| Endpoint | Method | Description |
|---|---|---|
| `/register` | POST | Register a new client (RFC 7591) |
| `/register/:client_id` | GET | Read registration (RFC 7592) |
| `/register/:client_id` | PUT | Update registration (RFC 7592) |
| `/register/:client_id` | DELETE | Delete registration |

---

## 🛠️ Admin API

All admin API routes are under `/api/` and require a Bearer token obtained via `POST /api/auth/login`.

> The admin portal uses session-based authentication (no token issued) to keep admin sessions out of the OIDC token store.

### Auth

| Method | Path | Description |
|---|---|---|
| POST | `/api/auth/login` | Admin login |
| POST | `/api/auth/logout` | Admin logout |

### Users

| Method | Path | Description |
|---|---|---|
| GET | `/api/users` | List users |
| POST | `/api/users` | Create user |
| GET | `/api/users/:id` | Get user |
| PUT | `/api/users/:id` | Update user |
| DELETE | `/api/users/:id` | Delete user |

### OAuth Clients

| Method | Path | Description |
|---|---|---|
| GET | `/api/clients` | List clients |
| POST | `/api/clients` | Create client |
| GET | `/api/clients/:id` | Get client |
| PUT | `/api/clients/:id` | Update client |
| DELETE | `/api/clients/:id` | Delete client |
| POST | `/api/clients/:id/regenerate-secret` | Rotate client secret |

### Signing Keys

| Method | Path | Description |
|---|---|---|
| GET | `/api/keys` | List all signing keys with cert details |
| POST | `/api/settings/rotate-keys` | Rotate active key (body: `{"validity_days":90}`) |
| GET | `/api/keys/:id/csr` | Generate PKCS#10 CSR for the key |
| POST | `/api/keys/:id/import-cert` | Import CA-signed cert (body: `{"cert_pem":"..."}`) |

### Tokens

| Method | Path | Description |
|---|---|---|
| GET | `/api/tokens` | List tokens (filter by subject, client, type, active) |
| DELETE | `/api/tokens/:id` | Revoke token |

### Audit Log

| Method | Path | Description |
|---|---|---|
| GET | `/api/audit` | Query audit log (filter by action, actor, date range) |

### Settings

| Method | Path | Description |
|---|---|---|
| GET | `/api/settings` | Get server settings |
| PUT | `/api/settings` | Update server settings |

---

## 🔑 Signing Key Lifecycle

```
Generate key pair + self-signed cert
           │
           ▼
    KID = x5t#S256(cert)        ← RFC 7517 §4.9
    ExpiresAt = cert.NotAfter
    JWKS includes x5c + x5t#S256
           │
           ├── (optional) Generate CSR ──► Submit to CA
           │                                    │
           │                         Receive signed cert
           │                                    │
           └────────────── Import Cert ◄─────────
                                │
                    KID re-derived from new cert
                    ExpiresAt updated from new cert
                    JWKS updated automatically
           │
           ▼
    Rotate Key
    Old key stays in JWKS until its cert expires
    New key becomes active
```

---

## 📋 Development Commands

```bash
make build-all     # Build frontend then Go binary
make build         # Build Go binary only (requires frontend/dist)
make build-frontend # Build React admin UI
make run           # Build frontend + run server
make test          # Run all Go tests
make lint          # Run golangci-lint
make fmt           # Run gofmt
make deps          # Download Go + npm dependencies
make clean         # Remove build artifacts
make install-tools # Install golangci-lint v2.5.0
```

---

## 🧪 Test Client

An interactive test client for exploring OIDC flows is included:

```bash
go run examples/test-client.go
# Visit http://localhost:9090
```

The test client walks through:
- Dynamic client registration
- Authorization Code Flow with PKCE
- Implicit Flow
- Token refresh
- UserInfo fetch
- Token introspection

---

## 📊 Audit Events

All security-relevant operations produce structured audit log entries:

| Category | Actions |
|---|---|
| **User** | `user.login`, `user.login_failed`, `user.consent_granted`, `user.consent_denied` |
| **Token** | `token.issued`, `token.revoked` |
| **Client** | `client.registered` |
| **Admin** | `admin.login`, `admin.user.*`, `admin.client.*`, `admin.settings.updated`, `admin.keys.rotated` |

Each entry records: timestamp, action, actor (type + ID), resource, status, IP address, user agent, and optional metadata.

---

## 🐳 Docker

```bash
# Build image locally
docker build -t openid-server .

# JSON-file storage (development)
docker-compose --profile json-storage up -d

# MongoDB-backed (production-like)
MONGO_USER=admin MONGO_PASSWORD=secret \
  docker-compose --profile with-mongodb up -d

# Logs
docker logs openid-server -f
```

Data is persisted to the `./data` volume mount.

---

## 🏗️ Storage Backends

### JSON File (default)

- Zero dependencies — single file on disk (`data/openid.json`)
- Suitable for development and small single-instance deployments
- Not suitable for multi-instance or high-write workloads

### MongoDB

Set `MONGODB_URI` to switch. Supports:
- Multi-instance deployments
- High-throughput workloads
- TTL indexes for automatic token expiry

---

## 🔒 Security Notes

- Passwords hashed with **bcrypt** (cost 10)
- PKCE enforced for public clients
- Nonce stored and checked to prevent replay attacks
- `auth_time` propagated through session for `max_age` enforcement
- Admin session uses server-side session store (not OIDC tokens)
- Signing keys backed by X.509 certificates; KIDs are RFC 7517-compliant thumbprints
- Old signing keys retained in JWKS until certificate expiry to avoid token validation gaps

For production hardening, additionally consider: TLS termination, rate limiting, MongoDB authentication, and HSM/KMS for key storage.

---

## 📄 License

MIT — free for personal and commercial use.

## 🤝 Contributing

Pull requests and issues are welcome. See `docs/CONTRIBUTING.md` for guidelines.
