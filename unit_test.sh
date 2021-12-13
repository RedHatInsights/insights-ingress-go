#!/bin/bash


export GO111MODULE="on"
export GOPATH="$(pwd)/gopath"
export GOROOT="/opt/go/1.16.10"
export PATH="${GOROOT}/bin:${PATH}"
ACG_CONFIG="$(pwd)/cdappconfig.json"  go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

if [ $? != 0 ]; then
    exit 1
fi
