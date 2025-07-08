# Storage Control Plane - Go Monolith

🚀 **Microservices-ready Go implementation** of the distributed storage system control plane, designed to coordinate with Python microservices running on separate EC2 instances.

## 🎯 **Architecture Overview**

This Go control plane acts as the **orchestration layer** for a distributed storage system where:
- ✅ **Python microservices** run on dedicated EC2 instances (data processing)
- ✅ **Go control plane** runs on separate EC2 instance (coordination & control)
- ✅ **Single binary deployment** - No Docker complexity for the Go component
- ✅ **Inter-service communication** - HTTP APIs between Go and Python services
- ✅ **Distributed coordination** - Real-time communication across EC2 instances

### **Services in Control Plane**
```
┌─────────────────────────────────────────────────────────────────┐
│               Go Control Plane (EC2: 15.207.184.150)           │
├─────────────────────────────────────────────────────────────────┤
│ 🔐 Auth Gateway      (8090) │ 🏢 Tenant Node       (8000) │
│ 🎯 Operation Node    (8081) │ 🧠 CBO Engine        (8082) │  
│ 📊 Metadata Catalog  (8083) │ 📈 Monitoring        (8084) │
│ 🔍 Query Interpreter (8085) │                              │
└─────────────────────────────────────────────────────────────────┘
                                    ↕ HTTP APIs
┌─────────────────────────────────────────────────────────────────┐
│              Python Services (EC2: 65.0.150.75)               │
├─────────────────────────────────────────────────────────────────┤
│ 🔐 Auth Gateway      (8080) │ 🏢 Tenant Node       (8001) │
│ 🎯 Operation Node    (8086) │ 🧠 CBO Engine        (8088) │  
│ 📊 Metadata Catalog  (8087) │ 📈 Monitoring        (8089) │
│ 🔍 Query Interpreter (8085) │                              │
└─────────────────────────────────────────────────────────────────┘
```

## 🚀 **Quick Start**

### **Prerequisites**
- **Go 1.24+** (latest stable)
- **Git** for version control
- **Air** for hot reload: `go install github.com/air-verse/air@latest`

### **1. Setup & Run**
```bash
# Clone and navigate
git clone <your-repo-url>
cd storage_control_plane

# Install dependencies
go mod download

# Copy environment template  
cp .env.example .env

# Start with hot reload (recommended)
air

# OR start normally
go run .
```

### **2. Verify Services**
```bash
# Check all services are running
curl http://localhost:8090/health  # Auth Gateway
curl http://localhost:8000/health  # Tenant Node
curl http://localhost:8081/health  # Operation Node  
curl http://localhost:8082/health  # CBO Engine
curl http://localhost:8083/health  # Metadata Catalog
curl http://localhost:8084/health  # Monitoring
curl http://localhost:8085/health  # Query Interpreter

# View dashboard
open http://localhost:8084/dashboard
```

## 📊 **Service Endpoints**

### **🔐 Auth Gateway (Port 8090)**
```bash
POST /auth/login      # User authentication
POST /auth/validate   # Token validation  
POST /auth/refresh    # Token refresh
POST /auth/logout     # User logout
GET  /health          # Health check
```

### **🏢 Tenant Node (Port 8000)**
```bash
POST /data/execute    # Execute queries on data
POST /data/store      # Store new data
GET  /data/retrieve   # Retrieve data
GET  /data/stats      # Data statistics
GET  /health          # Health check
```

### **🎯 Operation Node (Port 8081)**
```bash
POST /query/execute   # Execute distributed queries
GET  /query/plan      # Get query execution plan
GET  /query/status    # Query execution status  
GET  /nodes/status    # Node cluster status
GET  /health          # Health check
```

### **🧠 CBO Engine (Port 8082)**
```bash
POST /optimize/query  # Optimize query plans
GET  /optimize/stats  # Optimizer statistics
GET  /optimize/config # Optimizer configuration
GET  /health          # Health check  
```

