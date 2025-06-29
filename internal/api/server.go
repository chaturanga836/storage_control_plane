package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/clickhouse"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/routing"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/utils"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/wal"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

type Server struct {
	router             *routing.Router
	distributedManager *clickhouse.DistributedIndexManager
	isDistributedMode  bool
}

func NewServer(router *routing.Router) *Server {
	return &Server{
		router:            router,
		isDistributedMode: false,
	}
}

// NewDistributedServer creates a server with distributed index management
func NewDistributedServer(router *routing.Router, distributedManager *clickhouse.DistributedIndexManager) *Server {
	return &Server{
		router:            router,
		distributedManager: distributedManager,
		isDistributedMode: true,
	}
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
	case "/data/query":
		if r.Method == http.MethodPost {
			s.handleQueryData(w, r, backend, tenantID)
			return
		}
	// Distributed index management endpoints
	case "/distributed/indexes":
		if r.Method == http.MethodPost {
			s.handleCreateDistributedIndex(w, r, tenantID)
			return
		}
		if r.Method == http.MethodGet {
			s.handleListDistributedIndexes(w, r, tenantID)
			return
		}
	case "/distributed/indexes/optimize":
		if r.Method == http.MethodPost {
			s.handleOptimizeDistributedIndexes(w, r, tenantID)
			return
		}
	case "/distributed/cluster/topology":
		if r.Method == http.MethodGet {
			s.handleGetClusterTopology(w, r, tenantID)
			return
		}
	case "/distributed/query/optimize":
		if r.Method == http.MethodPost {
			s.handleOptimizeDistributedQuery(w, r, tenantID)
			return
		}
	case "/analytics/summary":
		if r.Method == http.MethodGet {
			s.handleAnalyticsSummary(w, r, backend, tenantID)
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

// New sorting-enabled query endpoint
func (s *Server) handleQueryData(w http.ResponseWriter, r *http.Request, backend *routing.Backend, tenantID string) {
	var req models.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	req.TenantID = tenantID

	// Validate sort fields based on query type
	var sortOptions utils.SortOptions
	switch req.QueryType {
	case models.QuerySQL, models.QueryAggregate:
		sortOptions = utils.AnalyticsSortOptions
	case models.QueryTimeSeries:
		sortOptions = utils.SortOptions{
			DefaultField:     "timestamp",
			DefaultDirection: models.SortDesc,
			AllowedFields:    []string{"timestamp", "value", "count"},
			MaxSortFields:    3,
		}
	default:
		sortOptions = utils.DataIngestionSortOptions
	}

	// Validate sort parameters
	validatedSorts, err := utils.ValidateSortFields(req.SortBy, sortOptions)
	if err != nil {
		http.Error(w, "invalid sort parameters: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.SortBy = validatedSorts

	// Use large-scale query executor for better performance
	config := utils.DefaultLargeScaleConfig

	// Type assert to get the actual ClickHouse store
	clickhouseStore, ok := backend.ClickHouse.(*clickhouse.Store)
	if !ok {
		http.Error(w, "ClickHouse backend not available", http.StatusInternalServerError)
		return
	}

	executor := clickhouse.NewLargeScaleQueryExecutor(clickhouseStore, config)

	// Execute query with large-scale optimizations
	var response *models.QueryResponse
	var queryErr error

	// Use distributed query optimization if available
	if s.isDistributedMode && s.distributedManager != nil {
		response, queryErr = s.executeDistributedQuery(r.Context(), &req, tenantID)
	} else {
		// Execute with large-scale query executor
		response, queryErr = executor.ExecuteLargeQuery(r.Context(), req)
	}

	if queryErr != nil {
		http.Error(w, "query execution failed: "+queryErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// executeDistributedQuery handles query execution in distributed mode
func (s *Server) executeDistributedQuery(ctx context.Context, req *models.QueryRequest, tenantID string) (*models.QueryResponse, error) {
	// Build where conditions for optimization
	whereConditions := map[string]interface{}{
		"tenant_id": tenantID,
	}

	// Add filters to where conditions
	for field, value := range req.Filters {
		whereConditions[field] = value
	}

	// Optimize query for distributed execution
	optimizedQuery, err := s.distributedManager.OptimizeDistributedQuery(
		req.Query,
		req.SortBy,
		whereConditions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize distributed query: %w", err)
	}

	// Execute optimized query (implementation would integrate with your query executor)
	// For now, return a placeholder response
	return &models.QueryResponse{
		QueryID:     "distributed_" + tenantID,
		Data:        []any{map[string]interface{}{"message": "distributed query executed", "query": optimizedQuery}},
		RowCount:    1,
		ExecutionMS: 10,
		FromCache:   false,
	}, nil
}

// handleCreateDistributedIndex creates indexes across all cluster nodes
func (s *Server) handleCreateDistributedIndex(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !s.isDistributedMode || s.distributedManager == nil {
		http.Error(w, "distributed mode not enabled", http.StatusBadRequest)
		return
	}

	var req struct {
		TableName string                    `json:"table_name"`
		IndexName string                    `json:"index_name"`
		Columns   []string                  `json:"columns"`
		IndexType clickhouse.IndexType     `json:"index_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := s.distributedManager.CreateDistributedIndex(
		r.Context(),
		req.TableName,
		req.IndexName,
		req.Columns,
		req.IndexType,
	)
	if err != nil {
		http.Error(w, "failed to create distributed index: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "distributed index created successfully",
		"index_name": req.IndexName,
		"table_name": req.TableName,
		"nodes": "all cluster nodes",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListDistributedIndexes lists indexes across all cluster nodes
func (s *Server) handleListDistributedIndexes(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !s.isDistributedMode || s.distributedManager == nil {
		http.Error(w, "distributed mode not enabled", http.StatusBadRequest)
		return
	}

	tableName := r.URL.Query().Get("table_name")
	if tableName == "" {
		tableName = "events" // default table
	}

	indexInfo, err := s.distributedManager.GetDistributedIndexInfo(r.Context(), tableName)
	if err != nil {
		http.Error(w, "failed to get distributed index info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"table_name": tableName,
		"node_indexes": indexInfo,
		"total_nodes": len(indexInfo),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleOptimizeDistributedIndexes performs maintenance on distributed indexes
func (s *Server) handleOptimizeDistributedIndexes(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !s.isDistributedMode || s.distributedManager == nil {
		http.Error(w, "distributed mode not enabled", http.StatusBadRequest)
		return
	}

	var req struct {
		TableName string `json:"table_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := s.distributedManager.OptimizeDistributedIndexes(r.Context(), req.TableName)
	if err != nil {
		http.Error(w, "failed to optimize distributed indexes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "distributed indexes optimized successfully",
		"table_name": req.TableName,
		"optimized_nodes": "all cluster nodes",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetClusterTopology returns information about the cluster topology
func (s *Server) handleGetClusterTopology(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !s.isDistributedMode || s.distributedManager == nil {
		http.Error(w, "distributed mode not enabled", http.StatusBadRequest)
		return
	}

	topology, err := s.distributedManager.GetClusterTopology(r.Context())
	if err != nil {
		http.Error(w, "failed to get cluster topology: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"cluster_topology": topology,
		"total_nodes": len(topology),
		"timestamp": "now",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleOptimizeDistributedQuery optimizes a query for distributed execution
func (s *Server) handleOptimizeDistributedQuery(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !s.isDistributedMode || s.distributedManager == nil {
		http.Error(w, "distributed mode not enabled", http.StatusBadRequest)
		return
	}

	var req struct {
		Query         string                `json:"query"`
		SortFields    []models.SortField    `json:"sort_fields"`
		WhereConditions map[string]interface{} `json:"where_conditions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Add tenant_id to where conditions
	if req.WhereConditions == nil {
		req.WhereConditions = make(map[string]interface{})
	}
	req.WhereConditions["tenant_id"] = tenantID

	optimizedQuery, err := s.distributedManager.OptimizeDistributedQuery(
		req.Query,
		req.SortFields,
		req.WhereConditions,
	)
	if err != nil {
		http.Error(w, "failed to optimize query: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"original_query": req.Query,
		"optimized_query": optimizedQuery,
		"optimizations_applied": []string{
			"partition pruning",
			"index hints",
			"distributed settings",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// New analytics endpoint with sorting support
func (s *Server) handleAnalyticsSummary(w http.ResponseWriter, r *http.Request, backend *routing.Backend, tenantID string) {
	// Parse query parameters for sorting
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Build sort parameters
	sortFields := utils.BuildSortParams(sortBy, sortOrder)

	// Validate sort fields
	validatedSorts, err := utils.ValidateSortFields(sortFields, utils.TenantSortOptions)
	if err != nil {
		http.Error(w, "invalid sort parameters: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse pagination
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// TODO: Implement actual analytics query with sorting
	summary := models.TenantSummary{
		TenantID:      tenantID,
		TotalFiles:    0,
		TotalRows:     0,
		TotalSizeGB:   0.0,
		SourceCount:   0,
		OldestRecord:  models.TimeRange{}.Start,
		NewestRecord:  models.TimeRange{}.End,
		LastUpdated:   models.TimeRange{}.End,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data":       summary,
		"sort_by":    validatedSorts,
		"pagination": map[string]int{"limit": limit, "offset": offset},
	})
}

// Data ingestion endpoints
func (s *Server) setupRoutes() {
	// ...existing routes...

	// Data ingestion endpoints
	s.router.HandleFunc("/api/v1/tenants/{tenant_id}/sources/{source_id}/ingest", s.ingestData).Methods("POST")
	s.router.HandleFunc("/api/v1/tenants/{tenant_id}/sources/{source_id}/batch", s.ingestBatch).Methods("POST")
	
	// Parquet file management endpoints
	s.router.HandleFunc("/api/v1/tenants/{tenant_id}/sources/{source_id}/files", s.listParquetFiles).Methods("GET")
	s.router.HandleFunc("/api/v1/tenants/{tenant_id}/files", s.listTenantFiles).Methods("GET")
	s.router.HandleFunc("/api/v1/files/{file_id}/stats", s.getFileStats).Methods("GET")
}

// ingestData handles single record ingestion
func (s *Server) ingestData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]
	sourceID := vars["source_id"]

	var record map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Add metadata
	enrichedRecord := map[string]interface{}{
		"tenant_id": tenantID,
		"source_id": sourceID,
		"data":      record,
		"ingested_at": time.Now().UTC().Format(time.RFC3339),
	}

	// Write to WAL for durability
	walDir := filepath.Join("data", "wal", tenantID, sourceID)
	if err := wal.AppendToWAL(walDir, enrichedRecord); err != nil {
		log.Printf("❌ Failed to write to WAL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status": "accepted",
		"message": "Data queued for processing",
		"tenant_id": tenantID,
		"source_id": sourceID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ingestBatch handles batch record ingestion
func (s *Server) ingestBatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]
	sourceID := vars["source_id"]

	var records []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&records); err != nil {
		http.Error(w, "Invalid JSON array", http.StatusBadRequest)
		return
	}

	walDir := filepath.Join("data", "wal", tenantID, sourceID)
	successCount := 0

	for _, record := range records {
		enrichedRecord := map[string]interface{}{
			"tenant_id": tenantID,
			"source_id": sourceID,
			"data":      record,
			"ingested_at": time.Now().UTC().Format(time.RFC3339),
		}

		if err := wal.AppendToWAL(walDir, enrichedRecord); err != nil {
			log.Printf("❌ Failed to write record to WAL: %v", err)
			continue
		}
		successCount++
	}

	response := map[string]interface{}{
		"status": "completed",
		"total_records": len(records),
		"successful_records": successCount,
		"failed_records": len(records) - successCount,
		"message": "Batch queued for processing",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// listParquetFiles lists Parquet files for a specific tenant/source
func (s *Server) listParquetFiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]
	sourceID := vars["source_id"]

	// Parse query parameters
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// List files from data directory
	basePath := filepath.Join("data", "parquet", tenantID, sourceID)
	files, err := s.discoverParquetFiles(basePath, limit, offset)
	if err != nil {
		log.Printf("❌ Failed to list files for %s/%s: %v", tenantID, sourceID, err)
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tenant_id": tenantID,
		"source_id": sourceID,
		"files":     files,
		"total":     len(files),
		"limit":     limit,
		"offset":    offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// listTenantFiles lists all Parquet files for a tenant across all sources
func (s *Server) listTenantFiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	basePath := filepath.Join("data", "parquet", tenantID)
	allFiles := make(map[string][]map[string]interface{})

	// Discover source directories
	sourceDirs, err := filepath.Glob(filepath.Join(basePath, "*"))
	if err != nil {
		log.Printf("❌ Failed to list sources for tenant %s: %v", tenantID, err)
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	for _, sourceDir := range sourceDirs {
		sourceID := filepath.Base(sourceDir)
		files, err := s.discoverParquetFiles(sourceDir, 50, 0)
		if err == nil && len(files) > 0 {
			allFiles[sourceID] = files
		}
	}

	response := map[string]interface{}{
		"tenant_id": tenantID,
		"sources":   allFiles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getFileStats returns detailed statistics for a specific Parquet file
func (s *Server) getFileStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["file_id"]

	// In a real system, you'd look up the file by ID in ClickHouse
	// For now, return a placeholder response
	response := map[string]interface{}{
		"file_id": fileID,
		"message": "File stats would be retrieved from ClickHouse metadata",
		"note":    "This endpoint needs ClickHouse integration for full functionality",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// discoverParquetFiles walks the directory structure to find Parquet files
func (s *Server) discoverParquetFiles(basePath string, limit, offset int) ([]map[string]interface{}, error) {
	var files []map[string]interface{}
	count := 0

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		if !info.IsDir() && strings.HasSuffix(path, ".parquet") {
			if count >= offset && len(files) < limit {
				// Try to read accompanying metadata
				metaPath := filepath.Join(filepath.Dir(path), "_stats.json")
				var metadata map[string]interface{}
				if metaData, err := os.ReadFile(metaPath); err == nil {
					json.Unmarshal(metaData, &metadata)
				}

				fileInfo := map[string]interface{}{
					"file_path":  path,
					"file_name":  info.Name(),
					"size_bytes": info.Size(),
					"modified":   info.ModTime().Format(time.RFC3339),
					"metadata":   metadata,
				}
				files = append(files, fileInfo)
			}
			count++
		}
		return nil
	})

	return files, err
}