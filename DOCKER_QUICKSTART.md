# Docker Quick Start Guide

This guide provides quick commands to get started with the OpenID Connect Server using Docker.

## Prerequisites

- Docker installed and running
- Docker Compose v2.0 or higher

## Environment Configuration (Optional)

Create a `.env` file from the example:

```bash
cp .env.example .env
```

Edit `.env` to customize MongoDB credentials:

```bash
MONGO_USER=admin
MONGO_PASSWORD=your-secure-password
MONGO_DATABASE=openid
```

## Deployment Options

### Option 1: JSON Storage Mode (Recommended for Development)

Best for: Development, testing, small deployments (<1000 users)

```bash
# Build and start
docker-compose --profile json-storage up -d

# View logs
docker-compose logs -f openid-server

# Stop
docker-compose --profile json-storage down
```

**Data Storage:**
- Configuration: `./data/config.json`
- Users/Clients/Tokens: `./data/openid.json`

### Option 2: MongoDB Storage Mode (Recommended for Production)

Best for: Production, large deployments (>1000 users), high availability

```bash
# Build and start (includes MongoDB container)
docker-compose --profile with-mongodb up -d

# View logs
docker-compose logs -f openid-server-mongodb

# Stop
docker-compose --profile with-mongodb down
```

**Data Storage:**
- Configuration: MongoDB `openid.config` collection
- Users/Clients/Tokens: MongoDB `openid` database

## First-Time Setup

After starting the server, visit the setup wizard:

```
http://localhost:8080/setup
```

Or use CLI setup:

```bash
# JSON storage mode
docker run --rm -v $(pwd)/data:/app/data \
  openid-server:latest setup \
  --issuer http://localhost:8080 \
  --admin-user admin \
  --admin-pass secret123 \
  --non-interactive

# MongoDB mode (start MongoDB first)
docker-compose --profile with-mongodb up -d mongodb
docker run --rm --network openid-golang_openid-network \
  -e MONGODB_URI=mongodb://admin:changeme@mongodb:27017/openid?authSource=admin \
  -e MONGODB_DATABASE=openid \
  openid-server:latest setup \
  --issuer http://localhost:8080 \
  --admin-user admin \
  --admin-pass secret123 \
  --non-interactive
```

## Access Points

- **Admin UI**: http://localhost:8080
- **OpenID Discovery**: http://localhost:8080/.well-known/openid-configuration
- **Setup Wizard**: http://localhost:8080/setup (first-time only)

## Useful Commands

### View Logs
```bash
# JSON mode
docker-compose logs -f openid-server

# MongoDB mode
docker-compose logs -f openid-server-mongodb
docker-compose logs -f mongodb
```

### Restart Server
```bash
# JSON mode
docker-compose restart openid-server

# MongoDB mode
docker-compose restart openid-server-mongodb
```

### Reset Configuration
```bash
# Stop containers
docker-compose down

# Remove data (WARNING: This deletes all data!)
rm -rf data/*

# Restart
docker-compose --profile json-storage up -d
```

### Access MongoDB Shell
```bash
docker exec -it openid-mongodb mongosh \
  -u admin -p changeme --authenticationDatabase admin
```

## Environment Variables

### Server Configuration
- `SERVER_HOST` - Server bind address (default: `0.0.0.0`)
- `SERVER_PORT` - Server port (default: `8080`)

### MongoDB Configuration
- `MONGO_USER` - MongoDB admin username (default: `admin`)
- `MONGO_PASSWORD` - MongoDB admin password (default: `changeme`)
- `MONGO_DATABASE` - MongoDB database name (default: `openid`)

### Setup Configuration (for non-interactive setup)
- `ISSUER_URL` - OpenID issuer URL
- `ADMIN_USER` - Initial admin username
- `ADMIN_PASS` - Initial admin password

## Troubleshooting

### Port Already in Use
```bash
# Change port in .env file
SERVER_PORT=9090

# Or override in docker-compose command
SERVER_PORT=9090 docker-compose --profile json-storage up -d
```

### MongoDB Connection Issues
```bash
# Check MongoDB is healthy
docker-compose ps mongodb

# View MongoDB logs
docker-compose logs mongodb

# Test connection
docker exec openid-mongodb mongosh --eval "db.adminCommand('ping')"
```

### Reset Everything
```bash
# Stop and remove all containers and volumes
docker-compose down -v

# Remove local data
rm -rf data/*

# Rebuild and start fresh
docker-compose build --no-cache
docker-compose --profile json-storage up -d
```

## Production Deployment

For production deployments:

1. **Use MongoDB mode** for better scalability
2. **Change default passwords** in `.env` file
3. **Use HTTPS** with a reverse proxy (nginx/traefik)
4. **Enable health checks** (already configured)
5. **Set up backups** for data volumes
6. **Use secrets management** for sensitive credentials

Example with external MongoDB:

```yaml
# docker-compose.override.yml
services:
  openid-server-mongodb:
    environment:
      - MONGODB_URI=mongodb://user:pass@mongo.example.com:27017/openid?authSource=admin&ssl=true
```

## Support

For detailed documentation, see:
- [docs/DOCKER.md](docs/DOCKER.md) - Complete Docker guide
- [README.md](README.md) - Main project documentation
