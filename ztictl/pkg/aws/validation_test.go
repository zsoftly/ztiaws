package aws

import (
	"testing"
)

func TestIsValidAWSRegionStrict(t *testing.T) {
	tests := []struct {
		name   string
		region string
		valid  bool
	}{
		// Valid standard regions
		{"Valid US East 1", "us-east-1", true},
		{"Valid US West 2", "us-west-2", true},
		{"Valid CA Central 1", "ca-central-1", true},
		{"Valid EU West 1", "eu-west-1", true},
		{"Valid EU Central 1", "eu-central-1", true},
		{"Valid AP Southeast 1", "ap-southeast-1", true},
		{"Valid AP Northeast 2", "ap-northeast-2", true},
		{"Valid SA East 1", "sa-east-1", true},
		{"Valid ME South 1", "me-south-1", true},
		{"Valid AF South 1", "af-south-1", true},
		{"Valid CN North 1", "cn-north-1", true},

		// Valid GovCloud regions
		{"Valid US GovCloud East 1", "us-gov-east-1", true},
		{"Valid US GovCloud West 1", "us-gov-west-1", true},

		// Invalid - wrong format
		{"Empty string", "", false},
		{"Single part", "us", false},
		{"Two parts", "us-east", false},
		{"Missing number", "us-east-", false},
		{"Too many parts", "us-east-1-extra", false},

		// Invalid - wrong prefix
		{"Invalid prefix xx", "xx-east-1", false},
		{"Invalid prefix ca alone", "ca-1", false},
		{"Typo caa", "caa-central-1", false},
		{"Typo uss", "uss-east-1", false},
		{"Invalid prefix abc", "abc-central-1", false},

		// Invalid - wrong direction
		{"Invalid direction", "us-center-1", false},
		{"Typo central", "ca-cent-1", false},
		{"Invalid direction middle", "us-middle-1", false},
		{"Wrong govcloud direction north", "us-gov-north-1", false},
		{"Wrong govcloud direction central", "us-gov-central-1", false},

		// Invalid - wrong number format
		{"Three digit number", "us-east-100", false},
		{"Zero", "us-east-0", false}, // Zero is invalid - AWS regions start from 1
		{"Letters in number", "us-east-1a", false},
		{"Special char in number", "us-east-1!", false},

		// Edge cases that should fail
		{"Looks valid but wrong", "ca-centralll-1", false},
		{"Extra hyphen", "us--east-1", false},
		{"Space in region", "us east 1", false},
		{"Tab in region", "us\teast\t1", false},
		{"Leading space", " us-east-1", false},
		{"Trailing space", "us-east-1 ", false},

		// Real regions that must pass
		{"Real CA Central", "ca-central-1", true},
		{"Real US East Ohio", "us-east-2", true},
		{"Real EU Ireland", "eu-west-1", true},
		{"Real EU Frankfurt", "eu-central-1", true},
		{"Real AP Singapore", "ap-southeast-1", true},
		{"Real AP Sydney", "ap-southeast-2", true},
		{"Real AP Tokyo", "ap-northeast-1", true},
		{"Real AP Mumbai", "ap-south-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidAWSRegion(tt.region)
			if result != tt.valid {
				t.Errorf("IsValidAWSRegion(%q) = %v, want %v", tt.region, result, tt.valid)
			}
		})
	}
}

func TestIsValidRegionShortcode(t *testing.T) {
	tests := []struct {
		name      string
		shortcode string
		valid     bool
	}{
		// Valid shortcodes
		{"Valid cac1", "cac1", true},
		{"Valid use1", "use1", true},
		{"Valid use2", "use2", true},
		{"Valid usw1", "usw1", true},
		{"Valid usw2", "usw2", true},
		{"Valid euw1", "euw1", true},
		{"Valid euc1", "euc1", true},
		{"Valid apse1", "apse1", true},

		// Invalid shortcodes
		{"Invalid empty", "", false},
		{"Invalid xxx", "xxx", false},
		{"Invalid ca", "ca", false},
		{"Invalid cac", "cac", false},
		{"Invalid caccc1", "caccc1", false},
		{"Invalid number", "123", false},
		{"Full region not shortcode", "ca-central-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidRegionShortcode(tt.shortcode)
			if result != tt.valid {
				t.Errorf("IsValidRegionShortcode(%q) = %v, want %v", tt.shortcode, result, tt.valid)
			}
		})
	}
}

func TestValidateRegionInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		// Valid shortcodes
		{"Shortcode cac1", "cac1", "ca-central-1", false},
		{"Shortcode use1", "use1", "us-east-1", false},
		{"Shortcode euw1", "euw1", "eu-west-1", false},

		// Valid full regions
		{"Full ca-central-1", "ca-central-1", "ca-central-1", false},
		{"Full us-east-1", "us-east-1", "us-east-1", false},
		{"Full eu-west-1", "eu-west-1", "eu-west-1", false},

		// With whitespace
		{"Trimmed shortcode", "  cac1  ", "ca-central-1", false},
		{"Trimmed full region", "  us-east-1  ", "us-east-1", false},

		// Invalid inputs
		{"Empty string", "", "", true},
		{"Only spaces", "   ", "", true},
		{"Invalid shortcode", "xxx", "", true},
		{"Invalid region typo", "caa-central-1", "", true},
		{"Invalid region format", "ca-cent-1", "", true},
		{"Partial region", "ca-central", "", true},
		{"Too many parts", "ca-central-1-extra", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateRegionInput(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("ValidateRegionInput(%q) expected error but got none", tt.input)
				}
				if result != "" {
					t.Errorf("ValidateRegionInput(%q) with error should return empty string, got %q", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateRegionInput(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ValidateRegionInput(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      ValidationError
		expected string
	}{
		{
			name: "Region error",
			err: ValidationError{
				Field:   "region",
				Value:   "invalid",
				Message: "must be a valid AWS region",
			},
			expected: "region 'invalid' is invalid: must be a valid AWS region",
		},
		{
			name: "Empty value",
			err: ValidationError{
				Field:   "region",
				Value:   "",
				Message: "cannot be empty",
			},
			expected: "region '' is invalid: cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("ValidationError.Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Benchmark tests to ensure performance
func BenchmarkIsValidAWSRegion(b *testing.B) {
	regions := []string{
		"us-east-1",
		"ca-central-1",
		"invalid-region",
		"us-gov-west-1",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, region := range regions {
			_ = IsValidAWSRegion(region)
		}
	}
}

func BenchmarkValidateRegionInput(b *testing.B) {
	inputs := []string{
		"cac1",
		"us-east-1",
		"invalid",
		"  use1  ",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			_, _ = ValidateRegionInput(input)
		}
	}
}
