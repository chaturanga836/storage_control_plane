package clickhouse

import (
	"database/sql"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) GetBusinessData(tenantID string) ([]models.BusinessData, error) {
	query := `SELECT tenant_id, data_id, payload, created_at 
	          FROM business_data 
	          WHERE tenant_id = ? 
	          ORDER BY created_at DESC`
	
	rows, err := s.db.Query(query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query business data: %w", err)
	}
	defer rows.Close()

	var results []models.BusinessData
	for rows.Next() {
		var data models.BusinessData
		if err := rows.Scan(&data.TenantID, &data.DataID, &data.Payload, &data.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, data)
	}
	
	return results, rows.Err()
}

func (s *Store) PutBusinessData(data models.BusinessData) error {
	query := `INSERT INTO business_data (tenant_id, data_id, payload, created_at) 
	          VALUES (?, ?, ?, ?)`
	
	_, err := s.db.Exec(query, data.TenantID, data.DataID, data.Payload, data.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert business data: %w", err)
	}
	
	return nil
}