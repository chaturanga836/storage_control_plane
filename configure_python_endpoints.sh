#!/bin/bash
# Configure Python Services Endpoints for Go Control Plane

echo "🔧 CONFIGURE PYTHON SERVICES ENDPOINTS"
echo "======================================"

echo ""
echo "This script will configure the Go Control Plane to connect to your Python services."
echo ""

# Get Python EC2 IP
echo "📍 Enter your Python services EC2 instance IP address:"
echo "   (This is the EC2 instance where your Python microservices are running)"
echo -n "Python EC2 IP: "
read -r PYTHON_IP

if [ -z "$PYTHON_IP" ]; then
    echo "❌ No IP address provided. Exiting."
    exit 1
fi

# Validate IP format (basic check)
if [[ ! $PYTHON_IP =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
    echo "⚠️  Warning: '$PYTHON_IP' doesn't look like a valid IP address."
    echo "   You can still continue if you're using a hostname."
    echo -n "Continue anyway? (y/N): "
    read -r confirm
    if [[ $confirm != "y" && $confirm != "Y" ]]; then
        exit 1
    fi
fi

echo ""
echo "🔧 Creating .env configuration file..."

# Create or update .env file
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

# Database Configuration (optional)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=storage_control
DB_USER=postgres
DB_PASSWORD=password

# Redis Configuration (optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Health Check Settings
HEALTH_CHECK_INTERVAL=30
SERVICE_TIMEOUT=10
RETRY_ATTEMPTS=3
EOF

echo "✅ Configuration file created successfully!"
echo ""
echo "📋 Configuration Summary:"
echo "  Python Services IP: $PYTHON_IP"
echo "  Go Control Plane Port: 8090"
echo ""
echo "🔗 Python Service URLs:"
echo "  • Auth Gateway:     http://${PYTHON_IP}:8080"
echo "  • Tenant Node:      http://${PYTHON_IP}:8001"
echo "  • Metadata Catalog: http://${PYTHON_IP}:8087"
echo "  • Operation Node:   http://${PYTHON_IP}:8086"
echo "  • CBO Engine:       http://${PYTHON_IP}:8088"
echo "  • Monitoring:       http://${PYTHON_IP}:8089"
echo "  • Query Interpreter:http://${PYTHON_IP}:8085"
echo ""
echo "🚀 Next steps:"
echo "1. Restart the Go Control Plane service:"
echo "   sudo systemctl restart storage-control-plane"
echo ""
echo "2. Check service status:"
echo "   sudo systemctl status storage-control-plane"
echo ""
echo "3. Test the health endpoint:"
echo "   curl http://localhost:8090/health"
echo ""
echo "4. View logs:"
echo "   sudo journalctl -u storage-control-plane -f"
echo ""

# Offer to restart service
echo -n "🔄 Would you like to restart the service now? (y/N): "
read -r restart_confirm
if [[ $restart_confirm == "y" || $restart_confirm == "Y" ]]; then
    echo "🔄 Restarting storage-control-plane service..."
    sudo systemctl restart storage-control-plane
    sleep 3
    
    if sudo systemctl is-active --quiet storage-control-plane; then
        echo "✅ Service restarted successfully!"
        echo "🧪 Testing health endpoint..."
        if curl -s http://localhost:8090/health > /dev/null; then
            echo "✅ Health check passed!"
        else
            echo "❌ Health check failed. Check logs: sudo journalctl -u storage-control-plane -f"
        fi
    else
        echo "❌ Service failed to restart. Check logs: sudo journalctl -u storage-control-plane -f"
    fi
fi

echo ""
echo "🎯 Your Go Control Plane is now configured to connect to Python services at $PYTHON_IP"
