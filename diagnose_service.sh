#!/bin/bash
# Diagnostic script for Go Control Plane service

echo "🔍 GO CONTROL PLANE DIAGNOSTICS"
echo "==============================="

echo ""
echo "📋 Service Status:"
sudo systemctl status storage-control-plane --no-pager

echo ""
echo "📋 Recent Logs:"
sudo journalctl -u storage-control-plane --no-pager -n 20

echo ""
echo "📋 Process Information:"
echo "Process ID: $(pgrep -f storage-control-plane || echo 'Not running')"
if pgrep -f storage-control-plane > /dev/null; then
    ps aux | grep storage-control-plane | grep -v grep
fi

echo ""
echo "📋 Port Listening:"
sudo netstat -tlnp | grep :8090 || echo "Port 8090 not listening"

echo ""
echo "📋 Environment File:"
if [ -f ".env" ]; then
    echo "Environment file exists:"
    cat .env
else
    echo "No .env file found"
fi

echo ""
echo "📋 Binary Information:"
if [ -f "storage-control-plane" ]; then
    ls -la storage-control-plane
    echo "Binary type: $(file storage-control-plane)"
else
    echo "Binary not found"
fi

echo ""
echo "📋 Network Test:"
echo "Testing localhost:8090..."
curl -v http://localhost:8090/ 2>&1 | head -10

echo ""
echo "📋 Health Check Test:"
echo "Testing localhost:8090/health..."
curl -v http://localhost:8090/health 2>&1 | head -10

echo ""
echo "📋 Manual Test:"
echo "Try running: ./storage-control-plane"
echo "Or check logs: sudo journalctl -u storage-control-plane -f"
