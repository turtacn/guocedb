#!/bin/bash

# This script runs all tests for the guocedb project.

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Starting guocedb tests..."

# Get the root directory of the project
ROOT_DIR=$(git rev-parse --show-toplevel)
cd "$ROOT_DIR"

# Run tests with coverage
# The -race flag detects race conditions
# The ... wildcard runs tests in all subdirectories
go test -v -race -cover ./...

echo "Tests completed successfully!"
