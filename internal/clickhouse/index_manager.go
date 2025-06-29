// index_manager.go - Manages database indexes for optimal sorting performance
package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// IndexManager handles database index creation and management
type IndexManager struct {
	store *Store
}

// NewIndexManager creates a new index manager
func NewIndexManager(store *Store) *IndexManager {
	return &IndexManager{store: store}
}

// IndexDefinition defines a database index
type IndexDefinition struct {
	Name        string            `json:"name"`
	Table       string            `json:"table"`
	Columns     []string          `json:"columns"`
	Type        IndexType         `json:"type"`
	Granularity int               `json:"granularity,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	Status      IndexStatus       `json:"status"`
}

type IndexType string

const (
	IndexTypeMinMax      IndexType = "minmax"
	IndexTypeBloomFilter IndexType = "bloom_filter"
	IndexTypeSet         IndexType = "set"
	IndexTypeNGramBF     IndexType = "ngrambf_v1"
	IndexTypeTokenBF     IndexType = "tokenbf_v1"
)

type IndexStatus string

const (
	IndexStatusActive   IndexStatus = "active"
	IndexStatusCreating IndexStatus = "creating"
	IndexStatusFailed   IndexStatus = "failed"
	IndexStatusDropped  IndexStatus = "dropped"
)

// PredefinedIndexes contains recommended indexes for different table types
var PredefinedIndexes = map[string][]IndexDefinition{
	"business_data": {
		{
			Name:        "idx_tenant_id",
			Table:       "business_data",
			Columns:     []string{"tenant_id"},
			Type:        IndexTypeBloomFilter,
			Granularity: 1,
		},
		{
			Name:        "idx_created_at",
			Table:       "business_data",
			Columns:     []string{"created_at"},
			Type:        IndexTypeMinMax,
			Granularity: 1,
		},
		{
			Name:        "idx_data_id",
			Table:       "business_data",
			Columns:     []string{"data_id"},
			Type:        IndexTypeBloomFilter,
			Granularity: 1,
		},
	},
	"analytics_events": {
		{
			Name:        "idx_timestamp",
			Table:       "analytics_events",
			Columns:     []string{"timestamp"},
			Type:        IndexTypeMinMax,
			Granularity: 1,
		},
		{
			Name:        "idx_tenant_id",
			Table:       "analytics_events",
			Columns:     []string{"tenant_id"},
			Type:        IndexTypeBloomFilter,
			Granularity: 1,
		},
		{
			Name:        "idx_event_id",
			Table:       "analytics_events",
			Columns:     []string{"event_id"},
			Type:        IndexTypeBloomFilter,
			Granularity: 1,
		},
	},
	"parquet_files": {
		{
			Name:        "idx_created_at",
			Table:       "parquet_files",
			Columns:     []string{"created_at"},
			Type:        IndexTypeMinMax,
			Granularity: 1,
		},
		{
			Name:        "idx_min_timestamp",
			Table:       "parquet_files",
			Columns:     []string{"min_timestamp"},
			Type:        IndexTypeMinMax,
			Granularity: 1,
		},
		{
			Name:        "idx_max_timestamp",
			Table:       "parquet_files",
			Columns:     []string{"max_timestamp"},
			Type:        IndexTypeMinMax,
			Granularity: 1,
		},
		{
			Name:        "idx_tenant_id",
			Table:       "parquet_files",
			Columns:     []string{"tenant_id"},
			Type:        IndexTypeBloomFilter,
			Granularity: 1,
		},
	},
}

// CreateIndex creates a new database index
func (im *IndexManager) CreateIndex(ctx context.Context, indexDef IndexDefinition) error {
	query := im.generateCreateIndexQuery(indexDef)
	
	_, err := im.store.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create index %s: %w", indexDef.Name, err)
	}
	
	return nil
}

// DropIndex drops a database index
func (im *IndexManager) DropIndex(ctx context.Context, tableName, indexName string) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP INDEX %s", tableName, indexName)
	
	_, err := im.store.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop index %s: %w", indexName, err)
	}
	
	return nil
}

// ListIndexes returns all indexes for a table
func (im *IndexManager) ListIndexes(ctx context.Context, tableName string) ([]IndexDefinition, error) {
	query := `
		SELECT name, type, expr, granularity 
		FROM system.data_skipping_indices 
		WHERE table = ? AND database = currentDatabase()
	`
	
	rows, err := im.store.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	defer rows.Close()
	
	var indexes []IndexDefinition
	for rows.Next() {
		var name, indexType, expr string
		var granularity int
		
		if err := rows.Scan(&name, &indexType, &expr, &granularity); err != nil {
			return nil, fmt.Errorf("failed to scan index row: %w", err)
		}
		
		indexes = append(indexes, IndexDefinition{
			Name:        name,
			Table:       tableName,
			Columns:     parseIndexColumns(expr),
			Type:        IndexType(indexType),
			Granularity: granularity,
			Status:      IndexStatusActive,
		})
	}
	
	return indexes, rows.Err()
}

// CreateRecommendedIndexes creates all recommended indexes for a table
func (im *IndexManager) CreateRecommendedIndexes(ctx context.Context, tableName string) error {
	indexes, exists := PredefinedIndexes[tableName]
	if !exists {
		return fmt.Errorf("no predefined indexes for table %s", tableName)
	}
	
	for _, indexDef := range indexes {
		// Check if index already exists
		existing, err := im.ListIndexes(ctx, tableName)
		if err != nil {
			return fmt.Errorf("failed to check existing indexes: %w", err)
		}
		
		indexExists := false
		for _, existingIndex := range existing {
			if existingIndex.Name == indexDef.Name {
				indexExists = true
				break
			}
		}
		
		if !indexExists {
			if err := im.CreateIndex(ctx, indexDef); err != nil {
				return fmt.Errorf("failed to create index %s: %w", indexDef.Name, err)
			}
		}
	}
	
	return nil
}

// AnalyzeIndexUsage analyzes which indexes are being used for queries
func (im *IndexManager) AnalyzeIndexUsage(ctx context.Context, tableName string) (*IndexUsageStats, error) {
	query := `
		SELECT 
			name,
			type,
			rows_read,
			marks_read,
			marks_selected,
			condition_selectivity
		FROM system.data_skipping_indices_stats 
		WHERE table = ? AND database = currentDatabase()
	`
	
	rows, err := im.store.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze index usage: %w", err)
	}
	defer rows.Close()
	
	stats := &IndexUsageStats{
		TableName: tableName,
		Indexes:   make(map[string]IndexStats),
	}
	
	for rows.Next() {
		var name, indexType string
		var rowsRead, marksRead, marksSelected int64
		var selectivity float64
		
		if err := rows.Scan(&name, &indexType, &rowsRead, &marksRead, &marksSelected, &selectivity); err != nil {
			return nil, fmt.Errorf("failed to scan index stats: %w", err)
		}
		
		stats.Indexes[name] = IndexStats{
			Name:         name,
			Type:         IndexType(indexType),
			RowsRead:     rowsRead,
			MarksRead:    marksRead,
			MarksSelected: marksSelected,
			Selectivity:  selectivity,
		}
	}
	
	return stats, rows.Err()
}

// OptimizeIndexes suggests index optimizations based on usage patterns
func (im *IndexManager) OptimizeIndexes(ctx context.Context, tableName string) ([]IndexOptimization, error) {
	stats, err := im.AnalyzeIndexUsage(ctx, tableName)
	if err != nil {
		return nil, err
	}
	
	var optimizations []IndexOptimization
	
	for indexName, indexStats := range stats.Indexes {
		// Low selectivity index
		if indexStats.Selectivity < 0.1 {
			optimizations = append(optimizations, IndexOptimization{
				IndexName:   indexName,
				Type:        OptimizationDrop,
				Reason:      "Low selectivity - index not effective",
				Priority:    PriorityMedium,
				Impact:      "Reduce storage overhead and improve write performance",
			})
		}
		
		// Unused index
		if indexStats.RowsRead == 0 {
			optimizations = append(optimizations, IndexOptimization{
				IndexName:   indexName,
				Type:        OptimizationDrop,
				Reason:      "Index never used in queries",
				Priority:    PriorityHigh,
				Impact:      "Reduce storage and maintenance overhead",
			})
		}
		
		// Heavily used index with poor performance
		if indexStats.RowsRead > 10000 && indexStats.Selectivity > 0.8 {
			optimizations = append(optimizations, IndexOptimization{
				IndexName:   indexName,
				Type:        OptimizationRebuild,
				Reason:      "High usage but poor selectivity",
				Priority:    PriorityHigh,
				Impact:      "Improve query performance",
			})
		}
	}
	
	return optimizations, nil
}

// Helper function to generate CREATE INDEX query
func (im *IndexManager) generateCreateIndexQuery(indexDef IndexDefinition) string {
	var query strings.Builder
	
	query.WriteString(fmt.Sprintf("ALTER TABLE %s ADD INDEX %s (", indexDef.Table, indexDef.Name))
	query.WriteString(strings.Join(indexDef.Columns, ", "))
	query.WriteString(fmt.Sprintf(") TYPE %s", indexDef.Type))
	
	if indexDef.Granularity > 0 {
		query.WriteString(fmt.Sprintf(" GRANULARITY %d", indexDef.Granularity))
	}
	
	return query.String()
}

// Helper function to parse index columns from expression
func parseIndexColumns(expr string) []string {
	// Simple parsing - in production, you'd want more robust parsing
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		expr = expr[1 : len(expr)-1]
	}
	
	columns := strings.Split(expr, ",")
	for i, col := range columns {
		columns[i] = strings.TrimSpace(col)
	}
	
	return columns
}

// Supporting types for index analysis
type IndexUsageStats struct {
	TableName string                `json:"table_name"`
	Indexes   map[string]IndexStats `json:"indexes"`
}

type IndexStats struct {
	Name          string    `json:"name"`
	Type          IndexType `json:"type"`
	RowsRead      int64     `json:"rows_read"`
	MarksRead     int64     `json:"marks_read"`
	MarksSelected int64     `json:"marks_selected"`
	Selectivity   float64   `json:"selectivity"`
}

type IndexOptimization struct {
	IndexName string             `json:"index_name"`
	Type      OptimizationType   `json:"type"`
	Reason    string             `json:"reason"`
	Priority  OptimizationPriority `json:"priority"`
	Impact    string             `json:"impact"`
}

type OptimizationType string

const (
	OptimizationDrop    OptimizationType = "drop"
	OptimizationRebuild OptimizationType = "rebuild"
	OptimizationCreate  OptimizationType = "create"
)

type OptimizationPriority string

const (
	PriorityHigh   OptimizationPriority = "high"
	PriorityMedium OptimizationPriority = "medium"
	PriorityLow    OptimizationPriority = "low"
)
