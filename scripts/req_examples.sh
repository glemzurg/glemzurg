#!/bin/bash
SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# We may have a test we want to run.
DEBUG="$1"

# We are in the script path directory.
cd $SCRIPT_PATH
[ $? -ne 0 ] && exit 1

# ================================================

# For easy iteration, rebuild the command line tool.
echo -e "\nUPDATE INSTALL\n"
(cd $SCRIPT_PATH/../apps/requirements/req && go install ./...)
[ $? -ne 0 ] && exit 1

# ================================================

# Output path.
INPUT_PATH="$SCRIPT_PATH/../examples/requirements/req/models"
OUTPUT_PATH="$SCRIPT_PATH/../examples/requirements/req/output"
MODEL="web_books"

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
$GOBIN/reqmodel -rootsource $INPUT_PATH -rootoutput $OUTPUT_PATH -model $MODEL $DEBUG

[ $? -ne 0 ] && exit 1

# Everything is fine.
exit 0