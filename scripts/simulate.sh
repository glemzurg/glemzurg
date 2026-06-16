#!/bin/bash
# Run the exercise simulator against a human-readable model.
#
# Example usage:
#   Default root (data_sandbox/model), jurisdiction class, seed 42:
#     ./scripts/simulate.sh evenplay 42
#
#   Explicit class scope:
#     ./scripts/simulate.sh evenplay 42 jurisdiction
#
#   Subdomain-qualified class (avoids matching same-named classes elsewhere):
#     ./scripts/simulate.sh evenplay 42 wallet/partner
#
#   Full step trace:
#     ./scripts/simulate.sh evenplay 42 jurisdiction --trace
#
#   Keep simulating after the first violation:
#     ./scripts/simulate.sh evenplay 42 jurisdiction --continue-on-violation
#
#   Custom model root:
#     ./scripts/simulate.sh data_sandbox/model evenplay 42 jurisdiction
#
#   Examples tree:
#     ./scripts/simulate.sh /data/examples/requirements/req/models/ web_books 1

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEFAULT_ROOTSOURCE="$REPO_ROOT/data_sandbox/model"

usage() {
    echo "Usage: $0 MODEL SEED [INCLUDE_CLASS] [OPTIONS...]"
    echo "       $0 ROOTSOURCE MODEL SEED [INCLUDE_CLASS] [OPTIONS...]"
    echo ""
    echo "  MODEL          Model name under the root source (e.g. evenplay)"
    echo "  SEED           Random seed for reproducible runs (e.g. 42)"
    echo "  INCLUDE_CLASS  Comma-separated classes: name, subdomain/class, or domain/subdomain/class (default: jurisdiction)"
    echo "  ROOTSOURCE     Human model root directory (default: data_sandbox/model)"
    echo ""
    echo "Options (passed to simulate):"
    echo "  --trace                  Include full step trace in output"
    echo "  --continue-on-violation  Keep simulating after violations"
    echo "  --max-steps N            Maximum simulation steps (default: 100)"
    echo "  --quiet                  Only output violations"
    echo "  --output FORMAT          text (default) or json"
}

resolve_relative_path() {
    local path="$1"
    if [[ "$path" != /* ]]; then
        echo "$REPO_ROOT/$path"
    else
        echo "$path"
    fi
}

if [ $# -lt 2 ]; then
    usage
    exit 1
fi

if [[ "$2" =~ ^-?[0-9]+$ ]]; then
    ROOTSOURCE="$DEFAULT_ROOTSOURCE"
    MODEL="$1"
    SEED="$2"
    shift 2
elif [ $# -ge 3 ] && [[ "$3" =~ ^-?[0-9]+$ ]]; then
    ROOTSOURCE="$(resolve_relative_path "$1")"
    MODEL="$2"
    SEED="$3"
    shift 3
else
    echo "ERROR: SEED must be an integer."
    usage
    exit 1
fi

if [ -z "$MODEL" ]; then
    echo "ERROR: MODEL is required."
    usage
    exit 1
fi

INCLUDE_CLASS="jurisdiction"
if [ -n "$1" ] && [[ "$1" != --* ]]; then
    INCLUDE_CLASS="$1"
    shift
fi

EXTRA_FLAGS=("$@")

if [[ "$ROOTSOURCE" != /* ]]; then
    ROOTSOURCE="$(resolve_relative_path "$ROOTSOURCE")"
fi

MODEL_PATH="$ROOTSOURCE/$MODEL"
if [ ! -d "$MODEL_PATH" ]; then
    echo "ERROR: Model directory not found: $MODEL_PATH"
    exit 1
fi

echo -e "\nBUILD simulate\n"
(cd "$REPO_ROOT/apps/requirements/req" && go build -buildvcs=false -o "$REPO_ROOT/bin/simulate" "./cmd/simulate")

SIMULATE_BIN="$REPO_ROOT/bin/simulate"
if [ ! -x "$SIMULATE_BIN" ]; then
    SIMULATE_BIN="/go/bin/simulate"
fi

CMD=(
    "$SIMULATE_BIN"
    -rootsource "$ROOTSOURCE"
    -model "$MODEL"
    -include-class "$INCLUDE_CLASS"
    -seed "$SEED"
)
if [ ${#EXTRA_FLAGS[@]} -gt 0 ]; then
    CMD+=("${EXTRA_FLAGS[@]}")
fi

echo -e "\n${CMD[*]}\n"
"${CMD[@]}"