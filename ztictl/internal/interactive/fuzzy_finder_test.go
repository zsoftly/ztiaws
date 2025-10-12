package interactive

import (
	"os"
	"testing"
)

func TestGetDisplayItemCount(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int
	}{
		{
			name:     "Default when env var not set",
			envValue: "",
			expected: 10,
		},
		{
			name:     "Valid value within range",
			envValue: "15",
			expected: 15,
		},
		{
			name:     "Minimum valid value",
			envValue: "1",
			expected: 1,
		},
		{
			name:     "Maximum allowed value",
			envValue: "20",
			expected: 20,
		},
		{
			name:     "Value too large (limited to 20)",
			envValue: "50",
			expected: 20,
		},
		{
			name:     "Invalid negative value (use default)",
			envValue: "-5",
			expected: 10,
		},
		{
			name:     "Invalid zero value (use default)",
			envValue: "0",
			expected: 10,
		},
		{
			name:     "Invalid non-numeric value (use default)",
			envValue: "invalid",
			expected: 10,
		},
		{
			name:     "Invalid empty spaces (use default)",
			envValue: "   ",
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envValue != "" {
				os.Setenv("ZTICTL_SELECTOR_HEIGHT", tt.envValue)
				defer os.Unsetenv("ZTICTL_SELECTOR_HEIGHT")
			} else {
				os.Unsetenv("ZTICTL_SELECTOR_HEIGHT")
			}

			// Test
			result := getDisplayItemCount()
			if result != tt.expected {
				t.Errorf("getDisplayItemCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetDisplayItemCountBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int
		wantWarn bool
	}{
		{
			name:     "Just below maximum (no warning)",
			envValue: "19",
			expected: 19,
			wantWarn: false,
		},
		{
			name:     "At maximum (no warning)",
			envValue: "20",
			expected: 20,
			wantWarn: false,
		},
		{
			name:     "Just above maximum (warning, capped)",
			envValue: "21",
			expected: 20,
			wantWarn: true,
		},
		{
			name:     "Way above maximum (warning, capped)",
			envValue: "100",
			expected: 20,
			wantWarn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ZTICTL_SELECTOR_HEIGHT", tt.envValue)
			defer os.Unsetenv("ZTICTL_SELECTOR_HEIGHT")

			result := getDisplayItemCount()
			if result != tt.expected {
				t.Errorf("getDisplayItemCount() = %d, want %d", result, tt.expected)
			}
		})
	}
}
