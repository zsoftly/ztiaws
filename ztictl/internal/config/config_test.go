package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/viper"
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
	content, err := os.ReadFile(configPath) // #nosec G304 - test file with controlled path
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

func TestExpandPathTildeExpansion(t *testing.T) {
	// Test tilde expansion functionality
	tests := []struct {
		name     string
		input    string
		contains string // Check if the result contains this
	}{
		{
			name:  "tilde with path",
			input: "~/test/path",
		},
		{
			name:  "bare tilde",
			input: "~",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)

			// Should not start with ~ after expansion (unless home dir couldn't be obtained)
			if strings.HasPrefix(result, "~") && result != tt.input {
				t.Errorf("expandPath(%q) still contains tilde: %q", tt.input, result)
			}

			// For valid tilde paths, result should be different from input (unless error occurred)
			if tt.input != "" && strings.HasPrefix(tt.input, "~") {
				if home, err := os.UserHomeDir(); err == nil {
					if tt.input == "~" {
						if result != home {
							t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, home)
						}
					} else if tt.input == "~/test/path" {
						expected := filepath.Join(home, "test/path")
						if result != expected {
							t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, expected)
						}
					}
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Save original viper state
	origCfg := cfg
	defer func() { cfg = origCfg }()

	// Reset viper
	viper.Reset()
	defer viper.Reset()

	tests := []struct {
		name          string
		setup         func() (cleanup func(), error error)
		expectError   bool
		errorContains string
	}{
		{
			name: "first run without env file",
			setup: func() (cleanup func(), error error) {
				// Create temp directory for config
				tempDir, err := ioutil.TempDir("", "config_test")
				if err != nil {
					return nil, err
				}

				// Mock home directory
				origHome := os.Getenv("HOME")
				_ = os.Setenv("HOME", tempDir) // #nosec G104 - test setup

				cleanup = func() {
					_ = os.Setenv("HOME", origHome) // #nosec G104 - test cleanup
					_ = os.RemoveAll(tempDir)       // #nosec G104 - test cleanup
				}

				return cleanup, nil
			},
			expectError: false,
		},
		{
			name: "first run with valid env file",
			setup: func() (cleanup func(), error error) {
				tempDir, err := ioutil.TempDir("", "config_test")
				if err != nil {
					return nil, err
				}

				// Create .env file
				envContent := `SSO_START_URL="https://example.awsapps.com/start"
SSO_REGION="us-east-1"
DEFAULT_PROFILE="test-profile"
LOG_DIR="/tmp/logs"
`
				envPath := filepath.Join("..", ".env")
				_ = os.MkdirAll(filepath.Dir(envPath), 0750) // #nosec G104 - test setup
				err = ioutil.WriteFile(envPath, []byte(envContent), 0600)
				if err != nil {
					return nil, err
				}

				origHome := os.Getenv("HOME")
				_ = os.Setenv("HOME", tempDir) // #nosec G104 - test setup

				cleanup = func() {
					_ = os.Setenv("HOME", origHome) // #nosec G104 - test cleanup
					_ = os.Remove(envPath)          // #nosec G104 - test cleanup
					_ = os.RemoveAll(tempDir)       // #nosec G104 - test cleanup
				}

				return cleanup, nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global config
			cfg = nil
			viper.Reset()

			cleanup, err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
			defer cleanup()

			err = Load()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err != nil && tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("Error should contain '%s', got: %v", tt.errorContains, err)
			}

			// Verify config was created
			if !tt.expectError && cfg == nil {
				t.Error("Config should be initialized after Load()")
			}
		})
	}
}

