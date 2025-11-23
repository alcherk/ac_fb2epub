#!/bin/bash

# Run script for fb2epub service
# Handles macOS LC_UUID issues by trying multiple methods

set -e

echo "Starting FB2 to EPUB Converter Service..."

# Method 1: Try Docker (if available)
if command -v docker &> /dev/null && docker info &> /dev/null; then
    echo "Using Docker..."
    docker-compose up -d
    echo "Service started! Access at http://localhost:8080"
    echo "View logs: docker-compose logs -f"
    echo "Stop: docker-compose down"
    exit 0
fi

# Method 2: Try go run
echo "Trying go run..."
if go run main.go 2>&1 | grep -q "LC_UUID"; then
    echo "Warning: LC_UUID error detected with go run"
    echo "Trying alternative methods..."
else
    # If no error, go run is working
    go run main.go
    exit 0
fi

# Method 3: Try go install
echo "Trying go install..."
go install . 2>/dev/null || true
if [ -f ~/go/bin/fb2epub ]; then
    echo "Using installed binary..."
    ~/go/bin/fb2epub
    exit 0
fi

# Method 4: Try local binary
if [ -f ./fb2epub ]; then
    echo "Trying local binary..."
    ./fb2epub
    exit 0
fi

# Fallback: Build and run
echo "Building and running..."
go build -trimpath -o fb2epub .
./fb2epub

