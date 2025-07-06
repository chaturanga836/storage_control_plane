#!/bin/bash
# Quick fix script for common deployment issues

echo "🔧 FIXING COMMON DEPLOYMENT ISSUES"
echo "=================================="

# Fix 1: Make deployment script executable
echo "✅ Making deploy_ec2.sh executable..."
chmod +x deploy_ec2.sh

# Fix 2: Handle Git ownership if needed
if git status &>/dev/null; then
    echo "✅ Git repository is accessible"
else
    echo "🔧 Fixing Git ownership issues..."
    git config --global --add safe.directory "$(pwd)"
    
    # If still failing, try ownership fix
    if ! git status &>/dev/null; then
        echo "🔧 Attempting ownership fix..."
        chown -R $(whoami):$(whoami) .
    fi
fi

# Fix 3: Ensure Go is in PATH if installed
if command -v go &> /dev/null; then
    echo "✅ Go is accessible: $(go version)"
else
    if [ -d "/usr/local/go" ]; then
        echo "🔧 Adding Go to PATH..."
        export PATH=$PATH:/usr/local/go/bin
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
fi

# Fix 4: Check if we have required files
echo "📋 Checking required files..."
if [ -f "main.go" ]; then
    echo "✅ main.go found"
else
    echo "❌ main.go not found - are you in the right directory?"
fi

if [ -f "go.mod" ]; then
    echo "✅ go.mod found"
else
    echo "❌ go.mod not found - are you in the right directory?"
fi

echo ""
echo "🚀 Common issues fixed! Now you can run:"
echo "   ./deploy_ec2.sh"
echo ""
