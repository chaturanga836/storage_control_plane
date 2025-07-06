#!/bin/bash
# Go Control Plane Deployment Script for EC2

echo "🚀 GO CONTROL PLANE DEPLOYMENT SCRIPT"
echo "====================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Check if we're on the right system
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    print_error "This script is designed for Linux (EC2). Please run on your EC2 instance."
    exit 1
fi

print_header "1. Checking Go installation..."
if ! command -v go &> /dev/null; then
    print_warning "Go not found. Installing Go 1.21..."
    
    # Install Go
    cd /tmp
    wget -q https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
    
    if [ -f "go1.21.6.linux-amd64.tar.gz" ]; then
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
        
        # Add to PATH
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        echo 'export GOPATH=$HOME/go' >> ~/.bashrc
        echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
        
        export PATH=$PATH:/usr/local/go/bin
        export GOPATH=$HOME/go
        export PATH=$PATH:$GOPATH/bin
        
        print_status "Go installed successfully"
        go version
    else
        print_error "Failed to download Go. Please install manually."
        exit 1
    fi
else
    print_status "Go is already installed: $(go version)"
fi

print_header "2. Setting up project directory..."

# Check if we're already in the storage_control_plane directory
if [ "$(basename $(pwd))" = "storage_control_plane" ]; then
    print_status "Already in storage_control_plane directory"
elif [ -d "storage_control_plane" ]; then
    cd storage_control_plane
    print_status "Found existing storage_control_plane directory"
else
    # Try to find storage_control_plane directory in home
    cd $HOME
    if [ -d "storage_control_plane" ]; then
        cd storage_control_plane
        print_status "Found storage_control_plane directory in home"
    else
        print_error "storage_control_plane directory not found. Please clone the repository first:"
        echo "  git clone https://github.com/chaturanga836/storage_control_plane.git"
        echo "  cd storage_control_plane"
        exit 1
    fi
fi

print_status "Working in: $(pwd)"

print_header "3. Installing dependencies..."
if [ -f "go.mod" ]; then
    go mod download
    print_status "Dependencies downloaded"
else
    print_warning "go.mod not found. Initializing module..."
    go mod init storage_control_plane
    
    # Add common dependencies
    go get github.com/joho/godotenv
    go get github.com/gorilla/mux
    go get github.com/rs/cors
    print_status "Module initialized with basic dependencies"
fi

print_header "4. Setting up environment..."
if [ ! -f ".env" ]; then
    if [ -f ".env.example" ]; then
        cp .env.example .env
        print_status "Environment file created from template"
    else
        # Create comprehensive .env file
        echo "Please provide your Python services EC2 IP address (or press Enter for localhost):"
        read -r PYTHON_IP
        if [ -z "$PYTHON_IP" ]; then
            PYTHON_IP="localhost"
        fi
        
        cat > .env << EOF
# Go Control Plane Configuration
PORT=8090
ENVIRONMENT=production
LOG_LEVEL=info

# Python Services Configuration
PYTHON_IP=${PYTHON_IP}

# Python Service Endpoints
AUTH_GATEWAY_URL=http://${PYTHON_IP}:8080
TENANT_NODE_URL=http://${PYTHON_IP}:8001
METADATA_CATALOG_URL=http://${PYTHON_IP}:8087
OPERATION_NODE_URL=http://${PYTHON_IP}:8086
CBO_ENGINE_URL=http://${PYTHON_IP}:8088
MONITORING_URL=http://${PYTHON_IP}:8089
QUERY_INTERPRETER_URL=http://${PYTHON_IP}:8085

# Distributed Mode
DISTRIBUTED_MODE=true
PYTHON_SERVICES_HOST=${PYTHON_IP}
GO_SERVICES_HOST=15.207.184.150

# Health Check Settings
HEALTH_CHECK_INTERVAL=30
SERVICE_TIMEOUT=10
RETRY_ATTEMPTS=3
EOF
        print_status "Environment file created with Python IP: $PYTHON_IP"
    fi
else
    print_status "Environment file already exists"
fi

print_header "5. Building Go application..."

# Fix: Exclude test files from main build by building only main package files
# or rename problematic test files
if [ -f "functional_test_suite.go" ]; then
    print_status "Renaming test files to prevent build conflicts..."
    mv functional_test_suite.go functional_test_suite_test.go 2>/dev/null || true
fi

# Build only the main application, excluding test files
if go build -o storage-control-plane main.go services.go services_extended.go; then
    print_status "Build successful: storage-control-plane binary created"
    ls -la storage-control-plane
