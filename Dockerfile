# Multi-stage Dockerfile for OpenID Connect Server
# Builds a minimal production image with embedded React UI

# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./

# Install all dependencies (including dev dependencies needed for build)
RUN npm ci

# Copy frontend source
COPY frontend/ ./

# Build frontend
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.24-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy backend go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy backend source
COPY backend/ ./

# Copy built frontend to embed location
COPY --from=frontend-builder /app/frontend/dist ./pkg/ui/dist/

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o openid-server .

# Stage 3: Final minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=backend-builder /app/openid-server .

# Create directories for config and data
RUN mkdir -p /app/config/keys /app/data

# Expose default port
EXPOSE 8080

# Set environment variables with defaults
ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${SERVER_PORT}/.well-known/openid-configuration || exit 1

# Run as non-root user
RUN addgroup -g 1000 openid && \
    adduser -D -u 1000 -G openid openid && \
    chown -R openid:openid /app

USER openid

# Run the server
ENTRYPOINT ["/app/openid-server"]
CMD ["serve"]
