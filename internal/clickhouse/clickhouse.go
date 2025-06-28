package clickhouse

import (
	"database/sql"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	"hybridstore/pkg/models"
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
	// TODO: SELECT * FROM business_data WHERE tenant_id = ?
	return nil, nil
}

func (s *Store) PutBusinessData(data models.BusinessData) error {
	// TODO: INSERT INTO business_data (tenant_id, data_id, payload, created_at) VALUES ...
	return nil
}