#!/bin/bash

# Build script for OpenID Connect Server with embedded Admin UI

set -e

# Get version from VERSION file or use dev
VERSION=$(cat VERSION 2>/dev/null || echo "dev")

echo "==> Building OpenID Connect Server v${VERSION}"
echo ""

echo "==> Building Admin UI..."
cd ui/admin
npm run build
cd ../..

echo "==> Copying UI to embed location..."
mkdir -p internal/ui/admin
rm -rf internal/ui/admin/dist
cp -r ui/admin/dist internal/ui/admin/

echo "==> Building Go server..."
go build -ldflags="-X main.Version=${VERSION}" -o bin/openid-server cmd/server/main.go

echo ""
echo "==> Build complete!"
echo "Version: ${VERSION}"
echo "Binary: ./bin/openid-server"
echo ""
echo "Run with: ./bin/openid-server"
echo "Check version: ./bin/openid-server --version"
