// directory_config.go - User-defined directory structure configuration
package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// DirectoryConfig defines how data should be organized in storage
type DirectoryConfig struct {
	ID               string                 `json:"id"`
	TenantID         string                 `json:"tenant_id"`
	SourceConnectionID string               `json:"source_connection_id"`
	Name             string                 `json:"name"`
	Pattern          string                 `json:"pattern"`          // e.g., "{tenant_id}/{source_id}/{year}/{month}/{day}"
	Variables        map[string]Variable    `json:"variables"`        // Variable definitions
	FileNaming       FileNamingConfig       `json:"file_naming"`      // How to name Parquet files
	MetadataConfig   MetadataConfig         `json:"metadata_config"`  // What metadata to generate
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	IsActive         bool                   `json:"is_active"`
}

// Variable defines a replaceable variable in the directory pattern
type Variable struct {
	Name        string     `json:"name"`        // e.g., "year", "month", "region"
	Type        VarType    `json:"type"`        // datetime, json_field, static, computed
	Source      string     `json:"source"`      // where to get the value from
	Format      string     `json:"format"`      // formatting rules
	DefaultValue string    `json:"default_value,omitempty"`
	Required    bool       `json:"required"`
	Validation  string     `json:"validation,omitempty"` // regex for validation
}

type VarType string

const (
	VarTypeDateTime   VarType = "datetime"    // Extract from timestamp: {year}, {month}, {day}, {hour}
	VarTypeJSONField  VarType = "json_field"  // Extract from JSON data: {user.region}, {event.type}
	VarTypeStatic     VarType = "static"      // Static value: {environment}, {version}
	VarTypeComputed   VarType = "computed"    // Computed value: {schema_hash}, {record_count}
)

// FileNamingConfig defines how to name Parquet files
type FileNamingConfig struct {
	Pattern     string `json:"pattern"`      // e.g., "data_{timestamp}_{sequence}.parquet"
	Timestamp   string `json:"timestamp"`    // timestamp format: "2006-01-02_15-04-05"
	Sequence    bool   `json:"sequence"`     // include sequence number
	SchemaHash  bool   `json:"schema_hash"`  // include schema hash
	MaxFileSize int64  `json:"max_file_size"` // max file size in bytes
}

// MetadataConfig defines what metadata files to generate
type MetadataConfig struct {
	GenerateSummary   bool                   `json:"generate_summary"`   // _summary.json
	GenerateIndexes   bool                   `json:"generate_indexes"`   // _indexes.json
	GenerateStats     bool                   `json:"generate_stats"`     // _stats.json
	GenerateSchema    bool                   `json:"generate_schema"`    // _schema.json
	CustomMetadata    map[string]interface{} `json:"custom_metadata"`    // custom metadata fields
	IndexedFields     []string               `json:"indexed_fields"`     // fields to create indexes for
	StatsFields       []string               `json:"stats_fields"`       // fields to compute stats for
}

// DirectoryResolver resolves directory patterns to actual paths
type DirectoryResolver struct {
	Config    *DirectoryConfig
	Timestamp time.Time
	Data      map[string]interface{}
}

// ResolvePath resolves the directory pattern to an actual path
func (dr *DirectoryResolver) ResolvePath() (string, error) {
	pattern := dr.Config.Pattern
	
	// Replace all variables in the pattern
	for varName, variable := range dr.Config.Variables {
		placeholder := "{" + varName + "}"
		value, err := dr.resolveVariable(variable)
		if err != nil {
			if variable.Required {
				return "", fmt.Errorf("failed to resolve required variable %s: %w", varName, err)
			}
			value = variable.DefaultValue
		}
		
		// Validate the value if validation regex is provided
		if variable.Validation != "" {
			matched, err := regexp.MatchString(variable.Validation, value)
			if err != nil {
				return "", fmt.Errorf("invalid validation regex for variable %s: %w", varName, err)
			}
			if !matched {
				return "", fmt.Errorf("variable %s value '%s' does not match validation pattern", varName, value)
			}
		}
		
		pattern = strings.ReplaceAll(pattern, placeholder, value)
	}
	
	// Check if any unresolved placeholders remain
	if strings.Contains(pattern, "{") && strings.Contains(pattern, "}") {
		return "", fmt.Errorf("unresolved placeholders in pattern: %s", pattern)
	}
	
	return pattern, nil
}

