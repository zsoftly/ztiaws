package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"ztictl/internal/config"
	"ztictl/internal/ssm"
)

// ssmCmd represents the ssm command
var ssmCmd = &cobra.Command{
	Use:   "ssm",
	Short: "AWS Systems Manager operations",
	Long: `Manage AWS Systems Manager operations including instance connections, command execution, 
file transfers, and port forwarding through SSM.

Examples:
  ztictl ssm connect <instance>         # Connect to instance via SSM
  ztictl ssm list [filters]             # List SSM-enabled instances
  ztictl ssm forward <instance> <ports> # Port forwarding via SSM
  ztictl ssm transfer <src> <dst>       # File transfer via SSM
  ztictl ssm command <instance> <cmd>   # Execute command via SSM
  ztictl ssm status [instance]          # Check SSM agent status`,
}

// ssmConnectCmd represents the ssm connect command
var ssmConnectCmd = &cobra.Command{
	Use:   "connect <instance-identifier>",
	Short: "Connect to an instance via SSM Session Manager",
	Long: `Connect to an EC2 instance using SSM Session Manager.
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		region, _ := cmd.Flags().GetString("region")
		if region == "" {
			region = config.Get().DefaultRegion
		}

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

// ssmListCmd represents the ssm list command
var ssmListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SSM-enabled instances",
	Long: `List EC2 instances that are available through AWS Systems Manager.
Optionally filter by tags, status, or name patterns.`,
	Run: func(cmd *cobra.Command, args []string) {
		region, _ := cmd.Flags().GetString("region")
		if region == "" {
			region = config.Get().DefaultRegion
		}

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

// ssmCommandCmd represents the ssm command command
var ssmCommandCmd = &cobra.Command{
	Use:   "command <instance-identifier> <command>",
	Short: "Execute a command on an instance via SSM",
	Long: `Execute a shell command on an EC2 instance using SSM Run Command.
Instance identifier can be an instance ID or instance name.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		region, _ := cmd.Flags().GetString("region")
		if region == "" {
			region = config.Get().DefaultRegion
		}

		instanceIdentifier := args[0]
		command := args[1]
		comment, _ := cmd.Flags().GetString("comment")

		logger.Info("Executing command", "instance", instanceIdentifier, "command", command, "region", region)

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		result, err := ssmManager.ExecuteCommand(ctx, instanceIdentifier, region, command, comment)
		if err != nil {
			logger.Error("Command execution failed", "error", err)
			os.Exit(1)
		}

		fmt.Printf("\n=== Command Execution Result ===\n")
		fmt.Printf("Instance: %s\n", result.InstanceID)
		fmt.Printf("Command: %s\n", result.Command)
		fmt.Printf("Status: %s\n", result.Status)
		if result.ExitCode != nil {
			fmt.Printf("Exit Code: %d\n", *result.ExitCode)
		}
		fmt.Printf("\n--- Output ---\n")
		fmt.Print(result.Output)
		if result.ErrorOutput != "" {
			fmt.Printf("\n--- Error Output ---\n")
			fmt.Print(result.ErrorOutput)
		}
	},
}

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
Files are transferred directly for small files or via S3 for large files.`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		region, _ := cmd.Flags().GetString("region")
		if region == "" {
			region = config.Get().DefaultRegion
		}

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
Files are transferred directly for small files or via S3 for large files.`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		region, _ := cmd.Flags().GetString("region")
		if region == "" {
			region = config.Get().DefaultRegion
		}

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

// ssmForwardCmd represents the ssm forward command
var ssmForwardCmd = &cobra.Command{
	Use:   "forward <instance-identifier> <local-port>:<remote-port>",
	Short: "Forward ports through SSM tunnel",
	Long: `Forward local ports to remote ports on an EC2 instance through SSM.
Format: local-port:remote-port (e.g., 8080:80)`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		region, _ := cmd.Flags().GetString("region")
		if region == "" {
			region = config.Get().DefaultRegion
		}

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
If no instance is specified, shows status for all instances in the region.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		region, _ := cmd.Flags().GetString("region")
		if region == "" {
			region = config.Get().DefaultRegion
		}

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
	rootCmd.AddCommand(ssmCmd)
	
	// Add subcommands
	ssmCmd.AddCommand(ssmConnectCmd)
	ssmCmd.AddCommand(ssmListCmd)
	ssmCmd.AddCommand(ssmCommandCmd)
	ssmCmd.AddCommand(ssmTransferCmd)
	ssmCmd.AddCommand(ssmForwardCmd)
	ssmCmd.AddCommand(ssmStatusCmd)
	
	// Add transfer subcommands
	ssmTransferCmd.AddCommand(ssmUploadCmd)
	ssmTransferCmd.AddCommand(ssmDownloadCmd)

	// Add region flag to all SSM commands
	for _, cmd := range []*cobra.Command{ssmConnectCmd, ssmListCmd, ssmCommandCmd, ssmUploadCmd, ssmDownloadCmd, ssmForwardCmd, ssmStatusCmd} {
		cmd.Flags().StringP("region", "r", "", "AWS region (default from config)")
	}

	// Add additional flags
	ssmListCmd.Flags().StringP("tag", "t", "", "Filter by tag (format: key=value)")
	ssmListCmd.Flags().StringP("status", "s", "", "Filter by status (running, stopped, etc.)")
	ssmListCmd.Flags().StringP("name", "n", "", "Filter by name pattern")
	
	ssmCommandCmd.Flags().StringP("comment", "c", "", "Comment for the command execution")
}
