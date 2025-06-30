# Testing Guide - Storage Control Plane

## 🧪 Testing Strategy Overview

This project uses a comprehensive testing approach with organized test directories:
- **Unit Tests** (`test/unit/`) - Test individual functions and components (no dependencies)
- **Integration Tests** (`test/integration/`) - Test component interactions (requires services)
- **End-to-End Tests** (`test/e2e/`) - Test complete user workflows (full system)
- **Test Fixtures** (`test/fixtures/`) - Shared test data generators and utilities

## 📁 Test Directory Structure

```
test/
├── unit/               # Unit tests - fast, no external dependencies
│   ├── large_scale_sort_test.go
│   ├── sort_utils_test.go
│   └── wal_test.go
├── integration/        # Integration tests - require external services
│   ├── schema_test.go
│   ├── index_manager_test.go
│   ├── distributed_index_manager_test.go
│   └── server_test.go
├── e2e/               # End-to-end tests - full system tests
│   ├── test_e2e.ps1
│   └── test_e2e.sh
├── testdata/          # Test data files and databases
│   └── duck.db
├── fixtures/          # Test fixtures and mock data generators
│   └── test_data_generator.go
├── run_tests.sh       # Unix test runner script
├── run_tests.ps1      # Windows test runner script
└── README.md          # Detailed test documentation
```

## 🏃‍♂️ Quick Test Commands

### Using Test Runner Scripts (Recommended)

**Windows:**
```powershell
# Run all tests
.\test\run_tests.ps1 all

# Run only unit tests (fast, no dependencies)
.\test\run_tests.ps1 unit

# Run integration tests (requires ClickHouse)
.\test\run_tests.ps1 integration

# Run E2E tests
.\test\run_tests.ps1 e2e

# Generate coverage report
.\test\run_tests.ps1 coverage
```

**Linux/Mac:**
```bash
# Run all tests
./test/run_tests.sh all

# Run only unit tests (fast, no dependencies)  
./test/run_tests.sh unit

# Run integration tests (requires ClickHouse)
./test/run_tests.sh integration

# Run E2E tests
./test/run_tests.sh e2e

# Generate coverage report
./test/run_tests.sh coverage
```

### Using Make
```bash
# Unit tests only (no dependencies required)
make test-unit

# Integration tests (requires ClickHouse)
make test-integration

# E2E tests
make test-e2e

# All tests (unit + integration)
make test

# All tests including E2E
make test-all
```

## 📝 Unit Testing

### Run Unit Tests
```bash
# All packages
go test ./internal/... ./pkg/... -v

# Specific packages
go test ./internal/clickhouse -v
go test ./internal/wal -v
go test ./internal/api -v

# With coverage
go test -cover ./internal/...
```

### Example Unit Test Output
```
=== RUN   TestMapJSONTypeToClickHouseType
=== RUN   TestMapJSONTypeToClickHouseType/string_type
=== RUN   TestMapJSONTypeToClickHouseType/bool_type
=== RUN   TestMapJSONTypeToClickHouseType/integer_float
=== RUN   TestMapJSONTypeToClickHouseType/decimal_float
--- PASS: TestMapJSONTypeToClickHouseType (0.00s)
    --- PASS: TestMapJSONTypeToClickHouseType/string_type (0.00s)
    --- PASS: TestMapJSONTypeToClickHouseType/bool_type (0.00s)
    --- PASS: TestMapJSONTypeToClickHouseType/integer_float (0.00s)
    --- PASS: TestMapJSONTypeToClickHouseType/decimal_float (0.00s)
PASS
```

## 🔗 Integration Testing

### API Integration Tests
```bash
# Test HTTP handlers
go test ./internal/api -v

# Test with external dependencies (requires services)
go test -tags=integration ./tests/integration/...
```

### Example Integration Test
```go
func TestServerRequiresTenantHeader(t *testing.T) {
    server := NewServer(mockRouter)
    req := httptest.NewRequest("GET", "/data", nil)
    w := httptest.NewRecorder()
    
    server.ServeHTTP(w, req)
    
    if w.Code != http.StatusBadRequest {
        t.Errorf("Expected 400, got %d", w.Code)
    }
}
```

## 🌐 End-to-End Testing

### Prerequisites
1. **Server must be running:**
   ```bash
   air  # or go run ./cmd/api
   ```

2. **Server should be accessible on `localhost:8081`**

### Run E2E Tests

**Windows PowerShell:**
```powershell
.\test_e2e.ps1
```

**Linux/Mac:**
```bash
bash test_e2e.sh
```

**Using Make:**
```bash
make test-e2e
```

### E2E Test Scenarios

The E2E tests cover:

