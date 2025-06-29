// large_scale_query.go
package clickhouse

import (
	"context"
	"fmt"
	"time"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/utils"
)

// LargeScaleQueryExecutor handles efficient querying of large datasets
type LargeScaleQueryExecutor struct {
	store  *Store
	config utils.LargeScaleSortConfig
}

// NewLargeScaleQueryExecutor creates a new large-scale query executor
func NewLargeScaleQueryExecutor(store *Store, config utils.LargeScaleSortConfig) *LargeScaleQueryExecutor {
	return &LargeScaleQueryExecutor{
		store:  store,
		config: config,
	}
}

// ExecuteLargeQuery executes a query optimized for large datasets
func (e *LargeScaleQueryExecutor) ExecuteLargeQuery(ctx context.Context, req models.QueryRequest) (*models.QueryResponse, error) {
	startTime := time.Now()
	
	// Estimate dataset size
	estimatedRows, err := e.estimateResultSize(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate result size: %w", err)
	}
	
	// Choose execution strategy based on size
	if estimatedRows > e.config.MaxMemoryRows {
		return e.executeStreamingQuery(ctx, req, estimatedRows)
	}
	
	return e.executeStandardQuery(ctx, req, estimatedRows, startTime)
}

// executeStreamingQuery handles large datasets with streaming
func (e *LargeScaleQueryExecutor) executeStreamingQuery(ctx context.Context, req models.QueryRequest, estimatedRows int64) (*models.QueryResponse, error) {
	// Use cursor-based pagination for very large results
	var allResults []map[string]any
	var lastValues map[string]any
	totalRows := int64(0)
	chunksProcessed := 0
	
	for {
		// Generate streaming query with cursor
		query, err := utils.GenerateStreamingQuery(req.Query, req.SortBy, e.config.ChunkSize, lastValues)
		if err != nil {
			return nil, fmt.Errorf("failed to generate streaming query: %w", err)
		}
		
		// Execute chunk
		chunkResults, err := e.store.executeQuery(query)
		if err != nil {
			return nil, fmt.Errorf("failed to execute chunk query: %w", err)
		}
		
		// Break if no more results
		if len(chunkResults) == 0 {
			break
		}
		
		// Accumulate results
		allResults = append(allResults, chunkResults...)
		totalRows += int64(len(chunkResults))
		chunksProcessed++
		
		// Update cursor for next iteration
		if len(chunkResults) > 0 {
			lastResult := chunkResults[len(chunkResults)-1]
			lastValues = make(map[string]any)
			
			// Extract values for cursor-based pagination
			for _, sortField := range req.SortBy {
				if val, exists := lastResult[sortField.Field]; exists {
					lastValues[sortField.Field] = val
				}
			}
		}
		
		// Safety check to prevent infinite loops
		if chunksProcessed > 1000 { // Max 1000 chunks
			break
		}
		
		// Check if we have enough results (if limit specified)
		if req.Limit > 0 && totalRows >= int64(req.Limit) {
			allResults = allResults[:req.Limit]
			break
		}
	}
	
	// Convert results to []any for the response
	var responseData []any
	for _, result := range allResults {
		responseData = append(responseData, result)
	}
	
	return &models.QueryResponse{
		QueryID:     generateQueryID(req.TenantID),
		Data:        responseData,
		Schema:      []models.ColumnInfo{}, // TODO: Extract schema
		RowCount:    int64(len(allResults)),
		ExecutionMS: time.Since(time.Now()).Milliseconds(),
		FromCache:   false,
		NextToken:   "", // Not needed for streaming approach
	}, nil
}

// executeStandardQuery handles smaller datasets with standard querying
func (e *LargeScaleQueryExecutor) executeStandardQuery(ctx context.Context, req models.QueryRequest, estimatedRows int64, startTime time.Time) (*models.QueryResponse, error) {
	// Generate optimized query
	query := utils.GenerateOptimizedClickHouseQuery(req.Query, req.SortBy, e.config, req.Limit, req.Offset)
	
	// Execute query
	results, err := e.store.executeQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	
	// Convert results to []any
	var resultData []any
	for _, result := range results {
		resultData = append(resultData, result)
	}
	
	return &models.QueryResponse{
		QueryID:     generateQueryID(req.TenantID),
		Data:        resultData,
		Schema:      []models.ColumnInfo{}, // TODO: Extract schema
		RowCount:    int64(len(results)),
		ExecutionMS: time.Since(startTime).Milliseconds(),
		FromCache:   false,
	}, nil
}

