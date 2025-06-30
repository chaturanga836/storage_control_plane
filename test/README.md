# Test Configuration

This directory contains all test-related files organized by test type:

## Directory Structure

```
test/
├── unit/               # Unit tests - fast, no external dependencies
├── integration/        # Integration tests - require external services
├── e2e/               # End-to-end tests - full system tests
├── testdata/          # Test data files and databases
├── fixtures/          # Test fixtures and mock data
└── README.md          # This file
```

## Test Types

### Unit Tests (`test/unit/`)
- **Purpose**: Test individual functions/components in isolation
- **Dependencies**: None (use mocks/stubs)
- **Speed**: Very fast (< 1s each)
- **Files moved here**:
  - `large_scale_sort_test.go` - Sort utilities testing
  - `sort_utils_test.go` - Basic sort function testing
  - `wal_test.go` - Write-ahead log testing

### Integration Tests (`test/integration/`)
- **Purpose**: Test components working together with real dependencies
- **Dependencies**: ClickHouse, databases, external services
- **Speed**: Moderate (1-10s each)
- **Files moved here**:
  - `schema_test.go` - ClickHouse schema testing
  - `index_manager_test.go` - Index management testing
  - `distributed_index_manager_test.go` - Distributed index testing
  - `server_test.go` - API server integration testing

### E2E Tests (`test/e2e/`)
- **Purpose**: Test complete workflows and user scenarios
- **Dependencies**: Full system stack
- **Speed**: Slow (10s+ each)
- **Files moved here**:
  - `test_e2e.ps1` - PowerShell E2E test script
  - `test_e2e.sh` - Shell E2E test script

### Test Data (`test/testdata/`)
- **Purpose**: Static test data files
- **Files**:
  - `duck.db` - Test database file
  - Add sample Parquet files, JSON fixtures, etc.

### Fixtures (`test/fixtures/`)
- **Purpose**: Test fixtures, mock data generators, test utilities
- **Usage**: Shared test setup code, mock services, data generators

## Running Tests

### Unit Tests Only (Fast)
```bash
go test ./test/unit/... -v
```

### Integration Tests (Requires ClickHouse)
```bash
# Start ClickHouse first
docker run -d --name clickhouse-test -p 9000:9000 clickhouse/clickhouse-server

# Run integration tests
go test ./test/integration/... -v
```

### All Tests
```bash
make test-all
```

### E2E Tests
```bash
# Windows
.\test\e2e\test_e2e.ps1

# Linux/Mac
./test/e2e/test_e2e.sh
```

## Test Environment Variables

Create `.env.test` in the root directory:
```env
# ClickHouse test configuration
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_DATABASE=storage_control_plane_test
CLICKHOUSE_USERNAME=default
CLICKHOUSE_PASSWORD=

# Test data paths
TEST_DATA_DIR=./test/testdata
TEST_FIXTURES_DIR=./test/fixtures
```

## Best Practices

1. **Unit tests** should be fast and not depend on external services
2. **Integration tests** can use real services but should clean up after themselves
3. **E2E tests** should test realistic user workflows
4. Use **testdata** for static test files
5. Use **fixtures** for generating dynamic test data
6. Each test should be independent and idempotent
7. Use descriptive test names that explain the scenario being tested

## Adding New Tests

### For a new unit test:
1. Create `*_test.go` file in `test/unit/`
2. Use package name matching the code being tested
3. Mock external dependencies

### For a new integration test:
1. Create `*_test.go` file in `test/integration/`
2. Include setup/teardown for external services
3. Use test containers when possible

### For E2E tests:
1. Add test scenarios to existing E2E scripts
2. Test complete user workflows
3. Verify end-to-end functionality