1. **🏥 Health Check** - Basic server connectivity
2. **📤 Data Ingestion** - POST JSON data with nested structures
3. **📥 Data Retrieval** - GET stored data
4. **🔄 Schema Evolution** - Handle different JSON schemas
5. **📦 Bulk Processing** - Multiple records ingestion

### Example E2E Test Output
```
🧪 Starting End-to-End Tests...
📋 Using Tenant ID: test-tenant-20250629203000
🔗 Using Source ID: test-source-001
🏥 Testing server health...
✅ Server is responding
📤 Testing data ingestion...
✅ Data ingestion successful - Status: 201
📥 Testing data retrieval...
✅ Data retrieval successful - Status: 200
📋 Retrieved data length: 156 chars
🔄 Testing schema evolution...
✅ Schema evolution test successful - Status: 201
🔄 Testing bulk data ingestion...
📦 Bulk record 1 sent
📦 Bulk record 2 sent
📦 Bulk record 3 sent
📦 Bulk record 4 sent
📦 Bulk record 5 sent
🎉 End-to-End Tests Completed!
```

## 📊 Coverage Reports

### Generate Coverage Report
```bash
# Generate coverage data
go test -coverprofile=coverage.out ./internal/... ./pkg/...

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# View in browser
start coverage.html  # Windows
open coverage.html   # Mac
```

### Coverage Goals
- **Unit Tests**: >80% coverage
- **Integration Tests**: >70% coverage
- **Critical Path**: >95% coverage

## ⚡ Performance Testing

### Benchmark Tests
```bash
# Run benchmarks
go test -bench=. -benchmem ./internal/...

# Specific benchmarks
go test -bench=BenchmarkFlattenJSON ./internal/clickhouse
```

### Memory Profiling
```bash
# Generate memory profile
go test -memprofile mem.prof -bench . ./internal/...

# Analyze profile
go tool pprof mem.prof
```

### Load Testing
```powershell
# Simple load test with PowerShell
1..100 | ForEach-Object -Parallel {
    $data = @{ data_id = "load-$_"; payload = @{ value = $_ } } | ConvertTo-Json
    Invoke-RestMethod -Uri "http://localhost:8081/data" `
        -Method POST -Body $data `
        -Headers @{"Content-Type"="application/json"; "X-Tenant-Id"="load-test"}
} -ThrottleLimit 10
```

## 🧹 Test Data Management

### Test Data Location
```
test_data/
├── rocksdb/           # Test RocksDB data
├── parquet/           # Test Parquet files
└── wal/               # Test WAL files
```

### Clean Test Data
```bash
# Clean all test artifacts
make clean

# Manual cleanup
Remove-Item -Recurse -Force test_data, tmp, coverage.*
```

### Test Environment
```bash
# Use test environment
cp .env.test .env

# Or set environment variables
$env:GO_ENV = "test"
$env:LOG_LEVEL = "debug"
```

## 🔍 Debugging Tests

### Verbose Test Output
```bash
# Show detailed test output
go test -v ./internal/...

# Show test logs
go test -v -args -test.v ./internal/...
```

### Debug Failed Tests
```bash
# Run specific test
go test -run TestSpecificFunction ./internal/clickhouse

# Debug with delve (if installed)
dlv test ./internal/clickhouse -- -test.run TestSpecificFunction
```

### Test Debugging Tips
1. **Add debug prints**: Use `t.Logf()` in tests
2. **Check test data**: Verify input/output in test failures
3. **Isolate tests**: Run one test at a time
4. **Check logs**: Application logs show what's happening

## 🚨 Continuous Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.24
      - run: go test ./...
      - run: go test -race ./...
```

### Test in Docker
```dockerfile
FROM golang:1.24
WORKDIR /app
COPY . .
RUN go test ./...
```

## 📋 Test Checklist

Before committing code:

- [ ] All unit tests pass: `make test-unit`
- [ ] Integration tests pass: `make test-integration`
- [ ] E2E tests pass: `make test-e2e`
- [ ] Code coverage >80%: `make test-coverage`
- [ ] No race conditions: `go test -race ./...`
- [ ] Benchmarks stable: `go test -bench=.`

## 🆘 Troubleshooting Tests

### Common Test Issues

**1. E2E Tests Fail - Server Not Running**
```
Error: Connection refused
```
**Solution**: Start server first: `air`

**2. Import Errors in Tests**
```
Error: package not found
```
**Solution**: `go mod tidy`

**3. Test Data Conflicts**
```
Error: file already exists
```
**Solution**: `make clean`

**4. Race Condition Warnings**
```
WARNING: DATA RACE
```
**Solution**: Add proper locking or use atomic operations

### Getting Help
1. Check test logs for detailed error messages
2. Run tests with `-v` flag for verbose output
3. Verify test environment configuration
4. Clean test data and try again

---

**Happy Testing! Your Storage Control Plane is thoroughly tested! 🧪✅**
