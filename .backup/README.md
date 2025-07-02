# Storage Control Plane - Development Guide

## 🚀 Quick Start

### Prerequisites
- **Go 1.24+** installed
- **Air** for hot reload: `go install github.com/air-verse/air@latest`
- **Make** (optional, for convenience commands)
- **Git** for version control

### 1. Clone and Setup
```bash
git clone <your-repo-url>
cd storage_control_plane

# Copy environment template
cp .env.example .env

# Install dependencies
go mod download
```

### 2. Start Development Server
```bash
# Method 1: Using Air (Hot Reload) - Recommended
air

# Method 2: Using Make
make dev

# Method 3: Manual
go run ./cmd/api
```

The server will start on `http://localhost:8081` with hot reload enabled.

## 🔧 Development Workflow

### Running with Air
Air automatically:
- ✅ Watches for file changes
- ✅ Rebuilds the application
- ✅ Restarts the server
- ✅ Shows build errors in real-time

**Air Configuration** (`.air.toml`):
```toml
[build]
  cmd = "go build -o ./tmp/main.exe ./cmd/api"
  bin = "tmp/main.exe"
  include_ext = ["go"]
  exclude_dir = ["tmp"]
```

### Environment Configuration

**Development (`.env`):**
```env
SERVER_ADDR=:8081
SHARED_ROCKSDB_PATH=./data/shared_rocksdb
SHARED_CLICKHOUSE_DSN=tcp://localhost:9000?debug=true
POSTGRES_DSN=postgres://user:pass@localhost:5432/core
GO_ENV=development
LOG_LEVEL=debug
```

### File Structure
```
storage_control_plane/
├── cmd/api/main.go           # Application entry point
├── internal/                 # Private application code
│   ├── api/                 # HTTP handlers
│   ├── clickhouse/          # ClickHouse integration
│   ├── config/              # Configuration management
│   ├── routing/             # Request routing
│   ├── wal/                 # Write-Ahead Log
│   └── writers/             # Data writers (Parquet, metadata)
├── pkg/models/              # Shared data models
├── data/                    # Local data storage
├── .env                     # Environment variables
└── .air.toml               # Air configuration
```

## 🧪 Testing Guide

### Unit Tests
```bash
# Run all unit tests
go test ./internal/... ./pkg/... -v

# Run specific package tests
go test ./internal/clickhouse -v
go test ./internal/wal -v

# Run with coverage
go test -cover ./internal/... ./pkg/...

# Generate coverage report
go test -coverprofile=coverage.out ./internal/... ./pkg/...
go tool cover -html=coverage.out -o coverage.html
```

### Integration Tests
```bash
# Make sure server is running first
air

# In another terminal, run E2E tests
.\test_e2e.ps1              # Windows PowerShell
# or
bash test_e2e.sh            # Linux/Mac
```

### Manual API Testing

**Health Check:**
```powershell
Invoke-RestMethod -Uri "http://localhost:8081/"
```

**Data Ingestion:**
```powershell
$testData = @{
    data_id = "user-001"
    payload = @{
        name = "John Doe"
        age = 30
        email = "john@example.com"
        profile = @{
            bio = "Software Engineer"
            skills = @("Go", "Python", "JavaScript")
        }
    }
} | ConvertTo-Json -Depth 5

Invoke-RestMethod -Uri "http://localhost:8081/data" `
    -Method POST `
    -Body $testData `
    -Headers @{
        "Content-Type" = "application/json"
        "X-Tenant-Id" = "test-tenant"
    }
```

**Data Retrieval:**
```powershell
Invoke-RestMethod -Uri "http://localhost:8081/data" `
    -Headers @{"X-Tenant-Id" = "test-tenant"}
```

## 🔧 Make Commands

```bash
make help           # Show all available commands
make build          # Build the application
make run            # Run the application
make dev            # Run with hot reload (air)
make test           # Run all tests
make test-unit      # Run unit tests only
make test-e2e       # Run end-to-end tests
make clean          # Clean build artifacts
make lint           # Lint code (requires golangci-lint)
make fmt            # Format code
```

## 🐛 Debugging

### Common Issues

**1. Port Already in Use:**
```
Error: listen tcp :8081: bind: Only one usage of each socket address...
```
**Solution:**
- Change port in `.env`: `SERVER_ADDR=:8082`
- Or kill existing process: `netstat -ano | findstr :8081`

**2. Module Import Errors:**
```
Error: package not found
```
**Solution:**
- Run `go mod tidy`
- Check import paths match `go.mod`

**3. Air Not Working:**
```
Error: air: command not found
```
**Solution:**
```bash
go install github.com/air-verse/air@latest
# Make sure $GOPATH/bin is in your PATH
```

### Debug Logging
Set `LOG_LEVEL=debug` in `.env` for verbose logging:
```env
LOG_LEVEL=debug
```

### Hot Reload Not Working
Check `.air.toml` configuration:
```toml
[build]
  include_ext = ["go"]
  exclude_dir = ["tmp", "data", ".git"]
```

## 📁 Data Storage

### Local Development Data
```
data/
├── shared_rocksdb/          # RocksDB storage
├── tenant_*/                # Per-tenant data
│   └── source_*/           # Per-source data
│       ├── *.parquet       # Parquet files
│       └── _stats.json     # Metadata
└── wal/                    # Write-Ahead Log files
```

### Cleanup Development Data
```bash
make clean
# or manually
rm -rf data/ tmp/
```

## 🔄 Git Workflow

### Ignored Files
```gitignore
# Build artifacts
tmp/
*.exe
*.log

# Environment
.env
.env.local

# Data
data/
```

### Development Branch
```bash
git checkout -b feature/your-feature
# Make changes, test with air
git add .
git commit -m "feat: your feature description"
git push origin feature/your-feature
```

## 🚀 Production Build

```bash
# Build optimized binary
go build -ldflags="-s -w" -o bin/storage-control-plane ./cmd/api

# Run production binary
./bin/storage-control-plane
```

## 📊 Performance Monitoring

### Memory Usage
```bash
go test -bench=. -benchmem ./internal/...
```

### CPU Profiling
```bash
go test -cpuprofile cpu.prof -bench . ./internal/...
go tool pprof cpu.prof
```

## 🆘 Getting Help

1. **Check Logs**: Application logs show detailed error information
2. **Run Tests**: `make test` to verify everything works
3. **Check Configuration**: Verify `.env` settings
4. **Restart Air**: `Ctrl+C` and `air` again
5. **Clean Build**: `make clean` then `air`

## 🎯 Next Steps

1. **Start Development**: `air`
2. **Run Tests**: `.\test_e2e.ps1`
3. **Add Features**: Modify code, Air will auto-reload
4. **Test Changes**: Use the test scripts
5. **Commit Changes**: Follow Git workflow

Happy coding! 🚀
