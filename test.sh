#!/bin/bash

# OpenID Golang - Quick Test Script
# This script will guide you through testing the OpenID server

echo "╔════════════════════════════════════════════════════════════╗"
echo "║  OpenID Connect Identity Server - Test Script             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed!"
    echo ""
    echo "Install Go with one of these commands:"
    echo "  sudo apt install golang-go"
    echo "  sudo snap install go --classic"
    echo ""
    exit 1
fi

echo "✅ Go is installed: $(go version)"
echo ""

# Check if setup has been run
if [ ! -f "config/keys/private.key" ]; then
    echo "⚠️  Setup has not been run yet!"
    echo ""
    echo "Running setup now..."
    ./setup.sh
    echo ""
fi

# Check if database exists
if [ ! -f "openid.db" ]; then
    echo "⚠️  Database not seeded yet!"
    echo ""
    echo "Running seed script..."
    go run scripts/seed.go
    echo ""
    echo "⚠️  IMPORTANT: Save the Client ID and Client Secret shown above!"
    echo ""
    read -p "Press Enter to continue..."
fi

echo ""
echo "Starting OpenID Connect server..."
echo ""
echo "The server will start on http://localhost:8080"
echo ""
echo "Available endpoints:"
echo "  - http://localhost:8080/health"
echo "  - http://localhost:8080/.well-known/openid-configuration"
echo "  - http://localhost:8080/.well-known/jwks.json"
echo ""
echo "To test the full OAuth flow:"
echo "  1. Keep this server running"
echo "  2. Open a new terminal"
echo "  3. Run: go run examples/test-client.go"
echo "  4. Visit http://localhost:9090 in your browser"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""
echo "════════════════════════════════════════════════════════════"
echo ""

go run cmd/server/main.go
