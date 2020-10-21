PROJECT_ROOT := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
ACG_CONFIG := $(PROJECT_ROOT)/cdappconfig.json

dep:
	go get -t -v ./...

# Run tests with ACG at the top level. Tests are at different depths.
test: dep
	ACG_CONFIG=$(ACG_CONFIG) go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...