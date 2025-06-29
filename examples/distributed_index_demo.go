package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/clickhouse"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

// DistributedIndexDemo demonstrates how indexes work with horizontal scaling
func main() {
	fmt.Println("=== Distributed Index Management Demo ===")
	
	// 1. Setup distributed cluster configuration
	fmt.Println("\n1. Setting up distributed cluster...")
	config := setupDistributedCluster()
	
	// 2. Initialize distributed index manager
	fmt.Println("\n2. Initializing distributed index manager...")
	manager, err := clickhouse.NewDistributedIndexManager(config)
	if err != nil {
		log.Fatal("Failed to create distributed manager:", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// 3. Create distributed tables and indexes
	fmt.Println("\n3. Creating distributed tables and indexes...")
	if err := createDistributedSchema(ctx, manager); err != nil {
		log.Fatal("Failed to create schema:", err)
	}

	// 4. Demonstrate different index strategies
	fmt.Println("\n4. Demonstrating index strategies...")
	demonstrateIndexStrategies(ctx, manager)

	// 5. Show query optimization for distributed systems
	fmt.Println("\n5. Demonstrating query optimization...")
	demonstrateQueryOptimization(ctx, manager)

	// 6. Monitor distributed index performance
	fmt.Println("\n6. Monitoring distributed index performance...")
	monitorDistributedPerformance(ctx, manager)

	// 7. Perform distributed maintenance
	fmt.Println("\n7. Performing distributed maintenance...")
	performDistributedMaintenance(ctx, manager)

	fmt.Println("\n=== Demo completed successfully! ===")
}

// setupDistributedCluster creates configuration for a 3-node ClickHouse cluster
func setupDistributedCluster() clickhouse.DistributedIndexConfig {
	return clickhouse.DistributedIndexConfig{
		ClusterName: "analytics_cluster",
		Nodes: []clickhouse.NodeInfo{
			{
				Host:     "localhost", // In production: actual hostnames
				Port:     9000,
				Database: "analytics_shard_1",
				Shard:    1,
				Replica:  1,
				Weight:   100,
			},
			{
				Host:     "localhost",
				Port:     9001, // Different port for demo
				Database: "analytics_shard_2", 
				Shard:    2,
				Replica:  1,
				Weight:   100,
			},
			{
				Host:     "localhost",
				Port:     9002, // Different port for demo
				Database: "analytics_shard_3",
				Shard:    3,
				Replica:  1,
				Weight:   100,
			},
		},
		ReplicationFactor: 1, // No replication for demo
		PartitionKey:     "tenant_id",
		ShardingKey:      "tenant_id", 
		IndexStrategy:    clickhouse.PartitionedIndexes,
	}
}

// createDistributedSchema creates tables and indexes across the cluster
func createDistributedSchema(ctx context.Context, manager *clickhouse.DistributedIndexManager) error {
	fmt.Println("  Creating distributed events table...")
	
	// Create indexes for different query patterns
	indexes := []struct {
		name    string
		columns []string
		itype   clickhouse.IndexType
		purpose string
	}{
		{"idx_tenant_time", []string{"tenant_id", "timestamp"}, clickhouse.IndexTypeMinMax, "Tenant-based time range queries"},
		{"idx_event_bloom", []string{"event_id"}, clickhouse.IndexTypeBloomFilter, "Exact event ID lookups"},
		{"idx_event_type", []string{"event_type"}, clickhouse.IndexTypeSet, "Event type filtering"},
		{"idx_user_bloom", []string{"user_id"}, clickhouse.IndexTypeBloomFilter, "User-specific queries"},
		{"idx_timestamp_minmax", []string{"timestamp"}, clickhouse.IndexTypeMinMax, "Time-based range queries"},
	}

	for _, idx := range indexes {
		fmt.Printf("    Creating %s (%s)...\n", idx.name, idx.purpose)
		err := manager.CreateDistributedIndex(ctx, "events", idx.name, idx.columns, idx.itype)
		if err != nil {
			return fmt.Errorf("failed to create index %s: %w", idx.name, err)
		}
	}

	return nil
}

// demonstrateIndexStrategies shows different indexing approaches
func demonstrateIndexStrategies(ctx context.Context, manager *clickhouse.DistributedIndexManager) {
	fmt.Println("  Analyzing index strategies across nodes...")

	// Get index information from all nodes
	indexInfo, err := manager.GetDistributedIndexInfo(ctx, "events")
	if err != nil {
		log.Printf("Failed to get index info: %v", err)
		return
	}

	for nodeIndex, indexes := range indexInfo {
		fmt.Printf("    Node %d indexes:\n", nodeIndex)
		for _, idx := range indexes {
			fmt.Printf("      %s (%s): %s [granularity: %d]\n", 
				idx.Name, idx.Type, idx.Expression, idx.Granularity)
		}
	}

	// Show cluster topology
	topology, err := manager.GetClusterTopology(ctx)
	if err == nil {
		fmt.Println("    Cluster topology:")
		for _, node := range topology {
			fmt.Printf("      Shard %d, Replica %d: %s:%d (weight: %d)\n",
				node.Shard, node.Replica, node.Host, node.Port, node.Weight)
		}
	}
}

// demonstrateQueryOptimization shows how queries are optimized for distribution
func demonstrateQueryOptimization(ctx context.Context, manager *clickhouse.DistributedIndexManager) {
	queries := []struct {
		name        string
		query       string
		sortFields  []models.SortField
		whereFields map[string]interface{}
		explanation string
	}{
		{
			name: "Tenant-specific query (optimal)",
			query: `
				SELECT event_type, COUNT(*) as count 
				FROM events_distributed 
				WHERE tenant_id = ? AND timestamp >= ?
				GROUP BY event_type
			`,
			sortFields: []models.SortField{
				{Field: "count", Order: models.DESC},
			},
			whereFields: map[string]interface{}{
				"tenant_id": "tenant_123",
				"timestamp": "2024-01-01",
			},
			explanation: "Single shard query with partition pruning",
		},
		{
			name: "Event ID lookup (good)",
			query: `
				SELECT tenant_id, timestamp, data 
				FROM events_distributed 
				WHERE event_id = ?
			`,
			sortFields: []models.SortField{
				{Field: "timestamp", Order: models.DESC},
			},
			whereFields: map[string]interface{}{
				"event_id": "event_456",
			},
			explanation: "Bloom filter index on all shards",
		},
		{
			name: "Time range query (acceptable)",
			query: `
				SELECT tenant_id, COUNT(*) as events
				FROM events_distributed 
				WHERE timestamp BETWEEN ? AND ?
				GROUP BY tenant_id
			`,
			sortFields: []models.SortField{
				{Field: "events", Order: models.DESC},
				{Field: "tenant_id", Order: models.ASC},
			},
			whereFields: map[string]interface{}{
				"timestamp": "2024-01-01",
			},
			explanation: "Time-based partition pruning across shards",
		},
	}

	for _, q := range queries {
		fmt.Printf("    %s:\n", q.name)
		fmt.Printf("      %s\n", q.explanation)
		
		optimizedQuery, err := manager.OptimizeDistributedQuery(q.query, q.sortFields, q.whereFields)
		if err != nil {
			fmt.Printf("      Error: %v\n", err)
			continue
		}
		
		fmt.Printf("      Optimized query includes:\n")
		if len(q.whereFields) > 0 {
			for field := range q.whereFields {
				if field == "tenant_id" {
					fmt.Printf("        - Partition pruning for %s\n", field)
				}
			}
		}
		
		for _, sortField := range q.sortFields {
			fmt.Printf("        - Index hint for %s\n", sortField.Field)
		}
		
		fmt.Printf("        - Distributed execution settings\n")
		fmt.Println()
	}
}

// monitorDistributedPerformance shows how to monitor index performance
func monitorDistributedPerformance(ctx context.Context, manager *clickhouse.DistributedIndexManager) {
	fmt.Println("  Simulating query performance monitoring...")

	// Simulate different query patterns and their performance characteristics
	scenarios := []struct {
		name               string
		queryType          string
		estimatedLatency   time.Duration
		indexHitRate       float64
		dataSkippedRatio   float64
		shardsInvolved     int
	}{
		{
			name:               "Tenant-specific query",
			queryType:          "partition_aligned",
			estimatedLatency:   10 * time.Millisecond,
			indexHitRate:       0.95,
			dataSkippedRatio:   0.90,
			shardsInvolved:     1,
		},
		{
			name:               "Event ID lookup",
			queryType:          "bloom_filter",
			estimatedLatency:   25 * time.Millisecond,
			indexHitRate:       0.98,
			dataSkippedRatio:   0.99,
			shardsInvolved:     3,
		},
		{
			name:               "Time range query",
			queryType:          "time_based",
			estimatedLatency:   100 * time.Millisecond,
			indexHitRate:       0.75,
			dataSkippedRatio:   0.60,
			shardsInvolved:     3,
		},
		{
			name:               "Full table scan",
			queryType:          "no_indexes",
			estimatedLatency:   2 * time.Second,
			indexHitRate:       0.10,
			dataSkippedRatio:   0.05,
			shardsInvolved:     3,
		},
	}

	fmt.Println("    Query performance analysis:")
	fmt.Printf("    %-25s %-15s %-12s %-12s %-12s %-8s\n", 
		"Query Type", "Strategy", "Latency", "Index Hit", "Skipped", "Shards")
	fmt.Println("    " + strings.Repeat("-", 85))

	for _, scenario := range scenarios {
		fmt.Printf("    %-25s %-15s %-12v %-12.1f%% %-12.1f%% %-8d\n",
			scenario.name,
			scenario.queryType,
			scenario.estimatedLatency,
			scenario.indexHitRate*100,
			scenario.dataSkippedRatio*100,
			scenario.shardsInvolved,
		)
	}

	// Show recommendations
	fmt.Println("\n    Performance recommendations:")
	fmt.Println("      ✅ Use tenant_id in WHERE clauses for single-shard queries")
	fmt.Println("      ✅ Include timestamp ranges to enable partition pruning")
	fmt.Println("      ✅ Use exact event_id lookups for bloom filter optimization")
	fmt.Println("      ⚠️  Avoid queries without indexed fields")
	fmt.Println("      ⚠️  Consider data locality for frequently queried patterns")
}

// performDistributedMaintenance demonstrates maintenance operations
func performDistributedMaintenance(ctx context.Context, manager *clickhouse.DistributedIndexManager) {
	fmt.Println("  Performing distributed maintenance...")

	// Simulate maintenance operations
	maintenanceTasks := []struct {
		name        string
		description string
		duration    time.Duration
	}{
		{"Index optimization", "Rebuilding indexes on all nodes", 200 * time.Millisecond},
		{"Table optimization", "Optimizing table parts and merges", 150 * time.Millisecond},
		{"Statistics update", "Updating index usage statistics", 50 * time.Millisecond},
		{"Cleanup unused indexes", "Removing unused index structures", 100 * time.Millisecond},
	}

	for _, task := range maintenanceTasks {
		fmt.Printf("    %s...\n", task.description)
		
		// Simulate the maintenance operation
		time.Sleep(task.duration)
		
		// In real implementation, this would call:
		// err := manager.OptimizeDistributedIndexes(ctx, "events")
		
		fmt.Printf("      ✅ %s completed\n", task.name)
	}

	// Show maintenance results
	fmt.Println("\n    Maintenance summary:")
	fmt.Println("      - All node indexes optimized")
	fmt.Println("      - Table parts merged and optimized")
	fmt.Println("      - Index statistics updated")
	fmt.Println("      - Unused structures cleaned up")
	fmt.Println("      - System ready for optimal query performance")
}

// Utility functions for the demo

import "strings"

// simulateQueryExecution shows how a distributed query would execute
func simulateQueryExecution(query string, partitionKey string, indexHints []string) {
	fmt.Printf("    Executing: %s\n", strings.TrimSpace(query))
	
	if partitionKey != "" {
		fmt.Printf("      🎯 Partition pruning: %s\n", partitionKey)
	}
	
	for _, hint := range indexHints {
		fmt.Printf("      📊 Index hint: %s\n", hint)
	}
	
	// Simulate execution time
	executionTime := time.Duration(rand.Intn(100)+10) * time.Millisecond
	fmt.Printf("      ⏱️  Execution time: %v\n", executionTime)
}

// demonstrateDataDistribution shows how data is distributed across shards
func demonstrateDataDistribution() {
	fmt.Println("  Data distribution across shards:")
	
	tenants := []string{"tenant_a", "tenant_b", "tenant_c", "tenant_d", "tenant_e"}
	
	for i, tenant := range tenants {
		shard := (i % 3) + 1
		fmt.Printf("    %s -> Shard %d\n", tenant, shard)
	}
	
	fmt.Println("\n  Query routing examples:")
	fmt.Println("    WHERE tenant_id = 'tenant_a' -> Routes to Shard 1 only")
	fmt.Println("    WHERE event_type = 'error'   -> Queries all shards")
	fmt.Println("    WHERE timestamp > '2024-01'  -> Uses time partitions")
}
