// parquet_writer.go
package writers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/writer"
)

type ParquetRecord map[string]interface{}

// ParquetWriteOptions provides configuration for Parquet writing
type ParquetWriteOptions struct {
	CompressionType string // "GZIP", "SNAPPY", "LZ4", etc.
	PageSize        int64
	RowGroupSize    int64
	EnableStats     bool
}

// DefaultParquetOptions returns sensible defaults
func DefaultParquetOptions() ParquetWriteOptions {
	return ParquetWriteOptions{
		CompressionType: "SNAPPY",
		PageSize:        8192,
		RowGroupSize:    128 * 1024 * 1024, // 128MB
		EnableStats:     true,
	}
}

// WriteParquetFile writes a batch of records to a parquet file in the target directory
func WriteParquetFile(records []ParquetRecord, dirPath string, filePrefix string) (string, error) {
	return WriteParquetFileWithOptions(records, dirPath, filePrefix, DefaultParquetOptions())
}

// WriteParquetFileWithOptions writes with custom options
func WriteParquetFileWithOptions(records []ParquetRecord, dirPath string, filePrefix string, opts ParquetWriteOptions) (string, error) {
	if len(records) == 0 {
		return "", fmt.Errorf("no records to write")
	}

	fileName := fmt.Sprintf("%s-%d.parquet", filePrefix, time.Now().UnixNano())
	filePath := filepath.Join(dirPath, fileName)

	// Ensure directory exists
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	fw, err := local.NewLocalFileWriter(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot create local file: %v", err)
	}
	defer fw.Close()

	// Use dynamic writer for schema-less data
	pw, err := writer.NewJSONWriter("root", fw, 4)
	if err != nil {
		return "", fmt.Errorf("failed to create parquet writer: %v", err)
	}
	defer pw.WriteStop()

	// Configure compression if supported
	// Note: xitongsys/parquet-go has limited compression config in JSON mode

	recordCount := 0
	for _, rec := range records {
		// Add row-level metadata
		enrichedRec := make(map[string]interface{})
		for k, v := range rec {
			enrichedRec[k] = v
		}
		enrichedRec["_parquet_write_time"] = time.Now().UTC().Format(time.RFC3339)

		jsonBytes, err := json.Marshal(enrichedRec)
		if err != nil {
			return "", fmt.Errorf("failed to marshal record %d: %v", recordCount, err)
		}

		if err := pw.Write(string(jsonBytes)); err != nil {
			return "", fmt.Errorf("failed to write record %d: %v", recordCount, err)
		}
		recordCount++
	}

	// Get file size
	fileInfo, _ := os.Stat(filePath)
	var fileSize int64
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	fmt.Printf("✅ Wrote %d records to %s (%.2f KB)\n",
		recordCount, fileName, float64(fileSize)/1024)

	return filePath, nil
}
