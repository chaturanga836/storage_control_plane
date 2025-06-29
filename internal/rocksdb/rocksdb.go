package rocksdb

import (
	"fmt"
	"github.com/chaturanga836/storage_system/go-control-plane/pkg/models"
)

type Store struct {
	path string
}

func NewStore(path string) (*Store, error) {
	// Mock implementation - no actual RocksDB dependency
	return &Store{path: path}, nil
}

func (s *Store) GetBusinessData(tenantID string) ([]models.BusinessData, error) {
	// TODO: List all keys for tenantID:* and return as BusinessData
	return nil, fmt.Errorf("not implemented")
}

func (s *Store) PutBusinessData(data models.BusinessData) error {
	// TODO: Use composite key tenantID:dataID for storage
	return fmt.Errorf("not implemented")
}