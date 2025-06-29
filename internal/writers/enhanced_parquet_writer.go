// enhanced_parquet_writer.go - Enhanced Parquet writer with metadata generation
package writers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/utils"
)

// EnhancedParquetWriter writes Parquet files with metadata according to directory configuration
type EnhancedParquetWriter struct {
	baseDir            string
	directoryConfigSvc *models.DirectoryConfigService
	metadataWriter     *MetadataWriter
}

// NewEnhancedParquetWriter creates a new enhanced Parquet writer
func NewEnhancedParquetWriter(baseDir string, directoryConfigSvc *models.DirectoryConfigService, metadataWriter *MetadataWriter) *EnhancedParquetWriter {
	return &EnhancedParquetWriter{
		baseDir:            baseDir,
		directoryConfigSvc: directoryConfigSvc,
		metadataWriter:     metadataWriter,
	}
}

// WriteData writes data to Parquet files using the source connection's directory configuration
func (w *EnhancedParquetWriter) WriteData(sourceConnection *models.SourceConnection, records []map[string]interface{}) (*WriteResult, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records to write")
	}

	// Get directory configuration for this source connection
	dirConfig, err := w.directoryConfigSvc.GetConfigBySourceConnection(sourceConnection.ID)
	if err != nil {
		// Use default configuration if none is specified
		dirConfig = w.getDefaultDirectoryConfig(sourceConnection)
	}

	// Process records in batches
	batchSize := sourceConnection.SyncConfig.BatchSize
	if batchSize <= 0 {
		batchSize = 1000 // default batch size
	}

	var allResults []*WriteResult
	
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		
		batch := records[i:end]
		result, err := w.writeBatch(sourceConnection, dirConfig, batch, i/batchSize)
		if err != nil {
			return nil, fmt.Errorf("failed to write batch %d: %w", i/batchSize, err)
		}
		
		allResults = append(allResults, result)
	}

	// Combine all results
	return w.combineResults(allResults), nil
}

// writeBatch writes a single batch of records
func (w *EnhancedParquetWriter) writeBatch(sourceConnection *models.SourceConnection, dirConfig *models.DirectoryConfig, records []map[string]interface{}, sequence int) (*WriteResult, error) {
	timestamp := time.Now()
	
	// Use the first record's timestamp if available
	if len(records) > 0 && dirConfig.Config.SchemaConfig.TimestampField != "" {
		if tsField, ok := records[0][dirConfig.Config.SchemaConfig.TimestampField]; ok {
			if ts, ok := tsField.(time.Time); ok {
				timestamp = ts
			} else if tsStr, ok := tsField.(string); ok {
				if parsed, err := time.Parse(dirConfig.Config.SchemaConfig.TimestampFormat, tsStr); err == nil {
					timestamp = parsed
				}
			}
		}
	}

	// Resolve directory path
	resolver := &models.DirectoryResolver{
		Config:    dirConfig,
		Timestamp: timestamp,
		Data:      records[0], // Use first record for variable resolution
	}

	dirPath, err := resolver.ResolvePath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory path: %w", err)
	}

	fullDirPath := filepath.Join(w.baseDir, dirPath)
	
	// Ensure directory exists
	if err := os.MkdirAll(fullDirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", fullDirPath, err)
	}

	// Resolve filename
	filename := resolver.ResolveFileName(sequence)
	filePath := filepath.Join(fullDirPath, filename)

	// Detect and compute schema hash
	schema := utils.FlattenJSONSchema(records[0])
	schemaHash := utils.ComputeSchemaHash(schema)

	// Write Parquet file
	parquetWriter := NewParquetWriter(filePath)
	defer parquetWriter.Close()

	for _, record := range records {
		if err := parquetWriter.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write record: %w", err)
		}
	}

	if err := parquetWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close parquet writer: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Generate metadata files
	metadata := &FileMetadata{
		FilePath:     filePath,
		RelativePath: filepath.Join(dirPath, filename),
		TenantID:     sourceConnection.TenantID,
		SourceID:     sourceConnection.ID,
		SchemaHash:   schemaHash,
		RecordCount:  int64(len(records)),
		FileSize:     fileInfo.Size(),
		MinTimestamp: timestamp,
		MaxTimestamp: timestamp,
		CreatedAt:    time.Now(),
		Schema:       schema,
		Records:      records,
	}

	if err := w.generateMetadataFiles(fullDirPath, dirConfig.MetadataConfig, metadata); err != nil {
		return nil, fmt.Errorf("failed to generate metadata: %w", err)
	}

	// Update ClickHouse metadata
	if err := w.metadataWriter.WriteParquetFileMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to write ClickHouse metadata: %w", err)
	}

	return &WriteResult{
		FilePath:         filePath,
		RelativePath:     metadata.RelativePath,
		RecordsWritten:   int64(len(records)),
		FileSize:         fileInfo.Size(),
		SchemaHash:       schemaHash,
		DirectoryPath:    fullDirPath,
		MetadataGenerated: true,
	}, nil
}

