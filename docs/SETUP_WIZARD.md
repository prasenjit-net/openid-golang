# Setup Wizard Guide

The OpenID Connect Server includes an interactive setup wizard that makes deployment simple and straightforward.

## Using the Setup Wizard

### 1. Download the Binary

Download the appropriate binary for your platform from the [GitHub Releases](https://github.com/prasenjit-net/openid-golang/releases) page:

- **Linux AMD64**: `openid-server-linux-amd64`
- **Linux ARM64**: `openid-server-linux-arm64`
- **macOS Intel**: `openid-server-darwin-amd64`
- **macOS Apple Silicon**: `openid-server-darwin-arm64`
- **Windows**: `openid-server-windows-amd64.exe`

### 2. Make it Executable (Linux/macOS only)

```bash
chmod +x openid-server-*
```

### 3. Run the Setup Wizard

```bash
# Linux/macOS
./openid-server-linux-amd64 --setup

# Windows
openid-server-windows-amd64.exe --setup
```

### 4. Follow the Interactive Prompts

The wizard will guide you through:

#### Step 1: RSA Key Generation
- Automatically generates 4096-bit RSA key pairs for JWT signing
- Saves keys to `config/keys/private.key` and `config/keys/public.key`
- If keys exist, you'll be asked if you want to regenerate them

#### Step 2: Server Configuration
- **Server Host**: IP address to bind to (default: `0.0.0.0` for all interfaces)
- **Server Port**: Port to listen on (default: `8080`)
- **Database Type**: Choose between `sqlite` (default) or `postgres`
- **Database Connection**: 
  - For SQLite: File path (default: `./openid.db`)
  - For PostgreSQL: Connection string (e.g., `postgres://user:pass@localhost/dbname`)
- **Issuer URL**: The base URL of your identity provider (default: `http://localhost:8080`)

The configuration is saved to `.env` file.

#### Step 3: Database Initialization
- Automatically initializes the database schema
- Creates all necessary tables

#### Step 4: Admin User Creation
- **Username**: Your admin username
- **Email**: Admin email address
- **Password**: Admin password (will be securely hashed with bcrypt)

#### Step 5: OAuth Client Creation (Optional)
You can create your first OAuth client during setup or skip and create it later via the admin UI.

If you choose to create one:
- **Client Name**: Friendly name for the client application
- **Redirect URI**: Where to redirect after authentication

The wizard will generate:
- **Client ID**: A random 32-character identifier
- **Client Secret**: A random 64-character secret

⚠️ **Important**: Save the client credentials securely - the secret won't be shown again!

### 5. Start the Server

```bash
# Linux/macOS
./openid-server-linux-amd64

# Windows
openid-server-windows-amd64.exe
```

### 6. Access the Admin UI

Open your browser and navigate to:
```
http://localhost:8080/
```

Login with the admin credentials you created during setup.

## Example Setup Session

```
🚀 OpenID Connect Server Setup Wizard
=====================================

Step 1: Generate RSA Keys
-------------------------
Generating 4096-bit RSA key pair...
✓ RSA keys generated at config/keys

Step 2: Server Configuration
-----------------------------
Server host (default: 0.0.0.0): 
Server port (default: 8080): 
Database type (sqlite/postgres, default: sqlite): 
Database file path (default: ./openid.db): 
Issuer URL (default: http://localhost:8080): 
✓ Configuration saved to .env

Step 3: Initialize Database
----------------------------
✓ Database initialized successfully

Step 4: Create Admin User
-------------------------
Admin username: admin
Admin email: admin@example.com
Admin password: ********
✓ Admin user 'admin' created successfully

Step 5: Create OAuth Client (Optional)
---------------------------------------
Do you want to create an OAuth client now? (y/N): y
Client name: My App
Redirect URI: http://localhost:9090/callback

✓ OAuth client created successfully!

Client ID:     abc123def456ghi789jkl012mno345pq
Client Secret: xyz789uvw456rst123opq890lmn567hij432gfe210dcb098pon765mlk654jih321

⚠️  Please save these credentials securely - the secret won't be shown again!

✅ Setup Complete!

You can now start the server with:
  ./openid-server

Access the admin UI at:
  http://localhost:8080/
```

## Command Line Options

```bash
# Run setup wizard
./openid-server --setup

# Show version
./openid-server --version

# Start server (normal mode)
./openid-server
```

## Configuration File

The setup wizard creates a `.env` file with the following format:

```bash
# OpenID Connect Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
DB_TYPE=sqlite
DB_CONNECTION=./openid.db
JWT_PRIVATE_KEY=config/keys/private.key
JWT_PUBLIC_KEY=config/keys/public.key
JWT_EXPIRY_MINUTES=60
ISSUER=http://localhost:8080
```

You can manually edit this file if needed.

## Re-running Setup

You can re-run the setup wizard at any time:

```bash
./openid-server --setup
```

The wizard will:
- Ask before overwriting existing RSA keys
- Ask before overwriting existing configuration
- Allow you to create additional users and clients

## Troubleshooting

### "Permission denied" error (Linux/macOS)
Make sure the binary is executable:
```bash
chmod +x openid-server-*
```

### "Cannot connect to database"
- For SQLite: Ensure you have write permissions in the directory
- For PostgreSQL: Verify the connection string and that the database exists

### "Port already in use"
Change the port during setup or modify the `.env` file:
```bash
SERVER_PORT=8081
```

### Keys not being recognized
Ensure the key paths in `.env` match where the keys were generated:
```bash
JWT_PRIVATE_KEY=config/keys/private.key
JWT_PUBLIC_KEY=config/keys/public.key
```

## Next Steps

After setup is complete:

1. **Test the server**: Visit `http://localhost:8080/health`
2. **Access admin UI**: Visit `http://localhost:8080/`
3. **Create more users**: Use the admin UI to add users
4. **Register clients**: Create OAuth clients for your applications
5. **Configure OIDC**: Update client applications with your issuer URL

## Production Deployment

For production deployments:

1. Use a strong admin password
2. Use PostgreSQL instead of SQLite
3. Set the issuer URL to your public domain (e.g., `https://auth.example.com`)
4. Configure HTTPS/TLS (use a reverse proxy like nginx)
5. Keep the RSA keys secure and backed up
6. Regularly rotate client secrets

## See Also

- [Getting Started Guide](GETTING_STARTED.md)
- [API Documentation](API.md)
- [Configuration Guide](../CONFIGURATION.md)
- [Deployment Guide](DEPLOYMENT.md)
