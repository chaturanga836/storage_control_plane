#!/bin/bash

# Storage Control Plane - Nginx Deployment Script
# Repository location: /opt/storage_control_plane

set -e

REPO_DIR="/opt/storage_control_plane"
NGINX_CONF="$REPO_DIR/nginx.conf"
NGINX_SITES_AVAILABLE="/etc/nginx/sites-available"
NGINX_SITES_ENABLED="/etc/nginx/sites-enabled"
SERVICE_NAME="storage-control-plane"

echo "🚀 Setting up Nginx reverse proxy for Storage Control Plane"
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
    exit 1
fi

# Check if nginx configuration exists
if [[ ! -f "$NGINX_CONF" ]]; then
    echo "❌ Nginx configuration not found at $NGINX_CONF"
    exit 1
fi

# Install nginx if not present
if ! command -v nginx &> /dev/null; then
    echo "📦 Installing nginx..."
    apt-get update
    apt-get install -y nginx
fi

# Stop nginx if running
echo "🛑 Stopping nginx..."
systemctl stop nginx || true

# Backup existing configuration if it exists
if [[ -f "$NGINX_SITES_AVAILABLE/$SERVICE_NAME" ]]; then
    echo "📋 Backing up existing configuration..."
    cp "$NGINX_SITES_AVAILABLE/$SERVICE_NAME" "$NGINX_SITES_AVAILABLE/$SERVICE_NAME.backup.$(date +%Y%m%d_%H%M%S)"
fi

# Copy nginx configuration
echo "📁 Installing nginx configuration..."
cp "$NGINX_CONF" "$NGINX_SITES_AVAILABLE/$SERVICE_NAME"

# Remove default site if it exists
if [[ -f "$NGINX_SITES_ENABLED/default" ]]; then
    echo "🗑️  Removing default nginx site..."
    rm -f "$NGINX_SITES_ENABLED/default"
fi

# Enable the site
echo "🔗 Enabling Storage Control Plane site..."
rm -f "$NGINX_SITES_ENABLED/$SERVICE_NAME"
ln -s "$NGINX_SITES_AVAILABLE/$SERVICE_NAME" "$NGINX_SITES_ENABLED/$SERVICE_NAME"

# Test nginx configuration
echo "🧪 Testing nginx configuration..."
if nginx -t; then
    echo "✅ Nginx configuration is valid"
else
    echo "❌ Nginx configuration test failed"
    exit 1
fi

# Start nginx
echo "🚀 Starting nginx..."
systemctl start nginx
systemctl enable nginx

# Check nginx status
if systemctl is-active --quiet nginx; then
    echo "✅ Nginx is running"
else
    echo "❌ Failed to start nginx"
    exit 1
fi

echo ""
echo "🎉 Nginx reverse proxy setup completed successfully!"
echo ""
echo "📊 Service Status:"
systemctl status nginx --no-pager -l
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
echo "📝 Next steps:"
echo "1. Update 'yourdomain.com' in $NGINX_SITES_AVAILABLE/$SERVICE_NAME"
echo "2. Start Go services: cd $REPO_DIR && go run ."
echo "3. Test endpoints: curl http://yourdomain.com/health"
echo ""
echo "🔧 Useful commands:"
echo "  sudo systemctl reload nginx    # Reload nginx config"
echo "  sudo systemctl restart nginx  # Restart nginx"
echo "  sudo nginx -t                 # Test nginx config"
echo "  sudo tail -f /var/log/nginx/access.log  # View access logs"
echo "  sudo tail -f /var/log/nginx/error.log   # View error logs"
