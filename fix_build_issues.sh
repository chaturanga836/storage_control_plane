#!/bin/bash
# Quick fix for Go build issues

echo "🔧 FIXING GO BUILD ISSUES"
echo "========================="

# Fix 1: Rename test files so they don't interfere with main build
if [ -f "functional_test_suite.go" ]; then
    echo "✅ Renaming functional_test_suite.go to functional_test_suite_test.go"
    mv functional_test_suite.go functional_test_suite_test.go
fi

# Fix 2: Try building with specific files only
echo "🔨 Attempting targeted build..."
if go build -o storage-control-plane main.go services.go services_extended.go; then
    echo "✅ Build successful with specific files!"
    ls -la storage-control-plane
    exit 0
fi

# Fix 3: Try building all non-test files
echo "🔨 Attempting build excluding test files..."
if go build -o storage-control-plane $(ls *.go | grep -v '_test.go'); then
    echo "✅ Build successful excluding test files!"
    ls -la storage-control-plane
    exit 0
fi

# Fix 4: Clean and try again
echo "🧹 Cleaning and retrying..."
go clean -cache
go mod tidy

if go build -o storage-control-plane .; then
    echo "✅ Build successful after cleanup!"
    ls -la storage-control-plane
else
    echo "❌ Build still failing. Manual intervention needed."
    echo ""
    echo "Debug information:"
    echo "Go version: $(go version)"
    echo "Go files in directory:"
    ls -la *.go
    echo ""
    echo "Try running: go build -v -o storage-control-plane main.go services.go services_extended.go"
fi
