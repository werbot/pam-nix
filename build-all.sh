#!/bin/bash
# Script to build wpam.so for all supported platforms
# Uses Docker for reliable cross-compilation

set -e

BIN_NAME=wpam
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")

echo "Building ${BIN_NAME} ${VERSION} for all platforms"
echo "================================================"

# Check if Docker is available
if command -v docker &> /dev/null; then
    echo "Using Docker for cross-compilation..."
    
    # Build Linux amd64
    echo "Building for linux/amd64..."
    docker run --rm -v "$(pwd)":/work -w /work golang:1.21 \
        bash -c "CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=c-shared -o ${BIN_NAME}-linux-amd64.so" || echo "Failed: linux/amd64"
    
    # Build Linux arm64
    echo "Building for linux/arm64..."
    docker run --rm -v "$(pwd)":/work -w /work golang:1.21 \
        bash -c "CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -buildmode=c-shared -o ${BIN_NAME}-linux-arm64.so" || echo "Failed: linux/arm64"
    
    echo ""
    echo "Build complete! Files:"
    ls -lh ${BIN_NAME}-*.so 2>/dev/null || echo "No binaries found"
else
    echo "Docker not found. Attempting direct cross-compilation..."
    echo "Note: This may fail if cross-compilers are not installed"
    echo ""
    make build-all
fi
