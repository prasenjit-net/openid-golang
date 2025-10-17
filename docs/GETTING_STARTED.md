# Getting Started - Step by Step

## Prerequisites ✓

Before you begin, ensure you have:
- [ ] Go 1.21 or higher installed
- [ ] OpenSSL installed (usually pre-installed on Linux)
- [ ] A terminal/command line
- [ ] A web browser
- [ ] curl (optional, for testing)

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

## Step 3: Run Setup

The setup script will prepare everything:

```bash
./setup.sh
```

This will:
1. ✓ Create `config/keys/` directory
2. ✓ Generate RSA private key (4096-bit)
3. ✓ Generate RSA public key
4. ✓ Create `.env` file from template
5. ✓ Download Go dependencies
6. ✓ Build the application to `bin/openid-server`

**Output example:**
```
Setting up OpenID Connect Identity Server...
Creating directories...
Generating RSA key pair...
✓ RSA keys generated
✓ Created .env file from .env.example
Downloading Go dependencies...
✓ Dependencies downloaded
Building application...
✓ Application built successfully

Setup complete! 🎉
```

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

Choose one of these methods:

### Option A: Using Makefile
```bash
make run
```

### Option B: Using the built binary
```bash
./bin/openid-server
```

### Option C: Using go run
```bash
go run cmd/server/main.go
```

### Option D: Using the test script
```bash
./test.sh
```

**Expected output:**
```
Starting OpenID Connect server on 0.0.0.0:8080
Issuer: http://localhost:8080
```

✅ **Server is now running!**

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
- Change the port in `.env`: `SERVER_PORT=8081`

### Error: "no such file or directory: config/keys/private.key"
Run the setup script: `./setup.sh`

### Error: "user not found"
Run the seed script: `go run scripts/seed.go`

### Error: "invalid client credentials"
Make sure you're using the correct Client ID and Client Secret from the seed output.

### Database is locked
Stop all running instances of the server and try again.

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
   - Switch to PostgreSQL
   - Add rate limiting
   - Implement proper session management
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
- `.env` - Configuration
- `scripts/seed.go` - Test data
- `internal/handlers/*.go` - Business logic
- `internal/models/models.go` - Data models

## Success! 🎉

You now have:
- ✅ A working OpenID Connect server
- ✅ Test user and client
- ✅ All endpoints functioning
- ✅ Example client application
- ✅ Complete documentation

**Happy coding!** 🚀
