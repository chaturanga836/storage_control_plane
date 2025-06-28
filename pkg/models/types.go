package models

import "time"

type BusinessData struct {
	TenantID  string    `json:"tenant_id"`
	DataID    string    `json:"data_id"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type DataStore interface {
	GetBusinessData(tenantID string) ([]BusinessData, error)
	PutBusinessData(data BusinessData) error
}