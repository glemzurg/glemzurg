#!/bin/bash
# Run the exercise simulator against a human-readable model.
#
# Example usage:
#   Default root (data_sandbox/model), finance/wallet subdomain, seed 42:
#     ./scripts/simulate.sh evenplay 42
#
#   Explicit subdomain scope:
#     ./scripts/simulate.sh evenplay 42 finance/wallet
#
#   One or more fully scoped classes (domain/subdomain/class) as positional args:
#     ./scripts/simulate.sh evenplay 42 finance/wallet/partner
#     ./scripts/simulate.sh evenplay 42 finance/wallet/partner finance/wallet/currency
#
#   Class scope via flag (class-only; no subdomain filter):
#     ./scripts/simulate.sh evenplay 42 --include-class finance/wallet/partner
#     ./scripts/simulate.sh evenplay 42 --include-class wallet/partner,finance/wallet/currency
#
#   Legacy class-only marker (- skips subdomain filter):
#     ./scripts/simulate.sh evenplay 42 - --include-class wallet/partner
#
#   Full step trace:
#     ./scripts/simulate.sh evenplay 42 finance/wallet --trace
#
#   Keep simulating after the first violation:
#     ./scripts/simulate.sh evenplay 42 finance/wallet --continue-on-violation
#
#   Custom model root:
#     ./scripts/simulate.sh data_sandbox/model evenplay 42 finance/wallet
#
#   Examples tree:
#     ./scripts/simulate.sh /data/examples/requirements/req/models/ web_books 1

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEFAULT_ROOTSOURCE="$REPO_ROOT/data_sandbox/model"

usage() {
    echo "Usage: $0 MODEL SEED [SCOPE...] [OPTIONS...]"
    echo "       $0 ROOTSOURCE MODEL SEED [SCOPE...] [OPTIONS...]"
    echo ""
    echo "  MODEL              Model name under the root source (e.g. evenplay)"
    echo "  SEED               Random seed for reproducible runs (e.g. 42)"
    echo "  SCOPE              Subdomain path (domain/subdomain or subdomain), fully scoped"
    echo "                     class path (domain/subdomain/class), or - for class-only scope"
    echo "                     (default subdomain when no class scope: finance/wallet)"
    echo "  ROOTSOURCE         Human model root directory (default: data_sandbox/model)"
    echo ""
    echo "Options (passed to simulate):"
    echo "  --include-class PATH     Narrow scope to specific class(es): name, subdomain/class,"
    echo "                           or domain/subdomain/class (comma-separated for multiple)"
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

# Three path segments (domain/subdomain/class) identify a fully scoped class.
is_fully_scoped_class() {
    local path="$1"
    local segment_count
    segment_count="$(echo "$path" | tr -cd '/' | wc -c)"
    [ "$segment_count" -eq 2 ]
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

INCLUDE_SUBDOMAIN=""
CLASS_ONLY=false
INCLUDE_CLASSES=()

while [ $# -gt 0 ] && [[ "$1" != --* ]]; do
    case "$1" in
        -)
            CLASS_ONLY=true
            shift
            ;;
        *)
            if is_fully_scoped_class "$1"; then
                INCLUDE_CLASSES+=("$1")
                CLASS_ONLY=true
                shift
            else
                INCLUDE_SUBDOMAIN="$1"
                shift
            fi
            ;;
    esac
done

EXTRA_FLAGS=("$@")

FILTERED_FLAGS=()
skip_next=false
for ((i = 0; i < ${#EXTRA_FLAGS[@]}; i++)); do
    if [ "$skip_next" = true ]; then
        skip_next=false
        continue
    fi
    if [ "${EXTRA_FLAGS[$i]}" = "--include-class" ]; then
        CLASS_ONLY=true
        if [ $((i + 1)) -lt ${#EXTRA_FLAGS[@]} ]; then
            INCLUDE_CLASSES+=("${EXTRA_FLAGS[$((i + 1))]}")
            skip_next=true
        fi
        continue
    fi
    FILTERED_FLAGS+=("${EXTRA_FLAGS[$i]}")
done
EXTRA_FLAGS=("${FILTERED_FLAGS[@]}")

if [ "$CLASS_ONLY" = true ]; then
    INCLUDE_SUBDOMAIN=""
elif [ -z "$INCLUDE_SUBDOMAIN" ]; then
    INCLUDE_SUBDOMAIN="finance/wallet"
fi

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
    -seed "$SEED"
)
if [ -n "$INCLUDE_SUBDOMAIN" ]; then
    CMD+=(-include-subdomain "$INCLUDE_SUBDOMAIN")
fi
if [ ${#INCLUDE_CLASSES[@]} -gt 0 ]; then
    # Flatten comma-separated --include-class values and positional class paths.
    INCLUDE_CLASS_CSV=""
    for entry in "${INCLUDE_CLASSES[@]}"; do
        if [ -z "$INCLUDE_CLASS_CSV" ]; then
            INCLUDE_CLASS_CSV="$entry"
        else
            INCLUDE_CLASS_CSV="$INCLUDE_CLASS_CSV,$entry"
        fi
    done
    CMD+=(-include-class "$INCLUDE_CLASS_CSV")
fi
if [ ${#EXTRA_FLAGS[@]} -gt 0 ]; then
    CMD+=("${EXTRA_FLAGS[@]}")
fi

echo -e "\n${CMD[*]}\n"
"${CMD[@]}"