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

  # # SQL library.
  # # go get -d "github.com/go-sql-driver/mysql"
  # # [ $? -ne 0 ] && exit 1

  # # Any imported libraries.
  # go get -d "glemzurg..."
  # [ $? -ne 0 ] && exit 1

# ================================================

# Update the schema in the test database.
# mysql -uroot psp_test < $SCRIPT_PATH/glemzurg/psp/sql/schema.sql 
# [ $? -ne 0 ] && exit 1

# Run unit tests.
echo -e "\nTEST\n" 
if [ -z "$TEST_TO_RUN" ]; then

  # No explicit test, running all tests.
  go test -p=1 -v ./...
  [ $? -ne 0 ] && exit 1

  # The core is working fine. Format.
  echo -e "\nFMT\n" ; go fmt ./...
  [ $? -ne 0 ] && exit 1

else 

  # An explicit test, run only that.
  go test -p=1 -v ./... -run "$TEST_TO_RUN"
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
echo -e "\$GOBIN/reqmd -config example/design/config.json -path example/design/requirements\n"

# Linting tool.
echo -e "\nLINTING\n"

# Run the linters on the source tree.
golangci-lint run ./...
[ $? -ne 0 ] && exit 1


# Everything is fine.
exit 0