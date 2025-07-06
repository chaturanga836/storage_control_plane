#!/bin/bash
# Test connectivity between Go Control Plane and Python Services

echo "🧪 TESTING DISTRIBUTED STORAGE SYSTEM CONNECTIVITY"
echo "=================================================="

# Configuration
PYTHON_IP="65.0.150.75"
GO_IP="15.207.184.150"

echo ""
echo "📋 Configuration:"
echo "  Python Services: $PYTHON_IP"
echo "  Go Control Plane: $GO_IP"
echo ""

echo "🔗 Testing Python Services Connectivity..."
echo "----------------------------------------"

# Test Python services
services=(
    "8080:Auth Gateway"
    "8001:Tenant Node"
    "8087:Metadata Catalog" 
    "8086:Operation Node"
    "8088:CBO Engine"
    "8089:Monitoring"
    "8085:Query Interpreter"
)

python_healthy=0
for service in "${services[@]}"; do
    IFS=':' read -r port name <<< "$service"
    echo -n "Testing $name (port $port)... "
    
    if timeout 5 curl -s "http://$PYTHON_IP:$port/health" > /dev/null 2>&1; then
        echo "✅ HEALTHY"
        ((python_healthy++))
    elif timeout 5 curl -s "http://$PYTHON_IP:$port/" > /dev/null 2>&1; then
        echo "🟡 RESPONDING (no /health endpoint)"
        ((python_healthy++))
    else
        echo "❌ NOT RESPONDING"
    fi
done

echo ""
echo "🚀 Testing Go Control Plane..."
echo "----------------------------"

go_healthy=0
echo -n "Testing Go Control Plane (port 8090)... "
if timeout 5 curl -s "http://localhost:8090/health" > /dev/null 2>&1; then
    echo "✅ HEALTHY"
    ((go_healthy++))
elif timeout 5 curl -s "http://localhost:8090/" > /dev/null 2>&1; then
    echo "🟡 RESPONDING (no /health endpoint)"
    ((go_healthy++))
else
    echo "❌ NOT RESPONDING"
fi

echo ""
echo "🌐 Testing External Access..."
echo "----------------------------"

echo -n "Testing Go Control Plane external access... "
if timeout 5 curl -s "http://$GO_IP:8090/health" > /dev/null 2>&1; then
    echo "✅ ACCESSIBLE EXTERNALLY"
else
    echo "❌ NOT ACCESSIBLE (check security groups)"
fi

echo ""
echo "📊 CONNECTIVITY SUMMARY"
echo "======================="
echo "Python Services Healthy: $python_healthy/7"
echo "Go Control Plane Healthy: $go_healthy/1"

if [ $python_healthy -eq 7 ] && [ $go_healthy -eq 1 ]; then
    echo ""
    echo "🎉 SUCCESS! Both systems are fully operational!"
    echo ""
    echo "🌐 Access URLs:"
    echo "  Go Control Plane:   http://$GO_IP:8090"
    echo "  Python Auth Gateway: http://$PYTHON_IP:8080"
    echo "  Python Tenant Node:  http://$PYTHON_IP:8001"
    echo ""
    echo "🧪 Test Commands:"
    echo "  curl http://$GO_IP:8090/health"
    echo "  curl http://$PYTHON_IP:8080/health"
    echo "  curl http://$PYTHON_IP:8001/health"
elif [ $python_healthy -gt 0 ] && [ $go_healthy -gt 0 ]; then
    echo ""
    echo "🟡 PARTIAL SUCCESS - Some services are running"
    echo "   Check individual service logs for issues"
else
    echo ""
    echo "❌ DEPLOYMENT ISSUES DETECTED"
    echo ""
    echo "🔧 Troubleshooting:"
    echo "  1. Check Python services: ssh ubuntu@$PYTHON_IP"
    echo "  2. Check Go service: sudo systemctl status storage-control-plane"
    echo "  3. Check security groups allow ports 8080, 8001, 8087, 8090"
    echo "  4. Check firewall rules on both instances"
fi

echo ""
echo "📋 Next Steps:"
echo "  1. Update security groups to allow cross-instance communication"
echo "  2. Test API endpoints on both systems"
echo "  3. Monitor logs: sudo journalctl -u storage-control-plane -f"
