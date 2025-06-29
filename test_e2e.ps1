# End-to-End Test Script for Storage Control Plane (PowerShell)
# Usage: .\test_e2e.ps1

Write-Host "🧪 Starting End-to-End Tests..." -ForegroundColor Cyan

# Configuration
$BaseUrl = "http://localhost:8081"
$TenantId = "test-tenant-$(Get-Date -Format 'yyyyMMddHHmmss')"
$SourceId = "test-source-001"

Write-Host "📋 Using Tenant ID: $TenantId" -ForegroundColor Yellow
Write-Host "🔗 Using Source ID: $SourceId" -ForegroundColor Yellow

# Test 1: Health Check
Write-Host "🏥 Testing server health..." -ForegroundColor Green
try {
    $healthResponse = Invoke-WebRequest -Uri "$BaseUrl/" -Method GET -TimeoutSec 5
    Write-Host "✅ Server is responding" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Health endpoint not implemented or server not running" -ForegroundColor Yellow
}

# Test 2: POST Data Ingestion
Write-Host "📤 Testing data ingestion..." -ForegroundColor Green

$testData = @{
    data_id = "user-001"
    payload = @{
        name = "John Doe"
        age = 30
        email = "john@example.com"
        profile = @{
            bio = "Software Engineer"
            location = "San Francisco"
            skills = @("Go", "Python", "JavaScript")
        }
        metadata = @{
            created_at = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
            version = 1
        }
    }
} | ConvertTo-Json -Depth 10

$headers = @{
    "Content-Type" = "application/json"
    "X-Tenant-Id" = $TenantId
}

try {
    $postResponse = Invoke-WebRequest -Uri "$BaseUrl/data" -Method POST -Body $testData -Headers $headers -TimeoutSec 10
    Write-Host "✅ Data ingestion successful - Status: $($postResponse.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "❌ Data ingestion failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
}

# Test 3: GET Data Retrieval
Write-Host "📥 Testing data retrieval..." -ForegroundColor Green

$getHeaders = @{
    "X-Tenant-Id" = $TenantId
}

try {
    $getResponse = Invoke-WebRequest -Uri "$BaseUrl/data" -Method GET -Headers $getHeaders -TimeoutSec 10
    Write-Host "✅ Data retrieval successful - Status: $($getResponse.StatusCode)" -ForegroundColor Green
    Write-Host "📋 Retrieved data length: $($getResponse.Content.Length) chars" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Data retrieval failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Schema Evolution Test
Write-Host "🔄 Testing schema evolution..." -ForegroundColor Green

$evolvedData = @{
    data_id = "user-002"
    payload = @{
        name = "Jane Smith"
        age = 25
        email = "jane@example.com"
        profile = @{
            bio = "Data Scientist"
            location = "New York"
            skills = @("Python", "R", "SQL")
            certifications = @("AWS", "GCP")
        }
        preferences = @{
            theme = "dark"
            notifications = $true
        }
        metadata = @{
            created_at = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
            version = 1
            source = "api_v2"
        }
    }
} | ConvertTo-Json -Depth 10

try {
    $schemaResponse = Invoke-WebRequest -Uri "$BaseUrl/data" -Method POST -Body $evolvedData -Headers $headers -TimeoutSec 10
    Write-Host "✅ Schema evolution test successful - Status: $($schemaResponse.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "❌ Schema evolution test failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 5: Bulk Data Test
Write-Host "🔄 Testing bulk data ingestion..." -ForegroundColor Green

for ($i = 1; $i -le 5; $i++) {
    $bulkData = @{
        data_id = "bulk-$i"
        payload = @{
            batch_id = $i
            timestamp = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
            data = @{
                value = $i * 10
                processed = $true
            }
        }
    } | ConvertTo-Json -Depth 10
    
    try {
        Invoke-WebRequest -Uri "$BaseUrl/data" -Method POST -Body $bulkData -Headers $headers -TimeoutSec 5 | Out-Null
        Write-Host "📦 Bulk record $i sent" -ForegroundColor Cyan
    } catch {
        Write-Host "❌ Bulk record $i failed" -ForegroundColor Red
    }
}

Write-Host "`n🎉 End-to-End Tests Completed!" -ForegroundColor Green
Write-Host "`n📝 Test Summary:" -ForegroundColor Yellow
Write-Host "   - Health Check: Basic connectivity"
Write-Host "   - Data Ingestion: JSON data with nested structures"
Write-Host "   - Data Retrieval: Reading stored data"
Write-Host "   - Schema Evolution: Different JSON structure"
Write-Host "   - Bulk Processing: Multiple records"
Write-Host "`n🔍 Check application logs for WAL/Parquet processing" -ForegroundColor Cyan
Write-Host "🗄️  Check ClickHouse for table creation and data storage" -ForegroundColor Cyan
