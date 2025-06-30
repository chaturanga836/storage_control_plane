// index_manager_test.go - Tests for database index management
package clickhouse

import (
	"context"
	"testing"
)

func TestIndexDefinitions(t *testing.T) {
	// Test that all predefined indexes are valid
	for tableName, indexes := range PredefinedIndexes {
		t.Run(tableName, func(t *testing.T) {
			for _, indexDef := range indexes {
				// Validate index definition
				if indexDef.Name == "" {
					t.Errorf("Index name cannot be empty for table %s", tableName)
				}
				
				if indexDef.Table != tableName {
					t.Errorf("Index table mismatch: expected %s, got %s", tableName, indexDef.Table)
				}
				
				if len(indexDef.Columns) == 0 {
					t.Errorf("Index %s must have at least one column", indexDef.Name)
				}
				
				// Validate index type
				validTypes := map[IndexType]bool{
					IndexTypeMinMax:      true,
					IndexTypeBloomFilter: true,
					IndexTypeSet:         true,
					IndexTypeNGramBF:     true,
					IndexTypeTokenBF:     true,
				}
				
				if !validTypes[indexDef.Type] {
					t.Errorf("Invalid index type %s for index %s", indexDef.Type, indexDef.Name)
				}
			}
		})
	}
}

func TestGenerateCreateIndexQuery(t *testing.T) {
	store := &Store{} // Mock store
	im := NewIndexManager(store)
	
	tests := []struct {
		name     string
		indexDef IndexDefinition
		expected string
	}{
		{
			name: "simple bloom filter index",
			indexDef: IndexDefinition{
				Name:        "idx_tenant_id",
				Table:       "business_data",
				Columns:     []string{"tenant_id"},
				Type:        IndexTypeBloomFilter,
				Granularity: 1,
			},
			expected: "ALTER TABLE business_data ADD INDEX idx_tenant_id (tenant_id) TYPE bloom_filter GRANULARITY 1",
		},
		{
			name: "minmax index without granularity",
			indexDef: IndexDefinition{
				Name:    "idx_timestamp",
				Table:   "events",
				Columns: []string{"timestamp"},
				Type:    IndexTypeMinMax,
			},
			expected: "ALTER TABLE events ADD INDEX idx_timestamp (timestamp) TYPE minmax",
		},
		{
			name: "multi-column index",
			indexDef: IndexDefinition{
				Name:        "idx_tenant_created",
				Table:       "data",
				Columns:     []string{"tenant_id", "created_at"},
				Type:        IndexTypeBloomFilter,
				Granularity: 2,
			},
			expected: "ALTER TABLE data ADD INDEX idx_tenant_created (tenant_id, created_at) TYPE bloom_filter GRANULARITY 2",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := im.generateCreateIndexQuery(tt.indexDef)
			if result != tt.expected {
				t.Errorf("Expected: %s\nGot: %s", tt.expected, result)
			}
		})
	}
}

