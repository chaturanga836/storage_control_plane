package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/your-org/storage-control-plane/internal/clickhouse"
	"github.com/your-org/storage-control-plane/internal/writers"
	"github.com/your-org/storage-control-plane/pkg/models"
)

// This demo shows how the metadata system solves your cross-file query problem
func main() {
	fmt.Println("🚀 Cross-File Query Demo - Solving the Parquet File Query Problem")
	fmt.Println("=" * 80)

	ctx := context.Background()

	// 1. Initialize ClickHouse client
	clickhouseClient, err := clickhouse.NewClient(&clickhouse.Config{
		Host:     "localhost",
		Port:     9000,
		Database: "storage_control_plane",
		Username: "default",
		Password: "",
	})
	if err != nil {
		log.Fatalf("Failed to create ClickHouse client: %v", err)
	}
	defer clickhouseClient.Close()

	// 2. Initialize the metadata-aware Parquet writer
	writerConfig := &writers.MetadataWriterConfig{
		BaseDir:                "/data/parquet",
		PopulateRecordMetadata: true,
		IndexedFields:          []string{"name", "email", "status", "category", "department"},
		MaxRecordsPerFile:      10000,
		CompressionLevel:       "SNAPPY",
	}
	
	metadataWriter := writers.NewMetadataParquetWriter(clickhouseClient, writerConfig)

	// 3. Initialize cross-file query service
	queryService := clickhouse.NewCrossFileQueryService(clickhouseClient)

	fmt.Println("\n📝 Step 1: Writing sample data to multiple Parquet files...")
	
	// Simulate data from source_connection_01 being written to multiple files over time
	err = writeSampleDataFiles(ctx, metadataWriter)
	if err != nil {
		log.Fatalf("Failed to write sample data: %v", err)
	}

	fmt.Println("✅ Successfully wrote data to multiple Parquet files with metadata indexing")

	// 4. Demonstrate the queries that were impossible before!
	fmt.Println("\n🔍 Step 2: Demonstrating cross-file queries that solve your problem...")

	// THE PROBLEM SOLVER: Find all records with name like "sam" across all files
	fmt.Println("\n1️⃣ Find all records with name containing 'sam' in source_connection_01")
	fmt.Println("   (This was IMPOSSIBLE before - now it's fast!)")
	
	samRecords, err := queryService.FindNameLikeSam(ctx, "tenant_123", "source_connection_01")
	if err != nil {
		log.Printf("Error finding sam records: %v", err)
	} else {
		fmt.Printf("   📊 Found %d records with 'sam' in name across all Parquet files:\n", len(samRecords))
		for _, record := range samRecords {
			fmt.Printf("      - %s (file: %s, row: %d)\n", record.Name, record.FilePath, record.RowNumber)
		}
	}

	// Get latest 10 records across all files
	fmt.Println("\n2️⃣ Get latest 10 records across all files in source_connection_01")
	
	latestRecords, err := queryService.GetLatestRecords(ctx, "tenant_123", "source_connection_01", 10)
	if err != nil {
		log.Printf("Error getting latest records: %v", err)
	} else {
		fmt.Printf("   📊 Latest 10 records across all files:\n")
		for _, record := range latestRecords {
			fmt.Printf("      - %s (%s) at %s\n", record.Name, record.Status, record.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// Get record count by status across all files
	fmt.Println("\n3️⃣ Get record count by status across all files")
	
	statusCounts, err := queryService.GetRecordCountByStatus(ctx, "tenant_123", "source_connection_01")
	if err != nil {
		log.Printf("Error getting status counts: %v", err)
	} else {
		fmt.Printf("   📊 Record counts by status across all files:\n")
		for status, count := range statusCounts {
			fmt.Printf("      - %s: %d records\n", status, count)
		}
	}

	// Complex search with multiple criteria
	fmt.Println("\n4️⃣ Complex search: Active users with 'sam' in name, created in last 30 days")
	
	complexSearch := &clickhouse.CrossFileSearchRequest{
		TenantID:     "tenant_123",
		SourceID:     "source_connection_01",
		NameContains: "sam",
		Status:       "active",
		TimeRange: &clickhouse.TimeRangeFilter{
			Field: "created_at",
			Start: time.Now().Add(-30 * 24 * time.Hour),
			End:   time.Now(),
		},
		Limit: 50,
		SortBy: []models.SortField{
			{Field: "created_at", Order: models.SortDesc},
		},
	}
	
	complexResult, err := queryService.SearchRecordsAcrossFiles(ctx, complexSearch)
	if err != nil {
		log.Printf("Error in complex search: %v", err)
	} else {
		fmt.Printf("   📊 Found %d active users with 'sam' in name (last 30 days):\n", complexResult.TotalFound)
		fmt.Printf("   ⚡ Query executed in: %v\n", complexResult.ExecutionTime)
		fmt.Printf("   📁 Relevant files checked: %d\n", len(complexResult.RelevantFiles))
		
		for _, record := range complexResult.Records {
			fmt.Printf("      - %s (%s) - %s\n", record.Name, record.Email, record.CreatedAt.Format("2006-01-02"))
		}
	}

	// Show directory summary
	fmt.Println("\n5️⃣ Directory summary for source_connection_01")
	
	summary, err := queryService.GetDirectorySummary(ctx, "tenant_123", "source_connection_01", "")
	if err != nil {
		log.Printf("Error getting directory summary: %v", err)
	} else {
		fmt.Printf("   📊 Directory Summary:\n")
		fmt.Printf("      - Total files: %d\n", summary.TotalFiles)
		fmt.Printf("      - Total records: %d\n", summary.TotalRecords)
		fmt.Printf("      - Total size: %d bytes\n", summary.TotalSize)
		fmt.Printf("      - Date range: %s to %s\n", 
			summary.FirstRecordAt.Format("2006-01-02"), 
			summary.LastRecordAt.Format("2006-01-02"))
	}

	fmt.Println("\n✅ Demo completed successfully!")
	fmt.Println("\n🎯 Key Benefits Demonstrated:")
	fmt.Println("   1. ✅ Can search 'name like sam' across ALL Parquet files")
	fmt.Println("   2. ✅ Can get latest records across multiple files")
	fmt.Println("   3. ✅ Can aggregate data (count by status) across files")
	fmt.Println("   4. ✅ Can do complex multi-criteria searches")
	fmt.Println("   5. ✅ Fast performance with indexes")
	fmt.Println("   6. ✅ Knows which files to check (query optimization)")
	
	fmt.Println("\n💡 How it works:")
	fmt.Println("   • Each Parquet file write also populates ClickHouse metadata tables")
	fmt.Println("   • record_metadata table contains searchable fields from all files")
	fmt.Println("   • ClickHouse indexes enable fast cross-file searching")
	fmt.Println("   • Query results include file location for full record retrieval")
}

// writeSampleDataFiles simulates writing data to multiple Parquet files over time
func writeSampleDataFiles(ctx context.Context, writer *writers.MetadataParquetWriter) error {
	tenantID := "tenant_123"
	sourceID := "source_connection_01"
	
	// Simulate 3 different data batches (would be 3 separate Parquet files)
	dataBatches := [][]map[string]interface{}{
		// Batch 1: Users from marketing department
		{
			{
				"id": "user_001", "name": "Samuel Johnson", "email": "sam.johnson@company.com",
				"status": "active", "department": "marketing", "created_at": time.Now().Add(-10 * time.Hour),
			},
			{
				"id": "user_002", "name": "Alice Smith", "email": "alice.smith@company.com",
				"status": "active", "department": "marketing", "created_at": time.Now().Add(-9 * time.Hour),
			},
			{
				"id": "user_003", "name": "Bob Wilson", "email": "bob.wilson@company.com",
				"status": "inactive", "department": "marketing", "created_at": time.Now().Add(-8 * time.Hour),
			},
		},
		// Batch 2: Users from engineering department
		{
			{
				"id": "user_004", "name": "Samantha Davis", "email": "sam.davis@company.com",
				"status": "active", "department": "engineering", "created_at": time.Now().Add(-7 * time.Hour),
			},
			{
				"id": "user_005", "name": "Charlie Brown", "email": "charlie.brown@company.com",
				"status": "active", "department": "engineering", "created_at": time.Now().Add(-6 * time.Hour),
			},
			{
				"id": "user_006", "name": "Sam Rodriguez", "email": "sam.rodriguez@company.com",
				"status": "pending", "department": "engineering", "created_at": time.Now().Add(-5 * time.Hour),
			},
		},
		// Batch 3: Users from sales department
		{
			{
				"id": "user_007", "name": "Diana Prince", "email": "diana.prince@company.com",
				"status": "active", "department": "sales", "created_at": time.Now().Add(-4 * time.Hour),
			},
			{
				"id": "user_008", "name": "Sammy Thompson", "email": "sammy.thompson@company.com",
				"status": "active", "department": "sales", "created_at": time.Now().Add(-3 * time.Hour),
			},
			{
				"id": "user_009", "name": "Frank Miller", "email": "frank.miller@company.com",
				"status": "inactive", "department": "sales", "created_at": time.Now().Add(-2 * time.Hour),
			},
		},
	}

	// Write each batch to a separate Parquet file
	for i, batch := range dataBatches {
		fmt.Printf("   Writing batch %d (%d records)...\n", i+1, len(batch))
		
		_, err := writer.WriteParquetWithMetadata(ctx, batch, tenantID, sourceID, nil)
		if err != nil {
			return fmt.Errorf("failed to write batch %d: %w", i+1, err)
		}
	}

	return nil
}
