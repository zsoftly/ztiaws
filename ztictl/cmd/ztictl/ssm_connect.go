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

// ssmConnectCmd represents the ssm connect command
var ssmConnectCmd = &cobra.Command{
	Use:   "connect [instance-identifier]",
	Short: "Connect to an instance via SSM Session Manager",
	Long: `Connect to an EC2 instance using SSM Session Manager.
If no instance identifier is provided, an interactive fuzzy finder will be launched.
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")

		var instanceIdentifier string
		if len(args) > 0 {
			instanceIdentifier = args[0]
		}

		if err := performConnection(regionCode, instanceIdentifier); err != nil {
			logging.LogError("Connection failed: %v", err)
			os.Exit(1)
		}
	},
}

// performConnection handles SSM connection logic and returns errors instead of calling os.Exit
func performConnection(regionCode, instanceIdentifier string) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()
	ssmManager := ssm.NewManager(logger)

	// Use the shared instance selection logic
	instanceID, err := ssmManager.GetInstanceService().SelectInstanceWithFallback(
		ctx,
		instanceIdentifier,
		region,
		nil, // No filters
	)
	if err != nil {
		return fmt.Errorf("instance selection failed: %w", err)
	}

	// Validate instance state before attempting connection
	if err := validateInstanceForSSM(ctx, ssmManager, instanceID, region, "connect"); err != nil {
		return err
	}

	logging.LogInfo("Connecting to instance %s in region: %s", instanceID, region)

	if err := ssmManager.StartSession(ctx, instanceID, region); err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	return nil
}

// validateInstanceForSSM checks if an instance is in a valid state for SSM operations
func validateInstanceForSSM(ctx context.Context, ssmManager *ssm.Manager, instanceID, region, operation string) error {
	// Get instance details
	instances, err := ssmManager.GetInstanceService().ListInstances(ctx, region, &awsservice.ListFilters{})
	if err != nil {
		return fmt.Errorf("failed to fetch instance details: %w", err)
	}

	// Find the specific instance
	var targetInstance *interactive.Instance
	for i := range instances {
		if instances[i].InstanceID == instanceID {
			targetInstance = &instances[i]
			break
		}
	}

	if targetInstance == nil {
		return fmt.Errorf("instance %s not found in region %s", instanceID, region)
	}

	// Check instance state
	if targetInstance.State != "running" {
		colors.PrintError("\n✗ Cannot %s to instance - Instance is not running\n", operation)
		fmt.Printf("\n")
		colors.PrintWarning("Instance Details:\n")
		fmt.Printf("  Instance ID: %s\n", colors.ColorData("%s", targetInstance.InstanceID))
		fmt.Printf("  Name:        %s\n", colors.ColorData("%s", targetInstance.Name))
		fmt.Printf("  State:       %s\n", getInstanceStateColor(targetInstance.State))
		fmt.Printf("\n")

		switch targetInstance.State {
		case "stopped":
			colors.PrintData("💡 Tip: Start the instance first:\n")
			fmt.Printf("   aws ec2 start-instances --instance-ids %s --region %s\n", targetInstance.InstanceID, region)
		case "stopping":
			colors.PrintWarning("⏳ Instance is currently stopping. Wait for it to stop, then start it.\n")
		case "pending":
			colors.PrintWarning("⏳ Instance is starting. Please wait a moment and try again.\n")
		case "terminated":
			colors.PrintError("✗ Instance has been terminated. Cannot perform operations on terminated instances.\n")
		default:
			colors.PrintWarning("⚠ Instance is in '%s' state. Only 'running' instances can accept SSM connections.\n", targetInstance.State)
		}
		fmt.Printf("\n")
		return fmt.Errorf("instance is in '%s' state, expected 'running'", targetInstance.State)
	}

	// Check SSM agent status
	if targetInstance.SSMStatus != "Online" {
		colors.PrintWarning("\n⚠ Warning: SSM Agent is not online\n")
		fmt.Printf("\n")
		fmt.Printf("  Instance ID:  %s\n", colors.ColorData("%s", targetInstance.InstanceID))
		fmt.Printf("  Name:         %s\n", colors.ColorData("%s", targetInstance.Name))
		fmt.Printf("  State:        %s\n", colors.ColorSuccess("%s", "● running"))
		fmt.Printf("  SSM Status:   %s\n", getSSMStatusColor(targetInstance.SSMStatus))
		fmt.Printf("\n")

		colors.PrintData("Possible reasons:\n")
		fmt.Printf("  1. SSM Agent not installed or not running\n")
		fmt.Printf("  2. Instance doesn't have required IAM role (AmazonSSMManagedInstanceCore)\n")
		fmt.Printf("  3. Network connectivity issues to SSM endpoints\n")
		fmt.Printf("  4. Instance recently started (agent may still be initializing)\n")
		fmt.Printf("\n")

		if targetInstance.SSMStatus == "ConnectionLost" {
			colors.PrintWarning("The agent was previously online but connection was lost.\n")
			fmt.Printf("This may be temporary. You can try to proceed, but connection may fail.\n\n")
		} else {
			colors.PrintError("Cannot %s - SSM Agent must be 'Online' to establish connection.\n\n", operation)
			return fmt.Errorf("SSM agent is '%s', expected 'Online'", targetInstance.SSMStatus)
		}
	}

	return nil
}

// getInstanceStateColor returns a colored state string
func getInstanceStateColor(state string) string {
	switch state {
	case "running":
		return colors.ColorSuccess("● %s", "running")
	case "stopped":
		return colors.ColorError("○ %s", "stopped")
	case "stopping":
		return colors.ColorWarning("◑ %s", "stopping")
	case "pending":
		return colors.ColorWarning("◐ %s", "pending")
	case "terminated":
		return colors.ColorError("✗ %s", "terminated")
	default:
		return colors.ColorData("%s", state)
	}
}

// getSSMStatusColor returns a colored SSM status string
func getSSMStatusColor(status string) string {
	switch status {
	case "Online":
		return colors.ColorSuccess("✓ %s", "Online")
	case "ConnectionLost":
		return colors.ColorWarning("⚠ %s", "Connection Lost")
	case "No Agent":
		return colors.ColorError("✗ %s", "No Agent")
	default:
		if status == "" {
			return colors.ColorError("✗ %s", "No Agent")
		}
		return colors.ColorWarning("? %s", status)
	}
}

func init() {
	ssmConnectCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
}
