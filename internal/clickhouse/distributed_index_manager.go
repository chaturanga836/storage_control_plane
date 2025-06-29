package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

// IndexInfo represents information about a database index
type IndexInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Expression  string `json:"expression"`
	Granularity int    `json:"granularity"`
}

// DistributedIndexManager handles index operations across multiple ClickHouse nodes
type DistributedIndexManager struct {
	nodes          []clickhouse.Conn
	clusterName    string
	replicationFactor int
	partitionKey   string
	mu            sync.RWMutex
}

// NodeInfo represents information about a ClickHouse node
type NodeInfo struct {
	Host     string
	Port     int
	Database string
	Shard    int
	Replica  int
	Weight   int
}

// DistributedIndexConfig configuration for distributed index management
type DistributedIndexConfig struct {
	ClusterName       string
	Nodes            []NodeInfo
	ReplicationFactor int
	PartitionKey     string
	ShardingKey      string
	IndexStrategy    IndexStrategy
}

// IndexStrategy defines how indexes are distributed
type IndexStrategy string

const (
	LocalIndexes      IndexStrategy = "local"      // Indexes on each node independently
	GlobalIndexes     IndexStrategy = "global"     // Centralized index management
	PartitionedIndexes IndexStrategy = "partitioned" // Indexes follow data partitioning
)

// NewDistributedIndexManager creates a new distributed index manager
func NewDistributedIndexManager(config DistributedIndexConfig) (*DistributedIndexManager, error) {
	manager := &DistributedIndexManager{
		nodes:           make([]clickhouse.Conn, 0, len(config.Nodes)),
		clusterName:     config.ClusterName,
		replicationFactor: config.ReplicationFactor,
		partitionKey:    config.PartitionKey,
	}

	// Connect to all nodes
	for _, node := range config.Nodes {
		conn, err := clickhouse.Open(&clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%d", node.Host, node.Port)},
			Auth: clickhouse.Auth{
				Database: node.Database,
			},
			Settings: clickhouse.Settings{
				"max_execution_time": 60,
			},
			DialTimeout: time.Second * 30,
			Compression: &clickhouse.Compression{
				Method: clickhouse.CompressionLZ4,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to node %s:%d: %w", node.Host, node.Port, err)
		}
		manager.nodes = append(manager.nodes, conn)
	}

	return manager, nil
}

// CreateDistributedIndex creates indexes across all nodes in the cluster
func (d *DistributedIndexManager) CreateDistributedIndex(ctx context.Context, tableName, indexName string, columns []string, indexType IndexType) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Create index on local tables first
	localTableName := fmt.Sprintf("%s_local", tableName)
	
	var wg sync.WaitGroup
	errors := make(chan error, len(d.nodes))

	for i, node := range d.nodes {
		wg.Add(1)
		go func(nodeIndex int, conn clickhouse.Conn) {
			defer wg.Done()
			
			// Create index on local table for this shard
			query := d.generateIndexQuery(localTableName, indexName, columns, indexType, nodeIndex)
			
			if err := conn.Exec(ctx, query); err != nil {
				errors <- fmt.Errorf("failed to create index on node %d: %w", nodeIndex, err)
				return
			}

			// Log successful index creation
			fmt.Printf("Index %s created on node %d for table %s\n", indexName, nodeIndex, localTableName)
		}(i, node)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		return err
	}

	// Create distributed table view that uses the indexed local tables
	return d.createDistributedTableView(ctx, tableName, localTableName)
}

