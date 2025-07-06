# 🚀 GO CONTROL PLANE - EC2 DEPLOYMENT GUIDE

## Prerequisites for EC2 Instance
- EC2 instance (Ubuntu 20.04+ or Amazon Linux 2)
- Go 1.21+ installed
- Git access to your repository
- At least 4GB RAM and 10GB free disk space

## Quick Start - Automated Deployment

### Option A: Use the Automated Script (RECOMMENDED)

```bash
# SSH to your EC2 instance
ssh -i your-key.pem ubuntu@your-ec2-ip

# Clone the repository
git clone https://github.com/chaturanga836/storage_control_plane.git
cd storage_control_plane

# Make the deployment script executable
chmod +x deploy_ec2.sh

# Run the automated deployment
./deploy_ec2.sh
```

**The script will automatically:**
- Install Go 1.21 if not present
- Download dependencies
- Create environment configuration
- Build the Go application
- Set up systemd service
- Start the service and run health checks

**Note:** The script can be run from inside the `storage_control_plane` directory or from its parent directory.

---

## Manual Deployment (Alternative)

If you prefer to deploy manually or need to customize the setup, follow these steps:

## Step 1: Setup EC2 Instance

### Launch and Configure EC2 Instance
```bash
# SSH to your EC2 instance
ssh -i your-key.pem ubuntu@your-ec2-ip

# Update system
sudo apt update && sudo apt upgrade -y

# Install Git
sudo apt install git -y
```

### Install Go
```bash
# Download and install Go 1.21
cd /tmp
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

### Install Air for Hot Reload (Optional but Recommended)
```bash
go install github.com/air-verse/air@latest
```

## Step 2: Clone and Setup Go Control Plane

```bash
# Clone your Go control plane repository
git clone https://github.com/chaturanga836/storage_control_plane.git
cd storage_control_plane

# Download Go dependencies
go mod download

# Copy environment template (create if doesn't exist)
if [ -f ".env.example" ]; then
    cp .env.example .env
else
    # Create basic .env file
    cat > .env << 'EOF'
# Go Control Plane Configuration
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info
EOF
fi
```

## Step 3: Configure Environment

Edit the `.env` file to configure your Go control plane:

```bash
nano .env
```

```env
# Go Control Plane Configuration
PORT=8080
ENVIRONMENT=production

# Database Configuration (optional - can use in-memory for demo)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=storage_control
DB_USER=postgres
DB_PASSWORD=your_secure_password

# Redis Configuration (optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Service Discovery (connect to Python microservices if on different instance)
PYTHON_SERVICES_HOST=your-python-ec2-ip
PYTHON_AUTH_GATEWAY=http://your-python-ec2-ip:8080
PYTHON_TENANT_NODE=http://your-python-ec2-ip:8001
PYTHON_METADATA_CATALOG=http://your-python-ec2-ip:8087

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## Step 4: Build and Deploy

### Option A: Quick Development Start
```bash
# Start with hot reload for development
air

# Server will start on http://0.0.0.0:8080
```

### Option B: Production Build and Run
```bash
# Build the binary
go build -o storage-control-plane .

# Run in background
nohup ./storage-control-plane > control-plane.log 2>&1 &

# Check it's running
ps aux | grep storage-control-plane
```

### Option C: Systemd Service (Recommended for Production)
```bash
# Create systemd service file
sudo tee /etc/systemd/system/storage-control-plane.service > /dev/null <<EOF
[Unit]
Description=Storage Control Plane
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/storage_control_plane
ExecStart=/home/ubuntu/storage_control_plane/storage-control-plane
Restart=always
RestartSec=10
Environment=PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
EnvironmentFile=/home/ubuntu/storage_control_plane/.env

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable storage-control-plane
sudo systemctl start storage-control-plane

# Check status
sudo systemctl status storage-control-plane
```

## Step 5: Verify Deployment

### Health Checks
```bash
# Check if service is running
curl http://localhost:8080/health

# Check all Go service endpoints
curl http://localhost:8080/auth/health     # Auth Gateway
curl http://localhost:8000/health          # Tenant Node
curl http://localhost:8081/health          # Operation Node
curl http://localhost:8082/health          # CBO Engine
curl http://localhost:8083/health          # Metadata Catalog
curl http://localhost:8084/health          # Monitoring
curl http://localhost:8085/health          # Query Interpreter
```

