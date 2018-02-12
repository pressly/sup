.PHONY: all build dist test install clean tools deps update-deps

all:
	@echo "build         - Build sup"
	@echo "dist          - Build sup distribution binaries"
	@echo "test          - Run tests"
	@echo "install       - Install binary"
	@echo "clean         - Clean up"
	@echo ""
	@echo "tools         - Install tools"
	@echo "vendor-list   - List vendor package tree"
	@echo "vendor-update - Update vendored packages"

build:
	@mkdir -p ./bin
	@rm -f ./bin/*
	go build -o ./bin/sup ./cmd/sup

dist:
	@mkdir -p ./bin
	@rm -f ./bin/*
	GOOS=darwin GOARCH=amd64 go build -o ./bin/sup-darwin64 ./cmd/sup
	GOOS=linux GOARCH=amd64 go build -o ./bin/sup-linux64 ./cmd/sup
	GOOS=linux GOARCH=386 go build -o ./bin/sup-linux386 ./cmd/sup
	GOOS=windows GOARCH=amd64 go build -o ./bin/sup-windows64.exe ./cmd/sup
	GOOS=windows GOARCH=386 go build -o ./bin/sup-windows386.exe ./cmd/sup

test:
	go test ./...

install:
	go install ./cmd/sup

clean:
	@rm -rf ./bin

tools:
	go get -u github.com/kardianos/govendor

vendor-list:
	@govendor list

vendor-update:
	@govendor update +external
