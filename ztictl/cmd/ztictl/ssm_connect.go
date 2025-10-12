package main

import (
	"context"
	"fmt"
	"os"

	"ztictl/internal/ssm"
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
	if err := ValidateInstanceState(ctx, ssmManager, instanceID, region, InstanceValidationRequirements{
		AllowedStates:    []string{"running"},
		RequireSSMOnline: true,
		Operation:        "connect",
	}); err != nil {
		return err
	}

	logging.LogInfo("Connecting to instance %s in region: %s", instanceID, region)

	if err := ssmManager.StartSession(ctx, instanceID, region); err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	return nil
}

func init() {
	ssmConnectCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
}
