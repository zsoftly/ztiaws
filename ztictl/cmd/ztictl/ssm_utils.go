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

	// Try to convert region shortcode to full AWS region name
	if fullRegion, err := awspkg.GetRegion(regionCode); err == nil {
		return fullRegion
	}

	// If conversion fails, assume it's already a full region name
	return regionCode
}
