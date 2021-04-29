#!/bin/bash

export GO111MODULE="on"
export GOPATH=/var/gopath
ACG_CONFIG="$(pwd)/cdappconfig.json"  go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

if [ $? != 0 ]; then
    exit 1
fi
