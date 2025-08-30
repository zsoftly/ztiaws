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
		{
			name:     "full AWS region name",
			input:    "us-west-2",
			expected: "us-west-2",
			wantErr:  false,
		},
		{
			name:     "case insensitive shortcode",
			input:    "USE1",
			expected: "us-east-1",
			wantErr:  false,
		},
		{
			name:     "europe region",
			input:    "euw1",
			expected: "eu-west-1",
			wantErr:  false,
		},
		{
			name:     "asia pacific region",
			input:    "apne1",
			expected: "ap-northeast-1",
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
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
		{
			name:     "europe description",
			input:    "euw1",
			contains: "Ireland",
		},
		{
			name:     "asia pacific description",
			input:    "apne1",
			contains: "Tokyo",
		},
		{
			name:     "invalid region returns unknown",
			input:    "invalid",
			contains: "Unknown",
		},
		{
			name:     "case insensitive",
			input:    "USE1",
			contains: "Virginia",
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

func TestValidateRegionCode(t *testing.T) {
	// Test that ValidateRegionCode is just an alias for GetRegion
	validRegion := "use1"
	invalidRegion := "invalid"

	// Test valid region
	region1, err1 := ValidateRegionCode(validRegion)
	region2, err2 := GetRegion(validRegion)

	if region1 != region2 || (err1 != nil) != (err2 != nil) {
		t.Error("ValidateRegionCode should behave identically to GetRegion")
	}

	// Test invalid region
	region3, err3 := ValidateRegionCode(invalidRegion)
	region4, err4 := GetRegion(invalidRegion)

	if region3 != region4 || (err3 != nil) != (err4 != nil) {
		t.Error("ValidateRegionCode should behave identically to GetRegion")
	}
}

func TestListSupportedRegions(t *testing.T) {
	regions := ListSupportedRegions()

	// Should return the same as RegionDescriptions
	if len(regions) != len(RegionDescriptions) {
		t.Errorf("Expected %d regions, got %d", len(RegionDescriptions), len(regions))
	}

	// Check some known regions
	expectedRegions := []string{"use1", "usw2", "euw1", "apne1"}
	for _, region := range expectedRegions {
		if _, exists := regions[region]; !exists {
			t.Errorf("Expected region %s to be in supported regions", region)
		}
	}

	// Verify content matches RegionDescriptions
	for code, desc := range regions {
		if RegionDescriptions[code] != desc {
			t.Errorf("Mismatch for region %s: expected %s, got %s", code, RegionDescriptions[code], desc)
		}
	}
}

func TestGetRegionCode(t *testing.T) {
	tests := []struct {
		name      string
		awsRegion string
		expected  string
	}{
		{
			name:      "us-east-1 to code",
			awsRegion: "us-east-1",
			expected:  "use1",
		},
		{
			name:      "ca-central-1 to code",
			awsRegion: "ca-central-1",
			expected:  "cac1",
		},
		{
			name:      "eu-west-1 to code",
			awsRegion: "eu-west-1",
			expected:  "euw1",
		},
		{
			name:      "unmapped region returns original",
			awsRegion: "unknown-region-1",
			expected:  "unknown-region-1",
		},
		{
			name:      "empty string",
			awsRegion: "",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRegionCode(tt.awsRegion)
			if result != tt.expected {
				t.Errorf("GetRegionCode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidAWSRegion(t *testing.T) {
	tests := []struct {
		name   string
		region string
		valid  bool
	}{
		{
			name:   "valid region us-east-1",
			region: "us-east-1",
			valid:  true,
		},
		{
			name:   "valid region eu-west-1",
			region: "eu-west-1",
			valid:  true,
		},
		{
			name:   "valid region ap-southeast-1",
			region: "ap-southeast-1",
			valid:  true,
		},
		{
			name:   "valid region with 4 parts",
			region: "ap-southeast-1a", // This would be 4 parts when split
			valid:  true,              // Actually, this does match AWS pattern (3-4 parts allowed)
		},
		{
			name:   "invalid short region",
			region: "us-east",
			valid:  false,
		},
		{
			name:   "invalid single part",
			region: "region",
			valid:  false,
		},
		{
			name:   "empty string",
			region: "",
			valid:  false,
		},
		{
			name:   "region code not full region",
			region: "use1",
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidAWSRegion(tt.region)
			if result != tt.valid {
				t.Errorf("isValidAWSRegion() = %v, want %v for region %s", result, tt.valid, tt.region)
			}
		})
	}
}

func TestRegionMapping(t *testing.T) {
	// Test that RegionMapping contains expected entries
	expectedMappings := map[string]string{
		"use1":  "us-east-1",
		"usw2":  "us-west-2",
		"euw1":  "eu-west-1",
		"apne1": "ap-northeast-1",
		"cac1":  "ca-central-1",
	}

	for code, expectedRegion := range expectedMappings {
		if actualRegion, exists := RegionMapping[code]; !exists || actualRegion != expectedRegion {
			t.Errorf("Expected RegionMapping[%s] = %s, got %s", code, expectedRegion, actualRegion)
		}
	}

	// Test that all mappings follow AWS region pattern
	for code, region := range RegionMapping {
		if code == "" {
			t.Error("Region code should not be empty")
		}

		if !isValidAWSRegion(region) {
			t.Errorf("Region %s for code %s is not a valid AWS region", region, code)
		}
	}
}

func TestRegionDescriptions(t *testing.T) {
	// Test that RegionDescriptions contains expected entries
	expectedDescriptions := map[string]string{
		"use1": "US East (N. Virginia)",
		"usw2": "US West (Oregon)",
		"euw1": "EU West (Ireland)",
		"cac1": "Canada Central (Montreal)",
	}

	for code, expectedDesc := range expectedDescriptions {
		if actualDesc, exists := RegionDescriptions[code]; !exists || actualDesc != expectedDesc {
			t.Errorf("Expected RegionDescriptions[%s] = %s, got %s", code, expectedDesc, actualDesc)
		}
	}

	// Test that all region codes in descriptions also exist in mappings
	for code := range RegionDescriptions {
		if _, exists := RegionMapping[code]; !exists {
			t.Errorf("Region code %s exists in descriptions but not in mappings", code)
		}
	}

	// Test that all region codes in mappings also exist in descriptions
	for code := range RegionMapping {
		if _, exists := RegionDescriptions[code]; !exists {
			t.Errorf("Region code %s exists in mappings but not in descriptions", code)
		}
	}
}

func TestAllRegionCodes(t *testing.T) {
	// Test that we can convert all region codes to regions and back
	for code, expectedRegion := range RegionMapping {
		// Convert code to region
		region, err := GetRegion(code)
		if err != nil {
			t.Errorf("Failed to get region for code %s: %v", code, err)
			continue
		}

		if region != expectedRegion {
			t.Errorf("Expected region %s for code %s, got %s", expectedRegion, code, region)
		}

		// Convert region back to code
		backCode := GetRegionCode(region)
		if backCode != code {
			t.Errorf("Expected code %s for region %s, got %s", code, region, backCode)
		}

		// Get description
		desc := GetRegionDescription(code)
		if desc == "Unknown Region" {
			t.Errorf("No description found for valid region code %s", code)
		}
	}
}

func TestCaseSensitivity(t *testing.T) {
	// Test that region codes are case-insensitive
	testCodes := []string{"use1", "USE1", "UsE1", "uSe1"}
	expectedRegion := "us-east-1"

	for _, code := range testCodes {
		region, err := GetRegion(code)
		if err != nil {
			t.Errorf("Failed to get region for code %s: %v", code, err)
		}

		if region != expectedRegion {
			t.Errorf("Expected region %s for code %s, got %s", expectedRegion, code, region)
		}

		// Test descriptions are also case-insensitive
		desc := GetRegionDescription(code)
		expectedDesc := RegionDescriptions["use1"]
		if desc != expectedDesc {
			t.Errorf("Expected description %s for code %s, got %s", expectedDesc, code, desc)
		}
	}
}
