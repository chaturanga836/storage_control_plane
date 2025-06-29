package utils

import (
	"testing"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

func TestLargeScaleSortValidation(t *testing.T) {
	tests := []struct {
		name           string
		sortFields     []models.SortField
		estimatedRows  int64
		expectError    bool
		expectStreaming bool
	}{
		{
			name: "small dataset - standard sorting",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
			},
			estimatedRows:   5000,
			expectError:     false,
			expectStreaming: false,
		},
		{
			name: "large dataset - indexed field",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
			},
			estimatedRows:   500000,
			expectError:     false,
			expectStreaming: true,
		},
		{
			name: "large dataset - non-indexed field with force index",
			sortFields: []models.SortField{
				{Field: "total_files", Direction: models.SortDesc},
			},
			estimatedRows:   500000,
			expectError:     true,
			expectStreaming: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := TenantSortOptions
			config := DefaultLargeScaleConfig
			
			validatedFields, optimizedConfig, err := ValidateSortFieldsForScale(
				tt.sortFields, opts, config, tt.estimatedRows)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if len(validatedFields) != len(tt.sortFields) {
				t.Errorf("expected %d fields, got %d", len(tt.sortFields), len(validatedFields))
			}
			
			if optimizedConfig.UseStreaming != tt.expectStreaming {
				t.Errorf("expected streaming=%v, got %v", tt.expectStreaming, optimizedConfig.UseStreaming)
			}
		})
	}
}

func TestStreamingQueryGeneration(t *testing.T) {
	tests := []struct {
		name       string
		baseQuery  string
		sortFields []models.SortField
		lastValues map[string]any
		chunkSize  int
		expected   string
	}{
		{
			name:      "simple query without cursor",
			baseQuery: "SELECT * FROM table",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
			},
			lastValues: nil,
			chunkSize:  1000,
			expected:   "SELECT * FROM table ORDER BY `created_at` DESC LIMIT 1000",
		},
		{
			name:      "query with cursor pagination",
			baseQuery: "SELECT * FROM table",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
			},
			lastValues: map[string]any{"created_at": "2025-06-29T10:00:00Z"},
			chunkSize:  1000,
			expected:   "SELECT * FROM table WHERE `created_at` < '2025-06-29T10:00:00Z' ORDER BY `created_at` DESC LIMIT 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateStreamingQuery(tt.baseQuery, tt.sortFields, tt.chunkSize, tt.lastValues)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestQueryComplexityEstimation(t *testing.T) {
	tests := []struct {
		name          string
		sortFields    []models.SortField
		estimatedRows int64
		expectedComplexity string
		minRecommendations int
	}{
		{
			name: "small dataset",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
			},
			estimatedRows:      5000,
			expectedComplexity: "LOW",
			minRecommendations: 1,
		},
		{
			name: "medium dataset",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
				{Field: "name", Direction: models.SortAsc},
			},
			estimatedRows:      100000,
			expectedComplexity: "MEDIUM",
			minRecommendations: 2,
		},
		{
			name: "large dataset",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
			},
			estimatedRows:      5000000,
			expectedComplexity: "HIGH",
			minRecommendations: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complexity, recommendations := EstimateQueryComplexity(tt.sortFields, tt.estimatedRows, DefaultLargeScaleConfig)
			
			if complexity != tt.expectedComplexity {
				t.Errorf("expected complexity %s, got %s", tt.expectedComplexity, complexity)
			}
			
			if len(recommendations) < tt.minRecommendations {
				t.Errorf("expected at least %d recommendations, got %d", tt.minRecommendations, len(recommendations))
			}
		})
	}
}

func TestOptimizedQueryGeneration(t *testing.T) {
	sortFields := []models.SortField{
		{Field: "created_at", Direction: models.SortDesc},
		{Field: "name", Direction: models.SortAsc},
	}
	
	config := LargeScaleSortConfig{
		IndexHints: map[string]string{
			"created_at": "idx_created_at",
		},
		UseStreaming: true,
	}
	
	baseQuery := "SELECT * FROM table"
	limit := 100
	offset := 50
	
	result := GenerateOptimizedClickHouseQuery(baseQuery, sortFields, config, limit, offset)
	
	expected := "SELECT * FROM table ORDER BY `created_at` DESC, `name` ASC LIMIT 100 OFFSET 50 /* INDEX_HINT: idx_created_at */"
	
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// Benchmark tests for large-scale operations
func BenchmarkSortValidation(b *testing.B) {
	sortFields := []models.SortField{
		{Field: "created_at", Direction: models.SortDesc},
		{Field: "tenant_id", Direction: models.SortAsc},
		{Field: "total_files", Direction: models.SortDesc},
	}
	
	opts := TenantSortOptions
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValidateSortFields(sortFields, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOrderByGeneration(b *testing.B) {
	sortFields := []models.SortField{
		{Field: "created_at", Direction: models.SortDesc},
		{Field: "tenant_id", Direction: models.SortAsc},
		{Field: "updated_at", Direction: models.SortDesc},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateClickHouseOrderBy(sortFields)
	}
}

func BenchmarkStreamingQueryGeneration(b *testing.B) {
	sortFields := []models.SortField{
		{Field: "timestamp", Direction: models.SortDesc},
	}
	
	lastValues := map[string]any{
		"timestamp": "2025-06-29T10:00:00Z",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateStreamingQuery("SELECT * FROM large_table", sortFields, 10000, lastValues)
		if err != nil {
			b.Fatal(err)
		}
	}
}
