package model

import (
	"testing"
)

// TestIsValidScanType tests scan type validation
func TestIsValidScanType(t *testing.T) {
	tests := []struct {
		name     string
		scanType ScanType
		want     bool
	}{
		{
			name:     "valid - all",
			scanType: ScanTypeAll,
			want:     true,
		},
		{
			name:     "valid - dns",
			scanType: ScanTypeDNS,
			want:     true,
		},
		{
			name:     "valid - whois",
			scanType: ScanTypeWHOIS,
			want:     true,
		},
		{
			name:     "valid - subdomain",
			scanType: ScanTypeSubdomain,
			want:     true,
		},
		{
			name:     "valid - port",
			scanType: ScanTypePort,
			want:     true,
		},
		{
			name:     "valid - ssl",
			scanType: ScanTypeSSL,
			want:     true,
		},
		{
			name:     "valid - asn",
			scanType: ScanTypeASN,
			want:     true,
		},
		{
			name:     "invalid - empty",
			scanType: ScanType(""),
			want:     false,
		},
		{
			name:     "invalid - unknown type",
			scanType: ScanType("unknown"),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidScanType(tt.scanType)
			if got != tt.want {
				t.Errorf("IsValidScanType(%q) = %v, want %v", tt.scanType, got, tt.want)
			}
		})
	}
}

// TestScanTypeCategory tests scan category classification
func TestScanTypeCategory(t *testing.T) {
	tests := []struct {
		name     string
		scanType ScanType
		want     ScanCategory
	}{
		{
			name:     "dns is passive",
			scanType: ScanTypeDNS,
			want:     ScanCategoryPassive,
		},
		{
			name:     "whois is passive",
			scanType: ScanTypeWHOIS,
			want:     ScanCategoryPassive,
		},
		{
			name:     "subdomain is passive",
			scanType: ScanTypeSubdomain,
			want:     ScanCategoryPassive,
		},
		{
			name:     "all is passive",
			scanType: ScanTypeAll,
			want:     ScanCategoryPassive,
		},
		{
			name:     "asn is passive",
			scanType: ScanTypeASN,
			want:     ScanCategoryPassive,
		},
		{
			name:     "port is active",
			scanType: ScanTypePort,
			want:     ScanCategoryActive,
		},
		{
			name:     "ssl is active",
			scanType: ScanTypeSSL,
			want:     ScanCategoryActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.scanType.Category()
			if got != tt.want {
				t.Errorf("%s.Category() = %v, want %v", tt.scanType, got, tt.want)
			}
		})
	}
}

// TestScanTypeRequiresPermission tests permission checking
func TestScanTypeRequiresPermission(t *testing.T) {
	tests := []struct {
		name     string
		scanType ScanType
		want     bool
	}{
		{
			name:     "dns - no permission needed",
			scanType: ScanTypeDNS,
			want:     false,
		},
		{
			name:     "whois - no permission needed",
			scanType: ScanTypeWHOIS,
			want:     false,
		},
		{
			name:     "subdomain - no permission needed",
			scanType: ScanTypeSubdomain,
			want:     false,
		},
		{
			name:     "port - permission required",
			scanType: ScanTypePort,
			want:     true,
		},
		{
			name:     "ssl - permission required",
			scanType: ScanTypeSSL,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.scanType.RequiresPermission()
			if got != tt.want {
				t.Errorf("%s.RequiresPermission() = %v, want %v", tt.scanType, got, tt.want)
			}
		})
	}
}

// TestScanTypeIsPassive tests passive scan detection
func TestScanTypeIsPassive(t *testing.T) {
	passiveScans := []ScanType{
		ScanTypeDNS,
		ScanTypeWHOIS,
		ScanTypeSubdomain,
		ScanTypeAll,
		ScanTypeASN,
		ScanTypeCertTrans,
	}

	for _, scanType := range passiveScans {
		t.Run(string(scanType), func(t *testing.T) {
			if !scanType.IsPassive() {
				t.Errorf("%s.IsPassive() = false, want true", scanType)
			}
			if scanType.IsActive() {
				t.Errorf("%s.IsActive() = true, want false", scanType)
			}
		})
	}
}

// TestScanTypeIsActive tests active scan detection
func TestScanTypeIsActive(t *testing.T) {
	activeScans := []ScanType{
		ScanTypePort,
		ScanTypeSSL,
	}

	for _, scanType := range activeScans {
		t.Run(string(scanType), func(t *testing.T) {
			if !scanType.IsActive() {
				t.Errorf("%s.IsActive() = false, want true", scanType)
			}
			if scanType.IsPassive() {
				t.Errorf("%s.IsPassive() = true, want false", scanType)
			}
		})
	}
}

// TestScanTypeDescription tests description generation
func TestScanTypeDescription(t *testing.T) {
	tests := []struct {
		name     string
		scanType ScanType
		contains string
	}{
		{
			name:     "dns description mentions DNS",
			scanType: ScanTypeDNS,
			contains: "DNS",
		},
		{
			name:     "whois description mentions WHOIS",
			scanType: ScanTypeWHOIS,
			contains: "WHOIS",
		},
		{
			name:     "port description mentions permission",
			scanType: ScanTypePort,
			contains: "permission",
		},
		{
			name:     "all description mentions passive",
			scanType: ScanTypeAll,
			contains: "passive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.scanType.Description()
			if desc == "" {
				t.Errorf("%s.Description() returned empty string", tt.scanType)
			}
		})
	}
}

// TestIsValidScanStatus tests scan status validation
func TestIsValidScanStatus(t *testing.T) {
	tests := []struct {
		name   string
		status ScanStatus
		want   bool
	}{
		{
			name:   "valid - pending",
			status: ScanStatusPending,
			want:   true,
		},
		{
			name:   "valid - running",
			status: ScanStatusRunning,
			want:   true,
		},
		{
			name:   "valid - completed",
			status: ScanStatusCompleted,
			want:   true,
		},
		{
			name:   "valid - failed",
			status: ScanStatusFailed,
			want:   true,
		},
		{
			name:   "valid - partial",
			status: ScanStatusPartial,
			want:   true,
		},
		{
			name:   "invalid - empty",
			status: ScanStatus(""),
			want:   false,
		},
		{
			name:   "invalid - unknown",
			status: ScanStatus("processing"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidScanStatus(tt.status)
			if got != tt.want {
				t.Errorf("IsValidScanStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// TestPassiveVsActiveClassification ensures scan classification
func TestPassiveVsActiveClassification(t *testing.T) {
	// All scans should be classified as either passive or active
	allScanTypes := []ScanType{
		ScanTypeAll,
		ScanTypeDNS,
		ScanTypeWHOIS,
		ScanTypeSubdomain,
		ScanTypeCertTrans,
		ScanTypeASN,
		ScanTypePort,
		ScanTypeSSL,
	}

	for _, scanType := range allScanTypes {
		t.Run(string(scanType), func(t *testing.T) {
			// Must be either passive or active, not both, not neither
			isPassive := scanType.IsPassive()
			isActive := scanType.IsActive()

			if isPassive == isActive {
				t.Errorf("%s should be either passive or active, not both or neither", scanType)
			}

			// Permission requirement should align with active status
			if scanType.RequiresPermission() != isActive {
				t.Errorf("%s permission requirement doesn't match active status", scanType)
			}
		})
	}
}

// BenchmarkScanTypeCategory benchmarks category lookup
func BenchmarkScanTypeCategory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ScanTypeDNS.Category()
	}
}

// BenchmarkIsValidScanType benchmarks scan type validation
func BenchmarkIsValidScanType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsValidScanType(ScanTypeDNS)
	}
}
