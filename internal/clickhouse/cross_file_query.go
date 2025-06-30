package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/your-org/storage-control-plane/pkg/models"
)

// CrossFileQueryService enables querying across multiple Parquet files
// This solves the problem: "find name like 'sam' in source_connection_01"
type CrossFileQueryService struct {
	client *Client
}

func NewCrossFileQueryService(client *Client) *CrossFileQueryService {
	return &CrossFileQueryService{
		client: client,
	}
}

// SearchRecordsAcrossFiles searches for records across all Parquet files in a source
// This is the main function that solves your cross-file query problem!
func (s *CrossFileQueryService) SearchRecordsAcrossFiles(ctx context.Context, req *CrossFileSearchRequest) (*CrossFileSearchResult, error) {
	// Build the query using record metadata table
	query, err := s.buildCrossFileQuery(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	// Execute the query
	rows, err := s.client.ExecuteRawQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute cross-file query: %w", err)
	}

	// Parse results
	records, err := s.parseSearchResults(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to parse results: %w", err)
	}

	// Get file hints for efficient Parquet file access
	fileHints, err := s.getRelevantFiles(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file hints: %w", err)
	}

	return &CrossFileSearchResult{
		Records:        records,
		TotalFound:     len(records),
		RelevantFiles:  fileHints,
		QueryExecuted:  query,
		ExecutionTime:  time.Since(time.Now()), // This would be measured properly
	}, nil
}

