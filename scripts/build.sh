#!/bin/bash

# scripts/build.sh
# A simple script to build the Guocedb server and client binaries.
# 一个简单的脚本，用于构建 Guocedb 服务器和客户端二进制文件。

# Exit immediately if a command exits with a non-zero status.
# 如果命令以非零状态退出，则立即退出。
set -euo pipefail

# Define the output directory for binaries.
# 定义二进制文件的输出目录。
BIN_DIR="./bin"

# Define the paths to the server and client main packages.
# 定义服务器和客户端主包的路径。
SERVER_PATH="./cmd/guocedb_server"
CLI_PATH="./cmd/guocedb_cli"

# Create the output directory if it doesn't exist.
# 如果输出目录不存在，则创建它。
mkdir -p "$BIN_DIR"
echo "Created output directory: $BIN_DIR" # 创建输出目录。

# Enable CGO if necessary for certain dependencies (like sqlite, although not used directly yet).
# If not needed, this line can be removed.
# 如果某些依赖项需要 CGO（例如 sqlite，尽管尚未直接使用），则启用 CGO。
# 如果不需要，可以删除此行。
# export CGO_ENABLED=1

echo "Building Guocedb server..." # 正在构建 Guocedb 服务器...
# Build the server binary.
# 构建服务器二进制文件。
# go build -o <output_path> <package_path>
go build -o "$BIN_DIR/guocedb_server" "$SERVER_PATH"
echo "Guocedb server built successfully: $BIN_DIR/guocedb_server" # Guocedb 服务器构建成功。

echo "Building Guocedb CLI..." # 正在构建 Guocedb CLI...
# Build the client binary.
# 构建客户端二进制文件。
go build -o "$BIN_DIR/guocedb_cli" "$CLI_PATH"
echo "Guocedb CLI built successfully: $BIN_DIR/guocedb_cli" # Guocedb CLI 构建成功。

echo "Build process completed." # 构建过程完成。