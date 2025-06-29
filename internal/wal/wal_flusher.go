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

		// Convert records to the correct type
		parquetRecords := make([]writers.ParquetRecord, len(records))
		for i, record := range records {
			parquetRecords[i] = writers.ParquetRecord(record)
		}

		filePath, err := writers.WriteParquetFile(parquetRecords, dir, "auto")
		if err != nil {
			fmt.Println("❌ Parquet write failed:", err)
			continue
		}

		// Save _stats.json (or update ClickHouse)
		meta := writers.FileMetadata{
			TenantID:     tenantID,
			SourceID:     sourceID,
			FilePath:     filePath,
			RowCount:     uint64(len(records)),
			MinTS:        time.Now().Format(time.RFC3339), // (placeholder)
			MaxTS:        time.Now().Format(time.RFC3339), // (placeholder)
			SchemaHash:   hash,
			ColumnStats:  "{}",
		}

		// Use the metadata (write to file or send to ClickHouse)
		if err := writers.WriteMetadata(meta, dir); err != nil {
			fmt.Println("⚠️ Metadata write failed:", err)
		}

		// NOTE: ClickHouse connection must be passed to this goroutine or refactored
		// writers.InsertParquetMetadata(chConn, meta)

		// Archive old WAL
		archived := strings.Replace(walPath, "wal_current.jsonl", fmt.Sprintf("wal_archived_%d.jsonl", time.Now().Unix()), 1)
		os.Rename(walPath, archived)
	}
}

// FlushAllTenants monitors and flushes WAL files for all tenants
func FlushAllTenants(walDir, dataDir string, flushThreshold int) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Discover tenant directories
		tenantDirs, err := filepath.Glob(filepath.Join(walDir, "*"))
		if err != nil {
			continue
		}

		for _, tenantDir := range tenantDirs {
			tenantID := filepath.Base(tenantDir)

			// Discover source directories
			sourceDirs, err := filepath.Glob(filepath.Join(tenantDir, "*"))
			if err != nil {
				continue
			}

			for _, sourceDir := range sourceDirs {
				sourceID := filepath.Base(sourceDir)

				// Check if WAL file exists and has enough records
				walPath := filepath.Join(sourceDir, "wal_current.jsonl")
				records, err := ReadWALFile(walPath)
				if err != nil || len(records) < flushThreshold {
					continue
				}

				fmt.Printf("🔄 Flushing WAL for tenant=%s, source=%s (%d records)\n",
					tenantID, sourceID, len(records))

				// Process the WAL file
				processWALFile(records, walPath, dataDir, tenantID, sourceID)
			}
		}
	}
}

func processWALFile(records []map[string]interface{}, walPath, dataDir, tenantID, sourceID string) {
	flat := utils.FlattenJSONSchema(records[0])
	hash := utils.ComputeSchemaHash(flat)
	dir := utils.BuildDirectoryPath(dataDir, tenantID, sourceID, hash, time.Now())
	os.MkdirAll(dir, os.ModePerm)

	// Convert records to the correct type
	parquetRecords := make([]writers.ParquetRecord, len(records))
	for i, record := range records {
		parquetRecords[i] = writers.ParquetRecord(record)
	}

	filePath, err := writers.WriteParquetFile(parquetRecords, dir, "auto")
	if err != nil {
		fmt.Printf("❌ Parquet write failed for %s/%s: %v\n", tenantID, sourceID, err)
		return
	}

	// Save metadata
	meta := writers.FileMetadata{
		TenantID:     tenantID,
		SourceID:     sourceID,
		FilePath:     filePath,
		RowCount:     uint64(len(records)),
		MinTS:        time.Now().Format(time.RFC3339),
		MaxTS:        time.Now().Format(time.RFC3339),
		SchemaHash:   hash,
		ColumnStats:  "{}",
	}

	if err := writers.WriteMetadata(meta, dir); err != nil {
		fmt.Printf("⚠️ Metadata write failed for %s/%s: %v\n", tenantID, sourceID, err)
	}

	// Archive processed WAL
	archived := strings.Replace(walPath, "wal_current.jsonl",
		fmt.Sprintf("wal_archived_%d.jsonl", time.Now().Unix()), 1)
	os.Rename(walPath, archived)

	fmt.Printf("✅ Flushed %d records from %s/%s to %s\n",
		len(records), tenantID, sourceID, filePath)
}