### **📊 Metadata Catalog (Port 8083)**
```bash
POST /metadata/partitions # Get partition metadata
GET  /metadata/tables     # Get table metadata
GET  /metadata/indexes    # Get index information
GET  /metadata/stats      # Metadata statistics
GET  /health              # Health check
```

### **📈 Monitoring (Port 8084)**
```bash
GET /metrics          # System & query metrics
GET /alerts           # Active alerts
GET /logs             # Recent logs  
GET /dashboard        # Web dashboard
GET /health           # Health check
```

### **🔍 Query Interpreter (Port 8085)**
```bash
POST /parse/sql       # Parse SQL queries
POST /parse/dsl       # Parse DSL queries  
POST /validate/query  # Validate query syntax
POST /transform/plan  # Transform to execution plan
GET  /health          # Health check
```

## 🧪 **Development Workflow**

### **1. Hot Reload Development**
```bash
# Install Air for hot reload
go install github.com/air-verse/air@latest

# Start development server (auto-restarts on file changes)
air

# The server will restart automatically when you modify .go files
```

### **2. Testing Endpoints**
```bash
# Test authentication
curl -X POST http://localhost:8090/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# Test query execution  
curl -X POST http://localhost:8081/query/execute \
  -H "Content-Type: application/json" \
  -d '{"query":"SELECT * FROM orders LIMIT 10"}'

# Test SQL parsing
curl -X POST http://localhost:8085/parse/sql \
  -H "Content-Type: application/json" \
  -d '{"query":"SELECT customer_id, SUM(amount) FROM orders GROUP BY customer_id"}'
```

### **3. Manual Testing**
```bash
# Build and run
go build -o storage-control-plane .
./storage-control-plane

# Or direct run
go run .
```

## 🔧 **Configuration**

### **Environment Variables (.env)**
```bash
# Go Control Plane Configuration
PORT=8090
ENVIRONMENT=production
LOG_LEVEL=info

# Python Services Configuration  
PYTHON_IP=65.0.150.75

# Python Service Endpoints
AUTH_GATEWAY_URL=http://65.0.150.75:8080
TENANT_NODE_URL=http://65.0.150.75:8001
METADATA_CATALOG_URL=http://65.0.150.75:8087
OPERATION_NODE_URL=http://65.0.150.75:8086
CBO_ENGINE_URL=http://65.0.150.75:8088
MONITORING_URL=http://65.0.150.75:8089
QUERY_INTERPRETER_URL=http://65.0.150.75:8085

# Distributed Mode
DISTRIBUTED_MODE=true
PYTHON_SERVICES_HOST=65.0.150.75
GO_SERVICES_HOST=15.207.184.150
```

## 🏗️ **Migration to Microservices**

When ready to split into individual microservices:

### **1. Service Extraction Pattern**
```bash
# Each service will become its own repository:
storage-auth-gateway/     # Auth Gateway service
storage-tenant-node/      # Tenant Node service  
storage-operation-node/   # Operation Node service
storage-cbo-engine/       # CBO Engine service
storage-metadata-catalog/ # Metadata Catalog service
storage-monitoring/       # Monitoring service
storage-query-interpreter/# Query Interpreter service
```

### **2. Code Structure**
```
main.go                  # Main application entry point
services.go              # Auth, Tenant, Operation node handlers  
services_extended.go     # CBO, Metadata, Query, Monitoring handlers
.env                     # Configuration
.env.example            # Environment template
go.mod                   # Dependencies
go.sum                   # Dependency checksums
.air.toml               # Hot reload configuration
Makefile                 # Build automation
```

### **3. Split Strategy**
1. **Create individual repositories** for each service
2. **Extract handler functions** to separate main.go files
3. **Add service-specific dependencies** to go.mod
4. **Implement inter-service communication** (HTTP/gRPC)
5. **Add Docker containers** for each service  
6. **Deploy with Docker Compose** or Kubernetes

## 🚀 **Production Deployment**

### **1. Build for Production**
```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o storage-control-plane .

# Or with Docker
docker build -t storage-control-plane .
docker run -p 8080-8085:8080-8085 storage-control-plane
```

