#!/bin/bash
SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $SCRIPT_PATH
[ $? -ne 0 ] && exit 1

echo -e "\nFMT\n" ; go fmt ./...
[ $? -ne 0 ] && exit 30

echo -e "\nLINTING\n"
golangci-lint run ./...
[ $? -ne 0 ] && exit 41

echo -e "\nCOMPLEXITY\n"
go-complexity-lint -warnings=none -exclude="*.generated.go" ./...
[ $? -ne 0 ] && exit 42

echo -e "\nTEST\n"
go test ./...
[ $? -ne 0 ] && exit 51

echo -e "\nINSTALL\n" ; go install -buildvcs=false ./...
[ $? -ne 0 ] && exit 60

exit 0
