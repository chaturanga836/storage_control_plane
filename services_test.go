package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setupTestServer creates a test server with all routes configured
func setupTestServer() *http.ServeMux {
	mux := http.NewServeMux()
	
	// Setup all service routes
	setupAuthGatewayRoutes(mux)
	setupTenantNodeRoutes(mux)
	setupOperationNodeRoutes(mux)
	setupCBOEngineRoutes(mux)
	setupMetadataCatalogRoutes(mux)
	setupQueryInterpreterRoutes(mux)
	setupMonitoringRoutes(mux)
	
	// Add health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":  "healthy",
			"service": "Storage Control Plane",
			"version": "1.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	return mux
}

// Test Auth Gateway handlers
func TestAuthLogin(t *testing.T) {
	mux := setupTestServer()
	
	loginData := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}

	jsonData, _ := json.Marshal(loginData)
	req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse JSON response")
	}

	if response["success"] != true {
		t.Errorf("Expected success true, got %v", response["success"])
	}

	if _, exists := response["token"]; !exists {
		t.Error("Expected token in response")
	}
}

func TestAuthValidateToken(t *testing.T) {
	mux := setupTestServer()
	
	req, err := http.NewRequest("GET", "/auth/validate", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer test-token")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse JSON response")
	}

	if response["valid"] != true {
		t.Errorf("Expected valid true, got %v", response["valid"])
	}
}

// Test Tenant Node handlers
func TestTenantDataQuery(t *testing.T) {
	mux := setupTestServer()
	
	queryData := map[string]interface{}{
		"query": "SELECT * FROM users WHERE id = 1",
		"filters": map[string]interface{}{
			"tenant_id": "tenant-1",
		},
	}

	jsonData, _ := json.Marshal(queryData)
	req, err := http.NewRequest("POST", "/data/execute", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse JSON response")
	}

	if response["success"] != true {
		t.Errorf("Expected success true, got %v", response["success"])
	}
}

func TestTenantDataInsert(t *testing.T) {
	mux := setupTestServer()
	
	insertData := map[string]interface{}{
		"table": "users",
		"data": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		},
	}

	jsonData, _ := json.Marshal(insertData)
	req, err := http.NewRequest("POST", "/data/store", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse JSON response")
	}

	if response["success"] != true {
		t.Errorf("Expected success true, got %v", response["success"])
	}
}

// Test Operation Node handlers
func TestOperationExecuteQuery(t *testing.T) {
	mux := setupTestServer()
	
	queryData := map[string]interface{}{
		"query": "SELECT COUNT(*) FROM transactions WHERE date > '2024-01-01'",
		"optimization_level": "high",
	}

	jsonData, _ := json.Marshal(queryData)
	req, err := http.NewRequest("POST", "/query/execute", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse JSON response")
	}

	if response["success"] != true {
		t.Errorf("Expected success true, got %v", response["success"])
	}
}

// Test CBO Engine handlers
func TestCBOOptimize(t *testing.T) {
	mux := setupTestServer()
	
	queryData := map[string]interface{}{
		"query": "SELECT u.name, COUNT(o.id) FROM users u LEFT JOIN orders o ON u.id = o.user_id GROUP BY u.id",
		"constraints": map[string]interface{}{
			"max_memory": "4GB",
			"timeout":    30,
		},
	}

	jsonData, _ := json.Marshal(queryData)
	req, err := http.NewRequest("POST", "/optimize/query", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse JSON response")
	}

	if response["success"] != true {
		t.Errorf("Expected success true, got %v", response["success"])
	}
}

// Test Query Interpreter handlers
func TestQueryParse(t *testing.T) {
	mux := setupTestServer()
	
	queryData := map[string]interface{}{
		"query": "SELECT * FROM products WHERE price > 100 AND category = 'electronics'",
		"dialect": "postgresql",
	}

	jsonData, _ := json.Marshal(queryData)
	req, err := http.NewRequest("POST", "/parse/sql", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse JSON response")
	}

	if response["success"] != true {
		t.Errorf("Expected success true, got %v", response["success"])
	}
}

// Test Health endpoint
func TestHealthEndpoint(t *testing.T) {
	mux := setupTestServer()
	
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse JSON response")
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}

	if response["service"] != "Storage Control Plane" {
		t.Errorf("Expected service 'Storage Control Plane', got %v", response["service"])
	}
}

// Integration test - Auth flow
func TestAuthFlow(t *testing.T) {
	mux := setupTestServer()
	
	// 1. Login
	loginData := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}

	jsonData, _ := json.Marshal(loginData)
	req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	var loginResponse map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &loginResponse); err != nil {
		t.Fatal("Failed to parse login response")
	}

	token, exists := loginResponse["token"].(string)
	if !exists {
		t.Fatal("No token in login response")
	}

	// 2. Validate token
	req, err = http.NewRequest("GET", "/auth/validate", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	var validateResponse map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &validateResponse); err != nil {
		t.Fatal("Failed to parse validate response")
	}

	if validateResponse["valid"] != true {
		t.Error("Token validation failed")
	}

	// 3. Logout
	req, err = http.NewRequest("POST", "/auth/logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("logout returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

// Additional comprehensive tests

// Test Monitoring endpoints
func TestMonitoringMetrics(t *testing.T) {
	mux := setupTestServer()
	
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse JSON response")
	}

	if _, exists := response["system_metrics"]; !exists {
		t.Error("Expected system_metrics in response")
	}

	if _, exists := response["query_metrics"]; !exists {
		t.Error("Expected query_metrics in response")
	}
}

// Test Metadata endpoints
func TestMetadataPartitions(t *testing.T) {
	mux := setupTestServer()
	
	// This endpoint requires POST method
	queryData := map[string]interface{}{
		"table": "orders",
	}
	
	jsonData, _ := json.Marshal(queryData)
	req, err := http.NewRequest("POST", "/metadata/partitions", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse JSON response")
	}

	if _, exists := response["partition_metadata"]; !exists {
		t.Error("Expected partition_metadata in response")
	}
}

// Test Query validation
func TestQueryValidation(t *testing.T) {
	mux := setupTestServer()
	
	queryData := map[string]interface{}{
		"query": "SELECT * FROM invalid_table WHERE invalid_column = 'test'",
		"dialect": "postgresql",
	}

	jsonData, _ := json.Marshal(queryData)
	req, err := http.NewRequest("POST", "/validate/query", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse JSON response")
	}

	// Query validation response has 'valid' field, not 'success'
	if _, exists := response["valid"]; !exists {
		t.Error("Expected valid field in response")
	}
}

// Test error handling for malformed requests
func TestErrorHandling(t *testing.T) {
	mux := setupTestServer()
	
	// Test with invalid JSON
	req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Should handle malformed JSON gracefully
	if status := rr.Code; status == http.StatusInternalServerError {
		t.Error("Should handle malformed JSON without internal server error")
	}
}

// Test concurrent requests
func TestConcurrentRequests(t *testing.T) {
	mux := setupTestServer()
	
	const numRequests = 10
	done := make(chan bool, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func() {
			req, err := http.NewRequest("GET", "/health", nil)
			if err != nil {
				t.Error(err)
				done <- false
				return
			}
			
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("concurrent request failed with status: %v", status)
				done <- false
				return
			}
			
			done <- true
		}()
	}
	
	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		if success := <-done; !success {
			t.Fatal("One or more concurrent requests failed")
		}
	}
}
