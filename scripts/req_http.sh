#!/bin/bash
# Start the HTTP server for a requirements model.
#
# Example usage:
#   Default (data/yaml format):
#     ./scripts/req_http.sh /data/examples/requirements/req/models/ web_books
#
#   With custom port:
#     ./scripts/req_http.sh /data/examples/requirements/req/models/ web_books 9090
#
#   With ai/json format:
#     ./scripts/req_http.sh /data/examples/requirements/req/ai_models/ web_books 8080 ai/json

SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Required parameters
INPUT_PATH="$1"
MODEL="$2"

# Optional parameters
PORT="${3:-8080}"           # Default to 8080
INPUT_FORMAT="${4:-data/yaml}"  # data/yaml (default) or ai/json

# Check required parameters
if [ -z "$INPUT_PATH" ]; then
    echo "ERROR: INPUT_PATH is required. Please provide the input path as the first argument."
    exit 1
fi

if [ -z "$MODEL" ]; then
    echo "ERROR: MODEL is required. Please provide the model name as the second argument."
    exit 1
fi

# We are in the script path directory.
cd $SCRIPT_PATH
[ $? -ne 0 ] && exit 1

# ================================================

# For easy iteration, rebuild the command line tool.
echo -e "\nUPDATE INSTALL\n"
(cd $SCRIPT_PATH/../apps/requirements/req && go install ./...)
[ $? -ne 0 ] && exit 1

# ================================================

# Run the HTTP server.
echo -e "\n/go/bin/req -http -port $PORT -rootsource $INPUT_PATH -model $MODEL -input $INPUT_FORMAT\n"
/go/bin/req -http -port $PORT -rootsource $INPUT_PATH -model $MODEL -input $INPUT_FORMAT

[ $? -ne 0 ] && exit 1

# Everything is fine.
exit 0
