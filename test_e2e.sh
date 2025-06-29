#!/bin/bash

# End-to-End Test Script for Storage Control Plane
# Compatible with Linux, macOS, and WSL
set -e

echo "🧪 Starting End-to-End Tests..."

# Load environment variables if .env.test exists
if [ -f ".env.test" ]; then
    echo "📄 Loading test environment variables..."
    export $(grep -v '^#' .env.test | xargs)
fi

# Configuration
BASE_URL="${TEST_BASE_URL:-http://localhost:8081}"
TENANT_ID="test-tenant-$(date +%s)"
SOURCE_ID="test-source-001"
TIMEOUT_SECONDS="${TEST_TIMEOUT:-30}"

echo "📋 Using Tenant ID: $TENANT_ID"
echo "🔗 Using Source ID: $SOURCE_ID"
echo "🌐 Testing against: $BASE_URL"

# Wait for server to be ready
echo "⏳ Waiting for server to be ready..."
for i in {1..10}; do
    if curl -s --connect-timeout 5 "$BASE_URL/" > /dev/null 2>&1; then
        echo "✅ Server is ready!"
        break
    fi
    if [ $i -eq 10 ]; then
        echo "❌ Server not responding after 10 attempts"
        echo "💡 Make sure the server is running with: ./dev.sh or make dev"
        exit 1
    fi
    echo "⏳ Attempt $i/10: Server not ready, waiting 2 seconds..."
    sleep 2
done

# Test 1: Health Check (if implemented)
echo "🏥 Testing server health..."
curl -f --connect-timeout $TIMEOUT_SECONDS "$BASE_URL/" || echo "⚠️  Health endpoint not implemented yet"

# Test 2: POST Data Ingestion
echo "📤 Testing data ingestion..."
TEST_DATA='{
  "data_id": "user-001",
  "payload": {
    "name": "John Doe",
    "age": 30,
    "email": "john@example.com",
    "profile": {
      "bio": "Software Engineer",
      "location": "San Francisco",
      "skills": ["Go", "Python", "JavaScript"]
    },
    "metadata": {
      "created_at": "2025-06-29T20:30:00Z",
      "version": 1
    }
  }
}'

RESPONSE=$(curl -s -w "%{http_code}" \
  --connect-timeout $TIMEOUT_SECONDS \
  --max-time $TIMEOUT_SECONDS \
  -X POST \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: $TENANT_ID" \
  -d "$TEST_DATA" \
  "$BASE_URL/data")

HTTP_CODE="${RESPONSE: -3}"
BODY="${RESPONSE%???}"

echo "📊 POST Response Code: $HTTP_CODE"
if [ "$HTTP_CODE" = "201" ]; then
  echo "✅ Data ingestion successful"
else
  echo "❌ Data ingestion failed"
  echo "Response: $BODY"
fi

# Test 3: GET Data Retrieval  
echo "📥 Testing data retrieval..."
GET_RESPONSE=$(curl -s -w "%{http_code}" \
  --connect-timeout $TIMEOUT_SECONDS \
  --max-time $TIMEOUT_SECONDS \
  -H "X-Tenant-Id: $TENANT_ID" \
  "$BASE_URL/data")

GET_HTTP_CODE="${GET_RESPONSE: -3}"
GET_BODY="${GET_RESPONSE%???}"

echo "📊 GET Response Code: $GET_HTTP_CODE"
if [ "$GET_HTTP_CODE" = "200" ]; then
  echo "✅ Data retrieval successful"
  echo "📋 Retrieved data: $GET_BODY"
else
  echo "❌ Data retrieval failed"
  echo "Response: $GET_BODY"
fi

# Test 4: Schema Evolution Test
echo "🔄 Testing schema evolution..."
EVOLVED_DATA='{
  "data_id": "user-002", 
  "payload": {
    "name": "Jane Smith",
    "age": 25,
    "email": "jane@example.com",
    "profile": {
      "bio": "Data Scientist",
      "location": "New York",
      "skills": ["Python", "R", "SQL"],
      "certifications": ["AWS", "GCP"]
    },
    "preferences": {
      "theme": "dark",
      "notifications": true
    },
    "metadata": {
      "created_at": "2025-06-29T20:35:00Z",
      "version": 1,
      "source": "api_v2"
    }
  }
}'

