#!/usr/bin/env bash
# Serve local pkgsite (godoc-style) docs for apps/requirements/req on :8080.
#
# Runs in the foreground so the terminal stays attached; stop with Ctrl+C.
#
# Usage:
#   ./scripts/req_godoc.sh
#   ./scripts/req_godoc.sh 9090          # optional port (default 8080)
#
# Then open (example for default port):
#   http://localhost:8080/github.com/glemzurg/glemzurg/apps/requirements/req
#   http://localhost:8080/github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/instance

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REQ_DIR="$REPO_ROOT/apps/requirements/req"
PORT="${1:-8080}"
MODULE_PATH="github.com/glemzurg/glemzurg/apps/requirements/req"

if [[ ! -d "$REQ_DIR" ]]; then
	echo "ERROR: requirements module not found at $REQ_DIR" >&2
	exit 1
fi

# Avoid root-owned /go/pkg/mod (common in this container) when downloading pkgsite.
export GOMODCACHE="${GOMODCACHE:-$HOME/go/pkg/mod}"
export GOCACHE="${GOCACHE:-$HOME/.cache/go-build}"
mkdir -p "$GOMODCACHE" "$GOCACHE"

echo "Serving pkgsite for $REQ_DIR on http://localhost:${PORT}"
echo "Module:    http://localhost:${PORT}/${MODULE_PATH}"
echo "Instance:  http://localhost:${PORT}/${MODULE_PATH}/internal/simulator/instance"
echo "Stop with Ctrl+C"
echo

cd "$REQ_DIR"
exec go run golang.org/x/pkgsite/cmd/pkgsite@latest -http=":${PORT}" .
