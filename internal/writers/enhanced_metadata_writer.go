// enhanced_metadata_writer.go - Enhanced metadata writer for the new system
package writers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

// EnhancedMetadataWriter writes metadata to ClickHouse with full directory structure support
type EnhancedMetadataWriter struct {
	conn clickhouse.Conn
}

// NewEnhancedMetadataWriter creates a new enhanced metadata writer
func NewEnhancedMetadataWriter(conn clickhouse.Conn) *EnhancedMetadataWriter {
	return &EnhancedMetadataWriter{
		conn: conn,
	}
}

// WriteParquetFileMetadata writes Parquet file metadata to ClickHouse
func (w *EnhancedMetadataWriter) WriteParquetFileMetadata(metadata *FileMetadata) error {
	ctx := context.Background()

	// Insert into parquet_files table
	query := `
		INSERT INTO parquet_files (
			id, tenant_id, source_id, file_path, schema_hash,
			record_count, file_size, min_timestamp, max_timestamp, created_at,
			compressed, compression_ratio
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	compressionRatio := w.calculateCompressionRatio(metadata)

	err := w.conn.Exec(ctx, query,
		uuid.New().String(),
		metadata.TenantID,
		metadata.SourceID,
		metadata.RelativePath, // Store relative path for portability
		metadata.SchemaHash,
		metadata.RecordCount,
		metadata.FileSize,
		metadata.MinTimestamp,
		metadata.MaxTimestamp,
		metadata.CreatedAt,
		1, // compressed flag
		compressionRatio,
	)

	if err != nil {
		return fmt.Errorf("failed to insert parquet file metadata: %w", err)
	}

	// Update tenant metadata summary
	if err := w.updateTenantMetadata(metadata); err != nil {
		return fmt.Errorf("failed to update tenant metadata: %w", err)
	}

	// Write schema version if new
	if err := w.writeSchemaVersion(metadata); err != nil {
		return fmt.Errorf("failed to write schema version: %w", err)
	}

	return nil
}

// WriteDirectoryStructure writes directory structure metadata
func (w *EnhancedMetadataWriter) WriteDirectoryStructure(dirConfig *models.DirectoryConfig, dirPath string, summary *DirectorySummary) error {
	ctx := context.Background()

	// Create a new table for directory structures if it doesn't exist
	createTable := `
		CREATE TABLE IF NOT EXISTS directory_structures (
			id String,
			tenant_id String,
			source_connection_id String,
			directory_path String,
			config_pattern String,
			total_files UInt64,
			total_records UInt64,
			total_size_bytes UInt64,
			schema_hash String,
			min_timestamp DateTime,
			max_timestamp DateTime,
			last_updated DateTime,
			unique_schemas UInt32,
			compression_ratio Float64,
			metadata_generated UInt8,
			
			INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
			INDEX idx_directory_path directory_path TYPE bloom_filter GRANULARITY 1,
			INDEX idx_last_updated last_updated TYPE minmax GRANULARITY 1
		) ENGINE = ReplacingMergeTree(last_updated)
		ORDER BY (tenant_id, source_connection_id, directory_path)
		PARTITION BY tenant_id
	`

	if err := w.conn.Exec(ctx, createTable); err != nil {
		return fmt.Errorf("failed to create directory_structures table: %w", err)
	}

	// Insert or update directory structure
	query := `
		INSERT INTO directory_structures (
			id, tenant_id, source_connection_id, directory_path, config_pattern,
			total_files, total_records, total_size_bytes, schema_hash,
			min_timestamp, max_timestamp, last_updated, unique_schemas,
			compression_ratio, metadata_generated
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	err := w.conn.Exec(ctx, query,
		uuid.New().String(),
		dirConfig.TenantID,
		dirConfig.SourceConnectionID,
		dirPath,
		dirConfig.Pattern,
		summary.TotalFiles,
		summary.TotalRecords,
		summary.TotalSizeBytes,
		summary.SchemaHash,
		summary.MinTimestamp,
		summary.MaxTimestamp,
		summary.LastUpdated,
		summary.UniqueSchemas,
		summary.CompressionRatio,
		1, // metadata_generated flag
	)

	return err
}

// WriteSourceConnectionMetrics updates source connection performance metrics
func (w *EnhancedMetadataWriter) WriteSourceConnectionMetrics(sourceConnectionID string, metrics *models.SourceMetrics) error {
	ctx := context.Background()

	// Create metrics table if it doesn't exist
	createTable := `
		CREATE TABLE IF NOT EXISTS source_connection_metrics (
			id String,
			source_connection_id String,
			timestamp DateTime,
			total_records_processed UInt64,
			total_bytes_processed UInt64,
			total_parquet_files UInt32,
			average_records_per_file UInt32,
			last_sync_duration_ms UInt64,
			average_sync_duration_ms UInt64,
			error_count UInt32,
			last_error String,
			last_error_time Nullable(DateTime),
			success_rate Float64,
			
			INDEX idx_source_connection_id source_connection_id TYPE bloom_filter GRANULARITY 1,
			INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1
		) ENGINE = MergeTree()
		ORDER BY (source_connection_id, timestamp)
		PARTITION BY toYYYYMM(timestamp)
	`

	if err := w.conn.Exec(ctx, createTable); err != nil {
		return fmt.Errorf("failed to create source_connection_metrics table: %w", err)
	}

	// Insert metrics
	query := `
		INSERT INTO source_connection_metrics (
			id, source_connection_id, timestamp, total_records_processed,
			total_bytes_processed, total_parquet_files, average_records_per_file,
			last_sync_duration_ms, average_sync_duration_ms, error_count,
			last_error, last_error_time, success_rate
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var lastErrorTime *time.Time
	if metrics.LastErrorTime != nil {
		lastErrorTime = metrics.LastErrorTime
	}

	err := w.conn.Exec(ctx, query,
		uuid.New().String(),
		sourceConnectionID,
		time.Now(),
		metrics.TotalRecordsProcessed,
		metrics.TotalBytesProcessed,
		metrics.TotalParquetFiles,
		metrics.AverageRecordsPerFile,
		metrics.LastSyncDuration,
		metrics.AverageSyncDuration,
		metrics.ErrorCount,
		metrics.LastError,
		lastErrorTime,
		metrics.SuccessRate,
	)

	return err
}

// QueryDataSources provides intelligent query routing based on metadata
func (w *EnhancedMetadataWriter) QueryDataSources(ctx context.Context, req *DataSourceQuery) (*DataSourceQueryResult, error) {
	// Query parquet files that match the criteria
	query := `
		SELECT 
			file_path,
			schema_hash,
			record_count,
			min_timestamp,
			max_timestamp,
			file_size
		FROM parquet_files 
		WHERE tenant_id = ?
		  AND min_timestamp <= ?
		  AND max_timestamp >= ?
	`

	args := []interface{}{
		req.TenantID,
		req.EndTime,
		req.StartTime,
	}

	// Add source filter if specified
	if req.SourceConnectionID != "" {
		query += " AND source_id = ?"
		args = append(args, req.SourceConnectionID)
	}

	// Add schema filter if specified
	if req.SchemaHash != "" {
		query += " AND schema_hash = ?"
		args = append(args, req.SchemaHash)
	}

	query += " ORDER BY min_timestamp"

	rows, err := w.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query data sources: %w", err)
	}
	defer rows.Close()

	var dataSources []DataSource
	var totalRecords int64
	var totalSize int64

	for rows.Next() {
		var ds DataSource
		err := rows.Scan(
			&ds.FilePath,
			&ds.SchemaHash,
			&ds.RecordCount,
			&ds.MinTimestamp,
			&ds.MaxTimestamp,
			&ds.FileSize,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		dataSources = append(dataSources, ds)
		totalRecords += ds.RecordCount
		totalSize += ds.FileSize
	}

	return &DataSourceQueryResult{
		DataSources:        dataSources,
		TotalFiles:         len(dataSources),
		TotalRecords:       totalRecords,
		TotalSize:          totalSize,
		QueryExecutionTime: time.Since(time.Now()), // This should be calculated properly
		OptimizationUsed:   w.determineOptimization(req, dataSources),
	}, nil
}

// GetDirectoryMetadata retrieves metadata for a specific directory
func (w *EnhancedMetadataWriter) GetDirectoryMetadata(ctx context.Context, tenantID, directoryPath string) (*DirectoryMetadata, error) {
	query := `
		SELECT 
			config_pattern,
			total_files,
			total_records,
			total_size_bytes,
			schema_hash,
			min_timestamp,
			max_timestamp,
			last_updated,
			unique_schemas,
			compression_ratio
		FROM directory_structures 
		WHERE tenant_id = ? AND directory_path = ?
		ORDER BY last_updated DESC
		LIMIT 1
	`

	row := w.conn.QueryRow(ctx, query, tenantID, directoryPath)

	var metadata DirectoryMetadata
	err := row.Scan(
		&metadata.ConfigPattern,
		&metadata.TotalFiles,
		&metadata.TotalRecords,
		&metadata.TotalSizeBytes,
		&metadata.SchemaHash,
		&metadata.MinTimestamp,
		&metadata.MaxTimestamp,
		&metadata.LastUpdated,
		&metadata.UniqueSchemas,
		&metadata.CompressionRatio,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get directory metadata: %w", err)
	}

	metadata.TenantID = tenantID
	metadata.DirectoryPath = directoryPath

	return &metadata, nil
}

// Helper methods

func (w *EnhancedMetadataWriter) updateTenantMetadata(metadata *FileMetadata) error {
	ctx := context.Background()

	// Use ReplacingMergeTree for tenant metadata to handle updates
	query := `
		INSERT INTO tenant_metadata (
			tenant_id, total_files, total_rows, total_size_gb, source_count,
			oldest_record, newest_record, last_updated, settings
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// This is a simplified version - in reality, you'd want to aggregate existing data
	totalSizeGB := float64(metadata.FileSize) / (1024 * 1024 * 1024)

	err := w.conn.Exec(ctx, query,
		metadata.TenantID,
		1, // This should be aggregated
		metadata.RecordCount,
		totalSizeGB,
		1, // This should be counted
		metadata.MinTimestamp,
		metadata.MaxTimestamp,
		time.Now(),
		"{}",
	)

	return err
}

func (w *EnhancedMetadataWriter) writeSchemaVersion(metadata *FileMetadata) error {
	ctx := context.Background()

	// Check if schema version already exists
	exists := false
	checkQuery := `
		SELECT 1 FROM schema_versions 
		WHERE tenant_id = ? AND source_id = ? AND hash = ?
		LIMIT 1
	`

	row := w.conn.QueryRow(ctx, checkQuery, metadata.TenantID, metadata.SourceID, metadata.SchemaHash)
	row.Scan(&exists)

	if exists {
		// Update last_seen
		updateQuery := `
			INSERT INTO schema_versions (
				hash, tenant_id, source_id, version, schema, first_seen,
				last_seen, backward_compatible, forward_compatible, record_count
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		schemaJSON, _ := json.Marshal(metadata.Schema)

		err := w.conn.Exec(ctx, updateQuery,
			metadata.SchemaHash,
			metadata.TenantID,
			metadata.SourceID,
			1, // version
			string(schemaJSON),
			time.Now(), // This should be the original first_seen
			time.Now(), // last_seen
			1, // backward_compatible
			1, // forward_compatible
			metadata.RecordCount,
		)

		return err
	}

	// Insert new schema version
	insertQuery := `
		INSERT INTO schema_versions (
			hash, tenant_id, source_id, version, schema, first_seen,
			last_seen, backward_compatible, forward_compatible, record_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	schemaJSON, _ := json.Marshal(metadata.Schema)

	err := w.conn.Exec(ctx, insertQuery,
		metadata.SchemaHash,
		metadata.TenantID,
		metadata.SourceID,
		1, // version
		string(schemaJSON),
		time.Now(),
		time.Now(),
		1, // backward_compatible
		1, // forward_compatible
		metadata.RecordCount,
	)

	return err
}

func (w *EnhancedMetadataWriter) calculateCompressionRatio(metadata *FileMetadata) float64 {
	// Estimate original JSON size (this is a rough approximation)
	estimatedJSONSize := metadata.RecordCount * 1000 // assume 1KB per record on average
	if estimatedJSONSize == 0 {
		return 0.0
	}
	return float64(metadata.FileSize) / float64(estimatedJSONSize)
}

func (w *EnhancedMetadataWriter) determineOptimization(req *DataSourceQuery, dataSources []DataSource) string {
	if len(dataSources) == 0 {
		return "none"
	}

	if len(dataSources) == 1 {
		return "single_file"
	}

	// Check if all files have the same schema
	firstSchema := dataSources[0].SchemaHash
	sameSchema := true
	for _, ds := range dataSources {
		if ds.SchemaHash != firstSchema {
			sameSchema = false
			break
		}
	}

	if sameSchema {
		return "schema_aligned"
	}

	return "multi_schema"
}

// Data structures for enhanced metadata

type DataSourceQuery struct {
	TenantID           string    `json:"tenant_id"`
	SourceConnectionID string    `json:"source_connection_id,omitempty"`
	StartTime          time.Time `json:"start_time"`
	EndTime            time.Time `json:"end_time"`
	SchemaHash         string    `json:"schema_hash,omitempty"`
	DirectoryPattern   string    `json:"directory_pattern,omitempty"`
	MaxFiles           int       `json:"max_files,omitempty"`
}

type DataSourceQueryResult struct {
	DataSources        []DataSource  `json:"data_sources"`
	TotalFiles         int           `json:"total_files"`
	TotalRecords       int64         `json:"total_records"`
	TotalSize          int64         `json:"total_size"`
	QueryExecutionTime time.Duration `json:"query_execution_time"`
	OptimizationUsed   string        `json:"optimization_used"`
	RecommendedQuery   string        `json:"recommended_query,omitempty"`
}

type DataSource struct {
	FilePath     string    `json:"file_path"`
	SchemaHash   string    `json:"schema_hash"`
	RecordCount  int64     `json:"record_count"`
	MinTimestamp time.Time `json:"min_timestamp"`
	MaxTimestamp time.Time `json:"max_timestamp"`
	FileSize     int64     `json:"file_size"`
}

type DirectoryMetadata struct {
	TenantID         string    `json:"tenant_id"`
	DirectoryPath    string    `json:"directory_path"`
	ConfigPattern    string    `json:"config_pattern"`
	TotalFiles       int64     `json:"total_files"`
	TotalRecords     int64     `json:"total_records"`
	TotalSizeBytes   int64     `json:"total_size_bytes"`
	SchemaHash       string    `json:"schema_hash"`
	MinTimestamp     time.Time `json:"min_timestamp"`
	MaxTimestamp     time.Time `json:"max_timestamp"`
	LastUpdated      time.Time `json:"last_updated"`
	UniqueSchemas    int32     `json:"unique_schemas"`
	CompressionRatio float64   `json:"compression_ratio"`
}
