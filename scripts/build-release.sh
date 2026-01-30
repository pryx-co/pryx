#!/bin/bash
set -e

VERSION="1.0.0"
DIST_DIR="dist"

echo "Building Pryx v${VERSION} release binaries..."

mkdir -p ${DIST_DIR}

# Build for macOS (Intel)
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.Version=${VERSION}" -o ${DIST_DIR}/pryx-v${VERSION}-darwin-amd64 ./cmd/pryx-core

# Build for macOS (Apple Silicon)
echo "Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.Version=${VERSION}" -o ${DIST_DIR}/pryx-v${VERSION}-darwin-arm64 ./cmd/pryx-core

# Build for Linux (x64)
echo "Building for Linux (x64)..."
GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Version=${VERSION}" -o ${DIST_DIR}/pryx-v${VERSION}-linux-amd64 ./cmd/pryx-core

# Build for Linux (ARM64)
echo "Building for Linux (ARM64)..."
GOOS=linux GOARCH=arm64 go build -ldflags="-X main.Version=${VERSION}" -o ${DIST_DIR}/pryx-v${VERSION}-linux-arm64 ./cmd/pryx-core

# Build for Windows (x64)
echo "Building for Windows (x64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-X main.Version=${VERSION}" -o ${DIST_DIR}/pryx-v${VERSION}-windows-amd64.exe ./cmd/pryx-core

echo "Generating checksums..."
cd ${DIST_DIR}
shasum -a 256 pryx-v${VERSION}-* > checksums.txt
cd ..

echo "✅ Build complete! Binaries in ${DIST_DIR}/"
echo ""
echo "Files created:"
ls -lh ${DIST_DIR}/
