package writers

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
	"github.com/your-org/storage-control-plane/internal/clickhouse"
	"github.com/your-org/storage-control-plane/pkg/models"
)

// MetadataParquetWriter writes Parquet files AND populates ClickHouse metadata tables
// This solves the problem of querying across multiple Parquet files
type MetadataParquetWriter struct {
	clickhouseClient *clickhouse.Client
	config           *MetadataWriterConfig
}

type MetadataWriterConfig struct {
	BaseDir                 string            `json:"base_dir"`
	PopulateRecordMetadata  bool              `json:"populate_record_metadata"`  // Whether to create searchable record index
	IndexedFields           []string          `json:"indexed_fields"`            // Which fields to index for searching
	CustomFieldExtractors   map[string]string `json:"custom_field_extractors"`   // JSONPath extractors for custom fields
	BloomFilterFields       []string          `json:"bloom_filter_fields"`       // Fields to create bloom filters for
	MaxRecordsPerFile       int               `json:"max_records_per_file"`
	CompressionLevel        string            `json:"compression_level"`
}

type FileMetadataCollector struct {
	FileID        string
	TenantID      string
	SourceID      string
	FilePath      string
	DirectoryPath string
	
	RecordCount   int64
	IndexedFields map[string]*FieldIndexCollector
	Statistics    *FileStatisticsCollector
	
	MinTimestamp  *time.Time
	MaxTimestamp  *time.Time
	MinCreatedAt  *time.Time
	MaxCreatedAt  *time.Time
}

type FieldIndexCollector struct {
	FieldName    string
	FieldType    string
	UniqueValues map[string]bool
	MinValue     interface{}
	MaxValue     interface{}
	HasNulls     bool
	ValueCounts  map[string]int64
}

type FileStatisticsCollector struct {
	DistinctCounts map[string]int64
	NullCounts     map[string]int64
	ValueCounts    map[string]map[string]int64
	NumericStats   map[string]*NumericStatCollector
}

type NumericStatCollector struct {
	Min    float64
	Max    float64
	Sum    float64
	Count  int64
	SumSq  float64 // For standard deviation
}

func NewMetadataParquetWriter(clickhouseClient *clickhouse.Client, config *MetadataWriterConfig) *MetadataParquetWriter {
	if config == nil {
		config = &MetadataWriterConfig{
			PopulateRecordMetadata: true,
			IndexedFields:          []string{"name", "email", "status", "category"},
			MaxRecordsPerFile:      100000,
			CompressionLevel:       "SNAPPY",
		}
	}
	
	return &MetadataParquetWriter{
		clickhouseClient: clickhouseClient,
		config:           config,
	}
}

// WriteParquetWithMetadata writes Parquet file AND populates ClickHouse metadata
// This is the key function that solves your cross-file query problem!
func (w *MetadataParquetWriter) WriteParquetWithMetadata(
	ctx context.Context,
	records []map[string]interface{},
	tenantID, sourceID string,
	directoryConfig *models.DirectoryConfig,
) (*models.ParquetFileMetadata, error) {

	if len(records) == 0 {
		return nil, fmt.Errorf("no records to write")
	}

	// 1. Generate file path using directory configuration
	dirPath, err := w.generateDirectoryPath(records[0], directoryConfig, tenantID, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate directory path: %w", err)
	}

	fileID := uuid.New().String()
	fileName := fmt.Sprintf("data_%d_%s.parquet", time.Now().Unix(), fileID[:8])
	filePath := filepath.Join(w.config.BaseDir, dirPath, fileName)

	// 2. Initialize metadata collector
	metadata := &FileMetadataCollector{
		FileID:        fileID,
		TenantID:      tenantID,
		SourceID:      sourceID,
		FilePath:      filePath,
		DirectoryPath: dirPath,
		IndexedFields: make(map[string]*FieldIndexCollector),
		Statistics:    &FileStatisticsCollector{
			DistinctCounts: make(map[string]int64),
			NullCounts:     make(map[string]int64),
			ValueCounts:    make(map[string]map[string]int64),
			NumericStats:   make(map[string]*NumericStatCollector),
		},
	}

	// 3. Write Parquet file while collecting metadata
	err = w.writeParquetFile(filePath, records, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to write parquet file: %w", err)
	}

	// 4. Store file metadata in ClickHouse
	fileMetadata, err := w.storeFileMetadata(ctx, metadata)
	if err != nil {
		log.Printf("Warning: Failed to store file metadata: %v", err)
		// Don't fail the entire operation, but log the error
	}

	// 5. Store record metadata in ClickHouse (for cross-file searching)
	if w.config.PopulateRecordMetadata {
		err = w.storeRecordMetadata(ctx, records, metadata)
		if err != nil {
			log.Printf("Warning: Failed to store record metadata: %v", err)
		}
	}

	// 6. Update directory summary
	err = w.updateDirectorySummary(ctx, metadata, records)
	if err != nil {
		log.Printf("Warning: Failed to update directory summary: %v", err)
	}

	return fileMetadata, nil
}

