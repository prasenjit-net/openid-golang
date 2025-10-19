#!/bin/bash

# Docker build script for OpenID Connect Server
# Builds optimized Docker images with proper tagging

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🐳 Building OpenID Connect Server Docker Image"
echo "=============================================="
echo ""

# Get version from VERSION file or use dev
VERSION=$(cat VERSION 2>/dev/null || echo "dev")
echo "Version: ${VERSION}"

# Image name
IMAGE_NAME="openid-server"
REGISTRY="${DOCKER_REGISTRY:-}"

# Build tags
if [ -n "$REGISTRY" ]; then
    IMAGE_TAG="${REGISTRY}/${IMAGE_NAME}:${VERSION}"
    IMAGE_TAG_LATEST="${REGISTRY}/${IMAGE_NAME}:latest"
else
    IMAGE_TAG="${IMAGE_NAME}:${VERSION}"
    IMAGE_TAG_LATEST="${IMAGE_NAME}:latest"
fi

echo "Building image: ${IMAGE_TAG}"
echo ""

# Build the Docker image
echo "📦 Building Docker image..."
docker build \
    --build-arg VERSION="${VERSION}" \
    --tag "${IMAGE_TAG}" \
    --tag "${IMAGE_TAG_LATEST}" \
    .

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ Docker image built successfully!${NC}"
    echo ""
    echo "Image tags:"
    echo "  - ${IMAGE_TAG}"
    echo "  - ${IMAGE_TAG_LATEST}"
    echo ""
    echo "Image size:"
    docker images "${IMAGE_NAME}" | grep -E "${VERSION}|latest" | head -2
    echo ""
    echo "To run the container:"
    echo "  docker run -p 8080:8080 ${IMAGE_TAG_LATEST}"
    echo ""
    echo "Or use docker-compose:"
    echo "  docker-compose up -d"
    echo ""
    
    # Offer to push to registry
    if [ -n "$REGISTRY" ]; then
        echo -e "${YELLOW}Push to registry? (y/N):${NC} "
        read -r PUSH_RESPONSE
        if [[ "$PUSH_RESPONSE" =~ ^[Yy]$ ]]; then
            echo "Pushing images..."
            docker push "${IMAGE_TAG}"
            docker push "${IMAGE_TAG_LATEST}"
            echo -e "${GREEN}✅ Images pushed to registry${NC}"
        fi
    fi
else
    echo ""
    echo -e "${RED}❌ Docker build failed${NC}"
    exit 1
fi
