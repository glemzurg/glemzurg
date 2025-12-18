#!/bin/bash
SCRIPT_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export GOPATH=$SCRIPT_PATH/..
export GOBIN=$SCRIPT_PATH/../bin

# go get -d code.google.com/p/go.tools
# go get -d golang.org/x/tools/blog
# go install code.google.com/p/go.tools/cmd/godoc

# godoc installed to /usr/local/bin/godoc
echo http://localhost:6060/pkg/github.com/glemzurg/glemzurg/apps/requirements/reqmd/
godoc -http=:6060