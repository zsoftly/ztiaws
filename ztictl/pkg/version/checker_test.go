package version

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		latest   string
		expected bool
	}{
		{"same version", "2.8.0", "2.8.0", false},
		{"current older", "2.7.0", "2.8.0", true},
		{"current newer", "2.9.0", "2.8.0", false},
		{"current with hash older", "2.7.0-abcd123", "2.8.0", true},
		{"current with hash same", "2.8.0-abcd123", "2.8.0", false},
		{"minor version diff", "2.8.0", "2.8.1", true},
		{"major version diff", "1.9.0", "2.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.current, tt.latest)
			if result != tt.expected {
				t.Errorf("compareVersions(%q, %q) = %v, want %v", tt.current, tt.latest, result, tt.expected)
			}
		})
	}
}
