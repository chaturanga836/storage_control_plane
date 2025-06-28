package migration

import (
	"fmt"
	"hybridstore/internal/rocksdb"
	"hybridstore/internal/clickhouse"
	"hybridstore/pkg/models"
)

// Outline for tenant migration: shared → dedicated
func MigrateTenantData(
	tenantID string,
	sharedRocks *rocksdb.Store,
	sharedCH *clickhouse.Store,
	dedicatedRocks *rocksdb.Store,
	dedicatedCH *clickhouse.Store,
) error {
	// 1. Export all tenant data from shared RocksDB
	bizData, err := sharedRocks.GetBusinessData(tenantID)
	if err != nil {
		return fmt.Errorf("export from shared rocksdb: %w", err)
	}
	// 2. Import into dedicated RocksDB
	for _, d := range bizData {
		err = dedicatedRocks.PutBusinessData(d)
		if err != nil {
			return fmt.Errorf("import to dedicated rocksdb: %w", err)
		}
	}
	// 3. Export all tenant data from shared ClickHouse
	chData, err := sharedCH.GetBusinessData(tenantID)
	if err != nil {
		return fmt.Errorf("export from shared clickhouse: %w", err)
	}
	// 4. Import into dedicated ClickHouse
	for _, d := range chData {
		err = dedicatedCH.PutBusinessData(d)
		if err != nil {
			return fmt.Errorf("import to dedicated clickhouse: %w", err)
		}
	}
	// 5. (Optional) Delete tenant data from shared stores after successful migration
	// TODO: Implement deletion logic

	return nil
}