#!/usr/bin/env pwsh
# PowerShell development script for Storage Control Plane

param(
    [Parameter(Position=0)]
    [ValidateSet("help", "test", "test-unit", "test-e2e", "build", "run", "dev", "clean", "fmt", "coverage", "lint")]
    [string]$Command = "help"
)

function Show-Help {
    Write-Host "🚀 Storage Control Plane - Development Commands" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\dev.ps1 <command>" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor Green
    Write-Host "  test          Run all tests" -ForegroundColor White
    Write-Host "  test-unit     Run unit tests" -ForegroundColor White
    Write-Host "  test-e2e      Run end-to-end tests" -ForegroundColor White
    Write-Host "  build         Build the application" -ForegroundColor White
    Write-Host "  run           Run the application" -ForegroundColor White
    Write-Host "  dev           Run with hot reload (air)" -ForegroundColor White
    Write-Host "  clean         Clean build artifacts" -ForegroundColor White
    Write-Host "  fmt           Format code" -ForegroundColor White
    Write-Host "  coverage      Generate coverage report" -ForegroundColor White
    Write-Host "  lint          Lint code (requires golangci-lint)" -ForegroundColor White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\dev.ps1 dev         # Start development server" -ForegroundColor Gray
    Write-Host "  .\dev.ps1 test        # Run all tests" -ForegroundColor Gray
    Write-Host "  .\dev.ps1 test-e2e    # Run E2E tests" -ForegroundColor Gray
}

function Test-All {
    Write-Host "🧪 Running all tests..." -ForegroundColor Green
    go test ./internal/... ./pkg/... -v
}

function Test-Unit {
    Write-Host "🧪 Running unit tests..." -ForegroundColor Green
    go test ./internal/... ./pkg/... -v
}

function Test-E2E {
    Write-Host "🌐 Running end-to-end tests..." -ForegroundColor Green
    Write-Host "⚠️  Make sure the server is running on :8081" -ForegroundColor Yellow
    & .\test_e2e.ps1
}

function Build-App {
    Write-Host "🔨 Building Storage Control Plane..." -ForegroundColor Green
    New-Item -ItemType Directory -Force -Path "bin" | Out-Null
    go build -o bin\storage-control-plane.exe .\cmd\api
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Build successful!" -ForegroundColor Green
    } else {
        Write-Host "❌ Build failed!" -ForegroundColor Red
    }
}

function Run-App {
    Write-Host "🚀 Starting Storage Control Plane..." -ForegroundColor Green
    go run .\cmd\api
}

function Start-Dev {
    Write-Host "🔥 Starting with hot reload..." -ForegroundColor Green
    if (Get-Command air -ErrorAction SilentlyContinue) {
        air
    } else {
        Write-Host "❌ Air not found. Install with: go install github.com/air-verse/air@latest" -ForegroundColor Red
        Write-Host "💡 Falling back to normal run..." -ForegroundColor Yellow
        go run .\cmd\api
    }
}

function Clean-All {
    Write-Host "🧹 Cleaning up..." -ForegroundColor Green
    
    $itemsToRemove = @("bin", "tmp", "test_data", "coverage.out", "coverage.html")
    
    foreach ($item in $itemsToRemove) {
        if (Test-Path $item) {
            Remove-Item -Recurse -Force $item
            Write-Host "  Removed: $item" -ForegroundColor Gray
        }
    }
    
    # Remove log files
    Get-ChildItem -Filter "*.log" | Remove-Item -Force
    Write-Host "✅ Cleanup complete!" -ForegroundColor Green
}

function Format-Code {
    Write-Host "💅 Formatting code..." -ForegroundColor Green
    go fmt ./...
    Write-Host "✅ Code formatted!" -ForegroundColor Green
}

function Generate-Coverage {
    Write-Host "📊 Generating coverage report..." -ForegroundColor Green
    go test -coverprofile=coverage.out ./internal/... ./pkg/...
    if ($LASTEXITCODE -eq 0) {
        go tool cover -html=coverage.out -o coverage.html
        Write-Host "✅ Coverage report generated: coverage.html" -ForegroundColor Green
        Start-Process coverage.html
    } else {
        Write-Host "❌ Coverage generation failed!" -ForegroundColor Red
    }
}

function Lint-Code {
    Write-Host "🔍 Linting code..." -ForegroundColor Green
    if (Get-Command golangci-lint -ErrorAction SilentlyContinue) {
        golangci-lint run
    } else {
        Write-Host "❌ golangci-lint not installed." -ForegroundColor Red
        Write-Host "💡 Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" -ForegroundColor Yellow
    }
}

# Execute command
switch ($Command) {
    "help" { Show-Help }
    "test" { Test-All }
    "test-unit" { Test-Unit }
    "test-e2e" { Test-E2E }
    "build" { Build-App }
    "run" { Run-App }
    "dev" { Start-Dev }
    "clean" { Clean-All }
    "fmt" { Format-Code }
    "coverage" { Generate-Coverage }
    "lint" { Lint-Code }
    default { Show-Help }
}
