# Quick Start Guide - Storage Control Plane

## 🚀 Get Running in 2 Minutes

### Step 1: Setup Environment
```bash
# 1. Copy environment file
copy .env.example .env

# 2. Install dependencies (if needed)
go mod download
```

### Step 2: Start Development Server
```bash
# Start with hot reload (recommended)
air
```

Your server will start at: **http://localhost:8081**

### Step 3: Test It Works
```powershell
# Quick health check
Invoke-RestMethod -Uri "http://localhost:8081/"

# Test data ingestion
$data = @{
    data_id = "test-001"
    payload = @{ name = "Test User"; age = 25 }
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8081/data" `
    -Method POST -Body $data `
    -Headers @{"Content-Type"="application/json"; "X-Tenant-Id"="test"}
```

## 🧪 Run Tests

### Unit Tests
```bash
go test ./internal/... -v
```

### End-to-End Tests
```powershell
# Make sure server is running first!
.\test_e2e.ps1
```

## 🔧 Development Commands

```bash
air                    # Start with hot reload
go run ./cmd/api       # Start without hot reload
go test ./...          # Run all tests
go build ./cmd/api     # Build binary
```

## 📁 Key Files

- **`.env`** - Environment configuration
- **`cmd/api/main.go`** - Application entry point
- **`internal/`** - Private application code
- **`.air.toml`** - Hot reload configuration

## 🆘 Common Issues

**Port already in use?**
- Edit `.env`: `SERVER_ADDR=:8082`

**Air not found?**
- Install: `go install github.com/air-verse/air@latest`

**Import errors?**
- Run: `go mod tidy`

## 📖 Full Documentation
See [README.md](README.md) for complete setup and development guide.

---
**That's it! Your Storage Control Plane is ready for development! 🎉**
