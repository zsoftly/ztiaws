package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"ztictl/internal/interactive"
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

		if err := executeSingleCommand(regionCode, instanceIdentifier, command); err != nil {
			logging.LogError("Command execution failed: %v", err)
			// Check if it's a non-zero exit code error and exit with that code
			if strings.Contains(err.Error(), "command exited with non-zero status:") {
				// Extract exit code from error message
				parts := strings.Split(err.Error(), ": ")
				if len(parts) > 1 {
					if exitCode, parseErr := strconv.Atoi(parts[len(parts)-1]); parseErr == nil {
						os.Exit(exitCode)
					}
				}
			}
			os.Exit(1)
		}
	},
}

// ssmExecTaggedCmd represents the exec-tagged command for multi-instance execution
var ssmExecTaggedCmd = &cobra.Command{
	Use:   "exec-tagged <region-shortcode> <command>",
	Short: "Execute a command on instances with specified tags (parallel execution)",
	Long: `Execute a command on EC2 instances that match the specified tags via SSM.
Region shortcuts supported: cac1, use1, euw1, etc.
Use --tags flag to specify one or more tag filters in key=value format, separated by commas.
Use --instances to explicitly specify instance IDs to target (comma-separated).
Use --parallel to control maximum concurrent executions (default: number of CPU cores).

ALL COMMANDS RUN IN PARALLEL BY DEFAULT for improved performance at scale.

Examples:
  ztictl ssm exec-tagged cac1 --tags Environment=Production "uptime"
  ztictl ssm exec-tagged use1 --tags Environment=dev,Component=fts --parallel 5 "sudo systemctl restart nginx"
  ztictl ssm exec-tagged cac1 --instances i-1234,i-5678 "ps aux | grep java"
  ztictl ssm exec-tagged use1 --tags Team=backend --parallel 10 "df -h"`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode := args[0]
		command := strings.Join(args[1:], " ")

		// Get flags
		tagsFlag, _ := cmd.Flags().GetString("tags")
		instancesFlag, _ := cmd.Flags().GetString("instances")
		parallelFlag, _ := cmd.Flags().GetInt("parallel")

		successful, err := executeTaggedCommand(regionCode, command, tagsFlag, instancesFlag, parallelFlag)
		if err != nil {
			logging.LogError("Tagged command execution failed: %v", err)
			os.Exit(1)
		}

		if !successful {
			os.Exit(1)
		}
	},
}

// ParallelExecutionResult represents the result of a parallel command execution
type ParallelExecutionResult struct {
	Instance interactive.Instance
	Result   *ssm.CommandResult
	Error    error
	Duration time.Duration
}

// executeCommandParallel runs commands in parallel across multiple instances
func executeCommandParallel(ctx context.Context, ssmManager *ssm.Manager, instances []interactive.Instance, region, command string, maxParallel int) []ParallelExecutionResult {
	// Create channels for work distribution and result collection
	instanceChan := make(chan interactive.Instance, len(instances))
	resultChan := make(chan ParallelExecutionResult, len(instances))

	// Send instances to work channel
	for _, instance := range instances {
		instanceChan <- instance
	}
	close(instanceChan)

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < maxParallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for instance := range instanceChan {
				startTime := time.Now()
				logging.LogInfo("Executing command on instance %s (%s)", instance.InstanceID, instance.Name)

				result, err := ssmManager.ExecuteCommand(ctx, instance.InstanceID, region, command, "")
				duration := time.Since(startTime)

				resultChan <- ParallelExecutionResult{
					Instance: instance,
					Result:   result,
					Error:    err,
					Duration: duration,
				}
			}
		}()
	}

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all results
	var results []ParallelExecutionResult
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// executeSingleCommand handles single instance command execution and returns errors instead of calling os.Exit
func executeSingleCommand(regionCode, instanceIdentifier, command string) error {
	region := resolveRegion(regionCode)

	logging.LogInfo("Executing command '%s' on instance %s in region: %s", command, instanceIdentifier, region)

	ssmManager := ssm.NewManager(logger)
	ctx := context.Background()

	result, err := ssmManager.ExecuteCommand(ctx, instanceIdentifier, region, command, "")
	if err != nil {
		colors.PrintError("✗ Failed to execute command on instance %s\n", instanceIdentifier)
		return fmt.Errorf("failed to execute command: %w", err)
	}

	colors.PrintHeader("Command executed successfully:\n")
	colors.PrintData("%s\n", result.Output)
	if result.ErrorOutput != "" {
		colors.PrintHeader("Error output:\n")
		colors.PrintData("%s\n", result.ErrorOutput)
	}

	if result.ExitCode != nil && *result.ExitCode != 0 {
		logging.LogWarn("Command exited with non-zero status: %d", *result.ExitCode)
		return fmt.Errorf("command exited with non-zero status: %d", *result.ExitCode)
	}

	return nil
}

