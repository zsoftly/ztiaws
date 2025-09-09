package config

import (
	"testing"
)

func TestIsValidRegion(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		expected bool
	}{
		// Valid regions
		{"Valid US East", "us-east-1", true},
		{"Valid US West", "us-west-2", true},
		{"Valid EU West", "eu-west-1", true},
		{"Valid EU Central", "eu-central-1", true},
		{"Valid Asia Pacific", "ap-southeast-1", true},
		{"Valid Asia Pacific Northeast", "ap-northeast-2", true},
		{"Valid Canada", "ca-central-1", true},
		{"Valid South America", "sa-east-1", true},
		{"Valid Middle East", "me-south-1", true},
		{"Valid Africa", "af-south-1", true},
		{"Valid China", "cn-north-1", true},
		{"Valid US GovCloud", "us-gov-west-1", true},
		{"Valid US GovCloud East", "us-gov-east-1", true},

		// Invalid GovCloud regions
		{"Invalid GovCloud north", "us-gov-north-1", false},
		{"Invalid GovCloud central", "us-gov-central-1", false},

		// Invalid regions
		{"Invalid prefix caa", "caa-central-1", false},
		{"Invalid prefix xx", "xx-central-1", false},
		{"Invalid direction", "us-invalid-1", false},
		{"Invalid direction too long", "ca-centralllll-1", false},
		{"Invalid format no hyphens", "useast1", false},
		{"Invalid format one hyphen", "us-east1", false},
		{"Invalid format four parts", "us-east-west-1", false},
		{"Invalid number part", "us-east-a", false},
		{"Invalid number part too long", "us-east-123", false},
		{"Empty string", "", false},
		{"Invalid special chars", "us-east-1!", false},
		{"Invalid with spaces", "us east 1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidRegion(tt.region)
			if result != tt.expected {
				t.Errorf("isValidRegion(%q) = %v, want %v", tt.region, result, tt.expected)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		// Valid URLs
		{"Valid HTTPS URL", "https://example.com", true},
		{"Valid HTTPS with path", "https://example.com/path", true},
		{"Valid HTTPS with port", "https://example.com:8080", true},
		{"Valid HTTP URL", "http://example.com", true},
		{"Valid AWS SSO URL", "https://d-1234567890.awsapps.com/start", true},
		{"Valid localhost", "http://localhost:3000", true},

		// Invalid URLs
		{"Missing protocol", "example.com", false},
		{"Invalid protocol", "ftp://example.com", false},
		{"Just protocol", "https://", false},
		{"Empty string", "", false},
		{"With quotes", "\"https://example.com\"", false},
		{"Random text", "not a url", false},
		{"File protocol", "file:///path/to/file", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidURL(tt.url)
			if result != tt.expected {
				t.Errorf("isValidURL(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		inputType string
		wantErr   bool
	}{
		// URL validation
		{"Valid URL", "https://example.com", "url", false},
		{"Invalid URL", "not-a-url", "url", true},
		{"Empty URL allowed", "", "url", false}, // Empty is allowed

		// Region validation
		{"Valid region", "us-east-1", "region", false},
		{"Invalid region", "invalid-region", "region", true},
		{"Empty region allowed", "", "region", false}, // Empty is allowed

		// Path validation - no longer checking for backslashes
		{"Path with forward slashes", "/path/to/file", "path", false},
		{"Path with backslashes", "C:\\path\\to\\file", "path", false}, // Now allowed
		{"Mixed slashes", "C:\\path/to\\file", "path", false},          // Now allowed
		{"Empty path", "", "path", false},

		// Unknown type (no validation)
		{"Unknown type", "anything", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInput(tt.input, tt.inputType)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateInput(%q, %q) error = %v, wantErr %v",
					tt.input, tt.inputType, err, tt.wantErr)
			}
		})
	}
}

func TestValidateLoadedConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid config",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "https://example.awsapps.com/start",
					Region:   "us-east-1",
				},
				DefaultRegion: "us-west-2",
				Logging: LoggingConfig{
					Directory: "/var/log/ztictl",
				},
				System: SystemConfig{
					TempDirectory: "/tmp",
				},
			},
			wantErr: false,
		},
		{
			name: "Empty config (valid)",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "",
					Region:   "",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid SSO URL",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "not-a-url",
					Region:   "us-east-1",
				},
			},
			wantErr: true,
			errMsg:  "SSO start URL is not a valid URL",
		},
		{
			name: "SSO URL with double quotes",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "\"\"https://example.com\"\"",
					Region:   "us-east-1",
				},
			},
			wantErr: true,
			errMsg:  "SSO start URL contains invalid quotes",
		},
		{
			name: "Invalid SSO region",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "https://example.com",
					Region:   "invalid-region",
				},
			},
			wantErr: true,
			errMsg:  "SSO region is not valid",
		},
		{
			name: "Invalid default region",
			config: &Config{
				SSO: SSOConfig{
					StartURL: "https://example.com",
					Region:   "us-east-1",
				},
				DefaultRegion: "xx-invalid-1",
			},
			wantErr: true,
			errMsg:  "Default region is not valid",
		},
		{
			name: "Paths with forward slashes (valid)",
			config: &Config{
				Logging: LoggingConfig{
					Directory: "C:/Users/test/logs",
				},
				System: SystemConfig{
					TempDirectory: "C:/Windows/Temp",
				},
			},
			wantErr: false,
		},
		{
			name: "Paths with backslashes (now valid)",
			config: &Config{
				Logging: LoggingConfig{
					Directory: "C:\\Users\\test\\logs",
				},
				System: SystemConfig{
					TempDirectory: "C:\\Windows\\Temp",
				},
			},
			wantErr: false, // Changed to allow backslashes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoadedConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLoadedConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
				t.Errorf("validateLoadedConfig() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}
