# Docker Support for OpenID Connect Server

This guide covers running the OpenID Connect Server in Docker containers.

## ⚠️ Important: VS Code Remote Connection Users

If you're using VS Code Remote connection and have just added yourself to the `docker` group, you'll need to **reload your VS Code window** to pick up the new group membership:

1. Press `F1` or `Ctrl+Shift+P`
2. Type: `Remote: Reload Window`
3. Press Enter

Alternatively, use `newgrp docker` in your terminal before running Docker commands.

You can verify Docker permissions by running: `./check-docker.sh`

## 🐳 Quick Start

### Using Docker Compose (Recommended)

The server will auto-initialize on first run:

```bash
# Build the image
docker-compose build

# Run with JSON storage (recommended for development/small deployments)
docker-compose --profile json-storage up -d

# Run with MongoDB (recommended for production)
docker-compose --profile with-mongodb up -d
```

The server will be available at `http://localhost:8080`

### First Time Setup

On first run, the server starts in **setup mode**:

1. Visit `http://localhost:8080/setup` in your browser
2. Enter your issuer URL (e.g., `http://localhost:8080`)
3. Optionally create an admin user
4. Click "Initialize" - JWT keys are auto-generated
5. Server automatically reloads in normal mode

**Alternative: CLI Setup**

You can also initialize via command line:

```bash
# Interactive setup
docker run --rm -it \
  -v $(pwd)/data:/app/data \
  openid-server:latest setup

# Non-interactive setup
docker run --rm \
  -v $(pwd)/data:/app/data \
  openid-server:latest setup \
  --issuer http://localhost:8080 \
  --admin-user admin \
  --admin-pass secret123 \
  --non-interactive
```

This creates:
- `data/config.json` - Configuration + JWT keys
- `data/openid.json` - Users, clients, tokens (JSON mode only)

### Using Docker CLI

```bash
# Run the container (after setup)
docker run -d \
  --name openid-server \
  -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/data:/app/data \
  openid-server:latest
```

## 📦 Building the Docker Image

### Automated Build

```bash
./docker-build.sh
```

This script:
- Builds a multi-stage Docker image
- Tags with version from `VERSION` file
- Shows image size
- Optionally pushes to registry

### Manual Build

```bash
docker build -t openid-server:latest .
```

### Build with Custom Registry

```bash
export DOCKER_REGISTRY=ghcr.io/prasenjit-net
./docker-build.sh
```

## 🏗️ Multi-Stage Build Architecture

The Dockerfile uses a 3-stage build process:

### Stage 1: Frontend Builder
- Base: `node:20-alpine`
- Builds React UI with Vite
- Output: `frontend/dist/`

### Stage 2: Backend Builder
- Base: `golang:1.21-alpine`
- Copies frontend build to embed location
- Builds Go binary with embedded UI
- Output: Statically linked `openid-server` binary

### Stage 3: Final Runtime
- Base: `alpine:latest` (minimal ~5MB base)
- Only contains the compiled binary
- Runs as non-root user
- Final image size: ~20-30MB

## 🚀 Running the Container

### Basic Run

```bash
docker run -p 8080:8080 openid-server:latest
```

### With Persistent Storage (JSON Mode)

```bash
docker run -d \
  --name openid-server \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  openid-server:latest
```

### With MongoDB Backend

```bash
docker run -d \
  --name openid-server \
  -p 8080:8080 \
  -e MONGODB_URI=mongodb://admin:changeme@mongodb:27017 \
  -e MONGODB_DATABASE=openid \
  -v $(pwd)/data:/app/data \
  --network openid-network \
  openid-server:latest
```

Note: Configuration (`data/config.json`) is still stored in the mounted volume even with MongoDB.

### Run Setup Command

```bash
# Interactive setup
docker run -it --rm \
  -v $(pwd)/data:/app/data \
  openid-server:latest setup

# Non-interactive with flags
docker run --rm \
  -v $(pwd)/data:/app/data \
  openid-server:latest setup \
  --issuer http://localhost:8080 \
  --admin-user admin \
  --admin-pass secret123 \
  --non-interactive

# Using environment variables
docker run --rm \
  -v $(pwd)/data:/app/data \
  -e ISSUER_URL=http://localhost:8080 \
  -e ADMIN_USER=admin \
  -e ADMIN_PASS=secret123 \
  openid-server:latest setup --non-interactive
```

## 🗂️ Storage Architecture

### Config Store System

The server uses a **config store** system instead of traditional config files:

- **Configuration**: Stored in `data/config.json` (always)
- **JWT Keys**: Generated and stored inside `data/config.json` as PEM strings
- **User Data**: Stored in `data/openid.json` (JSON mode) or MongoDB

### Auto-Detection Logic

