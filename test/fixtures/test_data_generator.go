package fixtures

import (
	"fmt"
	"time"
	"math/rand"
	
	"github.com/google/uuid"
	"github.com/your-org/storage-control-plane/pkg/models"
)

// TestDataGenerator provides utilities for generating test data
type TestDataGenerator struct {
	rand *rand.Rand
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateSourceConnection creates a test source connection
func (g *TestDataGenerator) GenerateSourceConnection(tenantID string) *models.SourceConnection {
	return &models.SourceConnection{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		Name:     fmt.Sprintf("test-source-%d", g.rand.Intn(1000)),
		Type:     models.SourceTypeAPI,
		Status:   models.ConnectionStatusActive,
		Config: map[string]any{
			"endpoint": fmt.Sprintf("https://api.example.com/v%d", g.rand.Intn(3)+1),
			"timeout":  30,
		},
		CreatedAt: time.Now().Add(-time.Duration(g.rand.Intn(30)) * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-time.Duration(g.rand.Intn(7)) * 24 * time.Hour),
		SyncConfig: models.SourceSyncConfig{
			Enabled:   true,
			Interval:  time.Duration(g.rand.Intn(60)+5) * time.Minute,
			BatchSize: g.rand.Intn(1000) + 100,
		},
		SchemaConfig: models.SourceSchemaConfig{
			AutoDetect:     true,
			StrictMode:     false,
			AllowEvolution: true,
		},
		Metrics: models.SourceMetrics{
			TotalRecords:    int64(g.rand.Intn(1000000)),
			LastSyncStatus:  "success",
			LastSyncTime:    time.Now().Add(-time.Duration(g.rand.Intn(24)) * time.Hour),
			ErrorCount:      int64(g.rand.Intn(5)),
		},
	}
}

// GenerateRecordMetadata creates test record metadata
func (g *TestDataGenerator) GenerateRecordMetadata(tenantID, sourceID, fileID string, count int) []models.RecordMetadata {
	records := make([]models.RecordMetadata, count)
	
	names := []string{"Samuel Johnson", "Alice Smith", "Bob Wilson", "Samantha Davis", "Charlie Brown", "Sam Rodriguez", "Diana Prince", "Sammy Thompson", "Frank Miller"}
	statuses := []string{"active", "inactive", "pending", "suspended"}
	categories := []string{"premium", "standard", "basic", "trial"}
	
	for i := 0; i < count; i++ {
		records[i] = models.RecordMetadata{
			RecordID:  fmt.Sprintf("%s_%d", fileID, i),
			TenantID:  tenantID,
			SourceID:  sourceID,
			FileID:    fileID,
			FilePath:  fmt.Sprintf("/data/%s/%s/2024/01/15/data_%s.parquet", tenantID, sourceID, fileID),
			Name:      names[g.rand.Intn(len(names))],
			Email:     fmt.Sprintf("user%d@example.com", g.rand.Intn(1000)),
			Status:    statuses[g.rand.Intn(len(statuses))],
			Category:  categories[g.rand.Intn(len(categories))],
			CreatedAt: time.Now().Add(-time.Duration(g.rand.Intn(365)) * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-time.Duration(g.rand.Intn(30)) * 24 * time.Hour),
			Timestamp: time.Now().Add(-time.Duration(g.rand.Intn(30)) * 24 * time.Hour),
			RowNumber: int64(i),
			CustomFields: map[string]any{
				"department": []string{"engineering", "marketing", "sales", "support"}[g.rand.Intn(4)],
				"score":      g.rand.Intn(100),
				"verified":   g.rand.Float32() > 0.5,
			},
		}
	}
	
	return records
}

// GenerateParquetFileMetadata creates test file metadata
func (g *TestDataGenerator) GenerateParquetFileMetadata(tenantID, sourceID string) *models.ParquetFileMetadata {
	fileID := uuid.New().String()
	recordCount := int64(g.rand.Intn(10000) + 1000)
	
	return &models.ParquetFileMetadata{
		FileID:        fileID,
		TenantID:      tenantID,
		SourceID:      sourceID,
		FilePath:      fmt.Sprintf("/data/%s/%s/2024/01/15/data_%s.parquet", tenantID, sourceID, fileID[:8]),
		DirectoryPath: fmt.Sprintf("%s/%s/2024/01/15", tenantID, sourceID),
		FileSize:      int64(g.rand.Intn(100000000) + 1000000), // 1MB to 100MB
		RecordCount:   recordCount,
		CreatedAt:     time.Now().Add(-time.Duration(g.rand.Intn(30)) * 24 * time.Hour),
		SchemaHash:    fmt.Sprintf("schema_%x", g.rand.Uint64()),
		IndexedFields: map[string]models.FieldIndex{
			"name": {
				FieldName:    "name",
				FieldType:    "string",
				UniqueValues: []string{"Samuel", "Alice", "Bob", "Samantha"},
				HasNulls:     false,
			},
			"status": {
				FieldName:    "status",
				FieldType:    "string",
				UniqueValues: []string{"active", "inactive", "pending"},
				HasNulls:     false,
			},
		},
		Stats: models.FileStatistics{
			DistinctCounts: map[string]int64{
				"name":   recordCount / 10,
				"status": 4,
				"email":  recordCount,
			},
			NullCounts: map[string]int64{
				"name":     0,
				"email":    g.rand.Int63n(recordCount / 100),
				"category": g.rand.Int63n(recordCount / 50),
			},
		},
	}
}

// GenerateDirectoryConfig creates test directory configuration
func (g *TestDataGenerator) GenerateDirectoryConfig(tenantID string) *models.DirectoryConfig {
	return &models.DirectoryConfig{
		ID:               uuid.New().String(),
		TenantID:         tenantID,
		Name:             fmt.Sprintf("config-%d", g.rand.Intn(1000)),
		Description:      "Test directory configuration",
		DirectoryPattern: "{{.tenant_id}}/{{.source_id}}/{{.year}}/{{.month}}/{{.day}}",
		FileNaming: models.FileNamingConfig{
			Pattern:   "data_{{.timestamp}}_{{.batch_id}}.parquet",
			Timestamp: "20060102_150405",
		},
		TimePartitioning: true,
		PartitionFormat:  "daily",
		Metadata: models.DirectoryMetadataConfig{
			GenerateManifest:    true,
			GenerateStatistics:  true,
			GenerateSchema:      true,
			CompressionLevel:    "SNAPPY",
			MaxRecordsPerFile:   100000,
		},
		CreatedAt: time.Now().Add(-time.Duration(g.rand.Intn(30)) * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-time.Duration(g.rand.Intn(7)) * 24 * time.Hour),
	}
}

// GenerateTestRecords creates raw test record data
func (g *TestDataGenerator) GenerateTestRecords(count int) []map[string]interface{} {
	records := make([]map[string]interface{}, count)
	
	names := []string{"Samuel Johnson", "Alice Smith", "Bob Wilson", "Samantha Davis", "Charlie Brown", "Sam Rodriguez"}
	departments := []string{"engineering", "marketing", "sales", "support"}
	statuses := []string{"active", "inactive", "pending"}
	
	for i := 0; i < count; i++ {
		records[i] = map[string]interface{}{
			"id":         fmt.Sprintf("user_%d", i+1),
			"name":       names[g.rand.Intn(len(names))],
			"email":      fmt.Sprintf("user%d@example.com", i+1),
			"status":     statuses[g.rand.Intn(len(statuses))],
			"department": departments[g.rand.Intn(len(departments))],
			"score":      g.rand.Intn(100),
			"verified":   g.rand.Float32() > 0.5,
			"created_at": time.Now().Add(-time.Duration(g.rand.Intn(365)) * 24 * time.Hour),
			"updated_at": time.Now().Add(-time.Duration(g.rand.Intn(30)) * 24 * time.Hour),
			"timestamp":  time.Now().Add(-time.Duration(g.rand.Intn(30)) * 24 * time.Hour),
		}
	}
	
	return records
}

// Common test constants
const (
	TestTenantID  = "test_tenant_123"
	TestSourceID  = "test_source_456"
	TestSourceID2 = "test_source_789"
)

// Test data sets for specific scenarios
var (
	// TestUsersWithSam contains test data with "sam" in names for search testing
	TestUsersWithSam = []map[string]interface{}{
		{
			"id": "user_001", "name": "Samuel Johnson", "email": "sam.johnson@company.com",
			"status": "active", "department": "marketing", "created_at": time.Now().Add(-10 * time.Hour),
		},
		{
			"id": "user_002", "name": "Samantha Davis", "email": "sam.davis@company.com",
			"status": "active", "department": "engineering", "created_at": time.Now().Add(-7 * time.Hour),
		},
		{
			"id": "user_003", "name": "Sam Rodriguez", "email": "sam.rodriguez@company.com",
			"status": "pending", "department": "engineering", "created_at": time.Now().Add(-5 * time.Hour),
		},
		{
			"id": "user_004", "name": "Sammy Thompson", "email": "sammy.thompson@company.com",
			"status": "active", "department": "sales", "created_at": time.Now().Add(-3 * time.Hour),
		},
	}
)
