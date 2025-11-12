package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"

	"ztictl/internal/ssm"
	"ztictl/pkg/aws"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
)

// ssmStartCmd represents the ssm start command
var ssmStartCmd = &cobra.Command{
	Use:   "start [instance-identifier]",
	Short: "Start stopped EC2 instance(s)",
	Long: `Start stopped EC2 instance(s).
If no instance identifier is provided, an interactive fuzzy finder will be launched.
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.
Use --instances flag to specify multiple instance IDs (comma-separated).
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Examples:
  ztictl ssm start --region cac1                        # Interactive fuzzy finder
  ztictl ssm start i-1234567890abcdef0 --region cac1   # Specific instance
  ztictl ssm start --instances i-1234,i-5678 --region use1  # Multiple instances`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		instancesFlag, _ := cmd.Flags().GetString("instances")
		parallelFlag, _ := cmd.Flags().GetInt("parallel")

		if err := performPowerOperation(args, regionCode, instancesFlag, parallelFlag, "start"); err != nil {
			logging.LogError("Start operation failed: %v", err)
			os.Exit(1)
		}
	},
}

// ssmStopCmd represents the ssm stop command
var ssmStopCmd = &cobra.Command{
	Use:   "stop [instance-identifier]",
	Short: "Stop running EC2 instance(s)",
	Long: `Stop running EC2 instance(s).
If no instance identifier is provided, an interactive fuzzy finder will be launched.
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.
Use --instances flag to specify multiple instance IDs (comma-separated).
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Examples:
  ztictl ssm stop --region cac1                        # Interactive fuzzy finder
  ztictl ssm stop i-1234567890abcdef0 --region cac1   # Specific instance
  ztictl ssm stop --instances i-1234,i-5678 --region use1  # Multiple instances`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		instancesFlag, _ := cmd.Flags().GetString("instances")
		parallelFlag, _ := cmd.Flags().GetInt("parallel")

		if err := performPowerOperation(args, regionCode, instancesFlag, parallelFlag, "stop"); err != nil {
			logging.LogError("Stop operation failed: %v", err)
			os.Exit(1)
		}
	},
}

// ssmRebootCmd represents the ssm reboot command
var ssmRebootCmd = &cobra.Command{
	Use:   "reboot [instance-identifier]",
	Short: "Reboot running EC2 instance(s)",
	Long: `Reboot running EC2 instance(s).
If no instance identifier is provided, an interactive fuzzy finder will be launched.
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.
Use --instances flag to specify multiple instance IDs (comma-separated).
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Examples:
  ztictl ssm reboot --region cac1                        # Interactive fuzzy finder
  ztictl ssm reboot i-1234567890abcdef0 --region cac1   # Specific instance
  ztictl ssm reboot --instances i-1234,i-5678 --region use1  # Multiple instances`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		instancesFlag, _ := cmd.Flags().GetString("instances")
		parallelFlag, _ := cmd.Flags().GetInt("parallel")

		if err := performPowerOperation(args, regionCode, instancesFlag, parallelFlag, "reboot"); err != nil {
			logging.LogError("Reboot operation failed: %v", err)
			os.Exit(1)
		}
	},
}

