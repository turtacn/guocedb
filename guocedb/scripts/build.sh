#!/bin/bash
set -e

# This script builds the guocedb server and cli binaries.

# Get the root directory of the project
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

echo "Building guocedb binaries..."
cd "$DIR"

# Create the output directory if it doesn't exist
mkdir -p ./bin

# Build server
echo "Building guocedb-server..."
go build -o ./bin/guocedb-server ./cmd/guocedb-server

# Build CLI
echo "Building guocedb-cli..."
go build -o ./bin/guocedb-cli ./cmd/guocedb-cli

echo "Build complete. Binaries are in the ./bin directory."
