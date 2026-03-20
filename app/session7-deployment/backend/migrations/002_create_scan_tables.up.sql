-- Migration: Create scan-related tables
-- Version: 002
-- Description: Tables for EASM scanning functionality

-- ============================================
-- SCAN JOBS TABLE
-- ============================================
-- Tracks scan operations (async job pattern)
CREATE TABLE IF NOT EXISTS scan_jobs (
    id UUID PRIMARY KEY,
    asset_id UUID NOT NULL,
    scan_type VARCHAR(50) NOT NULL 
        CHECK (scan_type IN ('subdomain', 'dns', 'whois', 'port', 'asn', 'ssl','all')),  -- Type of scan
    status VARCHAR(50) NOT NULL 
        CHECK (status IN ('pending', 'running', 'completed', 'failed', 'partial')),
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP NULL,  -- NULL while running
    error TEXT,  -- Error message if failed
    results INTEGER DEFAULT 0,  -- Count of results found
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key to assets table
    CONSTRAINT fk_scan_job_asset 
        FOREIGN KEY (asset_id) 
        REFERENCES assets(id) 
        ON DELETE CASCADE
);

-- Indexes for scan_jobs
CREATE INDEX idx_scan_jobs_asset_id ON scan_jobs(asset_id);
CREATE INDEX idx_scan_jobs_status ON scan_jobs(status);
CREATE INDEX idx_scan_jobs_type_status ON scan_jobs(scan_type, status);
CREATE INDEX idx_scan_jobs_created_at ON scan_jobs(created_at DESC);

