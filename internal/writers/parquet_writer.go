// parquet_writer.go
package writers

import (
	"encoding/json"
	"fmt"
	"time"
	"path/filepath"

	"github.com/xitongsys/parquet-go/writer"
	"github.com/xitongsys/parquet-go-source/local"
)

type ParquetRecord map[string]interface{}

// WriteParquetFile writes a batch of records to a parquet file in the target directory
func WriteParquetFile(records []ParquetRecord, dirPath string, filePrefix string) (string, error) {
	fileName := fmt.Sprintf("%s-%d.parquet", filePrefix, time.Now().UnixNano())
	filePath := filepath.Join(dirPath, fileName)

	fw, err := local.NewLocalFileWriter(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot create local file: %v", err)
	}

	// Use dynamic writer for schema-less data (optional: replace with typed writer)
	pw, err := writer.NewJSONWriter("root", fw, 4)
	if err != nil {
		return "", fmt.Errorf("failed to create parquet writer: %v", err)
	}
	defer pw.WriteStop()
	defer fw.Close()

	for _, rec := range records {
		jsonBytes, _ := json.Marshal(rec)
		if err := pw.Write(string(jsonBytes)); err != nil {
			return "", fmt.Errorf("failed to write record: %v", err)
		}
	}

	return filePath, nil
}
