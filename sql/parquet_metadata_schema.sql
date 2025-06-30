-- parquet_metadata_schema.sql
-- ClickHouse schema for Parquet file metadata and searchable record metadata

-- Table 1: Parquet File Metadata
-- This table contains metadata about each Parquet file for query optimization
CREATE TABLE IF NOT EXISTS parquet_file_metadata (
    file_id String COMMENT 'Unique file identifier',
    tenant_id String COMMENT 'Tenant identifier',
    source_id String COMMENT 'Source connection identifier', 
    file_path String COMMENT 'Full path to Parquet file',
    directory_path String COMMENT 'Directory containing the file',
    
    -- File properties
    file_size UInt64 COMMENT 'File size in bytes',
    record_count UInt64 COMMENT 'Number of records in file',
    created_at DateTime COMMENT 'File creation timestamp',
    schema_hash String COMMENT 'Hash of the schema used',
    
    -- Data boundaries for query optimization
    min_timestamp Nullable(DateTime) COMMENT 'Earliest timestamp in file',
    max_timestamp Nullable(DateTime) COMMENT 'Latest timestamp in file',
    min_created_at Nullable(DateTime) COMMENT 'Earliest created_at in file',
    max_created_at Nullable(DateTime) COMMENT 'Latest created_at in file',
    
    -- Statistics as JSON
    indexed_fields String COMMENT 'JSON of indexed field metadata',
    stats String COMMENT 'JSON of file statistics',
    
    -- Indexes for fast file discovery
    INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_source_id source_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_directory_path directory_path TYPE bloom_filter GRANULARITY 1,
    INDEX idx_created_at created_at TYPE minmax GRANULARITY 1,
    INDEX idx_schema_hash schema_hash TYPE bloom_filter GRANULARITY 1,
    INDEX idx_min_timestamp min_timestamp TYPE minmax GRANULARITY 1,
    INDEX idx_max_timestamp max_timestamp TYPE minmax GRANULARITY 1
) ENGINE = ReplacingMergeTree(created_at)
ORDER BY (tenant_id, source_id, file_id)
PARTITION BY (tenant_id, toYYYYMM(created_at))
SETTINGS index_granularity = 8192;

