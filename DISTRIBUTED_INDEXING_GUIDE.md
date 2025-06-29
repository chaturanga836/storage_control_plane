# Distributed Indexing and Horizontal Scaling Guide

## Overview

This guide explains how indexes work with horizontal nodes (distributed/partitioned systems) and how our storage control plane supports large-scale distributed deployments.

## Index Strategies in Distributed Systems

### 1. Local Indexes (Per-Node Strategy)

**How it works:**
- Each node maintains indexes only for its local data partition
- Query coordinator scans all nodes and merges results
- Best for: Write-heavy workloads, simple queries

**Advantages:**
- Fast writes (no cross-node coordination)
- Simple index maintenance
- Good fault tolerance

**Disadvantages:**
- Slower cross-partition queries
- Requires scatter-gather operations
- Higher query latency

```sql
-- Example: Local indexes on each shard
-- Node 1: CREATE INDEX idx_tenant_local_1 ON events_shard_1 (tenant_id)
-- Node 2: CREATE INDEX idx_tenant_local_2 ON events_shard_2 (tenant_id)
-- Node 3: CREATE INDEX idx_tenant_local_3 ON events_shard_3 (tenant_id)

-- Query hits all nodes:
SELECT * FROM events_distributed WHERE tenant_id = 'tenant_123'
-- Executes on all shards, each uses its local index
```

### 2. Global Indexes (Cross-Node Strategy)

**How it works:**
- Index spans across all nodes, pointing to data locations
- Centralized or distributed index management
- Best for: Read-heavy workloads, complex queries

**Advantages:**
- Fast cross-partition queries
- Single index lookup
- Better query performance

**Disadvantages:**
- Slower writes (must update global index)
- Complex maintenance
- Potential single point of failure

```sql
-- Example: Global index (conceptual)
-- Global Index Table: tenant_id -> [node1:partition1, node2:partition5, ...]
-- Single lookup to find all data locations for a tenant
```

### 3. Partitioned Indexes (Aligned Strategy)

**How it works:**
- Indexes are partitioned using the same strategy as data
- Most common in ClickHouse, Cassandra, BigQuery
- Best for: Balanced read/write workloads

**Advantages:**
- Optimal when queries align with partition key
- Balanced performance
- Scales linearly

**Disadvantages:**
- Poor performance for cross-partition queries
- Requires good partition key design

```sql
-- Example: Partitioned by tenant_id
-- Node 1: tenant_ids 'a'-'h' + indexes for those tenants
-- Node 2: tenant_ids 'i'-'p' + indexes for those tenants  
-- Node 3: tenant_ids 'q'-'z' + indexes for those tenants

-- Efficient query (single node):
SELECT * FROM events WHERE tenant_id = 'abc123' AND timestamp > '2024-01-01'
-- Routes to Node 1, uses local indexes

-- Inefficient query (all nodes):
SELECT * FROM events WHERE timestamp > '2024-01-01'
-- Must scan all nodes
```

## ClickHouse Distributed Architecture

### ClickHouse Cluster Setup

```xml
<!-- /etc/clickhouse-server/config.d/clusters.xml -->
<remote_servers>
    <storage_cluster>
        <shard>
            <replica>
                <host>node1.example.com</host>
                <port>9000</port>
            </replica>
            <replica>
                <host>node1-replica.example.com</host>
                <port>9000</port>
            </replica>
        </shard>
        <shard>
            <replica>
                <host>node2.example.com</host>
                <port>9000</port>
            </replica>
            <replica>
                <host>node2-replica.example.com</host>
                <port>9000</port>
            </replica>
        </shard>
        <shard>
            <replica>
                <host>node3.example.com</host>
                <port>9000</port>
            </replica>
            <replica>
                <host>node3-replica.example.com</host>
                <port>9000</port>
            </replica>
        </shard>
    </storage_cluster>
</remote_servers>
```

### Local Tables with Indexes

```sql
-- On each node, create local table with indexes
CREATE TABLE events_local ON CLUSTER storage_cluster (
    tenant_id String,
    event_id String,
    timestamp DateTime64(3),
    data String,
    created_at DateTime DEFAULT now()
) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/events', '{replica}')
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, timestamp)
SETTINGS index_granularity = 8192;

-- Add indexes on each local table
ALTER TABLE events_local ON CLUSTER storage_cluster 
ADD INDEX idx_tenant_minmax tenant_id TYPE minmax GRANULARITY 1;

ALTER TABLE events_local ON CLUSTER storage_cluster 
ADD INDEX idx_timestamp_minmax timestamp TYPE minmax GRANULARITY 1;

ALTER TABLE events_local ON CLUSTER storage_cluster 
ADD INDEX idx_event_bloom event_id TYPE bloom_filter GRANULARITY 1;
```

