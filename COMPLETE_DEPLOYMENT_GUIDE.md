# 🚀 GO CONTROL PLANE DEPLOYMENT GUIDE

## 📍 **System Overview**
- **Go Control Plane EC2**: `15.207.184.150` (Go binary service)
- **Python Services EC2**: `65.0.150.75` (Docker microservices)

---

## 🚀 **GO CONTROL PLANE DEPLOYMENT**

### **Step 1: Verify Environment Configuration**
```bash
# Navigate to storage control plane
cd /opt/storage_control_plane

# Verify .env file is properly configured:
cat .env
```

**Expected .env content:**
```properties
# Go Control Plane Configuration
PORT=8090
ENVIRONMENT=production
LOG_LEVEL=info

# Python Services Configuration  
PYTHON_IP=65.0.150.75
AUTH_GATEWAY_URL=http://65.0.150.75:8080
TENANT_NODE_URL=http://65.0.150.75:8001
# ... other endpoints

# Distributed Mode
DISTRIBUTED_MODE=true
PYTHON_SERVICES_HOST=65.0.150.75
GO_SERVICES_HOST=15.207.184.150
```

### **Step 2: Deploy Go Control Plane**
```bash
# Method 1: Use the deployment script (if not already deployed)
./deploy_ec2.sh

# Method 2: Manual deployment (if already built)
# Restart service to pick up new .env configuration
sudo systemctl restart storage-control-plane

# Check service status
sudo systemctl status storage-control-plane

# Check logs
sudo journalctl -u storage-control-plane -f
```

### **Step 3: Test Go Control Plane**
```bash
# Test local health endpoint
curl http://localhost:8090/health

# Test external access (from your local machine)
curl http://15.207.184.150:8090/health
```

---

## 🔗 **CONNECTIVITY TEST**

### **Test connectivity to Python services:**
```bash
# Test connectivity to Python services
chmod +x test_distributed_connectivity.sh
./test_distributed_connectivity.sh
```

### **Manual Tests:**
```bash
# From Go EC2, test Python services
curl http://65.0.150.75:8080/health
curl http://65.0.150.75:8001/health
curl http://65.0.150.75:8087/health
```

---

## 🔒 **SECURITY GROUP CONFIGURATION**

### **Go EC2 (15.207.184.150) Security Group:**
```
Type        Protocol    Port Range    Source
HTTP        TCP         8090         0.0.0.0/0          # Public → Go Control Plane
SSH         TCP         22           YOUR_IP/32         # SSH access
```

---

## 🧪 **VERIFICATION CHECKLIST**

### **✅ Go Control Plane (15.207.184.150)**
- [ ] Service running (`sudo systemctl status storage-control-plane`)
- [ ] Health endpoint responding (`curl http://localhost:8090/health`)
- [ ] External access working (`curl http://15.207.184.150:8090/health`)
- [ ] Can reach Python services (`./test_distributed_connectivity.sh`)

---

## 🌐 **ACCESS URLS**

### **Go Control Plane:**
- **Public Access**: `http://15.207.184.150:8090`
- **Health Endpoint**: `http://15.207.184.150:8090/health`

### **System Management:**
```bash
# Go Control Plane  
sudo systemctl status storage-control-plane
sudo journalctl -u storage-control-plane -f
sudo systemctl restart storage-control-plane
```

## 🎉 **SUCCESS INDICATORS**

When working correctly, you should see:

1. **Go Control Plane responding**: `{"status":"healthy","timestamp":"..."}`
2. **Can communicate with Python services**: Cross-system connectivity working
3. **External access working**: Control plane accessible from internet

Your Go Control Plane is now fully deployed! 🚀
