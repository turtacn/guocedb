#!/bin/bash
set -e

# This script runs the linter on the guocedb project.

# Get the root directory of the project
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
cd "$DIR"

# Check for golangci-lint and install if not present
if ! command -v golangci-lint &> /dev/null
then
    echo "golangci-lint not found, installing..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

echo "Running linter..."
golangci-lint run ./...
