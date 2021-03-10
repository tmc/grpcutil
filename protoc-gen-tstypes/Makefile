.PHONY: all build test examples
all: build test

build: gen
	go install

gen:
	go generate ./...

test:
	go test ./...

check:
	npx tsc --pretty testdata/output/defaults/*

checkall:
	npx tsc --pretty testdata/output/defaults/*

checkwatch:
	npx tsc -w --pretty testdata/output/defaults/*

examples:
	bash examples.sh
