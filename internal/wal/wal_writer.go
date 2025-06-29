// wal_writer.go
package wal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var walMu sync.Mutex

// AppendToWAL writes a JSON record to the WAL file for durability
func AppendToWAL(walDir string, record map[string]interface{}) error {
	walMu.Lock()
	defer walMu.Unlock()

	if err := os.MkdirAll(walDir, os.ModePerm); err != nil {
		return err
	}
	
	filePath := filepath.Join(walDir, "wal_current.jsonl")
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open WAL file: %v", err)
	}
	defer f.Close()

	timestamped := map[string]interface{}{
		"ts":   time.Now().UTC().Format(time.RFC3339),
		"data": record,
	}

	line, _ := json.Marshal(timestamped)
	_, err = f.WriteString(string(line) + "\n")
	return err
}

// ReadWALFile reads all records from WAL for replaying
func ReadWALFile(filePath string) ([]map[string]interface{}, error) {
	var records []map[string]interface{}
	f, err := os.Open(filePath)
	if err != nil {
		return records, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			if data, ok := entry["data"].(map[string]interface{}); ok {
				records = append(records, data)
			}
		}
	}
	return records, nil
}
