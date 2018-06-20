.PHONY: all build test examples
all: build test

build:
	go install

test:
	go test ./...

example:
	protoc -I. -I$(GOPATH)/src --tstypes_out=. testdata/example1.proto

examples:
	bash examples.sh
