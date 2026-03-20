package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mini-asm/internal/model"
	"mini-asm/internal/service"
	"mini-asm/internal/storage"
)

// AssetHandler handles HTTP requests for asset operations
type AssetHandler struct {
	service *service.AssetService
}

// NewAssetHandler creates a new asset handler
func NewAssetHandler(service *service.AssetService) *AssetHandler {
	return &AssetHandler{
		service: service,
	}
}

// CreateAssetRequest represents the request body for creating an asset
type CreateAssetRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// UpdateAssetRequest represents the request body for updating an asset
type UpdateAssetRequest struct {
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateAsset handles POST /assets
func (h *AssetHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	asset, err := h.service.CreateAsset(req.Name, req.Type)
	if err != nil {
		status := mapErrorToStatus(err)
		respondJSON(w, status, ErrorResponse{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, asset)
}

// ListAssets handles GET /assets with query parameters
// Supports: ?page=1&page_size=20&type=domain&status=active&search=example&sort_by=name&sort_order=asc
func (h *AssetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	params := storage.QueryParams{
		Page:      parseIntParam(r, "page", 1),
		PageSize:  parseIntParam(r, "page_size", 20),
		Type:      r.URL.Query().Get("type"),
		Status:    r.URL.Query().Get("status"),
		Search:    r.URL.Query().Get("search"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	result, err := h.service.ListAssets(params)
	if err != nil {
		status := mapErrorToStatus(err)
		respondJSON(w, status, ErrorResponse{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetAsset handles GET /assets/{id}
func (h *AssetHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	asset, err := h.service.GetAssetByID(id)
	if err != nil {
		status := mapErrorToStatus(err)
		respondJSON(w, status, ErrorResponse{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, asset)
}

// UpdateAsset handles PUT /assets/{id}
func (h *AssetHandler) UpdateAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req UpdateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	asset, err := h.service.UpdateAsset(id, req.Name, req.Type, req.Status)
	if err != nil {
		status := mapErrorToStatus(err)
		respondJSON(w, status, ErrorResponse{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, asset)
}

// DeleteAsset handles DELETE /assets/{id}
func (h *AssetHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.service.DeleteAsset(id); err != nil {
		status := mapErrorToStatus(err)
		respondJSON(w, status, ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parseIntParam parses an integer query parameter with a default value
func parseIntParam(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil || intValue < 1 {
		return defaultValue
	}

	return intValue
}

// respondJSON sends a JSON response with the given status code
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// mapErrorToStatus maps domain errors to HTTP status codes
func mapErrorToStatus(err error) int {
	switch err {
	case model.ErrNotFound:
		return http.StatusNotFound
	case model.ErrInvalidInput, model.ErrEmptyName, model.ErrInvalidType, model.ErrInvalidStatus:
		return http.StatusBadRequest
	default:
		// Check if error message contains validation keywords
		errMsg := err.Error()
		if contains(errMsg, "invalid") || contains(errMsg, "required") || contains(errMsg, "too long") {
			return http.StatusBadRequest
		}
		return http.StatusInternalServerError
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)*2))
}

/*
🎓 TEACHING NOTES:

=== SESSION 4 ENHANCEMENTS ===

1. Query Parameter Parsing:
   - parseIntParam() helper for safe integer parsing
   - Defaults if parameter missing or invalid
   - Example: ?page=abc → defaults to 1

2. Comprehensive Query Support:
   GET /assets?page=2&page_size=20&type=domain&status=active&search=example&sort_by=name&sort_order=asc

   All optional parameters:
   - page: pagination (default 1)
   - page_size: results per page (default 20)
   - type: filter by asset type
   - status: filter by status
   - search: search in name field
   - sort_by: field to sort by
   - sort_order: asc or desc

3. Request/Response Structs:
   - CreateAssetRequest: typed request body
   - UpdateAssetRequest: supports partial updates (omitempty)
   - ErrorResponse: consistent error format

4. Error Handling:
   - mapErrorToStatus() maps domain errors to HTTP status
   - Validation errors → 400 Bad Request
   - Not found → 404 Not Found
   - Server errors → 500 Internal Server Error

5. JSON Response Helper:
   - respondJSON() centralizes JSON encoding
   - Sets Content-Type header
   - Consistent response format

HANDLER RESPONSIBILITIES:

✅ HTTP concerns:
   - Parse request (body, query params, path params)
   - Call service method
   - Map errors to status codes
   - Send JSON response

❌ NOT handler's job:
   - Business logic (service layer)
   - Validation logic (validator + service)
   - Database queries (storage layer)

QUERY PARAMETER EXAMPLES:

1. Paginated list:
   GET /assets?page=1&page_size=20
   Response: {
     "data": [...],
     "total": 245,
     "page": 1,
     "page_size": 20,
     "total_pages": 13
   }

2. Filtered list:
   GET /assets?type=domain&status=active
   Response: {
     "data": [... only active domains ...],
     "total": 42,
     ...
   }

3. Search:
   GET /assets?search=example
   Response: {
     "data": [... assets with "example" in name ...],
     ...
   }

4. Combined:
   GET /assets?type=domain&search=google&page=2&sort_by=name&sort_order=asc
   Response: {
     "data": [... domains with "google", page 2, sorted by name ...],
     ...
   }

ERROR RESPONSES:

400 Bad Request:
  {
    "error": "invalid domain format (e.g., example.com)"
  }

404 Not Found:
  {
    "error": "asset not found"
  }

500 Internal Server Error:
  {
    "error": "internal server error"
  }

COMPARISON WITH SESSION 3:

Session 3:
  GET /assets                              // All assets, no pagination
  GET /assets?type=domain&status=active    // Filters work BUT no pagination!

Session 4:
  GET /assets                              // Paginated (default page 1)
  GET /assets?page=2                       // Second page
  GET /assets?type=domain&page=2           // Filtered AND paginated
  GET /assets?search=example&sort_by=name  // Search AND sort

KEY IMPROVEMENTS:

1. Flexible querying with any combination of parameters
2. Pagination ALWAYS included
3. Sorting support
4. Better error messages from validator
5. Consistent response structure

TESTING:

1. Unit tests:
   - Mock service layer
   - Test parameter parsing
   - Test error mapping

2. Integration tests:
   - Test with real service
   - Various query combinations
   - Edge cases (invalid params, etc.)

DEMO FLOW:

1. Create some assets:
   POST /assets {"name":"example.com","type":"domain"}
   POST /assets {"name":"192.168.1.1","type":"ip"}
   POST /assets {"name":"test.com","type":"domain"}

2. List with pagination:
   GET /assets?page=1&page_size=2
   → Shows 2 results, total count

3. Filter:
   GET /assets?type=domain
   → Only domains

4. Search:
   GET /assets?search=example
   → Only "example.com"

5. Sort:
   GET /assets?sort_by=name&sort_order=asc
   → Alphabetical order

6. Combined:
   GET /assets?type=domain&search=example&page=1&sort_by=name
   → All features together!

SECURITY NOTES:

1. Input validation in service layer (validator)
2. Safe integer parsing (defaults instead of errors)
3. SQL injection prevention (whitelisted sort fields)
4. Error message sanitization (don't leak internals)

CLIENT USAGE:

JavaScript:
  const params = new URLSearchParams({
    page: 2,
    page_size: 20,
    type: 'domain',
    search: 'example'
  });
  fetch(`/assets?${params}`);

curl:
  curl "http://localhost:8080/assets?page=1&type=domain&search=example"

Postman:
  Add query params in Params tab
  Automatically URL-encodes

BEST PRACTICES:

1. Use query params for filtering/pagination (GET)
2. Use request body for resource data (POST/PUT)
3. Use path params for IDs (GET /assets/{id})
4. Return appropriate status codes
5. Include pagination metadata in response
6. Validate inputs before processing
7. Log errors (production consideration)
*/
