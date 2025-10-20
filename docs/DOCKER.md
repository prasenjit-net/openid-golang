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

### First Time Setup

Before running the server, you need to initialize it:

```bash
# Build the image
./docker-build.sh

# Run setup to initialize configuration and keys
docker run --rm -it \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/data:/app/data \
  openid-server:1.1.0 setup
```

This will:
- Generate RSA key pair for JWT signing
- Create configuration file
- Set up initial admin user

### Using Docker Compose (Recommended)

The easiest way to run the server:

```bash
# Run with JSON storage (default)
docker-compose up -d

# Run with MongoDB
docker-compose --profile with-mongodb up -d
```

The server will be available at `http://localhost:8080`

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

### With Persistent Storage

```bash
docker run -d \
  --name openid-server \
  -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config.toml:/app/config.toml:ro \
  openid-server:latest
```

### With Environment Variables

```bash
docker run -d \
  --name openid-server \
  -p 8080:8080 \
  -e SERVER_HOST=0.0.0.0 \
  -e SERVER_PORT=8080 \
  openid-server:latest serve
```

### Run Setup Command

```bash
docker run -it --rm \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/config.toml:/app/config.toml \
  openid-server:latest setup
```

## 🗂️ Volume Mounts

### Recommended Volumes

```yaml
volumes:
  - ./config:/app/config          # RSA keys directory
  - ./data:/app/data              # JSON storage directory
  - ./config.toml:/app/config.toml:ro  # Configuration file
```

### Directory Structure in Container

```
/app/
├── openid-server           # Binary
├── config/
│   ├── config.toml        # Configuration file
│   └── keys/              # RSA private/public keys
└── data/
    └── data.json          # JSON storage (if using JSON mode)
```

## 📝 Docker Compose Configurations

### JSON Storage (Default)

```bash
docker-compose up -d
```

Uses `config/config.toml` with JSON file storage.

### MongoDB Storage

```bash
docker-compose --profile with-mongodb up -d
```

Starts OpenID server + MongoDB container.

Update your `config/config.toml`:
```toml
[storage]
type = "mongodb"
mongo_uri = "mongodb://admin:changeme@mongodb:27017/openid?authSource=admin"
```

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
json_file_path = "/app/data/data.json"
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
