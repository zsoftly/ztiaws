package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// getUserHomeForTest returns the appropriate home directory for the current platform
func getUserHomeForTest() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows
	}
	return home
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "absolute path",
			input:    "/tmp/test",
			expected: "/tmp/test",
		},
		{
			name:     "relative path",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "tilde path",
			input:    "~/test",
			expected: filepath.Join(getUserHomeForTest(), "test"),
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConfigValidationError(t *testing.T) {
	err := &ConfigValidationError{
		Field:   "SSO region",
		Value:   "invalid-region",
		Message: "not a valid AWS region",
	}

	expected := "SSO region 'invalid-region' is invalid: not a valid AWS region"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestValidateLoadedConfigDetailed(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorField  string
	}{
		{
			name: "valid config",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "https://example.awsapps.com/start",
					Region:   "us-east-1",
				},
				DefaultRegion: "us-west-2",
			},
			expectError: false,
		},
		{
			name: "placeholder SSO URL",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "https://d-xxxxxxxxxx.awsapps.com/start",
					Region:   "us-east-1",
				},
				DefaultRegion: "us-west-2",
			},
			expectError: true,
			errorField:  "SSO start URL",
		},
		{
			name: "invalid SSO region",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "https://example.awsapps.com/start",
					Region:   "invalid-region",
				},
				DefaultRegion: "us-west-2",
			},
			expectError: true,
			errorField:  "SSO region",
		},
		{
			name: "invalid default region",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "https://example.awsapps.com/start",
					Region:   "us-east-1",
				},
				DefaultRegion: "invalid-region",
			},
			expectError: true,
			errorField:  "Default region",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoadedConfigDetailed(tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else {
					if err.Field != tt.errorField {
						t.Errorf("Expected error field %q, got %q", tt.errorField, err.Field)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	setDefaults()

	if viper.GetString("default_region") != "ca-central-1" {
		t.Errorf("Default region not set correctly: %q", viper.GetString("default_region"))
	}

	if viper.GetString("sso.region") != "ca-central-1" {
		t.Errorf("SSO region default not set correctly: %q", viper.GetString("sso.region"))
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config with SSO",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "https://example.awsapps.com/start",
					Region:   "us-east-1",
				},
			},
			expectError: false,
		},
		{
			name: "empty SSO start URL",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "",
					Region:   "us-east-1",
				},
			},
			expectError: false, // Empty SSO is allowed (first run)
		},
		{
			name: "SSO URL without region",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "https://example.awsapps.com/start",
					Region:   "",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCreateSampleConfig(t *testing.T) {
	tempDir := t.TempDir()

	configPath := filepath.Join(tempDir, ".ztictl.yaml")
	err := CreateSampleConfig(configPath)
	if err != nil {
		t.Fatalf("CreateSampleConfig failed: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Check file content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if len(content) == 0 {
		t.Error("config file is empty")
	}

	// Check for expected sections
	if !contains(string(content), "sso:") || !contains(string(content), "logging:") || !contains(string(content), "system:") {
		t.Error("config file missing expected sections")
	}
}

// Helper functions
func contains(s, substr string) bool {
	return indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestExpandPathTildeExpansion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantHome bool
	}{
		{
			name:     "tilde with slash",
			input:    "~/test",
			wantHome: true,
		},
		{
			name:     "tilde alone",
			input:    "~",
			wantHome: true,
		},
		{
			name:     "no tilde",
			input:    "/absolute/path",
			wantHome: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			home := getUserHomeForTest()

			if tt.wantHome {
				if home != "" && !contains(result, home) {
					t.Errorf("expandPath(%q) should contain home directory", tt.input)
				}
			} else {
				if result != tt.input {
					t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.input)
				}
			}
		})
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name         string
		createFile   bool
		expectExists bool
	}{
		{
			name:         "config file exists",
			createFile:   true,
			expectExists: true,
		},
		{
			name:         "config file doesn't exist",
			createFile:   false,
			expectExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			tempDir := t.TempDir()

			// Temporarily override HOME
			origHome := os.Getenv("HOME")
			origUserProfile := os.Getenv("USERPROFILE")
			os.Setenv("HOME", tempDir)
			os.Setenv("USERPROFILE", tempDir)
			defer func() {
				os.Setenv("HOME", origHome)
				os.Setenv("USERPROFILE", origUserProfile)
			}()

			if tt.createFile {
				configPath := filepath.Join(tempDir, ".ztictl.yaml")
				err := os.WriteFile(configPath, []byte("test: value"), 0600)
				if err != nil {
					t.Fatalf("Failed to create config file: %v", err)
				}
			}

			result := Exists()
			if result != tt.expectExists {
				t.Errorf("Exists() = %v, want %v", result, tt.expectExists)
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	// Save original home
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	testHome := "/test/home"
	os.Setenv("HOME", testHome)
	os.Setenv("USERPROFILE", testHome)

	expected := filepath.Join(testHome, ".ztictl.yaml")
	result := getConfigPath()

	if result != expected {
		t.Errorf("getConfigPath() = %q, want %q", result, expected)
	}
}

func TestWriteInteractiveConfig(t *testing.T) {
	// Save original home
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")

	tempDir := t.TempDir()

	// Set temporary home
	os.Setenv("HOME", tempDir)
	os.Setenv("USERPROFILE", tempDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	config := &Config{
		SSO: SSOConfig{
			StartURL: "https://d-test123.awsapps.com/start",
			Region:   "us-west-2",
		},
		DefaultRegion: "us-east-1",
		Logging: LoggingConfig{
			Directory:   filepath.Join(tempDir, "logs"),
			FileLogging: true,
			Level:       "debug",
		},
		System: SystemConfig{
			IAMPropagationDelay: 10,
			S3BucketPrefix:      "test-bucket-prefix",
			FileSizeThreshold:   1048576,
			TempDirectory:       filepath.Join(tempDir, "tmp"),
		},
	}

	err := writeInteractiveConfig(config)
	if err != nil {
		t.Fatalf("writeInteractiveConfig failed: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tempDir, ".ztictl.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	expectedValues := []string{
		"https://d-test123.awsapps.com/start",
		"us-west-2",
		"us-east-1",
		filepath.Join(tempDir, "logs"),
		"debug",
		"10",
		"test-bucket-prefix",
		filepath.Join(tempDir, "tmp"), // Verify temp_directory is written
	}

	for _, val := range expectedValues {
		if !contains(string(content), val) {
			t.Errorf("Config file missing expected value: %q", val)
		}
	}
}

func TestCreateSampleConfigErrorHandling(t *testing.T) {
	// Test with invalid path
	invalidPath := "/nonexistent/directory/config.yaml"
	err := CreateSampleConfig(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}
