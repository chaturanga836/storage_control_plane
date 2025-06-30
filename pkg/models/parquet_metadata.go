package models

import "time"

// ParquetFileMetadata represents metadata for each Parquet file stored in ClickHouse
// This allows ClickHouse to search across multiple files without reading the actual Parquet data
type ParquetFileMetadata struct {
	// File identification
	FileID          string    `json:"file_id"`
	TenantID        string    `json:"tenant_id"`
	SourceID        string    `json:"source_id"`
	FilePath        string    `json:"file_path"`
	DirectoryPath   string    `json:"directory_path"`
	
	// File metadata
	FileSize        int64     `json:"file_size"`
	RecordCount     int64     `json:"record_count"`
	CreatedAt       time.Time `json:"created_at"`
	SchemaHash      string    `json:"schema_hash"`
	
	// Data boundaries for optimization
	MinTimestamp    *time.Time `json:"min_timestamp,omitempty"`
	MaxTimestamp    *time.Time `json:"max_timestamp,omitempty"`
	MinCreatedAt    *time.Time `json:"min_created_at,omitempty"`
	MaxCreatedAt    *time.Time `json:"max_created_at,omitempty"`
	
	// Indexed fields for fast searching
	IndexedFields   map[string]FieldIndex `json:"indexed_fields"`
	
	// Statistics
	Stats           FileStatistics `json:"stats"`
}

// FieldIndex contains searchable values for a specific field across the file
type FieldIndex struct {
	FieldName    string   `json:"field_name"`
	FieldType    string   `json:"field_type"`
	UniqueValues []string `json:"unique_values,omitempty"` // For fields with limited values
	MinValue     any      `json:"min_value,omitempty"`
	MaxValue     any      `json:"max_value,omitempty"`
	HasNulls     bool     `json:"has_nulls"`
	BloomFilter  string   `json:"bloom_filter,omitempty"`  // Base64 encoded bloom filter for text fields
}

// FileStatistics contains aggregated statistics for the file
type FileStatistics struct {
	DistinctCounts  map[string]int64  `json:"distinct_counts"`
	NullCounts      map[string]int64  `json:"null_counts"`
	ValueCounts     map[string]map[string]int64 `json:"value_counts"` // field -> value -> count
	NumericStats    map[string]NumericStat `json:"numeric_stats"`
}

type NumericStat struct {
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Avg    float64 `json:"avg"`
	Sum    float64 `json:"sum"`
	StdDev float64 `json:"stddev"`
}

// RecordMetadata represents searchable metadata for each record (for fast searching)
// Only stores key fields that are commonly searched
type RecordMetadata struct {
	// Record identification
	RecordID     string    `json:"record_id"`
	TenantID     string    `json:"tenant_id"`
	SourceID     string    `json:"source_id"`
	FileID       string    `json:"file_id"`
	FilePath     string    `json:"file_path"`
	
	// Searchable fields - these are the fields users commonly search by
	Name         string    `json:"name,omitempty"`         // For name searches like "sam"
	Email        string    `json:"email,omitempty"`        // For email searches
	Status       string    `json:"status,omitempty"`       // For status filtering
	Category     string    `json:"category,omitempty"`     // For category filtering
	Tags         []string  `json:"tags,omitempty"`         // For tag-based searches
	
	// Timestamps for temporal queries
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Timestamp    time.Time `json:"timestamp"`
	
	// Custom indexed fields (configured per source)
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	
	// Record location in file (for efficient retrieval)
	RowNumber    int64     `json:"row_number"`
	Offset       int64     `json:"offset,omitempty"`
}

// DirectorySummary provides aggregated information for a directory
type DirectorySummary struct {
	// Directory identification
	TenantID        string    `json:"tenant_id"`
	SourceID        string    `json:"source_id"`
	DirectoryPath   string    `json:"directory_path"`
	
	// Summary statistics
	TotalFiles      int64     `json:"total_files"`
	TotalRecords    int64     `json:"total_records"`
	TotalSize       int64     `json:"total_size"`
	
	// Time boundaries
	FirstRecordAt   time.Time `json:"first_record_at"`
	LastRecordAt    time.Time `json:"last_record_at"`
	LastUpdated     time.Time `json:"last_updated"`
	
	// Field summaries
	FieldSummaries  map[string]FieldSummary `json:"field_summaries"`
	
	// Schema information
	SchemaVersions  []string  `json:"schema_versions"`
	CurrentSchema   string    `json:"current_schema"`
}

type FieldSummary struct {
	FieldName       string   `json:"field_name"`
	FieldType       string   `json:"field_type"`
	UniqueValues    int64    `json:"unique_values"`
	NullCount       int64    `json:"null_count"`
	MostCommonValues []ValueCount `json:"most_common_values"`
}

type ValueCount struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// QueryHint helps ClickHouse know which files to check for a query
type QueryHint struct {
	TenantID      string              `json:"tenant_id"`
	SourceID      string              `json:"source_id,omitempty"`
	DirectoryPath string              `json:"directory_path,omitempty"`
	FieldFilters  map[string]any      `json:"field_filters,omitempty"`
	TimeRange     *TimeRange          `json:"time_range,omitempty"`
	RelevantFiles []string            `json:"relevant_files"`
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
