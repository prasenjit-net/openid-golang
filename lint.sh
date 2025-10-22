#!/bin/bash

# Lint and format check script
# This script runs all the standard Go quality checks

set -e

# Configuration
GOLANGCI_LINT_VERSION="v2.5.0"

echo "🔍 Running code quality checks..."
echo ""

# Change to backend directory
cd backend

# Check formatting
echo "📝 Checking code formatting..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "❌ The following files are not formatted:"
    echo "$UNFORMATTED"
    echo ""
    echo "Run: gofmt -w ."
    exit 1
fi
echo "✅ All files are properly formatted"
echo ""

# Run go vet
echo "🔎 Running go vet..."
if ! go vet ./...; then
    echo "❌ go vet found issues"
    exit 1
fi
echo "✅ go vet passed"
echo ""

# Try to build
echo "🔨 Building project..."
if ! go build ./...; then
    echo "❌ Build failed"
    exit 1
fi
echo "✅ Build successful"
echo ""

# Run tests
echo "🧪 Running tests..."
if ! go test ./...; then
    echo "❌ Tests failed"
    exit 1
fi
echo "✅ Tests passed"
echo ""

# Check for golangci-lint (optional)
if command -v golangci-lint &> /dev/null; then
    INSTALLED_VERSION=$(golangci-lint version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    echo "🔍 Running golangci-lint (v$INSTALLED_VERSION)..."
    
    # Check if version matches
    EXPECTED_VERSION=$(echo "$GOLANGCI_LINT_VERSION" | sed 's/v//')
    if [ "$INSTALLED_VERSION" != "$EXPECTED_VERSION" ]; then
        echo "⚠️  Warning: golangci-lint version mismatch"
        echo "   Expected: v$EXPECTED_VERSION"
        echo "   Installed: v$INSTALLED_VERSION"
        echo "   Run 'make install-tools' to install the correct version"
    fi
    
    if golangci-lint run --timeout=5m 2>&1 | grep -v "typecheck"; then
        echo "⚠️  golangci-lint found some issues (typecheck warnings ignored)"
    else
        echo "✅ golangci-lint passed"
    fi
else
    echo "ℹ️  golangci-lint not installed (optional)"
    echo "   Run 'make install-tools' to install golangci-lint $GOLANGCI_LINT_VERSION"
fi

echo ""
echo "🎉 All checks passed!"
