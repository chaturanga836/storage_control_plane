package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/api"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/clickhouse"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/config"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/routing"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}
	router := routing.NewRouter(cfg)

	// Check if distributed mode is enabled
	distributedMode := os.Getenv("DISTRIBUTED_MODE")
	
	var server *api.Server
	
	if strings.ToLower(distributedMode) == "true" {
		log.Println("Starting in distributed mode...")
		
		// Setup distributed index manager
		distributedManager, err := setupDistributedManager()
		if err != nil {
			log.Printf("Failed to setup distributed manager: %v", err)
			log.Println("Falling back to single-node mode...")
			server = api.NewServer(router)
		} else {
			log.Println("Distributed index manager initialized successfully")
			server = api.NewDistributedServer(router, distributedManager)
		}
	} else {
		log.Println("Starting in single-node mode...")
		server = api.NewServer(router)
	}

	addr := cfg.ServerAddress
	log.Printf("Starting API server at %s...", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// setupDistributedManager creates and configures the distributed index manager
func setupDistributedManager() (*clickhouse.DistributedIndexManager, error) {
	// Read cluster configuration from environment variables
	clusterName := os.Getenv("CLUSTER_NAME")
	if clusterName == "" {
		clusterName = "analytics_cluster"
	}

	// Parse node configuration from environment
	// Format: "host1:port1:db1:shard1:replica1,host2:port2:db2:shard2:replica2,..."
	nodesConfig := os.Getenv("CLUSTER_NODES")
	if nodesConfig == "" {
		// Default development configuration
		nodesConfig = "localhost:9000:analytics:1:1"
	}

	nodes, err := parseNodeConfig(nodesConfig)
	if err != nil {
		return nil, err
	}

	partitionKey := os.Getenv("PARTITION_KEY")
	if partitionKey == "" {
		partitionKey = "tenant_id"
	}

	shardingKey := os.Getenv("SHARDING_KEY")
	if shardingKey == "" {
		shardingKey = "tenant_id"
	}

	config := clickhouse.DistributedIndexConfig{
		ClusterName:       clusterName,
		Nodes:            nodes,
		ReplicationFactor: 1, // Can be configured via env var
		PartitionKey:     partitionKey,
		ShardingKey:      shardingKey,
		IndexStrategy:    clickhouse.PartitionedIndexes,
	}

	return clickhouse.NewDistributedIndexManager(config)
}

// parseNodeConfig parses node configuration from environment variable
func parseNodeConfig(nodesConfig string) ([]clickhouse.NodeInfo, error) {
	var nodes []clickhouse.NodeInfo
	
	nodeStrings := strings.Split(nodesConfig, ",")
	for i, nodeStr := range nodeStrings {
		parts := strings.Split(strings.TrimSpace(nodeStr), ":")
		if len(parts) < 3 {
			log.Printf("Warning: Invalid node config '%s', using defaults", nodeStr)
			continue
		}

		// Parse port
		port := 9000
		if len(parts) > 1 {
			if p, err := parsePort(parts[1]); err == nil {
				port = p
			}
		}

		// Parse shard and replica
		shard := i + 1
		replica := 1
		if len(parts) > 3 {
			if s, err := parsePort(parts[3]); err == nil {
				shard = s
			}
		}
		if len(parts) > 4 {
			if r, err := parsePort(parts[4]); err == nil {
				replica = r
			}
		}

		node := clickhouse.NodeInfo{
			Host:     parts[0],
			Port:     port,
			Database: parts[2],
			Shard:    shard,
			Replica:  replica,
			Weight:   100,
		}
		
		nodes = append(nodes, node)
	}

	if len(nodes) == 0 {
		// Fallback to single local node
		nodes = append(nodes, clickhouse.NodeInfo{
			Host:     "localhost",
			Port:     9000,
			Database: "analytics",
			Shard:    1,
			Replica:  1,
			Weight:   100,
		})
	}

	return nodes, nil
}

// parsePort safely parses a port number
func parsePort(portStr string) (int, error) {
	port := 9000
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 && p < 65536 {
			port = p
		}
	}
	return port, nil
}