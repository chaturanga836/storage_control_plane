// wal_flusher.go
package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/utils"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/writers"
)

// FlushWAL periodically flushes wal_current.jsonl into Parquet and updates metadata
func FlushWAL(walDir, dataDir, tenantID, sourceID string, flushThreshold int) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		walPath := filepath.Join(walDir, "wal_current.jsonl")
		records, err := ReadWALFile(walPath)
		if err != nil || len(records) < flushThreshold {
			continue // Skip if not enough or file missing
		}

		flat := utils.FlattenJSONSchema(records[0])
		hash := utils.ComputeSchemaHash(flat)
		dir := utils.BuildDirectoryPath(dataDir, tenantID, sourceID, hash, time.Now())
		os.MkdirAll(dir, os.ModePerm)

		filePath, err := writers.WriteParquetFile(records, dir, "auto")
		if err != nil {
			fmt.Println("❌ Parquet write failed:", err)
			continue
		}

		// Save _stats.json (or update ClickHouse)
		meta := writers.FileMetadata{
			TenantID:   tenantID,
			SourceID:   sourceID,
			FilePath:   filePath,
			RowCount:   uint64(len(records)),
			MinTS:      time.Now().Format(time.RFC3339), // (placeholder)
			MaxTS:      time.Now().Format(time.RFC3339), // (placeholder)
			SchemaHash: hash,
			ColumnStats: "{}",
		}

		// NOTE: ClickHouse connection must be passed to this goroutine or refactored
		// writers.InsertParquetMetadata(chConn, meta)

		// Archive old WAL
		archived := strings.Replace(walPath, "wal_current.jsonl", fmt.Sprintf("wal_archived_%d.jsonl", time.Now().Unix()), 1)
		os.Rename(walPath, archived)
	}
}
