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
if [ ! -f "data.json" ]; then
    echo "⚠️  Database not initialized yet!"
    echo ""
    echo "Please run setup first:"
    echo "  ./bin/openid-server setup"
    echo ""
    echo "Or run in demo mode:"
    echo "  ./bin/openid-server setup --demo"
    echo ""
    exit 1
fi

echo ""
echo "Starting OpenID Connect server..."
echo ""
echo "The server will start on http://localhost:8080"
echo ""
echo "Available endpoints:"
echo "  - http://localhost:8080/"
echo "  - http://localhost:8080/.well-known/openid-configuration"
echo "  - http://localhost:8080/.well-known/jwks.json"
echo "  - http://localhost:8080/api/admin/stats"
echo ""
echo "To test the full OAuth flow:"
echo "  1. Keep this server running"
echo "  2. Open a new terminal"
echo "  3. Run: cd backend && go run ../examples/test-client.go"
echo "  4. Visit http://localhost:9090 in your browser"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""
echo "════════════════════════════════════════════════════════════"
echo ""

./bin/openid-server serve