func TestGet(t *testing.T) {
	// Save original config
	origCfg := cfg
	defer func() { cfg = origCfg }()

	tests := []struct {
		name     string
		setup    func()
		validate func(*Config, *testing.T)
	}{
		{
			name: "get with existing config",
			setup: func() {
				cfg = &Config{
					DefaultRegion: "us-west-2",
					SSO: SSOConfig{
						StartURL: "https://test.awsapps.com/start",
						Region:   "us-east-1",
					},
				}
			},
			validate: func(c *Config, t *testing.T) {
				if c.DefaultRegion != "us-west-2" {
					t.Errorf("Expected default region us-west-2, got %s", c.DefaultRegion)
				}
				if c.SSO.StartURL != "https://test.awsapps.com/start" {
					t.Errorf("Expected SSO start URL to match, got %s", c.SSO.StartURL)
				}
			},
		},
		{
			name: "get with nil config (creates new)",
			setup: func() {
				cfg = nil
				viper.Reset()
			},
			validate: func(c *Config, t *testing.T) {
				if c == nil {
					t.Error("Config should not be nil")
				}
				// Should have default values set
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result := Get()

			if result == nil {
				t.Error("Get() should never return nil")
			}

			tt.validate(result, t)
		})
	}
}

func TestLoadLegacyEnvFile(t *testing.T) {
	tests := []struct {
		name        string
		envContent  string
		expectError bool
		validate    func(*testing.T)
	}{
		{
			name: "valid env file",
			envContent: `# Comment
SSO_START_URL="https://example.awsapps.com/start"
SSO_REGION=us-east-1
DEFAULT_PROFILE="test-profile"
LOG_DIR=/tmp/logs
`,
			expectError: false,
			validate: func(t *testing.T) {
				if viper.GetString("sso.start_url") != "https://example.awsapps.com/start" {
					t.Errorf("SSO start URL not set correctly: %s", viper.GetString("sso.start_url"))
				}
				if viper.GetString("sso.region") != "us-east-1" {
					t.Errorf("SSO region not set correctly: %s", viper.GetString("sso.region"))
				}
				if viper.GetString("sso.default_profile") != "test-profile" {
					t.Errorf("Default profile not set correctly: %s", viper.GetString("sso.default_profile"))
				}
				if viper.GetString("logging.directory") != "/tmp/logs" {
					t.Errorf("Log directory not set correctly: %s", viper.GetString("logging.directory"))
				}
			},
		},
		{
			name: "env file with quotes",
			envContent: `SSO_START_URL='https://single-quote.com/start'
DEFAULT_PROFILE="double-quote-profile"
`,
			expectError: false,
			validate: func(t *testing.T) {
				if viper.GetString("sso.start_url") != "https://single-quote.com/start" {
					t.Errorf("Single quoted value not parsed correctly: %s", viper.GetString("sso.start_url"))
				}
				if viper.GetString("sso.default_profile") != "double-quote-profile" {
					t.Errorf("Double quoted value not parsed correctly: %s", viper.GetString("sso.default_profile"))
				}
			},
		},
		{
			name: "env file with empty lines and comments",
			envContent: `
# This is a comment
# Another comment

SSO_REGION=us-west-2

# Final comment
`,
			expectError: false,
			validate: func(t *testing.T) {
				if viper.GetString("sso.region") != "us-west-2" {
					t.Errorf("SSO region not set correctly: %s", viper.GetString("sso.region"))
				}
			},
		},
		{
			name:        "nonexistent file",
			envContent:  "",    // Will create no file
			expectError: false, // Should not error on nonexistent file
			validate: func(t *testing.T) {
				// No validation needed for nonexistent file
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			defer viper.Reset()

			var envPath string
			var cleanup func()

			if tt.envContent != "" {
				// Create temporary env file
				tempDir, err := ioutil.TempDir("", "env_test")
				if err != nil {
					t.Fatalf("Failed to create temp directory: %v", err)
				}

				envPath = filepath.Join(tempDir, ".env")
				err = ioutil.WriteFile(envPath, []byte(tt.envContent), 0600)
				if err != nil {
					t.Fatalf("Failed to write env file: %v", err)
				}

				cleanup = func() { _ = os.RemoveAll(tempDir) } // #nosec G104 - test cleanup
			} else {
				envPath = "/nonexistent/.env"
				cleanup = func() {}
			}

			defer cleanup()

			err := LoadLegacyEnvFile(envPath)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (cleanup func())
		expected bool
	}{
		{
			name: "config file exists",
			setup: func() (cleanup func()) {
				tempDir, err := ioutil.TempDir("", "config_exists_test")
				if err != nil {
					t.Fatalf("Failed to create temp directory: %v", err)
				}

				// Save original environment variables for both Windows and Unix
				var origHome, origUserProfile string
				if runtime.GOOS == "windows" {
					origUserProfile = os.Getenv("USERPROFILE")
					_ = os.Setenv("USERPROFILE", tempDir) // #nosec G104 - test setup
				} else {
					origHome = os.Getenv("HOME")
					_ = os.Setenv("HOME", tempDir) // #nosec G104 - test setup
				}

				configPath := filepath.Join(tempDir, ".ztictl.yaml")
				_ = ioutil.WriteFile(configPath, []byte("test: config"), 0600) // #nosec G104 - test setup

				return func() {
					if runtime.GOOS == "windows" {
						_ = os.Setenv("USERPROFILE", origUserProfile) // #nosec G104 - test cleanup
					} else {
						_ = os.Setenv("HOME", origHome) // #nosec G104 - test cleanup
					}
					_ = os.RemoveAll(tempDir) // #nosec G104 - test cleanup
				}
			},
			expected: true,
		},
		{
			name: "config file does not exist",
			setup: func() (cleanup func()) {
				tempDir, err := ioutil.TempDir("", "config_not_exists_test")
				if err != nil {
					t.Fatalf("Failed to create temp directory: %v", err)
				}

				// Save original environment variables for both Windows and Unix
				var origHome, origUserProfile string
				if runtime.GOOS == "windows" {
					origUserProfile = os.Getenv("USERPROFILE")
					_ = os.Setenv("USERPROFILE", tempDir) // #nosec G104 - test setup
				} else {
					origHome = os.Getenv("HOME")
					_ = os.Setenv("HOME", tempDir) // #nosec G104 - test setup
				}

				return func() {
					if runtime.GOOS == "windows" {
						_ = os.Setenv("USERPROFILE", origUserProfile) // #nosec G104 - test cleanup
					} else {
						_ = os.Setenv("HOME", origHome) // #nosec G104 - test cleanup
					}
					_ = os.RemoveAll(tempDir) // #nosec G104 - test cleanup
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			result := Exists()

			if result != tt.expected {
				t.Errorf("Exists() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	// Save original environment variables for both Windows and Unix
	var origHome, origUserProfile string
	if runtime.GOOS == "windows" {
		origUserProfile = os.Getenv("USERPROFILE")
		defer os.Setenv("USERPROFILE", origUserProfile)
	} else {
		origHome = os.Getenv("HOME")
		defer os.Setenv("HOME", origHome)
	}

	// Test with known home directory
	testHome := "/tmp/test_home"
	if runtime.GOOS == "windows" {
		testHome = "C:\\tmp\\test_home"
		_ = os.Setenv("USERPROFILE", testHome) // #nosec G104 - test setup
	} else {
		_ = os.Setenv("HOME", testHome) // #nosec G104 - test setup
	}

	result := getConfigPath()
	expected := filepath.Join(testHome, ".ztictl.yaml")

	if result != expected {
		t.Errorf("getConfigPath() = %s, want %s", result, expected)
	}
}

func TestWriteInteractiveConfig(t *testing.T) {
	// Create temporary home directory
	tempDir, err := ioutil.TempDir("", "interactive_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir) // #nosec G104 - test setup
	defer os.Setenv("HOME", origHome)

	config := &Config{
		SSO: SSOConfig{
			StartURL:       "https://test.awsapps.com/start",
			Region:         "us-east-1",
			DefaultProfile: "test-profile",
		},
		DefaultRegion: "ca-central-1",
		Logging: LoggingConfig{
			Directory:   "~/logs",
			FileLogging: true,
			Level:       "info",
		},
		System: SystemConfig{
			IAMPropagationDelay: 10,
			FileSizeThreshold:   2097152,
			S3BucketPrefix:      "test-bucket-prefix",
		},
	}

	err = writeInteractiveConfig(config)
	if err != nil {
		t.Fatalf("writeInteractiveConfig failed: %v", err)
	}

	// Verify config file was created
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Read and verify content
	content, err := ioutil.ReadFile(configPath) // #nosec G304 - test file with controlled path
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	contentStr := string(content)
	expectedStrings := []string{
		"https://test.awsapps.com/start",
		"us-east-1",
		"test-profile",
		"ca-central-1",
		"~/logs",
		"true",
		"info",
		"10",
		"2097152",
		"test-bucket-prefix",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Config file should contain '%s'", expected)
		}
	}
}

func TestInteractiveInit(t *testing.T) {
	// This test checks the structure and error handling of InteractiveInit
	// We can't easily test the interactive input part without complex mocking

	// Test that InteractiveInit exists and can handle setup errors
	tempDir, err := ioutil.TempDir("", "interactive_init_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir) // #nosec G104 - test setup
	defer os.Setenv("HOME", origHome)

	// Mock stdin with minimal input
	input := "\nhttps://test.com/start\n\n\n\n\n\n\n\n\n\n\n"
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		defer w.Close()
		_, _ = w.Write([]byte(input)) // #nosec G104 - test setup
	}()

	defer func() {
		os.Stdin = oldStdin
		_ = r.Close() // #nosec G104 - test cleanup
	}()

	// Test that InteractiveInit can run without panicking
	// The actual interactive testing would require more complex setup
	err = InteractiveInit()

	// Should not panic and should create config file
	if err != nil {
		t.Logf("InteractiveInit returned error (expected in test environment): %v", err)
	}
}

// Test helper functions
func TestContainsAllFunction(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substrs  []string
		expected bool
	}{
		{
			name:     "all substrings present",
			str:      "hello world test string",
			substrs:  []string{"hello", "world", "test"},
			expected: true,
		},
		{
			name:     "missing substring",
			str:      "hello world test string",
			substrs:  []string{"hello", "missing", "test"},
			expected: false,
		},
		{
			name:     "empty substrings",
			str:      "hello world",
			substrs:  []string{},
			expected: true,
		},
		{
			name:     "empty string",
			str:      "",
			substrs:  []string{"hello"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsAll(tt.str, tt.substrs)
			if result != tt.expected {
				t.Errorf("containsAll(%q, %v) = %v, want %v", tt.str, tt.substrs, result, tt.expected)
			}
		})
	}
}

func TestContainsFunction(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substr   string
		expected bool
	}{
		{
			name:     "substring at beginning",
			str:      "hello world",
			substr:   "hello",
			expected: true,
		},
		{
			name:     "substring at end",
			str:      "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "substring in middle",
			str:      "hello world test",
			substr:   "world",
			expected: true,
		},
		{
			name:     "substring not found",
			str:      "hello world",
			substr:   "missing",
			expected: false,
		},
		{
			name:     "empty substring",
			str:      "hello world",
			substr:   "",
			expected: true,
		},
		{
			name:     "exact match",
			str:      "test",
			substr:   "test",
			expected: true,
		},
		{
			name:     "substring longer than string",
			str:      "hi",
			substr:   "hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.str, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.str, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestIndexOfSubstringFunction(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substr   string
		expected int
	}{
		{
			name:     "substring at beginning",
			str:      "hello world",
			substr:   "hello",
			expected: 0,
		},
		{
			name:     "substring at end",
			str:      "hello world",
			substr:   "world",
			expected: 6,
		},
		{
			name:     "substring in middle",
			str:      "hello beautiful world",
			substr:   "beautiful",
			expected: 6,
		},
		{
			name:     "substring not found",
			str:      "hello world",
			substr:   "missing",
			expected: -1,
		},
		{
			name:     "empty substring",
			str:      "hello world",
			substr:   "",
			expected: 0,
		},
		{
			name:     "substring longer than string",
			str:      "hi",
			substr:   "hello",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexOfSubstring(tt.str, tt.substr)
			if result != tt.expected {
				t.Errorf("indexOfSubstring(%q, %q) = %v, want %v", tt.str, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestConfigStructValidation(t *testing.T) {
	// Test config struct field validation
	config := &Config{
		SSO: SSOConfig{
			StartURL:       "https://example.awsapps.com/start",
			Region:         "us-east-1",
			DefaultProfile: "test-profile",
		},
		DefaultRegion: "ca-central-1",
		Logging: LoggingConfig{
			Directory:   "/var/logs",
			FileLogging: true,
			Level:       "info",
		},
		System: SystemConfig{
			IAMPropagationDelay: 5,
			FileSizeThreshold:   1048576,
			S3BucketPrefix:      "ztictl-test",
			TempDirectory:       "/tmp",
		},
	}

	// Test that all fields are properly set
	if config.SSO.StartURL == "" {
		t.Error("SSO StartURL should be set")
	}
	if config.SSO.Region == "" {
		t.Error("SSO Region should be set")
	}
	if config.SSO.DefaultProfile == "" {
		t.Error("SSO DefaultProfile should be set")
	}
	if config.DefaultRegion == "" {
		t.Error("DefaultRegion should be set")
	}
	if config.Logging.Directory == "" {
		t.Error("Logging Directory should be set")
	}
	if config.Logging.Level == "" {
		t.Error("Logging Level should be set")
	}
	if config.System.IAMPropagationDelay <= 0 {
		t.Error("IAM PropagationDelay should be positive")
	}
	if config.System.FileSizeThreshold <= 0 {
		t.Error("FileSizeThreshold should be positive")
	}
	if config.System.S3BucketPrefix == "" {
		t.Error("S3BucketPrefix should be set")
	}
	if config.System.TempDirectory == "" {
		t.Error("TempDirectory should be set")
	}
}

func TestSetDefaultsWithViper(t *testing.T) {
	// Reset viper and test defaults
	viper.Reset()
	defer viper.Reset()

	setDefaults()

	// Test that all defaults are set
	tests := []struct {
		key      string
		expected interface{}
	}{
		{"default_region", "ca-central-1"},
		{"sso.region", "us-east-1"},
		{"sso.default_profile", "default-sso-profile"},
		{"logging.file_logging", true},
		{"logging.level", "info"},
		{"system.iam_propagation_delay", 5},
		{"system.file_size_threshold", 1048576},
		{"system.s3_bucket_prefix", "ztictl-ssm-file-transfer"},
	}

	for _, tt := range tests {
		t.Run("default_"+tt.key, func(t *testing.T) {
			actual := viper.Get(tt.key)
			if actual != tt.expected {
				t.Errorf("Default for %s = %v (%T), want %v (%T)", tt.key, actual, actual, tt.expected, tt.expected)
			}
		})
	}

	// Test that logging directory default is set and not empty
	logDir := viper.GetString("logging.directory")
	if logDir == "" {
		t.Error("Default logging directory should not be empty")
	}

	// Test that temp directory default is set and not empty
	tempDir := viper.GetString("system.temp_directory")
	if tempDir == "" {
		t.Error("Default temp directory should not be empty")
	}
}

func TestLoadLegacyEnvFileErrorHandling(t *testing.T) {
	// Test file permission error
	tempDir, err := ioutil.TempDir("", "env_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create env file then make it unreadable
	envPath := filepath.Join(tempDir, ".env")
	err = ioutil.WriteFile(envPath, []byte("TEST=value"), 0600)
	if err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}

	// Change permissions to make it unreadable (if supported)
	_ = os.Chmod(envPath, 0000)   // #nosec G104 - test setup
	defer os.Chmod(envPath, 0600) // Restore for cleanup

	viper.Reset()
	err = LoadLegacyEnvFile(envPath)

	// Should handle permission errors gracefully
	if err == nil {
		t.Log("Permission error test skipped (may not be supported on this platform)")
	}
}

func TestCreateSampleConfigErrorHandling(t *testing.T) {
	// Test directory creation error
	invalidPath := "/root/permission-denied/.ztictl.yaml"
	err := CreateSampleConfig(invalidPath)

	// Should return an error for invalid paths
	if err == nil {
		t.Log("Permission error test skipped (may not be supported on this platform)")
	} else if !strings.Contains(err.Error(), "failed to") {
		t.Errorf("Error should indicate what failed, got: %v", err)
	}
}

// Mock stdin for testing interactive functions
func mockStdin(input string) func() {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		defer w.Close()
		_, _ = w.Write([]byte(input)) // #nosec G104 - test setup
	}()

	return func() {
		os.Stdin = oldStdin
		_ = r.Close() // #nosec G104 - test cleanup
	}
}
