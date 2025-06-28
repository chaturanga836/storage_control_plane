package routing

import (
	"database/sql"
	"errors"
	"fmt"
	"hybridstore/internal/config"
	"hybridstore/internal/rocksdb"
	"hybridstore/internal/clickhouse"
	"hybridstore/pkg/models"
	_ "github.com/lib/pq"
)

type Router struct {
	cfg         *config.Config
	pg          *sql.DB
	sharedRocks *rocksdb.Store
	sharedCH    *clickhouse.Store
}

func NewRouter(cfg *config.Config) *Router {
	// Open Postgres and shared data stores
	pg, err := sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		panic(fmt.Sprintf("postgres open: %v", err))
	}
	sharedRocks, err := rocksdb.NewStore(cfg.SharedRocksDBPath)
	if err != nil {
		panic(fmt.Sprintf("shared rocksdb: %v", err))
	}
	sharedCH, err := clickhouse.NewStore(cfg.SharedClickHouseDSN)
	if err != nil {
		panic(fmt.Sprintf("shared clickhouse: %v", err))
	}
	return &Router{
		cfg:         cfg,
		pg:          pg,
		sharedRocks: sharedRocks,
		sharedCH:    sharedCH,
	}
}

// Backend holds the endpoints for a tenant's data stores
type Backend struct {
	RocksDB   models.DataStore
	ClickHouse models.DataStore
}

// LookupBackend returns the correct backend for a tenant (shared or dedicated)
func (r *Router) LookupBackend(tenantID string) (*Backend, error) {
	var isDedicated bool
	var rocksdbDSN, clickhouseDSN string

	err := r.pg.QueryRow(`
		SELECT is_dedicated, rocksdb_endpoint, clickhouse_endpoint
		FROM tenant_routing
		WHERE tenant_id = $1
	`, tenantID).Scan(&isDedicated, &rocksdbDSN, &clickhouseDSN)
	if err != nil {
		return nil, err
	}

	if !isDedicated {
		return &Backend{
			RocksDB:   r.sharedRocks,
			ClickHouse: r.sharedCH,
		}, nil
	}

	dedicatedRocks, err := rocksdb.NewStore(rocksdbDSN)
	if err != nil {
		return nil, fmt.Errorf("dedicated rocksdb: %w", err)
	}
	dedicatedCH, err := clickhouse.NewStore(clickhouseDSN)
	if err != nil {
		return nil, fmt.Errorf("dedicated clickhouse: %w", err)
	}
	return &Backend{
		RocksDB:   dedicatedRocks,
		ClickHouse: dedicatedCH,
	}, nil
}