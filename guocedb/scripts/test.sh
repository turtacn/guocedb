#!/bin/bash
set -e

# This script runs all tests for the guocedb project.

# Get the root directory of the project
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
cd "$DIR"

echo "Running tests..."

# Tidy modules first to ensure all dependencies are present.
go mod tidy

# Run tests with coverage
# The -race flag detects race conditions.
# The -v flag provides verbose output.
go test -v -race -coverprofile=coverage.out ./...

# Display coverage summary
echo ""
echo "================="
echo "Test Coverage:"
echo "================="
go tool cover -func=coverage.out

echo ""
echo "Tests passed."