// resolveVariable resolves a single variable based on its type
func (dr *DirectoryResolver) resolveVariable(variable Variable) (string, error) {
	switch variable.Type {
	case VarTypeDateTime:
		return dr.resolveDateTimeVariable(variable)
	case VarTypeJSONField:
		return dr.resolveJSONFieldVariable(variable)
	case VarTypeStatic:
		return variable.Source, nil
	case VarTypeComputed:
		return dr.resolveComputedVariable(variable)
	default:
		return "", fmt.Errorf("unknown variable type: %s", variable.Type)
	}
}

// resolveDateTimeVariable resolves datetime-based variables
func (dr *DirectoryResolver) resolveDateTimeVariable(variable Variable) (string, error) {
	timestamp := dr.Timestamp
	
	switch variable.Source {
	case "year":
		return fmt.Sprintf("%04d", timestamp.Year()), nil
	case "month":
		return fmt.Sprintf("%02d", timestamp.Month()), nil
	case "day":
		return fmt.Sprintf("%02d", timestamp.Day()), nil
	case "hour":
		return fmt.Sprintf("%02d", timestamp.Hour()), nil
	case "date":
		if variable.Format != "" {
			return timestamp.Format(variable.Format), nil
		}
		return timestamp.Format("2006-01-02"), nil
	case "datetime":
		if variable.Format != "" {
			return timestamp.Format(variable.Format), nil
		}
		return timestamp.Format("2006-01-02_15-04-05"), nil
	default:
		return "", fmt.Errorf("unknown datetime source: %s", variable.Source)
	}
}

// resolveJSONFieldVariable extracts values from JSON data
func (dr *DirectoryResolver) resolveJSONFieldVariable(variable Variable) (string, error) {
	if dr.Data == nil {
		return "", fmt.Errorf("no data available for JSON field extraction")
	}
	
	// Support nested field access like "user.region" or "metadata.source.type"
	fieldPath := strings.Split(variable.Source, ".")
	current := dr.Data
	
	for i, field := range fieldPath {
		if value, ok := current[field]; ok {
			if i == len(fieldPath)-1 {
				// Last field, return the value
				if str, ok := value.(string); ok {
					return str, nil
				}
				return fmt.Sprintf("%v", value), nil
			} else {
				// Intermediate field, continue traversing
				if nested, ok := value.(map[string]interface{}); ok {
					current = nested
				} else {
					return "", fmt.Errorf("field %s is not a nested object", field)
				}
			}
		} else {
			return "", fmt.Errorf("field %s not found in data", strings.Join(fieldPath[:i+1], "."))
		}
	}
	
	return "", fmt.Errorf("failed to extract field %s", variable.Source)
}

// resolveComputedVariable resolves computed variables
func (dr *DirectoryResolver) resolveComputedVariable(variable Variable) (string, error) {
	switch variable.Source {
	case "schema_hash":
		if dr.Data == nil {
			return "", fmt.Errorf("no data available for schema hash computation")
		}
		// This would use the existing schema hasher
		return "computed_schema_hash", nil // Placeholder
	case "record_count":
		// This would be computed based on the batch size
		return "1000", nil // Placeholder
	case "file_size":
		// This would be computed based on the data size
		return "10MB", nil // Placeholder
	default:
		return "", fmt.Errorf("unknown computed variable: %s", variable.Source)
	}
}

// ResolveFileName resolves the file naming pattern
func (dr *DirectoryResolver) ResolveFileName(sequence int) string {
	config := dr.Config.FileNaming
	filename := config.Pattern
	
	// Replace timestamp
	if strings.Contains(filename, "{timestamp}") {
		timestampStr := dr.Timestamp.Format(config.Timestamp)
		filename = strings.ReplaceAll(filename, "{timestamp}", timestampStr)
	}
	
	// Replace sequence
	if config.Sequence && strings.Contains(filename, "{sequence}") {
		filename = strings.ReplaceAll(filename, "{sequence}", fmt.Sprintf("%05d", sequence))
	}
	
	// Replace schema hash
	if config.SchemaHash && strings.Contains(filename, "{schema_hash}") {
		// This would use the actual schema hash
		filename = strings.ReplaceAll(filename, "{schema_hash}", "schema123")
	}
	
	// Ensure .parquet extension
	if !strings.HasSuffix(filename, ".parquet") {
		filename += ".parquet"
	}
	
	return filename
}

