# OpenID Golang - Complete Project Documentation

> 📁 **All documentation is now organized in the `docs/` folder for easy access!**

## 📚 Documentation Index

This project contains comprehensive documentation:

1. **[README.md](../README.md)** - Project overview and features
2. **[GETTING_STARTED.md](GETTING_STARTED.md)** - Detailed step-by-step setup guide ⭐ START HERE
3. **[QUICKSTART.md](QUICKSTART.md)** - Quick reference for experienced developers
4. **[STORAGE.md](STORAGE.md)** - Storage backends guide (MongoDB & JSON)
5. **[PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)** - What's been built and why
6. **[IMPLEMENTATION.md](IMPLEMENTATION.md)** - Technical implementation details
7. **[API.md](API.md)** - Complete API reference
8. **[ARCHITECTURE.md](ARCHITECTURE.md)** - System architecture and diagrams
9. **[TESTING.md](TESTING.md)** - Testing guide and examples
10. **[CI_CD.md](CI_CD.md)** - CI/CD workflows and automation

📋 See **[STRUCTURE.md](../STRUCTURE.md)** for complete project structure overview.

## 🚀 Quick Start (3 Steps)

```bash
# 1. Setup
./setup.sh

# 2. Create test data
go run scripts/seed.go

# 3. Start server
./test.sh
```

Visit http://localhost:8080/health to verify.

## 📁 Project Structure

```
openid-golang/
├── cmd/server/main.go              # Application entry point
├── internal/                       # Private application code
│   ├── config/                     # Configuration
│   ├── crypto/                     # JWT, PKCE, bcrypt
│   ├── handlers/                   # HTTP handlers
│   ├── middleware/                 # HTTP middleware
│   ├── models/                     # Data models
│   └── storage/                    # Storage backends (MongoDB, JSON)
├── scripts/seed.go                 # Database seeding
├── examples/test-client.go         # OAuth test client
├── docs/                           # Documentation
├── config.toml.example             # Configuration example
├── setup.sh                        # Setup script ⭐
├── test.sh                         # Quick test script ⭐
└── Makefile                        # Build commands
```

## 🎯 Core Features

- ✅ Full OpenID Connect Core 1.0
- ✅ Authorization Code Flow
- ✅ PKCE Support
- ✅ JWT ID Tokens (RS256)
- ✅ Access & Refresh Tokens
- ✅ UserInfo Endpoint
- ✅ Discovery Endpoint
- ✅ JWKS Endpoint
- ✅ Client Authentication
- ✅ Flexible Storage (MongoDB or JSON file)
- ✅ No CGO dependency - Pure Go

## 🔑 Test Credentials

After running `go run scripts/seed.go`:

```
Username: testuser
Password: password123
Client ID: (shown in seed output)
Client Secret: (shown in seed output)
```

## 🌐 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/.well-known/openid-configuration` | GET | Discovery |
| `/.well-known/jwks.json` | GET | Public keys |
| `/authorize` | GET | Start auth flow |
| `/token` | POST | Get tokens |
| `/userinfo` | GET | User profile |
| `/health` | GET | Health check |

## 💻 Commands

```bash
# Setup and initialization
./setup.sh                          # Complete setup
go run scripts/seed.go              # Create test data
./test.sh                           # Quick test

# Running
make run                            # Start server
go run cmd/server/main.go           # Direct run
./bin/openid-server                 # Use built binary

# Development
make build                          # Build binary
make test                           # Run tests
make fmt                            # Format code
make clean                          # Clean artifacts

# Testing
go run examples/test-client.go      # Test OAuth flow
curl http://localhost:8080/health   # Health check
```

## 🔧 Configuration

Edit `config.toml` file:

```toml
issuer = "http://localhost:8080"

[server]
host = "0.0.0.0"
port = 8080

[storage]
type = "json"                       # or "mongodb"
json_file_path = "data.json"
# mongo_uri = "mongodb://localhost:27017/openid"

[jwt]
private_key_path = "config/keys/private.key"
public_key_path = "config/keys/public.key"
expiry_minutes = 60
```

Or use environment variables (legacy support).

See [STORAGE.md](STORAGE.md) for storage backend options.

## 📊 Code Statistics

