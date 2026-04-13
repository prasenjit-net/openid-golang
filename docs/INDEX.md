# OpenID Connect Identity Server — Documentation

A lightweight, production-ready OIDC identity provider in Go with an embedded React admin UI.

## Guides

| Document | Description |
|---|---|
| [Getting Started](GETTING_STARTED.md) | Step-by-step first-run guide |
| [Docker](DOCKER.md) | Running with Docker / Docker Compose |
| [Configuration](CONFIGURATION.md) | All configuration options |
| [Contributing](CONTRIBUTING.md) | Development workflow and guidelines |

## Reference

| Document | Description |
|---|---|
| [API Reference](API.md) | Complete endpoint documentation |
| [Architecture](ARCHITECTURE.md) | System design, data flow diagrams |
| [Storage Backends](STORAGE.md) | JSON file vs MongoDB |

## Feature Highlights

- **Authorization Code Flow** with PKCE
- **Implicit Flow** (`id_token`, `token id_token`)
- **Token Revocation** and **Introspection**
- **Dynamic Client Registration** (RFC 7591/7592)
- **RSA signing keys** with X.509 certificates
  - KID derived from cert SHA-256 thumbprint (`x5t#S256`)
  - Generate CSR → submit to CA → import signed cert
  - JWKS includes `x5c` and `x5t#S256` per RFC 7517
- **Key rotation** — old keys stay in JWKS until cert expiry
- **Audit logging** — structured records for all security events
- **Token management** — search, filter, revoke via admin UI
- **Modern admin UI** — light/dark theme, responsive design