// generateMetadataFiles generates various metadata files
func (w *EnhancedParquetWriter) generateMetadataFiles(dirPath string, config models.MetadataConfig, metadata *FileMetadata) error {
	// Generate summary file
	if config.GenerateSummary {
		if err := w.generateSummaryFile(dirPath, metadata); err != nil {
			return fmt.Errorf("failed to generate summary: %w", err)
		}
	}

	// Generate schema file
	if config.GenerateSchema {
		if err := w.generateSchemaFile(dirPath, metadata); err != nil {
			return fmt.Errorf("failed to generate schema: %w", err)
		}
	}

	// Generate stats file
	if config.GenerateStats {
		if err := w.generateStatsFile(dirPath, metadata, config.StatsFields); err != nil {
			return fmt.Errorf("failed to generate stats: %w", err)
		}
	}

	// Generate indexes file
	if config.GenerateIndexes {
		if err := w.generateIndexesFile(dirPath, metadata, config.IndexedFields); err != nil {
			return fmt.Errorf("failed to generate indexes: %w", err)
		}
	}

	// Generate custom metadata
	if len(config.CustomMetadata) > 0 {
		if err := w.generateCustomMetadata(dirPath, metadata, config.CustomMetadata); err != nil {
			return fmt.Errorf("failed to generate custom metadata: %w", err)
		}
	}

	return nil
}

// generateSummaryFile generates a summary of the data in the directory
func (w *EnhancedParquetWriter) generateSummaryFile(dirPath string, metadata *FileMetadata) error {
	summary := DirectorySummary{
		DirectoryPath:    dirPath,
		TotalFiles:       1, // This should be updated to count all files in directory
		TotalRecords:     metadata.RecordCount,
		TotalSizeBytes:   metadata.FileSize,
		SchemaHash:       metadata.SchemaHash,
		MinTimestamp:     metadata.MinTimestamp,
		MaxTimestamp:     metadata.MaxTimestamp,
		LastUpdated:      time.Now(),
		FileList:         []string{filepath.Base(metadata.FilePath)},
		UniqueSchemas:    1,
		CompressionRatio: w.calculateCompressionRatio(metadata),
	}

	summaryBytes, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	summaryPath := filepath.Join(dirPath, "_summary.json")
	return os.WriteFile(summaryPath, summaryBytes, 0644)
}

// generateSchemaFile generates a schema definition file
func (w *EnhancedParquetWriter) generateSchemaFile(dirPath string, metadata *FileMetadata) error {
	schema := SchemaInfo{
		SchemaHash:    metadata.SchemaHash,
		Fields:        metadata.Schema,
		RecordCount:   metadata.RecordCount,
		SampleRecord:  metadata.Records[0], // First record as sample
		Version:       1,
		CreatedAt:     time.Now(),
		LastSeen:      time.Now(),
		Compatibility: SchemaCompatibility{
			BackwardCompatible: true,
			ForwardCompatible:  true,
		},
	}

	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	schemaPath := filepath.Join(dirPath, "_schema.json")
	return os.WriteFile(schemaPath, schemaBytes, 0644)
}

