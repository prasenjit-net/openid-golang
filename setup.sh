#!/bin/bash

echo "Setting up OpenID Connect Development Environment..."
echo ""

# Create directories
echo "Step 1: Creating directories..."
mkdir -p config/keys
mkdir -p bin

# Download dependencies
echo ""
echo "Step 2: Downloading Go dependencies..."
go mod download
go mod tidy
echo "✓ Dependencies downloaded"

# Build the application
echo ""
echo "Step 3: Building application..."
go build -o bin/openid-server cmd/server/main.go
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi
echo "✓ Application built successfully"

# Run the setup wizard
echo ""
echo "Step 4: Running setup wizard..."
echo "This will generate keys, create config.toml, and set up users/clients."
echo ""
./bin/openid-server --setup

if [ $? -ne 0 ]; then
    echo ""
    echo "❌ Setup wizard failed or was cancelled"
    exit 1
fi

echo ""
echo "=========================================="
echo "Development Environment Setup Complete! 🎉"
echo "=========================================="
echo ""
echo "Next steps for development:"
echo "  1. Create test data: go run scripts/seed.go"
echo "  2. Start the server: make run"
echo "     or: ./bin/openid-server"
echo ""
echo "Configuration file: config.toml"
echo "RSA keys: config/keys/"
echo ""