// ssmStartTaggedCmd represents the ssm start-tagged command
var ssmStartTaggedCmd = &cobra.Command{
	Use:   "start-tagged",
	Short: "Start multiple stopped EC2 instances with specified tags (parallel execution)",
	Long: `Start multiple stopped EC2 instances that match the specified tags.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.
Use --tags flag to specify one or more tag filters in key=value format, separated by commas.
Use --instances to explicitly specify instance IDs to target (comma-separated).
Use --parallel to control maximum concurrent operations (default: number of CPU cores).

Examples:
  ztictl ssm start-tagged --region cac1 --tags Environment=Production
  ztictl ssm start-tagged --region use1 --tags Environment=dev,Component=fts --parallel 5
  ztictl ssm start-tagged --region cac1 --instances i-1234,i-5678`,
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		tagsFlag, _ := cmd.Flags().GetString("tags")
		instancesFlag, _ := cmd.Flags().GetString("instances")
		parallelFlag, _ := cmd.Flags().GetInt("parallel")

		region := resolveRegion(regionCode)

		// Validate arguments and flags
		if err := validateTaggedCommandArgs(tagsFlag, instancesFlag, parallelFlag); err != nil {
			colors.PrintError("✗ %v\n", err)
			logging.LogError("Validation error for start-tagged command: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		awsClient, err := aws.NewClient(ctx, aws.ClientOptions{Region: region})
		if err != nil {
			colors.PrintError("✗ Failed to create AWS client: %v\n", err)
			logging.LogError("Failed to create AWS client: %v", err)
			os.Exit(1)
		}

		var instanceIDs []string

		if instancesFlag != "" {
			// Use explicit instance IDs
			instanceIDs = strings.Split(instancesFlag, ",")
			for i, id := range instanceIDs {
				instanceIDs[i] = strings.TrimSpace(id)
			}
			logging.LogInfo("Starting %d explicit instance IDs in region: %s", len(instanceIDs), region)
		} else {
			// Use tag filtering to find instances
			instanceIDs, err = getInstanceIDsByTags(ctx, awsClient, tagsFlag)
			if err != nil {
				colors.PrintError("✗ Failed to find instances by tags: %v\n", err)
				logging.LogError("Failed to find instances by tags: %v", err)
				os.Exit(1)
			}
			logging.LogInfo("Starting %d instances with tags '%s' in region: %s", len(instanceIDs), tagsFlag, region)
		}

		if len(instanceIDs) == 0 {
			if instancesFlag != "" {
				logging.LogInfo("No instances specified")
			} else {
				logging.LogInfo("No instances found with tags: %s", tagsFlag)
			}
			return
		}

		// Create SSM manager for validation
		ssmManager := ssm.NewManager(logger)

		// Execute power operations in parallel
		startTime := time.Now()
		results := executePowerOperationParallel(ctx, awsClient, ssmManager, instanceIDs, "start", parallelFlag, region)
		totalDuration := time.Since(startTime)

		// Process and display results
		if err := displayPowerOperationResults(results, "start", totalDuration, parallelFlag); err != nil {
			os.Exit(1)
		}
	},
}

// ssmStopTaggedCmd represents the ssm stop-tagged command
var ssmStopTaggedCmd = &cobra.Command{
	Use:   "stop-tagged",
	Short: "Stop multiple running EC2 instances with specified tags (parallel execution)",
	Long: `Stop multiple running EC2 instances that match the specified tags.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.
Use --tags flag to specify one or more tag filters in key=value format, separated by commas.
Use --instances to explicitly specify instance IDs to target (comma-separated).
Use --parallel to control maximum concurrent operations (default: number of CPU cores).

Examples:
  ztictl ssm stop-tagged --region cac1 --tags Environment=Production
  ztictl ssm stop-tagged --region use1 --tags Environment=dev,Component=fts --parallel 5
  ztictl ssm stop-tagged --region cac1 --instances i-1234,i-5678`,
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		tagsFlag, _ := cmd.Flags().GetString("tags")
		instancesFlag, _ := cmd.Flags().GetString("instances")
		parallelFlag, _ := cmd.Flags().GetInt("parallel")

		region := resolveRegion(regionCode)

		// Validate arguments and flags
		if err := validateTaggedCommandArgs(tagsFlag, instancesFlag, parallelFlag); err != nil {
			colors.PrintError("✗ %v\n", err)
			logging.LogError("Validation error for stop-tagged command: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		awsClient, err := aws.NewClient(ctx, aws.ClientOptions{Region: region})
		if err != nil {
			colors.PrintError("✗ Failed to create AWS client: %v\n", err)
			logging.LogError("Failed to create AWS client: %v", err)
			os.Exit(1)
		}

		var instanceIDs []string

		if instancesFlag != "" {
			// Use explicit instance IDs
			instanceIDs = strings.Split(instancesFlag, ",")
			for i, id := range instanceIDs {
				instanceIDs[i] = strings.TrimSpace(id)
			}
			logging.LogInfo("Stopping %d explicit instance IDs in region: %s", len(instanceIDs), region)
		} else {
			// Use tag filtering to find instances
			instanceIDs, err = getInstanceIDsByTags(ctx, awsClient, tagsFlag)
			if err != nil {
				colors.PrintError("✗ Failed to find instances by tags: %v\n", err)
				logging.LogError("Failed to find instances by tags: %v", err)
				os.Exit(1)
			}
			logging.LogInfo("Stopping %d instances with tags '%s' in region: %s", len(instanceIDs), tagsFlag, region)
		}

		if len(instanceIDs) == 0 {
			if instancesFlag != "" {
				logging.LogInfo("No instances specified")
			} else {
				logging.LogInfo("No instances found with tags: %s", tagsFlag)
			}
			return
		}

		// Create SSM manager for validation
		ssmManager := ssm.NewManager(logger)

		// Execute power operations in parallel
		startTime := time.Now()
		results := executePowerOperationParallel(ctx, awsClient, ssmManager, instanceIDs, "stop", parallelFlag, region)
		totalDuration := time.Since(startTime)

		// Process and display results
		if err := displayPowerOperationResults(results, "stop", totalDuration, parallelFlag); err != nil {
			os.Exit(1)
		}
	},
}

// ssmRebootTaggedCmd represents the ssm reboot-tagged command
var ssmRebootTaggedCmd = &cobra.Command{
	Use:   "reboot-tagged",
	Short: "Reboot multiple running EC2 instances with specified tags (parallel execution)",
	Long: `Reboot multiple running EC2 instances that match the specified tags.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.
Use --tags flag to specify one or more tag filters in key=value format, separated by commas.
Use --instances to explicitly specify instance IDs to target (comma-separated).
Use --parallel to control maximum concurrent operations (default: number of CPU cores).

Examples:
  ztictl ssm reboot-tagged --region cac1 --tags Environment=Production
  ztictl ssm reboot-tagged --region use1 --tags Environment=dev,Component=fts --parallel 5
  ztictl ssm reboot-tagged --region cac1 --instances i-1234,i-5678`,
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		tagsFlag, _ := cmd.Flags().GetString("tags")
		instancesFlag, _ := cmd.Flags().GetString("instances")
		parallelFlag, _ := cmd.Flags().GetInt("parallel")

		region := resolveRegion(regionCode)

		// Validate arguments and flags
		if err := validateTaggedCommandArgs(tagsFlag, instancesFlag, parallelFlag); err != nil {
			colors.PrintError("✗ %v\n", err)
			logging.LogError("Validation error for reboot-tagged command: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		awsClient, err := aws.NewClient(ctx, aws.ClientOptions{Region: region})
		if err != nil {
			colors.PrintError("✗ Failed to create AWS client: %v\n", err)
			logging.LogError("Failed to create AWS client: %v", err)
			os.Exit(1)
		}

		var instanceIDs []string

		if instancesFlag != "" {
			// Use explicit instance IDs
			instanceIDs = strings.Split(instancesFlag, ",")
			for i, id := range instanceIDs {
				instanceIDs[i] = strings.TrimSpace(id)
			}
			logging.LogInfo("Rebooting %d explicit instance IDs in region: %s", len(instanceIDs), region)
		} else {
			// Use tag filtering to find instances
			instanceIDs, err = getInstanceIDsByTags(ctx, awsClient, tagsFlag)
			if err != nil {
				colors.PrintError("✗ Failed to find instances by tags: %v\n", err)
				logging.LogError("Failed to find instances by tags: %v", err)
				os.Exit(1)
			}
			logging.LogInfo("Rebooting %d instances with tags '%s' in region: %s", len(instanceIDs), tagsFlag, region)
		}

		if len(instanceIDs) == 0 {
			if instancesFlag != "" {
				logging.LogInfo("No instances specified")
			} else {
				logging.LogInfo("No instances found with tags: %s", tagsFlag)
			}
			return
		}

		// Create SSM manager for validation
		ssmManager := ssm.NewManager(logger)

		// Execute power operations in parallel
		startTime := time.Now()
		results := executePowerOperationParallel(ctx, awsClient, ssmManager, instanceIDs, "reboot", parallelFlag, region)
		totalDuration := time.Since(startTime)

		// Process and display results
		if err := displayPowerOperationResults(results, "reboot", totalDuration, parallelFlag); err != nil {
			os.Exit(1)
		}
	},
}

// PowerOperationResult represents the result of a power operation on an instance
type PowerOperationResult struct {
	InstanceID string
	Operation  string
	Error      error
	Duration   time.Duration
}

// performPowerOperation handles power operations with fuzzy finder support
func performPowerOperation(args []string, regionCode, instancesFlag string, parallelFlag int, operation string) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()

	// Case 1: Multiple instances via --instances flag
	if instancesFlag != "" {
		if len(args) > 0 {
			return fmt.Errorf("cannot specify both instance identifier and --instances flag")
		}

		instanceIDs := strings.Split(instancesFlag, ",")
		for i, id := range instanceIDs {
			instanceIDs[i] = strings.TrimSpace(id)
		}
		logging.LogInfo("%s %d instances in region: %s", capitalize(operation), len(instanceIDs), region)

		awsClient, err := aws.NewClient(ctx, aws.ClientOptions{Region: region})
		if err != nil {
			colors.PrintError("✗ Failed to create AWS client: %v\n", err)
			return fmt.Errorf("failed to create AWS client: %w", err)
		}

		// Create SSM manager for validation
		ssmManager := ssm.NewManager(logger)

		startTime := time.Now()
		results := executePowerOperationParallel(ctx, awsClient, ssmManager, instanceIDs, operation, parallelFlag, region)
		totalDuration := time.Since(startTime)
		return displayPowerOperationResults(results, operation, totalDuration, parallelFlag)
	}

	// Case 2: Single instance (direct or fuzzy finder)
	ssmManager := ssm.NewManager(logger)
	var instanceIdentifier string
	if len(args) > 0 {
		instanceIdentifier = args[0]
	}

	// Use SelectInstanceWithFallback to handle both direct and fuzzy finder modes
	instanceID, err := ssmManager.GetInstanceService().SelectInstanceWithFallback(
		ctx,
		instanceIdentifier,
		region,
		nil, // No filters for power commands
	)
	if err != nil {
		return fmt.Errorf("instance selection failed: %w", err)
	}

	logging.LogInfo("%s instance %s in region: %s", capitalize(operation), instanceID, region)

	// Validate instance state before attempting power operation
	requirements, err := buildRequirementsForOperation(operation)
	if err != nil {
		return err
	}
	if err := ValidateInstanceState(ctx, ssmManager, instanceID, region, requirements); err != nil {
		return err
	}

	awsClient, err := aws.NewClient(ctx, aws.ClientOptions{Region: region})
	if err != nil {
		colors.PrintError("✗ Failed to create AWS client: %v\n", err)
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Execute the power operation
	switch operation {
	case "start":
		_, err = awsClient.EC2.StartInstances(ctx, &ec2.StartInstancesInput{
			InstanceIds: []string{instanceID},
		})
	case "stop":
		_, err = awsClient.EC2.StopInstances(ctx, &ec2.StopInstancesInput{
			InstanceIds: []string{instanceID},
		})
	case "reboot":
		_, err = awsClient.EC2.RebootInstances(ctx, &ec2.RebootInstancesInput{
			InstanceIds: []string{instanceID},
		})
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}

	if err != nil {
		colors.PrintError("✗ Failed to %s instance %s\n", operation, instanceID)
		return fmt.Errorf("failed to %s instance: %w", operation, err)
	}

	colors.PrintSuccess("✓ Instance %s %s requested successfully\n", instanceID, operation)
	logging.LogInfo("Instance %s requested successfully", operation)
	return nil
}

// validateTaggedCommandArgs validates arguments and flags for tagged commands
func validateTaggedCommandArgs(tagsFlag, instancesFlag string, parallelFlag int) error {
	// Validate that we have either tags or instances specified
	if tagsFlag == "" && instancesFlag == "" {
		return fmt.Errorf("either --tags or --instances flag is required")
	}

	// Validate mutual exclusion - cannot specify both tags and instances
	if tagsFlag != "" && instancesFlag != "" {
		return fmt.Errorf("cannot specify both --tags and --instances flags")
	}

	// Validate parallel value
	if parallelFlag <= 0 {
		return fmt.Errorf("--parallel must be greater than 0")
	}

	return nil
}

// buildRequirementsForOperation creates validation requirements for a power operation
func buildRequirementsForOperation(operation string) (InstanceValidationRequirements, error) {
	req := InstanceValidationRequirements{
		RequireSSMOnline: false,
		Operation:        operation,
	}
	switch operation {
	case "start":
		req.AllowedStates = []string{"stopped"}
	case "stop", "reboot":
		req.AllowedStates = []string{"running"}
	default:
		return req, fmt.Errorf("unknown operation: %s", operation)
	}
	return req, nil
}

// capitalize returns a copy of the string with the first letter capitalized
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// getInstanceIDsByTags finds instance IDs by tag filters
func getInstanceIDsByTags(ctx context.Context, awsClient *aws.Client, tagsFlag string) ([]string, error) {
	// Parse tag filters
	filters := make([]types.Filter, 0)
	if tagsFlag != "" {
		tagPairs := strings.Split(tagsFlag, ",")
		for _, pair := range tagPairs {
			parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				filters = append(filters, types.Filter{
					Name:   awssdk.String("tag:" + key),
					Values: []string{value},
				})
			}
		}
	}

	result, err := awsClient.EC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	var instanceIDs []string
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instanceIDs = append(instanceIDs, *instance.InstanceId)
		}
	}

	return instanceIDs, nil
}

