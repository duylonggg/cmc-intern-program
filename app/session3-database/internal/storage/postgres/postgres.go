package postgres

import (
	"database/sql"
	"fmt"
	"log"

	"mini-asm/internal/config"
	"mini-asm/internal/model"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresStorage implements the Storage interface using PostgreSQL
// This is a concrete implementation that can be swapped with MemoryStorage
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

// GetAll retrieves all assets from the database
func (p *PostgresStorage) GetAll() ([]*model.Asset, error) {
	query := `
		SELECT id, name, type, status, created_at, updated_at
		FROM assets
		ORDER BY created_at DESC
	`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query assets: %w", err)
	}
	defer rows.Close()

	var assets []*model.Asset
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

	return assets, nil
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

// Filter returns assets matching the given criteria
func (p *PostgresStorage) Filter(assetType, status string) ([]*model.Asset, error) {
	query := `
		SELECT id, name, type, status, created_at, updated_at
		FROM assets
		WHERE 1=1
	`

	var args []interface{}
	argCount := 1

	if assetType != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, assetType)
		argCount++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	query += " ORDER BY created_at DESC"

	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to filter assets: %w", err)
	}
	defer rows.Close()

	var assets []*model.Asset
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

	return assets, nil
}

// Search finds assets by partial name match
func (p *PostgresStorage) Search(query string) ([]*model.Asset, error) {
	sqlQuery := `
		SELECT id, name, type, status, created_at, updated_at
		FROM assets
		WHERE name ILIKE $1
		ORDER BY created_at DESC
		LIMIT 100
	`

	// Add wildcards for partial matching
	searchPattern := "%" + query + "%"

	rows, err := p.db.Query(sqlQuery, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search assets: %w", err)
	}
	defer rows.Close()

	var assets []*model.Asset
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

	return assets, nil
}

