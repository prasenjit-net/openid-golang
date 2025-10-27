# Development Environment Setup

## Fixed Issues

### 1. Vite Proxy Configuration

Updated `frontend/vite.config.ts` to proxy all OpenID Connect and API endpoints to the backend server:

- `/api` - Admin API endpoints
- `/authorize` - OAuth authorization endpoint
- `/token` - Token endpoint  
- `/.well-known` - OpenID Discovery
- `/jwks` - JSON Web Key Set
- `/userinfo` - UserInfo endpoint
- `/setup` - Setup wizard

### 2. OAuth Redirect URI

Added development redirect URI to the `admin-ui` client in `data.json`:
- `http://localhost:3000/admin/callback`

This allows the OAuth flow to work in development mode.

### 3. Development Script

Created `dev.sh` script to run both backend and frontend servers concurrently:

```bash
./dev.sh
```

This script:
- Starts the backend server on `http://localhost:8080`
- Starts the frontend dev server on `http://localhost:3000`
- Handles cleanup on Ctrl+C

## Running in Development Mode

### Option 1: Use the dev script (recommended)
```bash
./dev.sh
```

### Option 2: Manual startup

Terminal 1 - Backend:
```bash
cd backend
go run main.go serve
```

Terminal 2 - Frontend:
```bash
cd frontend
npm run dev
```

## Access Points

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Admin Login**: http://localhost:3000/login (auto-redirects to OAuth)
- **Setup Wizard**: http://localhost:8080/setup (if not configured)
- **Health Check**: http://localhost:8080/health
- **OIDC Discovery**: http://localhost:8080/.well-known/openid-configuration

## OAuth Flow in Dev Mode

1. Frontend redirects to `/authorize` (proxied to backend)
2. User authenticates with backend
3. Backend redirects to `http://localhost:3000/admin/callback` with ID token
4. Frontend processes the token and logs in
5. User is redirected to `/dashboard`

## Test Credentials

From `data.json`:

**Admin User:**
- Username: `admin`
- Password: (check data.json password_hash)

**Regular User:**
- Username: `test`
- Email: `test@prasenjit.net`

## Notes

- The proxy configuration ensures all API calls from the frontend are forwarded to the backend
- Hot module replacement (HMR) works normally in the frontend
- Backend changes require restarting the Go server
- The OAuth implicit flow is used for admin UI authentication