// estimateResultSize estimates the number of rows a query will return
func (e *LargeScaleQueryExecutor) estimateResultSize(ctx context.Context, req models.QueryRequest) (int64, error) {
	// Create a COUNT query based on the original query
	countQuery := "SELECT COUNT(*) as row_count FROM (" + req.Query + ") as subquery"
	
	results, err := e.store.executeQuery(countQuery)
	if err != nil {
		// If estimation fails, assume it's a large dataset
		return e.config.MaxMemoryRows + 1, nil
	}
	
	if len(results) > 0 && len(results[0]) > 0 {
		if count, ok := results[0]["row_count"].(int64); ok {
			return count, nil
		}
	}
	
	// Default to large dataset assumption
	return e.config.MaxMemoryRows + 1, nil
}

// generateQueryID creates a unique query ID
func generateQueryID(tenantID string) string {
	return fmt.Sprintf("query_%s_%d", tenantID, time.Now().UnixNano())
}

// QueryOptimizer provides query optimization recommendations
type QueryOptimizer struct {
	config utils.LargeScaleSortConfig
}

// OptimizeQuery analyzes and optimizes a query for large datasets
func (o *QueryOptimizer) OptimizeQuery(req models.QueryRequest, estimatedRows int64) (*QueryOptimization, error) {
	optimization := &QueryOptimization{
		OriginalQuery:     req.Query,
		EstimatedRows:     estimatedRows,
		Recommendations:   []string{},
		OptimizedSortBy:   req.SortBy,
		SuggestedIndexes:  []string{},
		UseStreaming:      estimatedRows > o.config.MaxMemoryRows,
	}
	
	// Analyze sort fields
	complexity, recommendations := utils.EstimateQueryComplexity(req.SortBy, estimatedRows, o.config)
	optimization.Complexity = complexity
	optimization.Recommendations = recommendations
	
	// Suggest indexes for sort fields
	for _, sortField := range req.SortBy {
		optimization.SuggestedIndexes = append(optimization.SuggestedIndexes, 
			fmt.Sprintf("CREATE INDEX idx_%s ON table_name (%s)", sortField.Field, sortField.Field))
	}
	
	// Add specific optimizations based on query patterns
	if estimatedRows > 1000000 {
		optimization.Recommendations = append(optimization.Recommendations,
			"Consider data partitioning by time or tenant",
			"Use materialized views for frequently accessed aggregations",
			"Implement query result caching",
		)
	}
	
	return optimization, nil
}

// QueryOptimization contains optimization analysis and recommendations
type QueryOptimization struct {
	OriginalQuery     string                `json:"original_query"`
	EstimatedRows     int64                 `json:"estimated_rows"`
	Complexity        string                `json:"complexity"`
	Recommendations   []string              `json:"recommendations"`
	OptimizedSortBy   []models.SortField    `json:"optimized_sort_by"`
	SuggestedIndexes  []string              `json:"suggested_indexes"`
	UseStreaming      bool                  `json:"use_streaming"`
}

// PerformanceMonitor tracks query performance for optimization
type PerformanceMonitor struct {
	metrics map[string]*utils.SortPerformanceMetrics
}

// RecordQueryPerformance records performance metrics for a query
func (p *PerformanceMonitor) RecordQueryPerformance(queryID string, metrics *utils.SortPerformanceMetrics) {
	if p.metrics == nil {
		p.metrics = make(map[string]*utils.SortPerformanceMetrics)
	}
	p.metrics[queryID] = metrics
}

// GetPerformanceInsights provides insights based on recorded metrics
func (p *PerformanceMonitor) GetPerformanceInsights() []string {
	insights := []string{}
	
	if len(p.metrics) == 0 {
		return insights
	}
	
	// Analyze patterns in recorded metrics
	totalQueries := len(p.metrics)
	slowQueries := 0
	avgQueryTime := int64(0)
	
	for _, metric := range p.metrics {
		avgQueryTime += metric.QueryTime
		if metric.QueryTime > 5000 { // 5 seconds
			slowQueries++
		}
	}
	
	avgQueryTime /= int64(totalQueries)
	
	if slowQueries > totalQueries/2 {
		insights = append(insights, "More than 50% of queries are slow (>5s)")
		insights = append(insights, "Consider adding more database indexes")
		insights = append(insights, "Optimize query patterns and enable streaming")
	}
	
	if avgQueryTime > 2000 {
		insights = append(insights, fmt.Sprintf("Average query time is high: %dms", avgQueryTime))
	}
	
	return insights
}
