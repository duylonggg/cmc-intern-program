package validator

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"mini-asm/internal/model"
)

// AssetValidator provides validation for asset-related operations
type AssetValidator struct{}

// NewAssetValidator creates a new validator instance
func NewAssetValidator() *AssetValidator {
	return &AssetValidator{}
}

// ValidateCreate validates asset creation request
func (v *AssetValidator) ValidateCreate(name, assetType string) error {
	// Validate name
	if err := v.ValidateName(name); err != nil {
		return err
	}

	// Validate type
	if err := v.ValidateType(assetType); err != nil {
		return err
	}

	// Type-specific validation
	switch assetType {
	case model.TypeDomain:
		if err := v.ValidateDomain(name); err != nil {
			return fmt.Errorf("invalid domain: %w", err)
		}
	case model.TypeIP:
		if err := v.ValidateIP(name); err != nil {
			return fmt.Errorf("invalid IP address: %w", err)
		}
	case model.TypeService:
		if err := v.ValidateService(name); err != nil {
			return fmt.Errorf("invalid service: %w", err)
		}
	}

	return nil
}

// ValidateUpdate validates asset update request
func (v *AssetValidator) ValidateUpdate(name, assetType, status string) error {
	// Name validation (if provided)
	if name != "" {
		if err := v.ValidateName(name); err != nil {
			return err
		}

		// Type-specific validation
		switch assetType {
		case model.TypeDomain:
			if err := v.ValidateDomain(name); err != nil {
				return fmt.Errorf("invalid domain: %w", err)
			}
		case model.TypeIP:
			if err := v.ValidateIP(name); err != nil {
				return fmt.Errorf("invalid IP address: %w", err)
			}
		case model.TypeService:
			if err := v.ValidateService(name); err != nil {
				return fmt.Errorf("invalid service: %w", err)
			}
		}
	}

	// Type validation (if provided)
	if assetType != "" {
		if err := v.ValidateType(assetType); err != nil {
			return err
		}
	}

	// Status validation (if provided)
	if status != "" {
		if err := v.ValidateStatus(status); err != nil {
			return err
		}
	}

	return nil
}

// ValidateName checks if name is valid
func (v *AssetValidator) ValidateName(name string) error {
	if name == "" {
		return errors.New("name is required")
	}

	if len(name) > 255 {
		return errors.New("name too long (max 255 characters)")
	}

	// Check for null bytes (security)
	if strings.Contains(name, "\x00") {
		return errors.New("name contains invalid characters")
	}

	return nil
}

// ValidateType checks if asset type is valid
func (v *AssetValidator) ValidateType(assetType string) error {
	if !model.IsValidType(assetType) {
		return fmt.Errorf("invalid asset type: must be %s, %s, or %s",
			model.TypeDomain, model.TypeIP, model.TypeService)
	}
	return nil
}

// ValidateStatus checks if status is valid
func (v *AssetValidator) ValidateStatus(status string) error {
	if !model.IsValidStatus(status) {
		return fmt.Errorf("invalid status: must be %s or %s",
			model.StatusActive, model.StatusInactive)
	}
	return nil
}

// ValidateDomain checks if domain name is valid
func (v *AssetValidator) ValidateDomain(domain string) error {
	// Basic length check
	if len(domain) < 1 || len(domain) > 253 {
		return errors.New("domain length must be between 1 and 253 characters")
	}

	// Check if it's actually an IP address (should use TypeIP instead)
	if net.ParseIP(domain) != nil {
		return errors.New("IP address should use type 'ip' instead of 'domain'")
	}

	// Check for valid domain format
	// RFC 1035: labels must be 1-63 characters, only alphanumeric and hyphens
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)

	if !domainRegex.MatchString(domain) {
		return errors.New("invalid domain format (e.g., example.com)")
	}

	// Disallow leading/trailing dots or hyphens
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return errors.New("domain cannot start or end with a dot")
	}

	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return errors.New("domain cannot start or end with a hyphen")
	}

	return nil
}

// ValidateIP checks if IP address is valid (IPv4 or IPv6)
func (v *AssetValidator) ValidateIP(ip string) error {
	// Try parsing as IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return errors.New("invalid IP address format (e.g., 192.168.1.1 or 2001:db8::1)")
	}

	return nil
}

