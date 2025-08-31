#!/bin/bash

# This script builds the guocedb server and cli binaries.

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Starting guocedb build..."

# Get the root directory of the project
ROOT_DIR=$(git rev-parse --show-toplevel)
cd "$ROOT_DIR"

# Create the output directory if it doesn't exist
OUTPUT_DIR="$ROOT_DIR/bin"
mkdir -p "$OUTPUT_DIR"

# Build the server
echo "Building guocedb-server..."
go build -o "$OUTPUT_DIR/guocedb-server" ./cmd/guocedb-server

# Build the CLI
echo "Building guocedb-cli..."
go build -o "$OUTPUT_DIR/guocedb-cli" ./cmd/guocedb-cli

echo "Build successful!"
echo "Binaries are located in $OUTPUT_DIR"
ls -l "$OUTPUT_DIR"
