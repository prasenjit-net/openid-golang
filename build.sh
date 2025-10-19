#!/bin/bash

# Build script for OpenID Connect Server with embedded Admin UI

set -e

# Get version from VERSION file or use dev
VERSION=$(cat VERSION 2>/dev/null || echo "dev")

echo "==> Building OpenID Connect Server v${VERSION}"
echo ""

echo "==> Building Frontend (React Admin UI)..."
cd frontend
npm run build
cd ..

echo "==> Copying UI to embed location..."
mkdir -p backend/pkg/ui/admin
rm -rf backend/pkg/ui/admin/dist
cp -r frontend/dist backend/pkg/ui/admin/

echo "==> Building Go backend server..."
cd backend
go build -ldflags="-X main.Version=${VERSION}" -o ../bin/openid-server .
cd ..

echo ""
echo "==> Build complete!"
echo "Version: ${VERSION}"
echo "Binary: ./bin/openid-server"
echo ""
echo "Run with: ./bin/openid-server"
echo "Check version: ./bin/openid-server --version"
