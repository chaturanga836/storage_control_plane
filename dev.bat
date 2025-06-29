@echo off
REM Windows batch file for development tasks

if "%1"=="help" goto help
if "%1"=="test" goto test
if "%1"=="test-unit" goto test_unit
if "%1"=="test-e2e" goto test_e2e
if "%1"=="build" goto build
if "%1"=="run" goto run
if "%1"=="dev" goto dev
if "%1"=="clean" goto clean
if "%1"=="fmt" goto fmt
if "%1"=="coverage" goto coverage
if "%1"=="" goto help

:help
echo Available commands:
echo   dev.bat test          - Run all tests
echo   dev.bat test-unit     - Run unit tests
echo   dev.bat test-e2e      - Run end-to-end tests
echo   dev.bat build         - Build the application
echo   dev.bat run           - Run the application
echo   dev.bat dev           - Run with hot reload (air)
echo   dev.bat clean         - Clean build artifacts
echo   dev.bat fmt           - Format code
echo   dev.bat coverage      - Generate coverage report
goto end

:test
echo 🧪 Running all tests...
go test ./internal/... ./pkg/... -v
goto end

:test_unit
echo 🧪 Running unit tests...
go test ./internal/... ./pkg/... -v
goto end

:test_e2e
echo 🌐 Running end-to-end tests...
echo ⚠️  Make sure the server is running on :8081
powershell -File test_e2e.ps1
goto end

:build
echo 🔨 Building Storage Control Plane...
go build -o bin\storage-control-plane.exe .\cmd\api
goto end

:run
echo 🚀 Starting Storage Control Plane...
go run .\cmd\api
goto end

:dev
echo 🔥 Starting with hot reload...
air
goto end

:clean
echo 🧹 Cleaning up...
if exist bin rmdir /s /q bin
if exist tmp rmdir /s /q tmp
if exist test_data rmdir /s /q test_data
if exist *.log del *.log
goto end

:fmt
echo 💅 Formatting code...
go fmt ./...
goto end

:coverage
echo 📊 Generating coverage report...
go test -coverprofile=coverage.out ./internal/... ./pkg/...
go tool cover -html=coverage.out -o coverage.html
echo 📋 Coverage report generated: coverage.html
start coverage.html
goto end

:end
