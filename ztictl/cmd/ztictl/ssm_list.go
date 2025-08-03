package main

import (
	"context"
	"fmt"
	"os"
	"strings"

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
		region := resolveRegion(regionCode)

		tagFilter, _ := cmd.Flags().GetString("tag")
		statusFilter, _ := cmd.Flags().GetString("status")
		nameFilter, _ := cmd.Flags().GetString("name")

		logging.LogInfo("Listing SSM-enabled instances in region: %s", region)

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		filters := &ssm.ListFilters{
			Tag:    tagFilter,
			Status: statusFilter,
			Name:   nameFilter,
		}

		instances, err := ssmManager.ListInstances(ctx, region, filters)
		if err != nil {
			colors.PrintError("✗ Failed to list instances in region %s\n", region)
			logging.LogError("Failed to list instances: %v", err)
			os.Exit(1)
		}

		if len(instances) == 0 {
			logging.LogInfo("No EC2 instances found in region: %s", region)
			return
		}

		fmt.Printf("\n")
		colors.PrintHeader("All EC2 Instances in %s:\n", region)
		colors.PrintHeader("=====================================\n")
		colors.PrintHeader("%-20s %-19s %-15s %-10s %-15s %s\n", "Name", "Instance ID", "IP Address", "State", "SSM Status", "Platform")
		colors.PrintHeader("%s\n", strings.Repeat("-", 100))

		for _, instance := range instances {
			name := instance.Name
			if name == "" {
				name = "N/A"
			}

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

			colors.Data.Printf("%-20s %-19s %-15s %-10s ", name, instance.InstanceID, instance.PrivateIPAddress, instance.State)
			fmt.Printf("%-15s ", ssmStatus)
			colors.Data.Printf("%s\n", instance.Platform)
		}
		fmt.Printf("\n")
		colors.PrintData("Total: %d instances\n", len(instances))
		fmt.Printf("Note: Only instances with %s SSM status can be connected to via SSM\n", colors.ColorSuccess("'✓ Online'"))
		colors.PrintData("Usage: ztictl ssm connect <instance-id-or-name>\n")
	},
}

func init() {
	ssmListCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmListCmd.Flags().StringP("tag", "t", "", "Filter by tag (format: key=value)")
	ssmListCmd.Flags().StringP("status", "s", "", "Filter by status (running, stopped, etc.)")
	ssmListCmd.Flags().StringP("name", "n", "", "Filter by name pattern")
}
