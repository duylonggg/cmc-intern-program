package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"mini-asm/internal/config"
	"mini-asm/internal/model"
	"mini-asm/internal/storage"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresStorage implements the Storage interface using PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage instance
func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func NewPostgresStorageFromConfig(config *config.PostgresConfig) (*PostgresStorage, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost,
		config.DBPort,
		config.DBUser,
		config.DBPassword,
		config.DBName,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &PostgresStorage{db: db}, nil
}

// Create inserts a new asset into the database
func (p *PostgresStorage) Create(asset *model.Asset) error {
	query := `
		INSERT INTO assets (id, name, type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := p.db.Exec(
		query,
		asset.ID,
		asset.Name,
		asset.Type,
		asset.Status,
		asset.CreatedAt,
		asset.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create asset: %w", err)
	}

	return nil
}

// GetAll retrieves assets with filtering, sorting, and pagination
func (p *PostgresStorage) GetAll(params storage.QueryParams) (*storage.PaginatedResult, error) {
	// Build base query with filters
	query, args := p.buildQuery(params)

	// Add sorting
	query += p.buildOrderBy(params.SortBy, params.SortOrder)

	// Add pagination
	offset := (params.Page - 1) * params.PageSize
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, params.PageSize, offset)

	// Execute query
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query assets: %w", err)
	}
	defer rows.Close()

	// Scan results
	assets := []*model.Asset{}
	for rows.Next() {
		asset := &model.Asset{}
		err := rows.Scan(
			&asset.ID,
			&asset.Name,
			&asset.Type,
			&asset.Status,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, asset)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Get total count for pagination metadata
	total, err := p.Count(params)
	if err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	// Calculate total pages
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	return &storage.PaginatedResult{
		Data:       assets,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Count returns the total number of assets matching the filters
func (p *PostgresStorage) Count(params storage.QueryParams) (int64, error) {
	// Build count query with same filters
	countQuery, args := p.buildCountQuery(params)

	var total int64
	err := p.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to count assets: %w", err)
	}

	return total, nil
}

// buildQuery constructs the SELECT query with filters
func (p *PostgresStorage) buildQuery(params storage.QueryParams) (string, []interface{}) {
	query := `
		SELECT id, name, type, status, created_at, updated_at
		FROM assets
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	// Filter by type
	if params.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, params.Type)
		argCount++
	}

	// Filter by status
	if params.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, params.Status)
		argCount++
	}

	// Search by name (case-insensitive partial match)
	if params.Search != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argCount)
		args = append(args, "%"+params.Search+"%")
		argCount++
	}

	return query, args
}

// buildCountQuery constructs the COUNT query with filters
func (p *PostgresStorage) buildCountQuery(params storage.QueryParams) (string, []interface{}) {
	query := "SELECT COUNT(*) FROM assets WHERE 1=1"

	args := []interface{}{}
	argCount := 1

	// Apply same filters as main query
	if params.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, params.Type)
		argCount++
	}

	if params.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, params.Status)
		argCount++
	}

	if params.Search != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argCount)
		args = append(args, "%"+params.Search+"%")
		argCount++
	}

	return query, args
}