func (w *MetadataParquetWriter) writeParquetFile(filePath string, records []map[string]interface{}, metadata *FileMetadataCollector) error {
	// Ensure directory exists
	err := ensureDirectoryExists(filepath.Dir(filePath))
	if err != nil {
		return err
	}

	// Create Parquet writer
	fw, err := writer.NewLocalFileWriter(filePath)
	if err != nil {
		return err
	}
	defer fw.Close()

	pw, err := writer.NewParquetWriter(fw, nil, 4)
	if err != nil {
		return err
	}
	defer pw.WriteStop()

	// Configure compression
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	// Write records and collect metadata
	for i, record := range records {
		// Collect metadata for this record
		w.collectRecordMetadata(record, metadata, int64(i))

		// Convert to Parquet-compatible format
		parquetRecord := w.convertToParquetRecord(record)
		
		// Write to Parquet file
		err = pw.Write(parquetRecord)
		if err != nil {
			return fmt.Errorf("failed to write record %d: %w", i, err)
		}
	}

	metadata.RecordCount = int64(len(records))
	return nil
}

func (w *MetadataParquetWriter) collectRecordMetadata(record map[string]interface{}, metadata *FileMetadataCollector, rowIndex int64) {
	for fieldName, value := range record {
		// Initialize field collector if needed
		if _, exists := metadata.IndexedFields[fieldName]; !exists {
			metadata.IndexedFields[fieldName] = &FieldIndexCollector{
				FieldName:    fieldName,
				FieldType:    w.inferFieldType(value),
				UniqueValues: make(map[string]bool),
				ValueCounts:  make(map[string]int64),
			}
		}

		collector := metadata.IndexedFields[fieldName]

		// Collect field statistics
		if value == nil {
			collector.HasNulls = true
			metadata.Statistics.NullCounts[fieldName]++
		} else {
			valueStr := fmt.Sprintf("%v", value)
			collector.UniqueValues[valueStr] = true
			collector.ValueCounts[valueStr]++

			// Update min/max for comparable types
			w.updateMinMax(collector, value)

			// Collect numeric statistics
			if w.isNumeric(value) {
				w.updateNumericStats(metadata.Statistics, fieldName, value)
			}
		}

		// Collect timestamp boundaries
		if fieldName == "timestamp" || fieldName == "created_at" || fieldName == "updated_at" {
			if ts, ok := w.parseTimestamp(value); ok {
				if fieldName == "timestamp" {
					w.updateTimestampBounds(&metadata.MinTimestamp, &metadata.MaxTimestamp, ts)
				} else if fieldName == "created_at" {
					w.updateTimestampBounds(&metadata.MinCreatedAt, &metadata.MaxCreatedAt, ts)
				}
			}
		}
	}
}

// storeRecordMetadata stores searchable record metadata in ClickHouse
// THIS IS THE KEY FUNCTION that enables cross-file queries!
func (w *MetadataParquetWriter) storeRecordMetadata(ctx context.Context, records []map[string]interface{}, fileMetadata *FileMetadataCollector) error {
	if w.clickhouseClient == nil {
		return fmt.Errorf("ClickHouse client not available")
	}

	recordMetadataList := make([]models.RecordMetadata, 0, len(records))

	for i, record := range records {
		recordMetadata := models.RecordMetadata{
			RecordID:     fmt.Sprintf("%s_%d", fileMetadata.FileID, i),
			TenantID:     fileMetadata.TenantID,
			SourceID:     fileMetadata.SourceID,
			FileID:       fileMetadata.FileID,
			FilePath:     fileMetadata.FilePath,
			RowNumber:    int64(i),
			CustomFields: make(map[string]any),
		}

		// Extract standard searchable fields
		if name, ok := record["name"].(string); ok {
			recordMetadata.Name = name
		}
		if email, ok := record["email"].(string); ok {
			recordMetadata.Email = email
		}
		if status, ok := record["status"].(string); ok {
			recordMetadata.Status = status
		}
		if category, ok := record["category"].(string); ok {
			recordMetadata.Category = category
		}

		// Extract timestamps
		if createdAt, ok := w.parseTimestamp(record["created_at"]); ok {
			recordMetadata.CreatedAt = createdAt
		}
		if updatedAt, ok := w.parseTimestamp(record["updated_at"]); ok {
			recordMetadata.UpdatedAt = updatedAt
		}
		if timestamp, ok := w.parseTimestamp(record["timestamp"]); ok {
			recordMetadata.Timestamp = timestamp
		}

		// Extract custom indexed fields
		for _, fieldName := range w.config.IndexedFields {
			if value, exists := record[fieldName]; exists && value != nil {
				recordMetadata.CustomFields[fieldName] = value
			}
		}

		recordMetadataList = append(recordMetadataList, recordMetadata)
	}

	// Batch insert into ClickHouse
	return w.insertRecordMetadataBatch(ctx, recordMetadataList)
}

