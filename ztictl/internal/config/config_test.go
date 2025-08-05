package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "absolute path",
			input:    "/tmp/test",
			expected: "/tmp/test",
		},
		{
			name:     "relative path",
			input:    "test/path",
			expected: "test/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if tt.name == "empty path" || tt.name == "absolute path" || tt.name == "relative path" {
				if result != tt.expected {
					t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	// Test that setDefaults doesn't panic
	setDefaults()

	// Basic sanity checks - these shouldn't panic in CI
	if os.TempDir() == "" {
		t.Error("TempDir should not be empty")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		shouldErr bool
	}{
		{
			name: "empty SSO config should be valid (first run)",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "",
				},
			},
			shouldErr: false,
		},
		{
			name: "partial SSO config should fail",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "https://example.com",
					Region:   "", // Missing region
				},
			},
			shouldErr: true,
		},
		{
			name: "complete SSO config should pass",
			config: &Config{
				SSO: SSOConfig{
					StartURL:       "https://example.com",
					Region:         "us-east-1",
					DefaultProfile: "test-profile",
				},
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)
			if tt.shouldErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCreateSampleConfig(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	// Test that CreateSampleConfig doesn't panic and creates a file
	err := CreateSampleConfig(configPath)
	if err != nil {
		t.Fatalf("CreateSampleConfig failed: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Check that file has some content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if len(content) == 0 {
		t.Error("config file is empty")
	}

	// Basic sanity check for YAML structure
	if !containsAll(string(content), []string{"sso:", "logging:", "system:"}) {
		t.Error("config file missing expected sections")
	}
}

func containsAll(str string, substrings []string) bool {
	for _, substr := range substrings {
		if !contains(str, substr) {
			return false
		}
	}
	return true
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			(len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					indexOfSubstring(str, substr) >= 0)))
}

func indexOfSubstring(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