// Predefined directory patterns for common use cases
var PredefinedPatterns = map[string]DirectoryConfig{
	"tenant_source_date": {
		Name:    "Tenant/Source/Date",
		Pattern: "{tenant_id}/{source_id}/{year}/{month}/{day}",
		Variables: map[string]Variable{
			"tenant_id": {Name: "tenant_id", Type: VarTypeStatic, Required: true},
			"source_id": {Name: "source_id", Type: VarTypeStatic, Required: true},
			"year":      {Name: "year", Type: VarTypeDateTime, Source: "year", Required: true},
			"month":     {Name: "month", Type: VarTypeDateTime, Source: "month", Required: true},
			"day":       {Name: "day", Type: VarTypeDateTime, Source: "day", Required: true},
		},
		FileNaming: FileNamingConfig{
			Pattern:   "data_{timestamp}_{sequence}",
			Timestamp: "2006-01-02_15-04-05",
			Sequence:  true,
		},
		MetadataConfig: MetadataConfig{
			GenerateSummary: true,
			GenerateIndexes: true,
			GenerateStats:   true,
			GenerateSchema:  true,
		},
	},
	
	"hierarchical_json": {
		Name:    "Hierarchical by JSON Fields",
		Pattern: "{tenant_id}/{source_id}/{region}/{category}/{year}/{month}",
		Variables: map[string]Variable{
			"tenant_id": {Name: "tenant_id", Type: VarTypeStatic, Required: true},
			"source_id": {Name: "source_id", Type: VarTypeStatic, Required: true},
			"region":    {Name: "region", Type: VarTypeJSONField, Source: "user.region", DefaultValue: "unknown"},
			"category":  {Name: "category", Type: VarTypeJSONField, Source: "event.category", DefaultValue: "general"},
			"year":      {Name: "year", Type: VarTypeDateTime, Source: "year", Required: true},
			"month":     {Name: "month", Type: VarTypeDateTime, Source: "month", Required: true},
		},
		FileNaming: FileNamingConfig{
			Pattern:    "events_{timestamp}",
			Timestamp:  "2006-01-02_15",
			SchemaHash: true,
		},
		MetadataConfig: MetadataConfig{
			GenerateSummary: true,
			GenerateIndexes: true,
			IndexedFields:   []string{"user.region", "event.category", "timestamp"},
			StatsFields:     []string{"event.value", "user.age"},
		},
	},
}

// DirectoryConfigService manages directory configurations
type DirectoryConfigService struct {
	configs map[string]*DirectoryConfig
}

// NewDirectoryConfigService creates a new directory config service
func NewDirectoryConfigService() *DirectoryConfigService {
	return &DirectoryConfigService{
		configs: make(map[string]*DirectoryConfig),
	}
}

// CreateConfig creates a new directory configuration
func (s *DirectoryConfigService) CreateConfig(config *DirectoryConfig) error {
	if config.ID == "" {
		return fmt.Errorf("config ID is required")
	}
	
	// Validate the configuration
	if err := s.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	config.IsActive = true
	
	s.configs[config.ID] = config
	return nil
}

// GetConfig retrieves a directory configuration
func (s *DirectoryConfigService) GetConfig(id string) (*DirectoryConfig, error) {
	config, exists := s.configs[id]
	if !exists {
		return nil, fmt.Errorf("configuration not found: %s", id)
	}
	return config, nil
}

// GetConfigBySourceConnection retrieves config by source connection ID
func (s *DirectoryConfigService) GetConfigBySourceConnection(sourceConnectionID string) (*DirectoryConfig, error) {
	for _, config := range s.configs {
		if config.SourceConnectionID == sourceConnectionID && config.IsActive {
			return config, nil
		}
	}
	return nil, fmt.Errorf("no configuration found for source connection: %s", sourceConnectionID)
}

// validateConfig validates a directory configuration
func (s *DirectoryConfigService) validateConfig(config *DirectoryConfig) error {
	if config.Pattern == "" {
		return fmt.Errorf("pattern is required")
	}
	
	if config.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	
	// Check that all variables in pattern are defined
	pattern := config.Pattern
	placeholderRegex := regexp.MustCompile(`\{([^}]+)\}`)
	matches := placeholderRegex.FindAllStringSubmatch(pattern, -1)
	
	for _, match := range matches {
		varName := match[1]
		if _, exists := config.Variables[varName]; !exists {
			return fmt.Errorf("variable %s used in pattern but not defined", varName)
		}
	}
	
	// Validate each variable
	for name, variable := range config.Variables {
		if variable.Name == "" {
			variable.Name = name
		}
		
		if variable.Type == "" {
			return fmt.Errorf("variable %s: type is required", name)
		}
		
		if variable.Type == VarTypeJSONField && variable.Source == "" {
			return fmt.Errorf("variable %s: source is required for json_field type", name)
		}
		
		if variable.Type == VarTypeStatic && variable.Source == "" {
			return fmt.Errorf("variable %s: source is required for static type", name)
		}
	}
	
	return nil
}
