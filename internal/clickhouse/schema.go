// internal/clickhouse/schema.go

package clickhouse

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
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
