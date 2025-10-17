# Configuration and Setup - Complete Guide

## Overview

The OpenID Connect Server requires the `--setup` wizard to be run before first use. This creates all necessary configuration, keys, and initial data.

**Key Points:**
- ✅ `--setup` is **mandatory** - server won't start without it
- ✅ No auto-creation of config files - explicit setup required
- ✅ No external dependencies (OpenSSL, etc.) - pure Go implementation
- ✅ `setup.sh` automates everything for development

## Setup Methods

There are two ways to set up the server:

1. **Interactive Setup** (`--setup` flag) - For production/manual setup
2. **Development Setup** (`setup.sh` script) - For developers (calls `--setup` automatically)

## Development Setup (./setup.sh)

### Purpose
Automates the entire development environment setup in one command.

### What it does:
- ✅ Creates directories (`config/keys/`, `bin/`)
- ✅ Downloads Go dependencies
- ✅ Builds the application binary
- ✅ **Calls `./bin/openid-server --setup` automatically**
  - Generates RSA key pairs (pure Go crypto, no OpenSSL!)
  - Creates `config.toml` interactively
  - Sets up storage
  - Creates admin user
  - Creates OAuth clients (optional)

### Usage:
```bash
./setup.sh
```

### Interactive Experience:
The script will prompt you for all configuration options through the embedded `--setup` wizard.

### When to use:
- First-time development environment setup
- After cloning the repository
- When you want everything configured in one go

**Note:** This is the recommended method for developers as it handles everything automatically.

## Runtime Setup (--setup flag)

### Purpose
Interactive wizard for manual setup and production deployment.

### What it does:
- ✅ Generates RSA key pairs (4096-bit, pure Go implementation)
- ✅ Creates `config.toml` with your preferences
- ✅ Lets you choose storage backend (MongoDB or JSON)
- ✅ Initializes the selected storage
- ✅ Creates admin user
- ✅ Optionally creates OAuth clients

### Usage:
```bash
./openid-server --setup
```

### When to use:
- Production deployment (when not using `setup.sh`)
- Manual configuration
- Reconfiguring an existing installation
- When you downloaded just the binary

**Note:** This is **mandatory** before the server can run. The server will exit with an error if `config.toml` doesn't exist.

## Server Startup Behavior

### With config.toml present:
```bash
./openid-server
# Starts normally using config.toml
```

### Without config.toml:
```bash
./openid-server
# ❌ Configuration file not found!
# 
# Please run the setup wizard first:
#   ./openid-server --setup
# 
# This will:
#   - Generate RSA keys for JWT signing
#   - Create config.toml with your preferences
#   - Choose storage backend (MongoDB or JSON)
#   - Create admin user and OAuth clients
```

**The server will NOT auto-create any files.** Setup is mandatory and explicit.

## Configuration Priority

The server loads configuration in this order (first found wins):

1. **config.toml** file (if exists)
2. **Environment variables** (legacy support)
3. **Built-in defaults**

### Command-line overrides:
- `--json-store` - Force JSON storage regardless of config

## Complete Workflow Examples

### Example 1: Development Setup (Recommended)

```bash
# One command does everything!
./setup.sh

# Prompts you for configuration, then:
# - Builds binary
# - Generates keys  
# - Creates config.toml
# - Sets up admin user
# - Ready to go!

# Start server
./bin/openid-server
```

### Example 2: Production Deployment

```bash
# 1. Download binary
wget https://github.com/prasenjit-net/openid-golang/releases/download/v1.0.0/openid-server-linux-amd64
chmod +x openid-server-linux-amd64

# 2. Run setup (REQUIRED)
./openid-server-linux-amd64 --setup
# Follow interactive prompts

# 3. Start server
./openid-server-linux-amd64
```

### Example 3: Trying to skip setup (FAILS)

```bash
# Download and try to run immediately
./openid-server

# ❌ Configuration file not found!
# Please run the setup wizard first:
#   ./openid-server --setup
```

**Setup is mandatory - no shortcuts!**

## Configuration Files

### config.toml (Runtime Configuration)

**Created by:**
- `--setup` interactive wizard
- Auto-created on first server start
- Manually created/edited

**Location:** Project root directory

**Example:**
```toml
issuer = "https://auth.example.com"

[server]
host = "0.0.0.0"
port = 8443

[storage]
type = "mongodb"
mongo_uri = "mongodb://user:pass@localhost:27017/openid"

[jwt]
private_key_path = "/etc/openid/keys/private.key"
public_key_path = "/etc/openid/keys/public.key"
expiry_minutes = 30
```

### Environment Variables (Legacy)

Still supported for backward compatibility:

```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
STORAGE_TYPE=json
JSON_FILE_PATH=data.json
MONGO_URI=mongodb://localhost:27017/openid
JWT_PRIVATE_KEY=config/keys/private.key
JWT_PUBLIC_KEY=config/keys/public.key
JWT_EXPIRY_MINUTES=60
ISSUER=http://localhost:8080
```

## Storage Options

### JSON File Storage (Default)

**Pros:**
- No external dependencies
- Simple setup
- Easy backup (single file)
- Perfect for development

**Cons:**
- Not suitable for high concurrency
- Limited scalability

**Configuration:**
```toml
[storage]
type = "json"
json_file_path = "data.json"
```

### MongoDB Storage

**Pros:**
- Production-grade
- High performance
- Scalable
- Supports clustering

**Cons:**
- Requires MongoDB server
- More complex setup

**Configuration:**
```toml
[storage]
type = "mongodb"
mongo_uri = "mongodb://localhost:27017/openid"
```

See [STORAGE.md](STORAGE.md) for detailed information.

## Troubleshooting

### Config file not found
**Symptom:** Server starts but logs warning about config
**Solution:** Normal! Server creates default config.toml automatically

### Want to reconfigure
**Solution:** Run `./openid-server --setup` again

### Storage connection fails
**For MongoDB:**
- Ensure MongoDB is running: `sudo systemctl status mongod`
- Check connection URI
- Or switch to JSON: `./openid-server --json-store`

**For JSON:**
- Check file permissions
- Ensure disk space available

### RSA keys not found
**Solution:** 
- Run `./setup.sh` (development)
- Or run `./openid-server --setup` (production)

### Port already in use
**Solution:** Edit `config.toml` and change port number

## Summary

| Task | Command | When to Use |
|------|---------|-------------|
| Development setup | `./setup.sh` | First time dev setup (does everything) |
| Production setup | `./openid-server --setup` | Deploying to production |
| Reconfigure | `./openid-server --setup` | Change settings/users/clients |
| Start server | `./openid-server` | After setup is complete |

**Key Points:**
- ✅ `--setup` is **mandatory** before first run
- ✅ No auto-creation of config or keys
- ✅ `setup.sh` calls `--setup` for you (dev only)
- ✅ Pure Go implementation - no OpenSSL needed
- ✅ Explicit configuration - no surprises
