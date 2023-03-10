GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

test:
	go test -v -race ./...

cover:
	go test -v -coverprofile=coverage.txt -covermode=atomic ./...

dep:
	go get -u ./...
	go get -t ./...