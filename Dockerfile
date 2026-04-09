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

# Build frontend - outputs directly to /app/pkg/ui/uidist (per vite.config.ts outDir)
RUN mkdir -p /app/pkg/ui/uidist && npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go module files and download dependencies first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy Go source files
COPY main.go ./
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/

# Overlay the built frontend assets from stage 1
COPY --from=frontend-builder /app/pkg/ui/uidist ./pkg/ui/uidist/

# Build the Go binary with embedded UI
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o openid-server .

# Stage 3: Final minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/openid-server .

RUN mkdir -p /app/data

EXPOSE 8080

ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${SERVER_PORT}/.well-known/openid-configuration || exit 1

RUN addgroup -g 1000 openid && \
    adduser -D -u 1000 -G openid openid && \
    chown -R openid:openid /app

USER openid

ENTRYPOINT ["/app/openid-server"]
CMD ["serve"]
