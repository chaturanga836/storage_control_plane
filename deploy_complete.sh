#!/bin/bash

# Storage Control Plane - Complete Deployment Script
# Repository location: /opt/storage_control_plane

set -e

REPO_DIR="/opt/storage_control_plane"
SERVICE_NAME="storage-control-plane"
NGINX_CONF="$REPO_DIR/nginx.conf"
SYSTEMD_SERVICE="$REPO_DIR/$SERVICE_NAME.service"

echo "🚀 Complete Storage Control Plane Deployment"
echo "Repository: $REPO_DIR"
echo "============================================="

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo "❌ This script must be run as root (use sudo)"
   exit 1
fi

# Check if repository exists
if [[ ! -d "$REPO_DIR" ]]; then
    echo "❌ Repository not found at $REPO_DIR"
    echo "Please clone the repository to $REPO_DIR first"
    exit 1
fi

# Change to repository directory
cd "$REPO_DIR"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.24+ first"
    exit 1
fi

echo "📦 Installing system dependencies..."
apt-get update
apt-get install -y nginx curl netstat-nat

echo "🏗️  Building Go application..."
go mod download
go build -o storage-control-plane .

if [[ ! -f "storage-control-plane" ]]; then
    echo "❌ Failed to build Go application"
    exit 1
fi

echo "✅ Go application built successfully"

# Make the binary executable
chmod +x storage-control-plane

echo "🔧 Installing systemd service..."
cp "$SYSTEMD_SERVICE" "/etc/systemd/system/$SERVICE_NAME.service"
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"

echo "🌐 Setting up nginx reverse proxy..."
# Install nginx configuration
cp "$NGINX_CONF" "/etc/nginx/sites-available/$SERVICE_NAME"

# Remove default site if it exists
if [[ -f "/etc/nginx/sites-enabled/default" ]]; then
    rm -f "/etc/nginx/sites-enabled/default"
fi

# Enable the site
rm -f "/etc/nginx/sites-enabled/$SERVICE_NAME"
ln -s "/etc/nginx/sites-available/$SERVICE_NAME" "/etc/nginx/sites-enabled/$SERVICE_NAME"

# Test nginx configuration
if nginx -t; then
    echo "✅ Nginx configuration is valid"
else
    echo "❌ Nginx configuration test failed"
    exit 1
fi

echo "🚀 Starting services..."

# Start the Go service
systemctl start "$SERVICE_NAME"

# Wait a moment for the service to start
sleep 3

# Check if the service is running
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "✅ Storage Control Plane service is running"
else
    echo "❌ Failed to start Storage Control Plane service"
    echo "Check logs: sudo journalctl -u $SERVICE_NAME -f"
    exit 1
fi

# Start nginx
systemctl restart nginx

# Check nginx status
if systemctl is-active --quiet nginx; then
    echo "✅ Nginx is running"
else
    echo "❌ Failed to start nginx"
    exit 1
fi

echo ""
echo "🎉 Deployment completed successfully!"
echo ""
echo "📊 Service Status:"
echo "=================="
systemctl status "$SERVICE_NAME" --no-pager -l | head -10
echo ""
systemctl status nginx --no-pager -l | head -10
echo ""
echo "🌐 Service Endpoints (via nginx):"
echo "  Auth Gateway:      http://yourdomain.com/auth/"
echo "  Tenant Node:       http://yourdomain.com/data/"
echo "  Operation Node:    http://yourdomain.com/query/"
echo "  CBO Engine:        http://yourdomain.com/optimize/"
echo "  Metadata Catalog:  http://yourdomain.com/metadata/"
echo "  Monitoring:        http://yourdomain.com/monitor/"
echo "  Query Interpreter: http://yourdomain.com/parse/"
echo "  Health Check:      http://yourdomain.com/health"
echo ""
echo "🔧 Direct Service Access:"
echo "  Auth Gateway:      http://localhost:8090/health"
echo "  Tenant Node:       http://localhost:8000/health"
echo "  Operation Node:    http://localhost:8081/health"
echo "  CBO Engine:        http://localhost:8082/health"
echo "  Metadata Catalog:  http://localhost:8083/health"
echo "  Monitoring:        http://localhost:8084/health"
echo "  Query Interpreter: http://localhost:8085/health"
echo ""
echo "📝 Next steps:"
echo "1. Update 'yourdomain.com' in /etc/nginx/sites-available/$SERVICE_NAME"
echo "2. Test the deployment: ./test_nginx_proxy.sh"
echo "3. View logs: sudo journalctl -u $SERVICE_NAME -f"
echo ""
echo "🔧 Useful commands:"
echo "  sudo systemctl restart $SERVICE_NAME  # Restart Go service"
echo "  sudo systemctl reload nginx          # Reload nginx config"
echo "  sudo systemctl status $SERVICE_NAME  # Check service status"
echo "  sudo journalctl -u $SERVICE_NAME -f  # Follow service logs"
echo "  sudo tail -f /var/log/nginx/access.log  # Nginx access logs"
echo "  sudo tail -f /var/log/nginx/error.log   # Nginx error logs"
