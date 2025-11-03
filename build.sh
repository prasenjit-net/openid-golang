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
mkdir -p backend/pkg/ui/uidist
# Remove existing contents but preserve dotfiles used to keep the folder in git (like .gitignore or .gitkeep)
# This deletes everything inside uidist except files named .gitignore or .gitkeep
if [ -d backend/pkg/ui/uidist ]; then
	find backend/pkg/ui/uidist -mindepth 1 \!
		-name '.gitignore' -a \!
		-name '.gitkeep' -print0 | xargs -0 rm -rf -- || true
fi
# Copy frontend build into uidist. Use the dot/dir form to include hidden files from dist as well.
cp -a frontend/dist/. backend/pkg/ui/uidist/

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
