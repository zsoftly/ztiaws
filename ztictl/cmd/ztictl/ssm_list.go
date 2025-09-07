package main

import (
	"context"
	"fmt"
	"os"

	"ztictl/internal/ssm"
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

	logging.LogInfo("Listing SSM-enabled instances in region: %s", region)

	ssmManager := ssm.NewManager(logger)
	ctx := context.Background()

	instances, err := ssmManager.ListInstances(ctx, region, filters)
	if err != nil {
		colors.PrintError("✗ Failed to list instances in region %s\n", region)
		return fmt.Errorf("failed to list instances: %w", err)
	}

	if len(instances) == 0 {
		logging.LogInfo("No EC2 instances found in region: %s", region)
		return nil
	}

	// Prepare data for dynamic table formatting
	formatter := NewTableFormatter(2) // 2 spaces between columns

	// Prepare column data
	names := make([]string, len(instances))
	instanceIDs := make([]string, len(instances))
	ipAddresses := make([]string, len(instances))
	states := make([]string, len(instances))
	ssmStatuses := make([]string, len(instances))
	platforms := make([]string, len(instances))

	for i, instance := range instances {
		// Name
		name := instance.Name
		if name == "" {
			name = "N/A"
		}
		names[i] = name

		// Instance ID
		instanceIDs[i] = instance.InstanceID

		// IP Address
		ipAddresses[i] = instance.PrivateIPAddress

		// State
		states[i] = instance.State

		// Format SSM status with color indicators
		var ssmStatus string
		switch instance.SSMStatus {
		case "Online":
			ssmStatus = colors.ColorSuccess("✓ Online")
		case "ConnectionLost":
			ssmStatus = colors.ColorWarning("⚠ Lost")
		case "No Agent":
			ssmStatus = colors.ColorError("✗ No Agent")
		default:
			if instance.SSMStatus == "" {
				ssmStatus = colors.ColorError("✗ No Agent")
			} else {
				ssmStatus = colors.ColorWarning("? %s", instance.SSMStatus)
			}
		}
		ssmStatuses[i] = ssmStatus

		// Platform
		platforms[i] = instance.Platform
	}

	// Add columns to formatter
	formatter.AddColumn("Name", names, 8)
	formatter.AddColumn("Instance ID", instanceIDs, 12)
	formatter.AddColumn("IP Address", ipAddresses, 10)
	formatter.AddColumn("State", states, 8)
	formatter.AddColumn("SSM Status", ssmStatuses, 10)
	formatter.AddColumn("Platform", platforms, 8)

	fmt.Printf("\n")
	colors.PrintHeader("All EC2 Instances in %s:\n", region)
	colors.PrintHeader("=====================================\n")

	// Print formatted header
	headerStr := formatter.FormatHeader()
	colors.PrintHeader("%s\n", headerStr)

	// Print formatted rows
	for i := 0; i < formatter.GetRowCount(); i++ {
		rowStr := formatter.FormatRow(i)
		fmt.Printf("%s\n", rowStr)
	}

	fmt.Printf("\n")
	colors.PrintData("Total: %d instances\n", len(instances))
	fmt.Printf("Note: Only instances with %s SSM status can be connected to via SSM\n", colors.ColorSuccess("'✓ Online'"))
	colors.PrintData("Usage: ztictl ssm connect <instance-id-or-name>\n")

	return nil
}

func init() {
	ssmListCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmListCmd.Flags().StringP("tag", "t", "", "Filter by tag (format: key=value)")
	ssmListCmd.Flags().StringP("status", "s", "", "Filter by status (running, stopped, etc.)")
	ssmListCmd.Flags().StringP("name", "n", "", "Filter by name pattern")
}