// generateIndexQuery creates the appropriate index query for a specific node
func (d *DistributedIndexManager) generateIndexQuery(tableName, indexName string, columns []string, indexType IndexType, nodeIndex int) string {
	// Add node-specific suffix to index name to avoid conflicts
	nodeIndexName := fmt.Sprintf("%s_node_%d", indexName, nodeIndex)
	
	switch indexType {
	case IndexTypeMinMax:
		return fmt.Sprintf("ALTER TABLE %s ADD INDEX %s (%s) TYPE minmax GRANULARITY 1", 
			tableName, nodeIndexName, strings.Join(columns, ", "))
	case IndexTypeBloomFilter:
		return fmt.Sprintf("ALTER TABLE %s ADD INDEX %s (%s) TYPE bloom_filter GRANULARITY 1", 
			tableName, nodeIndexName, strings.Join(columns, ", "))
	case IndexTypeSet:
		return fmt.Sprintf("ALTER TABLE %s ADD INDEX %s (%s) TYPE set(1000) GRANULARITY 1", 
			tableName, nodeIndexName, strings.Join(columns, ", "))
	default:
		return fmt.Sprintf("ALTER TABLE %s ADD INDEX %s (%s) TYPE minmax GRANULARITY 1", 
			tableName, nodeIndexName, strings.Join(columns, ", "))
	}
}

