.PHONY: build run test clean deps

# Build the application
build:
	@echo "Building..."
	@go build -o bin/openid-server cmd/server/main.go

# Run the application
run:
	@echo "Running..."
	@go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Generate keys for JWT signing
generate-keys:
	@echo "Generating RSA key pair..."
	@mkdir -p config/keys
	@openssl genrsa -out config/keys/private.key 4096
	@openssl rsa -in config/keys/private.key -pubout -out config/keys/public.key
	@echo "Keys generated in config/keys/"
