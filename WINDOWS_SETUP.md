# Windows Development Commands - Storage Control Plane

## 🪟 Windows-Specific Setup

Since you're on Windows and may not have `make`, here are the alternatives:

### **Option 1: PowerShell Script (Recommended)**
```powershell
# Show all commands
.\dev.ps1 help

# Start development with hot reload
.\dev.ps1 dev

# Run tests
.\dev.ps1 test

# Run E2E tests
.\dev.ps1 test-e2e

# Build application
.\dev.ps1 build
```

### **Option 2: Batch Script**
```cmd
# Show all commands
dev.bat help

# Start development
dev.bat dev

# Run tests
dev.bat test
```

### **Option 3: Direct Go Commands**
```powershell
# Start development
air

# Run tests
go test ./internal/... ./pkg/... -v

# Run E2E tests (server must be running)
.\test_e2e.ps1

# Build
go build -o bin\storage-control-plane.exe .\cmd\api
```

## 🧪 **Testing on Windows**

### **1. Start the Server**
```powershell
# Method 1: PowerShell script
.\dev.ps1 dev

# Method 2: Direct command
air
```

### **2. Run Unit Tests**
```powershell
# Method 1: PowerShell script
.\dev.ps1 test-unit

# Method 2: Direct command
go test ./internal/... ./pkg/... -v
```

### **3. Run End-to-End Tests**
```powershell
# Method 1: PowerShell script
.\dev.ps1 test-e2e

# Method 2: Direct command
.\test_e2e.ps1
```

## 🔧 **Windows-Specific Commands**

### **Check if Server is Running**
```powershell
# Test connectivity
Invoke-RestMethod -Uri "http://localhost:8081/"

# Check what's using port 8081
netstat -ano | findstr :8081
```

### **Kill Process on Port**
```powershell
# Find process using port 8081
$proc = Get-NetTCPConnection -LocalPort 8081 -ErrorAction SilentlyContinue
if ($proc) { Stop-Process -Id $proc.OwningProcess -Force }
```

### **View Logs**
```powershell
# View Air logs
Get-Content air.log -Tail 20 -Wait

# View application logs
Get-Content tmp\*.log -Tail 20 -Wait
```

## 📊 **Coverage Report (Windows)**
```powershell
# Generate and open coverage
.\dev.ps1 coverage

# Manual generation
go test -coverprofile=coverage.out ./internal/... ./pkg/...
go tool cover -html=coverage.out -o coverage.html
start coverage.html
```

## 🚀 **Quick Test Workflow**

### **Complete Test Cycle**
```powershell
# 1. Start server
.\dev.ps1 dev

# 2. In another PowerShell window, run tests
.\dev.ps1 test-e2e

# 3. Check results
# ✅ Should see test results and server logs
```

### **Development Workflow**
```powershell
# 1. Start development
.\dev.ps1 dev

# 2. Make code changes (Air will auto-restart)

# 3. Test changes
.\test_e2e.ps1

# 4. Run unit tests
.\dev.ps1 test-unit
```

## 🔍 **Debugging on Windows**

### **Check Go Environment**
```powershell
go version
go env GOPATH
go env GOROOT
```

### **Check Air Installation**
```powershell
air -v
# If not found: go install github.com/air-verse/air@latest
```

### **Check Dependencies**
```powershell
go mod verify
go mod tidy
```

## 🎯 **Windows Best Practices**

1. **Use PowerShell** - Better than Command Prompt
2. **Enable Execution Policy** - For PowerShell scripts:
   ```powershell
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```
3. **Use Windows Terminal** - Better terminal experience
4. **Install Git Bash** - Alternative to PowerShell if preferred

## 🆘 **Common Windows Issues**

### **PowerShell Script Won't Run**
```
cannot be loaded because running scripts is disabled
```
**Solution:**
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### **Air Command Not Found**
```
'air' is not recognized as an internal or external command
```
**Solutions:**
```powershell
# Reinstall Air
go install github.com/air-verse/air@latest

# Add Go bin to PATH
$env:PATH += ";$(go env GOPATH)\bin"

# Or use full path
& "$(go env GOPATH)\bin\air.exe"
```

### **Port Permission Issues**
```
bind: Only one usage of each socket address is normally permitted
```
**Solutions:**
```powershell
# Kill process on port
netstat -ano | findstr :8081
taskkill /PID <PID> /F

# Or use different port in .env
SERVER_ADDR=:8082
```

---

**Your Windows development environment is now fully configured! 🪟🚀**
