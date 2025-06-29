package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/writers"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

// DataLakeDemo demonstrates the complete data lake functionality
func main() {
	fmt.Println("🏗️ Data Lake Architecture Demo")
	fmt.Println("==============================")

	// Initialize services
	dirConfigSvc := models.NewDirectoryConfigService()
	
	// 1. Setup Source Connection with custom directory structure
	fmt.Println("\n📡 1. Setting up Source Connection with Custom Directory Structure")
	sourceConnection := setupSourceConnection()
	
	// 2. Create custom directory configuration
	fmt.Println("\n📂 2. Creating Custom Directory Configuration")
	dirConfig := createCustomDirectoryConfig(sourceConnection)
	
	if err := dirConfigSvc.CreateConfig(dirConfig); err != nil {
		log.Fatalf("Failed to create directory config: %v", err)
	}
	
	// 3. Simulate API data ingestion
	fmt.Println("\n📊 3. Simulating API Data Ingestion")
	apiData := simulateAPIData()
	
	// 4. Process data through the enhanced Parquet writer
	fmt.Println("\n💾 4. Processing Data with Enhanced Parquet Writer")
	demonstrateParquetWriting(dirConfigSvc, sourceConnection, apiData)
	
	// 5. Query data using ClickHouse metadata
	fmt.Println("\n🔍 5. Querying Data using ClickHouse Metadata")
	demonstrateDataQuerying()
	
	// 6. Show directory structure and metadata
	fmt.Println("\n📁 6. Generated Directory Structure")
	demonstrateDirectoryStructure(dirConfig, apiData[0])
	
	fmt.Println("\n✅ Demo completed successfully!")
	fmt.Println("\nYour data lake now supports:")
	fmt.Println("- ✅ User-defined directory structures")
	fmt.Println("- ✅ Automatic Parquet file generation")
	fmt.Println("- ✅ Rich metadata and statistics")
	fmt.Println("- ✅ ClickHouse query optimization")
	fmt.Println("- ✅ Schema evolution tracking")
	fmt.Println("- ✅ Performance monitoring")
}

