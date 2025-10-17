# Getting Started - Step by Step

## Prerequisites ✓

Before you begin, ensure you have:
- [ ] Go 1.21 or higher installed (for development setup)
- [ ] Optional: MongoDB if you want to use MongoDB storage
- [ ] A terminal/command line
- [ ] A web browser

**Note:** 
- For production deployment, download pre-built binaries - no dependencies required!
- No OpenSSL needed - RSA keys generated using pure Go crypto!

## Step 1: Install Go

If Go is not installed:

```bash
# Check if Go is installed
go version

# If not installed, install it:
sudo apt install golang-go

# Or via snap:
sudo snap install go --classic

# Verify installation
go version
```

Expected output: `go version go1.21.x linux/amd64` (or similar)

## Step 2: Navigate to Project

```bash
cd /home/prasenjit/CodeProjects/openid-golang
```

## Step 3: Run Development Setup

The setup script will prepare **everything** for development:

```bash
./setup.sh
```

This comprehensive script will:
1. ✓ Create directories
2. ✓ Download Go dependencies  
3. ✓ Build the application
4. ✓ **Run the interactive setup wizard** which:
   - Generates RSA key pairs (4096-bit) using pure Go crypto
   - Creates `config.toml` with your configuration
   - Sets up storage (JSON or MongoDB)
   - Creates admin user
   - Creates OAuth clients (optional)

**You'll be prompted for:**
- Server configuration (host, port)
- Storage type (JSON file or MongoDB)
- Admin user credentials
- OAuth client details (optional)

**Output at completion:**
```
==========================================
Development Environment Setup Complete! 🎉
==========================================

Configuration file: config.toml
RSA keys: config/keys/
```

**Note:** Everything is done in one step - no need to run `--setup` separately!

## Step 4: Seed the Database

Create test user and client:

```bash
go run scripts/seed.go
```

**IMPORTANT:** This will output:
```
Seeding database with test data...
✓ Created test user: testuser (password: password123)
✓ Created test client:
  Client ID: abc-123-def-456
  Client Secret: xyz-789-uvw-012
  Redirect URIs: [http://localhost:3000/callback]

Seeding complete!
```

**📝 Write down the Client ID and Client Secret!**

You can also create custom clients by modifying `scripts/seed.go`.

## Step 5: Start the Server

Now that setup is complete, starting the server is simple:

```bash
./bin/openid-server
# OR
make run
```

**Expected output:**
```
Starting OpenID Connect Server vdev
Using JSON file storage: data.json
Starting OpenID Connect server on 0.0.0.0:8080
Issuer: http://localhost:8080
```

✅ **Server is now running!**

**If you see an error about missing config.toml:**
This means setup wasn't completed. Run `./bin/openid-server --setup` first.

## Step 6: Verify It's Working

Open a **new terminal** (keep the server running) and test:

### Test 1: Health Check
```bash
curl http://localhost:8080/health
```
Expected: `{"status":"ok"}`

### Test 2: Discovery Endpoint
```bash
curl http://localhost:8080/.well-known/openid-configuration | jq
```
Expected: JSON with OpenID Connect configuration

### Test 3: JWKS Endpoint
```bash
curl http://localhost:8080/.well-known/jwks.json | jq
```
Expected: JSON Web Key Set

## Step 7: Test the Full OAuth Flow

### Method A: Use the Test Client (Recommended)

In a **new terminal**:

```bash
go run examples/test-client.go
```

This starts a test client on port 9090.

**Expected output:**
```
OpenID Connect Test Client
===========================

Make sure the OpenID server is running at http://localhost:8080
Run the seed script first: go run scripts/seed.go

Client ID: your-client-id
Client Secret: your-client-secret

Starting test client on http://localhost:9090
Visit http://localhost:9090 to start the authorization flow
```

**Now:**
1. Open your browser
2. Go to http://localhost:9090
3. Click "Click here to authorize"
4. Login with:
   - Username: `testuser`
   - Password: `password123`
5. You'll be redirected back with an authorization code
6. Copy the curl command and run it to get tokens

### Method B: Manual Testing

#### 1. Get Authorization Code

Open in browser:
```
http://localhost:8080/authorize?client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost:3000/callback&response_type=code&scope=openid%20profile%20email&state=random123
```

