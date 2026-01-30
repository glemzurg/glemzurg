#!/bin/bash
# Example usage:
#   Default (data/yaml to md):
#     ./scripts/req_examples.sh /data/examples/requirements/req/models/ /data/examples/requirements/req/output/ web_books
#
#   With debug:
#     ./scripts/req_examples.sh /data/examples/requirements/req/models/ /data/examples/requirements/req/output/ web_books -debug
#
#   Convert data/yaml to ai/json:
#     ./scripts/req_examples.sh /data/examples/requirements/req/models/ /data/examples/requirements/req/ai_output/ web_books -debug "data/yaml" "ai/json"
#
#   Convert ai/json to md:
#     ./scripts/req_examples.sh /data/examples/requirements/req/ai_models/ /data/examples/requirements/req/output/ web_books "" "ai/json" "md"
#
#   Convert ai/json to data/yaml:
#     ./scripts/req_examples.sh /data/examples/requirements/req/ai_models/ /data/examples/requirements/req/models_output/ web_books "" "ai/json" "data/yaml"

SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Required parameters
INPUT_PATH="$1"
OUTPUT_PATH="$2"
MODEL="$3"

# Optional parameters
DEBUG="$4"           # -debug or empty
INPUT_FORMAT="$5"    # data/yaml (default) or ai/json
OUTPUT_FORMAT="$6"   # md (default), data/yaml, or ai/json

# Check required parameters
if [ -z "$INPUT_PATH" ]; then
    echo "ERROR: INPUT_PATH is required. Please provide the input path as the first argument."
    exit 1
fi

if [ -z "$OUTPUT_PATH" ]; then
    echo "ERROR: OUTPUT_PATH is required. Please provide the output path as the second argument."
    exit 1
fi

if [ -z "$MODEL" ]; then
    echo "ERROR: MODEL is required. Please provide the model name as the third argument."
    exit 1
fi

# Build the optional flags
OPTIONAL_FLAGS=""

if [ -n "$DEBUG" ]; then
    OPTIONAL_FLAGS="$OPTIONAL_FLAGS $DEBUG"
fi

if [ -n "$INPUT_FORMAT" ]; then
    OPTIONAL_FLAGS="$OPTIONAL_FLAGS -input $INPUT_FORMAT"
fi

if [ -n "$OUTPUT_FORMAT" ]; then
    OPTIONAL_FLAGS="$OPTIONAL_FLAGS -output $OUTPUT_FORMAT"
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

# Create the output path if it doesn't exist, or clear it.
mkdir -p $OUTPUT_PATH
[ $? -ne 0 ] && exit 1
#rm -fr $OUTPUT_PATH/*
#[ $? -ne 0 ] && exit 1

# Clear any dot path for debug files.
rm -fr $OUTPUT_PATH/*/dot
[ $? -ne 0 ] && exit 1

# Run the command to generate from the example.
echo -e "\n\$GOBIN/req -rootsource $INPUT_PATH -rootoutput $OUTPUT_PATH -model $MODEL$OPTIONAL_FLAGS\n"
/go/bin/req -rootsource $INPUT_PATH -rootoutput $OUTPUT_PATH -model $MODEL $OPTIONAL_FLAGS

[ $? -ne 0 ] && exit 1

# Everything is fine.
exit 0
