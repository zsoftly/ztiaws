package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"ztictl/internal/ssm"
)

// ssmCommandCmd represents the ssm command command
var ssmCommandCmd = &cobra.Command{
	Use:   "command <instance-identifier> <command>",
	Short: "Execute a command on an instance",
	Long: `Execute a command on an EC2 instance via SSM Run Command.
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		region := resolveRegion(regionCode)

		instanceIdentifier := args[0]
		command := strings.Join(args[1:], " ")
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

func init() {
	ssmCommandCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmCommandCmd.Flags().StringP("comment", "c", "", "Comment for the command execution")
}