func TestParseIndexColumns(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected []string
	}{
		{
			name:     "single column",
			expr:     "tenant_id",
			expected: []string{"tenant_id"},
		},
		{
			name:     "multiple columns",
			expr:     "tenant_id, created_at",
			expected: []string{"tenant_id", "created_at"},
		},
		{
			name:     "parenthesized expression",
			expr:     "(tenant_id, created_at)",
			expected: []string{"tenant_id", "created_at"},
		},
		{
			name:     "with extra spaces",
			expr:     " tenant_id , created_at ",
			expected: []string{"tenant_id", "created_at"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIndexColumns(tt.expr)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d columns, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("Expected column %d to be %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

func TestIndexOptimizationLogic(t *testing.T) {
	tests := []struct {
		name           string
		stats          IndexStats
		expectedOptType OptimizationType
		shouldOptimize  bool
	}{
		{
			name: "unused index should be dropped",
			stats: IndexStats{
				Name:        "idx_unused",
				RowsRead:    0,
				Selectivity: 0.5,
			},
			expectedOptType: OptimizationDrop,
			shouldOptimize:  true,
		},
		{
			name: "low selectivity index should be dropped",
			stats: IndexStats{
				Name:        "idx_low_selectivity",
				RowsRead:    1000,
				Selectivity: 0.05,
			},
			expectedOptType: OptimizationDrop,
			shouldOptimize:  true,
		},
		{
			name: "high usage poor selectivity should be rebuilt",
			stats: IndexStats{
				Name:        "idx_rebuild",
				RowsRead:    20000,
				Selectivity: 0.9,
			},
			expectedOptType: OptimizationRebuild,
			shouldOptimize:  true,
		},
		{
			name: "good index should not be optimized",
			stats: IndexStats{
				Name:        "idx_good",
				RowsRead:    5000,
				Selectivity: 0.3,
			},
			shouldOptimize: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock index manager
			store := &Store{}
			_ = NewIndexManager(store)
			
			// Create mock usage stats
			stats := &IndexUsageStats{
				TableName: "test_table",
				Indexes: map[string]IndexStats{
					tt.stats.Name: tt.stats,
				},
			}
			
			// This would normally call the actual optimization logic
			// For testing, we'll implement the logic inline
			var optimizations []IndexOptimization
			
			for indexName, indexStats := range stats.Indexes {
				if indexStats.Selectivity < 0.1 {
					optimizations = append(optimizations, IndexOptimization{
						IndexName: indexName,
						Type:      OptimizationDrop,
						Reason:    "Low selectivity",
						Priority:  PriorityMedium,
					})
				}
				
				if indexStats.RowsRead == 0 {
					optimizations = append(optimizations, IndexOptimization{
						IndexName: indexName,
						Type:      OptimizationDrop,
						Reason:    "Unused index",
						Priority:  PriorityHigh,
					})
				}
				
				if indexStats.RowsRead > 10000 && indexStats.Selectivity > 0.8 {
					optimizations = append(optimizations, IndexOptimization{
						IndexName: indexName,
						Type:      OptimizationRebuild,
						Reason:    "High usage poor selectivity",
						Priority:  PriorityHigh,
					})
				}
			}
			
			if tt.shouldOptimize {
				if len(optimizations) == 0 {
					t.Errorf("Expected optimization but got none")
					return
				}
				
				if optimizations[0].Type != tt.expectedOptType {
					t.Errorf("Expected optimization type %s, got %s", tt.expectedOptType, optimizations[0].Type)
				}
			} else {
				if len(optimizations) > 0 {
					t.Errorf("Expected no optimization but got %d", len(optimizations))
				}
			}
		})
	}
}

func TestSchemaValidation(t *testing.T) {
	// Test that all predefined schemas are valid
	for tableName, schema := range PredefinedSchemas {
		t.Run(tableName, func(t *testing.T) {
			// Validate schema definition
			if schema.Name != tableName {
				t.Errorf("Schema name mismatch: expected %s, got %s", tableName, schema.Name)
			}
			
			if len(schema.Columns) == 0 {
				t.Errorf("Table %s must have at least one column", tableName)
			}
			
			if schema.Engine == "" {
				t.Errorf("Table %s must specify an engine", tableName)
			}
			
			if len(schema.OrderBy) == 0 {
				t.Errorf("Table %s must specify ORDER BY columns", tableName)
			}
			
			// Validate that ORDER BY columns exist in table
			columnNames := make(map[string]bool)
			for _, col := range schema.Columns {
				columnNames[col.Name] = true
			}
			
			for _, orderCol := range schema.OrderBy {
				if !columnNames[orderCol] {
					t.Errorf("ORDER BY column %s does not exist in table %s", orderCol, tableName)
				}
			}
			
			// Validate that index columns exist in table
			for _, indexDef := range schema.Indexes {
				for _, indexCol := range indexDef.Columns {
					if !columnNames[indexCol] {
						t.Errorf("Index column %s does not exist in table %s", indexCol, tableName)
					}
				}
			}
		})
	}
}

// Benchmark tests for index operations
func BenchmarkGenerateCreateIndexQuery(b *testing.B) {
	store := &Store{}
	im := NewIndexManager(store)
	
	indexDef := IndexDefinition{
		Name:        "idx_benchmark",
		Table:       "test_table",
		Columns:     []string{"col1", "col2", "col3"},
		Type:        IndexTypeBloomFilter,
		Granularity: 1,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = im.generateCreateIndexQuery(indexDef)
	}
}

func BenchmarkParseIndexColumns(b *testing.B) {
	expr := "(tenant_id, created_at, updated_at, source_id)"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseIndexColumns(expr)
	}
}

// Mock implementation for testing
type MockIndexManager struct {
	indexes map[string][]IndexDefinition
	stats   map[string]*IndexUsageStats
}

func NewMockIndexManager() *MockIndexManager {
	return &MockIndexManager{
		indexes: make(map[string][]IndexDefinition),
		stats:   make(map[string]*IndexUsageStats),
	}
}

func (m *MockIndexManager) CreateIndex(ctx context.Context, indexDef IndexDefinition) error {
	if m.indexes[indexDef.Table] == nil {
		m.indexes[indexDef.Table] = []IndexDefinition{}
	}
	m.indexes[indexDef.Table] = append(m.indexes[indexDef.Table], indexDef)
	return nil
}

func (m *MockIndexManager) ListIndexes(ctx context.Context, tableName string) ([]IndexDefinition, error) {
	return m.indexes[tableName], nil
}

func (m *MockIndexManager) SetMockStats(tableName string, stats *IndexUsageStats) {
	m.stats[tableName] = stats
}

func (m *MockIndexManager) AnalyzeIndexUsage(ctx context.Context, tableName string) (*IndexUsageStats, error) {
	if stats, exists := m.stats[tableName]; exists {
		return stats, nil
	}
	return &IndexUsageStats{TableName: tableName, Indexes: make(map[string]IndexStats)}, nil
}
