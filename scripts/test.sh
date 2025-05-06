#!/bin/bash

# scripts/test.sh
# A simple script to run the Guocedb tests.
# 一个简单的脚本，用于运行 Guocedb 测试。

# Exit immediately if a command exits with a non-zero status.
# 如果命令以非零状态退出，则立即退出。
set -euo pipefail

echo "Running Guocedb tests..." # 正在运行 Guocedb 测试...

# Run all tests in the project.
# The -v flag provides verbose output, showing each test function.
# The ./... pattern matches all packages in the current directory and its subdirectories.
#
# 运行项目中的所有测试。
# -v 标志提供详细输出，显示每个测试函数。
# ./... 模式匹配当前目录及其子目录中的所有包。
go test -v ./...

# Note: For more complex test setups (e.g., integration tests requiring a running server),
# you might need to start services before running integration tests and stop them afterwards.
#
# 注意：对于更复杂的测试设置（例如，需要运行服务器的集成测试），
# 你可能需要在运行集成测试之前启动服务并在之后停止它们。
# You might also use Go test flags like -run, -short, or build tags to control which tests run.
# 你还可以使用 Go 测试标志，如 -run、-short 或构建标签来控制运行哪些测试。
# Example to run only unit tests (assuming unit tests are in packages ending with /unit):
# go test -v ./... -run Unit
# Example to run tests with a specific tag (requires adding // +build <tag_name> to files):
# go test -v -tags integration ./...

echo "Test process completed." # 测试过程完成。