SCHEMA_RESPONSE=$(curl -s -w "%{http_code}" \
  --connect-timeout $TIMEOUT_SECONDS \
  --max-time $TIMEOUT_SECONDS \
  -X POST \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: $TENANT_ID" \
  -d "$EVOLVED_DATA" \
  "$BASE_URL/data")

SCHEMA_HTTP_CODE="${SCHEMA_RESPONSE: -3}"
echo "📊 Schema Evolution Response Code: $SCHEMA_HTTP_CODE"

# Test 5: Bulk Data Test
echo "🔄 Testing bulk data ingestion..."
for i in {1..5}; do
  BULK_DATA='{
    "data_id": "bulk-'$i'",
    "payload": {
      "batch_id": '$i',
      "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
      "data": {
        "value": '$((i * 10))',
        "processed": true
      }
    }
  }'
  
  curl -s \
    --connect-timeout $TIMEOUT_SECONDS \
    --max-time $TIMEOUT_SECONDS \
    -X POST \
    -H "Content-Type: application/json" \
    -H "X-Tenant-Id: $TENANT_ID" \
    -d "$BULK_DATA" \
    "$BASE_URL/data" > /dev/null
    
  echo "📦 Bulk record $i sent"
done

# Test 6: Query with Sorting
echo "🔄 Testing query with sorting..."
SORT_QUERY='{
  "tenant_id": "'$TENANT_ID'",
  "query_type": "sql", 
  "query": "SELECT * FROM tenant_data",
  "sort_by": [
    {"field": "created_at", "direction": "desc"},
    {"field": "data_id", "direction": "asc"}
  ],
  "limit": 10,
  "offset": 0,
  "format": "json"
}'

SORT_RESPONSE=$(curl -s -w "%{http_code}" \
  --connect-timeout $TIMEOUT_SECONDS \
  --max-time $TIMEOUT_SECONDS \
  -X POST \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: $TENANT_ID" \
  -d "$SORT_QUERY" \
  "$BASE_URL/data/query")

SORT_HTTP_CODE="${SORT_RESPONSE: -3}"
echo "📊 Sort Query Response Code: $SORT_HTTP_CODE"

# Test 7: Analytics with Sorting
echo "🔄 Testing analytics with sorting..."
ANALYTICS_RESPONSE=$(curl -s -w "%{http_code}" \
  --connect-timeout $TIMEOUT_SECONDS \
  --max-time $TIMEOUT_SECONDS \
  -H "X-Tenant-Id: $TENANT_ID" \
  "$BASE_URL/analytics/summary?sort_by=created_at,total_files&sort_order=desc,asc&limit=5")

ANALYTICS_HTTP_CODE="${ANALYTICS_RESPONSE: -3}"
echo "📊 Analytics Response Code: $ANALYTICS_HTTP_CODE"

echo "🎉 End-to-End Tests Completed!"
echo ""
echo "📝 Test Summary:"
echo "   - Health Check: Basic connectivity"
echo "   - Data Ingestion: JSON data with nested structures"  
echo "   - Data Retrieval: Reading stored data"
echo "   - Schema Evolution: Different JSON structure"
echo "   - Bulk Processing: Multiple records"
echo "   - Query with Sorting: Advanced query capabilities"
echo "   - Analytics with Sorting: Sorted analytics data"
echo ""
echo "🔍 Check application logs for WAL/Parquet processing"
echo "🗄️  Check ClickHouse for table creation and data storage"
echo ""
echo "💡 Tips for Linux/Unix development:"
echo "   - Run with: chmod +x test_e2e.sh && ./test_e2e.sh"
echo "   - Use make commands: make test-e2e, make dev, make build"
echo "   - Configure .env.test for custom test settings"
echo "   - For WSL: Ensure Docker Desktop integration is enabled"
