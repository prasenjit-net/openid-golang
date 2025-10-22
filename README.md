# OpenID Connect Identity Server

[![CI](https://github.com/prasenjit-net/openid-golang/actions/workflows/ci.yml/badge.svg)](https://github.com/prasenjit-net/openid-golang/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/prasenjit-net/openid-golang)](https://github.com/prasenjit-net/openid-golang/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/prasenjit-net/openid-golang)](go.mod)
[![License](https://img.shields.io/github/license/prasenjit-net/openid-golang)](LICENSE)

A lightweight OpenID Connect (OIDC) identity provider implementation in Go with an embedded React admin UI.

> 💡 **Tip:** Run `./show-docs.sh` to see the complete documentation structure!

## 📚 Documentation

All documentation is organized in the **`docs/`** folder:

- **[Getting Started Guide](docs/GETTING_STARTED.md)** - Step-by-step setup tutorial ⭐ START HERE
- **[Docker Documentation](docs/DOCKER.md)** - Run with Docker and Docker Compose 🐳
- **[Quick Start](docs/QUICKSTART.md)** - Quick reference for experienced developers
- **[API Documentation](docs/API.md)** - Complete API reference
- **[Architecture](docs/ARCHITECTURE.md)** - System architecture and diagrams
- **[Storage Backends](docs/STORAGE.md)** - MongoDB and JSON storage options
- **[Testing Guide](docs/TESTING.md)** - How to test the server
- **[Implementation Details](docs/IMPLEMENTATION.md)** - Technical details
- **[Project Summary](docs/PROJECT_SUMMARY.md)** - What's been built
- **[Documentation Index](docs/INDEX.md)** - Complete documentation hub
- **[Admin UI Documentation](frontend/README.md)** - React admin interface guide
- **[CI/CD Documentation](docs/CI_CD.md)** - Continuous Integration and Deployment
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute to the project

## 🚀 Quick Start

### Option 1: Docker (Easiest)

```bash
# Using Docker Compose
docker-compose up -d

# Access the server
# http://localhost:8080
```

See **[Docker Documentation](docs/DOCKER.md)** for detailed instructions.

### Option 2: Using the Setup Wizard (Recommended for Binary)

Download the binary from [GitHub Releases](https://github.com/prasenjit-net/openid-golang/releases) and run:

```bash
# Make it executable (Linux/macOS)
chmod +x openid-server-*

# Run the setup wizard (REQUIRED - first time only)
./openid-server-linux-amd64 --setup

# Start the server
./openid-server-linux-amd64
```

The `--setup` wizard is **mandatory** and will:
- Generate RSA key pairs for JWT signing (no OpenSSL required!)
- Create configuration file (config.toml) interactively
- Choose storage backend (MongoDB or JSON file)
- Initialize storage
- Create an admin user
- Optionally create your first OAuth client

After setup, just run `./openid-server-linux-amd64` to start the server.

### Option 3: Development Setup

```bash
# 1. Run the setup script (does everything including --setup wizard)
./setup.sh

# This will:
# - Download Go dependencies
# - Build the binary
# - Run the interactive setup wizard
# - Generate keys, create config.toml, admin user, etc.

# 2. Create test data (optional)
go run scripts/seed.go

# 3. Start the server
./bin/openid-server
# OR
make run
```

**Note:** The `setup.sh` script automatically runs `--setup` for you, so everything is configured in one step. No OpenSSL dependency required!

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
- **Flexible Storage Backends:**
  - MongoDB for production (scalable, high-performance)
  - JSON file storage for development and small deployments
  - No CGO dependency - pure Go implementation
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

- **For Production (Binary)**: No prerequisites! Everything is self-contained.
- **For Development**: Go 1.21 or higher

**Note:** No OpenSSL required - RSA keys are generated using pure Go crypto!

## 🛠️ Installation

### Binary Distribution (Recommended for Production)

1. Download the latest binary for your platform from [GitHub Releases](https://github.com/prasenjit-net/openid-golang/releases)

2. Run the interactive setup wizard:
```bash
# Linux/macOS
chmod +x openid-server-*
./openid-server-* --setup

# Windows
openid-server-windows-amd64.exe --setup
```

3. Start the server:
```bash
# Linux/macOS
./openid-server-*

# Windows
openid-server-windows-amd64.exe
```

### Development Setup

Run the automated setup script:

```bash
./setup.sh
```

This will:
- Create necessary directories
- Generate RSA key pairs for JWT signing
- Create `config.toml` configuration file
- Download Go dependencies
- Build the application

## 🔧 Configuration

### Using config.toml (Recommended)

Create a `config/config.toml` file:

```toml
issuer = "http://localhost:8080"

[server]
host = "0.0.0.0"
port = 8080

[storage]
type = "json"  # or "mongodb"
json_file_path = "data/openid.json"
# For MongoDB:
# type = "mongodb"
# mongo_uri = "mongodb://localhost:27017/openid"

[jwt]
private_key_path = "config/keys/private.key"
public_key_path = "config/keys/public.key"
expiry_minutes = 60
```

### Using Environment Variables (Legacy)

You can still use environment variables:

```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
STORAGE_TYPE=json
JSON_FILE_PATH=data/openid.json
# or for MongoDB:
# STORAGE_TYPE=mongodb
# MONGO_URI=mongodb://localhost:27017/openid
JWT_PRIVATE_KEY=config/keys/private.key
JWT_PUBLIC_KEY=config/keys/public.key
JWT_EXPIRY_MINUTES=60
ISSUER=http://localhost:8080
```

### Storage Options

See [Storage Documentation](docs/STORAGE.md) for detailed information about MongoDB and JSON storage backends.

**Quick comparison:**
- **JSON Storage** (`--json-store` flag or `type = "json"`): Simple file-based storage, perfect for development and small deployments
- **MongoDB Storage** (`type = "mongodb"`): Production-grade database, recommended for high-traffic production environments

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
# Basic commands
make build         # Build the application
make run           # Run the server
make test          # Run tests
make fmt           # Format code
make lint          # Run linters
make clean         # Clean build artifacts
make deps          # Download dependencies

# Tool management
make install-tools # Install golangci-lint v2.5.0 and other dev tools
make check-tools   # Check versions of installed tools

# Or use the shell script
./lint.sh          # Run all quality checks (fmt, vet, test, lint)
```

### Code Quality Tools

This project uses **golangci-lint v2.5.0** consistently across:
- ✅ Local development (Makefile)
- ✅ CI/CD pipeline (GitHub Actions)
- ✅ Shell scripts (lint.sh)

To install or upgrade golangci-lint to the correct version:
```bash
make install-tools
```

To check your installed version:
```bash
make check-tools
golangci-lint version  # Should show v2.5.0
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
