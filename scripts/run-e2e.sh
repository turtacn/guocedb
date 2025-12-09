#!/bin/bash
set -e

echo "=== GuoceDB E2E Tests ==="

# Change to project root
cd "$(dirname "$0")/.."

# Build the project
echo "Building..."
make build || { echo "Build failed"; exit 1; }

# Run integration tests
echo "Running integration tests..."
go test -v ./integration/e2e_*.go -count=1 -timeout=10m || {
    echo "E2E tests failed"
    exit 1
}

echo "=== All E2E tests passed ==="
