#!/bin/bash
SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# We may have a test we want to run.
TEST_TO_RUN="$1"

# We are in the script path directory.
cd $SCRIPT_PATH
[ $? -ne 0 ] && exit 1

# ================================================

# Uncomment to get tools.

# # Get all the libraries we need.
# echo -e "\nGET\n"

# # Testing libraries.
# go get github.com/smartystreets/goconvey
# [ $? -ne 0 ] && exit 1
# go get github.com/stretchr/testify
# [ $? -ne 0 ] && exit 1

# # Get the linter.
# go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
# [ $? -ne 0 ] && exit 1

# # Get the PEG parser.
# go install github.com/mna/pigeon@latest
# [ $? -ne 0 ] && exit 1

# # SQL library.
# go get -d "github.com/lib/pq"
# [ $? -ne 0 ] && exit 1

# # Any imported libraries.
# go get "glemzurg..."
# [ $? -ne 0 ] && exit 1

# ================================================

# # Generate the parsing logic.
# pigeon -o parser/pigeon_parser/temp.txt parser/peg/file.peg
# [ $? -ne 0 ] && exit 1

# cat parser/peg/package_header.txt parser/pigeon_parser/temp.txt > parser/pigeon_parser/file.generated.go 
# [ $? -ne 0 ] && exit 1

# rm -fr parser/pigeon_parser/temp.txt 
# [ $? -ne 0 ] && exit 1

# Run unit tests.
echo -e "\nTEST\n" 
if [ -z "$TEST_TO_RUN" ]; then

  # No explicit test, running all tests.
  go test -count=1 -p=1 ./...
  [ $? -ne 0 ] && exit 1

  # The core is working fine. Format.
  echo -e "\nFMT\n" ; go fmt ./...
  [ $? -ne 0 ] && exit 1

else 

  # An explicit test, run only that.
  go test -count=1 -p=1 -v ./... -run "$TEST_TO_RUN"
  [ $? -ne 0 ] && exit 1

fi 

# Build and install any executables.
echo -e "\nINSTALL\n" ; go install ./...
[ $? -ne 0 ] && exit 1

# Setting up default data.
# echo -e "\nPOPULATING DATABASE\n"
# mysql -uroot psp_test < $GOPATH/../test_files/psp_test.sql
# [ $? -ne 0 ] && exit 1

# Indicate the command.
echo -e "\nLAUNCH COMMAND\n"
echo -e "\$GOBIN/reqmodel -rootsource example/models -rootoutput example/output/models -model model_a -plantuml /usr/bin/plantuml\n"
echo -e "\$GOBIN/reqmodelhttp -rootsource example/models -rootmd example/output/models -port 8087 -plantuml /usr/bin/plantuml -nodb\n"

# Linting tool.
echo -e "\nLINTING\n"

# Run the linters on the source tree.
golangci-lint run ./...
[ $? -ne 0 ] && exit 1

# Everything is fine.
exit 0