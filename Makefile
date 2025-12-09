# GuoceDB Makefile

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -X 'main.Version=$(VERSION)' \
           -X 'main.GitCommit=$(GIT_COMMIT)' \
           -X 'main.BuildTime=$(BUILD_TIME)'

.PHONY: all build test clean install lint fmt run docker help

all: build

## Build targets
build:  ## Build the GuoceDB binary
	@echo "Building GuoceDB $(VERSION)..."
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/guocedb ./cmd/guocedb

build-cli: build  ## Alias for build (backward compatibility)

build-server: build  ## Alias for build (backward compatibility)

build-static:  ## Build static binaries (no CGO)
	@echo "Building static binaries..."
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb ./cmd/guocedb

## Test targets
test:  ## Run all tests
	go test -v -race -cover ./...

test-unit:  ## Run unit tests only
	go test -v -short ./...

test-integration:  ## Run integration tests
	go test -v ./integration/... -count=1

test-e2e:  ## Run E2E tests
	./scripts/run-e2e.sh

test-config:  ## Run config tests only
	go test -v -race -cover ./config/...

test-server:  ## Run server tests only
	go test -v -race -cover ./server/...

test-short:  ## Run short tests only
	go test -short ./...

test-cover:  ## Generate test coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage: test-cover  ## Alias for test-cover

bench:  ## Run benchmarks
	go test -bench=. -benchmem ./benchmark/...

bench-full:  ## Run full benchmark suite
	./scripts/benchmark.sh

## Development targets
clean:  ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

install: build  ## Install binary to GOPATH/bin
	@echo "Installing guocedb..."
	go install -ldflags "$(LDFLAGS)" ./cmd/guocedb

lint:  ## Run linter (if golangci-lint is installed)
	@which golangci-lint > /dev/null && golangci-lint run ./... || echo "golangci-lint not installed, skipping..."

fmt:  ## Format code
	go fmt ./...

vet:  ## Run go vet
	go vet ./...

## Run targets
run:  ## Run server with default config
	go run ./cmd/guocedb --port 3306 --data-dir ./data

run-check:  ## Run config check
	go run ./cmd/guocedb check

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
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb-linux-amd64 ./cmd/guocedb

release-darwin:  ## Build macOS release binaries
	@mkdir -p bin
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb-darwin-amd64 ./cmd/guocedb

release-windows:  ## Build Windows release binaries
	@mkdir -p bin
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS) -s -w" -o bin/guocedb-windows-amd64.exe ./cmd/guocedb

release-all: release-linux release-darwin release-windows  ## Build all release binaries

## Help
help:  ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)