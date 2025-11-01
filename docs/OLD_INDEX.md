---
layout: default
title: Home
---

# OpenID Connect Identity Server Documentation

[![CI](https://github.com/prasenjit-net/openid-golang/actions/workflows/ci.yml/badge.svg)](https://github.com/prasenjit-net/openid-golang/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/prasenjit-net/openid-golang)](https://github.com/prasenjit-net/openid-golang/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/prasenjit-net/openid-golang)](../go.mod)
[![License](https://img.shields.io/github/license/prasenjit-net/openid-golang)](../LICENSE)

A lightweight, production-ready OpenID Connect (OIDC) identity provider implementation in Go with an embedded React admin UI.

---

## 🚀 Quick Navigation

### Getting Started
- **[Getting Started Guide](GETTING_STARTED.md)** ⭐ - Complete setup tutorial
- **[Quick Start](QUICKSTART.md)** - For experienced developers
- **[Docker Quick Start](DOCKER_QUICKSTART.md)** 🐳 - Run with Docker
- **[Setup Wizard](SETUP_WIZARD.md)** - Interactive configuration

### Core Documentation
- **[API Reference](API.md)** - Complete API documentation
- **[Architecture](ARCHITECTURE.md)** - System design and architecture
- **[Configuration](CONFIGURATION.md)** - Configuration options
- **[Storage Backends](STORAGE.md)** - MongoDB and JSON storage

### Development
- **[Development Setup](DEV_SETUP.md)** - Development environment
- **[Testing Guide](TESTING.md)** - Testing strategies
- **[Contributing](CONTRIBUTING.md)** - How to contribute
- **[Project Structure](STRUCTURE.md)** - Codebase organization

### Features & Implementation
- **[OIDC Compliance](OIDC_COMPLIANCE_PLAN.md)** - OpenID Connect compliance status
- **[OAuth2 Compliance](OAUTH2_COMPLIANCE_GAP_ANALYSIS.md)** - OAuth 2.0 compliance
- **[Scope-Based Claims](SCOPE_BASED_CLAIMS.md)** - Claims filtering
- **[Auth Time Verification](AUTH_TIME_VERIFICATION.md)** - Authentication time tracking
- **[Dynamic Registration](DYNAMIC_REGISTRATION_PLAN.md)** - Client registration
- **[Admin UI](ADMIN_UI.md)** - Admin interface documentation

### Advanced Features
- **[Back-Channel Logout](BACK_CHANNEL_LOGOUT_PLAN.md)** - Logout implementation
- **[Front-Channel Logout](FRONT_CHANNEL_LOGOUT_PLAN.md)** - Browser-based logout
- **[RP-Initiated Logout](RP_INITIATED_LOGOUT_PLAN.md)** - Relying party logout
- **[Audit Logging](AUDIT_LOGGING_PLAN.md)** - Security audit trails

### Operations
- **[Docker Deployment](DOCKER.md)** - Production Docker deployment
- **[CI/CD](CI_CD.md)** - Continuous integration and deployment
- **[CI/CD Implementation](CI_CD_IMPLEMENTATION.md)** - Pipeline details
- **[Linting](LINTING_RESOLUTION.md)** - Code quality standards

### Project Information
- **[Project Summary](PROJECT_SUMMARY.md)** - What's been built
- **[Implementation Details](IMPLEMENTATION.md)** - Technical details
- **[Admin UI Enhancement](ADMIN_UI_ENHANCEMENT_PLAN.md)** - Future plans
- **[Reorganization Notes](REORGANIZATION.md)** - Project restructuring

---

## 🎯 Key Features

- ✅ **Full OpenID Connect Core 1.0** - Complete OIDC implementation
- ✅ **Authorization Code Flow** - Secure authorization with PKCE support
- ✅ **Implicit Flow** - Single-page application support
- ✅ **JWT ID Tokens** - RS256 signed tokens with proper claims
- ✅ **Access & Refresh Tokens** - Token lifecycle management
- ✅ **UserInfo Endpoint** - Standard user information retrieval
- ✅ **Discovery Endpoint** - Auto-configuration support
- ✅ **JWKS Endpoint** - Public key distribution
- ✅ **Client Authentication** - Multiple auth methods
- ✅ **Flexible Storage** - MongoDB or JSON file storage
- ✅ **React Admin UI** - Modern web-based administration
- ✅ **No CGO Dependency** - Pure Go implementation

---

## 📦 Installation Options

### Docker (Recommended for Production)
```bash
docker-compose up -d
```

### Binary Release
Download from [GitHub Releases](https://github.com/prasenjit-net/openid-golang/releases)

### From Source
```bash
git clone https://github.com/prasenjit-net/openid-golang.git
cd openid-golang
./setup.sh
go run backend/main.go serve
```

---

## 🔧 Quick Configuration

### Environment Variables
```bash
STORAGE_TYPE=mongodb        # or "json"
MONGODB_URI=mongodb://localhost:27017
JWT_PRIVATE_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem
ISSUER_URL=https://auth.example.com
PORT=8080
```

### Using Setup Wizard
```bash
./openid-server setup
```

---

## 🌐 Standard Endpoints

| Endpoint | Description |
|----------|-------------|
| `/.well-known/openid-configuration` | OpenID Provider Configuration |
| `/.well-known/jwks.json` | JSON Web Key Set |
| `/authorize` | Authorization Endpoint |
| `/token` | Token Endpoint |
| `/userinfo` | UserInfo Endpoint |
| `/revoke` | Token Revocation |
| `/introspect` | Token Introspection |

---

## 🧪 Testing

```bash
# Run all tests
./test.sh

# Run specific tests
go test ./backend/pkg/handlers -v

# Run with coverage
go test -cover ./...
```

---

## 📊 Architecture Overview

```
┌─────────────┐
│   Client    │
│ Application │
└──────┬──────┘
       │ OIDC Flow
       ↓
┌─────────────────────────────────┐
│   OpenID Connect Server         │
│  ┌──────────────────────────┐  │
│  │   Authorization          │  │
│  │   Endpoint               │  │
│  └───────────┬──────────────┘  │
│              ↓                   │
│  ┌──────────────────────────┐  │
│  │   Token Endpoint         │  │
│  └───────────┬──────────────┘  │
│              ↓                   │
│  ┌──────────────────────────┐  │
│  │   UserInfo Endpoint      │  │
│  └──────────────────────────┘  │
└────────────┬────────────────────┘
             ↓
    ┌────────────────┐
    │   Storage      │
    │ MongoDB / JSON │
    └────────────────┘
```

---

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

## 🤝 Contributing

We welcome contributions! Please see the [Contributing Guide](CONTRIBUTING.md) for details.

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/prasenjit-net/openid-golang/issues)
- **Discussions**: [GitHub Discussions](https://github.com/prasenjit-net/openid-golang/discussions)
- **Documentation**: This site!

---

## 🔗 Quick Links

- [GitHub Repository](https://github.com/prasenjit-net/openid-golang)
- [Release Notes](https://github.com/prasenjit-net/openid-golang/releases)
- [CI/CD Status](https://github.com/prasenjit-net/openid-golang/actions)

---

**Last Updated**: November 2025
theme: jekyll-theme-cayman
title: OpenID Connect Server
description: A lightweight OpenID Connect (OIDC) identity provider implementation in Go
show_downloads: true
github:
  is_project_page: true