### **2. Performance Tuning**  
```bash
# Environment variables for production
export GOMAXPROCS=4
export CGO_ENABLED=0  
export GOOS=linux
```

## 📊 **System Monitoring**

### **Health Checks**
All services expose `/health` endpoints returning:
```json
{
  "status": "healthy",
  "service": "Auth Gateway", 
  "version": "1.0.0",
  "time": "2024-07-02T10:30:00Z"
}
```

### **Metrics Dashboard**
Visit: `http://localhost:8084/dashboard`

### **API Monitoring**
```bash
# Get system metrics
curl http://localhost:8084/metrics

# Check active alerts  
curl http://localhost:8084/alerts

# View recent logs
curl http://localhost:8084/logs
```

## 🔄 **Comparison with Python Version**

| Aspect | **Python Microservices** | **Go Monolith** |
|--------|---------------------------|------------------|
| **🏗️ Architecture** | 7 separate Docker containers | Single Go binary |
| **🚀 Startup Time** | 10-30 seconds (Docker) | 1-2 seconds |
| **🔧 Development** | Docker Compose required | `go run .` |
| **🧪 Testing** | Complex service orchestration | Simple local testing |
| **📦 Deployment** | Multiple containers | Single binary |
| **🔍 Debugging** | Distributed logs | Single process |
| **⚡ Performance** | Network overhead between services | In-memory communication |
| **🎯 Production** | True microservices | Monolith (for now) |

## 🎯 **Current Status**

✅ **COMPLETED:**
- ✅ **Service Architecture** - All 7 services implemented
- ✅ **HTTP Endpoints** - Complete API surface area
- ✅ **Configuration** - Environment-based config
- ✅ **Health Checks** - Individual service monitoring  
- ✅ **Graceful Shutdown** - Clean service termination
- ✅ **Development Setup** - Hot reload with Air
- ✅ **Mock Responses** - Realistic test data

🔄 **IN PROGRESS:**
- 🔄 **Unit Tests** - Comprehensive test coverage
- 🔄 **Integration Tests** - Service interaction testing
- 🔄 **Database Integration** - Real data persistence
- 🔄 **Performance Benchmarks** - Load testing

📋 **NEXT STEPS:**
1. **Add comprehensive unit tests** for all services
2. **Implement database integration** (PostgreSQL + Redis)
3. **Add real query parsing** with SQLGlot equivalent  
4. **Performance testing** and optimization
5. **Split into microservices** repositories
6. **Container deployment** with Docker/Kubernetes

## 💡 **Development Tips**

### **Air Configuration (.air.toml)**
```toml
# Automatically reload on .go file changes
[build]
  cmd = "go build -o ./tmp/main ."
  bin = "./tmp/main"
  include_ext = ["go"]
  exclude_dir = ["assets", "tmp", "vendor"]
```

### **VS Code Setup**
1. **Install Go extension**
2. **Configure settings.json:**
```json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.formatTool": "goimports"
}
```

### **Useful Commands**
```bash
go mod tidy              # Clean up dependencies
go fmt ./...             # Format all Go files  
go vet ./...             # Static analysis
go test ./...            # Run all tests
go build -race .         # Build with race detection
```

---

🎯 **This Go monolith provides the perfect foundation for rapid development while maintaining clean service boundaries for future microservices extraction.**

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
Check `.air.toml` configuration and ensure Air is installed:
```bash
go install github.com/air-verse/air@latest
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
go build -ldflags="-s -w" -o storage-control-plane .

# Run production binary
./storage-control-plane
```

## 📊 Performance Monitoring

### Memory Usage
```bash
go test -bench=. -benchmem ./...
```

### CPU Profiling
```bash
go test -cpuprofile cpu.prof -bench . ./...
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
2. **Run Tests**: `make test`
3. **Add Features**: Modify code, Air will auto-reload
4. **Deploy**: Use deployment guides in this repository
5. **Commit Changes**: Follow Git workflow

Happy coding! 🚀