// generateStatsFile generates statistical information about the data
func (w *EnhancedParquetWriter) generateStatsFile(dirPath string, metadata *FileMetadata, statsFields []string) error {
	stats := DataStatistics{
		RecordCount:   metadata.RecordCount,
		FieldStats:    make(map[string]FieldStatistics),
		GeneratedAt:   time.Now(),
		SchemaHash:    metadata.SchemaHash,
	}

	// Calculate field statistics
	for _, record := range metadata.Records {
		for fieldName, value := range record {
			// Only calculate stats for specified fields if configured
			if len(statsFields) > 0 && !contains(statsFields, fieldName) {
				continue
			}

			if _, exists := stats.FieldStats[fieldName]; !exists {
				stats.FieldStats[fieldName] = FieldStatistics{
					FieldName:  fieldName,
					DataType:   getDataType(value),
					NonNullCount: 0,
					NullCount:    0,
				}
			}

			fieldStat := stats.FieldStats[fieldName]
			if value == nil {
				fieldStat.NullCount++
			} else {
				fieldStat.NonNullCount++
				w.updateFieldStatistics(&fieldStat, value)
			}
			stats.FieldStats[fieldName] = fieldStat
		}
	}

	statsBytes, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return err
	}

	statsPath := filepath.Join(dirPath, "_stats.json")
	return os.WriteFile(statsPath, statsBytes, 0644)
}

// generateIndexesFile generates index information for fast lookups
func (w *EnhancedParquetWriter) generateIndexesFile(dirPath string, metadata *FileMetadata, indexedFields []string) error {
	indexes := DirectoryIndexes{
		SchemaHash:    metadata.SchemaHash,
		TotalRecords:  metadata.RecordCount,
		GeneratedAt:   time.Now(),
		FieldIndexes:  make(map[string]FieldIndex),
	}

	// Create indexes for specified fields
	for _, fieldName := range indexedFields {
		if fieldName == "" {
			continue
		}

		fieldIndex := FieldIndex{
			FieldName:    fieldName,
			DataType:     "",
			UniqueValues: make(map[string][]int), // value -> record indices
			MinValue:     nil,
			MaxValue:     nil,
		}

		// Build index by scanning records
		for i, record := range metadata.Records {
			if value, exists := record[fieldName]; exists && value != nil {
				valueStr := fmt.Sprintf("%v", value)
				if fieldIndex.DataType == "" {
					fieldIndex.DataType = getDataType(value)
				}

				// Add to unique values index
				if _, exists := fieldIndex.UniqueValues[valueStr]; !exists {
					fieldIndex.UniqueValues[valueStr] = []int{}
				}
				fieldIndex.UniqueValues[valueStr] = append(fieldIndex.UniqueValues[valueStr], i)

				// Update min/max for comparable types
				w.updateMinMax(&fieldIndex, value)
			}
		}

		indexes.FieldIndexes[fieldName] = fieldIndex
	}

	indexBytes, err := json.MarshalIndent(indexes, "", "  ")
	if err != nil {
		return err
	}

	indexPath := filepath.Join(dirPath, "_indexes.json")
	return os.WriteFile(indexPath, indexBytes, 0644)
}

// generateCustomMetadata generates custom metadata files
func (w *EnhancedParquetWriter) generateCustomMetadata(dirPath string, metadata *FileMetadata, customMetadata map[string]interface{}) error {
	customPath := filepath.Join(dirPath, "_custom.json")
	
	// Merge custom metadata with basic file info
	combined := map[string]interface{}{
		"file_info": map[string]interface{}{
			"file_path":      metadata.FilePath,
			"record_count":   metadata.RecordCount,
			"file_size":      metadata.FileSize,
			"schema_hash":    metadata.SchemaHash,
			"created_at":     metadata.CreatedAt,
		},
		"custom": customMetadata,
	}

	customBytes, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(customPath, customBytes, 0644)
}

// Helper functions