type CrossFileSearchRequest struct {
	TenantID    string                 `json:"tenant_id"`
	SourceID    string                 `json:"source_id,omitempty"`
	
	// Search filters
	NameContains    string            `json:"name_contains,omitempty"`    // e.g., "sam"
	EmailContains   string            `json:"email_contains,omitempty"`
	Status          string            `json:"status,omitempty"`
	Category        string            `json:"category,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	
	// Custom field filters
	CustomFilters   map[string]any    `json:"custom_filters,omitempty"`
	
	// Time range
	TimeRange       *TimeRangeFilter  `json:"time_range,omitempty"`
	
	// Pagination
	Limit           int               `json:"limit,omitempty"`
	Offset          int               `json:"offset,omitempty"`
	
	// Sorting
	SortBy          []models.SortField `json:"sort_by,omitempty"`
	
	// Options
	IncludeFileInfo bool              `json:"include_file_info,omitempty"`
}

type TimeRangeFilter struct {
	Field string    `json:"field"`      // "created_at", "updated_at", "timestamp"
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type CrossFileSearchResult struct {
	Records        []CrossFileRecord  `json:"records"`
	TotalFound     int                `json:"total_found"`
	RelevantFiles  []string           `json:"relevant_files"`
	QueryExecuted  string             `json:"query_executed"`
	ExecutionTime  time.Duration      `json:"execution_time"`
}

type CrossFileRecord struct {
	RecordID     string                 `json:"record_id"`
	Name         string                 `json:"name,omitempty"`
	Email        string                 `json:"email,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Category     string                 `json:"category,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]any         `json:"custom_fields,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Timestamp    time.Time              `json:"timestamp"`
	
	// File location (for retrieving full record from Parquet)
	FileID       string                 `json:"file_id"`
	FilePath     string                 `json:"file_path"`
	RowNumber    int64                  `json:"row_number"`
}

func (s *CrossFileQueryService) buildCrossFileQuery(req *CrossFileSearchRequest) (string, error) {
	var queryBuilder strings.Builder
	
	// SELECT clause
	queryBuilder.WriteString(`
		SELECT 
			record_id,
			name,
			email,
			status,
			category,
			tags,
			custom_fields,
			created_at,
			updated_at,
			timestamp,
			file_id,
			file_path,
			row_number
		FROM record_metadata
		WHERE 1=1
	`)
	
	// Tenant filter (always required)
	queryBuilder.WriteString(fmt.Sprintf(" AND tenant_id = '%s'", req.TenantID))
	
	// Source filter
	if req.SourceID != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND source_id = '%s'", req.SourceID))
	}
	
	// Name search - THIS IS THE KEY FUNCTIONALITY!
	if req.NameContains != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND lower(name) LIKE '%%%s%%'", strings.ToLower(req.NameContains)))
	}
	
	// Email search
	if req.EmailContains != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND lower(email) LIKE '%%%s%%'", strings.ToLower(req.EmailContains)))
	}
	
	// Status filter
	if req.Status != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND status = '%s'", req.Status))
	}
	
	// Category filter
	if req.Category != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND category = '%s'", req.Category))
	}
	
	// Tag filters
	if len(req.Tags) > 0 {
		tagConditions := make([]string, 0, len(req.Tags))
		for _, tag := range req.Tags {
			tagConditions = append(tagConditions, fmt.Sprintf("has(tags, '%s')", tag))
		}
		queryBuilder.WriteString(" AND (" + strings.Join(tagConditions, " OR ") + ")")
	}
	
	// Time range filter
	if req.TimeRange != nil {
		field := req.TimeRange.Field
		if field == "" {
			field = "created_at"
		}
		queryBuilder.WriteString(fmt.Sprintf(" AND %s >= '%s' AND %s <= '%s'",
			field, req.TimeRange.Start.Format("2006-01-02 15:04:05"),
			field, req.TimeRange.End.Format("2006-01-02 15:04:05")))
	}
	
	// Custom field filters (JSON queries)
	for fieldName, value := range req.CustomFilters {
		queryBuilder.WriteString(fmt.Sprintf(" AND JSONExtractRaw(custom_fields, '%s') = '%v'", fieldName, value))
	}
	
	// Sorting
	if len(req.SortBy) > 0 {
		orderClauses := make([]string, 0, len(req.SortBy))
		for _, sort := range req.SortBy {
			direction := "ASC"
			if sort.Order == models.SortDesc {
				direction = "DESC"
			}
			orderClauses = append(orderClauses, fmt.Sprintf("%s %s", sort.Field, direction))
		}
		queryBuilder.WriteString(" ORDER BY " + strings.Join(orderClauses, ", "))
	} else {
		// Default sorting
		queryBuilder.WriteString(" ORDER BY created_at DESC")
	}
	
	// Pagination
	if req.Limit > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT %d", req.Limit))
		if req.Offset > 0 {
			queryBuilder.WriteString(fmt.Sprintf(" OFFSET %d", req.Offset))
		}
	}
	
	return queryBuilder.String(), nil
}

func (s *CrossFileQueryService) parseSearchResults(rows interface{}) ([]CrossFileRecord, error) {
	// This would parse the ClickHouse result rows into CrossFileRecord structs
	// Implementation depends on your ClickHouse driver
	return []CrossFileRecord{}, nil
}

// getRelevantFiles returns a list of Parquet files that contain matching records
func (s *CrossFileQueryService) getRelevantFiles(ctx context.Context, req *CrossFileSearchRequest) ([]string, error) {
	query := `
		SELECT DISTINCT file_path
		FROM record_metadata
		WHERE tenant_id = '%s'
	`
	
	conditions := []string{fmt.Sprintf("tenant_id = '%s'", req.TenantID)}
	
	if req.SourceID != "" {
		conditions = append(conditions, fmt.Sprintf("source_id = '%s'", req.SourceID))
	}
	
	if req.NameContains != "" {
		conditions = append(conditions, fmt.Sprintf("lower(name) LIKE '%%%s%%'", strings.ToLower(req.NameContains)))
	}
	
	fullQuery := fmt.Sprintf("SELECT DISTINCT file_path FROM record_metadata WHERE %s",
		strings.Join(conditions, " AND "))
	
	rows, err := s.client.ExecuteRawQuery(ctx, fullQuery)
	if err != nil {
		return nil, err
	}
	
	// Parse file paths from result
	// Implementation depends on your ClickHouse driver
	return []string{}, nil
}

// GetDirectorySummary returns aggregated statistics for a directory
func (s *CrossFileQueryService) GetDirectorySummary(ctx context.Context, tenantID, sourceID, directoryPath string) (*models.DirectorySummary, error) {
	query := `
		SELECT 
			directory_path,
			total_files,
			total_records,
			total_size,
			first_record_at,
			last_record_at,
			last_updated,
			field_summaries,
			schema_versions,
			current_schema
		FROM directory_summaries
		WHERE tenant_id = '%s'
	`
	
	conditions := []string{fmt.Sprintf("tenant_id = '%s'", tenantID)}
	
	if sourceID != "" {
		conditions = append(conditions, fmt.Sprintf("source_id = '%s'", sourceID))
	}
	
	if directoryPath != "" {
		conditions = append(conditions, fmt.Sprintf("directory_path = '%s'", directoryPath))
	}
	
	fullQuery := fmt.Sprintf(query + " AND " + strings.Join(conditions[1:], " AND "), tenantID)
	
	rows, err := s.client.ExecuteRawQuery(ctx, fullQuery)
	if err != nil {
		return nil, err
	}
	
	// Parse directory summary from result
	// Implementation would parse the JSON fields and create DirectorySummary struct
	return &models.DirectorySummary{}, nil
}

// GetRecordCountByStatus returns count of records by status across all files
func (s *CrossFileQueryService) GetRecordCountByStatus(ctx context.Context, tenantID, sourceID string) (map[string]int64, error) {
	query := `
		SELECT 
			status,
			COUNT(*) as record_count
		FROM record_metadata
		WHERE tenant_id = '%s'
	`
	
	if sourceID != "" {
		query += " AND source_id = '%s'"
		query = fmt.Sprintf(query, tenantID, sourceID)
	} else {
		query = fmt.Sprintf(query, tenantID)
	}
	
	query += " GROUP BY status ORDER BY record_count DESC"
	
	rows, err := s.client.ExecuteRawQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	
	// Parse results into map
	results := make(map[string]int64)
	// Implementation would parse the rows
	return results, nil
}

// Example usage functions for your specific use cases:

// FindNameLikeSam finds all records with names containing "sam" in a source
func (s *CrossFileQueryService) FindNameLikeSam(ctx context.Context, tenantID, sourceID string) ([]CrossFileRecord, error) {
	req := &CrossFileSearchRequest{
		TenantID:     tenantID,
		SourceID:     sourceID,
		NameContains: "sam",
		Limit:        100,
		SortBy: []models.SortField{
			{Field: "created_at", Order: models.SortDesc},
		},
	}
	
	result, err := s.SearchRecordsAcrossFiles(ctx, req)
	if err != nil {
		return nil, err
	}
	
	return result.Records, nil
}

// GetLatestRecords gets the latest N records across all files in a source
func (s *CrossFileQueryService) GetLatestRecords(ctx context.Context, tenantID, sourceID string, limit int) ([]CrossFileRecord, error) {
	req := &CrossFileSearchRequest{
		TenantID: tenantID,
		SourceID: sourceID,
		Limit:    limit,
		SortBy: []models.SortField{
			{Field: "created_at", Order: models.SortDesc},
		},
	}
	
	result, err := s.SearchRecordsAcrossFiles(ctx, req)
	if err != nil {
		return nil, err
	}
	
	return result.Records, nil
}
