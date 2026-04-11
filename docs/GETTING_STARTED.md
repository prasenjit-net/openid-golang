# Getting Started

## Prerequisites

| Mode | Requirements |
|---|---|
| **Binary / Docker** | None — fully self-contained |
| **Development** | Go 1.24+, Node 20+ |

---

## Option 1 — Docker (fastest)

```bash
git clone https://github.com/prasenjit-net/openid-golang.git
cd openid-golang

# Start with JSON-file storage (no external dependencies)
docker-compose --profile json-storage up -d
```

The server is available at **http://localhost:8080**.

On first launch the setup wizard opens automatically at `http://localhost:8080/setup`.

---

## Option 2 — Pre-built Binary

1. Download the binary for your platform from [GitHub Releases](https://github.com/prasenjit-net/openid-golang/releases).

2. Make it executable and run the setup wizard:
```bash
chmod +x openid-server-*
./openid-server-* --setup
```

3. Start the server:
```bash
./openid-server-*
```

---

## Option 3 — Build from Source

```bash
git clone https://github.com/prasenjit-net/openid-golang.git
cd openid-golang

# Install Go + npm dependencies
make deps

# Build frontend + Go binary and run
make run
```

---

## First-Run Setup Wizard

Open **http://localhost:8080/setup** in your browser.

The wizard guides you through:

1. **Issuer URL** — the base URL clients will use (e.g. `http://localhost:8080`)
2. **Storage backend** — JSON file (default) or MongoDB
3. **Admin user** — username + password for the admin portal
4. **Initial OAuth client** (optional) — create a test client immediately

After setup, the wizard is disabled and the admin UI is available at **http://localhost:8080**.

---

## Admin UI Tour

| Page | Path | Description |
|---|---|---|
| Dashboard | `/` | Stats, recent activity |
| Users | `/users` | Create and manage OIDC users |
| Clients | `/clients` | Register and manage OAuth clients |
| Keys | `/keys` | Rotate signing keys, generate CSR, import cert |
| Tokens | `/tokens` | Search and revoke active tokens |
| Audit Log | `/audit` | Security event history |
| Settings | `/settings` | Issuer URL, token lifetimes, registration settings |

---

## Your First Authorization Code Flow

### 1. Register a client

In the admin UI → **Clients** → **New Client**, or via the API:

```bash
curl -X POST http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{
    "client_name": "My App",
    "redirect_uris": ["http://localhost:3000/callback"],
    "grant_types": ["authorization_code", "refresh_token"],
    "response_types": ["code"]
  }'
```

### 2. Create a user

Admin UI → **Users** → **New User**, or:

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"your-password"}' | jq -r .token)

curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"alice123","email":"alice@example.com"}'
```

### 3. Start the authorization flow

Open this URL in a browser (replace values):

```
http://localhost:8080/authorize
  ?client_id=YOUR_CLIENT_ID
  &redirect_uri=http://localhost:3000/callback
  &response_type=code
  &scope=openid%20profile%20email
  &state=random-state
  &code_challenge=YOUR_S256_CHALLENGE
  &code_challenge_method=S256
```

Log in as `alice`, grant consent, and you will be redirected with `?code=...`.

### 4. Exchange code for tokens

```bash
curl -X POST http://localhost:8080/token \
  -u "YOUR_CLIENT_ID:YOUR_CLIENT_SECRET" \
  -d "grant_type=authorization_code" \
  -d "code=THE_CODE" \
  -d "redirect_uri=http://localhost:3000/callback" \
  -d "code_verifier=YOUR_PKCE_VERIFIER"
```

---

## Interactive Test Client

A full interactive demo is bundled:

```bash
go run examples/test-client.go
# Open http://localhost:9090
```

The test client walks through dynamic registration, Authorization Code Flow with PKCE, Implicit Flow, token refresh, UserInfo, and introspection — with a UI that explains each step.

---

## Key Management

### Rotate signing key

```bash
curl -X POST http://localhost:8080/api/settings/rotate-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"validity_days": 365}'
```

### Get a CSR (for CA submission)

```bash
# 1. Find the key ID
KEY_ID=$(curl -s http://localhost:8080/api/keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')

# 2. Generate CSR
curl -s http://localhost:8080/api/keys/$KEY_ID/csr \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r .csr_pem
```

### Import CA-signed certificate

```bash
curl -X POST http://localhost:8080/api/keys/$KEY_ID/import-cert \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"cert_pem\": \"$(cat signed.crt | awk 'NF{printf "%s\\n",$0}')\"}"
```
