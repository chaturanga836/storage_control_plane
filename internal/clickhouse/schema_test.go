package clickhouse

import (
	"testing"
	"reflect"
)

func TestMapJSONTypeToClickHouseType(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string type", "hello", "String"},
		{"bool type", true, "UInt8"},
		{"integer float", 42.0, "Int64"},
		{"decimal float", 42.5, "Float64"},
		{"map type", map[string]any{"key": "value"}, "String"},
		{"array type", []any{1, 2, 3}, "String"},
		{"nil type", nil, "String"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapJSONTypeToClickHouseType(tt.input)
			if result != tt.expected {
				t.Errorf("mapJSONTypeToClickHouseType(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFlattenJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]string
	}{
		{
			name: "simple flat JSON",
			input: map[string]any{
				"name": "John",
				"age":  30.0,
			},
			expected: map[string]string{
				"name": "String",
				"age":  "Int64",
			},
		},
		{
			name: "nested JSON",
			input: map[string]any{
				"user": map[string]any{
					"name": "John",
					"profile": map[string]any{
						"bio": "Developer",
						"active": true,
					},
				},
				"count": 42.0,
			},
			expected: map[string]string{
				"user.name":           "String",
				"user.profile.bio":    "String",
				"user.profile.active": "UInt8",
				"count":               "Int64",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := map[string]string{}
			FlattenJSON(tt.input, "", result)
			
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FlattenJSON() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateCreateTableSQL(t *testing.T) {
	sample := map[string]any{
		"id":   "123",
		"name": "John",
		"age":  30.0,
		"user": map[string]any{
			"email": "john@example.com",
		},
	}

	result := GenerateCreateTableSQL("tenant1", "conn1", sample)
	
	// Check if table name is correct
	if !contains(result, "tenant_tenant1__conn_conn1") {
		t.Errorf("Table name not found in SQL: %s", result)
	}
	
	// Check if all fields are present
	expectedFields := []string{"`age` Int64", "`id` String", "`name` String", "`user.email` String"}
	for _, field := range expectedFields {
		if !contains(result, field) {
			t.Errorf("Field %s not found in SQL: %s", field, result)
		}
	}
	
	// Check required system fields
	systemFields := []string{"id String", "updated_at DateTime", "version_id String"}
	for _, field := range systemFields {
		if !contains(result, field) {
			t.Errorf("System field %s not found in SQL: %s", field, result)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 len(s) > len(substr) && s[1:len(substr)+1] == substr ||
		 findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
