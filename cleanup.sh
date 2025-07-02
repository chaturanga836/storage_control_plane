#!/bin/bash

# Cleanup script for storage_control_plane Go repository
# This script cleans up the existing code and prepares for a fresh microservices-based implementation

echo "🧹 Starting cleanup of storage_control_plane repository..."

# 1. Backup existing important files
echo "📋 Backing up important configuration files..."
mkdir -p .backup
cp go.mod .backup/ 2>/dev/null || echo "No go.mod to backup"
cp go.sum .backup/ 2>/dev/null || echo "No go.sum to backup"
cp .env .backup/ 2>/dev/null || echo "No .env to backup"
cp README.md .backup/ 2>/dev/null || echo "No README.md to backup"

# 2. Clean up old code directories (but preserve structure)
echo "🗑️  Cleaning up old code..."
rm -rf cmd/* 2>/dev/null || echo "cmd directory already clean"
rm -rf internal/* 2>/dev/null || echo "internal directory already clean"
rm -rf pkg/* 2>/dev/null || echo "pkg directory already clean"
rm -rf api/* 2>/dev/null || echo "api directory already clean"

# 3. Remove old documentation (we'll create new ones)
echo "📚 Cleaning up old documentation..."
rm -f CROSS_FILE_QUERY_SOLUTION.md
rm -f DATA_LAKE_COMPARISON.md  
rm -f DISTRIBUTED_INDEXING_GUIDE.md
rm -f HORIZONTAL_INDEXING_SUMMARY.md
rm -f LINUX_SETUP.md
rm -f OPTIMIZATION_SUMMARY.md
rm -f QUERY_EXAMPLES.md
rm -f QUERY_FLOW_GUIDE.md
rm -f README_MULTILANG.md
rm -f ROADMAP.md
rm -f TESTING.md
rm -f WINDOWS_SETUP.md
rm -f NEXT_STEPS.md
rm -f AIR_SETUP.md
rm -f DOCUMENTATION_INDEX.md

# 4. Clean up old build artifacts and temporary files
echo "🧽 Removing build artifacts..."
rm -f api.exe
rm -rf tmp/
rm -rf data/
rm -rf test/fixtures/data/
find . -name "*.log" -delete 2>/dev/null || true
find . -name "*.tmp" -delete 2>/dev/null || true

# 5. Clean up old scripts (we'll create new ones)
echo "🔧 Cleaning up old scripts..."
rm -f dev.bat
rm -f dev.ps1
rm -f dev.sh
rm -f startup.bat
rm -f startup.sh
rm -f setup.sh
rm -f test_e2e.ps1
rm -f test_e2e.sh

# 6. Remove old test files (we'll create new ones)
echo "🧪 Cleaning up old tests..."
rm -rf test/
rm -rf examples/

# 7. Remove old SQL files (we'll create new ones if needed)
echo "🗄️  Cleaning up SQL files..."
rm -rf sql/

# 8. Remove old tenant_node (Python remnant)
echo "🐍 Removing Python remnants..."
rm -rf tenant_node/

# 9. Clean up environment files (keep examples)
echo "🌍 Cleaning up environment files..."
rm -f .env.test
rm -f .env.test.example

# 10. Reset Git history (optional - uncomment if you want a fresh start)
# echo "🔄 Resetting Git history..."
# rm -rf .git
# git init
# git add .
# git commit -m "Initial commit: Clean slate for Go microservices implementation"

echo "✅ Cleanup completed!"
echo ""
echo "📁 Current directory structure:"
ls -la

echo ""
echo "🎯 Next steps:"
echo "1. Run the setup script to create new Go structure"
echo "2. Initialize new microservices architecture"
echo "3. Implement services one by one"
echo ""
echo "💡 Backed up files are in .backup/ directory"
