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

// ssmExecCmd represents the exec command for single instance command execution
var ssmExecCmd = &cobra.Command{
	Use:   "exec <region-shortcode> <instance-identifier> <command>",
	Short: "Execute a command on a single instance",
	Long: `Execute a command on a single EC2 instance via SSM.
Region shortcuts supported: cac1, use1, euw1, etc.
Instance identifier can be an instance ID or name.

Examples:
  ztictl ssm exec cac1 i-1234567890abcdef0 "uptime"
  ztictl ssm exec use1 web-server "sudo systemctl status nginx"`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode := args[0]
		instanceIdentifier := args[1]
		command := strings.Join(args[2:], " ")

		region := resolveRegion(regionCode)

		logging.LogInfo("Executing command '%s' on instance %s in region: %s", command, instanceIdentifier, region)

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		result, err := ssmManager.ExecuteCommand(ctx, instanceIdentifier, region, command, "")
		if err != nil {
			colors.PrintError("✗ Failed to execute command on instance %s\n", instanceIdentifier)
			logging.LogError("Failed to execute command: %v", err)
			os.Exit(1)
		}

		colors.PrintHeader("Command executed successfully:\n")
		colors.PrintData("%s\n", result.Output)
		if result.ErrorOutput != "" {
			colors.PrintHeader("Error output:\n")
			colors.PrintData("%s\n", result.ErrorOutput)
		}

		if result.ExitCode != nil && *result.ExitCode != 0 {
			logging.LogWarn("Command exited with non-zero status: %d", *result.ExitCode)
			os.Exit(int(*result.ExitCode))
		}
	},
}

// ssmExecTaggedCmd represents the exec-tagged command for multi-instance execution
var ssmExecTaggedCmd = &cobra.Command{
	Use:   "exec-tagged <region-shortcode> <tag-key> <tag-value> <command>",
	Short: "Execute a command on all instances with specified tag",
	Long: `Execute a command on all EC2 instances that have the specified tag via SSM.
Region shortcuts supported: cac1, use1, euw1, etc.

Examples:
  ztictl ssm exec-tagged cac1 Environment Production "uptime"
  ztictl ssm exec-tagged use1 Role web-server "sudo systemctl restart nginx"`,
	Args: cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode := args[0]
		tagKey := args[1]
		tagValue := args[2]
		command := strings.Join(args[3:], " ")

		region := resolveRegion(regionCode)

		logging.LogInfo("Executing command '%s' on instances tagged %s=%s in region: %s", command, tagKey, tagValue, region)

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		// First, list instances with the specified tag
		filters := &ssm.ListFilters{
			Tag: fmt.Sprintf("%s=%s", tagKey, tagValue),
		}

		instances, err := ssmManager.ListInstances(ctx, region, filters)
		if err != nil {
			colors.PrintError("✗ Failed to list instances in region %s\n", region)
			logging.LogError("Failed to list instances: %v", err)
			os.Exit(1)
		}

		if len(instances) == 0 {
			logging.LogInfo("No instances found with tag %s=%s", tagKey, tagValue)
			return
		}

		logging.LogInfo("Found %d instances to execute command on", len(instances))

		// Execute command on each instance
		successCount := 0
		for _, instance := range instances {
			logging.LogInfo("Executing command on instance %s (%s)", instance.InstanceID, instance.Name)

			result, err := ssmManager.ExecuteCommand(ctx, instance.InstanceID, region, command, "")
			if err != nil {
				logging.LogError("Failed to execute command on instance %s (%s): %v", instance.InstanceID, instance.Name, err)
				continue
			}

			fmt.Printf("\n")
			colors.PrintHeader("=== Instance: %s (%s) ===\n", instance.Name, instance.InstanceID)
			colors.PrintHeader("Command: %s\n", command)
			colors.PrintHeader("Output:\n")
			colors.PrintData("%s\n", result.Output)

			if result.ErrorOutput != "" {
				colors.PrintHeader("Error output:\n")
				colors.PrintData("%s\n", result.ErrorOutput)
			}

			if result.ExitCode == nil || *result.ExitCode == 0 {
				successCount++
				exitCode := 0
				if result.ExitCode != nil {
					exitCode = int(*result.ExitCode)
				}
				colors.PrintSuccess("✓ Success (exit code: %d)\n", exitCode)
			} else {
				colors.PrintError("✗ Failed (exit code: %d)\n", int(*result.ExitCode))
			}
		}

		logging.LogInfo("Command execution completed: %d successful, %d failed", successCount, len(instances)-successCount)

		if successCount < len(instances) {
			os.Exit(1)
		}
	},
}

func init() {
	// Register exec commands - this ensures they're available when ssm.go's init runs
	// Commands will be added to ssmCmd in ssm.go's init function
}
