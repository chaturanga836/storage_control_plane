package models

import "time"

// Data Pipeline Models
type DataRecord struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id"`
	SourceID     string            `json:"source_id"`
	Data         map[string]any    `json:"data"`
	Version      int64             `json:"version"`
	SchemaHash   string            `json:"schema_hash"`
	IngestedAt   time.Time         `json:"ingested_at"`
	Metadata     RecordMetadata    `json:"metadata"`
}

type RecordMetadata struct {
	SourceTimestamp *time.Time        `json:"source_timestamp,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
	ContentType     string            `json:"content_type,omitempty"`
	Size            int64             `json:"size,omitempty"`
}

type ParquetFile struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id"`
	SourceID     string            `json:"source_id"`
	FilePath     string            `json:"file_path"`
	SchemaHash   string            `json:"schema_hash"`
	RecordCount  int64             `json:"record_count"`
	FileSize     int64             `json:"file_size"`
	MinTimestamp time.Time         `json:"min_timestamp"`
	MaxTimestamp time.Time         `json:"max_timestamp"`
	CreatedAt    time.Time         `json:"created_at"`
	Compressed   bool              `json:"compressed"`
	Stats        FileStats         `json:"stats"`
}

type FileStats struct {
	ColumnStats    map[string]ColumnStat `json:"column_stats"`
	Partitions     []string              `json:"partitions"`
	BloomFilters   []string              `json:"bloom_filters,omitempty"`
	IndexColumns   []string              `json:"index_columns,omitempty"`
}

type ColumnStat struct {
	Type         string  `json:"type"`
	MinValue     any     `json:"min_value,omitempty"`
	MaxValue     any     `json:"max_value,omitempty"`
	NullCount    int64   `json:"null_count"`
	UniqueCount  int64   `json:"unique_count,omitempty"`
	AvgLength    float64 `json:"avg_length,omitempty"`
}

type SchemaVersion struct {
	Hash           string            `json:"hash"`
	TenantID       string            `json:"tenant_id"`
	SourceID       string            `json:"source_id"`
	Version        int               `json:"version"`
	Schema         map[string]string `json:"schema"` // field_name -> type
	FirstSeen      time.Time         `json:"first_seen"`
	LastSeen       time.Time         `json:"last_seen"`
	BackwardCompat bool              `json:"backward_compatible"`
	ForwardCompat  bool              `json:"forward_compatible"`
}

type WALEntry struct {
	ID         string         `json:"id"`
	TenantID   string         `json:"tenant_id"`
	SourceID   string         `json:"source_id"`
	Timestamp  time.Time      `json:"timestamp"`
	Data       map[string]any `json:"data"`
	Flushed    bool           `json:"flushed"`
	FlushTime  *time.Time     `json:"flush_time,omitempty"`
}
