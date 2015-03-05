# Old-skool build tools.

help:
	@echo "build:   Build code."
	@echo "test:    Run tests."
	@echo "install: Install binary."
	@echo "clean:   Clean up."
.PHONY: help

all: build test
.PHONY: all

build:
	go build ./...
.PHONY: build

test:
	go test ./... | grep -v "no test files" 
.PHONY: test

install: build
	go install ./...
.PHONY: install

