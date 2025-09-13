#!/bin/bash
set -e

# This script generates Go code from the Protobuf definitions.
# It requires protoc, protoc-gen-go, and protoc-gen-go-grpc to be installed.

# Get the root directory of the project
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
cd "$DIR"

echo "Generating Go code from Protobuf definitions..."

# Define the source .proto file
PROTO_FILE="api/protobuf/mgmt/v1/management.proto"

# Define the output directory
OUT_DIR="api/protobuf/mgmt/v1"

# Check if the proto file exists
if [ ! -f "$PROTO_FILE" ]; then
    echo "Error: Proto file not found at $PROTO_FILE"
    exit 1
fi

# Generate Go gRPC code
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    "$PROTO_FILE"

echo "Code generation complete for $PROTO_FILE."
echo "Generated code is in $OUT_DIR."

# Note: To generate the gRPC gateway code (for the REST API), you would add
# the --grpc-gateway_out flag and relevant options.
# Example:
# protoc --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
#   "$PROTO_FILE"