-- ============================================
-- SUBDOMAINS TABLE
-- ============================================
-- Stores discovered subdomains
CREATE TABLE IF NOT EXISTS subdomains (
    id UUID PRIMARY KEY,
    asset_id UUID NOT NULL,  -- Parent domain
    scan_job_id UUID NOT NULL,  -- Which scan found this
    name VARCHAR(255) NOT NULL,  -- e.g., "api.example.com"
    source VARCHAR(100) NOT NULL,  -- Discovery method: bruteforce, cert_transparency, web_scraping
    is_active BOOLEAN DEFAULT true,  -- Currently reachable
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign keys
    CONSTRAINT fk_subdomain_asset 
        FOREIGN KEY (asset_id) 
        REFERENCES assets(id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_subdomain_scan_job 
        FOREIGN KEY (scan_job_id) 
        REFERENCES scan_jobs(id) 
        ON DELETE CASCADE,
    
    -- Prevent duplicate subdomains per asset
    CONSTRAINT unique_subdomain_per_asset 
        UNIQUE (asset_id, name)
);

-- Indexes for subdomains
CREATE INDEX idx_subdomains_asset_id ON subdomains(asset_id);
CREATE INDEX idx_subdomains_scan_job_id ON subdomains(scan_job_id);
CREATE INDEX idx_subdomains_name ON subdomains(name);
CREATE INDEX idx_subdomains_is_active ON subdomains(is_active);

-- ============================================
-- DNS RECORDS TABLE
-- ============================================
-- Stores DNS records for domains/subdomains
CREATE TABLE IF NOT EXISTS dns_records (
    id UUID PRIMARY KEY,
    asset_id UUID NOT NULL,  -- Domain or subdomain
    scan_job_id UUID NOT NULL,  -- Which scan discovered this
    record_type VARCHAR(10) NOT NULL,  -- A, AAAA, CNAME, MX, NS, TXT, SOA
    name VARCHAR(255) NOT NULL,  -- Record name
    value TEXT NOT NULL,  -- Record value (can be long for TXT)
    ttl INTEGER,  -- Time to live (seconds)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign keys
    CONSTRAINT fk_dns_record_asset 
        FOREIGN KEY (asset_id) 
        REFERENCES assets(id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_dns_record_scan_job 
        FOREIGN KEY (scan_job_id) 
        REFERENCES scan_jobs(id) 
        ON DELETE CASCADE
);

-- Indexes for dns_records
CREATE INDEX idx_dns_records_asset_id ON dns_records(asset_id);
CREATE INDEX idx_dns_records_scan_job_id ON dns_records(scan_job_id);
CREATE INDEX idx_dns_records_type ON dns_records(record_type);
CREATE INDEX idx_dns_records_name ON dns_records(name);

-- ============================================
-- WHOIS RECORDS TABLE
-- ============================================
-- Stores WHOIS registration information
CREATE TABLE IF NOT EXISTS whois_records (
    id UUID PRIMARY KEY,
    asset_id UUID NOT NULL,  -- Domain only
    scan_job_id UUID NOT NULL,  -- Which scan discovered this
    registrar TEXT,  -- Domain registrar name (can be long)
    created_date TIMESTAMP NULL,  -- Domain registration date
    expiry_date TIMESTAMP NULL,  -- Domain expiration date
    name_servers TEXT,  -- JSON array: ["ns1.example.com", "ns2.example.com"]
    status TEXT,  -- Domain status (multiple statuses can be very long)
    emails TEXT,  -- JSON array: ["admin@example.com"]
    raw_data TEXT NOT NULL,  -- Full WHOIS response for reference
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign keys
    CONSTRAINT fk_whois_record_asset 
        FOREIGN KEY (asset_id) 
        REFERENCES assets(id) 
        ON DELETE CASCADE,
    CONSTRAINT fk_whois_record_scan_job 
        FOREIGN KEY (scan_job_id) 
        REFERENCES scan_jobs(id) 
        ON DELETE CASCADE,
    
    -- One WHOIS record per asset per scan
    CONSTRAINT unique_whois_per_scan 
        UNIQUE (asset_id, scan_job_id)
);

-- Indexes for whois_records
CREATE INDEX idx_whois_records_asset_id ON whois_records(asset_id);
CREATE INDEX idx_whois_records_scan_job_id ON whois_records(scan_job_id);
CREATE INDEX idx_whois_records_expiry_date ON whois_records(expiry_date);  -- Find expiring domains

/*
🎓 TEACHING NOTES - Database Design for EASM

=== TABLE DESIGN PATTERNS ===

1. PARENT-CHILD RELATIONSHIPS:
   
   assets (parent)
     ├── scan_jobs (children - what scans were run)
     ├── subdomains (children - what was discovered)
     ├── dns_records (children - DNS info)
     └── whois_records (children - registration info)
   
   scan_jobs (parent)
     ├── subdomains (children - results from this scan)
     ├── dns_records (children - results from this scan)
     └── whois_records (children - results from this scan)

2. FOREIGN KEY CONSTRAINTS:
   
   Purpose: Data integrity
   - Cannot create subdomain without asset
   - Cannot create scan result without scan job
   
   ON DELETE CASCADE:
   - When asset deleted → all scans/results deleted
   - Prevents orphaned records
   - Automatic cleanup
   
   Example:
   ```sql
   DELETE FROM assets WHERE id = 'xxx';
   -- Automatically deletes:
   --   - All scan_jobs for this asset
   --   - All subdomains for this asset
   --   - All dns_records for this asset
   --   - All whois_records for this asset
   ```

3. UNIQUE CONSTRAINTS:
   
   unique_subdomain_per_asset:
   - Same subdomain can't be added twice for one asset
   - Prevents duplicates from multiple scans
   - Database-level deduplication
   
   unique_whois_per_scan:
   - One WHOIS record per scan per asset
   - Can have multiple WHOIS records for same asset (historical)
   - But not within same scan

=== CHECK CONSTRAINTS ===

Validate data at database level:

```sql
CHECK (scan_type IN ('subdomain', 'dns', 'whois', ...))
```

Benefits:
- ✅ Prevents invalid data even if app has bugs
- ✅ Self-documenting (schema shows valid values)
- ✅ Performance (no need to validate on read)

Alternative approaches:
- ENUM type (PostgreSQL specific)
- Lookup table with foreign key
- Application-level validation only (❌ less safe)

=== NULLABLE vs NOT NULL ===

Strategy:
- NOT NULL: Required fields (name, type, created_at)
- NULL: Optional or conditional fields

Examples:
```sql
ended_at TIMESTAMP NULL      -- NULL while scan running
error TEXT                   -- NULL if successful
expiry_date TIMESTAMP NULL   -- NULL if not parseable
```

In Go:
```go
EndedAt   *time.Time  // Pointer = nullable
Error     string      // Empty string vs NULL (both work)
```

=== DATA TYPES CHOICE ===

1. UUID vs INTEGER:
   - UUID: globally unique, no coordination, harder to guess
   - INTEGER: simpler, sequential, smaller
   - We use UUID for security (can't enumerate all resources)

2. VARCHAR(n) vs TEXT:
   - VARCHAR(255): Limited length (domains, names)
   - TEXT: Unlimited (raw WHOIS data, long TXT records)
   - VARCHAR with limit = validation + documentation

3. BOOLEAN:
   - is_active: true/false, very clear
   - Alternative: status VARCHAR('active'/'inactive')
   - BOOLEAN is simpler when binary choice

4. INTEGER:
   - ttl: DNS time-to-live (seconds, 0-2147483647)
   - results: Count of results (0-n)
   - Could use BIGINT if expecting huge values

5. TIMESTAMP:
   - Date + Time + Timezone aware
   - vs DATE: only date, no time
   - vs TIMESTAMPTZ: explicit timezone (better for distributed systems)

=== INDEXING STRATEGY ===

1. Primary Key (id):
   - Automatic index (unique + fast lookup)
   - PostgreSQL: B-tree index

2. Foreign Keys (asset_id, scan_job_id):
   - Critical for JOIN performance
   - Without index: full table scan (slow!)
   - With index: fast lookup

3. Filter Fields (status, record_type, is_active):
   - Used in WHERE clauses
   - Example: "SELECT * FROM scan_jobs WHERE status = 'running'"
   - Index makes this fast

4. Composite Indexes (scan_type, status):
   - For queries with multiple conditions
   - Example: "WHERE scan_type = 'dns' AND status = 'running'"
   - More specific = faster

5. Sort Fields (created_at DESC):
   - Used in ORDER BY
   - DESC: Pre-sorted in descending order (newest first)
   - Common pattern: show recent items first

Index Trade-offs:
- ✅ Faster SELECT queries
- ❌ Slower INSERT/UPDATE/DELETE (index must be updated)
- ❌ More disk space
- Rule: Index what you query, not everything

=== MIGRATION BEST PRACTICES ===

1. Versioning:
   - 001_create_assets.up.sql (base tables)
   - 002_create_scan_tables.up.sql (new feature)
   - 003_add_column.up.sql (future change)

2. Up/Down Migrations:
   - Up: Apply change
   - Down: Revert change
   - Allows rollback if issues

3. IF NOT EXISTS:
   - Safe to run multiple times
   - Idempotent (same result every time)
   - No error if table already exists

4. Order Matters:
   - Create parent tables first (assets)
   - Then child tables with foreign keys
   - Otherwise: "table does not exist" error

=== QUERYING PATTERNS ===

1. Get all scan jobs for an asset:
```sql
SELECT * FROM scan_jobs 
WHERE asset_id = $1 
ORDER BY created_at DESC;
```

2. Get scan results:
```sql
SELECT s.* FROM subdomains s
JOIN scan_jobs j ON s.scan_job_id = j.id
WHERE j.asset_id = $1;
```

3. Find running scans:
```sql
SELECT * FROM scan_jobs 
WHERE status = 'running' 
AND started_at < NOW() - INTERVAL '1 hour';
-- Find stuck scans (running > 1 hour)
```

4. Dashboard query (composite index used):
```sql
SELECT scan_type, status, COUNT(*) 
FROM scan_jobs 
GROUP BY scan_type, status;
```

5. Find expiring domains:
```sql
SELECT a.name, w.expiry_date 
FROM assets a
JOIN whois_records w ON w.asset_id = a.id
WHERE w.expiry_date < NOW() + INTERVAL '30 days'
AND w.expiry_date > NOW();
-- Domains expiring in next 30 days
```

=== SCALING CONSIDERATIONS ===

For production systems:

1. Partitioning:
   - Partition scan_jobs by created_at (time-series data)
   - Move old scans to archive

2. Archiving:
   - Keep recent scans in main table
   - Move completed scans > 90 days to archive table

3. Materialized Views:
   - Pre-compute dashboard statistics
   - Refresh periodically

4. Read Replicas:
   - Heavy read traffic (dashboards, reports)
   - Route reads to replicas
   - Writes to primary only

Not needed for this teaching project, but good to know!

=== COMPARISON WITH SESSION 3 ===

Session 3: Single table (assets)
Session 5: 
  - 5 tables total (1 + 4 new)
  - Foreign key relationships
  - CASCADE deletes
  - UNIQUE constraints
  - More complex queries (JOIN)

New SQL concepts:
  - FOREIGN KEY with ON DELETE CASCADE
  - UNIQUE constraints (multi-column)
  - NULL vs NOT NULL strategy
  - Composite indexes
  - CHECK constraints

=== DEMO POINTS FOR STUDENTS ===

1. Show CASCADE delete in action:
   - Create asset → create scan → create results
   - Delete asset → show all related records gone

2. Show UNIQUE constraint:
   - Try inserting same subdomain twice
   - Error: constraint violation

3. Show JOIN query:
   - Get asset name + scan results in one query
   - Explain performance difference vs multiple queries

4. Show index usage:
   - EXPLAIN ANALYZE query
   - Show index scan vs sequential scan

5. Show CHECK constraint:
   - Try inserting invalid scan_type
   - Error: check constraint violation

These hands-on demos make concepts stick!
*/
