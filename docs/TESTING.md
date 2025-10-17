# Testing the OpenID Connect Server

## Quick Start

1. **Setup the server:**
   ```bash
   chmod +x setup.sh
   ./setup.sh
   ```

2. **Seed the database with test data:**
   ```bash
   go run scripts/seed.go
   ```
   
   This will create:
   - Test user: `testuser` / `password123`
   - Test client with client_id and client_secret (printed to console)

3. **Start the server:**
   ```bash
   make run
   # or
   go run cmd/server/main.go
   ```

4. **Test the discovery endpoint:**
   ```bash
   curl http://localhost:8080/.well-known/openid-configuration | jq
   ```

## Testing with cURL

### 1. Get Discovery Information
```bash
curl http://localhost:8080/.well-known/openid-configuration
```

### 2. Get JWKS
```bash
curl http://localhost:8080/.well-known/jwks.json
```

### 3. Authorization Flow (requires browser)

Start the test client:
```bash
go run examples/test-client.go
```

Then visit http://localhost:9090 and follow the authorization flow.

### 4. Manual Token Exchange

After getting an authorization code:
```bash
curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "CLIENT_ID:CLIENT_SECRET" \
  -d "grant_type=authorization_code&code=AUTHORIZATION_CODE&redirect_uri=http://localhost:9090/callback"
```

### 5. Get User Info

Using the access token from step 4:
```bash
curl http://localhost:8080/userinfo \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### 6. Refresh Token

Using the refresh token:
```bash
curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "CLIENT_ID:CLIENT_SECRET" \
  -d "grant_type=refresh_token&refresh_token=REFRESH_TOKEN"
```

## Testing with Postman

1. Import the OpenID Connect Discovery endpoint
2. Use the OAuth 2.0 authorization flow
3. Configure:
   - Authorization URL: http://localhost:8080/authorize
   - Token URL: http://localhost:8080/token
   - Client ID: (from seed output)
   - Client Secret: (from seed output)
   - Scope: openid profile email

## Testing PKCE Flow

Generate code verifier and challenge:
```bash
# Generate code verifier (random string)
CODE_VERIFIER=$(openssl rand -base64 32 | tr -d '=' | tr '+/' '-_')

# Generate code challenge (SHA256 hash)
CODE_CHALLENGE=$(echo -n $CODE_VERIFIER | openssl dgst -binary -sha256 | base64 | tr -d '=' | tr '+/' '-_')

echo "Code Verifier: $CODE_VERIFIER"
echo "Code Challenge: $CODE_CHALLENGE"
```

Use in authorization:
```
http://localhost:8080/authorize?client_id=CLIENT_ID&redirect_uri=http://localhost:9090/callback&response_type=code&scope=openid profile email&state=test&code_challenge=$CODE_CHALLENGE&code_challenge_method=S256
```

Then include the verifier in token exchange:
```bash
curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=CODE&redirect_uri=http://localhost:9090/callback&client_id=CLIENT_ID&code_verifier=$CODE_VERIFIER"
```

## Validating ID Tokens

The ID token is a JWT. You can decode it at https://jwt.io or using:

```bash
# Install jwt-cli: cargo install jwt-cli
jwt decode YOUR_ID_TOKEN
```

## Health Check

```bash
curl http://localhost:8080/health
```

Expected response: `{"status":"ok"}`