func setupSourceConnection() *models.SourceConnection {
	return &models.SourceConnection{
		ID:       "src_customer_api",
		TenantID: "tenant_acme_corp",
		Name:     "Customer Management API",
		Type:     models.SourceAPI,
		Status:   models.StatusActive,
		Config: map[string]any{
			"endpoint":    "https://api.acme.com/customers",
			"auth_token":  "bearer_token_here",
			"polling_interval": "5m",
		},
		SyncConfig: models.SourceSyncConfig{
			Enabled:        true,
			Interval:       5 * time.Minute,
			BatchSize:      1000,
			ConcurrentJobs: 2,
			RetryAttempts:  3,
			RetryBackoff:   30 * time.Second,
			TimeoutSeconds: 60,
		},
		SchemaConfig: models.SourceSchemaConfig{
			AutoDetectSchema:    true,
			RequiredFields:      []string{"customer_id", "email", "created_at"},
			TimestampField:      "created_at",
			TimestampFormat:     "2006-01-02T15:04:05Z",
			SchemaEvolution: models.SchemaEvolution{
				AllowNewFields:    true,
				AllowFieldRemoval: false,
				AllowTypeChanges:  false,
				StrictMode:        false,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createCustomDirectoryConfig(sourceConnection *models.SourceConnection) *models.DirectoryConfig {
	return &models.DirectoryConfig{
		ID:                 "config_customer_hierarchy",
		TenantID:           sourceConnection.TenantID,
		SourceConnectionID: sourceConnection.ID,
		Name:               "Customer Data Hierarchy",
		Pattern:            "{tenant_id}/{source_id}/{region}/{customer_type}/{year}/{month}/{day}",
		Variables: map[string]models.Variable{
			"tenant_id": {
				Name:     "tenant_id",
				Type:     models.VarTypeStatic,
				Source:   sourceConnection.TenantID,
				Required: true,
			},
			"source_id": {
				Name:     "source_id", 
				Type:     models.VarTypeStatic,
				Source:   sourceConnection.ID,
				Required: true,
			},
			"region": {
				Name:         "region",
				Type:         models.VarTypeJSONField,
				Source:       "customer.region",
				DefaultValue: "unknown",
				Validation:   "^(north|south|east|west|central)$",
			},
			"customer_type": {
				Name:         "customer_type",
				Type:         models.VarTypeJSONField,
				Source:       "customer.type",
				DefaultValue: "standard",
				Validation:   "^(premium|standard|basic)$",
			},
			"year": {
				Name:     "year",
				Type:     models.VarTypeDateTime,
				Source:   "year",
				Required: true,
			},
			"month": {
				Name:     "month",
				Type:     models.VarTypeDateTime,
				Source:   "month",
				Required: true,
			},
			"day": {
				Name:     "day",
				Type:     models.VarTypeDateTime,
				Source:   "day",
				Required: true,
			},
		},
		FileNaming: models.FileNamingConfig{
			Pattern:     "customers_{region}_{timestamp}_{sequence}",
			Timestamp:   "2006-01-02_15-04-05",
			Sequence:    true,
			SchemaHash:  true,
			MaxFileSize: 100 * 1024 * 1024, // 100MB
		},
		MetadataConfig: models.MetadataConfig{
			GenerateSummary:   true,
			GenerateIndexes:   true,
			GenerateStats:     true,
			GenerateSchema:    true,
			IndexedFields:     []string{"customer.region", "customer.type", "customer.id", "created_at"},
			StatsFields:       []string{"customer.age", "order_total", "last_login_days"},
			CustomMetadata: map[string]interface{}{
				"data_source":      "customer_api",
				"compliance_level": "gdpr",
				"retention_days":   2555, // 7 years
			},
		},
		IsActive: true,
	}
}

func simulateAPIData() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"customer_id": "cust_001",
			"email":       "john.doe@example.com",
			"created_at":  "2025-06-30T10:15:00Z",
			"customer": map[string]interface{}{
				"id":     "cust_001",
				"name":   "John Doe",
				"region": "north",
				"type":   "premium",
				"age":    35,
				"address": map[string]interface{}{
					"street": "123 Main St",
					"city":   "New York",
					"state":  "NY",
					"zip":    "10001",
				},
			},
			"orders": []map[string]interface{}{
				{
					"order_id": "ord_001",
					"total":    299.99,
					"date":     "2025-06-25T14:30:00Z",
				},
			},
			"order_total":      299.99,
			"last_login_days":  3,
		},
		{
			"customer_id": "cust_002", 
			"email":       "jane.smith@example.com",
			"created_at":  "2025-06-30T10:20:00Z",
			"customer": map[string]interface{}{
				"id":     "cust_002",
				"name":   "Jane Smith",
				"region": "south",
				"type":   "standard",
				"age":    28,
				"address": map[string]interface{}{
					"street": "456 Oak Ave",
					"city":   "Atlanta",
					"state":  "GA",
					"zip":    "30301",
				},
			},
			"orders": []map[string]interface{}{
				{
					"order_id": "ord_002",
					"total":    149.99,
					"date":     "2025-06-28T09:15:00Z",
				},
			},
			"order_total":      149.99,
			"last_login_days":  1,
		},
		{
			"customer_id": "cust_003",
			"email":       "bob.wilson@example.com", 
			"created_at":  "2025-06-30T10:25:00Z",
			"customer": map[string]interface{}{
				"id":     "cust_003",
				"name":   "Bob Wilson",
				"region": "west",
				"type":   "basic",
				"age":    42,
				"address": map[string]interface{}{
					"street": "789 Pine Rd",
					"city":   "Seattle",
					"state":  "WA",
					"zip":    "98101",
				},
			},
			"orders": []map[string]interface{}{
				{
					"order_id": "ord_003",
					"total":    79.99,
					"date":     "2025-06-29T16:45:00Z",
				},
			},
			"order_total":      79.99,
			"last_login_days":  7,
		},
	}
}

func demonstrateParquetWriting(dirConfigSvc *models.DirectoryConfigService, sourceConnection *models.SourceConnection, apiData []map[string]interface{}) {
	// This would normally be initialized with real ClickHouse connection
	// metadataWriter := writers.NewEnhancedMetadataWriter(clickhouseConn)
	// parquetWriter := writers.NewEnhancedParquetWriter("/data", dirConfigSvc, metadataWriter)

	fmt.Printf("📝 Processing %d records from source: %s\n", len(apiData), sourceConnection.Name)
	
	// Simulate the directory resolution
	dirConfig, _ := dirConfigSvc.GetConfigBySourceConnection(sourceConnection.ID)
	
	for i, record := range apiData {
		fmt.Printf("\n📄 Record %d: %s\n", i+1, record["customer_id"])
		
		// Resolve directory path
		resolver := &models.DirectoryResolver{
			Config:    dirConfig,
			Timestamp: time.Now(),
			Data:      record,
		}
		
		dirPath, err := resolver.ResolvePath()
		if err != nil {
			fmt.Printf("❌ Error resolving path: %v\n", err)
			continue
		}
		
		filename := resolver.ResolveFileName(i)
		
		fmt.Printf("   📁 Directory: %s\n", dirPath)
		fmt.Printf("   📄 Filename: %s\n", filename)
		
		// Show what metadata would be generated
		fmt.Printf("   📊 Metadata to generate:\n")
		if dirConfig.MetadataConfig.GenerateSummary {
			fmt.Printf("      ✅ _summary.json\n")
		}
		if dirConfig.MetadataConfig.GenerateSchema {
			fmt.Printf("      ✅ _schema.json\n")
		}
		if dirConfig.MetadataConfig.GenerateStats {
			fmt.Printf("      ✅ _stats.json (fields: %v)\n", dirConfig.MetadataConfig.StatsFields)
		}
		if dirConfig.MetadataConfig.GenerateIndexes {
			fmt.Printf("      ✅ _indexes.json (fields: %v)\n", dirConfig.MetadataConfig.IndexedFields)
		}
		if len(dirConfig.MetadataConfig.CustomMetadata) > 0 {
			fmt.Printf("      ✅ _custom.json\n")
		}
	}
}

func demonstrateDataQuerying() {
	fmt.Println("🔍 Example ClickHouse queries that would be generated:")
	
	queries := []struct {
		name        string
		description string
		sql         string
	}{
		{
			name:        "Get Source Connections by Tenant",
			description: "List all source connections for tenant_acme_corp",
			sql: `
SELECT id, name, type, status, last_sync, created_at
FROM source_connections 
WHERE tenant_id = 'tenant_acme_corp'
ORDER BY created_at DESC`,
		},
		{
			name:        "Find Customer Data Files",
			description: "Find all Parquet files containing customer data in North region",
			sql: `
SELECT file_path, record_count, min_timestamp, max_timestamp
FROM parquet_files pf
JOIN directory_structures ds ON ds.tenant_id = pf.tenant_id
WHERE pf.tenant_id = 'tenant_acme_corp'
  AND pf.source_id = 'src_customer_api'
  AND ds.directory_path LIKE '%/north/%'
  AND pf.min_timestamp >= '2025-06-30'
ORDER BY pf.min_timestamp DESC`,
		},
		{
			name:        "Get Directory Statistics",
			description: "Get statistics for customer data directories",
			sql: `
SELECT 
    directory_path,
    total_files,
    total_records,
    total_size_bytes / (1024*1024) as size_mb,
    unique_schemas,
    compression_ratio
FROM directory_structures
WHERE tenant_id = 'tenant_acme_corp'
  AND source_connection_id = 'src_customer_api'
ORDER BY last_updated DESC`,
		},
		{
			name:        "Performance Monitoring",
			description: "Monitor source connection performance",
			sql: `
SELECT 
    timestamp,
    total_records_processed,
    last_sync_duration_ms,
    success_rate,
    error_count
FROM source_connection_metrics
WHERE source_connection_id = 'src_customer_api'
  AND timestamp >= now() - INTERVAL 24 HOUR
ORDER BY timestamp DESC`,
		},
	}
	
	for _, query := range queries {
		fmt.Printf("\n📊 %s\n", query.name)
		fmt.Printf("   %s\n", query.description)
		fmt.Printf("   SQL: %s\n", query.sql)
	}
}

func demonstrateDirectoryStructure(dirConfig *models.DirectoryConfig, sampleRecord map[string]interface{}) {
	fmt.Println("📁 Generated Directory Structure:")
	
	resolver := &models.DirectoryResolver{
		Config:    dirConfig,
		Timestamp: time.Now(),
		Data:      sampleRecord,
	}
	
	dirPath, _ := resolver.ResolvePath()
	filename := resolver.ResolveFileName(0)
	
	fmt.Printf(`
data/
└── %s/
    ├── %s
    ├── _summary.json      # Directory statistics
    ├── _schema.json       # Schema definition
    ├── _stats.json        # Field statistics  
    ├── _indexes.json      # Search indexes
    └── _custom.json       # Custom metadata
`, dirPath, filename)

	// Show sample metadata content
	fmt.Println("\n📄 Sample _summary.json content:")
	summary := map[string]interface{}{
		"directory_path":    dirPath,
		"total_files":       1,
		"total_records":     3,
		"total_size_bytes":  2048,
		"schema_hash":       "sha1_abc123",
		"min_timestamp":     "2025-06-30T10:15:00Z",
		"max_timestamp":     "2025-06-30T10:25:00Z",
		"last_updated":      time.Now().Format(time.RFC3339),
		"unique_schemas":    1,
		"compression_ratio": 0.65,
	}
	
	summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
	fmt.Println(string(summaryJSON))
	
	fmt.Println("\n📄 Sample _indexes.json content:")
	indexes := map[string]interface{}{
		"field_indexes": map[string]interface{}{
			"customer.region": map[string]interface{}{
				"unique_values": map[string][]int{
					"north": {0},
					"south": {1}, 
					"west":  {2},
				},
				"data_type": "string",
			},
			"customer.type": map[string]interface{}{
				"unique_values": map[string][]int{
					"premium":  {0},
					"standard": {1},
					"basic":    {2},
				},
				"data_type": "string",
			},
		},
		"total_records": 3,
		"generated_at":  time.Now().Format(time.RFC3339),
	}
	
	indexJSON, _ := json.MarshalIndent(indexes, "", "  ")
	fmt.Println(string(indexJSON))
}
