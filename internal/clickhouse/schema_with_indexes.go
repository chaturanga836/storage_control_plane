// schema_with_indexes.go - Database schema definitions with optimized indexes
package clickhouse

import (
	"context"
	"fmt"
	"strings"
)

// TableSchema defines a complete table with indexes
type TableSchema struct {
	Name        string              `json:"name"`
	Columns     []ColumnDefinition  `json:"columns"`
	Engine      string              `json:"engine"`
	OrderBy     []string            `json:"order_by"`
	PartitionBy string              `json:"partition_by,omitempty"`
	Indexes     []IndexDefinition   `json:"indexes"`
	Settings    map[string]string   `json:"settings,omitempty"`
}

type ColumnDefinition struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

// PredefinedSchemas contains all table schemas with their recommended indexes
var PredefinedSchemas = map[string]TableSchema{
	"business_data": {
		Name: "business_data",
		Columns: []ColumnDefinition{
			{Name: "tenant_id", Type: "String", Comment: "Tenant identifier"},
			{Name: "data_id", Type: "String", Comment: "Unique data identifier"},
			{Name: "payload", Type: "String", Comment: "JSON payload data"},
			{Name: "created_at", Type: "DateTime", Comment: "Record creation timestamp"},
			{Name: "updated_at", Type: "DateTime", Default: "now()", Comment: "Last update timestamp"},
			{Name: "version", Type: "UInt64", Default: "1", Comment: "Record version for optimistic locking"},
			{Name: "size_bytes", Type: "UInt64", Default: "0", Comment: "Payload size in bytes"},
		},
		Engine:      "MergeTree()",
		OrderBy:     []string{"tenant_id", "created_at"},
		PartitionBy: "toYYYYMM(created_at)",
		Indexes: []IndexDefinition{
			{
				Name:        "idx_tenant_id",
				Table:       "business_data",
				Columns:     []string{"tenant_id"},
				Type:        IndexTypeBloomFilter,
				Granularity: 1,
			},
			{
				Name:        "idx_created_at",
				Table:       "business_data",
				Columns:     []string{"created_at"},
				Type:        IndexTypeMinMax,
				Granularity: 1,
			},
			{
				Name:        "idx_data_id",
				Table:       "business_data",
				Columns:     []string{"data_id"},
				Type:        IndexTypeBloomFilter,
				Granularity: 1,
			},
			{
				Name:        "idx_updated_at",
				Table:       "business_data",
				Columns:     []string{"updated_at"},
				Type:        IndexTypeMinMax,
				Granularity: 1,
			},
		},
		Settings: map[string]string{
			"index_granularity": "8192",
			"storage_policy":    "default",
		},
	},
	
	"analytics_events": {
		Name: "analytics_events",
		Columns: []ColumnDefinition{
			{Name: "tenant_id", Type: "String", Comment: "Tenant identifier"},
			{Name: "event_id", Type: "String", Comment: "Unique event identifier"},
			{Name: "timestamp", Type: "DateTime", Comment: "Event timestamp"},
			{Name: "event_type", Type: "String", Comment: "Type of event"},
			{Name: "value", Type: "Float64", Default: "0.0", Comment: "Numeric value associated with event"},
			{Name: "properties", Type: "String", Comment: "JSON properties"},
			{Name: "source_id", Type: "String", Comment: "Source system identifier"},
			{Name: "session_id", Type: "String", Nullable: true, Comment: "User session identifier"},
		},
		Engine:      "MergeTree()",
		OrderBy:     []string{"tenant_id", "timestamp"},
		PartitionBy: "(tenant_id, toYYYYMM(timestamp))",
		Indexes: []IndexDefinition{
			{
				Name:        "idx_timestamp",
				Table:       "analytics_events",
				Columns:     []string{"timestamp"},
				Type:        IndexTypeMinMax,
				Granularity: 1,
			},
			{
				Name:        "idx_tenant_id",
				Table:       "analytics_events",
				Columns:     []string{"tenant_id"},
				Type:        IndexTypeBloomFilter,
				Granularity: 1,
			},
			{
				Name:        "idx_event_type",
				Table:       "analytics_events",
				Columns:     []string{"event_type"},
				Type:        IndexTypeSet,
				Granularity: 1,
			},
			{
				Name:        "idx_source_id",
				Table:       "analytics_events",
				Columns:     []string{"source_id"},
				Type:        IndexTypeBloomFilter,
				Granularity: 1,
			},
			{
				Name:        "idx_value",
				Table:       "analytics_events",
				Columns:     []string{"value"},
				Type:        IndexTypeMinMax,
				Granularity: 1,
			},
		},
		Settings: map[string]string{
			"index_granularity": "8192",
			"storage_policy":    "hot_and_cold",
		},
	},
	
	"parquet_files": {
		Name: "parquet_files",
		Columns: []ColumnDefinition{
			{Name: "id", Type: "String", Comment: "Unique file identifier"},
			{Name: "tenant_id", Type: "String", Comment: "Tenant identifier"},
			{Name: "source_id", Type: "String", Comment: "Source system identifier"},
			{Name: "file_path", Type: "String", Comment: "File path in storage"},
			{Name: "schema_hash", Type: "String", Comment: "Schema version hash"},
			{Name: "record_count", Type: "UInt64", Comment: "Number of records in file"},
			{Name: "file_size", Type: "UInt64", Comment: "File size in bytes"},
			{Name: "min_timestamp", Type: "DateTime", Comment: "Earliest record timestamp"},
			{Name: "max_timestamp", Type: "DateTime", Comment: "Latest record timestamp"},
			{Name: "created_at", Type: "DateTime", Comment: "File creation timestamp"},
			{Name: "compressed", Type: "UInt8", Default: "1", Comment: "Whether file is compressed"},
			{Name: "compression_ratio", Type: "Float32", Default: "0.0", Comment: "Compression ratio achieved"},
		},
		Engine:      "MergeTree()",
		OrderBy:     []string{"tenant_id", "created_at"},
		PartitionBy: "toYYYYMM(created_at)",
		Indexes: []IndexDefinition{
			{
				Name:        "idx_tenant_id",
				Table:       "parquet_files",
				Columns:     []string{"tenant_id"},
				Type:        IndexTypeBloomFilter,
				Granularity: 1,
			},
			{
				Name:        "idx_created_at",
				Table:       "parquet_files",
				Columns:     []string{"created_at"},
				Type:        IndexTypeMinMax,
				Granularity: 1,
			},
			{
				Name:        "idx_min_timestamp",
				Table:       "parquet_files",
				Columns:     []string{"min_timestamp"},
				Type:        IndexTypeMinMax,
				Granularity: 1,
			},
			{
				Name:        "idx_max_timestamp",
				Table:       "parquet_files",
				Columns:     []string{"max_timestamp"},
				Type:        IndexTypeMinMax,
				Granularity: 1,
			},
			{
				Name:        "idx_source_id",
				Table:       "parquet_files",
				Columns:     []string{"source_id"},
				Type:        IndexTypeBloomFilter,
				Granularity: 1,
			},
			{
				Name:        "idx_file_path",
				Table:       "parquet_files",
				Columns:     []string{"file_path"},
				Type:        IndexTypeBloomFilter,
				Granularity: 1,
			},
		},
		Settings: map[string]string{
			"index_granularity": "8192",
			"storage_policy":    "default",
		},
	},
	
	"tenant_metadata": {
		Name: "tenant_metadata",
		Columns: []ColumnDefinition{
			{Name: "tenant_id", Type: "String", Comment: "Tenant identifier"},
			{Name: "total_files", Type: "UInt64", Default: "0", Comment: "Total number of files"},
			{Name: "total_rows", Type: "UInt64", Default: "0", Comment: "Total number of rows"},
			{Name: "total_size_gb", Type: "Float64", Default: "0.0", Comment: "Total size in GB"},
			{Name: "source_count", Type: "UInt32", Default: "0", Comment: "Number of sources"},
			{Name: "oldest_record", Type: "DateTime", Nullable: true, Comment: "Oldest record timestamp"},
			{Name: "newest_record", Type: "DateTime", Nullable: true, Comment: "Newest record timestamp"},
			{Name: "last_updated", Type: "DateTime", Default: "now()", Comment: "Last update timestamp"},
			{Name: "settings", Type: "String", Default: "{}", Comment: "Tenant-specific settings JSON"},
		},
		Engine:      "ReplacingMergeTree(last_updated)",
		OrderBy:     []string{"tenant_id"},
		PartitionBy: "",
		Indexes: []IndexDefinition{
			{
				Name:        "idx_tenant_id",
				Table:       "tenant_metadata",
				Columns:     []string{"tenant_id"},
				Type:        IndexTypeBloomFilter,
				Granularity: 1,
			},
			{
				Name:        "idx_last_updated",
				Table:       "tenant_metadata",
				Columns:     []string{"last_updated"},
				Type:        IndexTypeMinMax,
				Granularity: 1,
			},
		},
		Settings: map[string]string{
			"index_granularity": "8192",
		},
	},
}

