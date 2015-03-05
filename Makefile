# Old-skool build tools.

help:
	@echo "build:   Build code."
	@echo "test:    Run tests."
	@echo "install: Install binary."
	@echo "clean:   Clean up."
.PHONY: help

all: build test
.PHONY: all

build: build_pkgs
	@mkdir -p ./bin
	@rm -f ./bin/*
	go build -o ./bin/sup ./cmd/sup
.PHONY: build

build_pkgs:
	go build ./...
.PHONY: build_pkgs

test:
	@go test ./... | grep -v "no test files" 
.PHONY: test

install: build
	go install ./...
.PHONY: install

clean:
	@rm -rf ./bin
.PHONY: clean