// buildOrderBy constructs the ORDER BY clause
func (p *PostgresStorage) buildOrderBy(sortBy, sortOrder string) string {
	// Default sorting
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Whitelist validation is done in validator, but double-check here
	validFields := map[string]bool{
		"name":       true,
		"type":       true,
		"status":     true,
		"created_at": true,
		"updated_at": true,
	}

	if !validFields[sortBy] {
		sortBy = "created_at"
	}

	sortOrder = strings.ToUpper(sortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	return fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
}

// GetByID retrieves a single asset by its ID
func (p *PostgresStorage) GetByID(id string) (*model.Asset, error) {
	query := `
		SELECT id, name, type, status, created_at, updated_at
		FROM assets
		WHERE id = $1
	`

	asset := &model.Asset{}
	err := p.db.QueryRow(query, id).Scan(
		&asset.ID,
		&asset.Name,
		&asset.Type,
		&asset.Status,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return asset, nil
}

// Update modifies an existing asset in the database
func (p *PostgresStorage) Update(id string, asset *model.Asset) error {
	query := `
		UPDATE assets
		SET name = $1, type = $2, status = $3, updated_at = $4
		WHERE id = $5
	`

	result, err := p.db.Exec(
		query,
		asset.Name,
		asset.Type,
		asset.Status,
		asset.UpdatedAt,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return model.ErrNotFound
	}

	return nil
}

// Delete removes an asset from the database
func (p *PostgresStorage) Delete(id string) error {
	query := `DELETE FROM assets WHERE id = $1`

	result, err := p.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return model.ErrNotFound
	}

	return nil
}

/*
🎓 TEACHING NOTES:

=== SESSION 4 ENHANCEMENTS ===

1. Unified GetAll Method:
   - Replaces separate GetAll(), Filter(), Search() methods
   - Single method handles all query variations
   - More maintainable and flexible

2. Dynamic Query Building:
   - buildQuery() constructs WHERE clause dynamically
   - Only adds conditions for provided parameters
   - Maintains prepared statement security ($1, $2, ...)

3. Query Construction Pattern:
   ```
   WHERE 1=1              ← Always true, simplifies logic
   AND type = $1          ← Add condition if param present
   AND status = $2        ← Add condition if param present
   AND name ILIKE $3      ← Add condition if param present
   ORDER BY created_at DESC
   LIMIT $4 OFFSET $5
   ```

4. Pagination Implementation:
   - LIMIT: how many results to return
   - OFFSET: how many results to skip
   - Offset = (Page - 1) * PageSize
   - Example: Page 2, Size 10 → OFFSET 10

5. Sorting:
   - buildOrderBy() creates ORDER BY clause
   - Whitelist validation (critical for security!)
   - Default: ORDER BY created_at DESC
   - Never concatenate user input directly!

6. Count Query:
   - Separate query to get total count
   - Same filters as main query
   - Used for pagination metadata

7. Parameter Counting:
   - Track $1, $2, $3... positions
   - argCount increments for each parameter
   - Important: LIMIT and OFFSET use next available numbers

SECURITY NOTES:

1. SQL Injection Prevention:
   - ✅ Always use prepared statements ($1, $2)
   - ✅ Whitelist sort fields
   - ✅ Validate parameters before query
   - ❌ Never: fmt.Sprintf("ORDER BY %s", userInput)

2. Defense in Depth:
   - Validation in validator layer
   - Double-check in storage layer
   - Prepared statements in database

EXAMPLE QUERIES:

1. All assets, page 1:
   ```sql
   SELECT ... FROM assets WHERE 1=1
   ORDER BY created_at DESC
   LIMIT 20 OFFSET 0
   ```

2. Filter by type, page 2:
   ```sql
   SELECT ... FROM assets WHERE 1=1 AND type = $1
   ORDER BY created_at DESC
   LIMIT 20 OFFSET 20
   ```
   Args: ["domain", 20, 20]

3. Filter + Search + Sort:
   ```sql
   SELECT ... FROM assets WHERE 1=1
   AND type = $1 AND name ILIKE $2
   ORDER BY name ASC
   LIMIT 10 OFFSET 0
   ```
   Args: ["domain", "%example%", 10, 0]

PERFORMANCE CONSIDERATIONS:

1. Indexes Help:
   - idx_assets_type → speeds up type filter
   - idx_assets_name → speeds up ILIKE search
   - idx_assets_created_at → speeds up default sort

2. ILIKE Performance:
   - Case-insensitive search
   - Can be slow on large datasets
   - Consider full-text search for production

3. Count Query:
   - Can be expensive on large tables
   - Consider caching count results
   - Or estimate with EXPLAIN

COMPARISON WITH SESSION 3:

Session 3:
  - Separate methods for each operation
  - No pagination
  - No sorting options
  - Harder to combine filters

Session 4:
  - Single unified method
  - Built-in pagination
  - Flexible sorting
  - Easy to combine any filters

Example improvement:
  Session 3: Can't get "active domains" with pagination
  Session 4: params := QueryParams{Type: "domain", Status: "active", Page: 1}

TESTING STRATEGIES:

1. Unit tests with mock DB
2. Integration tests with real DB
3. Test edge cases:
   - Empty results
   - Large datasets
   - Special characters in search
   - Invalid sort fields (should use default)

DEMO POINTS:

1. Show query building with fmt.Printf
2. Show parameter array construction
3. Test with different filter combinations
4. Show pagination working (page 1, 2, 3)
5. Show sorting (asc vs desc)
6. Show combined: filter + search + sort + page
*/

// ============================================
// SCAN STORAGE METHODS (Session 5)
// ============================================

// CreateScanJob inserts a new scan job into the database
func (p *PostgresStorage) CreateScanJob(job *model.ScanJob) error {
	query := `
		INSERT INTO scan_jobs (id, asset_id, scan_type, status, started_at, ended_at, error, results, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := p.db.Exec(
		query,
		job.ID,
		job.AssetID,
		job.ScanType,
		job.Status,
		job.StartedAt,
		job.EndedAt,
		job.Error,
		job.Results,
		job.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create scan job: %w", err)
	}

	return nil
}

// GetScanJob retrieves a scan job by ID
func (p *PostgresStorage) GetScanJob(id string) (*model.ScanJob, error) {
	query := `
		SELECT id, asset_id, scan_type, status, started_at, ended_at, error, results, created_at
		FROM scan_jobs
		WHERE id = $1
	`

	job := &model.ScanJob{}
	err := p.db.QueryRow(query, id).Scan(
		&job.ID,
		&job.AssetID,
		&job.ScanType,
		&job.Status,
		&job.StartedAt,
		&job.EndedAt,
		&job.Error,
		&job.Results,
		&job.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scan job: %w", err)
	}

	return job, nil
}

// UpdateScanJob updates an existing scan job
func (p *PostgresStorage) UpdateScanJob(job *model.ScanJob) error {
	query := `
		UPDATE scan_jobs
		SET status = $1, ended_at = $2, error = $3, results = $4
		WHERE id = $5
	`

	result, err := p.db.Exec(
		query,
		job.Status,
		job.EndedAt,
		job.Error,
		job.Results,
		job.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update scan job: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return model.ErrNotFound
	}

	return nil
}

// ListScanJobsByAsset retrieves all scan jobs for an asset
func (p *PostgresStorage) ListScanJobsByAsset(assetID string) ([]*model.ScanJob, error) {
	query := `
		SELECT id, asset_id, scan_type, status, started_at, ended_at, error, results, created_at
		FROM scan_jobs
		WHERE asset_id = $1
		ORDER BY created_at DESC
	`

	rows, err := p.db.Query(query, assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query scan jobs: %w", err)
	}
	defer rows.Close()

	jobs := []*model.ScanJob{}
	for rows.Next() {
		job := &model.ScanJob{}
		err := rows.Scan(
			&job.ID,
			&job.AssetID,
			&job.ScanType,
			&job.Status,
			&job.StartedAt,
			&job.EndedAt,
			&job.Error,
			&job.Results,
			&job.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// CreateSubdomain inserts a new subdomain into the database
func (p *PostgresStorage) CreateSubdomain(subdomain *model.Subdomain) error {
	query := `
		INSERT INTO subdomains (id, asset_id, scan_job_id, name, source, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (asset_id, name) DO UPDATE SET
			scan_job_id = EXCLUDED.scan_job_id,
			is_active = EXCLUDED.is_active
	`

	_, err := p.db.Exec(
		query,
		subdomain.ID,
		subdomain.AssetID,
		subdomain.ScanJobID,
		subdomain.Name,
		subdomain.Source,
		subdomain.IsActive,
		subdomain.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create subdomain: %w", err)
	}

	return nil
}

// GetSubdomainsByAsset retrieves all subdomains for an asset
func (p *PostgresStorage) GetSubdomainsByAsset(assetID string) ([]*model.Subdomain, error) {
	query := `
		SELECT id, asset_id, scan_job_id, name, source, is_active, created_at
		FROM subdomains
		WHERE asset_id = $1
		ORDER BY created_at DESC
	`

	return p.querySubdomains(query, assetID)
}

// GetSubdomainsByScan retrieves all subdomains discovered by a scan
func (p *PostgresStorage) GetSubdomainsByScan(scanJobID string) ([]*model.Subdomain, error) {
	query := `
		SELECT id, asset_id, scan_job_id, name, source, is_active, created_at
		FROM subdomains
		WHERE scan_job_id = $1
		ORDER BY created_at DESC
	`

	return p.querySubdomains(query, scanJobID)
}

// querySubdomains is a helper method shared by subdomain query methods
func (p *PostgresStorage) querySubdomains(query string, arg string) ([]*model.Subdomain, error) {
	rows, err := p.db.Query(query, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to query subdomains: %w", err)
	}
	defer rows.Close()

	subdomains := []*model.Subdomain{}
	for rows.Next() {
		subdomain := &model.Subdomain{}
		err := rows.Scan(
			&subdomain.ID,
			&subdomain.AssetID,
			&subdomain.ScanJobID,
			&subdomain.Name,
			&subdomain.Source,
			&subdomain.IsActive,
			&subdomain.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subdomain: %w", err)
		}
		subdomains = append(subdomains, subdomain)
	}

	return subdomains, nil
}

// CreateDNSRecord inserts a new DNS record into the database
func (p *PostgresStorage) CreateDNSRecord(record *model.DNSRecord) error {
	query := `
		INSERT INTO dns_records (id, asset_id, scan_job_id, record_type, name, value, ttl, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := p.db.Exec(
		query,
		record.ID,
		record.AssetID,
		record.ScanJobID,
		record.RecordType,
		record.Name,
		record.Value,
		record.TTL,
		record.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create DNS record: %w", err)
	}

	return nil
}

// GetDNSRecordsByAsset retrieves all DNS records for an asset
func (p *PostgresStorage) GetDNSRecordsByAsset(assetID string) ([]*model.DNSRecord, error) {
	query := `
		SELECT id, asset_id, scan_job_id, record_type, name, value, ttl, created_at
		FROM dns_records
		WHERE asset_id = $1
		ORDER BY created_at DESC, record_type
	`

	return p.queryDNSRecords(query, assetID)
}

// GetDNSRecordsByScan retrieves all DNS records discovered by a scan
func (p *PostgresStorage) GetDNSRecordsByScan(scanJobID string) ([]*model.DNSRecord, error) {
	query := `
		SELECT id, asset_id, scan_job_id, record_type, name, value, ttl, created_at
		FROM dns_records
		WHERE scan_job_id = $1
		ORDER BY created_at DESC, record_type
	`

	return p.queryDNSRecords(query, scanJobID)
}

// queryDNSRecords is a helper method shared by DNS record query methods
func (p *PostgresStorage) queryDNSRecords(query string, arg string) ([]*model.DNSRecord, error) {
	rows, err := p.db.Query(query, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to query DNS records: %w", err)
	}
	defer rows.Close()

	records := []*model.DNSRecord{}
	for rows.Next() {
		record := &model.DNSRecord{}
		err := rows.Scan(
			&record.ID,
			&record.AssetID,
			&record.ScanJobID,
			&record.RecordType,
			&record.Name,
			&record.Value,
			&record.TTL,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan DNS record: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// CreateWHOISRecord inserts a new WHOIS record into the database
func (p *PostgresStorage) CreateWHOISRecord(record *model.WHOISRecord) error {
	query := `
		INSERT INTO whois_records (id, asset_id, scan_job_id, registrar, created_date, expiry_date, 
									name_servers, status, emails, raw_data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (asset_id, scan_job_id) DO UPDATE SET
			registrar = EXCLUDED.registrar,
			created_date = EXCLUDED.created_date,
			expiry_date = EXCLUDED.expiry_date,
			name_servers = EXCLUDED.name_servers,
			status = EXCLUDED.status,
			emails = EXCLUDED.emails,
			raw_data = EXCLUDED.raw_data
	`

	_, err := p.db.Exec(
		query,
		record.ID,
		record.AssetID,
		record.ScanJobID,
		record.Registrar,
		record.CreatedDate,
		record.ExpiryDate,
		record.NameServers,
		record.Status,
		record.Emails,
		record.RawData,
		record.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create WHOIS record: %w", err)
	}

	return nil
}

// GetWHOISRecordByAsset retrieves the latest WHOIS record for an asset
func (p *PostgresStorage) GetWHOISRecordByAsset(assetID string) (*model.WHOISRecord, error) {
	query := `
		SELECT id, asset_id, scan_job_id, registrar, created_date, expiry_date, 
		       name_servers, status, emails, raw_data, created_at
		FROM whois_records
		WHERE asset_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	record := &model.WHOISRecord{}
	err := p.db.QueryRow(query, assetID).Scan(
		&record.ID,
		&record.AssetID,
		&record.ScanJobID,
		&record.Registrar,
		&record.CreatedDate,
		&record.ExpiryDate,
		&record.NameServers,
		&record.Status,
		&record.Emails,
		&record.RawData,
		&record.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get WHOIS record: %w", err)
	}

	return record, nil
}

// GetWHOISRecordsByScan retrieves all WHOIS records discovered by a scan
func (p *PostgresStorage) GetWHOISRecordsByScan(scanJobID string) ([]*model.WHOISRecord, error) {
	query := `
		SELECT id, asset_id, scan_job_id, registrar, created_date, expiry_date, 
		       name_servers, status, emails, raw_data, created_at
		FROM whois_records
		WHERE scan_job_id = $1
		ORDER BY created_at DESC
	`

	rows, err := p.db.Query(query, scanJobID)
	if err != nil {
		return nil, fmt.Errorf("failed to query WHOIS records: %w", err)
	}
	defer rows.Close()

	records := []*model.WHOISRecord{}
	for rows.Next() {
		record := &model.WHOISRecord{}
		err := rows.Scan(
			&record.ID,
			&record.AssetID,
			&record.ScanJobID,
			&record.Registrar,
			&record.CreatedDate,
			&record.ExpiryDate,
			&record.NameServers,
			&record.Status,
			&record.Emails,
			&record.RawData,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan WHOIS record: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

/*
🎓 NOTES - Scan Storage (Session 5)

=== KEY PATTERNS ===

1. **UPSERT with ON CONFLICT**:
   
   Problem: Subdomain might be discovered multiple times
   Solution: ON CONFLICT (asset_id, name) DO UPDATE
   
   ```sql
   INSERT INTO subdomains (...)
   VALUES (...)
   ON CONFLICT (asset_id, name) DO UPDATE SET
       scan_job_id = EXCLUDED.EXCLUDED.scan_job_id,
       is_active = EXCLUDED.is_active
   ```
   
   Behavior:
   - First time: INSERT new subdomain
   - Second time: UPDATE existing subdomain
   - No duplicate key error!
   
   Use cases:
   - Subdomain found again → update scan_job_id, is_active
   - WHOIS record for same asset/scan → update with new data

2. **Helper Methods for DRY**:
   
   ```go
   func (p *PostgresStorage) GetSubdomainsByAsset(assetID) { ... }
   func (p *PostgresStorage) GetSubdomainsByScan(scanJobID) { ... }
   // Both use:
   func (p *PostgresStorage) querySubdomains(query, arg) { ... }
   ```
   
   Benefits:
   - Avoid duplicating scan logic
   - Single place to fix bugs
   - Consistent error handling

3. **Nullable Fields**:
   
   ```go
   &job.EndedAt,    // *time.Time (nullable)
   &job.Error,      // string (PostgreSQL NULL → empty string)
   ```
   
   PostgreSQL NULL handling:
   - Pointer types (*time.Time): SQL NULL → Go nil
   - String types: SQL NULL → Go empty string
   - No special handling needed!

4. **Ordering**:
   
   ```sql
   ORDER BY created_at DESC  -- Newest first
   ORDER BY created_at DESC, record_type  -- Newest, then by type
   ```
   
   Common patterns:
   - Scan jobs: Most recent first
   - Results: Group by type, newest first

=== INTEGRATION WITH SCANNERS ===

Typical flow:

```go
// 1. Create scan job
job := &model.ScanJob{
    ID: uuid.New().String(),
    AssetID: asset.ID,
    ScanType: model.ScanTypeDNS,
    Status: model.ScanStatusPending,
}
storage.CreateScanJob(job)

// 2. Update to running
job.Status = model.ScanStatusRunning
storage.UpdateScanJob(job)

// 3. Perform scan
scanner := NewDNSScanner()
records, err := scanner.Scan(asset)

// 4. Save results
for _, record := range records {
    record.AssetID = asset.ID
    record.ScanJobID = job.ID
    storage.CreateDNSRecord(record)
}

// 5. Update job to completed
job.Status = model.ScanStatusCompleted
job.Results = len(records)
now := time.Now()
job.EndedAt = &now
storage.UpdateScanJob(job)
```

=== ERROR HANDLING STRATEGIES ===

1. **Scan Failure**:
   ```go
   if err != nil {
       job.Status = model.ScanStatusFailed
       job.Error = err.Error()
       storage.UpdateScanJob(job)
   }
   ```

2. **Partial Results**:
   ```go
   if len(results) == 0 {
       job.Status = model.ScanStatusPartial
       job.Error = "no results found"
   }
   ```

3. **Database Errors**:
   - Wrap errors: fmt.Errorf("...: %w", err)
   - Provides context
   - Helps debugging

=== QUERY OPTIMIZATION ===

1. **Indexes Used** (from migration):
   - `idx_scan_jobs_asset_id` → Fast: WHERE asset_id = ?
   - `idx_subdomains_scan_job_id` → Fast: WHERE scan_job_id = ?
   - `idx_dns_records_asset_id` → Fast: WHERE asset_id = ?

2. **Without Indexes**:
   ```
   100,000 scan jobs → Sequential scan (slow!)
   With index → Index scan (fast!)
   ```

3 **Foreign Key Benefits**:
   - Integrity: Can't create result without valid asset/job
   - CASCADE: Delete asset → all scan data deleted
   - JOIN optimization: Database knows relationship

=== COMPARISON WITH SESSION 4 ===

Session 4: Single table (assets)
Session 5: 
  - Multiple related tables
  - Foreign key relationships
  - UPSERT operations
  - Helper methods for DRY

New SQL concepts:
  - ON CONFLICT (UPSERT)
  - Helper query methods
  - Nullable field handling
  - Multi-table queries

=== STUDENT EXERCISES ===

1. Add Batch Insert:
   ```go
   func (p *PostgresStorage) CreateDNSRecordsBatch(records []*DNSRecord) error
   // More efficient than loop of individual inserts
   ```

2. Add Scan Statistics:
   ```go
   func (p *PostgresStorage) GetScanStats(assetID string) (*ScanStats, error)
   // Count: subdomains, DNS records, etc.
   // Aggregate data for dashboard
   ```

3. Add Historical Comparison:
   ```go
   func (p *PostgresStorage) CompareDNSRecords(oldScanID, newScanID string) (*DNSChanges, error)
   // Find: Added, removed, modified records
   ```

4. Add Search:
   ```go
   func (p *PostgresStorage) SearchSubdomains(pattern string) ([]*Subdomain, error)
   // Search across all assets
   ```

These methods follow the same patterns - students can implement!

=== KEY TAKEAWAYS ===

1. UPSERT prevents duplicates (ON CONFLICT)
2. Helper methods reduce duplication
3. Nullable fields use pointers
4. Foreign keys maintain integrity
5. Indexes speed up queries
6. Error wrapping aids debugging
7. Consistent patterns across methods

Storage layer is the foundation - get it right and everything else is easier!
*/

