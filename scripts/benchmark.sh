#!/bin/bash
set -e

echo "=== GuoceDB Benchmark ==="

# Change to project root
cd "$(dirname "$0")/.."

# Build the project
echo "Building..."
make build || { echo "Build failed"; exit 1; }

# Run benchmarks
echo "Running benchmarks..."
go test -bench=. -benchmem -benchtime=3s -count=3 ./benchmark/... | tee benchmark_results.txt

echo ""
echo "=== Benchmark Results ==="
echo "Results saved to benchmark_results.txt"
