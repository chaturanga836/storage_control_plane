# Cross-File Query Solution: Solving the ClickHouse + Parquet Limitation

## 🎯 **The Problem You Described**

**ClickHouse can query individual Parquet files but cannot bring cumulative results across multiple files in a directory.**

For example, these queries were **IMPOSSIBLE** before:
- ❌ "Find all records with name like 'sam' in source_connection_01" (across multiple Parquet files)
- ❌ "Get count of records by status across all files in a directory"  
- ❌ "Find latest 10 records across multiple Parquet files"

## 🏗️ **The Solution: Metadata Tables in ClickHouse**

We solve this by maintaining **3 metadata tables** in ClickHouse that contain searchable indexes of your Parquet file contents:

### 1. **`record_metadata` Table** (The Key Solution!)
```sql
CREATE TABLE record_metadata (
    record_id String,
    tenant_id String,
    source_id String,
    file_id String,
    file_path String,
    
    -- Searchable fields (THE MAGIC!)
    name String,           -- Enables "name LIKE '%sam%'" queries
    email String,          -- Enables email searches
    status String,         -- Enables status filtering
    category String,       -- Enables category filtering
    
    -- Timestamps for temporal queries
    created_at DateTime,
    updated_at DateTime,
    
    -- Location in Parquet file
    row_number UInt64,     -- For retrieving full record
    
    -- Indexes for fast searching
    INDEX idx_name name TYPE bloom_filter GRANULARITY 1,
    INDEX idx_status status TYPE set GRANULARITY 1
) ENGINE = MergeTree()
ORDER BY (tenant_id, source_id, created_at);
```

### 2. **`parquet_file_metadata` Table**
- Contains metadata about each Parquet file
- Enables query optimization (which files to check)
- Stores file statistics and boundaries

### 3. **`directory_summaries` Table**  
- Contains aggregated statistics per directory
- Provides quick overview without scanning individual records

## 🔄 **How It Works**

### **Data Writing Process:**
```
1. API data arrives → 
2. Write to Parquet file → 
3. SIMULTANEOUSLY populate ClickHouse metadata tables →
4. ClickHouse can now search across ALL files!
```

### **Query Process:**
```
1. User query: "find name like 'sam' in source_connection_01" →
2. ClickHouse searches record_metadata table (FAST!) →
3. Returns matching records with file locations →
4. Optional: Retrieve full records from specific Parquet files
```

## 🚀 **Now These Queries Are POSSIBLE and FAST!**

### ✅ **1. Find name like "sam" across all files**
```sql
SELECT name, file_path, row_number, created_at
FROM record_metadata 
WHERE tenant_id = 'tenant_123'
    AND source_id = 'source_connection_01'  
    AND lower(name) LIKE '%sam%'
ORDER BY created_at DESC;
```

### ✅ **2. Get latest 10 records across all files**
```sql
SELECT name, status, file_path, created_at
FROM record_metadata
WHERE tenant_id = 'tenant_123'
    AND source_id = 'source_connection_01'
ORDER BY created_at DESC
LIMIT 10;
```

### ✅ **3. Count records by status across all files**
```sql
SELECT status, COUNT(*) as count
FROM record_metadata
WHERE tenant_id = 'tenant_123'
    AND source_id = 'source_connection_01'
GROUP BY status;
```

### ✅ **4. Complex multi-criteria search**
```sql
SELECT name, email, file_path, created_at
FROM record_metadata
WHERE tenant_id = 'tenant_123'
    AND source_id = 'source_connection_01'
    AND lower(name) LIKE '%sam%'
    AND status = 'active'
    AND created_at >= '2024-01-01'
ORDER BY created_at DESC;
```

## 💻 **Implementation Components**

### **1. Enhanced Parquet Writer** (`metadata_parquet_writer.go`)
- Writes Parquet files AND populates metadata tables
- Extracts searchable fields during write process
- Maintains statistics and indexes

