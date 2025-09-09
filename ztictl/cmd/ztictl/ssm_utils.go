package main

import (
	"ztictl/internal/config"
	awspkg "ztictl/pkg/aws"
)

// Helper function to resolve region code to full region name
// This is equivalent to the region resolution in 01_regions.sh
func resolveRegion(regionCode string) string {
	if regionCode == "" {
		return config.Get().DefaultRegion
	}

	// Handle both shortcodes and full region names
	// First normalize to shortcode if it's a full name
	normalized := config.NormalizeRegion(regionCode)

	// Then convert shortcode to full AWS region name
	if fullRegion, err := awspkg.GetRegion(normalized); err == nil {
		return fullRegion
	}

	// If it's not a known shortcode, check if it's already a valid full region
	if isValidAWSRegion(regionCode) {
		return regionCode
	}

	// If conversion fails, assume it's already a full region name
	return regionCode
}