func (w *EnhancedParquetWriter) getDefaultDirectoryConfig(sourceConnection *models.SourceConnection) *models.DirectoryConfig {
	return &models.DirectoryConfig{
		ID:       "default-" + sourceConnection.ID,
		TenantID: sourceConnection.TenantID,
		Name:     "Default Configuration",
		Pattern:  "{tenant_id}/{source_id}/{year}/{month}/{day}",
		Variables: map[string]models.Variable{
			"tenant_id": {Name: "tenant_id", Type: models.VarTypeStatic, Source: sourceConnection.TenantID, Required: true},
			"source_id": {Name: "source_id", Type: models.VarTypeStatic, Source: sourceConnection.ID, Required: true},
			"year":      {Name: "year", Type: models.VarTypeDateTime, Source: "year", Required: true},
			"month":     {Name: "month", Type: models.VarTypeDateTime, Source: "month", Required: true},
			"day":       {Name: "day", Type: models.VarTypeDateTime, Source: "day", Required: true},
		},
		FileNaming: models.FileNamingConfig{
			Pattern:   "data_{timestamp}_{sequence}",
			Timestamp: "2006-01-02_15-04-05",
			Sequence:  true,
		},
		MetadataConfig: models.MetadataConfig{
			GenerateSummary: true,
			GenerateSchema:  true,
			GenerateStats:   true,
			GenerateIndexes: false,
		},
	}
}

func (w *EnhancedParquetWriter) calculateCompressionRatio(metadata *FileMetadata) float64 {
	// Estimate original JSON size
	jsonBytes, _ := json.Marshal(metadata.Records)
	if len(jsonBytes) == 0 {
		return 0.0
	}
	return float64(metadata.FileSize) / float64(len(jsonBytes))
}

func (w *EnhancedParquetWriter) combineResults(results []*WriteResult) *WriteResult {
	if len(results) == 0 {
		return nil
	}
	
	if len(results) == 1 {
		return results[0]
	}

	// Combine multiple results
	combined := &WriteResult{
		FilePath:         results[0].DirectoryPath,
		RecordsWritten:   0,
		FileSize:         0,
		MetadataGenerated: true,
	}

	for _, result := range results {
		combined.RecordsWritten += result.RecordsWritten
		combined.FileSize += result.FileSize
	}

	return combined
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getDataType(value interface{}) string {
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
		return "datetime"
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	default:
		return "unknown"
	}
}

func (w *EnhancedParquetWriter) updateFieldStatistics(fieldStat *FieldStatistics, value interface{}) {
	// Update statistics based on data type
	switch v := value.(type) {
	case string:
		if fieldStat.MinValue == nil || v < fieldStat.MinValue.(string) {
			fieldStat.MinValue = v
		}
		if fieldStat.MaxValue == nil || v > fieldStat.MaxValue.(string) {
			fieldStat.MaxValue = v
		}
	case int, int32, int64:
		intVal := int64(0)
		switch v := v.(type) {
		case int:
			intVal = int64(v)
		case int32:
			intVal = int64(v)
		case int64:
			intVal = v
		}
		if fieldStat.MinValue == nil || intVal < fieldStat.MinValue.(int64) {
			fieldStat.MinValue = intVal
		}
		if fieldStat.MaxValue == nil || intVal > fieldStat.MaxValue.(int64) {
			fieldStat.MaxValue = intVal
		}
	case float32, float64:
		floatVal := float64(0)
		switch v := v.(type) {
		case float32:
			floatVal = float64(v)
		case float64:
			floatVal = v
		}
		if fieldStat.MinValue == nil || floatVal < fieldStat.MinValue.(float64) {
			fieldStat.MinValue = floatVal
		}
		if fieldStat.MaxValue == nil || floatVal > fieldStat.MaxValue.(float64) {
			fieldStat.MaxValue = floatVal
		}
	}
}

func (w *EnhancedParquetWriter) updateMinMax(fieldIndex *FieldIndex, value interface{}) {
	// Similar to updateFieldStatistics but for indexes
	switch v := value.(type) {
	case string:
		if fieldIndex.MinValue == nil || v < fieldIndex.MinValue.(string) {
			fieldIndex.MinValue = v
		}
		if fieldIndex.MaxValue == nil || v > fieldIndex.MaxValue.(string) {
			fieldIndex.MaxValue = v
		}
	case int, int32, int64:
		intVal := int64(0)
		switch v := v.(type) {
		case int:
			intVal = int64(v)
		case int32:
			intVal = int64(v)
		case int64:
			intVal = v
		}
		if fieldIndex.MinValue == nil || intVal < fieldIndex.MinValue.(int64) {
			fieldIndex.MinValue = intVal
		}
		if fieldIndex.MaxValue == nil || intVal > fieldIndex.MaxValue.(int64) {
			fieldIndex.MaxValue = intVal
		}
	}
}

