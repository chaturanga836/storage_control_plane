package rocksdb

import (
	"fmt"
	"github.com/tecbot/gorocksdb"
	"hybridstore/pkg/models"
)

type Store struct {
	db *gorocksdb.DB
}

func NewStore(path string) (*Store, error) {
	opts := gorocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	db, err := gorocksdb.OpenDb(opts, path)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) GetBusinessData(tenantID string) ([]models.BusinessData, error) {
	// TODO: List all keys for tenantID:* and return as BusinessData
	return nil, fmt.Errorf("not implemented")
}

func (s *Store) PutBusinessData(data models.BusinessData) error {
	// TODO: Use composite key tenantID:dataID for storage
	return fmt.Errorf("not implemented")
}