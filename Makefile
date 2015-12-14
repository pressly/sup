.PHONY: all build dist test install clean tools deps update-deps

all:
	@echo "build:      Build code."
	@echo "test:       Run tests."
	@echo "install:    Install binary."
	@echo ""
	@echo "tools       Install tools."
	@echo "deps        Install dependencies."
	@echo "update-deps Update dependencies."
	@echo "clean:      Clean up."

build:
	@mkdir -p ./bin
	@rm -f ./bin/*
	go build -o ./bin/sup ./cmd/sup

dist:
	@mkdir -p ./bin
	@rm -f ./bin/*
	GOOS=darwin GOARCH=amd64 go build -o ./bin/sup-darwin64 ./cmd/sup
	GOOS=linux GOARCH=amd64 go build -o ./bin/sup-linux64 ./cmd/sup
	tar -czf ./bin/sup-linux64.tar.gz ./bin/sup-linux64
	tar -czf ./bin/sup-darwin64.tar.gz ./bin/sup-darwin64

test:
	go test ./...

install: build
	go install ./...

clean:
	@rm -rf ./bin

tools:
	go get -u github.com/pressly/glock

deps:
	@glock sync -n github.com/pressly/sup < Glockfile

update-deps:
	@glock save -n github.com/pressly/sup > Glockfile
