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

	"ztictl/pkg/aws"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
)

// ssmStartCmd represents the ssm start command
var ssmStartCmd = &cobra.Command{
	Use:   "start [instance-identifier]",
	Short: "Start stopped EC2 instance(s)",
	Long: `Start stopped EC2 instance(s).
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.
Use --instances flag to specify multiple instance IDs (comma-separated).
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Examples:
  ztictl ssm start i-1234567890abcdef0 --region cac1
  ztictl ssm start --instances i-1234,i-5678 --region use1`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		instancesFlag, _ := cmd.Flags().GetString("instances")
		parallelFlag, _ := cmd.Flags().GetInt("parallel")

		region := resolveRegion(regionCode)

		// Validate arguments and flags
		if len(args) == 0 && instancesFlag == "" {
			colors.PrintError("✗ Either provide an instance identifier or use --instances flag\n")
			os.Exit(1)
		}

		// Validate mutual exclusion
		if len(args) > 0 && instancesFlag != "" {
			colors.PrintError("✗ Cannot specify both instance identifier and --instances flag\n")
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
			// Multiple instances via --instances flag
			instanceIDs = strings.Split(instancesFlag, ",")
			for i, id := range instanceIDs {
				instanceIDs[i] = strings.TrimSpace(id)
			}
			logging.LogInfo("Starting %d instances in region: %s", len(instanceIDs), region)

			// Execute in parallel
			startTime := time.Now()
			results := executePowerOperationParallel(ctx, awsClient, instanceIDs, region, "start", parallelFlag)
			totalDuration := time.Since(startTime)
			displayPowerOperationResults(results, "start", totalDuration, parallelFlag)
		} else {
			// Single instance
			instanceIdentifier := args[0]
			logging.LogInfo("Starting instance %s in region: %s", instanceIdentifier, region)

			// Resolve instance ID if name was provided
			instanceID, err := resolveInstanceID(ctx, awsClient, instanceIdentifier, region)
			if err != nil {
				colors.PrintError("✗ Failed to resolve instance: %v\n", err)
				logging.LogError("Failed to resolve instance: %v", err)
				os.Exit(1)
			}

			// Start the instance
			_, err = awsClient.EC2.StartInstances(ctx, &ec2.StartInstancesInput{
				InstanceIds: []string{instanceID},
			})
			if err != nil {
				colors.PrintError("✗ Failed to start instance %s\n", instanceID)
				logging.LogError("Failed to start instance: %v", err)
				os.Exit(1)
			}

			colors.PrintSuccess("✓ Instance %s (%s) start requested successfully\n", instanceIdentifier, instanceID)
			logging.LogInfo("Instance start requested successfully")
		}
	},
}

// ssmStopCmd represents the ssm stop command
var ssmStopCmd = &cobra.Command{
	Use:   "stop [instance-identifier]",
	Short: "Stop running EC2 instance(s)",
	Long: `Stop running EC2 instance(s).
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.
Use --instances flag to specify multiple instance IDs (comma-separated).
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Examples:
  ztictl ssm stop i-1234567890abcdef0 --region cac1
  ztictl ssm stop --instances i-1234,i-5678 --region use1`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		instancesFlag, _ := cmd.Flags().GetString("instances")
		parallelFlag, _ := cmd.Flags().GetInt("parallel")

		region := resolveRegion(regionCode)

		// Validate arguments and flags
		if len(args) == 0 && instancesFlag == "" {
			colors.PrintError("✗ Either provide an instance identifier or use --instances flag\n")
			os.Exit(1)
		}

		// Validate mutual exclusion
		if len(args) > 0 && instancesFlag != "" {
			colors.PrintError("✗ Cannot specify both instance identifier and --instances flag\n")
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
			// Multiple instances via --instances flag
			instanceIDs = strings.Split(instancesFlag, ",")
			for i, id := range instanceIDs {
				instanceIDs[i] = strings.TrimSpace(id)
			}
			logging.LogInfo("Stopping %d instances in region: %s", len(instanceIDs), region)

			// Execute in parallel
			startTime := time.Now()
			results := executePowerOperationParallel(ctx, awsClient, instanceIDs, region, "stop", parallelFlag)
			totalDuration := time.Since(startTime)
			displayPowerOperationResults(results, "stop", totalDuration, parallelFlag)
		} else {
			// Single instance
			instanceIdentifier := args[0]
			logging.LogInfo("Stopping instance %s in region: %s", instanceIdentifier, region)

			// Resolve instance ID if name was provided
			instanceID, err := resolveInstanceID(ctx, awsClient, instanceIdentifier, region)
			if err != nil {
				colors.PrintError("✗ Failed to resolve instance: %v\n", err)
				logging.LogError("Failed to resolve instance: %v", err)
				os.Exit(1)
			}

			// Stop the instance
			_, err = awsClient.EC2.StopInstances(ctx, &ec2.StopInstancesInput{
				InstanceIds: []string{instanceID},
			})
			if err != nil {
				colors.PrintError("✗ Failed to stop instance %s\n", instanceID)
				logging.LogError("Failed to stop instance: %v", err)
				os.Exit(1)
			}

			colors.PrintSuccess("✓ Instance %s (%s) stop requested successfully\n", instanceIdentifier, instanceID)
			logging.LogInfo("Instance stop requested successfully")
		}
	},
}

