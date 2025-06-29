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
	UpdatedAt    time.Time         `json:"updated_at"`
	
	// Directory structure configuration
	DirectoryConfigID string `json:"directory_config_id,omitempty"` // References DirectoryConfig
	
	// Sync configuration
	SyncConfig SourceSyncConfig `json:"sync_config"`
	
	// Data schema and validation
	SchemaConfig SourceSchemaConfig `json:"schema_config"`
	
	// Performance metrics
	Metrics SourceMetrics `json:"metrics"`
}

// SourceSyncConfig defines how often and how to sync data from this source
type SourceSyncConfig struct {
	Enabled         bool          `json:"enabled"`
	Interval        time.Duration `json:"interval"`        // e.g., 5m, 1h, 24h
	BatchSize       int           `json:"batch_size"`      // number of records per batch
	ConcurrentJobs  int           `json:"concurrent_jobs"` // number of parallel sync jobs
	RetryAttempts   int           `json:"retry_attempts"`
	RetryBackoff    time.Duration `json:"retry_backoff"`
	TimeoutSeconds  int           `json:"timeout_seconds"`
	LastSyncAttempt *time.Time    `json:"last_sync_attempt,omitempty"`
	NextSyncDue     *time.Time    `json:"next_sync_due,omitempty"`
}

// SourceSchemaConfig defines schema detection and validation rules
type SourceSchemaConfig struct {
	AutoDetectSchema    bool              `json:"auto_detect_schema"`
	RequiredFields      []string          `json:"required_fields"`      // fields that must be present
	ForbiddenFields     []string          `json:"forbidden_fields"`     // fields that should be excluded
	FieldMappings       map[string]string `json:"field_mappings"`       // rename fields: "old_name" -> "new_name"
	DefaultValues       map[string]any    `json:"default_values"`       // default values for missing fields
	ValidationRules     map[string]string `json:"validation_rules"`     // regex validation per field
	SchemaEvolution     SchemaEvolution   `json:"schema_evolution"`     // how to handle schema changes
	TimestampField      string            `json:"timestamp_field"`      // which field contains the timestamp
	TimestampFormat     string            `json:"timestamp_format"`     // timestamp parsing format
}

// SchemaEvolution defines how to handle schema changes
type SchemaEvolution struct {
	AllowNewFields    bool `json:"allow_new_fields"`    // allow new fields to be added
	AllowFieldRemoval bool `json:"allow_field_removal"` // allow fields to be removed
	AllowTypeChanges  bool `json:"allow_type_changes"`  // allow field type changes
	StrictMode        bool `json:"strict_mode"`         // reject any schema changes
}

// SourceMetrics tracks performance and health metrics
type SourceMetrics struct {
	TotalRecordsProcessed int64     `json:"total_records_processed"`
	TotalBytesProcessed   int64     `json:"total_bytes_processed"`
	TotalParquetFiles     int       `json:"total_parquet_files"`
	AverageRecordsPerFile int       `json:"average_records_per_file"`
	LastSyncDuration      int64     `json:"last_sync_duration_ms"`
	AverageSyncDuration   int64     `json:"average_sync_duration_ms"`
	ErrorCount            int       `json:"error_count"`
	LastError             string    `json:"last_error,omitempty"`
	LastErrorTime         *time.Time `json:"last_error_time,omitempty"`
	SuccessRate           float64   `json:"success_rate"` // percentage of successful syncs
}
