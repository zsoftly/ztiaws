package main

import (
	"context"
	"fmt"
	"os"

	"ztictl/internal/interactive"
	"ztictl/internal/ssm"
	awsservice "ztictl/pkg/aws"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
)

// ssmListCmd represents the ssm list command
var ssmListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all EC2 instances with their SSM status",
	Long: `List all EC2 instances in a region with their SSM agent status.
Shows all instances regardless of their state or SSM connectivity.
Optionally filter by tags, status, or name patterns.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		tagFilter, _ := cmd.Flags().GetString("tag")
		statusFilter, _ := cmd.Flags().GetString("status")
		nameFilter, _ := cmd.Flags().GetString("name")

		filters := &ssm.ListFilters{
			Tag:    tagFilter,
			Status: statusFilter,
			Name:   nameFilter,
		}

		if err := performInstanceListing(regionCode, filters); err != nil {
			logging.LogError("Instance listing failed: %v", err)
			os.Exit(1)
		}
	},
}

// performInstanceListing handles instance listing logic and returns errors instead of calling os.Exit
func performInstanceListing(regionCode string, filters *ssm.ListFilters) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()
	ssmManager := ssm.NewManager(logger)

	logging.LogInfo("Listing SSM-enabled instances in region: %s", region)

	// Convert SSM filters to AWS filters
	awsFilters := &awsservice.ListFilters{
		Tag:    filters.Tag,
		Tags:   filters.Tags,
		Status: filters.Status,
		Name:   filters.Name,
	}

	instances, err := ssmManager.GetInstanceService().ListInstances(ctx, region, awsFilters)
	if err != nil {
		colors.PrintError("✗ Failed to list instances in region %s\n", region)
		return fmt.Errorf("failed to list instances: %w", err)
	}

	if len(instances) == 0 {
		logging.LogInfo("No EC2 instances found in region: %s", region)
		return nil
	}

	// Use shared fuzzy finder (user can select or just browse)
	_, err = interactive.SelectInstance(instances, "Browse EC2 instances")
	if err != nil {
		// User cancelled - that's OK for list command
		if err.Error() == "instance selection cancelled" {
			return nil
		}
		return fmt.Errorf("instance selection failed: %w", err)
	}

	return nil
}

func init() {
	ssmListCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmListCmd.Flags().StringP("tag", "t", "", "Filter by tag (format: key=value)")
	ssmListCmd.Flags().StringP("status", "s", "", "Filter by status (running, stopped, etc.)")
	ssmListCmd.Flags().StringP("name", "n", "", "Filter by name pattern")
}
