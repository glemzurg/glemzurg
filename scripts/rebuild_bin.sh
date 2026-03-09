#!/bin/bash
# Rebuild the req and req_check binaries into glemzurg/bin

# Exit quickly on an error.
set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_DIR="$PROJECT_ROOT/bin"

# Create bin directory if it doesn't exist
mkdir -p "$BIN_DIR"

cd "$PROJECT_ROOT/apps/requirements/req"

echo "Building req..."
go build -o "$BIN_DIR/req" "./cmd/req"

echo "Building req_check..."
go build -o "$BIN_DIR/req_check" "./cmd/req_check"

echo ""
echo "Binaries built successfully:"
ls -la "$BIN_DIR"
