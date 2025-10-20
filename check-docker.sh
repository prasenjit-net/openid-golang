#!/bin/bash

# Docker setup check script

echo "🐳 Docker Setup Check"
echo "===================="
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed"
    echo ""
    echo "Install Docker:"
    echo "  Ubuntu/Debian: sudo apt-get install docker.io"
    echo "  Fedora: sudo dnf install docker"
    echo "  Arch: sudo pacman -S docker"
    exit 1
fi

echo "✅ Docker is installed: $(docker --version)"
echo ""

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    echo "⚠️  Docker daemon is not accessible"
    echo ""
    
    # Check if it's a permission issue
    if docker info 2>&1 | grep -q "permission denied"; then
        echo "❌ Permission denied - User not in docker group"
        echo ""
        echo "Fix this by running:"
        echo "  sudo usermod -aG docker \$USER"
        echo "  newgrp docker"
        echo ""
        echo "Or run docker commands with sudo:"
        echo "  sudo docker ps"
        exit 1
    else
        echo "❌ Docker daemon is not running"
        echo ""
        echo "Start Docker with:"
        echo "  sudo systemctl start docker"
        echo "  sudo systemctl enable docker"
        exit 1
    fi
fi

echo "✅ Docker daemon is accessible"
echo ""

# Check user groups
if groups | grep -q docker; then
    echo "✅ User is in docker group"
else
    echo "⚠️  User is NOT in docker group"
    echo ""
    echo "Add yourself to docker group:"
    echo "  sudo usermod -aG docker \$USER"
    echo "  newgrp docker"
fi

echo ""
echo "Docker Setup: OK ✅"
echo ""
echo "You can now run:"
echo "  ./docker-build.sh"
echo "  docker-compose up -d"
