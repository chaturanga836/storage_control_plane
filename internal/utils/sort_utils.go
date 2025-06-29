// sort_utils.go
package utils

import (
	"fmt"
	"strings"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

// SortOptions defines available sorting configurations
type SortOptions struct {
	DefaultField     string
	DefaultDirection models.SortOrder
	AllowedFields    []string
	MaxSortFields    int
	// New fields for large-scale optimization
	IndexedFields    []string  // Fields that have database indexes
	MaxResultSize    int64     // Maximum result set size
	ForceIndexUsage  bool      // Force using indexed fields for large queries
}

// LargeScaleSortConfig defines configuration for large dataset sorting
type LargeScaleSortConfig struct {
	MaxMemoryRows    int64   // Maximum rows to sort in memory
	UseStreaming     bool    // Use streaming for large results
	ChunkSize        int     // Size of chunks for streaming
	IndexHints       map[string]string // Database index hints
	QueryTimeout     int     // Query timeout in seconds
}

// ValidateSortFields validates and sanitizes sort fields
func ValidateSortFields(sortFields []models.SortField, opts SortOptions) ([]models.SortField, error) {
	if len(sortFields) == 0 {
		// Return default sorting if none specified
		if opts.DefaultField != "" {
			return []models.SortField{{
				Field:     opts.DefaultField,
				Direction: opts.DefaultDirection,
			}}, nil
		}
		return []models.SortField{}, nil
	}

	if len(sortFields) > opts.MaxSortFields {
		return nil, fmt.Errorf("too many sort fields: max %d allowed", opts.MaxSortFields)
	}

	var validatedFields []models.SortField
	seenFields := make(map[string]bool)

	for _, field := range sortFields {
		// Check for duplicates
		if seenFields[field.Field] {
			return nil, fmt.Errorf("duplicate sort field: %s", field.Field)
		}
		seenFields[field.Field] = true

		// Validate field name
		if !isAllowedField(field.Field, opts.AllowedFields) {
			return nil, fmt.Errorf("invalid sort field: %s", field.Field)
		}

		// Validate direction
		if field.Direction != models.SortAsc && field.Direction != models.SortDesc {
			field.Direction = models.SortAsc // Default to ascending
		}

		validatedFields = append(validatedFields, field)
	}

	return validatedFields, nil
}

// GenerateClickHouseOrderBy generates ORDER BY clause for ClickHouse
func GenerateClickHouseOrderBy(sortFields []models.SortField) string {
	if len(sortFields) == 0 {
		return ""
	}

	var clauses []string
	for _, field := range sortFields {
		// Sanitize field name to prevent SQL injection
		sanitizedField := sanitizeFieldName(field.Field)
		direction := strings.ToUpper(string(field.Direction))
		clauses = append(clauses, fmt.Sprintf("`%s` %s", sanitizedField, direction))
	}

	return "ORDER BY " + strings.Join(clauses, ", ")
}

// GeneratePostgreSQLOrderBy generates ORDER BY clause for PostgreSQL
func GeneratePostgreSQLOrderBy(sortFields []models.SortField) string {
	if len(sortFields) == 0 {
		return ""
	}

	var clauses []string
	for _, field := range sortFields {
		sanitizedField := sanitizeFieldName(field.Field)
		direction := strings.ToUpper(string(field.Direction))
		clauses = append(clauses, fmt.Sprintf(`"%s" %s`, sanitizedField, direction))
	}

	return "ORDER BY " + strings.Join(clauses, ", ")
}

// Common sort field definitions for different entity types with large-scale optimizations
var (
	TenantSortOptions = SortOptions{
		DefaultField:     "created_at",
		DefaultDirection: models.SortDesc,
		AllowedFields:    []string{"created_at", "updated_at", "tenant_id", "total_files", "total_rows", "total_size_gb"},
		IndexedFields:    []string{"created_at", "updated_at", "tenant_id"}, // These should have database indexes
		MaxSortFields:    3,
		MaxResultSize:    1000000, // 1M rows max
		ForceIndexUsage:  true,    // Force index usage for large queries
	}

	DataIngestionSortOptions = SortOptions{
		DefaultField:     "created_at",
		DefaultDirection: models.SortDesc,
		AllowedFields:    []string{"created_at", "updated_at", "data_id", "tenant_id", "source_id"},
		IndexedFields:    []string{"created_at", "tenant_id", "source_id"},
		MaxSortFields:    3,
		MaxResultSize:    5000000, // 5M rows max
		ForceIndexUsage:  false,   // Allow non-indexed sorting for smaller datasets
	}

	FileSortOptions = SortOptions{
		DefaultField:     "created_at",
		DefaultDirection: models.SortDesc,
		AllowedFields:    []string{"created_at", "file_path", "record_count", "file_size", "min_timestamp", "max_timestamp"},
		IndexedFields:    []string{"created_at", "min_timestamp", "max_timestamp"},
		MaxSortFields:    3,
		MaxResultSize:    100000, // 100K files max
		ForceIndexUsage:  true,
	}

	AnalyticsSortOptions = SortOptions{
		DefaultField:     "timestamp",
		DefaultDirection: models.SortDesc,
		AllowedFields:    []string{"timestamp", "value", "count", "avg", "sum", "min", "max"},
		IndexedFields:    []string{"timestamp"},
		MaxSortFields:    5,
		MaxResultSize:    10000000, // 10M analytics points max
		ForceIndexUsage:  true,
	}

	// Default large-scale configuration
	DefaultLargeScaleConfig = LargeScaleSortConfig{
		MaxMemoryRows: 100000,  // Sort in memory only for < 100K rows
		UseStreaming:  false,   // Enable based on dataset size
		ChunkSize:     10000,   // 10K rows per chunk when streaming
		IndexHints: map[string]string{
			"created_at": "idx_created_at",
			"timestamp":  "idx_timestamp",
			"tenant_id":  "idx_tenant_id",
		},
		QueryTimeout: 30, // 30 seconds max query time
	}
)

// Helper functions
func isAllowedField(field string, allowedFields []string) bool {
	for _, allowed := range allowedFields {
		if field == allowed {
			return true
		}
	}
	return false
}

func sanitizeFieldName(field string) string {
	// Remove any potentially dangerous characters
	// Allow only alphanumeric, underscore, and dot (for nested fields)
	var result strings.Builder
	for _, char := range field {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '.' {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// BuildSortParams builds sort parameters from query strings (for HTTP APIs)
func BuildSortParams(sortBy string, sortOrder string) []models.SortField {
	if sortBy == "" {
		return []models.SortField{}
	}

	fields := strings.Split(sortBy, ",")
	orders := strings.Split(sortOrder, ",")

	var sortFields []models.SortField
	for i, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		direction := models.SortAsc
		if i < len(orders) {
			orderStr := strings.ToLower(strings.TrimSpace(orders[i]))
			if orderStr == "desc" || orderStr == "descending" {
				direction = models.SortDesc
			}
		}

		sortFields = append(sortFields, models.SortField{
			Field:     field,
			Direction: direction,
		})
	}

	return sortFields
}

// Enhanced sort validation with large-scale considerations
func ValidateSortFieldsForScale(sortFields []models.SortField, opts SortOptions, config LargeScaleSortConfig, estimatedRows int64) ([]models.SortField, *LargeScaleSortConfig, error) {
	// Standard validation first
	validatedFields, err := ValidateSortFields(sortFields, opts)
	if err != nil {
		return nil, nil, err
	}

	// Large-scale optimizations
	if estimatedRows > config.MaxMemoryRows {
		// For large datasets, prefer indexed fields
		if opts.ForceIndexUsage {
			for _, field := range validatedFields {
				if !isIndexedField(field.Field, opts.IndexedFields) {
					return nil, nil, fmt.Errorf("field %s is not indexed and cannot be used for large dataset sorting", field.Field)
				}
			}
		}
		
		// Enable streaming for large results
		optimizedConfig := config
		optimizedConfig.UseStreaming = true
		
		return validatedFields, &optimizedConfig, nil
	}

	return validatedFields, &config, nil
}

// GenerateOptimizedClickHouseQuery generates optimized ClickHouse query for large datasets
func GenerateOptimizedClickHouseQuery(baseQuery string, sortFields []models.SortField, config LargeScaleSortConfig, limit int, offset int) string {
	var query strings.Builder
	query.WriteString(baseQuery)
	
	// Add ORDER BY clause
	orderBy := GenerateClickHouseOrderBy(sortFields)
	if orderBy != "" {
		query.WriteString(" ")
		query.WriteString(orderBy)
	}
	
	// Add LIMIT and OFFSET for pagination
	if limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", limit))
		if offset > 0 {
			query.WriteString(fmt.Sprintf(" OFFSET %d", offset))
		}
	}
	
	// Add query hints for large datasets
	if config.UseStreaming && len(config.IndexHints) > 0 {
		// Add index hints if supported by ClickHouse version
		for field, hint := range config.IndexHints {
			if containsField(sortFields, field) {
				query.WriteString(fmt.Sprintf(" /* INDEX_HINT: %s */", hint))
			}
		}
	}
	
	return query.String()
}

// GenerateStreamingQuery creates a query optimized for streaming large results
func GenerateStreamingQuery(baseQuery string, sortFields []models.SortField, chunkSize int, lastValues map[string]any) (string, error) {
	if len(sortFields) == 0 {
		return baseQuery + fmt.Sprintf(" LIMIT %d", chunkSize), nil
	}
	
	var query strings.Builder
	query.WriteString(baseQuery)
	
	// Build cursor-based pagination using sort fields
	if len(lastValues) > 0 {
		query.WriteString(" WHERE ")
		var conditions []string
		
		for i, field := range sortFields {
			if lastValue, exists := lastValues[field.Field]; exists {
				operator := ">"
				if field.Direction == models.SortDesc {
					operator = "<"
				}
				
				// Handle different data types properly
				switch v := lastValue.(type) {
				case string:
					conditions = append(conditions, fmt.Sprintf("`%s` %s '%s'", sanitizeFieldName(field.Field), operator, sanitizeStringValue(v)))
				case int, int64, float64:
					conditions = append(conditions, fmt.Sprintf("`%s` %s %v", sanitizeFieldName(field.Field), operator, v))
				default:
					conditions = append(conditions, fmt.Sprintf("`%s` %s '%v'", sanitizeFieldName(field.Field), operator, v))
				}
				
				// For multi-field sorting, we need progressive conditions
				if i < len(sortFields)-1 {
					break // Simplified: use only first sort field for cursor
				}
			}
		}
		
		if len(conditions) > 0 {
			query.WriteString(strings.Join(conditions, " OR "))
		}
	}
	
	// Add ORDER BY
	orderBy := GenerateClickHouseOrderBy(sortFields)
	if orderBy != "" {
		query.WriteString(" ")
		query.WriteString(orderBy)
	}
	
	// Add LIMIT for chunk size
	query.WriteString(fmt.Sprintf(" LIMIT %d", chunkSize))
	
	return query.String(), nil
}

// Helper functions for large-scale operations
func isIndexedField(field string, indexedFields []string) bool {
	for _, indexed := range indexedFields {
		if field == indexed {
			return true
		}
	}
	return false
}

func containsField(sortFields []models.SortField, fieldName string) bool {
	for _, field := range sortFields {
		if field.Field == fieldName {
			return true
		}
	}
	return false
}

func sanitizeStringValue(value string) string {
	// Escape single quotes and other SQL injection attempts
	return strings.ReplaceAll(value, "'", "''")
}

// Performance monitoring for sort operations
type SortPerformanceMetrics struct {
	QueryTime       int64  // milliseconds
	RowsProcessed   int64
	MemoryUsed      int64  // bytes
	IndexesUsed     []string
	StreamingUsed   bool
	ChunksProcessed int
}

// EstimateQueryComplexity estimates the complexity of a sort operation
func EstimateQueryComplexity(sortFields []models.SortField, estimatedRows int64, config LargeScaleSortConfig) (complexity string, recommendations []string) {
	var recs []string
	
	if estimatedRows < 10000 {
		return "LOW", []string{"Standard in-memory sorting suitable"}
	}
	
	if estimatedRows < 1000000 {
		complexity = "MEDIUM"
		recs = append(recs, "Consider using indexed fields for better performance")
		if len(sortFields) > 2 {
			recs = append(recs, "Multiple sort fields may impact performance")
		}
	} else {
		complexity = "HIGH"
		recs = append(recs, "Use streaming and cursor-based pagination")
		recs = append(recs, "Ensure sort fields are indexed")
		recs = append(recs, "Consider data partitioning strategies")
		
		if !config.UseStreaming {
			recs = append(recs, "Enable streaming mode for large datasets")
		}
	}
	
	return complexity, recs
}
