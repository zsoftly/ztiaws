package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"ztictl/internal/ssm"
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

		logger.Info("Executing command on instance",
			"identifier", instanceIdentifier,
			"region", region,
			"command", command)

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		result, err := ssmManager.ExecuteCommand(ctx, instanceIdentifier, region, command, "")
		if err != nil {
			logger.Error("Failed to execute command", "error", err)
			os.Exit(1)
		}

		fmt.Printf("Command executed successfully:\n%s\n", result.Output)
		if result.ErrorOutput != "" {
			fmt.Printf("Error output:\n%s\n", result.ErrorOutput)
		}

		if result.ExitCode != nil && *result.ExitCode != 0 {
			logger.Warn("Command exited with non-zero status", "exitCode", *result.ExitCode)
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

		logger.Info("Executing command on tagged instances",
			"tag", fmt.Sprintf("%s=%s", tagKey, tagValue),
			"region", region,
			"command", command)

		ssmManager := ssm.NewManager(logger)
		ctx := context.Background()

		// First, list instances with the specified tag
		filters := &ssm.ListFilters{
			Tag: fmt.Sprintf("%s=%s", tagKey, tagValue),
		}

		instances, err := ssmManager.ListInstances(ctx, region, filters)
		if err != nil {
			logger.Error("Failed to list instances", "error", err)
			os.Exit(1)
		}

		if len(instances) == 0 {
			logger.Info("No instances found with specified tag", "tag", fmt.Sprintf("%s=%s", tagKey, tagValue))
			return
		}

		logger.Info("Found instances", "count", len(instances))

		// Execute command on each instance
		successCount := 0
		for _, instance := range instances {
			logger.Info("Executing on instance", "instanceId", instance.InstanceID, "name", instance.Name)

			result, err := ssmManager.ExecuteCommand(ctx, instance.InstanceID, region, command, "")
			if err != nil {
				logger.Error("Failed to execute command on instance",
					"instanceId", instance.InstanceID,
					"name", instance.Name,
					"error", err)
				continue
			}

			fmt.Printf("\n=== Instance: %s (%s) ===\n", instance.Name, instance.InstanceID)
			fmt.Printf("Command: %s\n", command)
			fmt.Printf("Output:\n%s\n", result.Output)

			if result.ErrorOutput != "" {
				fmt.Printf("Error output:\n%s\n", result.ErrorOutput)
			}

			if result.ExitCode == nil || *result.ExitCode == 0 {
				successCount++
				exitCode := 0
				if result.ExitCode != nil {
					exitCode = int(*result.ExitCode)
				}
				fmt.Printf("✓ Success (exit code: %d)\n", exitCode)
			} else {
				fmt.Printf("✗ Failed (exit code: %d)\n", int(*result.ExitCode))
			}
		}

		logger.Info("Command execution completed",
			"total", len(instances),
			"successful", successCount,
			"failed", len(instances)-successCount)

		if successCount < len(instances) {
			os.Exit(1)
		}
	},
}

func init() {
	// Register exec commands - this ensures they're available when ssm.go's init runs
	// Commands will be added to ssmCmd in ssm.go's init function
}