### Distributed Table

```sql
-- Create distributed table that aggregates all local tables
CREATE TABLE events_distributed ON CLUSTER storage_cluster AS events_local
ENGINE = Distributed(storage_cluster, currentDatabase(), events_local, sipHash64(tenant_id));
```

## Using Our Distributed Index Manager

### 1. Initialize Distributed Manager

```go
package main

import (
    "context"
    "log"
    
    "github.com/chaturanga836/storage_system/go-control-plane/internal/clickhouse"
)

func main() {
    config := clickhouse.DistributedIndexConfig{
        ClusterName: "storage_cluster",
        Nodes: []clickhouse.NodeInfo{
            {Host: "node1.example.com", Port: 9000, Database: "analytics", Shard: 1, Replica: 1},
            {Host: "node2.example.com", Port: 9000, Database: "analytics", Shard: 2, Replica: 1},
            {Host: "node3.example.com", Port: 9000, Database: "analytics", Shard: 3, Replica: 1},
        },
        ReplicationFactor: 2,
        PartitionKey: "tenant_id",
        ShardingKey: "tenant_id",
        IndexStrategy: clickhouse.PartitionedIndexes,
    }

    manager, err := clickhouse.NewDistributedIndexManager(config)
    if err != nil {
        log.Fatal(err)
    }
    defer manager.Close()

    ctx := context.Background()
    
    // Create distributed indexes
    err = manager.CreateDistributedIndex(ctx, "events", "idx_tenant_performance", 
        []string{"tenant_id", "timestamp"}, clickhouse.IndexTypeMinMax)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 2. Optimize Queries for Distribution

```go
// Original query
baseQuery := `
    SELECT tenant_id, COUNT(*) as event_count, AVG(processing_time) as avg_time
    FROM events_distributed 
    WHERE timestamp >= ? AND timestamp <= ?
    GROUP BY tenant_id
    ORDER BY event_count DESC
`

// Add distributed optimizations
sortFields := []models.SortField{
    {Field: "event_count", Order: models.DESC},
    {Field: "tenant_id", Order: models.ASC},
}

whereConditions := map[string]interface{}{
    "timestamp": "2024-01-01",
    "tenant_id": "high_priority_tenant",
}

optimizedQuery, err := manager.OptimizeDistributedQuery(baseQuery, sortFields, whereConditions)
// Result includes index hints, partition pruning, and distributed settings
```

### 3. Monitor Distributed Index Performance

```go
// Get index information across all nodes
indexInfo, err := manager.GetDistributedIndexInfo(ctx, "events")
if err != nil {
    log.Fatal(err)
}

for nodeIndex, indexes := range indexInfo {
    fmt.Printf("Node %d indexes:\n", nodeIndex)
    for _, idx := range indexes {
        fmt.Printf("  %s (%s): %s\n", idx.Name, idx.Type, idx.Expression)
    }
}

// Get cluster topology
topology, err := manager.GetClusterTopology(ctx)
if err != nil {
    log.Fatal(err)
}

for _, node := range topology {
    fmt.Printf("Shard %d, Replica %d: %s:%d (weight: %d)\n", 
        node.Shard, node.Replica, node.Host, node.Port, node.Weight)
}
```

## Performance Optimization Strategies

### 1. Partition Pruning

```sql
-- Good: Query includes partition key
SELECT * FROM events_distributed 
WHERE tenant_id = 'specific_tenant' AND timestamp >= '2024-01-01'
-- Only queries nodes containing this tenant's data

-- Bad: Query without partition key
SELECT * FROM events_distributed 
WHERE event_type = 'error' AND timestamp >= '2024-01-01'
-- Must query all nodes
```

### 2. Index Hints and Query Planning

```sql
-- Our system automatically adds hints like:
SELECT * FROM events_distributed 
WHERE tenant_id = 'tenant_123' 
ORDER BY timestamp DESC
/* PARTITION_PRUNE: tenant_id = tenant_123, USE_INDEX: idx_tenant_id */
SETTINGS 
    distributed_product_mode = 'global',
    distributed_aggregation_memory_efficient = 1,
    prefer_localhost_replica = 1