// SchemaManager handles database schema operations
type SchemaManager struct {
	store        *Store
	indexManager *IndexManager
}

// NewSchemaManager creates a new schema manager
func NewSchemaManager(store *Store) *SchemaManager {
	return &SchemaManager{
		store:        store,
		indexManager: NewIndexManager(store),
	}
}

// CreateTable creates a table with its indexes
func (sm *SchemaManager) CreateTable(ctx context.Context, tableName string) error {
	schema, exists := PredefinedSchemas[tableName]
	if !exists {
		return fmt.Errorf("no schema defined for table %s", tableName)
	}
	
	// Create table
	createTableQuery := sm.generateCreateTableQuery(schema)
	if _, err := sm.store.db.ExecContext(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}
	
	// Create indexes
	for _, indexDef := range schema.Indexes {
		if err := sm.indexManager.CreateIndex(ctx, indexDef); err != nil {
			// Log warning but don't fail table creation
			fmt.Printf("Warning: failed to create index %s: %v\n", indexDef.Name, err)
		}
	}
	
	return nil
}

// CreateAllTables creates all predefined tables
func (sm *SchemaManager) CreateAllTables(ctx context.Context) error {
	for tableName := range PredefinedSchemas {
		if err := sm.CreateTable(ctx, tableName); err != nil {
			return fmt.Errorf("failed to create table %s: %w", tableName, err)
		}
	}
	return nil
}

