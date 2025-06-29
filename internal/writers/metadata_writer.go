// metadata_writer.go
package writers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
)

type FileMetadata struct {
	TenantID    string
	SourceID    string
	FilePath    string
	RowCount    uint64
	MinTS       string // ISO8601
	MaxTS       string
	SchemaHash  string
	ColumnStats string // JSON string
}

func InsertParquetMetadata(conn clickhouse.Conn, meta FileMetadata) error {
	table := fmt.Sprintf("parquet_meta_%s_%s", meta.TenantID, meta.SourceID)

	insert := fmt.Sprintf(`
		INSERT INTO %s (
			id, file_path, row_count, min_ts, max_ts, column_stats, schema_hash, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, now())
	`, table)

	return conn.Exec(context.Background(), insert,
		uuid.New(),
		meta.FilePath,
		meta.RowCount,
		meta.MinTS,
		meta.MaxTS,
		meta.ColumnStats,
		meta.SchemaHash,
	)
}

func UpsertTenantSummary(conn clickhouse.Conn, meta FileMetadata) error {
	upsert := `
		INSERT INTO tenant_summary (
			tenant_id, source_id, total_files, total_rows, last_updated,
			last_file_path, last_min_ts, last_max_ts
		) VALUES (?, ?, ?, ?, now(), ?, ?, ?)
	`

	return conn.Exec(context.Background(), upsert,
		meta.TenantID,
		meta.SourceID,
		uint64(1),
		meta.RowCount,
		meta.FilePath,
		meta.MinTS,
		meta.MaxTS,
	)
}

// WriteMetadata writes metadata to a JSON file
func WriteMetadata(meta FileMetadata, dirPath string) error {
	metaFile := filepath.Join(dirPath, "_stats.json")
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}
	
	return os.WriteFile(metaFile, data, 0644)
}