// ValidateService checks if service name is valid
func (v *AssetValidator) ValidateService(service string) error {
	// Service format: protocol://host:port or just service name

	// Minimum length
	if len(service) < 1 {
		return errors.New("service name is required")
	}

	// Check for common patterns
	// Allow: http://example.com, https://example.com:443, ssh, ftp, etc.
	serviceRegex := regexp.MustCompile(`^([a-zA-Z0-9\-]+)(://[a-zA-Z0-9\-\.]+)?(:[\d]+)?(/.*)?$`)

	if !serviceRegex.MatchString(service) {
		return errors.New("invalid service format (e.g., http://example.com or ssh)")
	}

	return nil
}

// ValidatePaginationParams validates pagination parameters
func (v *AssetValidator) ValidatePaginationParams(page, pageSize int) error {
	if page < 1 {
		return errors.New("page must be >= 1")
	}

	if pageSize < 1 {
		return errors.New("page_size must be >= 1")
	}

	if pageSize > 100 {
		return errors.New("page_size too large (max 100)")
	}

	return nil
}

// ValidateSortParams validates sort parameters
func (v *AssetValidator) ValidateSortParams(sortBy, sortOrder string) error {
	// Whitelist of allowed sort fields (prevent SQL injection)
	validSortFields := map[string]bool{
		"name":       true,
		"type":       true,
		"status":     true,
		"created_at": true,
		"updated_at": true,
	}

	if sortBy != "" && !validSortFields[sortBy] {
		return fmt.Errorf("invalid sort field: %s (allowed: name, type, status, created_at, updated_at)", sortBy)
	}

	if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
		return errors.New("sort order must be 'asc' or 'desc'")
	}

	return nil
}

// ValidateSearchQuery validates search query
func (v *AssetValidator) ValidateSearchQuery(query string) error {
	if len(query) > 255 {
		return errors.New("search query too long (max 255 characters)")
	}

	// Check for SQL injection patterns (defense in depth)
	dangerousPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
	}

	lowerQuery := strings.ToLower(query)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerQuery, pattern) {
			return fmt.Errorf("search query contains invalid characters: %s", pattern)
		}
	}

	return nil
}

/*
🎓 NOTES:

1. Input Validation Strategy:
   - Validate early (at boundary)
   - Fail fast with clear error messages
   - Multi-layer: type checking + format + business rules

2. Security Considerations:
   - Null byte injection prevention
   - SQL injection pattern detection
   - Whitelist approach for sort fields
   - Length limits on all inputs

3. Type-Specific Validation:
   - Domain: RFC 1035 compliance
   - IP: IPv4 and IPv6 support
   - Service: Flexible format for various protocols

4. Regex Patterns:
   - Domain: ^([a-zA-Z0-9]...)$ → strict format
   - Keep regex simple and testable
   - Comment complex patterns

5. Error Messages:
   - User-friendly: "domain length must be between 1 and 253"
   - Include examples: "(e.g., example.com)"
   - Don't leak internals: "invalid" not "SQL error"

6. Validator as Service:
   - Reusable across layers
   - Easy to unit test
   - Single responsibility

7. Pagination Validation:
   - Range checks (page >= 1)
   - Reasonable limits (page_size <= 100)
   - Prevent abuse

8. Sort Validation:
   - Whitelist fields (CRITICAL for security!)
   - If not whitelisted → reject or use default
   - Never concatenate user input into SQL ORDER BY

COMMON ATTACKS PREVENTED:

1. SQL Injection:
   - Whitelist sort fields
   - Check for SQL keywords in search
   - Use parameterized queries (in storage layer)

2. NoSQL Injection:
   - Validate input types
   - Reject special characters

3. DoS:
   - Limit page_size to reasonable number
   - Limit search query length

4. Path Traversal:
   - Not applicable here but check file paths if added

TESTING STRATEGY:

Write tests for:
- ✅ Valid inputs (happy path)
- ❌ Empty/null inputs
- ❌ Too long inputs
- ❌ Invalid formats
- ❌ SQL injection attempts
- ❌ Edge cases (max length, special chars)

Example test cases:
- Domain: "example.com" ✓, "invalid..com" ✗, "toolongdomainname..." ✗
- IP: "192.168.1.1" ✓, "999.999.999.999" ✗, "not-an-ip" ✗
- Service: "http://example.com" ✓, "ssh" ✓, "invalid service!!!" ✗

BEST PRACTICES:

1. Validate at API boundary (handler/service)
2. Use validator consistently across all operations
3. Return meaningful errors
4. Log validation failures (may indicate attack)
5. Don't trust client-side validation
6. Defense in depth: validate + parameterized queries
*/
