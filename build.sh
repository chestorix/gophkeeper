#!/bin/bash

mkdir -p dist

echo "Building clients for different platforms..."

# Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o dist/gophkeeper-client-linux cmd/client/main.go

# Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o dist/gophkeeper-client.exe cmd/client/main.go

# MacOS (Intel)
echo "Building for MacOS (Intel)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o dist/gophkeeper-client-macos-intel cmd/client/main.go

# MacOS (Apple Silicon)
echo "Building for MacOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o dist/gophkeeper-client-macos-apple cmd/client/main.go

echo "Build complete! Clients are in ./dist directory:"
ls -la dist/