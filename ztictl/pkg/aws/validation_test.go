package aws

import (
	"testing"
)

func TestIsValidAWSRegion(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		expected bool
	}{
		// Valid regions
		{"Valid US East 1", "us-east-1", true},
		{"Valid US West 2", "us-west-2", true},
		{"Valid Canada Central 1", "ca-central-1", true},
		{"Valid EU West 1", "eu-west-1", true},
		{"Valid AP Southeast 1", "ap-southeast-1", true},
		{"Valid GovCloud East", "us-gov-east-1", true},
		{"Valid GovCloud West", "us-gov-west-1", true},
		{"Valid EU North 1", "eu-north-1", true},
		{"Valid AP Northeast 1", "ap-northeast-1", true},
		{"Valid AP Southeast 2", "ap-southeast-2", true},
		{"Valid AP South 1", "ap-south-1", true},
		{"Valid EU Central 1", "eu-central-1", true},
		{"Valid EU West 2", "eu-west-2", true},
		{"Valid EU West 3", "eu-west-3", true},
		{"Valid EU South 1", "eu-south-1", true},
		{"Valid ME South 1", "me-south-1", true},
		{"Valid ME Central 1", "me-central-1", true},
		{"Valid AF South 1", "af-south-1", true},
		{"Valid CN North 1", "cn-north-1", true},
		{"Valid CN Northwest 1", "cn-northwest-1", true},
		{"Valid SA East 1", "sa-east-1", true},

		// Invalid regions - should be rejected
		{"Invalid trailing digits", "ca-central-11", false},
		{"Invalid too short", "ca", false},
		{"Invalid format", "useast1", false},
		{"Invalid format", "us-east", false},
		{"Invalid format", "us-east-", false},
		{"Invalid format", "us-east-1-extra", false},
		{"Invalid prefix", "xx-east-1", false},
		{"Invalid direction", "us-invalid-1", false},
		{"Invalid number", "us-east-0", false},
		{"Invalid number", "us-east-100", false},
		{"Invalid number", "us-east-abc", false},
		{"Empty string", "", false},
		{"Invalid GovCloud", "us-gov-invalid-1", false},
		{"Invalid GovCloud direction", "us-gov-north-1", false},
		{"Invalid GovCloud format", "us-gov-east", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidAWSRegion(tt.region)
			if result != tt.expected {
				t.Errorf("IsValidAWSRegion(%q) = %v, expected %v", tt.region, result, tt.expected)
			}
		})
	}
}

func TestIsPlaceholderSSOURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"Placeholder URL", "https://d-xxxxxxxxxx.awsapps.com/start", true},
		{"Valid SSO URL", "https://d-1234567890.awsapps.com/start", false},
		{"Valid SSO URL with company name", "https://zsoftly.awsapps.com/start", false},
		{"Invalid URL", "https://d-xxxxxxxxxx.awsapps.com/", false},
		{"Invalid URL", "https://d-xxxxxxxxxx.awsapps.com/end", false},
		{"Empty URL", "", false},
		{"Different placeholder", "https://d-abcdefghij.awsapps.com/start", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPlaceholderSSOURL(tt.url)
			if result != tt.expected {
				t.Errorf("IsPlaceholderSSOURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestValidateSSOURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
		errorField  string
	}{
		// Valid URLs
		{"Valid SSO URL", "https://d-1234567890.awsapps.com/start", false, ""},
		{"Valid SSO URL with company name", "https://zsoftly.awsapps.com/start", false, ""},
		{"Valid SSO URL with different domain", "https://mycompany.awsapps.com/start", false, ""},

		// Invalid URLs - should return errors
		{"Placeholder URL", "https://d-xxxxxxxxxx.awsapps.com/start", true, "SSO start URL"},
		{"Empty URL", "", true, "SSO start URL"},
		{"Invalid protocol", "http://d-1234567890.awsapps.com/start", true, "SSO start URL"},
		{"Invalid protocol", "ftp://d-1234567890.awsapps.com/start", true, "SSO start URL"},
		{"Invalid domain", "https://d-1234567890.awsapps.com/", true, "SSO start URL"},
		{"Invalid domain", "https://d-1234567890.awsapps.com/end", true, "SSO start URL"},
		{"Invalid domain", "https://d-1234567890.example.com/start", true, "SSO start URL"},
		{"Invalid format", "not-a-url", true, "SSO start URL"},
		{"Invalid format", "https://", true, "SSO start URL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSSOURL(tt.url)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateSSOURL(%q) expected error but got none", tt.url)
					return
				}
				
				if valErr, ok := err.(*ValidationError); ok {
					if valErr.Field != tt.errorField {
						t.Errorf("ValidateSSOURL(%q) error field = %q, expected %q", tt.url, valErr.Field, tt.errorField)
					}
				} else {
					t.Errorf("ValidateSSOURL(%q) returned error of wrong type: %T", tt.url, err)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateSSOURL(%q) unexpected error: %v", tt.url, err)
				}
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "test field",
		Value:   "test value",
		Message: "test message",
	}
	
	expected := "test field 'test value' is invalid: test message"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %q, expected %q", err.Error(), expected)
	}
}