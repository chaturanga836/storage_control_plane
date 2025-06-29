// directory_planner.go
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BuildDirectoryPath returns the full directory path for a parquet file based on tenant, source, schema, and timestamp
func BuildDirectoryPath(basePath, tenantID, sourceID, schemaHash string, ts time.Time) string {
	return filepath.Join(
		basePath,
		tenantID,
		sourceID,
		"sh_"+schemaHash,
		fmt.Sprintf("year=%d", ts.Year()),
		fmt.Sprintf("month=%02d", ts.Month()),
		fmt.Sprintf("day=%02d", ts.Day()),
	)
}

// EnsureDirectoryExists creates the directory structure if it doesn't exist
func EnsureDirectoryExists(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}
