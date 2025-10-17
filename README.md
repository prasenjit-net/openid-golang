# OpenID Connect Identity Server

A lightweight OpenID Connect (OIDC) identity provider implementation in Go with an embedded React admin UI.

> 💡 **Tip:** Run `./show-docs.sh` to see the complete documentation structure!

## 📚 Documentation

All documentation is organized in the **`docs/`** folder:

- **[Getting Started Guide](docs/GETTING_STARTED.md)** - Step-by-step setup tutorial ⭐ START HERE
- **[Quick Start](docs/QUICKSTART.md)** - Quick reference for experienced developers
- **[API Documentation](docs/API.md)** - Complete API reference
- **[Architecture](docs/ARCHITECTURE.md)** - System architecture and diagrams
- **[Testing Guide](docs/TESTING.md)** - How to test the server
- **[Implementation Details](docs/IMPLEMENTATION.md)** - Technical details
- **[Project Summary](docs/PROJECT_SUMMARY.md)** - What's been built
- **[Documentation Index](docs/INDEX.md)** - Complete documentation hub
- **[Admin UI Documentation](ui/admin/README.md)** - React admin interface guide

## 🚀 Quick Start

```bash
# 1. Setup (generates keys, downloads dependencies)
./setup.sh

# 2. Create test data (user and OAuth client)
go run scripts/seed.go

# 3. Start the server
./test.sh

# OR - Build with embedded admin UI
./build.sh
./bin/openid-server
```

Visit http://localhost:8080/health to verify the server is running.
Access the admin UI at http://localhost:8080/

## ✨ Features

- **OpenID Connect Core 1.0** implementation
- Authorization Code Flow with PKCE support
- Token endpoint for access tokens, refresh tokens, and ID tokens
- UserInfo endpoint
- OpenID Connect Discovery (/.well-known/openid-configuration)
- JWT-based ID tokens (RS256 signing)
- Client authentication and management
- User authentication with bcrypt password hashing
- **React Admin UI** with:
  - User management
  - OAuth client registration and management
  - Server settings configuration
  - Signing key rotation
  - Initial setup wizard
  - Dashboard with statistics

## Project Structure

```
.
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers for OIDC endpoints
│   ├── middleware/      # HTTP middleware (logging, auth, etc.)
│   ├── models/          # Data models
│   ├── storage/         # Database/storage interfaces
│   └── crypto/          # Cryptographic utilities (JWT, signing)
├── pkg/
│   └── oidc/            # Public OIDC utilities
├── config/              # Configuration files
├── go.mod
└── README.md
```

## 📋 Prerequisites

- Go 1.21 or higher
- OpenSSL (for key generation)
- SQLite (for development) or PostgreSQL (for production)

## 🛠️ Installation

Run the automated setup script:

```bash
./setup.sh
```

This will:
- Create necessary directories
- Generate RSA key pairs for JWT signing
- Create `.env` configuration file
- Download Go dependencies
- Build the application

## 🔧 Configuration

Configure via `.env` file or environment variables:

```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
DB_TYPE=sqlite
DB_CONNECTION=./openid.db
JWT_PRIVATE_KEY=config/keys/private.key
JWT_PUBLIC_KEY=config/keys/public.key
JWT_EXPIRY_MINUTES=60
ISSUER=http://localhost:8080
```

## 🧪 Testing

Create test data:
```bash
go run scripts/seed.go
```

Start the server:
```bash
make run
# or
./test.sh
```

Run the test OAuth client:
```bash
go run examples/test-client.go
```

Then visit http://localhost:9090 in your browser.

## OpenID Connect Endpoints

- **Discovery**: `GET /.well-known/openid-configuration`
- **Authorization**: `GET /authorize`
- **Token**: `POST /token`
- **UserInfo**: `GET /userinfo`
- **JWKS**: `GET /.well-known/jwks.json`

## Configuration

The server can be configured using a YAML configuration file or environment variables.

Key configuration options:
- `server.host` - Server host address
- `server.port` - Server port
- `issuer` - OIDC issuer URL
- `database.type` - Database type (sqlite, postgres)
- `database.connection` - Database connection string
- `jwt.signing_key` - JWT signing key (RS256)

## 💻 Development Commands

```bash
make build      # Build the application
make run        # Run the server
make test       # Run tests
make fmt        # Format code
make clean      # Clean build artifacts
make deps       # Download dependencies
```

## 🔐 Security Considerations

⚠️ This is a development/learning implementation. For production:

- Use HTTPS/TLS
- Switch to PostgreSQL
- Implement proper session management
- Add rate limiting
- Enable CSRF protection
- Implement account lockout
- Add audit logging
- Use HSM/KMS for key management
- Add monitoring and backups

See [docs/QUICKSTART.md](docs/QUICKSTART.md) for complete production checklist.

## 📖 Learn More

- Read the [Getting Started Guide](docs/GETTING_STARTED.md) for detailed instructions
- Check the [API Documentation](docs/API.md) for endpoint details
- Study the [Architecture](docs/ARCHITECTURE.md) to understand the design
- Follow the [Testing Guide](docs/TESTING.md) to test all features

## 📄 License

MIT License - Free for personal and commercial use

## 🤝 Contributing

Contributions are welcome! The codebase is well-structured and documented.

- Check [docs/IMPLEMENTATION.md](docs/IMPLEMENTATION.md) for technical details
- Review the code - it's well-commented
- Submit issues and pull requests

## 🌟 Standards Compliance

- OpenID Connect Core 1.0
- OAuth 2.0 (RFC 6749)
- JWT (RFC 7519)
- PKCE (RFC 7636)
- JWKS (RFC 7517)
