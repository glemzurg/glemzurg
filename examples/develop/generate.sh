#!/bin/bash
# Generate markdown docs for the develop model.
#
# Usage:
#   ./examples/develop/generate.sh
#   ./examples/develop/generate.sh -debug
#
# Writes to examples/develop/req/output/develop

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

MODEL="develop"
ROOTSOURCE="$SCRIPT_DIR/req/models"
ROOTOUTPUT="$SCRIPT_DIR/req/output"

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "  Generate the $MODEL model from $ROOTSOURCE"
    echo "  into $ROOTOUTPUT/$MODEL"
    echo ""
    echo "Options:"
    echo "  -debug     Enable req debug logging"
    echo "  -skipdb    Skip database validation"
    echo "  -h, --help Show this help"
}

OPTIONAL_FLAGS=()
while [ $# -gt 0 ]; do
    case "$1" in
        -h|--help)
            usage
            exit 0
            ;;
        -debug|-skipdb)
            OPTIONAL_FLAGS+=("$1")
            shift
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

mkdir -p "$ROOTOUTPUT"
rm -rf "$ROOTOUTPUT/$MODEL"
mkdir -p "$ROOTOUTPUT/$MODEL"

REQ_BIN="/go/bin/req"
if [ ! -x "$REQ_BIN" ]; then
    REQ_BIN="$(command -v req)"
fi

echo -e "\n$REQ_BIN -rootsource $ROOTSOURCE -rootoutput $ROOTOUTPUT -model $MODEL ${OPTIONAL_FLAGS[*]}\n"
"$REQ_BIN" -rootsource "$ROOTSOURCE" -rootoutput "$ROOTOUTPUT" -model "$MODEL" "${OPTIONAL_FLAGS[@]}"
