package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"ztictl/internal/ssm"
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
		region := resolveRegion(regionCode)

		instanceIdentifier := args[0]
		localFile := args[1]
		remotePath := args[2]

		logger.Info("Uploading file", "instance", instanceIdentifier, "local", localFile, "remote", remotePath)

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		if err := ssmManager.UploadFile(ctx, instanceIdentifier, region, localFile, remotePath); err != nil {
			logger.Error("File upload failed", "error", err)
			os.Exit(1)
		}

		logger.Info("File upload completed successfully")
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
		region := resolveRegion(regionCode)

		instanceIdentifier := args[0]
		remoteFile := args[1]
		localPath := args[2]

		logger.Info("Downloading file", "instance", instanceIdentifier, "remote", remoteFile, "local", localPath)

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		if err := ssmManager.DownloadFile(ctx, instanceIdentifier, region, remoteFile, localPath); err != nil {
			logger.Error("File download failed", "error", err)
			os.Exit(1)
		}

		logger.Info("File download completed successfully")
	},
}

func init() {
	ssmTransferCmd.AddCommand(ssmUploadCmd)
	ssmTransferCmd.AddCommand(ssmDownloadCmd)

	ssmUploadCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmDownloadCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
}