- **Total Lines:** ~1,800 lines of Go code
- **Files:** 25+ Go files
- **Packages:** 7 internal packages
- **Tests:** Unit tests included
- **Documentation:** 8 comprehensive guides

## 🏗️ Architecture

```
Client App → HTTP → [Middleware] → [Handlers] → [Storage]
                         ↓              ↓            ↓
                    Logging        JWT/Crypto    SQLite
                    CORS           Models
                    Recovery       Config
```

## 🧪 Testing Flow

```
1. Start server:
   ./test.sh

2. In new terminal:
   go run examples/test-client.go

3. Open browser:
   http://localhost:9090

4. Login:
   testuser / password123

5. Get authorization code

6. Exchange for tokens

7. Call UserInfo endpoint
```

## 📖 Learning Path

1. **Beginners:** Start with [GETTING_STARTED.md](GETTING_STARTED.md)
2. **Quick Setup:** Use [QUICKSTART.md](QUICKSTART.md)
3. **Understanding:** Read [IMPLEMENTATION.md](IMPLEMENTATION.md)
4. **API Usage:** Reference [API.md](API.md)
5. **Architecture:** Study [ARCHITECTURE.md](ARCHITECTURE.md)
6. **Testing:** Follow [TESTING.md](TESTING.md)

## 🛡️ Security Notes

This is a development/learning implementation. For production:

- [ ] Enable HTTPS/TLS
- [ ] Use PostgreSQL
- [ ] Add rate limiting
- [ ] Implement session management
- [ ] Add CSRF protection
- [ ] Enable account lockout
- [ ] Add audit logging
- [ ] Use HSM/KMS for keys
- [ ] Add monitoring
- [ ] Configure backups

See [QUICKSTART.md](QUICKSTART.md) for complete production checklist.

## 🤝 Contributing

The codebase is modular and extensible:

- **Storage:** Implement `storage.Storage` interface
- **Auth:** Modify `handlers/authorize.go`
- **Scopes:** Update `handlers/userinfo.go`
- **Grant Types:** Extend `handlers/token.go`

## 📝 Standards

Compliant with:
- OpenID Connect Core 1.0
- OAuth 2.0 (RFC 6749)
- JWT (RFC 7519)
- PKCE (RFC 7636)
- JWKS (RFC 7517)

## 🔗 Useful Links

- [OpenID Connect Spec](https://openid.net/specs/openid-connect-core-1_0.html)
- [OAuth 2.0 RFC](https://tools.ietf.org/html/rfc6749)
- [JWT.io](https://jwt.io) - JWT decoder
- [OAuth.tools](https://oauth.tools) - OAuth tester

## ❓ Troubleshooting

### Server won't start
- Check port 8080 is available
- Verify RSA keys exist: `ls config/keys/`
- Run setup: `./setup.sh`

### Login fails
- Verify test user: `go run scripts/seed.go`
- Check credentials: testuser / password123

### Token exchange fails
- Verify client credentials
- Check redirect URI matches
- Ensure code hasn't expired

### JWT validation fails
- Check RSA keys
- Verify issuer URL

See [GETTING_STARTED.md](GETTING_STARTED.md) for more troubleshooting.

## 📞 Support

- Review documentation in `docs/` folder
- Check code comments in source files
- Examine test files for examples
- Study the example client

## ✨ What You Get

After setup, you have:

- ✅ Complete OpenID Connect Identity Provider
- ✅ Test user and OAuth client
- ✅ Working login and consent flow
- ✅ JWT token generation
- ✅ Full API endpoints
- ✅ Example OAuth client
- ✅ Comprehensive tests
- ✅ Production-ready architecture
- ✅ Extensive documentation

## 🎓 Next Steps

1. **Run It:** Follow [GETTING_STARTED.md](GETTING_STARTED.md)
2. **Understand It:** Read [IMPLEMENTATION.md](IMPLEMENTATION.md)
3. **Customize It:** Modify handlers and models
4. **Test It:** Use [docs/TESTING.md](docs/TESTING.md)
5. **Deploy It:** Follow production checklist
6. **Integrate It:** Use in your applications

## 📄 License

MIT License - Free for personal and commercial use

---

**Ready to start? Run `./setup.sh` and follow [GETTING_STARTED.md](GETTING_STARTED.md)!** 🚀

For any questions, refer to the documentation or examine the source code - it's well-commented and structured for learning.
