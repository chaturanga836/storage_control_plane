#!/bin/bash

# Deploy Go Control Plane with Nginx Reverse Proxy
# This script sets up nginx as a reverse proxy for the Go microservices

set -e

echo "🚀 Deploying Go Control Plane with Nginx Reverse Proxy"
echo "======================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NGINX_CONF_DIR="/etc/nginx/sites-available"
NGINX_ENABLED_DIR="/etc/nginx/sites-enabled"
SITE_NAME="storage-control-plane"
DOMAIN_NAME=${1:-"yourdomain.com"}

echo -e "${YELLOW}Domain: ${DOMAIN_NAME}${NC}"
echo

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install nginx if not present
if ! command_exists nginx; then
    echo -e "${YELLOW}Installing nginx...${NC}"
    sudo apt update
    sudo apt install -y nginx
    echo -e "${GREEN}✅ Nginx installed${NC}"
else
    echo -e "${GREEN}✅ Nginx already installed${NC}"
fi

# Stop nginx temporarily
echo -e "${YELLOW}Stopping nginx...${NC}"
sudo systemctl stop nginx

# Backup existing default site
if [ -f "${NGINX_ENABLED_DIR}/default" ]; then
    echo -e "${YELLOW}Backing up default nginx site...${NC}"
    sudo mv "${NGINX_ENABLED_DIR}/default" "${NGINX_ENABLED_DIR}/default.backup"
fi

# Create nginx configuration
echo -e "${YELLOW}Creating nginx configuration...${NC}"
sudo tee "${NGINX_CONF_DIR}/${SITE_NAME}" > /dev/null << EOF
# Go Storage Control Plane - Nginx Configuration
# Generated on $(date)

upstream go_control_plane {
    # Health check configuration
    least_conn;
    
    # Auth Gateway
    server localhost:8090 max_fails=3 fail_timeout=30s;
}

upstream go_auth_gateway {
    server localhost:8090 max_fails=3 fail_timeout=30s;
}

upstream go_tenant_node {
    server localhost:8000 max_fails=3 fail_timeout=30s;
}

upstream go_operation_node {
    server localhost:8081 max_fails=3 fail_timeout=30s;
}

upstream go_cbo_engine {
    server localhost:8082 max_fails=3 fail_timeout=30s;
}

upstream go_metadata_catalog {
    server localhost:8083 max_fails=3 fail_timeout=30s;
}

upstream go_monitoring {
    server localhost:8084 max_fails=3 fail_timeout=30s;
}

upstream go_query_interpreter {
    server localhost:8085 max_fails=3 fail_timeout=30s;
}

# Rate limiting
limit_req_zone \$binary_remote_addr zone=api_limit:10m rate=10r/s;

server {
    listen 80;
    server_name ${DOMAIN_NAME};
    
    # Logging
    access_log /var/log/nginx/storage-control-plane.access.log;
    error_log /var/log/nginx/storage-control-plane.error.log;
    
    # Rate limiting
    limit_req zone=api_limit burst=20 nodelay;
    
    # Common proxy headers
    proxy_set_header Host \$host;
    proxy_set_header X-Real-IP \$remote_addr;
    proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto \$scheme;
    proxy_set_header X-Forwarded-Host \$server_name;
    
    # Timeout configurations
    proxy_connect_timeout 30s;
    proxy_send_timeout 30s;
    proxy_read_timeout 30s;
    proxy_buffering off;

    # 🔐 Auth Gateway (Port 8090)
    location /auth/ {
        proxy_pass http://go_auth_gateway/;
        rewrite ^/auth/(.*)$ /\$1 break;
    }

    # 🏢 Tenant Node (Port 8000)
    location /data/ {
        proxy_pass http://go_tenant_node/;
        rewrite ^/data/(.*)$ /\$1 break;
    }

    # 🎯 Operation Node (Port 8081)
    location /query/ {
        proxy_pass http://go_operation_node/;
        rewrite ^/query/(.*)$ /\$1 break;
    }

    # 🧠 CBO Engine (Port 8082)
    location /optimize/ {
        proxy_pass http://go_cbo_engine/;
        rewrite ^/optimize/(.*)$ /\$1 break;
    }

    # 📊 Metadata Catalog (Port 8083)
    location /metadata/ {
        proxy_pass http://go_metadata_catalog/;
        rewrite ^/metadata/(.*)$ /\$1 break;
    }

    # 📈 Monitoring (Port 8084)
    location /monitor/ {
        proxy_pass http://go_monitoring/;
        rewrite ^/monitor/(.*)$ /\$1 break;
    }

    # 🔍 Query Interpreter (Port 8085)
    location /parse/ {
        proxy_pass http://go_query_interpreter/;
        rewrite ^/parse/(.*)$ /\$1 break;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://go_auth_gateway/health;
        access_log off;
    }

    # Version endpoint
    location /version {
        proxy_pass http://go_auth_gateway/version;
    }

    # Admin/Dashboard
    location /admin/ {
        proxy_pass http://go_monitoring/;
        rewrite ^/admin/(.*)$ /\$1 break;
    }

    # Status page for nginx
    location /nginx-status {
        stub_status on;
        access_log off;
        allow 127.0.0.1;
        deny all;
    }

    # Default route
    location / {
        return 404 "Go Storage Control Plane - Service not found";
        add_header Content-Type text/plain;
    }
}
EOF

