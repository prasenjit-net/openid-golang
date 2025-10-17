#!/bin/bash

echo "Setting up OpenID Connect Identity Server..."

# Create directories
echo "Creating directories..."
mkdir -p config/keys
mkdir -p bin

# Generate RSA keys
echo "Generating RSA key pair..."
if [ ! -f config/keys/private.key ]; then
    openssl genrsa -out config/keys/private.key 4096
    openssl rsa -in config/keys/private.key -pubout -out config/keys/public.key
    echo "✓ RSA keys generated"
else
    echo "✓ RSA keys already exist"
fi

# Copy environment file if it doesn't exist
if [ ! -f .env ]; then
    cp .env.example .env
    echo "✓ Created .env file from .env.example"
else
    echo "✓ .env file already exists"
fi

# Download dependencies
echo "Downloading Go dependencies..."
go mod download
go mod tidy
echo "✓ Dependencies downloaded"

# Build the application
echo "Building application..."
go build -o bin/openid-server cmd/server/main.go
echo "✓ Application built successfully"

echo ""
echo "Setup complete! 🎉"
echo ""
echo "To create a test user and client, run:"
echo "  go run scripts/seed.go"
echo ""
echo "To start the server, run:"
echo "  make run"
echo "  or"
echo "  ./bin/openid-server"
echo ""
