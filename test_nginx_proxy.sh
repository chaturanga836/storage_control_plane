#!/bin/bash

# Test Go Control Plane via Nginx Reverse Proxy
# This script validates that all services are accessible through nginx

set -e

echo "🧪 Testing Go Control Plane via Nginx"
echo "====================================="

# Configuration
DOMAIN="${1:-localhost}"
BASE_URL="http://${DOMAIN}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Test function
test_endpoint() {
    local endpoint="$1"
    local method="${2:-GET}"
    local expected_status="${3:-200}"
    local description="$4"
    
    echo -n "Testing ${endpoint} ... "
    
    local url="${BASE_URL}${endpoint}"
    local response
    local status_code
    
    if response=$(curl -s -w "%{http_code}" -X "${method}" "${url}" 2>/dev/null); then
        status_code="${response: -3}"
        response_body="${response%???}"
        
        if [ "${status_code}" = "${expected_status}" ]; then
            echo -e "${GREEN}✅ OK (${status_code})${NC}"
            if [ -n "$description" ]; then
                echo -e "   ${CYAN}${description}${NC}"
            fi
            return 0
        else
            echo -e "${YELLOW}⚠️  Unexpected status: ${status_code}${NC}"
            return 1
        fi
    else
        echo -e "${RED}❌ Failed to connect${NC}"
        return 1
    fi
}

# Test POST endpoint with JSON
test_post_endpoint() {
    local endpoint="$1"
    local json_data="$2"
    local description="$3"
    
    echo -n "Testing POST ${endpoint} ... "
    
    local url="${BASE_URL}${endpoint}"
    local response
    local status_code
    
    if response=$(curl -s -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "${json_data}" "${url}" 2>/dev/null); then
        status_code="${response: -3}"
        response_body="${response%???}"
        
        if [ "${status_code}" = "200" ] || [ "${status_code}" = "201" ]; then
            echo -e "${GREEN}✅ OK (${status_code})${NC}"
            if [ -n "$description" ]; then
                echo -e "   ${CYAN}${description}${NC}"
            fi
            return 0
        else
            echo -e "${YELLOW}⚠️  Status: ${status_code}${NC}"
            return 1
        fi
    else
        echo -e "${RED}❌ Failed to connect${NC}"
        return 1
    fi
}

echo "🌐 Testing domain: ${DOMAIN}"
echo

# Health checks for all services
echo -e "${YELLOW}🔍 Health Checks:${NC}"
test_endpoint "/health" "GET" "200" "Main health endpoint"

# Auth Gateway tests (Port 8090)
echo
echo -e "${YELLOW}🔐 Auth Gateway Tests:${NC}"
test_post_endpoint "/auth/login" '{"username":"admin","password":"password"}' "Login endpoint"
test_endpoint "/auth/validate" "GET" "200" "Token validation"

# Tenant Node tests (Port 8000)
echo
echo -e "${YELLOW}🏢 Tenant Node Tests:${NC}"
test_endpoint "/data/stats" "GET" "200" "Data statistics"
test_post_endpoint "/data/store" '{"key":"test","value":"data"}' "Store data"

# Operation Node tests (Port 8081)
echo
echo -e "${YELLOW}🎯 Operation Node Tests:${NC}"
test_endpoint "/query/status" "GET" "200" "Query status"
test_post_endpoint "/query/execute" '{"query":"SELECT * FROM test"}' "Execute query"

# CBO Engine tests (Port 8082)
echo
echo -e "${YELLOW}🧠 CBO Engine Tests:${NC}"
test_endpoint "/optimize/stats" "GET" "200" "Optimizer statistics"
test_post_endpoint "/optimize/query" '{"query":"SELECT * FROM users"}' "Query optimization"

# Metadata Catalog tests (Port 8083)
echo
echo -e "${YELLOW}📊 Metadata Catalog Tests:${NC}"
test_endpoint "/metadata/stats" "GET" "200" "Metadata statistics"
test_endpoint "/metadata/tables" "GET" "200" "Table metadata"

# Monitoring tests (Port 8084)
echo
echo -e "${YELLOW}📈 Monitoring Tests:${NC}"
test_endpoint "/monitor/metrics" "GET" "200" "System metrics"
test_endpoint "/monitor/health" "GET" "200" "Monitoring health"

# Query Interpreter tests (Port 8085)
echo
echo -e "${YELLOW}🔍 Query Interpreter Tests:${NC}"
test_endpoint "/parse/health" "GET" "200" "Parser health"
test_post_endpoint "/parse/sql" '{"query":"SELECT id FROM users"}' "SQL parsing"

# Version endpoint
echo
echo -e "${YELLOW}ℹ️  Version Information:${NC}"
test_endpoint "/version" "GET" "200" "Version info"

echo
echo -e "${GREEN}🎉 Testing Complete!${NC}"
echo

# Summary
echo -e "${CYAN}📋 Service Summary:${NC}"
echo "  🔐 Auth Gateway:      ${BASE_URL}/auth/"
echo "  🏢 Tenant Node:       ${BASE_URL}/data/"
echo "  🎯 Operation Node:    ${BASE_URL}/query/"
echo "  🧠 CBO Engine:        ${BASE_URL}/optimize/"
echo "  📊 Metadata Catalog:  ${BASE_URL}/metadata/"
echo "  📈 Monitoring:        ${BASE_URL}/monitor/"
echo "  🔍 Query Interpreter: ${BASE_URL}/parse/"

echo
echo -e "${CYAN}🔧 Debugging:${NC}"
echo "  Nginx status:  systemctl status nginx"
echo "  Nginx logs:    tail -f /var/log/nginx/storage-control-plane.*.log"
echo "  Test direct:   curl http://localhost:8090/health"

echo
echo -e "${CYAN}💡 Tips:${NC}"
echo "  - Ensure Go Control Plane is running: go run ."
echo "  - Check nginx config: nginx -t"
echo "  - Reload nginx: systemctl reload nginx"