Replace `YOUR_CLIENT_ID` with the client ID from Step 4.

Login with:
- Username: `testuser`
- Password: `password123`

You'll be redirected to:
```
http://localhost:3000/callback?code=AUTHORIZATION_CODE&state=random123
```

Copy the `AUTHORIZATION_CODE` from the URL.

#### 2. Exchange Code for Tokens

```bash
curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "YOUR_CLIENT_ID:YOUR_CLIENT_SECRET" \
  -d "grant_type=authorization_code" \
  -d "code=AUTHORIZATION_CODE" \
  -d "redirect_uri=http://localhost:3000/callback"
```

Response:
```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "...",
  "id_token": "eyJhbGc..."
}
```

#### 3. Get User Info

```bash
curl http://localhost:8080/userinfo \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

Response:
```json
{
  "sub": "user-id",
  "name": "Test User",
  "given_name": "Test",
  "family_name": "User",
  "email": "test@example.com"
}
```

#### 4. Decode ID Token

Go to https://jwt.io and paste the `id_token` to decode it.

You'll see:
- Header: Algorithm (RS256), Key ID
- Payload: User claims (sub, name, email, etc.)
- Signature: Verification status

## Step 8: Stop the Server

Press `Ctrl+C` in the terminal where the server is running.

```
^C
Shutting down server...
Server stopped
```

## Troubleshooting

### Error: "bind: address already in use"
Port 8080 is already in use. Either:
- Stop the other application using port 8080
- Change the port in `config.toml`: `port = 8081` under `[server]` section

### Error: "no such file or directory: config/keys/private.key"
Run the setup script: `./setup.sh`

### Error: "user not found"
Run the seed script: `go run scripts/seed.go`

### Error: "invalid client credentials"
Make sure you're using the correct Client ID and Client Secret from the seed output.

### Error: "failed to connect to MongoDB"
If using MongoDB storage:
- Ensure MongoDB is running: `sudo systemctl status mongod`
- Check the connection URI in `config.toml`
- Or switch to JSON storage: edit `config.toml` and set `type = "json"`

### Can't find config.toml
The server will work with environment variables as a fallback. Copy `config.toml.example` to `config.toml` and customize it.

## What's Next?

Now that you have a working OpenID Connect server:

1. **Read the Documentation**
   - `README.md` - Overview
   - `docs/API.md` - API reference
   - `docs/ARCHITECTURE.md` - System design
   - `docs/TESTING.md` - More testing scenarios

2. **Customize It**
   - Modify `internal/handlers/authorize.go` to customize login
   - Add new user fields in `internal/models/models.go`
   - Implement your own storage in `internal/storage/`

3. **Integrate with Your App**
   - Use the OIDC endpoints in your application
   - Implement OAuth2 client flow
   - Validate ID tokens

4. **Production Deployment**
   - Enable HTTPS
   - Use MongoDB for production storage
   - Add rate limiting
   - Implement proper session management
   - See `docs/STORAGE.md` for storage backend options
   - See `QUICKSTART.md` for production checklist

## Quick Reference

### Test Credentials
- Username: `testuser`
- Password: `password123`

### Server Endpoints
- Base URL: `http://localhost:8080`
- Discovery: `/.well-known/openid-configuration`
- JWKS: `/.well-known/jwks.json`
- Authorize: `/authorize`
- Token: `/token`
- UserInfo: `/userinfo`
- Health: `/health`

### Commands
```bash
./setup.sh                    # Initial setup
go run scripts/seed.go        # Create test data
make run                      # Start server
go run examples/test-client.go # Test client
make test                     # Run tests
make build                    # Build binary
```

### Files to Edit
- `config.toml` - Configuration (server, storage, JWT settings)
- `scripts/seed.go` - Test data
- `internal/handlers/*.go` - Business logic
- `internal/models/models.go` - Data models

### Storage Options
- **JSON File** (default): Simple file-based storage in `data.json`
- **MongoDB**: Production-grade database storage
- See `docs/STORAGE.md` for details

## Success! 🎉

You now have:
- ✅ A working OpenID Connect server
- ✅ Test user and client
- ✅ All endpoints functioning
- ✅ Example client application
- ✅ Complete documentation

**Happy coding!** 🚀
