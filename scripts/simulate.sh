#!/bin/bash
# Run the exercise simulator against a human-readable model.
#
# Scope is a union: a class is in the run if it is listed explicitly OR it belongs
# to an included subdomain (or domain). Both can be set together on the command line.
#
# Example usage:
#   Default root (data_sandbox/model), finance/wallet subdomain, seed 42:
#     ./scripts/simulate.sh evenplay 42
#
#   Explicit subdomain scope (all classes in that subdomain):
#     ./scripts/simulate.sh evenplay 42 finance/wallet
#     ./scripts/simulate.sh evenplay 42 --include-subdomain finance/wallet
#
#   One or more fully scoped classes (domain/subdomain/class) as positional args:
#     ./scripts/simulate.sh evenplay 42 finance/wallet/partner
#     ./scripts/simulate.sh evenplay 42 finance/wallet/partner finance/wallet/currency
#
#   Subdomain + extra classes (union — both contribute to scope):
#     ./scripts/simulate.sh evenplay 42 finance/wallet finance/operations/fee
#     ./scripts/simulate.sh evenplay 42 --include-subdomain finance/wallet \
#         --include-class finance/operations/fee
#
#   Class scope via flag:
#     ./scripts/simulate.sh evenplay 42 --include-class finance/wallet/partner
#     ./scripts/simulate.sh evenplay 42 --include-class wallet/partner,finance/wallet/currency
#
#   Class-only (no default subdomain):
#     ./scripts/simulate.sh evenplay 42 - --include-class finance/wallet/partner
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
    echo "  SCOPE              Zero or more of:"
    echo "                       domain/subdomain  — include all classes in the subdomain"
    echo "                       domain/subdomain/class — include that class only"
    echo "                       -                 — do not apply the default subdomain"
    echo "                     Scope is a union: subdomain includes + class includes."
    echo "                     Default when no scope is given: finance/wallet"
    echo "  ROOTSOURCE         Human model root directory (default: data_sandbox/model)"
    echo ""
    echo "Options (passed to simulate):"
    echo "  --include-subdomain PATH  Subdomain(s) to include (comma-separated; may repeat)"
    echo "  --include-class PATH      Class(es) to include: name, subdomain/class, or"
    echo "                            domain/subdomain/class (comma-separated; may repeat)"
    echo "  --trace                   Include full step trace in output"
    echo "  --continue-on-violation   Keep simulating after violations"
    echo "  --max-steps N             Maximum simulation steps (default: 100)"
    echo "  --quiet                   Only output violations"
    echo "  --output FORMAT           text (default) or json"
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

# Append comma-separated entries to a bash array (name passed by reference via nameref-like eval).
append_csv_to_array() {
    local -n _arr=$1
    local csv="$2"
    local IFS=','
    local part
    for part in $csv; do
        part="$(echo "$part" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
        if [ -n "$part" ]; then
            _arr+=("$part")
        fi
    done
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

INCLUDE_SUBDOMAINS=()
INCLUDE_CLASSES=()
# When true, do not apply the default finance/wallet subdomain if nothing was scoped.
SKIP_DEFAULT_SUBDOMAIN=false

while [ $# -gt 0 ] && [[ "$1" != --* ]]; do
    case "$1" in
        -)
            SKIP_DEFAULT_SUBDOMAIN=true
            shift
            ;;
        *)
            if is_fully_scoped_class "$1"; then
                INCLUDE_CLASSES+=("$1")
            else
                INCLUDE_SUBDOMAINS+=("$1")
            fi
            shift
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
    case "${EXTRA_FLAGS[$i]}" in
        --include-class)
            if [ $((i + 1)) -lt ${#EXTRA_FLAGS[@]} ]; then
                append_csv_to_array INCLUDE_CLASSES "${EXTRA_FLAGS[$((i + 1))]}"
                skip_next=true
            fi
            ;;
        --include-subdomain)
            if [ $((i + 1)) -lt ${#EXTRA_FLAGS[@]} ]; then
                append_csv_to_array INCLUDE_SUBDOMAINS "${EXTRA_FLAGS[$((i + 1))]}"
                skip_next=true
            fi
            ;;
        *)
            FILTERED_FLAGS+=("${EXTRA_FLAGS[$i]}")
            ;;
    esac
done
EXTRA_FLAGS=("${FILTERED_FLAGS[@]}")

# Default subdomain only when neither subdomain nor class scope was provided.
if [ ${#INCLUDE_SUBDOMAINS[@]} -eq 0 ] && [ ${#INCLUDE_CLASSES[@]} -eq 0 ]; then
    if [ "$SKIP_DEFAULT_SUBDOMAIN" = false ]; then
        INCLUDE_SUBDOMAINS=("finance/wallet")
    fi
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

if [ ${#INCLUDE_SUBDOMAINS[@]} -gt 0 ]; then
    INCLUDE_SUBDOMAIN_CSV=""
    for entry in "${INCLUDE_SUBDOMAINS[@]}"; do
        if [ -z "$INCLUDE_SUBDOMAIN_CSV" ]; then
            INCLUDE_SUBDOMAIN_CSV="$entry"
        else
            INCLUDE_SUBDOMAIN_CSV="$INCLUDE_SUBDOMAIN_CSV,$entry"
        fi
    done
    CMD+=(-include-subdomain "$INCLUDE_SUBDOMAIN_CSV")
fi

if [ ${#INCLUDE_CLASSES[@]} -gt 0 ]; then
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
