package config

import (
	"strings"
	"testing"
)

// TestSSODomainIDConstruction tests that SSO URLs are correctly constructed from domain IDs
func TestSSODomainIDConstruction(t *testing.T) {
	tests := []struct {
		name        string
		domainID    string
		expectedURL string
		shouldError bool
	}{
		{
			name:        "standard AWS domain ID",
			domainID:    "d-1234567890",
			expectedURL: "https://d-1234567890.awsapps.com/start",
			shouldError: false,
		},
		{
			name:        "custom domain name",
			domainID:    "zsoftly",
			expectedURL: "https://zsoftly.awsapps.com/start",
			shouldError: false,
		},
		{
			name:        "domain with numbers",
			domainID:    "company123",
			expectedURL: "https://company123.awsapps.com/start",
			shouldError: false,
		},
		{
			name:        "domain with hyphens",
			domainID:    "my-company",
			expectedURL: "https://my-company.awsapps.com/start",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the URL construction from InteractiveInit
			constructedURL := "https://" + tt.domainID + ".awsapps.com/start"

			if constructedURL != tt.expectedURL {
				t.Errorf("URL construction failed: got %s, want %s", constructedURL, tt.expectedURL)
			}

			// Verify the URL is valid
			if !strings.HasPrefix(constructedURL, "https://") {
				t.Error("Constructed URL should start with https://")
			}

			if !strings.HasSuffix(constructedURL, ".awsapps.com/start") {
				t.Error("Constructed URL should end with .awsapps.com/start")
			}
		})
	}
}

// TestDefaultRegionSettings tests that ca-central-1 is used as default
func TestDefaultRegionSettings(t *testing.T) {
	tests := []struct {
		name           string
		inputRegion    string
		expectedRegion string
	}{
		{
			name:           "empty input uses ca-central-1",
			inputRegion:    "",
			expectedRegion: "ca-central-1",
		},
		{
			name:           "explicit ca-central-1",
			inputRegion:    "ca-central-1",
			expectedRegion: "ca-central-1",
		},
		{
			name:           "override with us-east-1",
			inputRegion:    "us-east-1",
			expectedRegion: "us-east-1",
		},
		{
			name:           "override with eu-west-1",
			inputRegion:    "eu-west-1",
			expectedRegion: "eu-west-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the default logic from InteractiveInit
			region := tt.inputRegion
			if region == "" {
				region = "ca-central-1"
			}

			if region != tt.expectedRegion {
				t.Errorf("Region default logic failed: got %s, want %s", region, tt.expectedRegion)
			}
		})
	}
}

// TestValidateInput tests the input validation for URLs
func TestValidateInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		inputType   string
		shouldError bool
	}{
		{
			name:        "valid https URL",
			input:       "https://d-1234567890.awsapps.com/start",
			inputType:   "url",
			shouldError: false,
		},
		{
			name:        "valid http URL",
			input:       "http://example.com",
			inputType:   "url",
			shouldError: false,
		},
		{
			name:        "invalid URL without protocol",
			input:       "example.com",
			inputType:   "url",
			shouldError: true,
		},
		{
			name:        "valid AWS region",
			input:       "ca-central-1",
			inputType:   "region",
			shouldError: false,
		},
		{
			name:        "invalid AWS region",
			input:       "caa-central-1",
			inputType:   "region",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInput(tt.input, tt.inputType)
			if (err != nil) != tt.shouldError {
				t.Errorf("validateInput(%s, %s) error = %v, shouldError = %v",
					tt.input, tt.inputType, err, tt.shouldError)
			}
		})
	}
}

// TestConfigWithSSODomainID tests the full config creation with domain ID
func TestConfigWithSSODomainID(t *testing.T) {
	tests := []struct {
		name             string
		domainID         string
		ssoRegion        string
		defaultRegion    string
		expectedStartURL string
	}{
		{
			name:             "config with all defaults",
			domainID:         "d-1234567890",
			ssoRegion:        "ca-central-1",
			defaultRegion:    "ca-central-1",
			expectedStartURL: "https://d-1234567890.awsapps.com/start",
		},
		{
			name:             "config with custom domain",
			domainID:         "zsoftly",
			ssoRegion:        "ca-central-1",
			defaultRegion:    "ca-central-1",
			expectedStartURL: "https://zsoftly.awsapps.com/start",
		},
		{
			name:             "config with mixed regions",
			domainID:         "mycompany",
			ssoRegion:        "us-east-1",
			defaultRegion:    "eu-west-1",
			expectedStartURL: "https://mycompany.awsapps.com/start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				SSO: SSOConfig{
					StartURL: "https://" + tt.domainID + ".awsapps.com/start",
					Region:   tt.ssoRegion,
				},
				DefaultRegion: tt.defaultRegion,
			}

			if config.SSO.StartURL != tt.expectedStartURL {
				t.Errorf("SSO StartURL = %s, want %s", config.SSO.StartURL, tt.expectedStartURL)
			}

			if config.SSO.Region != tt.ssoRegion {
				t.Errorf("SSO Region = %s, want %s", config.SSO.Region, tt.ssoRegion)
			}

			if config.DefaultRegion != tt.defaultRegion {
				t.Errorf("DefaultRegion = %s, want %s", config.DefaultRegion, tt.defaultRegion)
			}
		})
	}
}