### Test Service Communication
```bash
# Test authentication
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "test_user", "password": "test_password"}'

# Test query execution
curl -X POST http://localhost:8081/query/execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your_token_here" \
  -d '{"query": "SELECT * FROM orders LIMIT 10"}'
```

## Step 6: Configure Security Groups

### Allow Traffic Between Instances
```bash
# On Python EC2 instance (if separate), allow Go instance access:
# Security Group Rules:
# - Allow inbound on ports 8080, 8001, 8087, 8088, 8086, 8089 from Go EC2 IP

# On Go EC2 instance, allow external access:
# Security Group Rules:
# - Allow inbound on port 8080 from 0.0.0.0/0 (or specific IPs)
# - Allow inbound on ports 8000-8085 from Python EC2 IP (if separate)
```

## Step 7: Test Cross-Service Communication

### Test Go Control Plane → Python Services (if separate instances)
```bash
# From Go EC2 instance, test connection to Python services
curl http://your-python-ec2-ip:8080/health  # Python auth-gateway
curl http://your-python-ec2-ip:8001/health  # Python tenant-node
curl http://your-python-ec2-ip:8087/health  # Python metadata-catalog
```

### Test Complete Workflow
```bash
# 1. Authenticate via Go control plane
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "test_user", "password": "test_password"}' | \
  jq -r '.token')

# 2. Execute query via Go that delegates to Python services
curl -X POST http://localhost:8081/query/execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query": "SELECT COUNT(*) FROM orders", "tenant_id": "tenant_001"}'
```

## ✅ **GO DEPLOYMENT SUCCESS CONFIRMATION**

🎉 **Congratulations!** Your Go control plane is now running!

### **Success Indicators:**
- ✅ **Go control plane running** on EC2 instance
- ✅ **All Go service endpoints responding** (ports 8080-8085)
- ✅ **Authentication and query flow operational**

### **Your Go Control Plane Status:**
```
🖥️  Go Control Plane Services:
├── Auth Gateway:     http://go-ec2-ip:8080
├── Tenant Node:      http://go-ec2-ip:8000
├── Operation Node:   http://go-ec2-ip:8081
├── CBO Engine:       http://go-ec2-ip:8082
├── Metadata Catalog: http://go-ec2-ip:8083
├── Monitoring:       http://go-ec2-ip:8084
└── Query Interpreter: http://go-ec2-ip:8085
```

## Troubleshooting

### Common Issues

#### 1. Go Binary Build Fails
```bash
# Check Go version
go version

# Clean module cache and rebuild
go clean -modcache
go mod download
go build -v .
```

#### 2. Service Won't Start
```bash
# Check logs
sudo journalctl -u storage-control-plane -f

# Check port conflicts
sudo netstat -tulpn | grep :8080

# Check permissions
ls -la storage-control-plane
chmod +x storage-control-plane
```

#### 3. Cross-Instance Communication Fails (if using separate instances)
```bash
# Test network connectivity
ping python-ec2-ip

# Test specific ports
telnet python-ec2-ip 8080

# Check security groups in AWS console
```

#### 4. Memory Issues
```bash
# Check memory usage
free -h

# Monitor Go process
top -p $(pgrep storage-control-plane)

# Optimize Go runtime (if needed)
export GOGC=100
export GOMEMLIMIT=1024MiB
```

---

## Troubleshooting Common Issues

### 1. Permission Denied Error
If you get `Permission denied` when running the script:
```bash
# Make the script executable
chmod +x deploy_ec2.sh

# Then run it
./deploy_ec2.sh
```

### 2. Git Ownership Issues
If you see `fatal: detected dubious ownership in repository`:
```bash
# Option 1: Add safe directory (recommended)
git config --global --add safe.directory /opt/storage_control_plane

# Option 2: Fix ownership (if switching between users)
sudo chown -R ubuntu:ubuntu /opt/storage_control_plane

# Option 3: If running as root and want to continue as root
chown -R root:root /opt/storage_control_plane
```

### 3. Go Installation Issues
If Go installation fails:
```bash
# Manual Go installation
cd /tmp
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

### 4. Build Failures
If the Go build fails:
```bash
# Clean and retry
go clean -cache
go mod download
go build -o storage-control-plane .
```

### 5. Service Startup Issues
If the systemd service fails to start:
```bash
# Check service status
sudo systemctl status storage-control-plane

# Check logs
sudo journalctl -u storage-control-plane -f

# Restart manually
sudo systemctl restart storage-control-plane
```

---

**🎯 Your Go control plane is now ready to run alongside or independently from your Python microservices!**
