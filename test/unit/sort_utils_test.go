package unit

import (
	"testing"
	"github.com/your-org/storage-control-plane/pkg/models"
	"github.com/your-org/storage-control-plane/internal/utils"
)

func TestValidateSortFields(t *testing.T) {
	tests := []struct {
		name        string
		sortFields  []models.SortField
		options     SortOptions
		expectError bool
		expected    []models.SortField
	}{
		{
			name:       "empty sort fields returns default",
			sortFields: []models.SortField{},
			options: SortOptions{
				DefaultField:     "created_at",
				DefaultDirection: models.SortDesc,
				AllowedFields:    []string{"created_at", "name"},
				MaxSortFields:    3,
			},
			expectError: false,
			expected: []models.SortField{{
				Field:     "created_at",
				Direction: models.SortDesc,
			}},
		},
		{
			name: "valid sort fields",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
				{Field: "name", Direction: models.SortAsc},
			},
			options: SortOptions{
				AllowedFields: []string{"created_at", "name", "id"},
				MaxSortFields: 3,
			},
			expectError: false,
			expected: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
				{Field: "name", Direction: models.SortAsc},
			},
		},
		{
			name: "too many sort fields",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
				{Field: "name", Direction: models.SortAsc},
				{Field: "id", Direction: models.SortAsc},
			},
			options: SortOptions{
				AllowedFields: []string{"created_at", "name", "id"},
				MaxSortFields: 2,
			},
			expectError: true,
		},
		{
			name: "invalid field name",
			sortFields: []models.SortField{
				{Field: "invalid_field", Direction: models.SortDesc},
			},
			options: SortOptions{
				AllowedFields: []string{"created_at", "name"},
				MaxSortFields: 3,
			},
			expectError: true,
		},
		{
			name: "duplicate field names",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
				{Field: "created_at", Direction: models.SortAsc},
			},
			options: SortOptions{
				AllowedFields: []string{"created_at", "name"},
				MaxSortFields: 3,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateSortFields(tt.sortFields, tt.options)
			
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
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d fields, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i].Field != expected.Field || result[i].Direction != expected.Direction {
					t.Errorf("field %d: expected %+v, got %+v", i, expected, result[i])
				}
			}
		})
	}
}

func TestGenerateClickHouseOrderBy(t *testing.T) {
	tests := []struct {
		name       string
		sortFields []models.SortField
		expected   string
	}{
		{
			name:       "empty sort fields",
			sortFields: []models.SortField{},
			expected:   "",
		},
		{
			name: "single field ascending",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortAsc},
			},
			expected: "ORDER BY `created_at` ASC",
		},
		{
			name: "single field descending",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
			},
			expected: "ORDER BY `created_at` DESC",
		},
		{
			name: "multiple fields",
			sortFields: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
				{Field: "name", Direction: models.SortAsc},
				{Field: "id", Direction: models.SortDesc},
			},
			expected: "ORDER BY `created_at` DESC, `name` ASC, `id` DESC",
		},
		{
			name: "nested field names",
			sortFields: []models.SortField{
				{Field: "user.profile.name", Direction: models.SortAsc},
			},
			expected: "ORDER BY `user.profile.name` ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateClickHouseOrderBy(tt.sortFields)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBuildSortParams(t *testing.T) {
	tests := []struct {
		name      string
		sortBy    string
		sortOrder string
		expected  []models.SortField
	}{
		{
			name:      "empty parameters",
			sortBy:    "",
			sortOrder: "",
			expected:  []models.SortField{},
		},
		{
			name:      "single field default order",
			sortBy:    "created_at",
			sortOrder: "",
			expected: []models.SortField{
				{Field: "created_at", Direction: models.SortAsc},
			},
		},
		{
			name:      "single field with order",
			sortBy:    "created_at",
			sortOrder: "desc",
			expected: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
			},
		},
		{
			name:      "multiple fields with orders",
			sortBy:    "created_at,name,id",
			sortOrder: "desc,asc,desc",
			expected: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
				{Field: "name", Direction: models.SortAsc},
				{Field: "id", Direction: models.SortDesc},
			},
		},
		{
			name:      "more fields than orders",
			sortBy:    "created_at,name,id",
			sortOrder: "desc,asc",
			expected: []models.SortField{
				{Field: "created_at", Direction: models.SortDesc},
				{Field: "name", Direction: models.SortAsc},
				{Field: "id", Direction: models.SortAsc},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSortParams(tt.sortBy, tt.sortOrder)
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d fields, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if result[i].Field != expected.Field || result[i].Direction != expected.Direction {
					t.Errorf("field %d: expected %+v, got %+v", i, expected, result[i])
				}
			}
		})
	}
}

func TestSanitizeFieldName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple field",
			input:    "created_at",
			expected: "created_at",
		},
		{
			name:     "nested field",
			input:    "user.profile.name",
			expected: "user.profile.name",
		},
		{
			name:     "field with dangerous characters",
			input:    "user'; DROP TABLE users; --",
			expected: "userDROPTABLEusers",
		},
		{
			name:     "field with spaces and special chars",
			input:    "user name@domain.com",
			expected: "usernamedomain.com",
		},
		{
			name:     "alphanumeric with underscores",
			input:    "user_123_name_v2",
			expected: "user_123_name_v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFieldName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
