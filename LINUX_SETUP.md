# Linux Setup Guide

This guide covers setup and development on Linux distributions (Ubuntu, Fedora, CentOS, etc.) and WSL (Windows Subsystem for Linux).

## Quick Start

```bash
# Clone and setup
git clone <repository-url>
cd storage_control_plane

# Setup development environment
make setup
make install-tools

# Start development
make dev-linux
```

## Prerequisites

### 1. Go Installation

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install golang-go
```

**Fedora/RHEL/CentOS:**
```bash
sudo dnf install golang
# or for older versions:
sudo yum install golang
```

**Arch Linux:**
```bash
sudo pacman -S go
```

**Manual Installation:**
```bash
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. Build Tools
```bash
# Ubuntu/Debian
sudo apt install build-essential

# Fedora/RHEL/CentOS
sudo dnf groupinstall "Development Tools"
# or
sudo yum groupinstall "Development Tools"

# Arch Linux
sudo pacman -S base-devel
```

### 3. Git
```bash
# Ubuntu/Debian
sudo apt install git

# Fedora/RHEL/CentOS
sudo dnf install git

# Arch Linux
sudo pacman -S git
```

## Development Environment Setup

### 1. One-Command Setup
```bash
make setup install-tools
```

### 2. Manual Setup
```bash
# Create directories
mkdir -p data/rocksdb data/parquet data/wal tmp bin

# Copy environment file
cp .env.example .env

# Install Go dependencies
go mod download

# Install development tools
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
```

### 3. Environment Variables
Edit `.env` file with your settings:
```bash
nano .env
# or
vim .env
# or
code .env  # if VS Code is installed
```

## Development Workflow

### Option 1: Using Make Commands
```bash
# Start with hot reload
make dev

# Run tests
make test

# Run E2E tests
make test-e2e

# Build application
make build

# Clean artifacts
make clean
```

### Option 2: Using dev.sh Script
```bash
# Make executable and run
chmod +x dev.sh
./dev.sh

# Or use the make target
make dev-linux
```

### Option 3: Manual Commands
```bash
# Hot reload development
air

# Or build and run
go build -o tmp/main cmd/api/main.go
./tmp/main

# Run tests
go test -v ./...
```

## Testing

### Unit Tests
```bash
go test -v ./internal/... ./pkg/...
# or
make test-unit
```

### Integration Tests
```bash
# Ensure services are running first
make test-integration
```

### End-to-End Tests
```bash
# Start server in one terminal
make dev

# Run E2E tests in another terminal
chmod +x test_e2e.sh
./test_e2e.sh
# or
make test-e2e
```

## Dependencies

### External Services

**ClickHouse (using Docker):**
```bash
docker run -d \
  --name clickhouse-server \
  -p 8123:8123 \
  -p 9000:9000 \
  clickhouse/clickhouse-server
```

**RocksDB:**
```bash
# Ubuntu/Debian
sudo apt install librocksdb-dev

# Fedora/RHEL/CentOS
sudo dnf install rocksdb-devel

# Build from source if needed
git clone https://github.com/facebook/rocksdb.git
cd rocksdb
make shared_lib
sudo make install-shared
```

## Common Issues & Solutions

### 1. Go Not in PATH
```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$(go env GOPATH)/bin
source ~/.bashrc
```

### 2. Air Not Found
```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Add GOPATH/bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### 3. Permission Denied for Scripts
```bash
chmod +x dev.sh test_e2e.sh
```

### 4. RocksDB Build Issues
```bash
# Install CGO dependencies
sudo apt install gcc g++  # Ubuntu/Debian
sudo dnf install gcc gcc-c++  # Fedora/RHEL

# Set CGO flags if needed
export CGO_CFLAGS="-I/usr/local/include/rocksdb"
export CGO_LDFLAGS="-L/usr/local/lib -lrocksdb"
```

### 5. WSL-Specific Issues

**Docker Integration:**
- Enable Docker Desktop WSL 2 integration
- Install Docker inside WSL: `curl -fsSL https://get.docker.com | sudo sh`

**File Permissions:**
```bash
# If working in Windows filesystem from WSL
cd /mnt/c/your/project  # Avoid this
cd ~/your/project       # Prefer this (Linux filesystem)
```

**Memory/Performance:**
```bash
# Configure WSL memory in ~/.wslconfig
[wsl2]
memory=4GB
processors=4
```

## IDE Setup

### VS Code
```bash
# Install VS Code
wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > packages.microsoft.gpg
sudo install -o root -g root -m 644 packages.microsoft.gpg /etc/apt/trusted.gpg.d/
sudo sh -c 'echo "deb [arch=amd64,arm64,armhf signed-by=/etc/apt/trusted.gpg.d/packages.microsoft.gpg] https://packages.microsoft.com/repos/code stable main" > /etc/apt/sources.list.d/vscode.list'
sudo apt update
sudo apt install code

# Install Go extension
code --install-extension golang.go
```

### Vim/Neovim
```bash
# Install vim-go plugin
git clone https://github.com/fatih/vim-go.git ~/.vim/pack/plugins/start/vim-go

# For Neovim with LSP
# Use nvim-lspconfig with gopls
```

## Performance Tips

### 1. Use Go Module Proxy
```bash
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org
```

### 2. Enable Build Cache
```bash
export GOCACHE=/home/$USER/.cache/go-build
```

### 3. Parallel Builds
```bash
export GOMAXPROCS=$(nproc)
```

## Monitoring & Debugging

### 1. Application Logs
```bash
# Follow logs in real-time
tail -f tmp/air.log

# Search logs
grep "ERROR" tmp/air.log
```

### 2. System Monitoring
```bash
# Monitor resource usage
htop
# or
top

# Monitor disk usage
df -h
du -sh data/
```

### 3. Debugging Tools
```bash
# Install Delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug with Delve
dlv debug cmd/api/main.go
```

## Deployment

### 1. Build for Production
```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/storage-control-plane cmd/api/main.go

# Cross-compile for multiple platforms
make build-all
```

### 2. Systemd Service
```bash
# Create service file
sudo nano /etc/systemd/system/storage-control-plane.service

# Enable and start
sudo systemctl enable storage-control-plane
sudo systemctl start storage-control-plane
```

## Support

- **Documentation:** See `DOCUMENTATION_INDEX.md`
- **Windows:** See `WINDOWS_SETUP.md`
- **Testing:** See `TESTING.md`
- **Air Hot Reload:** See `AIR_SETUP.md`

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make changes and test: `make test`
4. Commit: `git commit -m "Add feature"`
5. Push: `git push origin feature-name`
6. Create a Pull Request
