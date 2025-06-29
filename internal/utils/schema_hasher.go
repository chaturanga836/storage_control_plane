// schema_hasher.go
package utils

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
)

// ComputeSchemaHash generates a stable hash based on flattened JSON structure
func ComputeSchemaHash(flatSchema map[string]string) string {
	jsonBytes, _ := json.Marshal(flatSchema)
	hash := sha1.Sum(jsonBytes)
	return fmt.Sprintf("%x", hash)
} 

// FlattenJSONSchema returns a flat map of key paths to types, e.g. { "user.id": "int64" }
func FlattenJSONSchema(data map[string]interface{}) map[string]string {
	flat := make(map[string]string)
	flattenHelper("", data, flat)
	return flat
}

func flattenHelper(prefix string, data map[string]interface{}, out map[string]string) {
	for k, v := range data {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]interface{}:
			flattenHelper(fullKey, val, out)
		case float64:
			out[fullKey] = "float64"
		case string:
			out[fullKey] = "string"
		case bool:
			out[fullKey] = "bool"
		case []interface{}:
			out[fullKey] = "array"
		default:
			out[fullKey] = "unknown"
		}
	}
}
