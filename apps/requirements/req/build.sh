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

# # Complexity linter.
# go install github.com/glemzurg/go-complexity-lint/cmd/go-complexity-lint@latest
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

# ================================================

# The core is working fine. Format.
echo -e "\nFMT\n" ; go fmt ./...
[ $? -ne 0 ] && exit 30

# ================================================

# Linting tool.
echo -e "\nLINTING\n"
[ $? -ne 0 ] && exit 40

# Run the linters on the source tree.
golangci-lint run ./...
[ $? -ne 0 ] && exit 41

# Complexity linter.
echo -e "\nCOMPLEXITY\n"
go-complexity-lint -exclude="*.generated.go,test_model.go" -nestdepth.warn=6 -nestdepth.fail=6 -cyclo.warn=14 -cyclo.fail=14 -params.warn=7 -params.fail=7 -fanout.warn=9 -fanout.fail=9 ./...
[ $? -ne 0 ] && exit 42

# ================================================

# Run unit tests.
echo -e "\nTEST\n" 
[ $? -ne 0 ] && exit 50

# Run all tests (except slow database tests).
go test ./...
[ $? -ne 0 ] && exit 51

# Run slow database tests
go test ./internal/database/... -dbtests
[ $? -ne 0 ] && exit 52

# ================================================

# Build and install any executables.
echo -e "\nINSTALL\n" ; go install ./...
[ $? -ne 0 ] && exit 60

# Indicate the command.
echo -e "\nLAUNCH COMMAND\n"
# echo -e "\$GOBIN/reqmodel -rootsource example/models -rootoutput example/output/models -model model_a -plantuml /usr/bin/plantuml\n"
# echo -e "\$GOBIN/reqmodelhttp -rootsource example/models -rootmd example/output/models -port 8087 -plantuml /usr/bin/plantuml -nodb\n"

# Everything is fine.
exit 0