-- Table 2: Record Metadata (Searchable Index)
-- This table contains searchable metadata for individual records across all Parquet files
-- This is the KEY table that solves your problem!
CREATE TABLE IF NOT EXISTS record_metadata (
    record_id String COMMENT 'Unique record identifier',
    tenant_id String COMMENT 'Tenant identifier',
    source_id String COMMENT 'Source connection identifier',
    file_id String COMMENT 'File containing this record',
    file_path String COMMENT 'Path to the Parquet file',
    
    -- Searchable fields - these enable cross-file queries
    name String DEFAULT '' COMMENT 'Name field for searching (e.g., contains "sam")',
    email String DEFAULT '' COMMENT 'Email field for searching',
    status String DEFAULT '' COMMENT 'Status field for filtering',
    category String DEFAULT '' COMMENT 'Category field for filtering',
    tags Array(String) DEFAULT [] COMMENT 'Tags for tag-based searches',
    
    -- Timestamps for temporal queries
    created_at DateTime COMMENT 'Record creation timestamp',
    updated_at DateTime COMMENT 'Record update timestamp', 
    timestamp DateTime COMMENT 'Record timestamp',
    
    -- Custom fields as JSON (configured per source)
    custom_fields String DEFAULT '{}' COMMENT 'JSON of custom indexed fields',
    
    -- Location in file for efficient retrieval
    row_number UInt64 COMMENT 'Row number in the Parquet file',
    offset Nullable(UInt64) COMMENT 'Byte offset in file',
    
    -- Indexes for fast searching - THIS IS THE MAGIC!
    INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_source_id source_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_name name TYPE bloom_filter GRANULARITY 1,
    INDEX idx_email email TYPE bloom_filter GRANULARITY 1,
    INDEX idx_status status TYPE set GRANULARITY 1,
    INDEX idx_category category TYPE set GRANULARITY 1,
    INDEX idx_created_at created_at TYPE minmax GRANULARITY 1,
    INDEX idx_updated_at updated_at TYPE minmax GRANULARITY 1,
    INDEX idx_timestamp timestamp TYPE minmax GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (tenant_id, source_id, created_at, record_id)
PARTITION BY (tenant_id, toYYYYMM(created_at))
SETTINGS index_granularity = 8192;

-- Table 3: Directory Summaries
-- This table contains aggregated information for each directory
CREATE TABLE IF NOT EXISTS directory_summaries (
    tenant_id String COMMENT 'Tenant identifier',
    source_id String COMMENT 'Source connection identifier',
    directory_path String COMMENT 'Directory path',
    
    -- Summary statistics
    total_files UInt64 COMMENT 'Total number of files in directory',
    total_records UInt64 COMMENT 'Total number of records across all files',
    total_size UInt64 COMMENT 'Total size of all files in bytes',
    
    -- Time boundaries
    first_record_at DateTime COMMENT 'Timestamp of first record',
    last_record_at DateTime COMMENT 'Timestamp of last record',
    last_updated DateTime COMMENT 'When this summary was last updated',
    
    -- Field summaries and schema info as JSON
    field_summaries String COMMENT 'JSON of field summaries',
    schema_versions Array(String) COMMENT 'List of schema versions in directory',
    current_schema String COMMENT 'Current/latest schema hash',
    
    -- Indexes
    INDEX idx_tenant_id tenant_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_source_id source_id TYPE bloom_filter GRANULARITY 1,
    INDEX idx_directory_path directory_path TYPE bloom_filter GRANULARITY 1,
    INDEX idx_last_updated last_updated TYPE minmax GRANULARITY 1
) ENGINE = ReplacingMergeTree(last_updated)
ORDER BY (tenant_id, source_id, directory_path)
PARTITION BY tenant_id
SETTINGS index_granularity = 8192;

-- Example Queries that solve your problem:

-- 1. Find all records with name like "sam" in source_connection_01
-- This was IMPOSSIBLE before, now it's fast!
/*
SELECT 
    r.record_id,
    r.name,
    r.file_path,
    r.row_number,
    r.created_at
FROM record_metadata r
WHERE r.tenant_id = 'tenant_123'
    AND r.source_id = 'source_connection_01'
    AND lower(r.name) LIKE '%sam%'
ORDER BY r.created_at DESC
LIMIT 100;
*/

-- 2. Get count of records by status across all files in source_connection_01
/*
SELECT 
    status,
    COUNT(*) as record_count,
    COUNT(DISTINCT file_id) as file_count
FROM record_metadata
WHERE tenant_id = 'tenant_123'
    AND source_id = 'source_connection_01'
GROUP BY status
ORDER BY record_count DESC;
*/

-- 3. Find latest 10 records across all files in a source
/*
SELECT 
    r.record_id,
    r.name,
    r.status,
    r.file_path,
    r.created_at
FROM record_metadata r
WHERE r.tenant_id = 'tenant_123'
    AND r.source_id = 'source_connection_01'
ORDER BY r.created_at DESC
LIMIT 10;
*/

-- 4. Get directory summary for a source
/*
SELECT 
    directory_path,
    total_files,
    total_records,
    total_size,
    last_updated
FROM directory_summaries
WHERE tenant_id = 'tenant_123'
    AND source_id = 'source_connection_01'
ORDER BY directory_path;
*/

-- 5. Find which files contain records matching criteria (for efficient Parquet file access)
/*
SELECT DISTINCT 
    file_path,
    file_id,
    COUNT(*) as matching_records
FROM record_metadata
WHERE tenant_id = 'tenant_123'
    AND source_id = 'source_connection_01'
    AND lower(name) LIKE '%sam%'
GROUP BY file_path, file_id
ORDER BY matching_records DESC;
*/