# Enable the site
echo -e "${YELLOW}Enabling nginx site...${NC}"
sudo ln -sf "${NGINX_CONF_DIR}/${SITE_NAME}" "${NGINX_ENABLED_DIR}/${SITE_NAME}"

# Test nginx configuration
echo -e "${YELLOW}Testing nginx configuration...${NC}"
if sudo nginx -t; then
    echo -e "${GREEN}✅ Nginx configuration is valid${NC}"
else
    echo -e "${RED}❌ Nginx configuration error${NC}"
    exit 1
fi

# Start nginx
echo -e "${YELLOW}Starting nginx...${NC}"
sudo systemctl start nginx
sudo systemctl enable nginx

echo -e "${GREEN}✅ Nginx started and enabled${NC}"

# Create monitoring script
echo -e "${YELLOW}Creating monitoring script...${NC}"
cat > monitor_services.sh << 'EOF'
#!/bin/bash

# Monitor Go Control Plane Services
echo "🔍 Go Storage Control Plane Service Status"
echo "=========================================="

services=(
    "8090:Auth Gateway"
    "8000:Tenant Node"
    "8081:Operation Node"
    "8082:CBO Engine"
    "8083:Metadata Catalog"
    "8084:Monitoring"
    "8085:Query Interpreter"
)

for service in "${services[@]}"; do
    port="${service%%:*}"
    name="${service##*:}"
    
    if curl -s "http://localhost:${port}/health" > /dev/null; then
        echo "✅ ${name} (${port}) - Healthy"
    else
        echo "❌ ${name} (${port}) - Unhealthy"
    fi
done

echo
echo "🌐 Nginx Status:"
if systemctl is-active --quiet nginx; then
    echo "✅ Nginx - Running"
else
    echo "❌ Nginx - Stopped"
fi

echo
echo "🔗 Access URLs:"
echo "  Health Check: http://${DOMAIN_NAME}/health"
echo "  Auth:         http://${DOMAIN_NAME}/auth/"
echo "  Data:         http://${DOMAIN_NAME}/data/"
echo "  Query:        http://${DOMAIN_NAME}/query/"
echo "  Optimize:     http://${DOMAIN_NAME}/optimize/"
echo "  Metadata:     http://${DOMAIN_NAME}/metadata/"
echo "  Monitor:      http://${DOMAIN_NAME}/monitor/"
echo "  Parse:        http://${DOMAIN_NAME}/parse/"
EOF

chmod +x monitor_services.sh

echo
echo -e "${GREEN}🎉 Deployment Complete!${NC}"
echo "========================"
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Start the Go Control Plane: go run ."
echo "2. Test services: ./monitor_services.sh"
echo "3. Access via: http://${DOMAIN_NAME}/"
echo
echo -e "${YELLOW}Service URLs:${NC}"
echo "  Health: http://${DOMAIN_NAME}/health"
echo "  Auth:   http://${DOMAIN_NAME}/auth/"
echo "  Monitor: http://${DOMAIN_NAME}/monitor/"
echo
echo -e "${YELLOW}Logs:${NC}"
echo "  Nginx Access: sudo tail -f /var/log/nginx/storage-control-plane.access.log"
echo "  Nginx Error:  sudo tail -f /var/log/nginx/storage-control-plane.error.log"
