.PHONY: build install test

build:
	go build -o bin/marc .

install: build
	mkdir -p $(HOME)/.local/bin
	cp bin/marc $(HOME)/.local/bin/marc

test:
	go test ./...
