package aws

import (
	"fmt"
	"strings"
)

// IsValidAWSRegion validates if a string is a properly formatted AWS region.
// This is the authoritative region validator for the entire application.
// Valid formats:
//   - Standard: xx-xxxx-n (e.g., us-east-1, ca-central-1)
//   - GovCloud: us-gov-xxxx-n (e.g., us-gov-east-1)
func IsValidAWSRegion(region string) bool {
	if region == "" {
		return false
	}

	// Must contain at least two hyphens
	parts := strings.Split(region, "-")

	// Handle special case for us-gov regions (4 parts)
	if len(parts) == 4 && parts[0] == "us" && parts[1] == "gov" {
		// us-gov-east-1, us-gov-west-1
		parts = []string{"us-gov", parts[2], parts[3]}
	}

	// Now must have exactly 3 parts
	if len(parts) != 3 {
		return false
	}

	// First part: valid region codes only
	validPrefixes := map[string]bool{
		"us":     true, // United States
		"eu":     true, // Europe
		"ap":     true, // Asia Pacific
		"ca":     true, // Canada
		"sa":     true, // South America
		"me":     true, // Middle East
		"af":     true, // Africa
		"cn":     true, // China
		"us-gov": true, // GovCloud
	}

	if !validPrefixes[parts[0]] {
		return false
	}

	// Second part: valid direction/area names
	if parts[0] == "us-gov" {
		// GovCloud only has east and west
		if parts[1] != "east" && parts[1] != "west" {
			return false
		}
	} else {
		validDirections := map[string]bool{
			"east":      true,
			"west":      true,
			"north":     true,
			"south":     true,
			"central":   true,
			"northeast": true,
			"southeast": true,
			"northwest": true,
			"southwest": true,
		}

		if !validDirections[parts[1]] {
			return false
		}
	}

	// Third part: must be a number (1-99)
	if len(parts[2]) < 1 || len(parts[2]) > 2 {
		return false
	}

	// Check if third part is a number
	for _, char := range parts[2] {
		if char < '0' || char > '9' {
			return false
		}
	}

	// AWS regions start from 1, not 0
	if parts[2] == "0" {
		return false
	}

	return true
}

// IsValidRegionShortcode checks if a string is a valid region shortcode
// (e.g., cac1, use1, euw2)
func IsValidRegionShortcode(shortcode string) bool {
	// Check if it exists in our region mapping
	_, exists := RegionMapping[shortcode]
	return exists
}

// ValidateRegionInput validates and normalizes region input.
// It accepts either a shortcode (e.g., cac1) or a full region name (e.g., ca-central-1).
// Returns the full region name if valid, or an error.
func ValidateRegionInput(input string) (string, error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return "", &ValidationError{Field: "region", Value: input, Message: "region cannot be empty"}
	}

	// First check if it's a valid shortcode
	if fullRegion, exists := RegionMapping[input]; exists {
		return fullRegion, nil
	}

	// Then check if it's a valid full region name
	if IsValidAWSRegion(input) {
		return input, nil
	}

	// Not valid as either shortcode or full name
	return "", &ValidationError{
		Field:   "region",
		Value:   input,
		Message: "must be a valid AWS region (e.g., us-east-1, ca-central-1) or shortcode (e.g., use1, cac1)",
	}
}

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s '%s' is invalid: %s", e.Field, e.Value, e.Message)
}
