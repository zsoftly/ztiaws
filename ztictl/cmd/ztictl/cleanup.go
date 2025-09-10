package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"ztictl/internal/ssm"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup [region]",
	Short: "Clean up temporary resources created by ztictl",
	Long: `Clean up temporary IAM policies, S3 objects, and other resources
created during file transfer operations. This includes:

- Removing stale IAM policies attached for S3 access
- Cleaning up old registry entries
- Removing stale lock files

Use this command if file transfer operations were interrupted
and temporary resources were not cleaned up automatically.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		region := args[0]

		if region == "" {
			return fmt.Errorf("region is required")
		}

		ssmManager := ssm.NewManager(logger)

		ctx := context.Background()

		logger.Info("Starting cleanup operation", "region", region)

		// Perform routine cleanup
		if err := ssmManager.Cleanup(ctx, region); err != nil {
			return fmt.Errorf("cleanup failed: %w", err)
		}

		logger.Info("Cleanup completed successfully")
		return nil
	},
}

var emergencyCleanupCmd = &cobra.Command{
	Use:   "emergency-cleanup [region]",
	Short: "Perform emergency cleanup of all temporary resources",
	Long: `Perform emergency cleanup of all temporary resources created by ztictl.
This is a more aggressive cleanup that:

- Removes all IAM policies created by ztictl for S3 access
- Cleans up all registry entries regardless of age
- Removes all lock files
- Attempts to detach policies from instance roles

Use this command if normal cleanup fails or if you need to
ensure all temporary resources are removed.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		region := args[0]

		if region == "" {
			return fmt.Errorf("region is required")
		}

		ssmManager := ssm.NewManager(logger)

		ctx := context.Background()

		logger.Info("Starting emergency cleanup operation", "region", region)

		// Perform emergency cleanup
		if err := ssmManager.EmergencyCleanup(ctx, region); err != nil {
			return fmt.Errorf("emergency cleanup failed: %w", err)
		}

		logger.Info("Emergency cleanup completed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
	rootCmd.AddCommand(emergencyCleanupCmd)
}
