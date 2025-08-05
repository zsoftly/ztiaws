package main

import (
	"context"

	"ztictl/internal/ssm"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
)

// ssmCleanupCmd represents the ssm cleanup command
var ssmCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up temporary resources created by ztictl",
	Long: `Clean up temporary IAM policies, S3 objects, and other resources
created during file transfer operations. This includes:

- Removing stale IAM policies attached for S3 access
- Cleaning up old registry entries
- Removing stale lock files

Use this command if file transfer operations were interrupted
and temporary resources were not cleaned up automatically.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		region := resolveRegion(regionCode)

		if region == "" {
			logging.LogError("Region is required. Use --region flag or set default region in config")
			return
		}

		logging.LogInfo("Starting cleanup operation in region: %s", region)

		ssmManager := ssm.NewManager(GetLogger())
		ctx := context.Background()

		// Perform routine cleanup
		if err := ssmManager.Cleanup(ctx, region); err != nil {
			logging.LogError("Cleanup failed: %v", err)
			return
		}

		logging.LogSuccess("Cleanup completed successfully")

		// Show colored success message
		colors.PrintSuccess("✓ Cleanup completed successfully\n")
	},
}

// ssmEmergencyCleanupCmd represents the ssm emergency-cleanup command
var ssmEmergencyCleanupCmd = &cobra.Command{
	Use:   "emergency-cleanup",
	Short: "Perform emergency cleanup of all temporary resources",
	Long: `Perform emergency cleanup of all temporary resources created by ztictl.
This is a more aggressive cleanup that:

- Removes all IAM policies created by ztictl for S3 access
- Cleans up all registry entries regardless of age
- Removes all lock files
- Attempts to detach policies from instance roles

Use this command if normal cleanup fails or if you need to
ensure all temporary resources are removed.`,
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		region := resolveRegion(regionCode)

		if region == "" {
			logging.LogError("Region is required. Use --region flag or set default region in config")
			return
		}

		logging.LogInfo("Starting emergency cleanup operation in region: %s", region)

		ssmManager := ssm.NewManager(GetLogger())
		ctx := context.Background()

		// Perform emergency cleanup
		if err := ssmManager.EmergencyCleanup(ctx, region); err != nil {
			logging.LogError("Emergency cleanup failed: %v", err)
			return
		}

		logging.LogSuccess("Emergency cleanup completed successfully")

		// Show colored success message
		colors.PrintSuccess("✓ Emergency cleanup completed successfully\n")
	},
}

func init() {
	ssmCleanupCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmEmergencyCleanupCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
}
