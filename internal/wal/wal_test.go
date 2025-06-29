package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestAppendToWAL(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	
	record := map[string]interface{}{
		"id":   "test-123",
		"name": "Test Record",
		"data": map[string]interface{}{
			"value": 42,
		},
	}
	
	err := AppendToWAL(tempDir, record)
	if err != nil {
		t.Fatalf("AppendToWAL failed: %v", err)
	}
	
	// Check if WAL file was created
	walPath := filepath.Join(tempDir, "wal_current.jsonl")
	if _, err := os.Stat(walPath); os.IsNotExist(err) {
		t.Fatal("WAL file was not created")
	}
	
	// Read back the records
	records, err := ReadWALFile(walPath)
	if err != nil {
		t.Fatalf("ReadWALFile failed: %v", err)
	}
	
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}
	
	// Verify record content
	readRecord := records[0]
	if readRecord["id"] != "test-123" {
		t.Errorf("Expected id 'test-123', got %v", readRecord["id"])
	}
}

func TestConcurrentWALWrites(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test concurrent writes
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			record := map[string]interface{}{
				"id":    fmt.Sprintf("record-%d", id),
				"value": id,
			}
			err := AppendToWAL(tempDir, record)
			if err != nil {
				t.Errorf("Concurrent write failed: %v", err)
			}
			done <- true
		}(i)
	}
	
	// Wait for all writes to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all records were written
	walPath := filepath.Join(tempDir, "wal_current.jsonl")
	records, err := ReadWALFile(walPath)
	if err != nil {
		t.Fatalf("ReadWALFile failed: %v", err)
	}
	
	if len(records) != 10 {
		t.Fatalf("Expected 10 records, got %d", len(records))
	}
}
