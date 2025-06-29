package api

import (
	"encoding/json"
	"net/http"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/routing"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

type Server struct {
	router *routing.Router
}

func NewServer(router *routing.Router) *Server {
	return &Server{router: router}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-Id")
	if tenantID == "" {
		http.Error(w, "missing X-Tenant-Id header", http.StatusBadRequest)
		return
	}
	backend, err := s.router.LookupBackend(tenantID)
	if err != nil {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}
	switch r.URL.Path {
	case "/data":
		if r.Method == http.MethodPost {
			s.handlePutData(w, r, backend, tenantID)
			return
		}
		if r.Method == http.MethodGet {
			s.handleGetData(w, r, backend, tenantID)
			return
		}
	}
	http.NotFound(w, r)
}

func (s *Server) handlePutData(w http.ResponseWriter, r *http.Request, backend *routing.Backend, tenantID string) {
	var req models.BusinessData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	req.TenantID = tenantID
	if err := backend.RocksDB.PutBusinessData(req); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if err := backend.ClickHouse.PutBusinessData(req); err != nil {
		http.Error(w, "analytics error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleGetData(w http.ResponseWriter, r *http.Request, backend *routing.Backend, tenantID string) {
	data, err := backend.RocksDB.GetBusinessData(tenantID)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(data)
}