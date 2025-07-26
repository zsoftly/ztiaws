package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"ztictl/internal/ssm"
)

// ssmForwardCmd represents the ssm forward command
var ssmForwardCmd = &cobra.Command{
	Use:   "forward <instance-identifier> <local-port>:<remote-port>",
	Short: "Forward ports through SSM tunnel",
	Long: `Forward local ports to remote ports on an EC2 instance through SSM.
Format: local-port:remote-port (e.g., 8080:80)
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		region := resolveRegion(regionCode)

		instanceIdentifier := args[0]
		portMapping := args[1]

		// Parse port mapping
		parts := strings.Split(portMapping, ":")
		if len(parts) != 2 {
			logger.Error("Invalid port format. Use local-port:remote-port (e.g., 8080:80)")
			os.Exit(1)
		}

		localPort, err := strconv.Atoi(parts[0])
		if err != nil {
			logger.Error("Invalid local port", "port", parts[0])
			os.Exit(1)
		}

		remotePort, err := strconv.Atoi(parts[1])
		if err != nil {
			logger.Error("Invalid remote port", "port", parts[1])
			os.Exit(1)
		}

		logger.Info("Starting port forwarding", "instance", instanceIdentifier, "local", localPort, "remote", remotePort)

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		if err := ssmManager.ForwardPort(ctx, instanceIdentifier, region, localPort, remotePort); err != nil {
			logger.Error("Port forwarding failed", "error", err)
			os.Exit(1)
		}
	},
}

// ssmStatusCmd represents the ssm status command
var ssmStatusCmd = &cobra.Command{
	Use:   "status [instance-identifier]",
	Short: "Check SSM agent status",
	Long: `Check the status of the SSM agent on instances.
If no instance is specified, shows status for all instances in the region.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		region := resolveRegion(regionCode)

		var instanceIdentifier string
		if len(args) > 0 {
			instanceIdentifier = args[0]
		}

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		if instanceIdentifier != "" {
			// Show status for specific instance
			status, err := ssmManager.GetInstanceStatus(ctx, instanceIdentifier, region)
			if err != nil {
				logger.Error("Failed to get instance status", "error", err)
				os.Exit(1)
			}

			fmt.Printf("\nSSM Agent Status for %s:\n", instanceIdentifier)
			fmt.Println("==========================")
			fmt.Printf("Instance ID: %s\n", status.InstanceID)
			fmt.Printf("SSM Status: %s\n", status.SSMStatus)
			fmt.Printf("Last Ping: %s\n", status.LastPingDateTime)
			fmt.Printf("Agent Version: %s\n", status.SSMAgentVersion)
			fmt.Printf("Platform: %s\n", status.Platform)
			fmt.Printf("State: %s\n", status.State)
		} else {
			// Show status for all instances
			statuses, err := ssmManager.ListInstanceStatuses(ctx, region)
			if err != nil {
				logger.Error("Failed to list instance statuses", "error", err)
				os.Exit(1)
			}

			if len(statuses) == 0 {
				logger.Info("No SSM-managed instances found", "region", region)
				return
			}

			fmt.Printf("\nSSM Agent Status in %s:\n", region)
			fmt.Println("==============================")
			fmt.Printf("%-20s %-15s %-20s %s\n", "Instance ID", "Status", "Last Ping", "Agent Version")
			fmt.Println(strings.Repeat("-", 75))

			for _, status := range statuses {
				fmt.Printf("%-20s %-15s %-20s %s\n",
					status.InstanceID, status.SSMStatus, status.LastPingDateTime, status.SSMAgentVersion)
			}
		}
	},
}

func init() {
	ssmForwardCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmStatusCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
}
