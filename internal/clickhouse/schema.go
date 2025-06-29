// internal/clickhouse/schema.go

package clickhouse

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
	"github.com/chaturanga836/storage_system/go-control-plane/internal/utils"
)

type FieldDef struct {
	Name string
	Type string
}

func mapJSONTypeToClickHouseType(value any) string {
	switch v := value.(type) {
	case string:
		return "String"
	case bool:
		return "UInt8"
	case float64:
		if v == float64(int64(v)) {
			return "Int64"
		}
		return "Float64"
	case map[string]any, []any:
		return "String" // store as JSON string for now
	default:
		return "String"
	}
}

func FlattenJSON(data map[string]any, prefix string, out map[string]string) {
	for k, v := range data {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}
		switch vv := v.(type) {
		case map[string]any:
			FlattenJSON(vv, fullKey, out)
		default:
			out[fullKey] = mapJSONTypeToClickHouseType(vv)
		}
	}
}

func GenerateCreateTableSQL(tenantID, connID string, sample map[string]any) string {
	flatSchema := map[string]string{}
	FlattenJSON(sample, "", flatSchema)

	fields := make([]FieldDef, 0, len(flatSchema))
	for k, v := range flatSchema {
		fields = append(fields, FieldDef{Name: k, Type: v})
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })

	cols := make([]string, 0, len(fields)+2)
	cols = append(cols, "id String", "updated_at DateTime", "version_id String")
	for _, f := range fields {
		cols = append(cols, fmt.Sprintf("`%s` %s", f.Name, f.Type))
	}

	table := fmt.Sprintf("tenant_%s__conn_%s", tenantID, connID)
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s)
ENGINE = MergeTree()
ORDER BY (id, updated_at);`, table, strings.Join(cols, ", "))
}

func ExecuteSQL(sql string, clickhouseURL string) error {
	body := bytes.NewBufferString(sql)
	req, err := http.NewRequest("POST", clickhouseURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ClickHouse returned error: %s", resp.Status)
	}

	return nil
}

func EnsureClickHouseTable(tenantID, connID string, sampleJSON map[string]any, clickhouseURL string) error {
	sql := GenerateCreateTableSQL(tenantID, connID, sampleJSON)
	return ExecuteSQL(sql, clickhouseURL)
}

// QueryWithSorting executes a ClickHouse query with sorting applied
func (store *Store) QueryWithSorting(query string, sortFields []models.SortField, sortOptions utils.SortOptions) ([]map[string]any, error) {
	// Validate sort fields
	validatedSorts, err := utils.ValidateSortFields(sortFields, sortOptions)
	if err != nil {
		return nil, fmt.Errorf("invalid sort parameters: %w", err)
	}

	// Generate ORDER BY clause
	orderByClause := utils.GenerateClickHouseOrderBy(validatedSorts)

	// Append ORDER BY to query if not already present
	if orderByClause != "" && !strings.Contains(strings.ToUpper(query), "ORDER BY") {
		query = query + " " + orderByClause
	}

	// Execute query (placeholder - implement actual ClickHouse execution)
	return store.executeQuery(query)
}

// GetTenantDataSorted retrieves tenant data with sorting
func (store *Store) GetTenantDataSorted(tenantID string, sortFields []models.SortField) ([]map[string]any, error) {
	// Validate sort fields for tenant data
	validatedSorts, err := utils.ValidateSortFields(sortFields, utils.TenantSortOptions)
	if err != nil {
		return nil, fmt.Errorf("invalid sort parameters: %w", err)
	}

	// Build query with sorting
	tableName := fmt.Sprintf("tenant_%s_data", tenantID)
	baseQuery := fmt.Sprintf("SELECT * FROM %s", tableName)

	orderByClause := utils.GenerateClickHouseOrderBy(validatedSorts)
	if orderByClause != "" {
		baseQuery += " " + orderByClause
	}

	return store.executeQuery(baseQuery)
}

// executeQuery executes a ClickHouse query and returns results
func (store *Store) executeQuery(query string) ([]map[string]any, error) {
	rows, err := store.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column information
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]any
	
	for rows.Next() {
		// Create slice of interface{} to hold each column value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		
		// Scan the row into our slice
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		// Create a map for this row
		rowMap := make(map[string]any)
		for i, col := range columns {
			var v interface{}
			val := values[i]
			
			// Handle different types that ClickHouse might return
			if b, ok := val.([]byte); ok {
				v = string(b)
			} else {
				v = val
			}
			rowMap[col] = v
		}
		
		results = append(results, rowMap)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	
	return results, nil
}
