# GuoceDB Makefile

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -X 'github.com/turtacn/guocedb/cli/commands.Version=$(VERSION)' \
           -X 'github.com/turtacn/guocedb/cli/commands.GitCommit=$(GIT_COMMIT)' \
           -X 'github.com/turtacn/guocedb/cli/commands.BuildTime=$(BUILD_TIME)'

.PHONY: all build build-cli build-server test clean install lint fmt run docker help

all: build

## Build targets
build: build-cli build-server  ## Build both CLI and server binaries

build-cli:  ## Build the CLI binary
	@echo "Building GuoceDB CLI..."
	go build -ldflags "$(LDFLAGS)" -o bin/guocedb ./cli

build-server:  ## Build the server binary
	@echo "Building GuoceDB Server..."
	go build -ldflags "$(LDFLAGS)" -o bin/guocedb-server ./cmd/guocedb-server

build-static:  ## Build static binaries (no CGO)
	@echo "Building static binaries..."
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb ./cli
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb-server ./cmd/guocedb-server

## Test targets
test:  ## Run all tests
	go test -race -cover ./...

test-cli:  ## Run CLI tests only
	go test -race -cover ./cli/... ./internal/...

test-integration:  ## Run integration tests
	go test -race -v ./integration/...

test-short:  ## Run short tests only
	go test -short ./...

coverage:  ## Generate test coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Development targets
clean:  ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

install: build-cli  ## Install CLI binary to GOPATH/bin
	cp bin/guocedb $(shell go env GOPATH)/bin/

lint:  ## Run linter
	golangci-lint run ./...

fmt:  ## Format code
	go fmt ./...

## Run targets
run:  ## Run server with default config
	go run ./cli serve --port 3306 --data-dir ./data

run-cli:  ## Run CLI with example command
	go run ./cli version

## Docker targets
docker:  ## Build Docker image
	docker build -t guocedb:$(VERSION) .

docker-run:  ## Run Docker container
	docker run -p 3306:3306 -v $(PWD)/data:/data guocedb:$(VERSION)

## Utility targets
deps:  ## Download dependencies
	go mod download
	go mod tidy

update-deps:  ## Update dependencies
	go get -u ./...
	go mod tidy

vendor:  ## Vendor dependencies
	go mod vendor

## Release targets
release-linux:  ## Build Linux release binaries
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb-linux-amd64 ./cli
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb-server-linux-amd64 ./cmd/guocedb-server

release-darwin:  ## Build macOS release binaries
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb-darwin-amd64 ./cli
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb-server-darwin-amd64 ./cmd/guocedb-server

release-windows:  ## Build Windows release binaries
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb-windows-amd64.exe ./cli
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb-server-windows-amd64.exe ./cmd/guocedb-server

release-all: release-linux release-darwin release-windows  ## Build all release binaries

## Help
help:  ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)