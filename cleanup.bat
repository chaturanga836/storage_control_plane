@echo off
REM Cleanup script for storage_control_plane Go repository (Windows version)
REM This script cleans up the existing code and prepares for a fresh microservices-based implementation

echo 🧹 Starting cleanup of storage_control_plane repository...

REM 1. Backup existing important files
echo 📋 Backing up important configuration files...
if not exist .backup mkdir .backup
if exist go.mod copy go.mod .backup\ >nul 2>&1
if exist go.sum copy go.sum .backup\ >nul 2>&1
if exist .env copy .env .backup\ >nul 2>&1
if exist README.md copy README.md .backup\ >nul 2>&1

REM 2. Clean up old code directories (but preserve structure)
echo 🗑️  Cleaning up old code...
if exist cmd rmdir /s /q cmd >nul 2>&1
if exist internal rmdir /s /q internal >nul 2>&1
if exist pkg rmdir /s /q pkg >nul 2>&1
if exist api rmdir /s /q api >nul 2>&1

REM 3. Remove old documentation (we'll create new ones)
echo 📚 Cleaning up old documentation...
del /q CROSS_FILE_QUERY_SOLUTION.md >nul 2>&1
del /q DATA_LAKE_COMPARISON.md >nul 2>&1
del /q DISTRIBUTED_INDEXING_GUIDE.md >nul 2>&1
del /q HORIZONTAL_INDEXING_SUMMARY.md >nul 2>&1
del /q LINUX_SETUP.md >nul 2>&1
del /q OPTIMIZATION_SUMMARY.md >nul 2>&1
del /q QUERY_EXAMPLES.md >nul 2>&1
del /q QUERY_FLOW_GUIDE.md >nul 2>&1
del /q README_MULTILANG.md >nul 2>&1
del /q ROADMAP.md >nul 2>&1
del /q TESTING.md >nul 2>&1
del /q WINDOWS_SETUP.md >nul 2>&1
del /q NEXT_STEPS.md >nul 2>&1
del /q AIR_SETUP.md >nul 2>&1
del /q DOCUMENTATION_INDEX.md >nul 2>&1

REM 4. Clean up old build artifacts and temporary files
echo 🧽 Removing build artifacts...
del /q api.exe >nul 2>&1
if exist tmp rmdir /s /q tmp >nul 2>&1
if exist data rmdir /s /q data >nul 2>&1
del /q *.log >nul 2>&1
del /q *.tmp >nul 2>&1

REM 5. Clean up old scripts (we'll create new ones)
echo 🔧 Cleaning up old scripts...
del /q dev.bat >nul 2>&1
del /q dev.ps1 >nul 2>&1
del /q dev.sh >nul 2>&1
del /q startup.bat >nul 2>&1
del /q startup.sh >nul 2>&1
del /q setup.sh >nul 2>&1
del /q test_e2e.ps1 >nul 2>&1
del /q test_e2e.sh >nul 2>&1

REM 6. Remove old test files (we'll create new ones)
echo 🧪 Cleaning up old tests...
if exist test rmdir /s /q test >nul 2>&1
if exist examples rmdir /s /q examples >nul 2>&1

REM 7. Remove old SQL files (we'll create new ones if needed)
echo 🗄️  Cleaning up SQL files...
if exist sql rmdir /s /q sql >nul 2>&1

REM 8. Remove old tenant_node (Python remnant)
echo 🐍 Removing Python remnants...
if exist tenant_node rmdir /s /q tenant_node >nul 2>&1

REM 9. Clean up environment files (keep examples)
echo 🌍 Cleaning up environment files...
del /q .env.test >nul 2>&1
del /q .env.test.example >nul 2>&1

echo ✅ Cleanup completed!
echo.
echo 📁 Current directory structure:
dir /b

echo.
echo 🎯 Next steps:
echo 1. Run the setup script to create new Go structure
echo 2. Initialize new microservices architecture
echo 3. Implement services one by one
echo.
echo 💡 Backed up files are in .backup\ directory

pause
