#!/bin/bash
# Example: ./scripts/req_examples.sh /data/examples/requirements/req/models/ /data/examples/requirements/req/output/ web_books -debug
SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Output path.
INPUT_PATH="$1"
OUTPUT_PATH="$2"
MODEL="$3"

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

# We may have a test we want to run.
DEBUG="$4"

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
echo -e "\$GOBIN/reqmodel -rootsource $INPUT_PATH -rootoutput $OUTPUT_PATH -model $MODEL -debug\n"
/go/bin/reqmodel -rootsource $INPUT_PATH -rootoutput $OUTPUT_PATH -model $MODEL $DEBUG

[ $? -ne 0 ] && exit 1

# Everything is fine.
exit 0