else
    print_warning "Specific file build failed, trying alternative build..."
    # Alternative: build with exclusions
    if go build -o storage-control-plane .; then
        print_status "Build successful: storage-control-plane binary created"
        ls -la storage-control-plane
    else
        print_error "Build failed. Please check the error messages above."
        print_error "Common issues:"
        print_error "- Test files mixed with main package"
        print_error "- Missing imports or undefined functions"
        exit 1
    fi
fi

print_header "6. Setting up systemd service..."
CURRENT_DIR=$(pwd)
sudo tee /etc/systemd/system/storage-control-plane.service > /dev/null <<EOF
[Unit]
Description=Storage Control Plane
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$CURRENT_DIR
ExecStart=$CURRENT_DIR/storage-control-plane
Restart=always
RestartSec=10
Environment=PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
EnvironmentFile=$CURRENT_DIR/.env

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable storage-control-plane
print_status "Systemd service configured"

print_header "7. Starting service..."
sudo systemctl start storage-control-plane

# Wait a moment for service to start
sleep 3

if sudo systemctl is-active --quiet storage-control-plane; then
    print_status "Service started successfully"
else
    print_error "Service failed to start. Checking logs..."
    sudo journalctl -u storage-control-plane --no-pager -l
    exit 1
fi

print_header "8. Running health checks..."
sleep 2

# Function to check endpoint
check_endpoint() {
    local url=$1
    local name=$2
    
    # Fix URL formatting
    if [[ "$url" != http://* ]]; then
        url="http://$url"
    fi
    
    if curl -s --connect-timeout 5 "$url" > /dev/null; then
        print_status "✅ $name is healthy"
        return 0
    else
        print_warning "❌ $name is not responding ($url)"
        return 1
    fi
}

# Check all endpoints
healthy_count=0
total_endpoints=1

# Check if our Go control plane service is responding
print_status "Checking Go control plane service..."
if check_endpoint "http://localhost:8090/health" "Go Control Plane"; then
    ((healthy_count++))
    print_status "✅ Go control plane is healthy!"
else
    print_warning "❌ Go control plane not responding on port 8090"
    # Try alternative endpoint
    if check_endpoint "http://localhost:8090/" "Go Control Plane Root"; then
        print_status "✅ Go control plane root endpoint responding"
        ((healthy_count++))
    fi
fi

print_header "9. Deployment Summary"
echo "======================================"
echo "Service Status: $(sudo systemctl is-active storage-control-plane)"
echo "Service Enabled: $(sudo systemctl is-enabled storage-control-plane)"
echo "Healthy Endpoints: $healthy_count/$total_endpoints"
echo "Process ID: $(pgrep -f storage-control-plane || echo 'Not running')"
echo "Memory Usage: $(ps -o pid,vsz,rss,comm -p $(pgrep -f storage-control-plane) 2>/dev/null | tail -1 || echo 'N/A')"

# Get public IP - handle metadata service issues gracefully
PUBLIC_IP=$(timeout 5 curl -s http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null | head -1 || echo "unknown")
PRIVATE_IP=$(timeout 5 curl -s http://169.254.169.254/latest/meta-data/local-ipv4 2>/dev/null | head -1 || echo "unknown")

# Clean up any HTML error responses
if [[ "$PUBLIC_IP" == *"<"* ]]; then
    PUBLIC_IP="unknown"
fi
if [[ "$PRIVATE_IP" == *"<"* ]]; then
    PRIVATE_IP="unknown"
fi

echo ""
echo "🌐 Access URLs:"
echo "  Public:  http://$PUBLIC_IP:8090"
echo "  Private: http://$PRIVATE_IP:8090"
echo "  Local:   http://localhost:8090"

if [ $healthy_count -eq $total_endpoints ]; then
    echo ""
    echo "🎉 DEPLOYMENT SUCCESSFUL! 🎉"
    echo "Your Go Control Plane is fully operational!"
    echo ""
    echo "📋 Next Steps:"
    echo "1. Update security groups to allow access on port 8090"
    echo "2. Test endpoints: curl http://localhost:8090/health"
    echo "3. Monitor logs: sudo journalctl -u storage-control-plane -f"
    echo "4. Configure connection to Python services in .env file (if on separate instance)"
else
    echo ""
    echo "⚠️  PARTIAL DEPLOYMENT"
    echo "Some endpoints are not responding. Check logs:"
    echo "  sudo journalctl -u storage-control-plane -f"
fi

echo ""
echo "📊 Service Management Commands:"
echo "  Start:   sudo systemctl start storage-control-plane"
echo "  Stop:    sudo systemctl stop storage-control-plane"
echo "  Restart: sudo systemctl restart storage-control-plane"
echo "  Status:  sudo systemctl status storage-control-plane"
echo "  Logs:    sudo journalctl -u storage-control-plane -f"
