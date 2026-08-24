#!/usr/bin/env bash
set -e

echo "📦 Building Nova Desktop Application..."
mkdir -p bin

# Build with optimization and symbol stripping
go build -ldflags="-s -w" -o bin/desktop_app .

echo "✅ Build Successful!"
echo "🚀 Run binary: ./bin/desktop_app"