// Data structures for metadata files

type WriteResult struct {
	FilePath         string `json:"file_path"`
	RelativePath     string `json:"relative_path"`
	RecordsWritten   int64  `json:"records_written"`
	FileSize         int64  `json:"file_size"`
	SchemaHash       string `json:"schema_hash"`
	DirectoryPath    string `json:"directory_path"`
	MetadataGenerated bool  `json:"metadata_generated"`
}

type FileMetadata struct {
	FilePath     string                 `json:"file_path"`
	RelativePath string                 `json:"relative_path"`
	TenantID     string                 `json:"tenant_id"`
	SourceID     string                 `json:"source_id"`
	SchemaHash   string                 `json:"schema_hash"`
	RecordCount  int64                  `json:"record_count"`
	FileSize     int64                  `json:"file_size"`
	MinTimestamp time.Time              `json:"min_timestamp"`
	MaxTimestamp time.Time              `json:"max_timestamp"`
	CreatedAt    time.Time              `json:"created_at"`
	Schema       map[string]string      `json:"schema"`
	Records      []map[string]interface{} `json:"-"` // Don't serialize records
}

type DirectorySummary struct {
	DirectoryPath    string    `json:"directory_path"`
	TotalFiles       int       `json:"total_files"`
	TotalRecords     int64     `json:"total_records"`
	TotalSizeBytes   int64     `json:"total_size_bytes"`
	SchemaHash       string    `json:"schema_hash"`
	MinTimestamp     time.Time `json:"min_timestamp"`
	MaxTimestamp     time.Time `json:"max_timestamp"`
	LastUpdated      time.Time `json:"last_updated"`
	FileList         []string  `json:"file_list"`
	UniqueSchemas    int       `json:"unique_schemas"`
	CompressionRatio float64   `json:"compression_ratio"`
}

type SchemaInfo struct {
	SchemaHash    string                 `json:"schema_hash"`
	Fields        map[string]string      `json:"fields"`
	RecordCount   int64                  `json:"record_count"`
	SampleRecord  map[string]interface{} `json:"sample_record"`
	Version       int                    `json:"version"`
	CreatedAt     time.Time              `json:"created_at"`
	LastSeen      time.Time              `json:"last_seen"`
	Compatibility SchemaCompatibility    `json:"compatibility"`
}

type SchemaCompatibility struct {
	BackwardCompatible bool `json:"backward_compatible"`
	ForwardCompatible  bool `json:"forward_compatible"`
}

type DataStatistics struct {
	RecordCount int64                      `json:"record_count"`
	FieldStats  map[string]FieldStatistics `json:"field_stats"`
	GeneratedAt time.Time                  `json:"generated_at"`
	SchemaHash  string                     `json:"schema_hash"`
}

type FieldStatistics struct {
	FieldName    string      `json:"field_name"`
	DataType     string      `json:"data_type"`
	NonNullCount int64       `json:"non_null_count"`
	NullCount    int64       `json:"null_count"`
	MinValue     interface{} `json:"min_value"`
	MaxValue     interface{} `json:"max_value"`
	UniqueCount  int64       `json:"unique_count,omitempty"`
	AvgLength    float64     `json:"avg_length,omitempty"`
}

type DirectoryIndexes struct {
	SchemaHash   string                `json:"schema_hash"`
	TotalRecords int64                 `json:"total_records"`
	GeneratedAt  time.Time             `json:"generated_at"`
	FieldIndexes map[string]FieldIndex `json:"field_indexes"`
}

type FieldIndex struct {
	FieldName    string              `json:"field_name"`
	DataType     string              `json:"data_type"`
	UniqueValues map[string][]int    `json:"unique_values"` // value -> record indices
	MinValue     interface{}         `json:"min_value"`
	MaxValue     interface{}         `json:"max_value"`
}
