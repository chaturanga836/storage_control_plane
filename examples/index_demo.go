// index_demo.go - Simple indexing demonstration
package main

import (
	"fmt"
	"strings"

	"github.com/chaturanga836/storage_system/go-control-plane/internal/utils"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

func runIndexDemo() {
	fmt.Println("\n🏗️  Database Indexing Demo")
	fmt.Println("=" + strings.Repeat("=", 30))

	// Show index optimization for sorting
	fmt.Println("\n📊 Index Types and Performance:")
	
	indexInfo := map[string]string{
		"bloom_filter": "Fast tenant_id lookups",
		"minmax":       "Timestamp range queries",
		"set":          "Low cardinality fields",
	}
	
	for indexType, description := range indexInfo {
		fmt.Printf("  • %-12s: %s\n", indexType, description)
	}

	// Show sorting with indexes
	fmt.Println("\n⚡ Efficient Sorting (uses indexes):")
	efficientSorts := []models.SortField{
		{Field: "created_at", Direction: models.SortDesc},
		{Field: "tenant_id", Direction: models.SortAsc},
	}
	
	orderBy := utils.GenerateClickHouseOrderBy(efficientSorts)
	fmt.Printf("  SQL: %s\n", orderBy)
	fmt.Printf("  ✅ Uses idx_created_at and idx_tenant_id\n")

	// Show inefficient sorting
	fmt.Println("\n❌ Inefficient Sorting (no indexes):")
	inefficientSorts := []models.SortField{
		{Field: "total_files", Direction: models.SortDesc},
	}
	
	estimatedRows := int64(1000000)
	_, _, err := utils.ValidateSortFieldsForScale(
		inefficientSorts, utils.TenantSortOptions, utils.DefaultLargeScaleConfig, estimatedRows)
	
	if err != nil {
		fmt.Printf("  ❌ Rejected: %v\n", err)
		fmt.Printf("  💡 System prevents slow queries automatically!\n")
	}
	
	fmt.Println("\n✅ Index demo complete!")
}
