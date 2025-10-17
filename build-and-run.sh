#!/bin/bash
# Quick build and run script for OpenID Connect Server with Admin UI

set -e

echo "🏗️  Building OpenID Connect Server with Embedded Admin UI"
echo "============================================================"

# Step 1: Build React Admin UI
echo ""
echo "📦 Step 1: Building React Admin UI..."
cd ui/admin
npm run build
cd ../..

# Step 2: Copy dist files to embed location
echo ""
echo "📋 Step 2: Copying build files for embedding..."
mkdir -p internal/ui/admin
rm -rf internal/ui/admin/dist
cp -r ui/admin/dist internal/ui/admin/

# Step 3: Build Go binary with embedded UI
echo ""
echo "🔨 Step 3: Building Go binary with embedded UI..."
go build -o bin/openid-server ./cmd/server

echo ""
echo "✅ Build Complete!"
echo ""
echo "To run the server:"
echo "  ./bin/openid-server"
echo ""
echo "Or run in background:"
echo "  nohup ./bin/openid-server > server.log 2>&1 &"
echo ""
echo "Access points:"
echo "  - Admin UI:  http://localhost:8080/"
echo "  - Admin API: http://localhost:8080/api/admin/*"
echo "  - Health:    http://localhost:8080/health"
echo "  - OIDC:      http://localhost:8080/.well-known/openid-configuration"
echo ""
