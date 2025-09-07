package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"ztictl/internal/ssm"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
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
		instanceIdentifier := args[0]
		command := strings.Join(args[1:], " ")
		comment, _ := cmd.Flags().GetString("comment")

		if err := performCommandExecution(regionCode, instanceIdentifier, command, comment); err != nil {
			logging.LogError("Command execution failed: %v", err)
			os.Exit(1)
		}
	},
}

// performCommandExecution handles command execution logic and returns errors instead of calling os.Exit
func performCommandExecution(regionCode, instanceIdentifier, command, comment string) error {
	region := resolveRegion(regionCode)

	logging.LogInfo("Executing command '%s' on instance %s in region: %s", command, instanceIdentifier, region)

	ssmManager := ssm.NewManager(logger)
	ctx := context.Background()

	result, err := ssmManager.ExecuteCommand(ctx, instanceIdentifier, region, command, comment)
	if err != nil {
		colors.PrintError("✗ Command execution failed on instance %s\n", instanceIdentifier)
		return fmt.Errorf("command execution failed: %w", err)
	}

	fmt.Printf("\n")
	colors.PrintHeader("=== Command Execution Result ===\n")
	colors.PrintData("Instance: %s\n", result.InstanceID)
	colors.PrintData("Command: %s\n", result.Command)
	colors.PrintData("Status: %s\n", result.Status)
	if result.ExitCode != nil {
		colors.PrintData("Exit Code: %d\n", *result.ExitCode)
	}
	colors.PrintHeader("\n--- Output ---\n")
	colors.PrintData("%s", result.Output)
	if result.ErrorOutput != "" {
		colors.PrintHeader("\n--- Error Output ---\n")
		colors.PrintData("%s", result.ErrorOutput)
	}

	return nil
}

func init() {
	ssmCommandCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmCommandCmd.Flags().StringP("comment", "c", "", "Comment for the command execution")
}