### **2. Cross-File Query Service** (`cross_file_query.go`)
- Provides high-level API for cross-file searches
- Handles complex query building
- Returns file hints for optimization

### **3. ClickHouse Schema** (`parquet_metadata_schema.sql`)
- Defines metadata tables with proper indexes
- Optimized for fast searching and aggregation

## 🎯 **Your Specific Use Cases - Now Solved!**

### **Use Case 1: Get all source connections under tenant**
```go
// Already solved by existing source_connections table
records, err := queryService.SearchRecordsAcrossFiles(ctx, &CrossFileSearchRequest{
    TenantID: "tenant_123",
    Limit:    100,
})
```

### **Use Case 2: Get latest 10 source connections**
```go
records, err := queryService.GetLatestRecords(ctx, "tenant_123", "", 10)
```

### **Use Case 3: Find all which contain name like 'sam' under particular tenant**
```go
records, err := queryService.FindNameLikeSam(ctx, "tenant_123", "source_connection_01")
```

### **Use Case 4: Get latest updated source connections under tenant**
```go
req := &CrossFileSearchRequest{
    TenantID: "tenant_123",
    SortBy: []models.SortField{{Field: "updated_at", Order: models.SortDesc}},
    Limit: 20,
}
result, err := queryService.SearchRecordsAcrossFiles(ctx, req)
```

## 🏆 **Key Benefits**

### **Performance:**
- ✅ **Fast searches** with ClickHouse indexes
- ✅ **Query optimization** - only checks relevant files  
- ✅ **Aggregations** across multiple files are now possible

### **Functionality:**
- ✅ **Cross-file text search** ("name like 'sam'")
- ✅ **Complex filtering** (multiple criteria)
- ✅ **Temporal queries** (date ranges)
- ✅ **Aggregations** (count by status, etc.)

### **Scalability:**
- ✅ **Works with thousands of Parquet files**
- ✅ **Efficient storage** (metadata is much smaller than full data)
- ✅ **Partition pruning** for large datasets

## 🛠️ **Implementation Steps**

### **Step 1: Deploy ClickHouse Schema**
```bash
# Run the metadata schema setup
clickhouse-client < sql/parquet_metadata_schema.sql
```

### **Step 2: Update Data Writing Process**
```go
// Replace regular Parquet writer with metadata-aware writer
metadataWriter := writers.NewMetadataParquetWriter(clickhouseClient, config)
fileMetadata, err := metadataWriter.WriteParquetWithMetadata(ctx, records, tenantID, sourceID, dirConfig)
```

### **Step 3: Use Cross-File Query Service**
```go
// Initialize query service  
queryService := clickhouse.NewCrossFileQueryService(clickhouseClient)

// Now you can search across all files!
results, err := queryService.SearchRecordsAcrossFiles(ctx, searchRequest)
```

## 📊 **Example Output**

```bash
🔍 Find all records with name containing 'sam' in source_connection_01
📊 Found 4 records with 'sam' in name across all Parquet files:
   - Samuel Johnson (file: /data/tenant_123/source_connection_01/2024/01/15/data_001.parquet, row: 0)
   - Samantha Davis (file: /data/tenant_123/source_connection_01/2024/01/15/data_002.parquet, row: 1)  
   - Sam Rodriguez (file: /data/tenant_123/source_connection_01/2024/01/15/data_002.parquet, row: 3)
   - Sammy Thompson (file: /data/tenant_123/source_connection_01/2024/01/15/data_003.parquet, row: 2)

⚡ Query executed in: 15ms
📁 Relevant files checked: 3
```

## 🎉 **Result: Problem Solved!**

**Before:** ❌ "Cannot find name like 'sam' across multiple Parquet files"

**After:** ✅ "Fast cross-file search with full metadata indexing"

The metadata approach transforms your Parquet files from isolated data islands into a unified, searchable data lake with ClickHouse as the query engine!
