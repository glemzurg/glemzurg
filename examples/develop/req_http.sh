#!/bin/bash
# Start the HTTP server for the develop model.
#
# Usage:
#   ./examples/develop/req_http.sh
#   ./examples/develop/req_http.sh 9090
#   ./examples/develop/req_http.sh 9090 -debug
#
# Serves examples/develop/req/models/develop (data/yaml).

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

MODEL="develop"
ROOTSOURCE="$SCRIPT_DIR/req/models"
PORT="${1:-8080}"
INPUT_FORMAT="data/yaml"

usage() {
    echo "Usage: $0 [PORT] [-debug]"
    echo ""
    echo "  Start the req HTTP server for $MODEL from $ROOTSOURCE"
    echo "  Default PORT is 8080."
}

OPTIONAL_FLAGS=()
if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
    usage
    exit 0
fi

if [ $# -gt 0 ]; then
    case "$1" in
        -debug)
            OPTIONAL_FLAGS+=("-debug")
            shift
            ;;
        *)
            PORT="$1"
            shift
            ;;
    esac
fi

while [ $# -gt 0 ]; do
    case "$1" in
        -debug)
            OPTIONAL_FLAGS+=("-debug")
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "ERROR: unknown argument: $1"
            usage
            exit 1
            ;;
    esac
done

echo -e "\nUPDATE INSTALL\n"
(cd "$REPO_ROOT/apps/requirements/req" && go install -buildvcs=false ./...)

REQ_BIN="/go/bin/req"
if [ ! -x "$REQ_BIN" ]; then
    REQ_BIN="$(command -v req)"
fi

echo -e "\n$REQ_BIN -http -port $PORT -rootsource $ROOTSOURCE -model $MODEL -input $INPUT_FORMAT ${OPTIONAL_FLAGS[*]}\n"
"$REQ_BIN" -http -port "$PORT" -rootsource "$ROOTSOURCE" -model "$MODEL" -input "$INPUT_FORMAT" "${OPTIONAL_FLAGS[@]}"
