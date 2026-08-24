#!/usr/bin/env bash
set -e

echo "📦 Compiling Nova IP Scanner standalone desktop application..."
mkdir -p bin
go build -ldflags="-s -w" -o bin/ip_scanner .

echo "✅ Build Successful!"
echo "🚀 Run binary: ./bin/ip_scanner"
