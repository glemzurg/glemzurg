#!/bin/bash
# Generate markdown docs for the development-process model.
#
# Usage:
#   ./examples/requirements/req/generate.sh
#   ./examples/requirements/req/generate.sh -debug
#
# Writes to examples/requirements/req/output/development-process

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

MODEL="development-process"
ROOTSOURCE="$SCRIPT_DIR/models"
ROOTOUTPUT="$SCRIPT_DIR/output"

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
