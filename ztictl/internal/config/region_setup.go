package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ztictl/pkg/aws"
	"ztictl/pkg/colors"

	"gopkg.in/yaml.v3"
)

// InteractiveRegionSetup prompts the user to configure regions interactively
func InteractiveRegionSetup() error {
	colors.PrintHeader("\n=== Region Configuration Setup ===\n")
	colors.PrintData("Let's configure your AWS regions for multi-region operations.\n\n")

	reader := bufio.NewReader(os.Stdin)

	// Ask if they want to configure regions
	colors.PrintData("Would you like to configure regions now? (yes/no): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "yes" && response != "y" {
		colors.PrintWarning("Region configuration skipped. You can configure it later in ~/.ztictl.yaml\n")
		return fmt.Errorf("region configuration skipped; you can configure it later in ~/.ztictl.yaml")
	}

	// Get enabled regions
	colors.PrintHeader("\n1. Configure Enabled Regions\n")
	colors.PrintData("Enter the regions you want to use (comma-separated).\n")
	colors.PrintData("You can use shortcodes (e.g., cac1, use1) or full names (e.g., ca-central-1, us-east-1).\n")
	colors.PrintData("Common regions: cac1 (Canada), use1 (US East), use2 (US East 2), usw2 (US West 2), euw1 (EU Ireland)\n")
	colors.PrintData("\nEnter regions: ")

	regionsInput, _ := reader.ReadString('\n')
	regionsInput = strings.TrimSpace(regionsInput)

	if regionsInput == "" {
		colors.PrintError("No regions provided. Configuration cancelled.\n")
		return fmt.Errorf("no regions provided")
	}

	// Parse and normalize regions
	enabledRegions := parseAndNormalizeRegions(regionsInput)
	if len(enabledRegions) == 0 {
		colors.PrintError("No valid regions provided. Configuration cancelled.\n")
		return fmt.Errorf("no valid regions")
	}

	colors.PrintSuccess("✓ Enabled regions: %s\n", strings.Join(enabledRegions, ", "))

	// Ask about region groups
	regionGroups := make(map[string][]string)

	colors.PrintHeader("\n2. Configure Region Groups (Optional)\n")
	colors.PrintData("Would you like to create region groups? (e.g., production, development) (yes/no): ")

	groupResponse, _ := reader.ReadString('\n')
	groupResponse = strings.TrimSpace(strings.ToLower(groupResponse))

	if groupResponse == "yes" || groupResponse == "y" {
		for {
			colors.PrintData("\nEnter group name (or 'done' to finish): ")
			groupName, _ := reader.ReadString('\n')
			groupName = strings.TrimSpace(groupName)

			if groupName == "done" || groupName == "" {
				break
			}

			// Allow flexible naming
			colors.PrintData("Enter regions for '%s' group (comma-separated): ", groupName)
			groupRegionsInput, _ := reader.ReadString('\n')
			groupRegionsInput = strings.TrimSpace(groupRegionsInput)

			if groupRegionsInput != "" {
				groupRegions := parseAndNormalizeRegions(groupRegionsInput)
				if len(groupRegions) > 0 {
					regionGroups[groupName] = groupRegions
					colors.PrintSuccess("✓ Created group '%s' with regions: %s\n", groupName, strings.Join(groupRegions, ", "))
				}
			}
		}
	}

	// Add 'all' group with all enabled regions
	regionGroups["all"] = enabledRegions

	// Save configuration
	if err := saveRegionConfig(enabledRegions, regionGroups); err != nil {
		colors.PrintError("Failed to save configuration: %v\n", err)
		return err
	}

	colors.PrintSuccess("\n✓ Region configuration saved to ~/.ztictl.yaml\n")
	colors.PrintData("You can now use:\n")
	colors.PrintData("  - --all-regions to use all enabled regions\n")
	for groupName := range regionGroups {
		if groupName != "all" {
			colors.PrintData("  - --region-group %s to use the %s group\n", groupName, groupName)
		}
	}

	return nil
}

// parseAndNormalizeRegions parses region input and normalizes to shortcodes
func parseAndNormalizeRegions(input string) []string {
	parts := strings.Split(input, ",")
	var normalized []string
	seen := make(map[string]bool)

	for _, part := range parts {
		region := strings.TrimSpace(part)
		if region == "" {
			continue
		}

		// Check if it's already a shortcode
		if _, exists := aws.RegionMapping[region]; exists {
			if !seen[region] {
				normalized = append(normalized, region)
				seen[region] = true
			}
			continue
		}

		// Check if it's a full region name - convert to shortcode
		shortcode := aws.GetRegionCode(region)
		if shortcode != region {
			// Successfully found a shortcode
			if !seen[shortcode] {
				normalized = append(normalized, shortcode)
				seen[shortcode] = true
			}
			continue
		}

		// Try to validate as AWS region format
		if aws.IsValidAWSRegion(region) {
			// It's a valid AWS region format but not in our mapping
			// Store it as-is (might be a new region)
			if !seen[region] {
				normalized = append(normalized, region)
				seen[region] = true
			}
		} else {
			colors.PrintWarning("⚠ Skipping invalid region: %s\n", region)
		}
	}

	return normalized
}

// saveRegionConfig saves the region configuration to the config file
func saveRegionConfig(enabledRegions []string, regionGroups map[string][]string) error {
	configPath := filepath.Join(os.Getenv("HOME"), ".ztictl.yaml")

	// Read existing config if it exists - preserve all existing settings
	var existingConfig map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, &existingConfig); err != nil {
			// If unmarshal fails, log error but continue with empty config
			fmt.Printf("Warning: Could not parse existing config: %v\n", err)
			existingConfig = make(map[string]interface{})
		}
	} else {
		existingConfig = make(map[string]interface{})
	}

	// Update only the regions section
	existingConfig["regions"] = map[string]interface{}{
		"enabled": enabledRegions,
		"groups":  regionGroups,
	}

	// Marshal back to YAML with proper formatting
	data, err := yaml.Marshal(existingConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Reload the configuration
	return Load()
}

// NormalizeRegion converts a region (shortcode or full name) to shortcode
func NormalizeRegion(region string) string {
	// If it's already a shortcode, return it
	if _, exists := aws.RegionMapping[region]; exists {
		return region
	}

	// Try to get shortcode from full name
	shortcode := aws.GetRegionCode(region)
	if shortcode != region {
		return shortcode
	}

	// Return as-is if we can't normalize
	return region
}

// ResolveRegionInput takes a region input (shortcode or full) and returns the full AWS region name
func ResolveRegionInput(regionInput string) string {
	// First check if it's a shortcode
	if fullRegion, exists := aws.RegionMapping[regionInput]; exists {
		return fullRegion
	}

	// Check if it's already a full region name
	if aws.IsValidAWSRegion(regionInput) {
		return regionInput
	}

	// Default: return as-is
	return regionInput
}
