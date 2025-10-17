#!/bin/bash

# Build script for OpenID Connect Server with embedded Admin UI

set -e

echo "==> Building Admin UI..."
cd ui/admin
npm run build
cd ../..

echo "==> Building Go server..."
go build -o bin/openid-server cmd/server/main.go

echo "==> Build complete!"
echo "Binary: ./bin/openid-server"
