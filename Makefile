.PHONY: build build-all build-frontend run test clean deps fmt lint generate-keys help install-tools check-tools dev setup

# Configuration
GOLANGCI_LINT_VERSION := v2.5.0

# Default target
.DEFAULT_GOAL := help

# Build frontend then backend binary (for producing a release binary)
build-all: build-frontend build

# Build the frontend UI directly into the embed directory
build-frontend:
	@echo "Building frontend..."
	@mkdir -p ui/dist
	@cd frontend && npm run build
	@echo "Frontend built successfully"

# Build the binary (requires frontend to be built first via build-all)
build:
	@echo "Building..."
	@go build -o bin/openid-server .

# Run the application (builds frontend first to ensure embedded UI is up-to-date)
run: build-frontend
	@echo "Running server..."
	@go run .

# Run tests (frontend must be built first for the go:embed to compile)
test: build-frontend
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts and temporary files
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf ui/dist
	@cd frontend && rm -rf node_modules/.vite
	@echo "Clean complete"

# Download dependencies
deps:
	@echo "Downloading Go dependencies..."
	@go mod download && go mod tidy
	@echo "Installing frontend dependencies..."
	@cd frontend && npm install

# Format code
fmt:
	@echo "Formatting Go code..."
	@go fmt ./...

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	@echo "✅ golangci-lint $(GOLANGCI_LINT_VERSION) installed successfully"

# Check installed tools
check-tools:
	@echo "Checking development tools..."
	@command -v golangci-lint >/dev/null 2>&1 && echo "✅ golangci-lint $$(golangci-lint version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)" || echo "❌ golangci-lint not found. Run 'make install-tools'"
	@echo "✅ go $$(go version | grep -oE 'go[0-9]+\.[0-9]+(\.[0-9]+)?' | sed 's/go//')"

# Run linter
lint:
	@echo "Running Go linter..."
	@golangci-lint run --timeout=5m
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
setup: build-frontend
	@echo "Running setup wizard..."
	@go run . setup

# Development mode - build frontend and run via go run (no binary produced)
dev: build-frontend
	@echo "Starting development server..."
	@go run . serve

# Help target
help:
	@echo "OpenID Connect Server - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make dev           - Build frontend + run server via go run (development)"
	@echo "  make run           - Build frontend + run server via go run"
	@echo "  make setup         - Build frontend + run setup wizard via go run"
	@echo "  make build-all     - Build frontend + compile binary (for releases)"
	@echo "  make build         - Compile backend binary only (build-frontend required first)"
	@echo "  make build-frontend - Build React UI into embed directory"
	@echo "  make test          - Build frontend + run backend tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make deps          - Download all dependencies (Go + NPM)"
	@echo "  make fmt           - Format Go code"
	@echo "  make lint          - Run linters (Go + frontend)"
	@echo "  make install-tools - Install development tools (golangci-lint v$(GOLANGCI_LINT_VERSION))"
	@echo "  make check-tools   - Check installed development tools"
	@echo "  make generate-keys - Generate RSA keys for JWT"
	@echo "  make help          - Show this help message"
