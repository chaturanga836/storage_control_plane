# Test Runner PowerShell Script for Storage Control Plane
# This script provides an easy way to run different types of tests on Windows

param(
    [Parameter(Position=0)]
    [ValidateSet("unit", "integration", "e2e", "all", "coverage", "clean", "help")]
    [string]$Command = "all"
)

# Colors for output
$Colors = @{
    Red    = "Red"
    Green  = "Green"  
    Yellow = "Yellow"
    Blue   = "Blue"
    White  = "White"
}

# Test directories
$UnitDir = "./test/unit"
$IntegrationDir = "./test/integration"
$E2EDir = "./test/e2e"

# Functions
function Write-Header {
    param([string]$Message)
    Write-Host "================================" -ForegroundColor $Colors.Blue
    Write-Host $Message -ForegroundColor $Colors.Blue
    Write-Host "================================" -ForegroundColor $Colors.Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor $Colors.Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor $Colors.Red
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor $Colors.Yellow
}

# Load test environment
function Load-TestEnvironment {
    if (Test-Path ".env.test") {
        Get-Content ".env.test" | ForEach-Object {
            if ($_ -match "^([^#].*)=(.*)$") {
                [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
            }
        }
        Write-Success "Loaded test environment from .env.test"
    } else {
        Write-Warning "No .env.test file found, using defaults"
    }
}

# Check if ClickHouse is running
function Test-ClickHouse {
    try {
        $result = & clickhouse-client --query "SELECT 1" 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "ClickHouse is running"
            return $true
        }
    } catch {
        # Command not found or failed
    }
    
    Write-Error "ClickHouse is not running or not accessible"
    return $false
}

# Run unit tests
function Invoke-UnitTests {
    Write-Header "Running Unit Tests"
    
    if (!(Test-Path $UnitDir)) {
        Write-Error "Unit test directory not found: $UnitDir"
        return $false
    }
    
    go test -v "$UnitDir/..." -timeout=5m
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Unit tests passed"
        return $true
    } else {
        Write-Error "Unit tests failed"
        return $false
    }
}

# Run integration tests
function Invoke-IntegrationTests {
    Write-Header "Running Integration Tests"
    
    if (!(Test-Path $IntegrationDir)) {
        Write-Error "Integration test directory not found: $IntegrationDir"
        return $false
    }
    
    # Check if required services are running
    if (!(Test-ClickHouse)) {
        Write-Error "Integration tests require ClickHouse to be running"
        Write-Warning "Start ClickHouse with: docker run -d --name clickhouse-test -p 9000:9000 clickhouse/clickhouse-server"
        return $false
    }
    
    go test -v "$IntegrationDir/..." -timeout=10m
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Integration tests passed"
        return $true
    } else {
        Write-Error "Integration tests failed"
        return $false
    }
}

# Run E2E tests
function Invoke-E2ETests {
    Write-Header "Running End-to-End Tests"
    
    $e2eScript = "$E2EDir/test_e2e.ps1"
    if (!(Test-Path $e2eScript)) {
        Write-Error "E2E test script not found: $e2eScript"
        return $false
    }
    
    & PowerShell -ExecutionPolicy Bypass -File $e2eScript
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "E2E tests passed"
        return $true
    } else {
        Write-Error "E2E tests failed"
        return $false
    }
}

# Generate test coverage
function New-CoverageReport {
    Write-Header "Generating Test Coverage"
    
    if (!(Test-Path "coverage")) {
        New-Item -ItemType Directory -Path "coverage" | Out-Null
    }
    
    # Run tests with coverage
    go test -v "$UnitDir/..." -coverprofile=coverage/unit.out -timeout=5m
    go test -v "$IntegrationDir/..." -coverprofile=coverage/integration.out -timeout=10m
    
    # Merge coverage files
    "mode: set" | Out-File -FilePath coverage/total.out -Encoding utf8
    if (Test-Path "coverage/unit.out") {
        Get-Content "coverage/unit.out" | Select-Object -Skip 1 | Add-Content coverage/total.out
    }
    if (Test-Path "coverage/integration.out") {
        Get-Content "coverage/integration.out" | Select-Object -Skip 1 | Add-Content coverage/total.out
    }
    
    # Generate HTML report
    go tool cover -html=coverage/total.out -o coverage/coverage.html
    
    # Display coverage summary
    go tool cover -func=coverage/total.out | Select-Object -Last 1
    
    Write-Success "Coverage report generated at coverage/coverage.html"
}

# Clean test artifacts
function Remove-TestArtifacts {
    Write-Header "Cleaning Test Artifacts"
    
    if (Test-Path "coverage") { Remove-Item -Recurse -Force "coverage" }
    if (Test-Path "test/tmp") { Remove-Item -Recurse -Force "test/tmp" }
    Get-ChildItem -Path "test/testdata" -Filter "temp_*" | Remove-Item -Recurse -Force
    
    Write-Success "Test artifacts cleaned"
}

# Show help
function Show-Help {
    Write-Host "Test Runner for Storage Control Plane" -ForegroundColor $Colors.Blue
    Write-Host ""
    Write-Host "Usage: .\test\run_tests.ps1 [COMMAND]" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor $Colors.White
    Write-Host "  unit          Run unit tests only" -ForegroundColor $Colors.White
    Write-Host "  integration   Run integration tests only" -ForegroundColor $Colors.White
    Write-Host "  e2e           Run end-to-end tests only" -ForegroundColor $Colors.White
    Write-Host "  all           Run all tests" -ForegroundColor $Colors.White
    Write-Host "  coverage      Generate test coverage report" -ForegroundColor $Colors.White
    Write-Host "  clean         Clean test artifacts" -ForegroundColor $Colors.White
    Write-Host "  help          Show this help message" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor $Colors.Yellow
    Write-Host "  .\test\run_tests.ps1 unit                    # Run unit tests" -ForegroundColor $Colors.White
    Write-Host "  .\test\run_tests.ps1 integration            # Run integration tests" -ForegroundColor $Colors.White
    Write-Host "  .\test\run_tests.ps1 all                    # Run all tests" -ForegroundColor $Colors.White
    Write-Host "  .\test\run_tests.ps1 coverage               # Generate coverage report" -ForegroundColor $Colors.White
}

# Main execution
switch ($Command) {
    "unit" {
        Load-TestEnvironment
        $success = Invoke-UnitTests
        if (!$success) { exit 1 }
    }
    "integration" {
        Load-TestEnvironment
        $success = Invoke-IntegrationTests
        if (!$success) { exit 1 }
    }
    "e2e" {
        Load-TestEnvironment
        $success = Invoke-E2ETests
        if (!$success) { exit 1 }
    }
    "all" {
        Load-TestEnvironment
        $unitSuccess = Invoke-UnitTests
        $integrationSuccess = Invoke-IntegrationTests
        $e2eSuccess = Invoke-E2ETests
        
        if (!$unitSuccess -or !$integrationSuccess -or !$e2eSuccess) {
            exit 1
        }
    }
    "coverage" {
        Load-TestEnvironment
        New-CoverageReport
    }
    "clean" {
        Remove-TestArtifacts
    }
    "help" {
        Show-Help
    }
    default {
        Write-Error "Unknown command: $Command"
        Write-Host ""
        Show-Help
        exit 1
    }
}
