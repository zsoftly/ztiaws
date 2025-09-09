package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"ztictl/internal/config"
	"ztictl/internal/ssm"
	awspkg "ztictl/pkg/aws"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Constants for output formatting
const (
	// MaxOutputLength is the maximum length of command output to display
	MaxOutputLength = 100
	// OutputTruncateLength is the length at which output is truncated
	OutputTruncateLength = 97
	// DefaultRegionParallelism is the default number of regions to process in parallel
	DefaultRegionParallelism = 5
)

// ssmExecMultiCmd represents the exec-multi command for multi-region command execution
var ssmExecMultiCmd = &cobra.Command{
	Use:   "exec-multi [regions|--regions|--all-regions|--region-group] <command>",
	Short: "Execute a command across multiple regions",
	Long: `Execute a command on EC2 instances across multiple AWS regions via SSM.
Configure your regions in ~/.ztictl.yaml or override with --regions flag.

Examples:
  # Basic multi-region with tags (positional)
  ztictl ssm exec-multi cac1,use1,euw1 --tags Environment=prod "uptime"
  
  # Override with --regions flag (supports both shortcodes and full names)
  ztictl ssm exec-multi --regions cac1,us-east-1,eu-west-1 --tags Environment=prod "uptime"
  
  # All configured regions from ~/.ztictl.yaml
  ztictl ssm exec-multi --all-regions --tags Component=web "systemctl status nginx"
  
  # Using region groups from config
  ztictl ssm exec-multi --region-group production --tags App=api "health-check"
  
  # With explicit instances
  ztictl ssm exec-multi --regions cac1,use1 --instances i-123,i-456 "hostname"
  
  # Control parallelism
  ztictl ssm exec-multi --regions cac1,use1,euw1 --tags Type=worker --parallel 5 "ps aux"
  
  # Control region parallelism
  ztictl ssm exec-multi --all-regions --tags App=api --parallel-regions 10 "health-check.sh"
  
  # Continue on region failures
  ztictl ssm exec-multi --regions cac1,use1 --tags App=api --continue-on-error "health-check.sh"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		allRegions, _ := cmd.Flags().GetBool("all-regions")
		regionsFlag, _ := cmd.Flags().GetString("regions")
		regionGroup, _ := cmd.Flags().GetString("region-group")
		tagsFlag, _ := cmd.Flags().GetString("tags")
		instancesFlag, _ := cmd.Flags().GetString("instances")
		parallelFlag, _ := cmd.Flags().GetInt("parallel")
		parallelRegionsFlag, _ := cmd.Flags().GetInt("parallel-regions")
		continueOnError, _ := cmd.Flags().GetBool("continue-on-error")

		// Parse regions
		var regions []string
		var command string

		// Priority: --regions flag > --all-regions > --region-group > positional argument
		if regionsFlag != "" {
			// Use regions from --regions flag (overrides everything)
			inputRegions := strings.Split(regionsFlag, ",")
			for _, r := range inputRegions {
				r = strings.TrimSpace(r)
				if r != "" {
					// Normalize to shortcode for consistency
					normalized := config.NormalizeRegion(r)
					regions = append(regions, normalized)
				}
			}
			command = strings.Join(args, " ")
		} else if allRegions {
			// Use configured regions
			cfg := config.Get()
			if len(cfg.Regions.Enabled) > 0 {
				regions = cfg.Regions.Enabled
			} else if len(cfg.Regions.Groups["all"]) > 0 {
				regions = cfg.Regions.Groups["all"]
			} else {
				// No regions configured - prompt for interactive setup
				colors.PrintWarning("⚠ No regions configured.\n")
				if err := config.InteractiveRegionSetup(); err != nil {
					colors.PrintError("✗ Region configuration required to use --all-regions\n")
					os.Exit(1)
				}
				// Reload config after setup
				cfg = config.Get()
				if len(cfg.Regions.Enabled) > 0 {
					regions = cfg.Regions.Enabled
				} else if len(cfg.Regions.Groups["all"]) > 0 {
					regions = cfg.Regions.Groups["all"]
				} else {
					colors.PrintError("✗ Failed to configure regions\n")
					os.Exit(1)
				}
			}
			command = strings.Join(args, " ")
		} else if regionGroup != "" {
			// Use region group
			cfg := config.Get()
			if len(cfg.Regions.Groups[regionGroup]) == 0 {
				colors.PrintError("✗ Region group '%s' not found in configuration\n", regionGroup)
				colors.PrintData("Available groups: ")
				if len(cfg.Regions.Groups) > 0 {
					var groups []string
					for name := range cfg.Regions.Groups {
						groups = append(groups, name)
					}
					colors.PrintData("%s\n", strings.Join(groups, ", "))
				} else {
					colors.PrintData("none\n")
				}
				os.Exit(1)
			}
			regions = cfg.Regions.Groups[regionGroup]
			command = strings.Join(args, " ")
		} else {
			// Check if first argument looks like regions (backward compatibility)
			if len(args) >= 2 {
				// Check if first arg contains comma or looks like a region
				firstArg := args[0]
				if strings.Contains(firstArg, ",") || looksLikeRegion(firstArg) {
					// Parse as regions list
					inputRegions := strings.Split(firstArg, ",")
					for _, r := range inputRegions {
						r = strings.TrimSpace(r)
						if r != "" {
							// Normalize to shortcode for consistency
							normalized := config.NormalizeRegion(r)
							regions = append(regions, normalized)
						}
					}
					command = strings.Join(args[1:], " ")
				} else {
					// No regions specified, show error
					colors.PrintError("✗ Must specify regions using one of:\n")
					colors.PrintData("  --regions cac1,use1,euw1\n")
					colors.PrintData("  --all-regions (uses ~/.ztictl.yaml)\n")
					colors.PrintData("  --region-group <group-name>\n")
					colors.PrintData("  cac1,use1,euw1 (positional argument)\n")
					os.Exit(1)
				}
			} else {
				colors.PrintError("✗ Must specify regions and a command\n")
				os.Exit(1)
			}
		}

		// Validate that we have either tags or instances specified
		if tagsFlag == "" && instancesFlag == "" {
			colors.PrintError("✗ Either --tags or --instances flag is required\n")
			os.Exit(1)
		}

		// Execute multi-region command
		success := executeMultiRegionCommand(regions, command, tagsFlag, instancesFlag, parallelFlag, parallelRegionsFlag, continueOnError)
		if !success {
			os.Exit(1)
		}
	},
}

// MultiRegionResult represents the result of execution in a single region
type MultiRegionResult struct {
	Region     string
	RegionName string
	Instances  []InstanceResult
	Error      error
	Duration   time.Duration
}

// InstanceResult represents the result of execution on a single instance
type InstanceResult struct {
	Instance    ssm.Instance
	Output      string
	ErrorOutput string
	ExitCode    int
	Success     bool
	Error       error
}

// RegionExecutionRequest represents a request to execute command in a region
type RegionExecutionRequest struct {
	RegionCode    string
	Command       string
	TagsFlag      string
	InstancesFlag string
	ParallelFlag  int
}

// executeMultiRegionCommand handles multi-region command execution with parallel processing
func executeMultiRegionCommand(regions []string, command, tagsFlag, instancesFlag string, parallelFlag, parallelRegionsFlag int, continueOnError bool) bool {
	startTime := time.Now()
	isDebug := viper.GetBool("debug")

	// Clean up and normalize region codes
	for i, region := range regions {
		region = strings.TrimSpace(region)
		// Normalize to shortcode for consistency
		regions[i] = config.NormalizeRegion(region)
	}

	colors.PrintHeader("=== MULTI-REGION EXECUTION STARTING ===\n")
	colors.PrintData("Regions: %s\n", strings.Join(regions, ", "))
	colors.PrintData("Command: %s\n", command)
	if tagsFlag != "" {
		colors.PrintData("Tags: %s\n", tagsFlag)
	}
	if instancesFlag != "" {
		colors.PrintData("Instances: %s\n", instancesFlag)
	}
	colors.PrintData("Parallelism per region: %d\n", parallelFlag)
	colors.PrintData("Parallel regions: %d\n", parallelRegionsFlag)
	colors.PrintData("Continue on error: %v\n\n", continueOnError)

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create channels for worker pool pattern
	regionChan := make(chan RegionExecutionRequest, len(regions))
	resultChan := make(chan MultiRegionResult, len(regions))

	// Send all region requests to the channel
	for _, regionCode := range regions {
		regionChan <- RegionExecutionRequest{
			RegionCode:    regionCode,
			Command:       command,
			TagsFlag:      tagsFlag,
			InstancesFlag: instancesFlag,
			ParallelFlag:  parallelFlag,
		}
	}
	close(regionChan)

	// Start worker goroutines for parallel region processing
	var wg sync.WaitGroup
	for i := 0; i < parallelRegionsFlag && i < len(regions); i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case request, ok := <-regionChan:
					if !ok {
						return // Channel closed, exit worker
					}

					// Check if context is cancelled before processing
					if ctx.Err() != nil {
						return
					}

					regionStartTime := time.Now()

					// Get full region name and description
					fullRegion := resolveRegion(request.RegionCode)
					regionDesc := awspkg.GetRegionDescription(request.RegionCode)

					// Only log if debug is enabled
					if isDebug {
						logging.LogInfo("[Worker %d] Processing region %s (%s)", workerID, request.RegionCode, fullRegion)
					}

					// Execute command in this region
					result := executeRegionCommandWithOutput(
						request.RegionCode,
						request.Command,
						request.TagsFlag,
						request.InstancesFlag,
						request.ParallelFlag,
						isDebug,
					)

					result.Duration = time.Since(regionStartTime)
					result.RegionName = regionDesc

					// Send result to channel
					select {
					case resultChan <- result:
						// Result sent successfully
					case <-ctx.Done():
						// Context cancelled, exit
						return
					}

					// Check if we should stop on error
					if result.Error != nil && !continueOnError {
						if isDebug {
							logging.LogError("Stopping execution due to failure in region %s", request.RegionCode)
						}
						// Cancel context to signal all workers to stop
						cancel()
						return
					}
				case <-ctx.Done():
					// Context cancelled, exit worker
					return
				}
			}
		}(i)
	}

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all results
	var results []MultiRegionResult
	overallSuccess := true
	for result := range resultChan {
		results = append(results, result)

		// Print region result with command outputs
		printRegionResult(result)

		if result.Error != nil || hasFailedInstances(result) {
			overallSuccess = false
		}
	}

	// Print multi-region summary
	printMultiRegionSummary(results, time.Since(startTime))

	return overallSuccess
}

// executeRegionCommandWithOutput executes command in a single region and returns detailed results
func executeRegionCommandWithOutput(regionCode, command, tagsFlag, instancesFlag string, parallelFlag int, isDebug bool) MultiRegionResult {
	result := MultiRegionResult{
		Region: regionCode,
	}

	region := resolveRegion(regionCode)
	ssmManager := ssm.NewManager(logger)
	ctx := context.Background()

	var instances []ssm.Instance
	var err error

	if instancesFlag != "" {
		// Use explicit instance IDs
		instanceIDs := strings.Split(instancesFlag, ",")
		for i, id := range instanceIDs {
			instanceIDs[i] = strings.TrimSpace(id)
		}

		if isDebug {
			logging.LogInfo("Targeting %d explicit instance IDs in region: %s", len(instanceIDs), region)
		}

		// Create Instance objects from IDs
		for _, instanceID := range instanceIDs {
			instances = append(instances, ssm.Instance{
				InstanceID: instanceID,
				Name:       instanceID,
			})
		}
	} else {
		// Use tag filtering
		if isDebug {
			logging.LogInfo("Finding instances with tags '%s' in region: %s", tagsFlag, region)
		}

		// List instances with the specified tags
		filters := &ssm.ListFilters{
			Tags: tagsFlag,
		}

		instances, err = ssmManager.ListInstances(ctx, region, filters)
		if err != nil {
			result.Error = fmt.Errorf("failed to list instances: %w", err)
			return result
		}
	}

	if len(instances) == 0 {
		if isDebug {
			logging.LogInfo("No instances found in region %s", region)
		}
		return result
	}

	if isDebug {
		logging.LogInfo("Executing command on %d instances in region %s", len(instances), region)
	}

	// Execute commands in parallel using existing function
	execResults := executeCommandParallel(ctx, ssmManager, instances, region, command, parallelFlag)

	// Convert results to our format
	for _, execResult := range execResults {
		instResult := InstanceResult{
			Instance: execResult.Instance,
		}

		if execResult.Error != nil {
			instResult.Error = execResult.Error
		} else if execResult.Result != nil {
			instResult.Output = execResult.Result.Output
			instResult.ErrorOutput = execResult.Result.ErrorOutput
			if execResult.Result.ExitCode != nil {
				instResult.ExitCode = int(*execResult.Result.ExitCode)
				instResult.Success = *execResult.Result.ExitCode == 0
			} else {
				instResult.Success = true
			}
		}

		result.Instances = append(result.Instances, instResult)
	}

	return result
}

// printRegionResult prints the result for a single region with command outputs
func printRegionResult(result MultiRegionResult) {
	fmt.Printf("\n")
	colors.PrintHeader("=== REGION: %s (%s) ===\n", result.Region, result.RegionName)

	if result.Error != nil {
		colors.PrintError("✗ Failed to execute in region: %v\n", result.Error)
		return
	}

	if len(result.Instances) == 0 {
		colors.PrintData("No instances found\n")
		return
	}

	// Print results for each instance
	for _, inst := range result.Instances {
		fmt.Printf("\n")
		if inst.Error != nil {
			colors.PrintError("✗ %s (%s): %v\n", inst.Instance.Name, inst.Instance.InstanceID, inst.Error)
		} else if inst.Success {
			colors.PrintSuccess("✓ %s (%s): success (exit code: %d)\n", inst.Instance.Name, inst.Instance.InstanceID, inst.ExitCode)

			// Show command output
			if inst.Output != "" {
				colors.PrintHeader("Output:\n")
				colors.PrintData("%s\n", inst.Output)
			}

			if inst.ErrorOutput != "" {
				colors.PrintHeader("Error Output:\n")
				colors.PrintData("%s\n", inst.ErrorOutput)
			}
		} else {
			colors.PrintError("✗ %s (%s): failed (exit code: %d)\n", inst.Instance.Name, inst.Instance.InstanceID, inst.ExitCode)

			// Show error output for failed commands
			if inst.ErrorOutput != "" {
				colors.PrintHeader("Error Output:\n")
				colors.PrintData("%s\n", inst.ErrorOutput)
			}

			if inst.Output != "" {
				colors.PrintHeader("Output:\n")
				colors.PrintData("%s\n", inst.Output)
			}
		}
	}

	// Region summary
	successful := 0
	failed := 0
	for _, inst := range result.Instances {
		if inst.Success && inst.Error == nil {
			successful++
		} else {
			failed++
		}
	}

	fmt.Printf("\n")
	colors.PrintData("Region Summary: %d/%d successful\n", successful, len(result.Instances))
	colors.PrintData("Duration: %v\n", result.Duration.Round(time.Millisecond))
}

// hasFailedInstances checks if any instances in the result failed
func hasFailedInstances(result MultiRegionResult) bool {
	for _, inst := range result.Instances {
		if !inst.Success || inst.Error != nil {
			return true
		}
	}
	return false
}

// printMultiRegionSummary prints the final summary of multi-region execution
func printMultiRegionSummary(results []MultiRegionResult, totalDuration time.Duration) {
	fmt.Printf("\n")
	colors.PrintHeader("=== MULTI-REGION SUMMARY ===\n")

	totalRegions := len(results)
	successfulRegions := 0
	totalInstances := 0
	totalSuccessful := 0
	totalFailed := 0

	for _, result := range results {
		instanceCount := len(result.Instances)
		successful := 0
		failed := 0

		for _, inst := range result.Instances {
			if inst.Success && inst.Error == nil {
				successful++
			} else {
				failed++
			}
		}

		totalInstances += instanceCount
		totalSuccessful += successful
		totalFailed += failed

		if result.Error == nil && failed == 0 {
			successfulRegions++
		}

		// Print per-region summary line
		status := "✓"
		if result.Error != nil || failed > 0 {
			status = "✗"
		}
		colors.PrintData("%s %s (%s): %d instances, %d successful, %d failed [%v]\n",
			status, result.Region, result.RegionName,
			instanceCount, successful, failed,
			result.Duration.Round(time.Millisecond))
	}

	fmt.Printf("\n")
	colors.PrintData("Regions processed: %d\n", totalRegions)
	colors.PrintData("Regions successful: %d\n", successfulRegions)
	colors.PrintData("Total instances: %d\n", totalInstances)
	colors.PrintData("Total successful: %d\n", totalSuccessful)
	colors.PrintData("Total failed: %d\n", totalFailed)
	colors.PrintData("Total execution time: %v\n", totalDuration.Round(time.Millisecond))

	if totalFailed > 0 {
		colors.PrintError("\n✗ Multi-region execution completed with failures\n")
	} else if totalInstances == 0 {
		colors.PrintWarning("\n⚠ No instances found matching criteria\n")
	} else {
		colors.PrintSuccess("\n✓ Multi-region execution completed successfully\n")
	}
}

// looksLikeRegion checks if a string looks like a region code or name
func looksLikeRegion(s string) bool {
	// Check if it's a known shortcode
	if _, exists := awspkg.RegionMapping[s]; exists {
		return true
	}

	// Check if it's a valid AWS region format (xx-xxxx-n)
	parts := strings.Split(s, "-")
	if len(parts) >= 3 && len(parts) <= 4 {
		return true
	}

	return false
}

func init() {
	// Add flags for exec-multi command
	ssmExecMultiCmd.Flags().StringP("regions", "r", "", "Override regions (comma-separated, supports shortcodes and full names)")
	ssmExecMultiCmd.Flags().BoolP("all-regions", "a", false, "Execute across all configured regions from ~/.ztictl.yaml")
	ssmExecMultiCmd.Flags().String("region-group", "", "Use predefined region group from config")
	ssmExecMultiCmd.Flags().StringP("tags", "t", "", "Tag filters in key=value format, separated by commas")
	ssmExecMultiCmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs to target")
	ssmExecMultiCmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent executions per region")
	ssmExecMultiCmd.Flags().IntP("parallel-regions", "P", DefaultRegionParallelism, "Maximum number of regions to process in parallel")
	ssmExecMultiCmd.Flags().BoolP("continue-on-error", "c", false, "Continue execution even if a region fails")
}
