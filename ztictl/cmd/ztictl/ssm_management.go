package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"ztictl/internal/ssm"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
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
		instanceIdentifier := args[0]
		portMapping := args[1]

		if err := performPortForwarding(regionCode, instanceIdentifier, portMapping); err != nil {
			logging.LogError("Port forwarding failed: %v", err)
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
		var instanceIdentifier string
		if len(args) > 0 {
			instanceIdentifier = args[0]
		}

		if err := performStatusCheck(regionCode, instanceIdentifier); err != nil {
			logging.LogError("Status check failed: %v", err)
			os.Exit(1)
		}
	},
}

// performPortForwarding handles port forwarding logic and returns errors instead of calling os.Exit
func performPortForwarding(regionCode, instanceIdentifier, portMapping string) error {
	region := resolveRegion(regionCode)

	// Parse port mapping
	parts := strings.Split(portMapping, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid port format. Use local-port:remote-port (e.g., 8080:80)")
	}

	localPort, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid local port: %s", parts[0])
	}

	remotePort, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid remote port: %s", parts[1])
	}

	logging.LogInfo("Starting port forwarding %d:%d on instance %s in region: %s", localPort, remotePort, instanceIdentifier, region)

	ssmManager := ssm.NewManager(logger)
	ctx := context.Background()

	if err := ssmManager.ForwardPort(ctx, instanceIdentifier, region, localPort, remotePort); err != nil {
		return fmt.Errorf("port forwarding failed: %w", err)
	}

	return nil
}

// performStatusCheck handles status checking logic and returns errors instead of calling os.Exit
func performStatusCheck(regionCode, instanceIdentifier string) error {
	region := resolveRegion(regionCode)

	ssmManager := ssm.NewManager(logger)
	ctx := context.Background()

	if instanceIdentifier != "" {
		// Show status for specific instance
		status, err := ssmManager.GetInstanceStatus(ctx, instanceIdentifier, region)
		if err != nil {
			return fmt.Errorf("failed to get instance status: %w", err)
		}

		fmt.Printf("\n")
		colors.PrintHeader("SSM Agent Status for %s:\n", instanceIdentifier)
		colors.PrintHeader("==========================\n")
		colors.PrintData("Instance ID: %s\n", status.InstanceID)
		colors.PrintData("SSM Status: %s\n", status.SSMStatus)
		colors.PrintData("Last Ping: %s\n", status.LastPingDateTime)
		colors.PrintData("Agent Version: %s\n", status.SSMAgentVersion)
		colors.PrintData("Platform: %s\n", status.Platform)
		colors.PrintData("State: %s\n", status.State)
	} else {
		// Show status for all instances
		statuses, err := ssmManager.ListInstanceStatuses(ctx, region)
		if err != nil {
			return fmt.Errorf("failed to list instance statuses: %w", err)
		}

		if len(statuses) == 0 {
			logging.LogInfo("No SSM-managed instances found in region: %s", region)
			return nil
		}

		fmt.Printf("\n")
		colors.PrintHeader("SSM Agent Status in %s:\n", region)
		colors.PrintHeader("==============================\n")
		colors.PrintHeader("%-20s %-15s %-20s %s\n", "Instance ID", "Status", "Last Ping", "Agent Version")
		colors.PrintHeader("%s\n", strings.Repeat("-", 75))

		for _, status := range statuses {
			colors.PrintData("%-20s %-15s %-20s %s\n",
				status.InstanceID, status.SSMStatus, status.LastPingDateTime, status.SSMAgentVersion)
		}
	}

	return nil
}

func init() {
	ssmForwardCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmStatusCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
}
