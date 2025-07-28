package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"ztictl/internal/ssm"
)

// ssmConnectCmd represents the ssm connect command
var ssmConnectCmd = &cobra.Command{
	Use:   "connect <instance-identifier>",
	Short: "Connect to an instance via SSM Session Manager",
	Long: `Connect to an EC2 instance using SSM Session Manager.
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		region := resolveRegion(regionCode)

		instanceIdentifier := args[0]

		logger.Info("Connecting to instance", "identifier", instanceIdentifier, "region", region)

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		if err := ssmManager.StartSession(ctx, instanceIdentifier, region); err != nil {
			logger.Error("Failed to start session", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	ssmConnectCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
}