// DropTable drops a table and all its indexes
func (sm *SchemaManager) DropTable(ctx context.Context, tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := sm.store.db.ExecContext(ctx, query)
	return err
}

// GetTableSchema returns the schema for a table
func (sm *SchemaManager) GetTableSchema(tableName string) (TableSchema, bool) {
	schema, exists := PredefinedSchemas[tableName]
	return schema, exists
}

// ValidateSchema checks if a table matches its expected schema
func (sm *SchemaManager) ValidateSchema(ctx context.Context, tableName string) error {
	expectedSchema, exists := PredefinedSchemas[tableName]
	if !exists {
		return fmt.Errorf("no schema defined for table %s", tableName)
	}
	
	// Check if table exists
	var count int
	query := "SELECT count() FROM system.tables WHERE name = ? AND database = currentDatabase()"
	if err := sm.store.db.QueryRowContext(ctx, query, tableName).Scan(&count); err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}
	
	if count == 0 {
		return fmt.Errorf("table %s does not exist", tableName)
	}
	
	// Check indexes
	existingIndexes, err := sm.indexManager.ListIndexes(ctx, tableName)
	if err != nil {
		return fmt.Errorf("failed to list indexes: %w", err)
	}
	
	expectedIndexNames := make(map[string]bool)
	for _, idx := range expectedSchema.Indexes {
		expectedIndexNames[idx.Name] = true
	}
	
	for _, existingIndex := range existingIndexes {
		if !expectedIndexNames[existingIndex.Name] {
			fmt.Printf("Warning: unexpected index %s on table %s\n", existingIndex.Name, tableName)
		}
	}
	
	return nil
}

// generateCreateTableQuery generates the CREATE TABLE SQL
func (sm *SchemaManager) generateCreateTableQuery(schema TableSchema) string {
	var query strings.Builder
	
	query.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", schema.Name))
	
	// Add columns
	for i, col := range schema.Columns {
		query.WriteString(fmt.Sprintf("  %s %s", col.Name, col.Type))
		
		if !col.Nullable {
			query.WriteString(" NOT NULL")
		}
		
		if col.Default != "" {
			query.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}
		
		if col.Comment != "" {
			query.WriteString(fmt.Sprintf(" COMMENT '%s'", col.Comment))
		}
		
		if i < len(schema.Columns)-1 {
			query.WriteString(",")
		}
		query.WriteString("\n")
	}
	
	query.WriteString(fmt.Sprintf(") ENGINE = %s\n", schema.Engine))
	
	// Add ORDER BY
	if len(schema.OrderBy) > 0 {
		query.WriteString(fmt.Sprintf("ORDER BY (%s)\n", strings.Join(schema.OrderBy, ", ")))
	}
	
	// Add PARTITION BY
	if schema.PartitionBy != "" {
		query.WriteString(fmt.Sprintf("PARTITION BY %s\n", schema.PartitionBy))
	}
	
	// Add settings
	if len(schema.Settings) > 0 {
		query.WriteString("SETTINGS ")
		settingPairs := make([]string, 0, len(schema.Settings))
		for key, value := range schema.Settings {
			settingPairs = append(settingPairs, fmt.Sprintf("%s = %s", key, value))
		}
		query.WriteString(strings.Join(settingPairs, ", "))
	}
	
	return query.String()
}
