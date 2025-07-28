package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"ztictl/internal/ssm"
)

// ssmListCmd represents the ssm list command
var ssmListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SSM-enabled instances",
	Long: `List EC2 instances that are available through AWS Systems Manager.
Optionally filter by tags, status, or name patterns.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		region := resolveRegion(regionCode)

		tagFilter, _ := cmd.Flags().GetString("tag")
		statusFilter, _ := cmd.Flags().GetString("status")
		nameFilter, _ := cmd.Flags().GetString("name")

		logger.Info("Listing SSM-enabled instances", "region", region)

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		filters := &ssm.ListFilters{
			Tag:    tagFilter,
			Status: statusFilter,
			Name:   nameFilter,
		}

		instances, err := ssmManager.ListInstances(ctx, region, filters)
		if err != nil {
			logger.Error("Failed to list instances", "error", err)
			os.Exit(1)
		}

		if len(instances) == 0 {
			logger.Info("No SSM-enabled instances found", "region", region)
			return
		}

		fmt.Printf("\nSSM-Enabled Instances in %s:\n", region)
		fmt.Println("=====================================")
		fmt.Printf("%-20s %-19s %-15s %-10s %s\n", "Name", "Instance ID", "IP Address", "State", "Platform")
		fmt.Println(strings.Repeat("-", 85))

		for _, instance := range instances {
			name := instance.Name
			if name == "" {
				name = "N/A"
			}
			fmt.Printf("%-20s %-19s %-15s %-10s %s\n",
				name, instance.InstanceID, instance.PrivateIPAddress, instance.State, instance.Platform)
		}
		fmt.Printf("\nTotal: %d instances\n", len(instances))
		fmt.Printf("Usage: ztictl ssm connect <instance-id-or-name>\n")
	},
}

func init() {
	ssmListCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmListCmd.Flags().StringP("tag", "t", "", "Filter by tag (format: key=value)")
	ssmListCmd.Flags().StringP("status", "s", "", "Filter by status (running, stopped, etc.)")
	ssmListCmd.Flags().StringP("name", "n", "", "Filter by name pattern")
}
