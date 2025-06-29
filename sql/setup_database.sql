-- setup_database.sql - Complete database setup with optimized indexes
-- Run this script to create all tables with their recommended indexes

-- Business Data Table
CREATE TABLE IF NOT EXISTS business_data (
    tenant_id String COMMENT 'Tenant identifier',
    data_id String COMMENT 'Unique data identifier',
    payload String COMMENT 'JSON payload data',
    created_at DateTime COMMENT 'Record creation timestamp',
    updated_at DateTime DEFAULT now() COMMENT 'Last update timestamp',
    version UInt64 DEFAULT 1 COMMENT 'Record version for optimistic locking',
    size_bytes UInt64 DEFAULT 0 COMMENT 'Payload size in bytes',
    
    INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_created_at created_at TYPE minmax GRANULARITY 1,
    INDEX idx_data_id data_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_updated_at updated_at TYPE minmax GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (tenant_id, created_at)
PARTITION BY toYYYYMM(created_at)
SETTINGS index_granularity = 8192, storage_policy = 'default';

-- Analytics Events Table
CREATE TABLE IF NOT EXISTS analytics_events (
    tenant_id String COMMENT 'Tenant identifier',
    event_id String COMMENT 'Unique event identifier',
    timestamp DateTime COMMENT 'Event timestamp',
    event_type String COMMENT 'Type of event',
    value Float64 DEFAULT 0.0 COMMENT 'Numeric value associated with event',
    properties String COMMENT 'JSON properties',
    source_id String COMMENT 'Source system identifier',
    session_id Nullable(String) COMMENT 'User session identifier',
    
    INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1,
    INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_event_type event_type TYPE set GRANULARITY 1,
    INDEX idx_source_id source_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_value value TYPE minmax GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (tenant_id, timestamp)
PARTITION BY (tenant_id, toYYYYMM(timestamp))
SETTINGS index_granularity = 8192, storage_policy = 'hot_and_cold';

-- Parquet Files Metadata Table
CREATE TABLE IF NOT EXISTS parquet_files (
    id String COMMENT 'Unique file identifier',
    tenant_id String COMMENT 'Tenant identifier',
    source_id String COMMENT 'Source system identifier',
    file_path String COMMENT 'File path in storage',
    schema_hash String COMMENT 'Schema version hash',
    record_count UInt64 COMMENT 'Number of records in file',
    file_size UInt64 COMMENT 'File size in bytes',
    min_timestamp DateTime COMMENT 'Earliest record timestamp',
    max_timestamp DateTime COMMENT 'Latest record timestamp',
    created_at DateTime COMMENT 'File creation timestamp',
    compressed UInt8 DEFAULT 1 COMMENT 'Whether file is compressed',
    compression_ratio Float32 DEFAULT 0.0 COMMENT 'Compression ratio achieved',
    
    INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_created_at created_at TYPE minmax GRANULARITY 1,
    INDEX idx_min_timestamp min_timestamp TYPE minmax GRANULARITY 1,
    INDEX idx_max_timestamp max_timestamp TYPE minmax GRANULARITY 1,
    INDEX idx_source_id source_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_file_path file_path TYPE bloom_filter GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (tenant_id, created_at)
PARTITION BY toYYYYMM(created_at)
SETTINGS index_granularity = 8192, storage_policy = 'default';

-- Tenant Metadata Table (for tenant summaries)
CREATE TABLE IF NOT EXISTS tenant_metadata (
    tenant_id String COMMENT 'Tenant identifier',
    total_files UInt64 DEFAULT 0 COMMENT 'Total number of files',
    total_rows UInt64 DEFAULT 0 COMMENT 'Total number of rows',
    total_size_gb Float64 DEFAULT 0.0 COMMENT 'Total size in GB',
    source_count UInt32 DEFAULT 0 COMMENT 'Number of sources',
    oldest_record Nullable(DateTime) COMMENT 'Oldest record timestamp',
    newest_record Nullable(DateTime) COMMENT 'Newest record timestamp',
    last_updated DateTime DEFAULT now() COMMENT 'Last update timestamp',
    settings String DEFAULT '{}' COMMENT 'Tenant-specific settings JSON',
    
    INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_last_updated last_updated TYPE minmax GRANULARITY 1
) ENGINE = ReplacingMergeTree(last_updated)
ORDER BY (tenant_id)
SETTINGS index_granularity = 8192;

-- WAL (Write-Ahead Log) Table for reliability
CREATE TABLE IF NOT EXISTS wal_entries (
    id String COMMENT 'Unique WAL entry identifier',
    tenant_id String COMMENT 'Tenant identifier',
    source_id String COMMENT 'Source system identifier',
    timestamp DateTime COMMENT 'WAL entry timestamp',
    operation String COMMENT 'Operation type (INSERT, UPDATE, DELETE)',
    table_name String COMMENT 'Target table name',
    data String COMMENT 'JSON data for the operation',
    flushed UInt8 DEFAULT 0 COMMENT 'Whether entry has been flushed',
    flush_time Nullable(DateTime) COMMENT 'When entry was flushed',
    retry_count UInt32 DEFAULT 0 COMMENT 'Number of retry attempts',
    
    INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1,
    INDEX idx_flushed flushed TYPE set GRANULARITY 1,
    INDEX idx_table_name table_name TYPE set GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (tenant_id, timestamp)
PARTITION BY toYYYYMM(timestamp)
SETTINGS index_granularity = 8192;

-- Schema Versions Table for tracking data evolution
CREATE TABLE IF NOT EXISTS schema_versions (
    hash String COMMENT 'Schema hash identifier',
    tenant_id String COMMENT 'Tenant identifier',
    source_id String COMMENT 'Source system identifier',
    version UInt32 COMMENT 'Schema version number',
    schema String COMMENT 'JSON schema definition',
    first_seen DateTime COMMENT 'When schema was first encountered',
    last_seen DateTime COMMENT 'When schema was last used',
    backward_compatible UInt8 DEFAULT 0 COMMENT 'Whether backward compatible',
    forward_compatible UInt8 DEFAULT 0 COMMENT 'Whether forward compatible',
    record_count UInt64 DEFAULT 0 COMMENT 'Number of records using this schema',
    
    INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_hash hash TYPE bloom_filter GRANULARITY 1,
    INDEX idx_first_seen first_seen TYPE minmax GRANULARITY 1,
    INDEX idx_last_seen last_seen TYPE minmax GRANULARITY 1
) ENGINE = ReplacingMergeTree(last_seen)
ORDER BY (tenant_id, hash)
PARTITION BY tenant_id
SETTINGS index_granularity = 8192;

-- Query Performance Metrics Table
CREATE TABLE IF NOT EXISTS query_metrics (
    query_id String COMMENT 'Unique query identifier',
    tenant_id String COMMENT 'Tenant identifier',
    timestamp DateTime COMMENT 'Query execution timestamp',
    query_type String COMMENT 'Type of query (SQL, timeseries, etc)',
    execution_time_ms UInt64 COMMENT 'Execution time in milliseconds',
    rows_processed UInt64 COMMENT 'Number of rows processed',
    memory_used UInt64 COMMENT 'Memory used in bytes',
    indexes_used String COMMENT 'JSON array of indexes used',
    streaming_used UInt8 DEFAULT 0 COMMENT 'Whether streaming was used',
    chunks_processed UInt32 DEFAULT 0 COMMENT 'Number of chunks processed',
    sort_fields String COMMENT 'JSON array of sort fields used',
    complexity String COMMENT 'Query complexity (LOW, MEDIUM, HIGH)',
    
    INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1,
    INDEX idx_query_type query_type TYPE set GRANULARITY 1,
    INDEX idx_execution_time execution_time_ms TYPE minmax GRANULARITY 1,
    INDEX idx_complexity complexity TYPE set GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (tenant_id, timestamp)
PARTITION BY (tenant_id, toYYYYMM(timestamp))
SETTINGS index_granularity = 8192, storage_policy = 'hot_and_cold';

-- Create materialized views for common aggregations
CREATE MATERIALIZED VIEW IF NOT EXISTS tenant_daily_stats
ENGINE = SummingMergeTree()
ORDER BY (tenant_id, date)
AS SELECT
    tenant_id,
    toDate(timestamp) as date,
    count() as query_count,
    avg(execution_time_ms) as avg_execution_time,
    sum(rows_processed) as total_rows_processed,
    max(execution_time_ms) as max_execution_time,
    countIf(streaming_used = 1) as streaming_queries
FROM query_metrics
GROUP BY tenant_id, toDate(timestamp);

-- Create view for index usage analysis
CREATE VIEW IF NOT EXISTS index_usage_summary AS
SELECT
    table,
    name as index_name,
    type as index_type,
    sum(rows_read) as total_rows_read,
    sum(marks_read) as total_marks_read,
    avg(condition_selectivity) as avg_selectivity,
    count() as usage_count
FROM system.data_skipping_indices_stats
WHERE database = currentDatabase()
GROUP BY table, name, type
ORDER BY total_rows_read DESC;

-- Optimization recommendations view
CREATE VIEW IF NOT EXISTS optimization_candidates AS
SELECT
    table,
    index_name,
    index_type,
    total_rows_read,
    avg_selectivity,
    CASE
        WHEN total_rows_read = 0 THEN 'DROP - Never used'
        WHEN avg_selectivity < 0.1 THEN 'DROP - Low selectivity'
        WHEN total_rows_read > 10000 AND avg_selectivity > 0.8 THEN 'REBUILD - High usage, poor selectivity'
        ELSE 'OK'
    END as recommendation,
    CASE
        WHEN total_rows_read = 0 OR avg_selectivity < 0.1 THEN 'HIGH'
        WHEN total_rows_read > 10000 AND avg_selectivity > 0.8 THEN 'HIGH'
        ELSE 'LOW'
    END as priority
FROM index_usage_summary
WHERE recommendation != 'OK'
ORDER BY 
    CASE priority WHEN 'HIGH' THEN 1 ELSE 2 END,
    total_rows_read DESC;

-- Insert sample data for testing (optional)
-- Uncomment the following lines to add test data

/*
INSERT INTO business_data (tenant_id, data_id, payload, created_at) VALUES
('tenant_001', 'data_001', '{"type": "test", "value": 123}', now() - INTERVAL 1 DAY),
('tenant_001', 'data_002', '{"type": "test", "value": 456}', now() - INTERVAL 2 HOUR),
('tenant_002', 'data_003', '{"type": "prod", "value": 789}', now() - INTERVAL 1 HOUR);

INSERT INTO analytics_events (tenant_id, event_id, timestamp, event_type, value, source_id) VALUES
('tenant_001', 'evt_001', now() - INTERVAL 1 DAY, 'page_view', 1.0, 'web'),
('tenant_001', 'evt_002', now() - INTERVAL 1 HOUR, 'button_click', 1.0, 'web'),
('tenant_002', 'evt_003', now() - INTERVAL 30 MINUTE, 'api_call', 1.0, 'api');
*/

-- Show created tables and their indexes
SELECT 
    table,
    name as index_name,
    type as index_type,
    granularity
FROM system.data_skipping_indices 
WHERE database = currentDatabase()
ORDER BY table, name;