// ssmRebootCmd represents the ssm reboot command
var ssmRebootCmd = &cobra.Command{
	Use:   "reboot [instance-identifier]",
	Short: "Reboot running EC2 instance(s)",
	Long: `Reboot running EC2 instance(s).
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.
Use --instances flag to specify multiple instance IDs (comma-separated).
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Examples:
  ztictl ssm reboot i-1234567890abcdef0 --region cac1
  ztictl ssm reboot --instances i-1234,i-5678 --region use1`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		instancesFlag, _ := cmd.Flags().GetString("instances")
		parallelFlag, _ := cmd.Flags().GetInt("parallel")

		region := resolveRegion(regionCode)

		// Validate arguments and flags
		if len(args) == 0 && instancesFlag == "" {
			colors.PrintError("✗ Either provide an instance identifier or use --instances flag\n")
			os.Exit(1)
		}

		// Validate mutual exclusion
		if len(args) > 0 && instancesFlag != "" {
			colors.PrintError("✗ Cannot specify both instance identifier and --instances flag\n")
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
			// Multiple instances via --instances flag
			instanceIDs = strings.Split(instancesFlag, ",")
			for i, id := range instanceIDs {
				instanceIDs[i] = strings.TrimSpace(id)
			}
			logging.LogInfo("Rebooting %d instances in region: %s", len(instanceIDs), region)

			// Execute in parallel
			startTime := time.Now()
			results := executePowerOperationParallel(ctx, awsClient, instanceIDs, region, "reboot", parallelFlag)
			totalDuration := time.Since(startTime)
			displayPowerOperationResults(results, "reboot", totalDuration, parallelFlag)
		} else {
			// Single instance
			instanceIdentifier := args[0]
			logging.LogInfo("Rebooting instance %s in region: %s", instanceIdentifier, region)

			// Resolve instance ID if name was provided
			instanceID, err := resolveInstanceID(ctx, awsClient, instanceIdentifier, region)
			if err != nil {
				colors.PrintError("✗ Failed to resolve instance: %v\n", err)
				logging.LogError("Failed to resolve instance: %v", err)
				os.Exit(1)
			}

			// Reboot the instance
			_, err = awsClient.EC2.RebootInstances(ctx, &ec2.RebootInstancesInput{
				InstanceIds: []string{instanceID},
			})
			if err != nil {
				colors.PrintError("✗ Failed to reboot instance %s\n", instanceID)
				logging.LogError("Failed to reboot instance: %v", err)
				os.Exit(1)
			}

			colors.PrintSuccess("✓ Instance %s (%s) reboot requested successfully\n", instanceIdentifier, instanceID)
			logging.LogInfo("Instance reboot requested successfully")
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

		// Validate that we have either tags or instances specified
		if tagsFlag == "" && instancesFlag == "" {
			colors.PrintError("✗ Either --tags or --instances flag is required\n")
			logging.LogError("No tags or instances specified for start-tagged command")
			os.Exit(1)
		}

		// Validate mutual exclusion - cannot specify both tags and instances
		if tagsFlag != "" && instancesFlag != "" {
			colors.PrintError("✗ Cannot specify both --tags and --instances flags\n")
			logging.LogError("Both tags and instances flags provided - only one is allowed")
			os.Exit(1)
		}

		// Validate parallel value
		if parallelFlag <= 0 {
			colors.PrintError("✗ --parallel must be greater than 0\n")
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
			instanceIDs, err = getInstanceIDsByTags(ctx, awsClient, tagsFlag, region)
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

		// Execute power operations in parallel
		startTime := time.Now()
		results := executePowerOperationParallel(ctx, awsClient, instanceIDs, region, "start", parallelFlag)
		totalDuration := time.Since(startTime)

		// Process and display results
		displayPowerOperationResults(results, "start", totalDuration, parallelFlag)
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

		// Validate that we have either tags or instances specified
		if tagsFlag == "" && instancesFlag == "" {
			colors.PrintError("✗ Either --tags or --instances flag is required\n")
			logging.LogError("No tags or instances specified for stop-tagged command")
			os.Exit(1)
		}

		// Validate mutual exclusion - cannot specify both tags and instances
		if tagsFlag != "" && instancesFlag != "" {
			colors.PrintError("✗ Cannot specify both --tags and --instances flags\n")
			logging.LogError("Both tags and instances flags provided - only one is allowed")
			os.Exit(1)
		}

		// Validate parallel value
		if parallelFlag <= 0 {
			colors.PrintError("✗ --parallel must be greater than 0\n")
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
			instanceIDs, err = getInstanceIDsByTags(ctx, awsClient, tagsFlag, region)
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

		// Execute power operations in parallel
		startTime := time.Now()
		results := executePowerOperationParallel(ctx, awsClient, instanceIDs, region, "stop", parallelFlag)
		totalDuration := time.Since(startTime)

		// Process and display results
		displayPowerOperationResults(results, "stop", totalDuration, parallelFlag)
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

		// Validate that we have either tags or instances specified
		if tagsFlag == "" && instancesFlag == "" {
			colors.PrintError("✗ Either --tags or --instances flag is required\n")
			logging.LogError("No tags or instances specified for reboot-tagged command")
			os.Exit(1)
		}

		// Validate mutual exclusion - cannot specify both tags and instances
		if tagsFlag != "" && instancesFlag != "" {
			colors.PrintError("✗ Cannot specify both --tags and --instances flags\n")
			logging.LogError("Both tags and instances flags provided - only one is allowed")
			os.Exit(1)
		}

		// Validate parallel value
		if parallelFlag <= 0 {
			colors.PrintError("✗ --parallel must be greater than 0\n")
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
			instanceIDs, err = getInstanceIDsByTags(ctx, awsClient, tagsFlag, region)
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

		// Execute power operations in parallel
		startTime := time.Now()
		results := executePowerOperationParallel(ctx, awsClient, instanceIDs, region, "reboot", parallelFlag)
		totalDuration := time.Since(startTime)

		// Process and display results
		displayPowerOperationResults(results, "reboot", totalDuration, parallelFlag)
	},
}

// PowerOperationResult represents the result of a power operation on an instance
type PowerOperationResult struct {
	InstanceID string
	Operation  string
	Error      error
	Duration   time.Duration
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

// resolveInstanceID converts instance name to instance ID if needed
func resolveInstanceID(ctx context.Context, awsClient *aws.Client, instanceIdentifier, region string) (string, error) {
	// If it's already an instance ID (starts with i-), return as is
	if strings.HasPrefix(instanceIdentifier, "i-") {
		return instanceIdentifier, nil
	}

	// Otherwise, treat it as a name and resolve to ID
	result, err := awsClient.EC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   &[]string{"tag:Name"}[0],
				Values: []string{instanceIdentifier},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe instances: %v", err)
	}

	if len(result.Reservations) == 0 {
		return "", fmt.Errorf("no instance found with name: %s", instanceIdentifier)
	}

	if len(result.Reservations) > 1 || len(result.Reservations[0].Instances) > 1 {
		return "", fmt.Errorf("multiple instances found with name: %s", instanceIdentifier)
	}

	return *result.Reservations[0].Instances[0].InstanceId, nil
}

// getInstanceIDsByTags finds instance IDs by tag filters
func getInstanceIDsByTags(ctx context.Context, awsClient *aws.Client, tagsFlag, region string) ([]string, error) {
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
					Name:   &[]string{fmt.Sprintf("tag:%s", key)}[0],
					Values: []string{value},
				})
			}
		}
	}

	result, err := awsClient.EC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %v", err)
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
func executePowerOperationParallel(ctx context.Context, awsClient *aws.Client, instanceIDs []string, region, operation string, maxParallel int) []PowerOperationResult {
	// Create channels for work distribution and result collection
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

				var err error
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
	var results []PowerOperationResult
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// displayPowerOperationResults displays the results of power operations
func displayPowerOperationResults(results []PowerOperationResult, operation string, totalDuration time.Duration, maxParallel int) {
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
		os.Exit(1)
	} else {
		logging.LogSuccess("All %s operations completed successfully", operation)
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
