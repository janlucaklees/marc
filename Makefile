.PHONY: build install test

build:
	go build -o bin/marc ./cmd/marc

install: build
	go install ./cmd/marc

test:
	go test ./...
