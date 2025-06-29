#!/bin/bash

# Cross-Platform Setup Script for Storage Control Plane
# Detects OS and runs appropriate setup commands

set -e

echo "🌍 Storage Control Plane - Cross-Platform Setup"
echo "==============================================="

# Detect operating system
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
    echo "🐧 Detected: Linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"
    echo "🍎 Detected: macOS"
elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    OS="windows"
    echo "🪟 Detected: Windows"
else
    OS="unknown"
    echo "❓ Unknown OS: $OSTYPE"
fi

# Check if running in WSL
if [[ -f "/proc/version" ]] && grep -q "microsoft" /proc/version; then
    echo "📱 Running in WSL (Windows Subsystem for Linux)"
    WSL=true
else
    WSL=false
fi

echo ""

# Check prerequisites
echo "🔍 Checking prerequisites..."

# Check Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | cut -d' ' -f3)
    echo "✅ Go is installed: $GO_VERSION"
else
    echo "❌ Go is not installed"
    echo ""
    case $OS in
        "linux")
            echo "💡 Install Go on Linux:"
            echo "   sudo apt install golang-go  # Ubuntu/Debian"
            echo "   sudo dnf install golang     # Fedora/RHEL"
            echo "   sudo pacman -S go           # Arch Linux"
            ;;
        "macos")
            echo "💡 Install Go on macOS:"
            echo "   brew install go"
            echo "   # Or download from: https://golang.org/dl/"
            ;;
        "windows")
            echo "💡 Install Go on Windows:"
            echo "   choco install golang        # With Chocolatey"
            echo "   # Or download from: https://golang.org/dl/"
            ;;
    esac
    exit 1
fi

# Check Make
if command -v make &> /dev/null; then
    echo "✅ Make is available"
    HAS_MAKE=true
else
    echo "⚠️  Make is not available"
    HAS_MAKE=false
    case $OS in
        "linux")
            echo "💡 Install Make: sudo apt install build-essential"
            ;;
        "macos")
            echo "💡 Install Make: xcode-select --install"
            ;;
        "windows")
            echo "💡 Install Make: choco install make"
            ;;
    esac
fi

echo ""

# Setup project
echo "🔧 Setting up project..."

# Create directories
echo "📁 Creating directories..."
mkdir -p data/rocksdb data/parquet data/wal tmp bin

# Setup environment
if [ ! -f ".env" ]; then
    echo "📄 Creating .env from .env.example..."
    cp .env.example .env
else
    echo "📄 .env already exists"
fi

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod download

echo ""

# Install development tools
echo "🛠️  Installing development tools..."

# Air for hot reload
if command -v air &> /dev/null; then
    echo "✅ Air is already installed"
else
    echo "🔥 Installing Air..."
    go install github.com/cosmtrek/air@latest
fi

# Other tools (optional)
echo "🔧 Installing optional development tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest 2>/dev/null || echo "⚠️  golangci-lint installation skipped"
go install golang.org/x/tools/cmd/goimports@latest 2>/dev/null || echo "⚠️  goimports installation skipped"

echo ""

# Make scripts executable
echo "🔐 Making scripts executable..."
chmod +x dev.sh 2>/dev/null || echo "⚠️  dev.sh not found"
chmod +x test_e2e.sh 2>/dev/null || echo "⚠️  test_e2e.sh not found"

echo ""
echo "✅ Setup complete!"
echo ""
echo "🚀 Next steps:"

if [ "$HAS_MAKE" = true ]; then
    echo "   make dev          # Start with hot reload"
    echo "   make test         # Run tests"
    echo "   make help         # See all commands"
else
    case $OS in
        "windows")
            echo "   .\\dev.ps1        # Start development (PowerShell)"
            echo "   .\\dev.bat        # Start development (Batch)"
            ;;
        *)
            echo "   ./dev.sh          # Start development"
            echo "   go run cmd/api/main.go  # Manual start"
            ;;
    esac
fi

echo ""
echo "📚 Documentation:"
echo "   README.md         # Complete development guide"
echo "   QUICKSTART.md     # Quick setup guide"

case $OS in
    "linux")
        echo "   LINUX_SETUP.md    # Linux-specific setup"
        ;;
    "windows")
        echo "   WINDOWS_SETUP.md  # Windows-specific setup"
        ;;
esac

if [ "$WSL" = true ]; then
    echo ""
    echo "📱 WSL-specific tips:"
    echo "   - Ensure Docker Desktop WSL integration is enabled"
    echo "   - Work in Linux filesystem (~/project) for better performance"
    echo "   - Use VS Code with Remote-WSL extension"
fi

echo ""
echo "🎉 Happy coding!"
