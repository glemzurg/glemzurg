#!/bin/bash
# Print human-readable class association facts for one model subdomain.
#
# Example usage:
#   Default (data_sandbox evenplay finance/wallet):
#     ./scripts/associationfacts.sh data_sandbox/model evenplay finance/wallet
#
#   With debug:
#     ./scripts/associationfacts.sh data_sandbox/model evenplay finance/wallet -debug
#
#   Examples tree:
#     ./scripts/associationfacts.sh /data/examples/requirements/req/models/ web_books finance/wallet

SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Required parameters
INPUT_PATH="$1"
MODEL="$2"
SUBDOMAIN="$3"

# Optional parameters
DEBUG="$4"           # -debug or empty

# Check required parameters
if [ -z "$INPUT_PATH" ]; then
    echo "ERROR: INPUT_PATH is required. Please provide the model root path as the first argument."
    exit 1
fi

if [ -z "$MODEL" ]; then
    echo "ERROR: MODEL is required. Please provide the model name as the second argument."
    exit 1
fi

if [ -z "$SUBDOMAIN" ]; then
    echo "ERROR: SUBDOMAIN is required. Please provide the domain/subdomain path as the third argument (e.g. finance/wallet)."
    exit 1
fi

# Resolve relative input paths against the repository root (parent of scripts/).
REPO_ROOT="$( cd "$SCRIPT_PATH/.." && pwd )"
if [[ "$INPUT_PATH" != /* ]]; then
    INPUT_PATH="$REPO_ROOT/$INPUT_PATH"
fi

# Build the optional flags
OPTIONAL_FLAGS=""

if [ -n "$DEBUG" ]; then
    OPTIONAL_FLAGS="$OPTIONAL_FLAGS $DEBUG"
fi

# We are in the script path directory.
cd $SCRIPT_PATH
[ $? -ne 0 ] && exit 1

# ================================================

# For easy iteration, rebuild the command line tool.
echo -e "\nUPDATE INSTALL\n"
(cd $SCRIPT_PATH/../apps/requirements/req && go install -buildvcs=false ./...)
[ $? -ne 0 ] && exit 1

# ================================================

# Run the command to print association facts for the subdomain.
echo -e "\n/go/bin/req -associationfacts -rootsource $INPUT_PATH -model $MODEL -subdomain $SUBDOMAIN$OPTIONAL_FLAGS\n"
/go/bin/req -associationfacts -rootsource $INPUT_PATH -model $MODEL -subdomain $SUBDOMAIN $OPTIONAL_FLAGS

[ $? -ne 0 ] && exit 1

# Everything is fine.
exit 0