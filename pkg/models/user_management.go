package models

import "time"

// User Management Models
type SuperAdmin struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Account struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"` // SuperAdmin ID
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"` // active, suspended, deleted
}

type User struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

type UserRole string

const (
	RoleAccountAdmin UserRole = "account_admin"
	RoleDataAdmin    UserRole = "data_admin"
	RoleAnalyst      UserRole = "analyst"
	RoleViewer       UserRole = "viewer"
)

type Tenant struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Config      TenantConfig `json:"config"`
}

type TenantConfig struct {
	RetentionDays     int    `json:"retention_days"`
	MaxStorageGB      int    `json:"max_storage_gb"`
	CompressionLevel  string `json:"compression_level"`
	EncryptionEnabled bool   `json:"encryption_enabled"`
}

type SourceConnection struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id"`
	Name         string            `json:"name"`
	Type         SourceType        `json:"type"`
	Config       map[string]any    `json:"config"`
	Status       ConnectionStatus  `json:"status"`
	LastSync     *time.Time        `json:"last_sync,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

type SourceType string

const (
	SourceAPI      SourceType = "api"
	SourceDatabase SourceType = "database"
	SourceFile     SourceType = "file"
	SourceStream   SourceType = "stream"
)

type ConnectionStatus string

const (
	StatusActive     ConnectionStatus = "active"
	StatusInactive   ConnectionStatus = "inactive"
	StatusError      ConnectionStatus = "error"
	StatusSyncing    ConnectionStatus = "syncing"
)
