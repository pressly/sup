.PHONY: help build build_pkgs test install clean

help:
	@echo "build:   Build code."
	@echo "test:    Run tests."
	@echo "install: Install binary."
	@echo "clean:   Clean up."

build: build_pkgs
	@mkdir -p ./bin
	@rm -f ./bin/*
	go build -o ./bin/sup ./cmd/sup

build_pkgs:
	go build ./...

test:
	go test

install: build
	go install ./...

clean:
	@rm -rf ./bin

deps:
	@glock sync -n github.com/pressly/sup < Glockfile

update_deps:
	@glock save -n github.com/pressly/sup > Glockfile