// createDistributedTableView creates a distributed table that aggregates local tables
func (d *DistributedIndexManager) createDistributedTableView(ctx context.Context, distTableName, localTableName string) error {
	// This creates a distributed table that queries all local tables
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s AS %s_local
		ENGINE = Distributed(%s, currentDatabase(), %s, rand())
	`, distTableName, distTableName, d.clusterName, localTableName)

	// Execute on the first node (coordinator)
	return d.nodes[0].Exec(ctx, query)
}

// OptimizeDistributedQuery optimizes a query for distributed execution with indexes
func (d *DistributedIndexManager) OptimizeDistributedQuery(
	baseQuery string, 
	sortFields []models.SortField, 
	whereConditions map[string]interface{},
) (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var optimizations []string

	// 1. Add partition pruning hints
	if partitionHint := d.generatePartitionPruningHint(whereConditions); partitionHint != "" {
		optimizations = append(optimizations, partitionHint)
	}

	// 2. Add index hints for sorting fields
	for _, field := range sortFields {
		if indexHint := d.generateIndexHint(field.Field); indexHint != "" {
			optimizations = append(optimizations, indexHint)
		}
	}

	// 3. Add distributed query settings
	distributedSettings := []string{
		"SETTINGS distributed_product_mode = 'global'",
		"distributed_aggregation_memory_efficient = 1",
		"prefer_localhost_replica = 1",
	}

	// Combine base query with optimizations
	optimizedQuery := baseQuery
	if len(optimizations) > 0 {
		optimizedQuery += " /* " + strings.Join(optimizations, ", ") + " */"
	}
	optimizedQuery += " " + strings.Join(distributedSettings, ", ")

	return optimizedQuery, nil
}

// generatePartitionPruningHint creates hints to eliminate unnecessary partitions/shards
func (d *DistributedIndexManager) generatePartitionPruningHint(whereConditions map[string]interface{}) string {
	if d.partitionKey == "" {
		return ""
	}

	// Check if partition key is in where conditions
	if value, exists := whereConditions[d.partitionKey]; exists {
		return fmt.Sprintf("PARTITION_PRUNE: %s = %v", d.partitionKey, value)
	}

	return ""
}

// generateIndexHint creates index hints for specific fields
func (d *DistributedIndexManager) generateIndexHint(fieldName string) string {
	// Common indexed fields in our system
	indexedFields := map[string]string{
		"tenant_id":   "idx_tenant_id",
		"created_at":  "idx_created_at",
		"updated_at":  "idx_updated_at",
		"file_path":   "idx_file_path",
		"timestamp":   "idx_timestamp",
	}

	if indexName, exists := indexedFields[fieldName]; exists {
		return fmt.Sprintf("USE_INDEX: %s", indexName)
	}

	return ""
}

// GetDistributedIndexInfo returns information about indexes across all nodes
func (d *DistributedIndexManager) GetDistributedIndexInfo(ctx context.Context, tableName string) (map[int][]IndexInfo, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[int][]IndexInfo)
	var wg sync.WaitGroup
	mu := sync.Mutex{}

	for i, node := range d.nodes {
		wg.Add(1)
		go func(nodeIndex int, conn clickhouse.Conn) {
			defer wg.Done()

			query := `
				SELECT 
					name, 
					type, 
					expr,
					granularity
				FROM system.data_skipping_indices 
				WHERE table = ? AND database = currentDatabase()
			`

			rows, err := conn.Query(ctx, query, fmt.Sprintf("%s_local", tableName))
			if err != nil {
				return
			}
			defer rows.Close()

			var nodeIndexes []IndexInfo
			for rows.Next() {
				var info IndexInfo
				if err := rows.Scan(&info.Name, &info.Type, &info.Expression, &info.Granularity); err != nil {
					continue
				}
				nodeIndexes = append(nodeIndexes, info)
			}

			mu.Lock()
			result[nodeIndex] = nodeIndexes
			mu.Unlock()
		}(i, node)
	}

	wg.Wait()
	return result, nil
}

// DropDistributedIndex removes indexes from all nodes
func (d *DistributedIndexManager) DropDistributedIndex(ctx context.Context, tableName, indexName string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var wg sync.WaitGroup
	errors := make(chan error, len(d.nodes))

	for i, node := range d.nodes {
		wg.Add(1)
		go func(nodeIndex int, conn clickhouse.Conn) {
			defer wg.Done()

			nodeIndexName := fmt.Sprintf("%s_node_%d", indexName, nodeIndex)
			localTableName := fmt.Sprintf("%s_local", tableName)
			
			query := fmt.Sprintf("ALTER TABLE %s DROP INDEX IF EXISTS %s", localTableName, nodeIndexName)
			
			if err := conn.Exec(ctx, query); err != nil {
				errors <- fmt.Errorf("failed to drop index on node %d: %w", nodeIndex, err)
			}
		}(i, node)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		return err
	}

	return nil
}

// OptimizeDistributedIndexes performs maintenance on indexes across all nodes
func (d *DistributedIndexManager) OptimizeDistributedIndexes(ctx context.Context, tableName string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var wg sync.WaitGroup
	errors := make(chan error, len(d.nodes))

	for i, node := range d.nodes {
		wg.Add(1)
		go func(nodeIndex int, conn clickhouse.Conn) {
			defer wg.Done()

			localTableName := fmt.Sprintf("%s_local", tableName)
			
			// Optimize table and rebuild indexes
			queries := []string{
				fmt.Sprintf("OPTIMIZE TABLE %s FINAL", localTableName),
				fmt.Sprintf("ALTER TABLE %s MATERIALIZE INDEX", localTableName),
			}

			for _, query := range queries {
				if err := conn.Exec(ctx, query); err != nil {
					errors <- fmt.Errorf("failed to optimize on node %d: %w", nodeIndex, err)
					return
				}
			}
		}(i, node)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		return err
	}

	return nil
}

// Close closes all node connections
func (d *DistributedIndexManager) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, node := range d.nodes {
		if err := node.Close(); err != nil {
			return fmt.Errorf("failed to close node %d: %w", i, err)
		}
	}

	return nil
}

// GetClusterTopology returns information about the cluster topology
func (d *DistributedIndexManager) GetClusterTopology(ctx context.Context) ([]NodeInfo, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `
		SELECT 
			host_name,
			port,
			shard_num,
			replica_num,
			shard_weight
		FROM system.clusters 
		WHERE cluster = ?
	`

	rows, err := d.nodes[0].Query(ctx, query, d.clusterName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topology []NodeInfo
	for rows.Next() {
		var node NodeInfo
		if err := rows.Scan(&node.Host, &node.Port, &node.Shard, &node.Replica, &node.Weight); err != nil {
			continue
		}
		topology = append(topology, node)
	}

	return topology, nil
}
