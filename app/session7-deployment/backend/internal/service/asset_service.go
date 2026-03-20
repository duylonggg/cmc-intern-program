package service

import (
	"mini-asm/internal/model"
	"mini-asm/internal/storage"
	"mini-asm/internal/validator"
	"time"

	"github.com/google/uuid"
)

// AssetService handles business logic for asset operations
// It sits between handlers (HTTP layer) and storage (data layer)
type AssetService struct {
	storage   storage.Storage           // Dependency injection - any Storage implementation
	validator *validator.AssetValidator // Validator for input validation
}

// NewAssetService creates a new asset service
// Takes Storage interface - can be memory, database, or mock
func NewAssetService(storage storage.Storage) *AssetService {
	return &AssetService{
		storage:   storage,
		validator: validator.NewAssetValidator(),
	}
}

// CreateAsset creates a new asset with validation
// Returns the created asset or an error
func (s *AssetService) CreateAsset(name, assetType string) (*model.Asset, error) {
	// Validation using validator package
	if err := s.validator.ValidateCreate(name, assetType); err != nil {
		return nil, err
	}

	// Business logic - create asset with defaults
	asset := &model.Asset{
		ID:        uuid.New().String(), // Auto-generate UUID
		Name:      name,
		Type:      assetType,
		Status:    model.StatusActive, // Default status
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Delegate to storage layer
	if err := s.storage.Create(asset); err != nil {
		return nil, err
	}

	return asset, nil
}

// ListAssets retrieves assets with filtering, sorting, and pagination
func (s *AssetService) ListAssets(params storage.QueryParams) (*storage.PaginatedResult, error) {
	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100 // Max page size
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	// Validate parameters
	if err := s.validator.ValidatePaginationParams(params.Page, params.PageSize); err != nil {
		return nil, err
	}

	if err := s.validator.ValidateSortParams(params.SortBy, params.SortOrder); err != nil {
		return nil, err
	}

	// Validate filters if provided
	if params.Type != "" {
		if err := s.validator.ValidateType(params.Type); err != nil {
			return nil, err
		}
	}

	if params.Status != "" {
		if err := s.validator.ValidateStatus(params.Status); err != nil {
			return nil, err
		}
	}

	if params.Search != "" {
		if err := s.validator.ValidateSearchQuery(params.Search); err != nil {
			return nil, err
		}
	}

	// Delegate to storage
	return s.storage.GetAll(params)
}

// GetAssetByID retrieves a single asset by ID
func (s *AssetService) GetAssetByID(id string) (*model.Asset, error) {
	if id == "" {
		return nil, model.ErrInvalidInput
	}

	return s.storage.GetByID(id)
}

// UpdateAsset updates an existing asset
// Only updates provided fields (partial update)
func (s *AssetService) UpdateAsset(id string, name, assetType, status string) (*model.Asset, error) {
	// Validate ID
	if id == "" {
		return nil, model.ErrInvalidInput
	}

	// Get existing asset
	existing, err := s.storage.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Store original type for validation
	originalType := existing.Type

	// Apply updates (only if provided)
	if name != "" {
		existing.Name = name
	}

	if assetType != "" {
		existing.Type = assetType
	}

	if status != "" {
		existing.Status = status
	}

	// Validate updates
	if err := s.validator.ValidateUpdate(existing.Name, existing.Type, existing.Status); err != nil {
		return nil, err
	}

	// If type changed, validate new name against new type
	if assetType != "" && assetType != originalType {
		if err := s.validator.ValidateCreate(existing.Name, existing.Type); err != nil {
			return nil, err
		}
	}

	// Update timestamp
	existing.UpdatedAt = time.Now()

	// Save to storage
	if err := s.storage.Update(id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteAsset removes an asset
func (s *AssetService) DeleteAsset(id string) error {
	if id == "" {
		return model.ErrInvalidInput
	}

	return s.storage.Delete(id)
}

/*
🎓 TEACHING NOTES:

=== SESSION 4 ENHANCEMENTS ===

1. Validator Integration:
   - Service now has validator field
   - All validations delegated to validator package
   - Single responsibility: service = business logic

   Before (Session 3):
   if name == "" {
       return nil, model.ErrEmptyName
   }

   After (Session 4):
   if err := s.validator.ValidateCreate(name, assetType); err != nil {
       return nil, err
   }

2. ListAssets with Query Parameters:
   - Replaced separate GetAll(), Filter(), Search()
   - Single method with QueryParams struct
   - Sets sensible defaults
   - Validates all parameters

3. Default Values:
   - Page 1 if not specified
   - PageSize 20 if not specified
   - Max PageSize 100 (prevent abuse)
   - SortBy "created_at" if not specified
   - SortOrder "desc" if not specified

4. Validation Strategy:
   - Validate pagination params
   - Validate sort params
   - Validate filters if provided
   - Catch errors early before hitting database

5. Enhanced UpdateAsset:
   - Validates before update
   - Special handling for type changes
   - Re-validates name if type changes
   - Example: changing from "domain" to "ip"
     Must check if name is valid IP format

COMPARISON:

Session 3:
  service.GetAllAssets()                      // No parameters
  service.FilterAssets("domain", "active")    // Limited filters
  service.SearchAssets("example")             // Separate method

Session 4:
  service.ListAssets(QueryParams{})           // All with pagination
  service.ListAssets(QueryParams{             // Combined filters
    Type: "domain",
    Status: "active",
    Search: "example",
    Page: 2,
    PageSize: 20,
    SortBy: "name",
    SortOrder: "asc",
  })

KEY IMPROVEMENTS:

1. Type Safety:
   - QueryParams struct vs many parameters
   - Compiler helps catch mistakes

2. Extensibility:
   - Add new filter = add field to struct
   - No need to change method signature

3. Default Handling:
   - Service sets defaults, not handler
   - Consistent across all callers

4. Validation:
   - Comprehensive validation before storage
   - Clear error messages
   - Prevents invalid state

VALIDATION FLOW:

1. Handler receives request
2. Parses query parameters
3. Calls service.ListAssets(params)
4. Service validates params:
   - Pagination valid? (page >= 1, size <= 100)
   - Sort fields valid? (whitelisted)
   - Filters valid? (type in [domain, ip, service])
   - Search valid? (no SQL injection patterns)
5. If valid → storage.GetAll(params)
6. If invalid → return error to handler
7. Handler maps error to HTTP status code

ERROR MAPPING (in Handler layer):

Validation error → 400 Bad Request
Not found → 404 Not Found
Server error → 500 Internal Server Error

BUSINESS LOGIC EXAMPLES:

1. Auto-generate UUID:
   - Business rule: all assets have unique IDs
   - Service responsibility, not handler

2. Default status:
   - Business rule: new assets are active by default
   - Can override later with UpdateAsset

3. Timestamps:
   - Business rule: track creation and modification
   - Auto-set, user can't override

4. Partial updates:
   - Business rule: update only provided fields
   - Fetch existing, merge, save

TESTING CONSIDERATIONS:

1. Unit test with mock storage:
   mockStorage := &MockStorage{}
   service := NewAssetService(mockStorage)

2. Test validation:
   - Invalid page number → error
   - Invalid sort field → error
   - Too large page size → clamped to 100

3. Test defaults:
   - No page → defaults to 1
   - No sort → defaults to created_at desc

4. Test business logic:
   - UUID generated
   - Timestamps set
   - Status defaults to active

DEMO POINTS:

1. Show validator catching invalid domain
2. Show service setting defaults
3. Show ListAssets with various parameter combinations
4. Show type-specific validation on update
5. Compare with Session 3 code (much cleaner!)
*/

/*
🎓 NOTES:

1. Service Layer Responsibilities:
   ✅ Business logic
   ✅ Validation
   ✅ Orchestration (coordinate multiple operations)
   ✅ Default values
   ❌ HTTP concerns (status codes, JSON)
   ❌ Database details (SQL, queries)

2. Dependency Injection:
   func NewAssetService(storage storage.Storage) *AssetService

   Q: Tại sao không dùng global variable?
   A: Testability! Có thể inject mock storage trong tests

   Example:
   // Production
   service := NewAssetService(memory.NewMemoryStorage())

   // Testing
   service := NewAssetService(&MockStorage{})

3. Validation Strategy:
   - Validate BEFORE business logic
   - Return specific errors (ErrEmptyName, ErrInvalidType)
   - Handler layer maps to HTTP status codes

4. Business Logic Examples:
   - Auto-generate UUID
   - Set default status = active
   - Auto-set timestamps
   - Partial updates (only update provided fields)

5. Error Propagation:
   if err := s.storage.Create(asset); err != nil {
       return nil, err  // Let caller handle
   }

   Q: Tại sao không handle error ở đây?
   A: Service không biết context (HTTP? CLI? gRPC?)
      Handler layer sẽ decide status code

6. Comparison với "Fat Controller":

   ❌ BAD (All in handler):
   func CreateAssetHandler(w, r) {
       // Parse JSON
       // Validate
       // Generate UUID
       // Set defaults
       // Save to DB
       // Return response
   }
   → Hard to test, hard to reuse

   ✅ GOOD (Layered):
   Handler: Parse JSON, call service, return HTTP response
   Service: Validate, business logic
   Storage: Data persistence
   → Easy to test each layer, reusable logic

7. UUID Generation:
   uuid.New().String() → "550e8400-e29b-41d4-a716-446655440000"
   - Globally unique
   - No need for database auto-increment
   - Can generate offline
   - URL-safe

8. Timestamps:
   CreatedAt: time.Now()  // Set once
   UpdatedAt: time.Now()  // Update on every change

   Q: Tại sao không dùng int64 (Unix timestamp)?
   A: time.Time có timezone, human-readable, JSON support

📝 TESTING SERVICE LAYER:

func TestCreateAsset(t *testing.T) {
    // Arrange
    mockStorage := &MockStorage{}
    service := NewAssetService(mockStorage)

    // Act
    asset, err := service.CreateAsset("example.com", "domain")

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, asset.ID)
    assert.Equal(t, "active", asset.Status) // Default
}

❓ QUESTIONS TO ASK:

1. Tại sao CreateAsset return (*Asset, error) thay vì chỉ error?
   → Need to return created asset with ID to client

2. Tại sao UpdateAsset là partial update?
   → Flexible - client chỉ gửi fields muốn update

3. Service layer có nên biết về HTTP không?
   → KHÔNG! Separation of concerns

4. Làm sao test service layer mà không cần database?
   → Mock storage! (Buổi 5)
*/
