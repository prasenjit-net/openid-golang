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
cd backend
go mod download
go mod tidy
cd ..
echo "✓ Dependencies downloaded"

# Build the application
echo ""
echo "Step 3: Building application..."
cd backend
go build -o ../bin/openid-server .
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi
cd ..
echo "✓ Application built successfully"

# Run the setup wizard
echo ""
echo "Step 4: Running setup wizard..."
echo "This will generate keys, create config.toml, and set up users/clients."
echo ""
./bin/openid-server setup

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
echo "  1. Start the server: ./bin/openid-server serve"
echo "     or: make run"
echo ""
echo "Configuration file: config.toml"
echo "RSA keys: config/keys/"
echo ""
