package aws

import (
	"fmt"
	"strings"

	"ztictl/pkg/errors"
)

// RegionMapping maps region codes to AWS region names
var RegionMapping = map[string]string{
	// Canada
	"cac1": "ca-central-1", // Montreal
	"caw1": "ca-west-1",    // Calgary

	// United States
	"use1": "us-east-1", // N. Virginia
	"use2": "us-east-2", // Ohio
	"usw1": "us-west-1", // N. California
	"usw2": "us-west-2", // Oregon

	// Europe
	"euw1": "eu-west-1",    // Ireland
	"euw2": "eu-west-2",    // London
	"euw3": "eu-west-3",    // Paris
	"euc1": "eu-central-1", // Frankfurt
	"euc2": "eu-central-2", // Zurich
	"eun1": "eu-north-1",   // Stockholm
	"eus1": "eu-south-1",   // Milan
	"eus2": "eu-south-2",   // Spain

	// Asia Pacific
	"aps1":  "ap-south-1",     // Mumbai
	"aps2":  "ap-south-2",     // Hyderabad
	"apse1": "ap-southeast-1", // Singapore
	"apse2": "ap-southeast-2", // Sydney
	"apse3": "ap-southeast-3", // Jakarta
	"apse4": "ap-southeast-4", // Melbourne
	"apne1": "ap-northeast-1", // Tokyo
	"apne2": "ap-northeast-2", // Seoul
	"apne3": "ap-northeast-3", // Osaka

	// South America
	"sae1": "sa-east-1", // São Paulo

	// Africa
	"afs1": "af-south-1", // Cape Town

	// Middle East
	"mes1": "me-south-1",   // Bahrain
	"mec1": "me-central-1", // UAE
}

// RegionDescriptions provides human-readable descriptions for regions
var RegionDescriptions = map[string]string{
	"cac1":  "Canada Central (Montreal)",
	"caw1":  "Canada West (Calgary)",
	"use1":  "US East (N. Virginia)",
	"use2":  "US East (Ohio)",
	"usw1":  "US West (N. California)",
	"usw2":  "US West (Oregon)",
	"euw1":  "EU West (Ireland)",
	"euw2":  "EU West (London)",
	"euw3":  "EU West (Paris)",
	"euc1":  "EU Central (Frankfurt)",
	"euc2":  "EU Central (Zurich)",
	"eun1":  "EU North (Stockholm)",
	"eus1":  "EU South (Milan)",
	"eus2":  "EU South (Spain)",
	"aps1":  "Asia Pacific South (Mumbai)",
	"aps2":  "Asia Pacific South (Hyderabad)",
	"apse1": "Asia Pacific Southeast (Singapore)",
	"apse2": "Asia Pacific Southeast (Sydney)",
	"apse3": "Asia Pacific Southeast (Jakarta)",
	"apse4": "Asia Pacific Southeast (Melbourne)",
	"apne1": "Asia Pacific Northeast (Tokyo)",
	"apne2": "Asia Pacific Northeast (Seoul)",
	"apne3": "Asia Pacific Northeast (Osaka)",
	"sae1":  "South America East (São Paulo)",
	"afs1":  "Africa South (Cape Town)",
	"mes1":  "Middle East South (Bahrain)",
	"mec1":  "Middle East Central (UAE)",
}

// GetRegion converts a region code to an AWS region name
func GetRegion(regionCode string) (string, error) {
	// If it's already a full AWS region name, return it
	if IsValidAWSRegion(regionCode) {
		return regionCode, nil
	}

	// Look up the region code
	region, exists := RegionMapping[strings.ToLower(regionCode)]
	if !exists {
		return "", errors.NewValidationError(fmt.Sprintf("invalid region code: %s", regionCode))
	}

	return region, nil
}

// ValidateRegionCode validates a region code and returns the AWS region name
func ValidateRegionCode(regionCode string) (string, error) {
	return GetRegion(regionCode)
}

// GetRegionDescription returns a human-readable description for a region code
func GetRegionDescription(regionCode string) string {
	desc, exists := RegionDescriptions[strings.ToLower(regionCode)]
	if !exists {
		return "Unknown Region"
	}
	return desc
}

// ListSupportedRegions returns all supported region codes and their descriptions
func ListSupportedRegions() map[string]string {
	return RegionDescriptions
}

// GetRegionCode returns the region code for an AWS region name
func GetRegionCode(awsRegion string) string {
	for code, region := range RegionMapping {
		if region == awsRegion {
			return code
		}
	}
	return awsRegion // Return the original if no mapping found
}