On startup, the server checks in order:

1. **MongoDB**: Check for `MONGODB_URI` environment variable
2. **JSON File**: Check for `data/config.json` file
3. **Setup Mode**: If neither exists, start setup wizard at `/setup`

### Volume Mounts

#### JSON Storage Mode (Single Volume)

```yaml
volumes:
  - ./data:/app/data    # Configuration + JWT keys + user data
```

Contents of `./data/`:
```
data/
├── config.json    # Server config + JWT keys
└── openid.json    # Users, clients, tokens
```

#### MongoDB Storage Mode (Single Volume)

```yaml
volumes:
  - ./data:/app/data    # Configuration + JWT keys only
```

Contents of `./data/`:
```
data/
└── config.json    # Server config + JWT keys
```

User/client/token data is stored in MongoDB.

### Directory Structure in Container

```
/app/
├── openid-server    # Binary (with embedded React UI)
└── data/            # Persistent data directory
    ├── config.json  # Config store (always)
    └── openid.json  # JSON storage (JSON mode only)
```

## 📝 Docker Compose Configurations

### JSON Storage Mode

```bash
docker-compose --profile json-storage up -d
```

Best for:
- Development environments
- Small deployments (<1000 users)
- Simplified setup

### MongoDB Storage Mode

```bash
docker-compose --profile with-mongodb up -d
```

Best for:
- Production environments
- Large deployments (>1000 users)
- High availability setups

The `MONGODB_URI` environment variable tells the server to use MongoDB for storage.

### Custom Configuration

Create `docker-compose.override.yml`:

```yaml
version: '3.8'
services:
  openid-server:
    environment:
      - SERVER_PORT=9090
    ports:
      - "9090:9090"
```

## 🔧 Configuration

### Environment Variables

- `SERVER_HOST` - Server bind address (default: `0.0.0.0`)
- `SERVER_PORT` - Server port (default: `8080`)

### Configuration File

Mount your `config.toml` file:

```toml
issuer = "http://localhost:8080"

[server]
host = "0.0.0.0"
port = 8080

[storage]
type = "json"
json_file_path = "/app/data/openid.json"
```

## 🏥 Health Checks

The Docker image includes a health check:

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --spider http://localhost:8080/.well-known/openid-configuration
```

Check health status:
```bash
docker ps
docker inspect --format='{{.State.Health.Status}}' openid-server
```

## 🔐 Security

### Non-Root User

The container runs as user `openid` (UID 1000):

```dockerfile
USER openid
```

### File Permissions

Ensure mounted volumes are readable by UID 1000:

```bash
chown -R 1000:1000 config/ data/
```

## 📊 Monitoring

### View Logs

```bash
# Docker Compose
docker-compose logs -f openid-server

# Docker CLI
docker logs -f openid-server
```

### Resource Usage

```bash
docker stats openid-server
```

## 🎯 Production Deployment

### Using Docker Swarm

```bash
docker stack deploy -c docker-compose.yml openid
```

### Using Kubernetes

Example deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openid-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: openid-server
  template:
    metadata:
      labels:
        app: openid-server
    spec:
      containers:
      - name: openid-server
        image: openid-server:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /app/config
        - name: data
          mountPath: /app/data
      volumes:
      - name: config
        persistentVolumeClaim:
          claimName: openid-config
      - name: data
        persistentVolumeClaim:
          claimName: openid-data
```

## 🛠️ Troubleshooting

### Container Won't Start

Check logs:
```bash
docker logs openid-server
```

### Permission Denied

Ensure volumes are writable:
```bash
chmod 755 config/ data/
chown -R 1000:1000 config/ data/
```

### Can't Connect to Server

Check if port is exposed:
```bash
docker port openid-server
```

### MongoDB Connection Issues

Verify MongoDB is running:
```bash
docker-compose ps
docker-compose logs mongodb
```

## 🧹 Cleanup

### Stop and Remove Containers

```bash
docker-compose down
```

### Remove Volumes Too

```bash
docker-compose down -v
```

### Clean Up Images

```bash
docker rmi openid-server:latest
docker system prune -a
```

## 📦 Image Registry

### Push to Docker Hub

```bash
docker tag openid-server:latest username/openid-server:latest
docker push username/openid-server:latest
```

### Push to GitHub Container Registry

```bash
docker tag openid-server:latest ghcr.io/prasenjit-net/openid-server:latest
docker push ghcr.io/prasenjit-net/openid-server:latest
```

## 🔄 Updates

### Update to Latest Version

```bash
docker-compose pull
docker-compose up -d
```

### Rollback

```bash
docker tag openid-server:v1.1.0 openid-server:latest
docker-compose up -d
```

## 📚 Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)
- [Best Practices for Writing Dockerfiles](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