// validateExecTaggedArgs validates arguments and flags for exec-tagged command
func validateExecTaggedArgs(tagsFlag, instancesFlag string, parallelFlag int) error {
	// Validate that we have either tags or instances specified
	if tagsFlag == "" && instancesFlag == "" {
		colors.PrintError("✗ Either --tags or --instances flag is required\n")
		return fmt.Errorf("no tags or instances specified for exec-tagged command")
	}

	// Validate mutual exclusion - cannot specify both tags and instances
	if tagsFlag != "" && instancesFlag != "" {
		colors.PrintError("✗ Cannot specify both --tags and --instances flags\n")
		return fmt.Errorf("both tags and instances flags provided - only one is allowed")
	}

	// Validate parallel value
	if parallelFlag <= 0 {
		colors.PrintError("✗ --parallel must be greater than 0\n")
		return fmt.Errorf("parallel must be greater than 0")
	}

	return nil
}

// executeTaggedCommand handles tagged command execution and returns success status and errors instead of calling os.Exit
func executeTaggedCommand(regionCode, command, tagsFlag, instancesFlag string, parallelFlag int) (bool, error) {
	if err := validateExecTaggedArgs(tagsFlag, instancesFlag, parallelFlag); err != nil {
		return false, err
	}

	region := resolveRegion(regionCode)
	ssmManager := ssm.NewManager(logger)
	ctx := context.Background()

	var instances []interactive.Instance
	var err error

	if instancesFlag != "" {
		// Use explicit instance IDs
		instanceIDs := strings.Split(instancesFlag, ",")
		for i, id := range instanceIDs {
			instanceIDs[i] = strings.TrimSpace(id)
		}

		logging.LogInfo("Targeting %d explicit instance IDs in region: %s", len(instanceIDs), region)

		// Create Instance objects from IDs (we'll resolve details during execution)
		for _, instanceID := range instanceIDs {
			instances = append(instances, interactive.Instance{
				InstanceID: instanceID,
				Name:       instanceID, // Will be resolved later if needed
			})
		}
	} else {
		// Use tag filtering
		logging.LogInfo("Executing command '%s' on instances with tags '%s' in region: %s", command, tagsFlag, region)

		// First, list instances with the specified tags
		filters := &ssm.ListFilters{
			Tags: tagsFlag,
		}

		instances, err = ssmManager.ListInstances(ctx, region, filters)
		if err != nil {
			colors.PrintError("✗ Failed to list instances in region %s\n", region)
			return false, fmt.Errorf("failed to list instances: %w", err)
		}
	}

	if len(instances) == 0 {
		if instancesFlag != "" {
			logging.LogInfo("No instances specified")
		} else {
			logging.LogInfo("No instances found with tags: %s", tagsFlag)
		}
		return true, nil
	}

	logging.LogInfo("Executing command on %d instances with parallelism: %d", len(instances), parallelFlag)

	// Execute commands in parallel
	startTime := time.Now()
	results := executeCommandParallel(ctx, ssmManager, instances, region, command, parallelFlag)
	totalDuration := time.Since(startTime)

	// Process and display results
	successCount := 0
	for _, result := range results {
		fmt.Printf("\n")
		colors.PrintHeader("=== Instance: %s (%s) ===\n", result.Instance.Name, result.Instance.InstanceID)
		colors.PrintHeader("Command: %s\n", command)
		colors.PrintData("Execution Time: %v\n", result.Duration.Round(time.Millisecond))

		if result.Error != nil {
			colors.PrintError("✗ Execution failed: %v\n", result.Error)
			continue
		}

		colors.PrintHeader("Output:\n")
		colors.PrintData("%s\n", result.Result.Output)

		if result.Result.ErrorOutput != "" {
			colors.PrintHeader("Error output:\n")
			colors.PrintData("%s\n", result.Result.ErrorOutput)
		}

		if result.Result.ExitCode == nil || *result.Result.ExitCode == 0 {
			successCount++
			exitCode := 0
			if result.Result.ExitCode != nil {
				exitCode = int(*result.Result.ExitCode)
			}
			colors.PrintSuccess("✓ Success (exit code: %d)\n", exitCode)
		} else {
			colors.PrintError("✗ Failed (exit code: %d)\n", int(*result.Result.ExitCode))
		}
	}

	// Summary
	fmt.Printf("\n")
	colors.PrintHeader("=== Execution Summary ===\n")
	colors.PrintData("Total instances: %d\n", len(instances))
	colors.PrintData("Successful: %d\n", successCount)
	colors.PrintData("Failed: %d\n", len(instances)-successCount)
	colors.PrintData("Total execution time: %v\n", totalDuration.Round(time.Millisecond))
	colors.PrintData("Max parallelism: %d\n", parallelFlag)

	if successCount < len(instances) {
		logging.LogWarn("Some executions failed: %d successful, %d failed", successCount, len(instances)-successCount)
		return false, nil
	} else {
		logging.LogSuccess("All executions completed successfully")
		return true, nil
	}
}

func init() {
	// Add flags for exec-tagged command
	ssmExecTaggedCmd.Flags().StringP("tags", "t", "", "Tag filters in key=value format, separated by commas")
	ssmExecTaggedCmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs to target explicitly")
	ssmExecTaggedCmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent executions")

	// Register exec commands - this ensures they're available when ssm.go's init runs
	// Commands will be added to ssmCmd in ssm.go's init function
}
