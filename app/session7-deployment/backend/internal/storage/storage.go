package storage

import "mini-asm/internal/model"

// QueryParams contains all query parameters for listing assets
type QueryParams struct {
	// Pagination
	Page     int
	PageSize int

	// Filtering
	Type   string
	Status string
	Search string

	// Sorting
	SortBy    string
	SortOrder string
}

// PaginatedResult contains paginated data and metadata
type PaginatedResult struct {
	Data       []*model.Asset `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// Storage defines the interface for data access operations
// This is the contract that any storage implementation must fulfill
//
// Why interface?
// - Allows multiple implementations (memory, postgres, mongodb, etc.)
// - Makes testing easy (can mock)
// - Follows Dependency Inversion Principle
type Storage interface {
	// Create adds a new asset to storage
	Create(asset *model.Asset) error

	// GetAll retrieves all assets with optional filtering, sorting, and pagination
	// This is the enhanced version that replaces the old GetAll, Filter, and Search
	GetAll(params QueryParams) (*PaginatedResult, error)

	// GetByID retrieves a single asset by its ID
	// Returns ErrNotFound if asset doesn't exist
	GetByID(id string) (*model.Asset, error)

	// Update modifies an existing asset
	// Returns ErrNotFound if asset doesn't exist
	Update(id string, asset *model.Asset) error

	// Delete removes an asset from storage
	// Returns ErrNotFound if asset doesn't exist
	Delete(id string) error

	// Count returns the total number of assets matching the filters
	Count(params QueryParams) (int64, error)
}

// ScanStorage defines the interface for scan-related data access operations
type ScanStorage interface {
	// Scan Job operations
	CreateScanJob(job *model.ScanJob) error
	GetScanJob(id string) (*model.ScanJob, error)
	UpdateScanJob(job *model.ScanJob) error
	ListScanJobsByAsset(assetID string) ([]*model.ScanJob, error)

	// Subdomain operations
	CreateSubdomain(subdomain *model.Subdomain) error
	GetSubdomainsByAsset(assetID string) ([]*model.Subdomain, error)
	GetSubdomainsByScan(scanJobID string) ([]*model.Subdomain, error)

	// DNS Record operations
	CreateDNSRecord(record *model.DNSRecord) error
	GetDNSRecordsByAsset(assetID string) ([]*model.DNSRecord, error)
	GetDNSRecordsByScan(scanJobID string) ([]*model.DNSRecord, error)

	// WHOIS Record operations
	CreateWHOISRecord(record *model.WHOISRecord) error
	GetWHOISRecordByAsset(assetID string) (*model.WHOISRecord, error)
	GetWHOISRecordsByScan(scanJobID string) ([]*model.WHOISRecord, error)
}

/*
🎓 NOTES - Session 5:

=== SESSION 4 ENHANCEMENTS ===

1. QueryParams Struct:
   - Consolidates all query parameters in one place
   - Easier to pass around than many individual parameters
   - Extensible: add new filters without changing signature

   Before (Session 3):
     Filter(assetType, status string) // Only 2 filters
     Search(query string)               // Separate method

   After (Session 4):
     GetAll(params QueryParams)         // Everything unified!

2. PaginatedResult:
   - Contains both data and metadata
   - Total: for "Showing 1-10 of 245"
   - TotalPages: for pagination UI
   - JSON tags for API response

3. Why Combine Methods?
   Session 3 had: GetAll(), Filter(), Search()
   Session 4 has: GetAll(params)

   Benefits:
   - Single method handles all cases
   - Easier to combine filters (type + search + pagination)
   - Consistent interface
   - Less code duplication

4. Pagination Math:
   - Offset = (Page - 1) * PageSize
   - Example: Page 2, PageSize 10 → Offset 10
   - TotalPages = ceil(Total / PageSize)

5. Default Values:
   - Page 1 if not specified
   - PageSize 20 if not specified
   - SortBy "created_at" if not specified
   - SortOrder "desc" if not specified

6. Filtering Logic:
   - Empty string = ignore filter
   - Type "domain" = filter by type
   - Status "active" = filter by status
   - Search "example" = partial name match
   - Can combine multiple filters!

EXAMPLE QUERIES:

1. Simple pagination:
   QueryParams{Page: 1, PageSize: 10}
   → First 10 assets

2. Filter + pagination:
   QueryParams{Type: "domain", Page: 1, PageSize: 10}
   → First 10 domains

3. Search + filter + sort:
   QueryParams{
     Search: "example",
     Type: "domain",
     SortBy: "name",
     SortOrder: "asc",
   }
   → Domains matching "example", sorted by name A-Z

4. Everything combined:
   QueryParams{
     Type: "domain",
     Status: "active",
     Search: "google",
     Page: 2,
     PageSize: 20,
     SortBy: "created_at",
     SortOrder: "desc",
   }
   → Active domains matching "google", page 2, newest first

IMPLEMENTATION TIPS:

PostgreSQL:
  SELECT * FROM assets
  WHERE type = $1 AND status = $2 AND name ILIKE $3
  ORDER BY created_at DESC
  LIMIT $4 OFFSET $5

Memory (in-memory):
  1. Filter slice
  2. Sort slice
  3. Calculate pagination
  4. Slice [offset:offset+limit]

COMPARISON:

Session 3:
  storage.GetAll()                    // All assets
  storage.Filter("domain", "active")  // Filtered, no pagination
  storage.Search("example")           // Search only

Session 4:
  storage.GetAll(QueryParams{})                        // All assets with pagination
  storage.GetAll(QueryParams{Type: "domain"})          // Filtered + paginated
  storage.GetAll(QueryParams{Search: "example"})       // Searched + paginated
  storage.GetAll(QueryParams{Type: "domain", Page: 2}) // Everything combined!

KEY IMPROVEMENT: Single flexible method instead of multiple rigid methods!
*/

/*
🎓 NOTES:

1. Interface Design:
   - Define behavior, not implementation
   - Methods should be atomic and clear
   - Return errors, don't panic

2. Why Pointers?
   - []*model.Asset vs []model.Asset
   - Pointers avoid copying large structs
   - Allows modification through reference
   - More memory efficient for large datasets

3. Error Handling:
   - Return error as last value
   - Use model.ErrNotFound for consistency
   - Caller decides how to handle

4. Method Signatures:
   - Create(asset *model.Asset) - pointer: will be modified (ID, timestamps)
   - GetByID(id string) - string: immutable lookup
   - Filter/Search - flexible parameters

5. Interface Benefits:

   Buổi 2: MemoryStorage implements this
   type MemoryStorage struct { ... }
   func (m *MemoryStorage) Create(asset *model.Asset) error { ... }

   Buổi 3: PostgresStorage implements the SAME interface
   type PostgresStorage struct { ... }
   func (p *PostgresStorage) Create(asset *model.Asset) error { ... }

   Service layer doesn't change!
   type AssetService struct {
       storage Storage  // Works with ANY implementation!
   }

6. Testing Benefits:
   type MockStorage struct { ... }
   func (m *MockStorage) Create(asset *model.Asset) error {
       return nil // or test-specific behavior
   }

📝 COMPARISON:

Without Interface (BAD):
    type AssetService struct {
        storage *MemoryStorage  // Coupled to specific implementation
    }
    // Can't swap to database without changing service!

With Interface (GOOD):
    type AssetService struct {
        storage Storage  // Any implementation works
    }
    // Easy to swap implementations!
*/
