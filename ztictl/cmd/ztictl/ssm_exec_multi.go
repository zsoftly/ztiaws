package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"ztictl/internal/ssm"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
)

// RegionResult holds results for a single region
type RegionResult struct {
	Region    string
	Instances []InstanceResult
	Error     error
}

// InstanceResult holds results for a single instance
type InstanceResult struct {
	InstanceID  string
	Name        string
	Output      string
	ErrorOutput string
	ExitCode    *int32
	Error       error
}

// ssmExecMultiCmd represents the exec-multi command for multi-region execution
var ssmExecMultiCmd = &cobra.Command{
	Use:   "exec-multi [flags] <command>",
	Short: "Execute commands across multiple regions",
	Long: `Execute commands on instances across multiple AWS regions based on tag filters.

Examples:
  # Execute on web servers in multiple regions
  ztictl ssm exec-multi --regions cac1,use1 --tags "Role=web" "systemctl status nginx"
  
  # Execute across all configured regions
  ztictl ssm exec-multi --all-regions --tags "Environment=prod" "df -h"
  
  # With custom concurrency and timeout
  ztictl ssm exec-multi --regions cac1,use1 --tags "Role=web" --concurrent 10 --timeout 60s "uptime"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()
		command := args[0]

		// Parse flags
		regions, _ := cmd.Flags().GetStringSlice("regions")
		allRegions, _ := cmd.Flags().GetBool("all-regions")
		tagsFlag, _ := cmd.Flags().GetString("tags")
		concurrent, _ := cmd.Flags().GetInt("concurrent")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		// Validate and parse regions
		regionList, err := parseRegions(regions, allRegions)
		if err != nil {
			colors.PrintError("Region parsing error: %v\n", err)
			os.Exit(1)
		}

		// Parse tags
		tags, err := parseTags(tagsFlag)
		if err != nil {
			colors.PrintError("Tag parsing error: %v\n", err)
			os.Exit(1)
		}

		if len(tags) == 0 {
			colors.PrintError("At least one tag filter is required\n")
			os.Exit(1)
		}

		logging.LogInfo("Starting multi-region execution across %d regions", len(regionList))
		logging.LogInfo("Command: %s", command)
		logging.LogInfo("Regions: %s", strings.Join(regionList, ", "))
		logging.LogInfo("Tags: %s", tagsFlag)

		// Execute across regions
		ctx := context.Background()
		results := executeMultiRegion(ctx, regionList, tags, command, concurrent, timeout)

		// Display results
		displayMultiRegionResults(results, command, startTime)

		// Exit with error if any region failed
		hasFailures := false
		for _, result := range results {
			if result.Error != nil {
				hasFailures = true
				break
			}
			for _, instance := range result.Instances {
				if instance.Error != nil || (instance.ExitCode != nil && *instance.ExitCode != 0) {
					hasFailures = true
					break
				}
			}
		}

		if hasFailures {
			os.Exit(1)
		}
	},
}

// parseRegions converts region input to full region names
func parseRegions(regions []string, allRegions bool) ([]string, error) {
	if allRegions {
		// Define all configured regions - you might want to make this configurable
		// Based on the existing bash script regions
		return []string{"ca-central-1", "ca-west-1", "us-east-1", "us-west-1", "eu-west-1", "ap-southeast-1"}, nil
	}

	if len(regions) == 0 {
		return nil, fmt.Errorf("must specify either --regions or --all-regions")
	}

	var result []string
	for _, region := range regions {
		fullRegion := resolveRegion(region) // Use existing function
		result = append(result, fullRegion)
	}

	return removeDuplicates(result), nil
}

// parseTags parses tag string in format "Key1=Value1,Key2=Value2"
func parseTags(tagsFlag string) (map[string]string, error) {
	if tagsFlag == "" {
		return nil, fmt.Errorf("tags flag is required")
	}

	tags := make(map[string]string)
	pairs := strings.Split(tagsFlag, ",")

	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tag format '%s', expected 'Key=Value'", pair)
		}
		tags[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	return tags, nil
}

// removeDuplicates removes duplicate regions
func removeDuplicates(regions []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, region := range regions {
		if !seen[region] {
			seen[region] = true
			result = append(result, region)
		}
	}
	return result
}

// executeMultiRegion executes command across multiple regions concurrently
func executeMultiRegion(ctx context.Context, regions []string, tags map[string]string, command string, concurrent int, timeout time.Duration) []RegionResult {
	// Create channels for coordinating goroutines
	resultChan := make(chan RegionResult, len(regions))

	// Use a semaphore to limit concurrency
	sem := make(chan struct{}, concurrent)

	var wg sync.WaitGroup

	// Launch goroutines for each region
	for _, region := range regions {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Create context with timeout for this region
			regionCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			result := executeInRegion(regionCtx, region, tags, command)
			resultChan <- result
		}(region)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results in order
	var results []RegionResult
	regionOrder := make(map[string]int)
	for i, region := range regions {
		regionOrder[region] = i
	}

	tempResults := make(map[string]RegionResult)
	for result := range resultChan {
		tempResults[result.Region] = result
	}

	// Sort results by original region order
	for _, region := range regions {
		if result, exists := tempResults[region]; exists {
			results = append(results, result)
		}
	}

	return results
}

// executeInRegion executes command in a single region
func executeInRegion(ctx context.Context, region string, tags map[string]string, command string) RegionResult {
	result := RegionResult{
		Region: region,
	}

	ssmManager := ssm.NewManager(logger)

	// Find instances matching tags in this region
	var tagFilter string
	var tagPairs []string
	for key, value := range tags {
		tagPairs = append(tagPairs, fmt.Sprintf("%s=%s", key, value))
	}
	tagFilter = strings.Join(tagPairs, ",")

	filters := &ssm.ListFilters{
		Tag: tagFilter,
	}

	instances, err := ssmManager.ListInstances(ctx, region, filters)
	if err != nil {
		result.Error = fmt.Errorf("failed to list instances in %s: %w", region, err)
		return result
	}

	if len(instances) == 0 {
		result.Error = fmt.Errorf("no instances found matching tags in %s", region)
		return result
	}

	logging.LogInfo("Found %d instances in region %s", len(instances), region)

	// Execute command on each instance
	for _, instance := range instances {
		instanceResult := InstanceResult{
			InstanceID: instance.InstanceID,
			Name:       instance.Name,
		}

		cmdResult, err := ssmManager.ExecuteCommand(ctx, instance.InstanceID, region, command, "")
		if err != nil {
			instanceResult.Error = err
		} else {
			instanceResult.Output = cmdResult.Output
			instanceResult.ErrorOutput = cmdResult.ErrorOutput
			instanceResult.ExitCode = cmdResult.ExitCode
		}

		result.Instances = append(result.Instances, instanceResult)
	}

	return result
}

// displayMultiRegionResults displays the execution results
func displayMultiRegionResults(results []RegionResult, command string, startTime time.Time) {
	colors.PrintHeader("=== Multi-Region Command Execution ===\n")
	fmt.Printf("Command: %s\n", command)
	fmt.Printf("Started: %s\n", startTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Duration: %s\n", time.Since(startTime).Round(time.Millisecond))
	fmt.Println()

	successfulRegions := 0
	totalInstances := 0
	successfulInstances := 0

	for _, result := range results {
		colors.PrintHeader("--- %s ---\n", strings.ToUpper(result.Region))

		if result.Error != nil {
			colors.PrintError("❌ Region Error: %s\n", result.Error.Error())
			fmt.Println()
			continue
		}

		if len(result.Instances) == 0 {
			colors.PrintData("ℹ️  No instances found with matching tags\n")
			fmt.Println()
			continue
		}

		successfulRegions++
		totalInstances += len(result.Instances)

		for _, instance := range result.Instances {
			instanceName := instance.Name
			if instanceName == "" {
				instanceName = instance.InstanceID
			}

			if instance.Error != nil {
				colors.PrintError("❌ %s (%s): %s\n", instanceName, instance.InstanceID, instance.Error.Error())
			} else if instance.ExitCode != nil && *instance.ExitCode != 0 {
				colors.PrintError("❌ %s (%s): Command failed (exit code: %d)\n", instanceName, instance.InstanceID, int(*instance.ExitCode))
				if instance.ErrorOutput != "" {
					colors.PrintData("   Error: %s\n", strings.TrimSpace(instance.ErrorOutput))
				}
			} else {
				colors.PrintSuccess("✅ %s (%s):\n", instanceName, instance.InstanceID)
				successfulInstances++

				// Indent the output
				output := strings.TrimSpace(instance.Output)
				if output != "" {
					for _, line := range strings.Split(output, "\n") {
						colors.PrintData("   %s\n", line)
					}
				} else {
					colors.PrintData("   (no output)\n")
				}
			}
		}
		fmt.Println()
	}

	// Summary
	colors.PrintHeader("=== Summary ===\n")
	fmt.Printf("Total regions processed: %d\n", len(results))
	fmt.Printf("Successful regions: %d\n", successfulRegions)
	fmt.Printf("Total instances: %d\n", totalInstances)
	fmt.Printf("Successful instances: %d\n", successfulInstances)

	if successfulInstances < totalInstances {
		fmt.Printf("Failed instances: %d\n", totalInstances-successfulInstances)
	}
}

func init() {
	// Add flags
	ssmExecMultiCmd.Flags().StringSlice("regions", []string{}, "Comma-separated list of regions (e.g., cac1,use1)")
	ssmExecMultiCmd.Flags().Bool("all-regions", false, "Execute across all configured regions")
	ssmExecMultiCmd.Flags().String("tags", "", "Tag filters as key=value pairs (e.g., 'Role=web,Environment=prod')")
	ssmExecMultiCmd.Flags().Int("concurrent", 5, "Number of concurrent region operations")
	ssmExecMultiCmd.Flags().Duration("timeout", 30*time.Second, "Timeout per region operation")

	// Mark required flags
	ssmExecMultiCmd.MarkFlagRequired("tags")
	ssmExecMultiCmd.MarkFlagsMutuallyExclusive("regions", "all-regions")
}
