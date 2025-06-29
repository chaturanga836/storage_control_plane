# Air Hot Reload Setup - Storage Control Plane

## 🔥 Air Configuration Guide

Air provides automatic hot reload for Go applications during development. This guide covers setup and troubleshooting.

## 📦 Installation

### Install Air
```bash
# Install latest version
go install github.com/air-verse/air@latest

# Verify installation
air -v
```

### Add to PATH (if needed)
```bash
# Windows PowerShell
$env:PATH += ";$env:GOPATH\bin"

# Linux/Mac
export PATH=$PATH:$(go env GOPATH)/bin
```

## ⚙️ Configuration

### Current `.air.toml` Setup
```toml
# .air.toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main.exe ./cmd/api"
  bin = "tmp/main.exe"
  include_ext = ["go"]
  exclude_dir = ["tmp"]
  exclude_file = []
  exclude_regex = []
  delay = 1000
  log = "air.log"
  send_interrupt = true
  kill_delay = 1000

[log]
  color = "true"
  timestamp = "true"

[screen]
  clear_on_rebuild = true
```

### What Air Watches
- ✅ **Included**: All `.go` files
- ❌ **Excluded**: `tmp/` directory, build artifacts
- 🔄 **Auto-rebuild**: On any Go file change
- 📝 **Logging**: Build output and errors

## 🚀 Usage

### Start Air
```bash
# Standard start
air

# With custom config
air -c .air.toml

# Debug mode
air -d
```

### Air Output Example
```
  __    _   ___
 / /\  | | | |_)
/_/--\ |_| |_| \_ v1.62.0, built with Go go1.24.4

watching .
watching cmd
watching cmd\api
watching internal
watching internal\api
watching internal\clickhouse
!exclude tmp
building...
running...
2025/06/29 20:30:00 Starting API server at :8081...
```

## 🔄 Development Workflow

### 1. Start Air
```bash
air
```

### 2. Make Changes
Edit any `.go` file and save. Air will:
1. 🔍 Detect file change
2. 🔨 Build new binary
3. 🔄 Restart application
4. 📝 Show build results

### 3. Watch Console
```
main.go has changed
building...
running...
2025/06/29 20:30:15 Starting API server at :8081...
```

## 🎛️ Customization

### Advanced Configuration
```toml
[build]
  # Build command
  cmd = "go build -o ./tmp/main.exe ./cmd/api"
  
  # Output binary
  bin = "tmp/main.exe"
  
  # Watch file extensions
  include_ext = ["go", "html", "css", "js"]
  
  # Exclude directories
  exclude_dir = ["tmp", "vendor", "node_modules", ".git"]
  
  # Exclude files
  exclude_file = ["main_test.go"]
  
  # Exclude regex patterns
  exclude_regex = [".*_test\\.go$"]
  
  # Delay before rebuild (ms)
  delay = 1000
  
  # Build arguments
  args_bin = ["-env", "development"]
  
  # Environment variables
  env = ["GO_ENV=development"]
  
  # Build with race detector
  race = true
  
  # Enable Go modules
  mod = "vendor"

[log]
  # Colored output
  color = true
  
  # Timestamp logs
  timestamp = true
  
  # Log level
  level = "info"

[misc]
  # Clear screen on rebuild
  clear_on_rebuild = true
  
  # Show app pid
  show_pid = true
```

### Environment-Specific Configs

**Development (`.air.dev.toml`):**
```toml
[build]
  cmd = "go build -tags=dev -o ./tmp/main.exe ./cmd/api"
  env = ["GO_ENV=development", "LOG_LEVEL=debug"]
```

**Testing (`.air.test.toml`):**
```toml
[build]
  cmd = "go build -tags=test -o ./tmp/main.exe ./cmd/api"
  env = ["GO_ENV=test", "LOG_LEVEL=debug"]
  race = true
```

## 🐛 Troubleshooting

### Common Issues

**1. Air Command Not Found**
```
'air' is not recognized as an internal or external command
```
**Solutions:**
- Reinstall: `go install github.com/air-verse/air@latest`
- Check PATH: `echo $env:PATH` (Windows) or `echo $PATH` (Linux/Mac)
- Use full path: `$env:GOPATH\bin\air.exe`

**2. Build Errors**
```
building...
# command-line-arguments
./main.go:10:2: package not found
failed to build, error: exit status 1
```
**Solutions:**
- Run `go mod tidy`
- Check import paths
- Verify Go version compatibility

**3. Port Already in Use**
```
listen tcp :8081: bind: Only one usage of each socket address
```
**Solutions:**
- Kill existing process: `netstat -ano | findstr :8081`
- Change port in `.env`: `SERVER_ADDR=:8082`
- Stop other Air instances

**4. Permission Denied**
```
mkdir tmp: permission denied
```
**Solutions:**
- Run as administrator (Windows)
- Check directory permissions
- Change `tmp_dir` in `.air.toml`

**5. Files Not Being Watched**
```
# Changes not triggering rebuild
```
**Solutions:**
- Check `include_ext` includes `.go`
- Verify file is not in `exclude_dir`
- Check for syntax errors in `.air.toml`

### Debug Air Issues
```bash
# Debug mode
air -d

# Check config
air -c .air.toml -d

# Manual test
go build -o tmp/test.exe ./cmd/api
./tmp/test.exe
```

## 🔧 Integration with IDEs

### VS Code
Add to `.vscode/tasks.json`:
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Start Air",
      "type": "shell",
      "command": "air",
      "group": "build",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": false,
        "panel": "new"
      },
      "problemMatcher": []
    }
  ]
}
```

### GoLand/IntelliJ
1. Run Configuration → Go Build
2. Command: `air`
3. Working Directory: Project root

## ⚡ Performance Tips

### Faster Builds
```toml
[build]
  # Use Go cache
  cmd = "go build -o ./tmp/main.exe ./cmd/api"
  
  # Reduce delay
  delay = 500
  
  # Exclude test files
  exclude_regex = [".*_test\\.go$"]
```

### Memory Usage
```toml
[build]
  # Limit concurrent builds
  delay = 1000
  
  # Clean tmp directory
  pre_cmd = ["rm -rf tmp/*"]
```

## 📊 Monitoring Air

### Log Files
```bash
# View Air logs
tail -f air.log

# View build errors
tail -f tmp/build-errors.log
```

### Process Monitoring
```bash
# Check Air process
ps aux | grep air

# Check application process  
ps aux | grep main.exe
```

## 🚀 Production Notes

### Don't Use Air in Production
Air is for development only. For production:

```bash
# Build optimized binary
go build -ldflags="-s -w" -o storage-control-plane ./cmd/api

# Run production binary
./storage-control-plane
```

### CI/CD Integration
```yaml
# GitHub Actions - don't use Air
- name: Build
  run: go build ./cmd/api

# Use Air only for development containers
- name: Dev Environment
  run: air
  if: github.ref == 'refs/heads/develop'
```

## 🎯 Air Best Practices

1. **Keep `.air.toml` simple** - Only customize what you need
2. **Exclude test files** - Faster builds
3. **Use environment variables** - Different configs per environment
4. **Monitor logs** - Watch for build issues
5. **Clean tmp directory** - Avoid stale binaries
6. **Use proper exclusions** - Don't watch vendor/, .git/, etc.

---

**Air Setup Complete! Enjoy hot reload development! 🔥🚀**
