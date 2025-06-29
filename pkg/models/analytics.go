package models

import "time"

// Query and Analytics Models
type QueryRequest struct {
	TenantID    string            `json:"tenant_id"`
	SourceID    string            `json:"source_id,omitempty"`
	Query       string            `json:"query"`
	QueryType   QueryType         `json:"query_type"`
	Filters     map[string]any    `json:"filters,omitempty"`
	TimeRange   *TimeRange        `json:"time_range,omitempty"`
	Limit       int               `json:"limit,omitempty"`
	Offset      int               `json:"offset,omitempty"`
	Format      ResponseFormat    `json:"format"`
}

type QueryType string

const (
	QuerySQL        QueryType = "sql"
	QueryAggregate  QueryType = "aggregate"
	QueryTimeSeries QueryType = "timeseries"
	QuerySearch     QueryType = "search"
)

type ResponseFormat string

const (
	FormatJSON    ResponseFormat = "json"
	FormatCSV     ResponseFormat = "csv"
	FormatParquet ResponseFormat = "parquet"
)

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type QueryResponse struct {
	QueryID     string         `json:"query_id"`
	Data        []any          `json:"data"`
	Schema      []ColumnInfo   `json:"schema"`
	RowCount    int64          `json:"row_count"`
	ExecutionMS int64          `json:"execution_ms"`
	FromCache   bool           `json:"from_cache"`
	NextToken   string         `json:"next_token,omitempty"`
}

type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

type TenantSummary struct {
	TenantID      string    `json:"tenant_id"`
	TotalFiles    int64     `json:"total_files"`
	TotalRows     int64     `json:"total_rows"`
	TotalSizeGB   float64   `json:"total_size_gb"`
	SourceCount   int       `json:"source_count"`
	OldestRecord  time.Time `json:"oldest_record"`
	NewestRecord  time.Time `json:"newest_record"`
	LastUpdated   time.Time `json:"last_updated"`
}

type SourceSummary struct {
	SourceID      string    `json:"source_id"`
	TenantID      string    `json:"tenant_id"`
	RecordCount   int64     `json:"record_count"`
	FileCount     int64     `json:"file_count"`
	SizeGB        float64   `json:"size_gb"`
	SchemaVersions int      `json:"schema_versions"`
	LastIngestion time.Time `json:"last_ingestion"`
	Status        string    `json:"status"`
}