func (w *MetadataParquetWriter) insertRecordMetadataBatch(ctx context.Context, records []models.RecordMetadata) error {
	// Build INSERT query
	query := `
		INSERT INTO record_metadata (
			record_id, tenant_id, source_id, file_id, file_path,
			name, email, status, category, tags,
			created_at, updated_at, timestamp,
			custom_fields, row_number
		) VALUES
	`

	values := make([]string, 0, len(records))
	for _, record := range records {
		customFieldsJSON, _ := json.Marshal(record.CustomFields)
		tagsJSON, _ := json.Marshal(record.Tags)
		
		value := fmt.Sprintf(
			"(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, '%s', '%s', '%s', %s, %d)",
			w.quoteString(record.RecordID),
			w.quoteString(record.TenantID),
			w.quoteString(record.SourceID),
			w.quoteString(record.FileID),
			w.quoteString(record.FilePath),
			w.quoteString(record.Name),
			w.quoteString(record.Email),
			w.quoteString(record.Status),
			w.quoteString(record.Category),
			string(tagsJSON),
			record.CreatedAt.Format("2006-01-02 15:04:05"),
			record.UpdatedAt.Format("2006-01-02 15:04:05"),
			record.Timestamp.Format("2006-01-02 15:04:05"),
			string(customFieldsJSON),
			record.RowNumber,
		)
		values = append(values, value)
	}

	fullQuery := query + strings.Join(values, ",")
	
	// Execute the query
	return w.clickhouseClient.ExecuteRawQuery(ctx, fullQuery)
}

// Helper methods
func (w *MetadataParquetWriter) generateDirectoryPath(record map[string]interface{}, config *models.DirectoryConfig, tenantID, sourceID string) (string, error) {
	if config == nil {
		// Default directory structure
		now := time.Now()
		return filepath.Join(tenantID, sourceID, now.Format("2006/01/02")), nil
	}

	// Apply user-defined directory pattern
	// This would implement the template engine for directory structure
	// For now, using a simple implementation
	basePath := filepath.Join(tenantID, sourceID)
	
	if config.TimePartitioning {
		now := time.Now()
		switch config.PartitionFormat {
		case "daily":
			basePath = filepath.Join(basePath, now.Format("2006/01/02"))
		case "monthly":
			basePath = filepath.Join(basePath, now.Format("2006/01"))
		case "yearly":
			basePath = filepath.Join(basePath, now.Format("2006"))
		}
	}

	return basePath, nil
}

func (w *MetadataParquetWriter) quoteString(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}

func (w *MetadataParquetWriter) parseTimestamp(value interface{}) (time.Time, bool) {
	switch v := value.(type) {
	case time.Time:
		return v, true
	case string:
		// Try common timestamp formats
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

// Additional helper methods would go here...
func (w *MetadataParquetWriter) inferFieldType(value interface{}) string {
	if value == nil {
		return "null"
	}
	
	switch value.(type) {
	case string:
		return "string"
	case int, int32, int64:
		return "integer"
	case float32, float64:
		return "float"
	case bool:
		return "boolean"
	case time.Time:
		return "timestamp"
	default:
		return "unknown"
	}
}

func (w *MetadataParquetWriter) isNumeric(value interface{}) bool {
	switch value.(type) {
	case int, int32, int64, float32, float64:
		return true
	default:
		return false
	}
}

func (w *MetadataParquetWriter) updateTimestampBounds(min, max **time.Time, current time.Time) {
	if *min == nil || current.Before(**min) {
		*min = &current
	}
	if *max == nil || current.After(**max) {
		*max = &current
	}
}

// More helper methods for metadata collection, numeric stats, etc...
func (w *MetadataParquetWriter) updateMinMax(collector *FieldIndexCollector, value interface{}) {
	// Implementation for updating min/max values
}

func (w *MetadataParquetWriter) updateNumericStats(stats *FileStatisticsCollector, fieldName string, value interface{}) {
	// Implementation for updating numeric statistics
}

func (w *MetadataParquetWriter) convertToParquetRecord(record map[string]interface{}) interface{} {
	// Convert record to Parquet-compatible format
	return record
}

func (w *MetadataParquetWriter) storeFileMetadata(ctx context.Context, metadata *FileMetadataCollector) (*models.ParquetFileMetadata, error) {
	// Implementation for storing file metadata
	return nil, nil
}

func (w *MetadataParquetWriter) updateDirectorySummary(ctx context.Context, metadata *FileMetadataCollector, records []map[string]interface{}) error {
	// Implementation for updating directory summaries
	return nil
}

func ensureDirectoryExists(dir string) error {
	// Implementation for creating directories
	return nil
}
