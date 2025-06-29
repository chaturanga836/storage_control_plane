// query_flow_demo.go - Demonstrates how data queries work from API to ClickHouse to Storage
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/utils"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

func main() {
	fmt.Println("🔄 Data Query Flow Demo: API → ClickHouse → Storage")
	fmt.Println("=" + strings.Repeat("=", 59))

	// 1. Simulate incoming HTTP request with sorting parameters
	fmt.Println("\n📥 1. HTTP Request with Sorting Parameters")
	queryRequest := models.QueryRequest{
		TenantID:  "tenant_123",
		Query:     "SELECT * FROM analytics_events WHERE tenant_id = 'tenant_123'",
		QueryType: models.QueryTimeSeries,
		SortBy: []models.SortField{
			{Field: "timestamp", Direction: models.SortDesc},
			{Field: "value", Direction: models.SortAsc},
		},
		Limit:  1000,
		Offset: 0,
		TimeRange: &models.TimeRange{
			Start: time.Now().Add(-24 * time.Hour),
			End:   time.Now(),
		},
	}
	
	requestJSON, _ := json.MarshalIndent(queryRequest, "", "  ")
	fmt.Printf("Request: %s\n", requestJSON)

	// 2. API Server validates and processes sort parameters
	fmt.Println("\n🔍 2. API Server: Sort Validation & Processing")
	
	// Choose appropriate sort options based on query type
	var sortOptions utils.SortOptions
	switch queryRequest.QueryType {
	case models.QueryTimeSeries:
		sortOptions = utils.AnalyticsSortOptions
	case models.QuerySQL:
		sortOptions = utils.DataIngestionSortOptions
	default:
		sortOptions = utils.TenantSortOptions
	}
	
	fmt.Printf("Selected Sort Options: %+v\n", sortOptions)
	
	// Validate sort fields
	validatedSorts, err := utils.ValidateSortFields(queryRequest.SortBy, sortOptions)
	if err != nil {
		log.Fatalf("Sort validation failed: %v", err)
	}
	
	fmt.Printf("Validated Sort Fields: %+v\n", validatedSorts)

	// 3. Large-Scale Query Planning
	fmt.Println("\n📊 3. Large-Scale Query Planning")
	
	// Estimate dataset size (simulated)
	estimatedRows := int64(2500000) // 2.5M rows
	fmt.Printf("Estimated dataset size: %d rows\n", estimatedRows)
	
	// Apply large-scale optimizations
	config := utils.DefaultLargeScaleConfig
	optimizedSorts, optimizedConfig, err := utils.ValidateSortFieldsForScale(
		validatedSorts, sortOptions, config, estimatedRows)
	if err != nil {
		fmt.Printf("❌ Large-scale validation failed (expected): %v\n", err)
		fmt.Println("💡 This shows the system correctly prevents inefficient queries!")
		
		// Try with only indexed fields
		fmt.Println("\n🔄 Retrying with only indexed fields...")
		indexedOnlySorts := []models.SortField{
			{Field: "timestamp", Direction: models.SortDesc},
		}
		
		validatedIndexedSorts, err := utils.ValidateSortFields(indexedOnlySorts, sortOptions)
		if err != nil {
			log.Fatalf("Indexed sort validation failed: %v", err)
		}
		
		optimizedSorts, optimizedConfig, err = utils.ValidateSortFieldsForScale(
			validatedIndexedSorts, sortOptions, config, estimatedRows)
		if err != nil {
			log.Fatalf("Large-scale validation failed even with indexed fields: %v", err)
		}
		
		fmt.Println("✅ Success with indexed fields!")
	}
	
	fmt.Printf("Large-scale config: %+v\n", optimizedConfig)
	fmt.Printf("Streaming enabled: %v\n", optimizedConfig.UseStreaming)
	
	// Analyze query complexity
	complexity, recommendations := utils.EstimateQueryComplexity(
		optimizedSorts, estimatedRows, *optimizedConfig)
	
	fmt.Printf("Query complexity: %s\n", complexity)
	fmt.Printf("Recommendations: %v\n", recommendations)

	// 4. Query Generation
	fmt.Println("\n🔧 4. ClickHouse Query Generation")
	
	// Generate optimized ClickHouse query
	optimizedQuery := utils.GenerateOptimizedClickHouseQuery(
		queryRequest.Query, optimizedSorts, *optimizedConfig, 
		queryRequest.Limit, queryRequest.Offset)
	
	fmt.Printf("Optimized Query:\n%s\n", optimizedQuery)
	
	// Generate ORDER BY clause
	orderByClause := utils.GenerateClickHouseOrderBy(optimizedSorts)
	fmt.Printf("ORDER BY clause: %s\n", orderByClause)

	// 5. Streaming Query Execution (for large datasets)
	fmt.Println("\n🚀 5. Streaming Query Execution")
	
	if optimizedConfig.UseStreaming {
		fmt.Println("Using streaming execution for large dataset...")
		
		// Simulate streaming query generation
		lastValues := map[string]any{
			"timestamp": time.Now().Add(-12 * time.Hour),
			"value":     100.5,
		}
		
		streamingQuery, err := utils.GenerateStreamingQuery(
			queryRequest.Query, optimizedSorts, optimizedConfig.ChunkSize, lastValues)
		if err != nil {
			log.Fatalf("Streaming query generation failed: %v", err)
		}
		
		fmt.Printf("Streaming Query (with cursor):\n%s\n", streamingQuery)
		fmt.Printf("Chunk size: %d rows\n", optimizedConfig.ChunkSize)
	}

	// 6. ClickHouse Execution Flow
	fmt.Println("\n💾 6. ClickHouse Database Execution")
	
	// This would normally connect to actual ClickHouse
	// For demo, we'll show the execution flow
	fmt.Println("Steps in ClickHouse execution:")
	fmt.Println("  1. Parse SQL query")
	fmt.Println("  2. Apply WHERE filters (tenant_id, time_range)")
	fmt.Println("  3. Use indexes for sort fields (timestamp, value)")
	fmt.Println("  4. Apply ORDER BY with optimized sorting")
	fmt.Println("  5. Apply LIMIT/OFFSET for pagination")
	fmt.Println("  6. Return result set")
	
	// Example result structure
	sampleResult := []map[string]any{
		{
			"timestamp": time.Now().Add(-1 * time.Hour),
			"value":     95.7,
			"tenant_id": "tenant_123",
			"event_id":  "event_001",
		},
		{
			"timestamp": time.Now().Add(-2 * time.Hour),
			"value":     87.3,
			"tenant_id": "tenant_123",
			"event_id":  "event_002",
		},
	}
	
	fmt.Printf("Sample result (first 2 rows): %+v\n", sampleResult)

	// 7. Storage Layer Access
	fmt.Println("\n📁 7. Storage Layer Access")
	fmt.Println("ClickHouse accesses data from:")
	fmt.Println("  • Parquet files (columnar storage)")
	fmt.Println("  • Memory tables (recent data)")
	fmt.Println("  • Distributed tables (if clustered)")
	fmt.Println("  • Index files (for fast sorting)")
	
	// Storage optimization info
	fmt.Println("\nStorage optimizations for large-scale sorting:")
	fmt.Println("  ✓ Columnar format reduces I/O for specific fields")
	fmt.Println("  ✓ Compression reduces data transfer")
	fmt.Println("  ✓ Index hints improve sort performance")
	fmt.Println("  ✓ Partitioning by tenant/time for parallel processing")

	// 8. Performance Monitoring
	fmt.Println("\n📈 8. Performance Monitoring")
	
	performanceMetrics := utils.SortPerformanceMetrics{
		QueryTime:       1250,  // milliseconds
		RowsProcessed:   2500000,
		MemoryUsed:      512 * 1024 * 1024, // 512MB
		IndexesUsed:     []string{"idx_timestamp", "idx_tenant_id"},
		StreamingUsed:   true,
		ChunksProcessed: 250, // 250 chunks of 10K rows each
	}
	
	fmt.Printf("Performance metrics: %+v\n", performanceMetrics)
	
	// Calculate performance stats
	rowsPerSecond := float64(performanceMetrics.RowsProcessed) / (float64(performanceMetrics.QueryTime) / 1000)
	fmt.Printf("Processing rate: %.0f rows/second\n", rowsPerSecond)
	fmt.Printf("Memory efficiency: %.2f MB per million rows\n", 
		float64(performanceMetrics.MemoryUsed/1024/1024)/float64(performanceMetrics.RowsProcessed/1000000))

	fmt.Println("\n✅ Query Flow Demo Complete!")
	fmt.Println("\nKey Points:")
	fmt.Println("• Sort validation ensures security and performance")
	fmt.Println("• Large datasets automatically use streaming")
	fmt.Println("• Index hints optimize database query execution")
	fmt.Println("• Cursor-based pagination handles massive results")
	fmt.Println("• Performance monitoring tracks efficiency")
	
	// Run the index demo
	runIndexDemo()
}