// GetStats returns aggregated statistics about all assets
func (p *PostgresStorage) GetStats() (*model.Stats, error) {
	stats := &model.Stats{
		ByType:   make(map[string]int),
		ByStatus: make(map[string]int),
	}

	// Total count
	if err := p.db.QueryRow(`SELECT COUNT(*) FROM assets`).Scan(&stats.Total); err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	// Count by type
	rows, err := p.db.Query(`SELECT type, COUNT(*) FROM assets GROUP BY type`)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats by type: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		var count int
		if err := rows.Scan(&t, &count); err != nil {
			return nil, fmt.Errorf("failed to scan type row: %w", err)
		}
		stats.ByType[t] = count
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating type rows: %w", err)
	}

	// Count by status
	rows2, err := p.db.Query(`SELECT status, COUNT(*) FROM assets GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats by status: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var s string
		var count int
		if err := rows2.Scan(&s, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status row: %w", err)
		}
		stats.ByStatus[s] = count
	}
	if err = rows2.Err(); err != nil {
		return nil, fmt.Errorf("error iterating status rows: %w", err)
	}

	return stats, nil
}

// Count returns the number of assets matching the optional type and status filters
func (p *PostgresStorage) Count(assetType, status string) (int, error) {
	query := `SELECT COUNT(*) FROM assets WHERE 1=1`
	var args []interface{}
	argCount := 1

	if assetType != "" {
		query += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, assetType)
		argCount++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	var count int
	if err := p.db.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count assets: %w", err)
	}

	return count, nil
}

// BatchCreate inserts multiple assets in a single transaction (all-or-nothing)
func (p *PostgresStorage) BatchCreate(assets []*model.Asset) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && rbErr.Error() != "sql: transaction has already been committed or rolled back" {
			log.Printf("warning: transaction rollback error: %v", rbErr)
		}
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO assets (id, name, type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, asset := range assets {
		if _, err := stmt.Exec(
			asset.ID,
			asset.Name,
			asset.Type,
			asset.Status,
			asset.CreatedAt,
			asset.UpdatedAt,
		); err != nil {
			return fmt.Errorf("failed to insert asset %q: %w", asset.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BatchDelete removes assets by IDs; IDs that don't exist are silently skipped.
// Returns the count of deleted assets and the count of IDs not found.
func (p *PostgresStorage) BatchDelete(ids []string) (deleted int, notFound int, err error) {
	if len(ids) == 0 {
		return 0, 0, nil
	}

	for _, id := range ids {
		result, execErr := p.db.Exec(`DELETE FROM assets WHERE id = $1`, id)
		if execErr != nil {
			return deleted, notFound, fmt.Errorf("failed to delete asset %q: %w", id, execErr)
		}
		rows, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return deleted, notFound, fmt.Errorf("failed to get affected rows for asset %q: %w", id, rowsErr)
		}
		if rows == 0 {
			notFound++
		} else {
			deleted++
		}
	}

	return deleted, notFound, nil
}

// ListPaginated retrieves a page of assets with optional type/status filters.
// Returns the matching assets, the total count of matching assets, and any error.
func (p *PostgresStorage) ListPaginated(page, limit int, assetType, status string) ([]*model.Asset, int, error) {
	where := " WHERE 1=1"
	var args []interface{}
	argCount := 1

	if assetType != "" {
		where += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, assetType)
		argCount++
	}

	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	// Total matching count
	var total int
	if err := p.db.QueryRow("SELECT COUNT(*) FROM assets"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count assets: %w", err)
	}

	// Paginated rows
	offset := (page - 1) * limit
	dataArgs := append(args, limit, offset)
	query := fmt.Sprintf(
		"SELECT id, name, type, status, created_at, updated_at FROM assets%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		where, argCount, argCount+1,
	)

	rows, err := p.db.Query(query, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list assets: %w", err)
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		asset := &model.Asset{}
		if err := rows.Scan(
			&asset.ID,
			&asset.Name,
			&asset.Type,
			&asset.Status,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan asset: %w", err)
		}
		assets = append(assets, asset)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return assets, total, nil
}

/*
🎓 TEACHING NOTES:

1. Interface Implementation:
   - PostgresStorage implements Storage interface
   - Same methods as MemoryStorage
   - Main.go CHỈ CẦN THAY 1 DÒNG!

2. SQL Queries:
   - Parameterized queries ($1, $2) prevent SQL injection
   - NEVER concatenate user input into SQL!
   - ❌ BAD: query := "SELECT * FROM assets WHERE id = '" + id + "'"
   - ✅ GOOD: query := "SELECT * FROM assets WHERE id = $1"

3. Error Handling:
   - Check sql.ErrNoRows → return model.ErrNotFound
   - Wrap errors with fmt.Errorf("context: %w", err)
   - RowsAffected() to verify UPDATE/DELETE success

4. Scanning Rows:
   - rows.Scan() maps columns to struct fields
   - ORDER MATTERS! Must match SELECT order
   - Don't forget defer rows.Close()

5. Dynamic Query Building:
   - Filter() builds query dynamically based on parameters
   - Track parameter count for $1, $2, $3...
   - WHERE 1=1 trick for easier dynamic AND conditions

6. ILIKE vs LIKE:
   - LIKE: case-sensitive
   - ILIKE: case-insensitive (PostgreSQL specific)
   - % wildcards for partial matching

7. Connection Pool:
   - sql.DB maintains connection pool automatically
   - db.Exec(), db.Query(), db.QueryRow() reuse connections
   - Don't close db in storage methods!

8. Transaction Support (Buổi 4):
   - Current: each operation is auto-committed
   - Future: db.Begin() for multi-step operations

COMPARISON: Memory vs Postgres

MemoryStorage:
- data := make(map[string]*model.Asset)
- Fast: O(1) lookups

PostgresStorage:
- data in database on disk
- Slightly slower but persistent
- Can handle millions of records
- Support for advanced queries (JOIN, aggregation)

KEY POINT: Clean Architecture cho phép swap giữa 2 implementations này
mà KHÔNG THAY ĐỔI business logic!
*/
