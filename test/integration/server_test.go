package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/config"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/routing"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

func TestServerRequiresTenantHeader(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":8081",
	}
	router := routing.NewRouter(cfg)
	server := NewServer(router)
	
	req := httptest.NewRequest("GET", "/data", nil)
	w := httptest.NewRecorder()
	
	server.ServeHTTP(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
	
	if !bytes.Contains(w.Body.Bytes(), []byte("missing X-Tenant-Id header")) {
		t.Error("Expected error message about missing tenant header")
	}
}

func TestPutDataEndpoint(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":8081",
	}
	router := routing.NewRouter(cfg)
	server := NewServer(router)
	
	// Create test data
	testData := models.BusinessData{
		DataID:  "test-123",
		Payload: []byte(`{"name": "test", "value": 42}`),
	}
	
	jsonData, _ := json.Marshal(testData)
	req := httptest.NewRequest("POST", "/data", bytes.NewReader(jsonData))
	req.Header.Set("X-Tenant-Id", "test-tenant")
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	// Note: This will fail until we implement proper mock backends
	// But it tests the HTTP handling logic
	if w.Code == http.StatusBadRequest {
		t.Error("Request should not fail due to bad request format")
	}
}

func TestGetDataEndpoint(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: ":8081",
	}
	router := routing.NewRouter(cfg)
	server := NewServer(router)
	
	req := httptest.NewRequest("GET", "/data", nil)
	req.Header.Set("X-Tenant-Id", "test-tenant")
	
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	
	// Should not be a bad request (tenant header is present)
	if w.Code == http.StatusBadRequest {
		t.Error("GET request with tenant header should not be bad request")
	}
}
