.PHONY: all build test examples
all: build test

build:
	go install

test:
	go test ./...

check:
	tsc --strict --pretty testdata/output/defaults/*

checkwatch:
	tsc -w --strict --pretty testdata/output/defaults/*

examples:
	bash examples.sh