// executePowerOperationParallel runs power operations in parallel across multiple instances
func executePowerOperationParallel(ctx context.Context, awsClient *aws.Client, ssmManager *ssm.Manager, instanceIDs []string, operation string, maxParallel int, region string) []PowerOperationResult {
	// Create channels for work distribution and result collection
	// Buffers sized to instance count for simplicity - memory scales linearly with instance count.
	// For typical operations (< 1000 instances), memory overhead is negligible (~100KB).
	// For very large deployments (> 10000 instances), consider refactoring to smaller buffers.
	instanceChan := make(chan string, len(instanceIDs))
	resultChan := make(chan PowerOperationResult, len(instanceIDs))

	// Send instance IDs to work channel
	for _, instanceID := range instanceIDs {
		instanceChan <- instanceID
	}
	close(instanceChan)

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < maxParallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for instanceID := range instanceChan {
				startTime := time.Now()
				logging.LogInfo("Executing %s operation on instance %s", operation, instanceID)

				// Validate instance state before attempting operation
				requirements, err := buildRequirementsForOperation(operation)

				// Only proceed with validation and operation if no error yet
				if err == nil {
					err = ValidateInstanceState(ctx, ssmManager, instanceID, region, requirements)
				}

				// Execute power operation only if validation passed
				if err == nil {
					switch operation {
					case "start":
						_, err = awsClient.EC2.StartInstances(ctx, &ec2.StartInstancesInput{
							InstanceIds: []string{instanceID},
						})
					case "stop":
						_, err = awsClient.EC2.StopInstances(ctx, &ec2.StopInstancesInput{
							InstanceIds: []string{instanceID},
						})
					case "reboot":
						_, err = awsClient.EC2.RebootInstances(ctx, &ec2.RebootInstancesInput{
							InstanceIds: []string{instanceID},
						})
					default:
						err = fmt.Errorf("unknown operation: %s", operation)
					}
				}

				duration := time.Since(startTime)

				resultChan <- PowerOperationResult{
					InstanceID: instanceID,
					Operation:  operation,
					Error:      err,
					Duration:   duration,
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
	results := make([]PowerOperationResult, 0, len(instanceIDs))
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// displayPowerOperationResults displays the results of power operations and returns error if any operations failed
func displayPowerOperationResults(results []PowerOperationResult, operation string, totalDuration time.Duration, maxParallel int) error {
	successCount := 0
	for _, result := range results {
		fmt.Printf("\n")
		colors.PrintHeader("=== Instance: %s ===\n", result.InstanceID)
		colors.PrintHeader("Operation: %s\n", capitalize(operation))
		colors.PrintData("Execution Time: %v\n", result.Duration.Round(time.Millisecond))

		if result.Error != nil {
			colors.PrintError("✗ Operation failed: %v\n", result.Error)
		} else {
			successCount++
			colors.PrintSuccess("✓ %s requested successfully\n", capitalize(operation))
		}
	}

	// Summary
	fmt.Printf("\n")
	colors.PrintHeader("=== Operation Summary ===\n")
	colors.PrintData("Total instances: %d\n", len(results))
	colors.PrintData("Successful: %d\n", successCount)
	colors.PrintData("Failed: %d\n", len(results)-successCount)
	colors.PrintData("Total execution time: %v\n", totalDuration.Round(time.Millisecond))
	colors.PrintData("Max parallelism: %d\n", maxParallel)

	if successCount < len(results) {
		logging.LogWarn("Some %s operations failed: %d successful, %d failed", operation, successCount, len(results)-successCount)
		return fmt.Errorf("some %s operations failed: %d successful, %d failed", operation, successCount, len(results)-successCount)
	} else {
		logging.LogSuccess("All %s operations completed successfully", operation)
		return nil
	}
}

func init() {
	// Add flags for single instance commands
	ssmStartCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmStartCmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs to target explicitly")
	ssmStartCmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

	ssmStopCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmStopCmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs to target explicitly")
	ssmStopCmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

	ssmRebootCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmRebootCmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs to target explicitly")
	ssmRebootCmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

	// Add flags for tagged commands
	ssmStartTaggedCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmStartTaggedCmd.Flags().StringP("tags", "t", "", "Tag filters in key=value format, separated by commas")
	ssmStartTaggedCmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs to target explicitly")
	ssmStartTaggedCmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

	ssmStopTaggedCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmStopTaggedCmd.Flags().StringP("tags", "t", "", "Tag filters in key=value format, separated by commas")
	ssmStopTaggedCmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs to target explicitly")
	ssmStopTaggedCmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

	ssmRebootTaggedCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmRebootTaggedCmd.Flags().StringP("tags", "t", "", "Tag filters in key=value format, separated by commas")
	ssmRebootTaggedCmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs to target explicitly")
	ssmRebootTaggedCmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")
}
