package main

import (
	"context"
	"fmt"
	"os"

	"ztictl/internal/ssm"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
)

// ssmTransferCmd represents the ssm transfer command
var ssmTransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "File transfer operations via SSM",
	Long: `Transfer files to/from EC2 instances using SSM.
For files < 1MB: Direct transfer via SSM (faster)
For files ≥ 1MB: Transfer via S3 intermediary (reliable for large files)`,
}

// ssmUploadCmd represents the upload subcommand
var ssmUploadCmd = &cobra.Command{
	Use:   "upload <instance-identifier> <local-file> <remote-path>",
	Short: "Upload a file to an instance",
	Long: `Upload a local file to an EC2 instance via SSM.
Files are transferred directly for small files or via S3 for large files.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		instanceIdentifier := args[0]
		localFile := args[1]
		remotePath := args[2]

		if err := performFileUpload(regionCode, instanceIdentifier, localFile, remotePath); err != nil {
			logging.LogError("File upload failed: %v", err)
			os.Exit(1)
		}
	},
}

// ssmDownloadCmd represents the download subcommand
var ssmDownloadCmd = &cobra.Command{
	Use:   "download <instance-identifier> <remote-file> <local-path>",
	Short: "Download a file from an instance",
	Long: `Download a file from an EC2 instance via SSM.
Files are transferred directly for small files or via S3 for large files.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		instanceIdentifier := args[0]
		remoteFile := args[1]
		localPath := args[2]

		if err := performFileDownload(regionCode, instanceIdentifier, remoteFile, localPath); err != nil {
			logging.LogError("File download failed: %v", err)
			os.Exit(1)
		}
	},
}

// performFileUpload handles file upload logic and returns errors instead of calling os.Exit
func performFileUpload(regionCode, instanceIdentifier, localFile, remotePath string) error {
	region := resolveRegion(regionCode)

	logging.LogInfo("Uploading file %s to instance %s at path: %s", localFile, instanceIdentifier, remotePath)

	ssmManager := ssm.NewManager(logger)
	ctx := context.Background()

	if err := ssmManager.UploadFile(ctx, instanceIdentifier, region, localFile, remotePath); err != nil {
		colors.PrintError("✗ File upload failed: %s -> %s\n", localFile, remotePath)
		return fmt.Errorf("file upload failed: %w", err)
	}

	logging.LogSuccess("File upload completed successfully")

	// Show colored success message
	colors.PrintSuccess("✓ File upload completed successfully: %s -> %s\n", localFile, remotePath)
	return nil
}

// performFileDownload handles file download logic and returns errors instead of calling os.Exit
func performFileDownload(regionCode, instanceIdentifier, remoteFile, localPath string) error {
	region := resolveRegion(regionCode)

	logging.LogInfo("Downloading file %s from instance %s to local path: %s", remoteFile, instanceIdentifier, localPath)

	ssmManager := ssm.NewManager(logger)
	ctx := context.Background()

	if err := ssmManager.DownloadFile(ctx, instanceIdentifier, region, remoteFile, localPath); err != nil {
		colors.PrintError("✗ File download failed: %s -> %s\n", remoteFile, localPath)
		return fmt.Errorf("file download failed: %w", err)
	}

	logging.LogSuccess("File download completed successfully")

	// Show colored success message
	colors.PrintSuccess("✓ File download completed successfully: %s -> %s\n", remoteFile, localPath)
	return nil
}

func init() {
	ssmTransferCmd.AddCommand(ssmUploadCmd)
	ssmTransferCmd.AddCommand(ssmDownloadCmd)

	ssmUploadCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmDownloadCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
}
