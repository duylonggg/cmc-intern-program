package model

import (
	"testing"
)

// TestIsValidType tests the asset type validation
func TestIsValidType(t *testing.T) {
	tests := []struct {
		name      string
		assetType string
		want      bool
	}{
		{
			name:      "valid domain type",
			assetType: TypeDomain,
			want:      true,
		},
		{
			name:      "valid ip type",
			assetType: TypeIP,
			want:      true,
		},
		{
			name:      "valid service type",
			assetType: TypeService,
			want:      true,
		},
		{
			name:      "invalid type - empty",
			assetType: "",
			want:      false,
		},
		{
			name:      "invalid type - random string",
			assetType: "invalid",
			want:      false,
		},
		{
			name:      "invalid type - case sensitive",
			assetType: "DOMAIN",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidType(tt.assetType)
			if got != tt.want {
				t.Errorf("IsValidType(%q) = %v, want %v", tt.assetType, got, tt.want)
			}
		})
	}
}

// TestIsValidStatus tests the asset status validation
func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "valid active status",
			status: StatusActive,
			want:   true,
		},
		{
			name:   "valid inactive status",
			status: StatusInactive,
			want:   true,
		},
		{
			name:   "invalid status - empty",
			status: "",
			want:   false,
		},
		{
			name:   "invalid status - random string",
			status: "pending",
			want:   false,
		},
		{
			name:   "invalid status - case sensitive",
			status: "ACTIVE",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidStatus(tt.status)
			if got != tt.want {
				t.Errorf("IsValidStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// BenchmarkIsValidType benchmarks the type validation function
func BenchmarkIsValidType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsValidType(TypeDomain)
	}
}

// BenchmarkIsValidStatus benchmarks the status validation function
func BenchmarkIsValidStatus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsValidStatus(StatusActive)
	}
}
