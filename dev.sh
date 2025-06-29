#!/bin/bash

# Development Script for Storage Control Plane (Linux/Unix/macOS)
set -e

echo "🚀 Storage Control Plane - Development Environment"
echo "=================================================="

# Load environment variables
if [ -f ".env" ]; then
    echo "📄 Loading environment variables from .env..."
    export $(grep -v '^#' .env | xargs)
else
    echo "⚠️  No .env file found. Using defaults..."
fi

# Default values
export PORT="${PORT:-8081}"
export ROCKSDB_PATH="${ROCKSDB_PATH:-./data/rocksdb}"
export PARQUET_PATH="${PARQUET_PATH:-./data/parquet}"
export WAL_PATH="${WAL_PATH:-./data/wal}"

echo "📋 Configuration:"
echo "   Port: $PORT"
echo "   RocksDB: $ROCKSDB_PATH"
echo "   Parquet: $PARQUET_PATH"
echo "   WAL: $WAL_PATH"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    echo "💡 Please install Go: https://golang.org/dl/"
    exit 1
fi

# Check if Air is installed for hot reload
if ! command -v air &> /dev/null; then
    echo "🔧 Installing Air for hot reload..."
    go install github.com/cosmtrek/air@latest
    
    # Check if Air is now available
    if ! command -v air &> /dev/null; then
        echo "⚠️  Air not found in PATH. You may need to add $GOPATH/bin to PATH"
        echo "💡 Run: export PATH=\$PATH:\$(go env GOPATH)/bin"
        echo "💡 Or add it to your ~/.bashrc or ~/.zshrc"
    fi
fi

# Create necessary directories
echo "📁 Creating data directories..."
mkdir -p data/rocksdb data/parquet data/wal tmp

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod download

# Option selection
echo ""
echo "Choose an option:"
echo "1) 🔥 Start with Air (Hot Reload)"
echo "2) ▶️  Build and Run"
echo "3) 🧪 Run Tests"
echo "4) 🌐 Run E2E Tests"
echo "5) 🏗️  Build Only"
echo "6) 🧹 Clean Build Cache"
echo ""
read -p "Enter your choice (1-6): " choice

case $choice in
    1)
        echo "🔥 Starting with Air hot reload..."
        if command -v air &> /dev/null; then
            air
        else
            echo "❌ Air not available. Running normal build instead..."
            go run cmd/api/main.go
        fi
        ;;
    2)
        echo "▶️  Building and running..."
        go build -o tmp/main cmd/api/main.go
        ./tmp/main
        ;;
    3)
        echo "🧪 Running tests..."
        go test -v ./...
        ;;
    4)
        echo "🌐 Running E2E tests..."
        if [ -f "test_e2e.sh" ]; then
            chmod +x test_e2e.sh
            ./test_e2e.sh
        else
            echo "❌ test_e2e.sh not found"
            exit 1
        fi
        ;;
    5)
        echo "🏗️  Building only..."
        go build -o tmp/main cmd/api/main.go
        echo "✅ Build complete: tmp/main"
        ;;
    6)
        echo "🧹 Cleaning build cache..."
        go clean -cache -modcache -testcache
        rm -rf tmp/*
        echo "✅ Cache cleaned"
        ;;
    *)
        echo "❌ Invalid choice"
        exit 1
        ;;
esac

echo ""
echo "✅ Done!"
