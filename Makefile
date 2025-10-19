.PHONY: build run test clean deps fmt lint generate-keys help

# Default target
.DEFAULT_GOAL := help

# Build the application
build:
	@echo "Building backend..."
	@cd backend && go build -o ../bin/openid-server .

# Build the frontend UI
build-frontend:
	@echo "Building frontend..."
	@cd frontend && npm run build
	@mkdir -p backend/pkg/ui
	@rm -rf backend/pkg/ui/dist
	@cp -r frontend/dist backend/pkg/ui/dist
	@echo "Frontend built successfully"

# Build the backend server

# Run the application
run:
	@echo "Running server..."
	@cd backend && go run .

# Run tests
test:
	@echo "Running backend tests..."
	@cd backend && go test -v ./...

# Clean build artifacts and temporary files
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf backend/pkg/ui/dist
	@cd frontend && rm -rf dist node_modules/.vite
	@echo "Clean complete"

# Download dependencies
deps:
	@echo "Downloading Go dependencies..."
	@cd backend && go mod download && go mod tidy
	@echo "Installing frontend dependencies..."
	@cd frontend && npm install

# Format code
fmt:
	@echo "Formatting Go code..."
	@cd backend && go fmt ./...

# Run linter
lint:
	@echo "Running Go linter..."
	@cd backend && golangci-lint run
	@echo "Running frontend linter..."
	@cd frontend && npm run lint || true

# Generate keys for JWT signing
generate-keys:
	@echo "Generating RSA key pair..."
	@mkdir -p config/keys
	@openssl genrsa -out config/keys/private.key 4096
	@openssl rsa -in config/keys/private.key -pubout -out config/keys/public.key
	@echo "✅ Keys generated in config/keys/"

# Run setup wizard
setup:
	@./bin/openid-server setup

# Development mode - build and run
dev:
	@$(MAKE) build-all
	@./bin/openid-server serve

# Help target
help:
	@echo "OpenID Connect Server - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make build        - Build backend only"
	@echo "  make build-all    - Build frontend + backend with embedded UI"
	@echo "  make run          - Run backend in development mode"
	@echo "  make dev          - Build all and run server"
	@echo "  make test         - Run backend tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Download all dependencies (Go + NPM)"
	@echo "  make fmt          - Format Go code"
	@echo "  make lint         - Run linters (Go + frontend)"
	@echo "  make generate-keys - Generate RSA keys for JWT"
	@echo "  make setup        - Run setup wizard"
	@echo "  make help         - Show this help message"
