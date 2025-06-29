// parquet_demo.go - Demonstrates Parquet file writing workflow
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/wal"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/writers"
)

func main() {
	fmt.Println("🚀 Parquet Writing Demo")
	fmt.Println("=======================")

	// Setup directories
	walDir := "demo_data/wal/demo_tenant/demo_source"
	dataDir := "demo_data/parquet"

	// Clean up previous demo data
	os.RemoveAll("demo_data")
	defer os.RemoveAll("demo_data")

	// Demo 1: Direct Parquet Writing
	fmt.Println("\n📝 Demo 1: Direct Parquet Writing")
	demonstrateDirectParquetWrite(dataDir)

	// Demo 2: WAL-to-Parquet Workflow
	fmt.Println("\n📝 Demo 2: WAL-to-Parquet Workflow")
	demonstrateWALWorkflow(walDir, dataDir)

	// Demo 3: Reading Parquet Metadata
	fmt.Println("\n📝 Demo 3: Reading Parquet Metadata")
	demonstrateMetadataReading(dataDir)

	fmt.Println("\n✅ Parquet demo completed!")
}

func demonstrateDirectParquetWrite(dataDir string) {
	// Create sample records
	records := []writers.ParquetRecord{
		{
			"user_id":    "user_123",
			"event_type": "page_view",
			"page":       "/home",
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"metadata": map[string]interface{}{
				"user_agent": "Mozilla/5.0",
				"ip":         "192.168.1.1",
			},
		},
		{
			"user_id":    "user_456",
			"event_type": "button_click",
			"button_id":  "signup_btn",
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"metadata": map[string]interface{}{
				"session_id": "sess_789",
				"experiment": "test_variant_a",
			},
		},
		{
			"user_id":    "user_123",
			"event_type": "purchase",
			"amount":     29.99,
			"currency":   "USD",
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"items": []map[string]interface{}{
				{"sku": "ITEM_001", "quantity": 2},
				{"sku": "ITEM_002", "quantity": 1},
			},
		},
	}

	// Write to Parquet
	targetDir := fmt.Sprintf("%s/demo_tenant/demo_source/schema_001/%s", 
		dataDir, time.Now().Format("2006-01-02"))

	filePath, err := writers.WriteParquetFile(records, targetDir, "events")
	if err != nil {
		log.Fatalf("❌ Failed to write Parquet file: %v", err)
	}

	// Write metadata
	metadata := writers.FileMetadata{
		TenantID:    "demo_tenant",
		SourceID:    "demo_source",
		FilePath:    filePath,
		RowCount:    uint64(len(records)),
		MinTS:       time.Now().Add(-1*time.Hour).Format(time.RFC3339),
		MaxTS:       time.Now().Format(time.RFC3339),
		SchemaHash:  "schema_001",
		ColumnStats: `{"user_id": {"unique": 2}, "event_type": {"unique": 3}}`,
	}

	if err := writers.WriteMetadata(metadata, targetDir); err != nil {
		log.Printf("⚠️ Failed to write metadata: %v", err)
	}

	fmt.Printf("   ✅ Wrote %d records to: %s\n", len(records), filePath)
	
	// Show file size
	if info, err := os.Stat(filePath); err == nil {
		fmt.Printf("   📊 File size: %.2f KB\n", float64(info.Size())/1024)
	}
}

func demonstrateWALWorkflow(walDir, dataDir string) {
	// Step 1: Write records to WAL
	fmt.Println("   Step 1: Writing records to WAL...")
	
	sampleEvents := []map[string]interface{}{
		{
			"event": "user_registration",
			"user_id": "new_user_001",
			"email": "user@example.com",
			"source": "website",
		},
		{
			"event": "product_view",
			"user_id": "user_002", 
			"product_id": "prod_123",
			"category": "electronics",
		},
		{
			"event": "cart_addition",
			"user_id": "user_002",
			"product_id": "prod_123",
			"quantity": 2,
		},
		{
			"event": "purchase_complete",
			"user_id": "user_002",
			"order_id": "order_456",
			"total": 199.99,
		},
		{
			"event": "user_logout",
			"user_id": "user_002",
			"session_duration": 1800, // 30 minutes
		},
	}

	for i, event := range sampleEvents {
		if err := wal.AppendToWAL(walDir, event); err != nil {
			log.Printf("❌ Failed to write WAL record %d: %v", i, err)
		}
	}

	fmt.Printf("   ✅ Wrote %d records to WAL\n", len(sampleEvents))

	// Step 2: Read WAL file to verify
	walFile := fmt.Sprintf("%s/wal_current.jsonl", walDir)
	records, err := wal.ReadWALFile(walFile)
	if err != nil {
		log.Printf("❌ Failed to read WAL file: %v", err)
		return
	}

	fmt.Printf("   📖 Read %d records from WAL\n", len(records))

	// Step 3: Manually trigger flush (normally done by background process)
	fmt.Println("   Step 2: Flushing WAL to Parquet...")
	// In the real system, this would be done by the background flusher
	// For demo purposes, we'll simulate it
	if len(records) >= 3 { // Simulate threshold
		fmt.Printf("   🔄 Threshold reached (%d >= 3), flushing to Parquet...\n", len(records))
		// The actual flush would be handled by wal.FlushWAL() in production
	}
}

func demonstrateMetadataReading(dataDir string) {
	// Walk the data directory to find metadata files
	metadataFiles := []string{}
	
	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "_stats.json" {
			metadataFiles = append(metadataFiles, path)
		}
		return nil
	})

	if err != nil {
		log.Printf("❌ Failed to walk directory: %v", err)
		return
	}

	fmt.Printf("   📊 Found %d metadata files:\n", len(metadataFiles))

	for _, metaFile := range metadataFiles {
		data, err := os.ReadFile(metaFile)
		if err != nil {
			continue
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(data, &metadata); err != nil {
			continue
		}

		fmt.Printf("   • %s:\n", metaFile)
		if tenant, ok := metadata["TenantID"]; ok {
			fmt.Printf("     - Tenant: %v\n", tenant)
		}
		if source, ok := metadata["SourceID"]; ok {
			fmt.Printf("     - Source: %v\n", source)
		}
		if rows, ok := metadata["RowCount"]; ok {
			fmt.Printf("     - Rows: %v\n", rows)
		}
		if filePath, ok := metadata["FilePath"]; ok {
			fmt.Printf("     - File: %v\n", filePath)
		}
	}
}