```

### 3. Parallel Processing

```go
// Execute queries in parallel across shards
func (d *DistributedIndexManager) ExecuteParallelQuery(ctx context.Context, query string) ([][]interface{}, error) {
    var wg sync.WaitGroup
    results := make(chan [][]interface{}, len(d.nodes))
    errors := make(chan error, len(d.nodes))

    for _, node := range d.nodes {
        wg.Add(1)
        go func(conn clickhouse.Conn) {
            defer wg.Done()
            
            rows, err := conn.Query(ctx, query)
            if err != nil {
                errors <- err
                return
            }
            defer rows.Close()

            var nodeResults [][]interface{}
            for rows.Next() {
                var row []interface{}
                if err := rows.Scan(&row); err != nil {
                    errors <- err
                    return
                }
                nodeResults = append(nodeResults, row)
            }
            results <- nodeResults
        }(node)
    }

    wg.Wait()
    close(results)
    close(errors)

    // Check for errors
    for err := range errors {
        return nil, err
    }

    // Merge results from all nodes
    var allResults [][]interface{}
    for nodeResults := range results {
        allResults = append(allResults, nodeResults...)
    }

    return allResults, nil
}
```

## Best Practices for Distributed Indexes

### 1. Choose the Right Partition Key

```go
// Good partition keys:
// - High cardinality (many unique values)
// - Evenly distributed
// - Commonly used in WHERE clauses

// Examples:
// ✅ tenant_id (if tenants are evenly sized)
// ✅ toYYYYMM(timestamp) (time-based partitioning)
// ✅ sipHash64(user_id) % 100 (hash-based)

// ❌ country (low cardinality, uneven distribution)
// ❌ is_active (boolean, only 2 values)
```

### 2. Index Selection Strategy

```sql
-- Primary indexes (ORDER BY) - automatically created
ORDER BY (tenant_id, timestamp, event_id)

-- Secondary indexes - choose carefully
ADD INDEX idx_tenant_minmax tenant_id TYPE minmax GRANULARITY 1;     -- Range queries
ADD INDEX idx_event_bloom event_id TYPE bloom_filter GRANULARITY 1;  -- Exact matches
ADD INDEX idx_status_set status TYPE set(100) GRANULARITY 1;         -- Limited values
```

### 3. Query Patterns

```sql
-- Optimal: Single shard query
SELECT * FROM events_distributed 
WHERE tenant_id = 'tenant_123' AND timestamp >= '2024-01-01'
-- Performance: Excellent (1 shard, local indexes)

-- Good: Multi-shard with selectivity
SELECT * FROM events_distributed 
WHERE event_id = 'specific_event' 
-- Performance: Good (bloom filter on all shards)

-- Acceptable: Time-based with partition pruning
SELECT * FROM events_distributed 
WHERE timestamp BETWEEN '2024-01-01' AND '2024-01-02'
-- Performance: OK (time partitions reduce scope)

-- Poor: Full table scan
SELECT * FROM events_distributed 
WHERE data LIKE '%error%'
-- Performance: Poor (must scan all data on all shards)
```

## Monitoring and Maintenance

### 1. Index Usage Monitoring

```sql
-- Monitor index usage
SELECT 
    table,
    name,
    type,
    expr,
    rows_read,
    bytes_read
FROM system.query_log 
WHERE type = 'QueryFinish' 
AND used_data_skipping_indices != ''
ORDER BY event_time DESC;
```

### 2. Performance Metrics

```go
// Monitor distributed query performance
type DistributedQueryMetrics struct {
    NodeLatencies    map[int]time.Duration
    IndexHitRates    map[int]float64
    DataSkippedRatio map[int]float64
    TotalRows        int64
    ProcessedRows    int64
}

func (d *DistributedIndexManager) GetQueryMetrics(ctx context.Context, queryID string) (*DistributedQueryMetrics, error) {
    // Implementation to collect metrics from all nodes
    // Returns performance data for optimization
}
```

### 3. Automated Optimization

```go
// Periodic index maintenance
func (d *DistributedIndexManager) PerformMaintenance(ctx context.Context) error {
    // 1. Optimize tables on all nodes
    if err := d.OptimizeDistributedIndexes(ctx, "events"); err != nil {
        return err
    }
    
    // 2. Analyze query patterns and suggest new indexes
    suggestions, err := d.AnalyzeQueryPatterns(ctx)
    if err != nil {
        return err
    }
    
    // 3. Remove unused indexes
    return d.CleanupUnusedIndexes(ctx, suggestions.UnusedIndexes)
}
```

## Conclusion

Our distributed index manager provides:

1. **Automatic index distribution** across cluster nodes
2. **Query optimization** with partition pruning and index hints  
3. **Parallel execution** for better performance
4. **Monitoring and maintenance** tools
5. **Flexible strategies** for different workload patterns

This enables our storage control plane to scale horizontally while maintaining excellent query performance through intelligent index management.
