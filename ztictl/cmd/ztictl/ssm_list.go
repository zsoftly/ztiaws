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
		tableFormat, _ := cmd.Flags().GetBool("table")

		filters := &ssm.ListFilters{
			Tag:    tagFilter,
			Status: statusFilter,
			Name:   nameFilter,
		}

		if err := performInstanceListing(regionCode, filters, tableFormat); err != nil {
			logging.LogError("Instance listing failed: %v", err)
			os.Exit(1)
		}
	},
}

// performInstanceListing handles instance listing logic and returns errors instead of calling os.Exit
func performInstanceListing(regionCode string, filters *ssm.ListFilters, tableFormat bool) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()
	ssmManager := ssm.NewManager(logger)

	colors.PrintData("🔍 Fetching instances from region %s...\n", region)

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
		colors.PrintWarning("⚠ No EC2 instances found in region: %s\n", region)
		return nil
	}

	colors.PrintSuccess("✓ Found %d instance(s) in region %s\n", len(instances), region)

	// Use table format if requested, otherwise use interactive fuzzy finder
	if tableFormat {
		printInstanceTable(instances, region)
		return nil
	}

	logging.LogInfo("Launching interactive instance browser...")

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

// printInstanceTable prints instances in a traditional table format
func printInstanceTable(instances []interactive.Instance, region string) {
	formatter := NewTableFormatter(2) // 2 spaces between columns

	// Prepare column data
	names := make([]string, len(instances))
	instanceIDs := make([]string, len(instances))
	privateIPs := make([]string, len(instances))
	publicIPs := make([]string, len(instances))
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

		// Private IP Address
		privateIPs[i] = instance.PrivateIPAddress

		// Public IP Address
		publicIP := instance.PublicIPAddress
		if publicIP == "" {
			publicIP = "N/A"
		}
		publicIPs[i] = publicIP

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
	formatter.AddColumn("Private IP", privateIPs, 10)
	formatter.AddColumn("Public IP", publicIPs, 10)
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
}

func init() {
	ssmListCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmListCmd.Flags().StringP("tag", "t", "", "Filter by tag (format: key=value)")
	ssmListCmd.Flags().StringP("status", "s", "", "Filter by status (running, stopped, etc.)")
	ssmListCmd.Flags().StringP("name", "n", "", "Filter by name pattern")
	ssmListCmd.Flags().Bool("table", false, "Display instances in table format instead of interactive fuzzy finder")
}
