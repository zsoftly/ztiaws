package aws

import (
	"strings"
	"testing"
)

func TestGetRegion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid region shortcode",
			input:    "use1",
			expected: "us-east-1",
			wantErr:  false,
		},
		{
			name:     "another valid region",
			input:    "cac1",
			expected: "ca-central-1",
			wantErr:  false,
		},
		{
			name:     "invalid region",
			input:    "invalid",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region, err := GetRegion(tt.input)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRegion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if region != tt.expected {
				t.Errorf("GetRegion() = %v, want %v", region, tt.expected)
			}
		})
	}
}

func TestGetRegionDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "us-east-1 description",
			input:    "use1",
			contains: "Virginia",
		},
		{
			name:     "canada-central description",
			input:    "cac1",
			contains: "Canada",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := GetRegionDescription(tt.input)
			if !strings.Contains(desc, tt.contains) {
				t.Errorf("GetRegionDescription() = %v, expected to contain %v", desc, tt.contains)
			}
		})
	}
}