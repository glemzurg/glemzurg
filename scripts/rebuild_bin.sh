#!/bin/bash
# Rebuild the req and req_check binaries into glemzurg/bin

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_DIR="$PROJECT_ROOT/bin"

# Create bin directory if it doesn't exist
mkdir -p "$BIN_DIR"

echo "Building req..."
go build -o "$BIN_DIR/req" "$PROJECT_ROOT/apps/requirements/req/cmd/req"

echo "Building req_check..."
go build -o "$BIN_DIR/req_check" "$PROJECT_ROOT/apps/requirements/req_check/cmd/req_check"

echo ""
echo "Binaries built successfully:"
ls -la "$BIN_DIR"
