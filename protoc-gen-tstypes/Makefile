.PHONY: all build test examples
all: build test

build: gen
	go install

gen:
	go generate ./...

test:
	go test ./...

check:
	npx typescript --pretty testdata/output/defaults/*

checkall:
	npx typescript --pretty testdata/output/defaults/*

checkwatch:
	npx typescript -w --pretty testdata/output/defaults/*

examples:
	bash examples.sh
