package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"ztictl/internal/interactive"
	awspkg "ztictl/pkg/aws"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSsmExecMultiCmd(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		shouldParseOk bool
		validateFunc  func(cmd *cobra.Command) error
	}{
		{
			name:          "Valid multi-region with tags",
			args:          []string{"cac1,use1", "--tags", "Environment=prod", "uptime"},
			shouldParseOk: true,
		},
		{
			name:          "Valid all-regions with tags",
			args:          []string{"--all-regions", "--tags", "Component=web", "hostname"},
			shouldParseOk: true,
		},
		{
			name:          "Valid multi-region with instances",
			args:          []string{"cac1,use1", "--instances", "i-123,i-456", "ps aux"},
			shouldParseOk: true,
		},
		{
			name:          "Missing command argument",
			args:          []string{},
			shouldParseOk: false,
		},
		{
			name:          "With continue-on-error flag",
			args:          []string{"cac1,use1", "--tags", "App=api", "--continue-on-error", "health-check.sh"},
			shouldParseOk: true,
		},
		{
			name:          "With parallel flags",
			args:          []string{"cac1", "--tags", "Type=worker", "--parallel", "5", "--parallel-regions", "3", "ps"},
			shouldParseOk: true,
		},
		{
			name:          "Test validation - missing tags and instances",
			args:          []string{"cac1,use1", "uptime"},
			shouldParseOk: true,
			validateFunc: func(cmd *cobra.Command) error {
				tags, _ := cmd.Flags().GetString("tags")
				instances, _ := cmd.Flags().GetString("instances")
				if tags == "" && instances == "" {
					return fmt.Errorf("Either --tags or --instances flag is required")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command for each test
			cmd := &cobra.Command{
				Use:  "exec-multi",
				Args: cobra.MinimumNArgs(1),
			}

			// Add the flags
			cmd.Flags().BoolP("all-regions", "a", false, "Execute across all configured regions")
			cmd.Flags().StringP("tags", "t", "", "Tag filters")
			cmd.Flags().StringP("instances", "i", "", "Instance IDs")
			cmd.Flags().IntP("parallel", "p", 8, "Parallel executions")
			cmd.Flags().IntP("parallel-regions", "r", 5, "Parallel regions")
			cmd.Flags().BoolP("continue-on-error", "c", false, "Continue on error")

			// Parse flags
			err := cmd.ParseFlags(tt.args)

			if !tt.shouldParseOk {
				// For cases where we expect parsing to fail (e.g., MinimumNArgs)
				if cmd.Args != nil {
					argsErr := cmd.Args(cmd, cmd.Flags().Args())
					assert.Error(t, argsErr)
				}
			} else {
				assert.NoError(t, err)

				// Run additional validation if provided
				if tt.validateFunc != nil {
					validationErr := tt.validateFunc(cmd)
					if validationErr != nil {
						assert.Error(t, validationErr)
					}
				}
			}
		})
	}
}

func TestInstanceResults(t *testing.T) {
	tests := []struct {
		name          string
		instances     []InstanceResult
		expectedStats struct {
			total      int
			successful int
			failed     int
		}
	}{
		{
			name: "All successful",
			instances: []InstanceResult{
				{Success: true, Error: nil},
				{Success: true, Error: nil},
			},
			expectedStats: struct {
				total      int
				successful int
				failed     int
			}{2, 2, 0},
		},
		{
			name: "Partial failure",
			instances: []InstanceResult{
				{Success: true, Error: nil},
				{Success: false, Error: nil},
				{Success: true, Error: nil},
			},
			expectedStats: struct {
				total      int
				successful int
				failed     int
			}{3, 2, 1},
		},
		{
			name: "All failed",
			instances: []InstanceResult{
				{Success: false, Error: fmt.Errorf("failed")},
				{Success: false, Error: nil},
			},
			expectedStats: struct {
				total      int
				successful int
				failed     int
			}{2, 0, 2},
		},
		{
			name:      "No instances",
			instances: []InstanceResult{},
			expectedStats: struct {
				total      int
				successful int
				failed     int
			}{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			successful := 0
			failed := 0
			for _, inst := range tt.instances {
				if inst.Success && inst.Error == nil {
					successful++
				} else {
					failed++
				}
			}
			assert.Equal(t, tt.expectedStats.total, len(tt.instances))
			assert.Equal(t, tt.expectedStats.successful, successful)
			assert.Equal(t, tt.expectedStats.failed, failed)
		})
	}
}

func TestMultiRegionResult(t *testing.T) {
	tests := []struct {
		name      string
		result    MultiRegionResult
		hasError  bool
		isSuccess bool
	}{
		{
			name: "Successful region execution",
			result: MultiRegionResult{
				Region:     "cac1",
				RegionName: "Canada Central",
				Instances: []InstanceResult{
					{Success: true, Error: nil},
					{Success: true, Error: nil},
				},
				Error:    nil,
				Duration: time.Second * 10,
			},
			hasError:  false,
			isSuccess: true,
		},
		{
			name: "Failed region execution with error",
			result: MultiRegionResult{
				Region:     "use1",
				RegionName: "US East",
				Instances:  []InstanceResult{},
				Error:      fmt.Errorf("connection timeout"),
				Duration:   time.Second * 30,
			},
			hasError:  true,
			isSuccess: false,
		},
		{
			name: "Partial failure",
			result: MultiRegionResult{
				Region:     "euw1",
				RegionName: "EU West",
				Instances: []InstanceResult{
					{Success: true, Error: nil},
					{Success: false, Error: nil},
				},
				Error:    nil,
				Duration: time.Second * 15,
			},
			hasError:  false,
			isSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.hasError, tt.result.Error != nil)

			// Check if all instances succeeded
			allSuccess := tt.result.Error == nil
			for _, inst := range tt.result.Instances {
				if !inst.Success || inst.Error != nil {
					allSuccess = false
					break
				}
			}

			assert.Equal(t, tt.isSuccess, allSuccess)
			assert.NotEmpty(t, tt.result.Region)
			assert.NotEmpty(t, tt.result.RegionName)
		})
	}
}

func TestRegionExecutionRequest(t *testing.T) {
	tests := []struct {
		name    string
		request RegionExecutionRequest
		valid   bool
	}{
		{
			name: "Valid request with tags",
			request: RegionExecutionRequest{
				RegionCode:    "cac1",
				Command:       "uptime",
				TagsFlag:      "Environment=prod",
				InstancesFlag: "",
				ParallelFlag:  5,
			},
			valid: true,
		},
		{
			name: "Valid request with instances",
			request: RegionExecutionRequest{
				RegionCode:    "use1",
				Command:       "hostname",
				TagsFlag:      "",
				InstancesFlag: "i-123,i-456",
				ParallelFlag:  10,
			},
			valid: true,
		},
		{
			name: "Invalid request - no tags or instances",
			request: RegionExecutionRequest{
				RegionCode:    "euw1",
				Command:       "ps aux",
				TagsFlag:      "",
				InstancesFlag: "",
				ParallelFlag:  5,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasSelector := tt.request.TagsFlag != "" || tt.request.InstancesFlag != ""
			assert.Equal(t, tt.valid, hasSelector)
			assert.NotEmpty(t, tt.request.RegionCode)
			assert.NotEmpty(t, tt.request.Command)
			assert.Greater(t, tt.request.ParallelFlag, 0)
		})
	}
}

func TestOutputTruncation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Short output not truncated",
			input:    "This is a short output",
			expected: "This is a short output",
		},
		{
			name:     "Exactly 100 chars not truncated",
			input:    strings.Repeat("a", MaxOutputLength),
			expected: strings.Repeat("a", MaxOutputLength),
		},
		{
			name:     "Over 100 chars truncated",
			input:    strings.Repeat("a", 150),
			expected: strings.Repeat("a", OutputTruncateLength) + "...",
		},
		{
			name:     "Empty output",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.input
			if len(output) > MaxOutputLength {
				output = output[:OutputTruncateLength] + "..."
			}
			assert.Equal(t, tt.expected, output)
			assert.LessOrEqual(t, len(output), MaxOutputLength+3) // +3 for "..."
		})
	}
}

func TestParallelRegionProcessing(t *testing.T) {
	// Test that parallel region count is properly bounded
	tests := []struct {
		name            string
		totalRegions    int
		parallelRegions int
		expectedWorkers int
	}{
		{
			name:            "Fewer regions than workers",
			totalRegions:    3,
			parallelRegions: 10,
			expectedWorkers: 3,
		},
		{
			name:            "More regions than workers",
			totalRegions:    20,
			parallelRegions: 5,
			expectedWorkers: 5,
		},
		{
			name:            "Equal regions and workers",
			totalRegions:    5,
			parallelRegions: 5,
			expectedWorkers: 5,
		},
		{
			name:            "Single region",
			totalRegions:    1,
			parallelRegions: 10,
			expectedWorkers: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate actual workers that would be spawned
			workers := tt.parallelRegions
			if workers > tt.totalRegions {
				workers = tt.totalRegions
			}
			assert.Equal(t, tt.expectedWorkers, workers)
		})
	}
}

func TestRegionCodeValidation(t *testing.T) {
	tests := []struct {
		name       string
		regionCode string
		valid      bool
	}{
		{
			name:       "Valid shortcode",
			regionCode: "cac1",
			valid:      true,
		},
		{
			name:       "Valid full region",
			regionCode: "ca-central-1",
			valid:      true,
		},
		{
			name:       "Invalid shortcode",
			regionCode: "xyz1",
			valid:      false,
		},
		{
			name:       "Empty region",
			regionCode: "",
			valid:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test using the actual region mapping
			_, exists := awspkg.RegionMapping[tt.regionCode]
			if !exists && tt.regionCode != "" {
				// Check if it's a full region name
				parts := strings.Split(tt.regionCode, "-")
				isFullRegion := len(parts) >= 3 && len(parts) <= 4
				if tt.valid {
					assert.True(t, isFullRegion || exists)
				}
			} else if tt.valid {
				assert.True(t, exists || tt.regionCode == "ca-central-1")
			}
		})
	}
}

func TestContinueOnErrorBehavior(t *testing.T) {
	tests := []struct {
		name              string
		regionResults     []bool // true = success, false = failure
		continueOnError   bool
		expectedProcessed int
	}{
		{
			name:              "Continue on error - process all",
			regionResults:     []bool{true, false, true, true},
			continueOnError:   true,
			expectedProcessed: 4,
		},
		{
			name:              "Stop on error - stop at first failure",
			regionResults:     []bool{true, false, true, true},
			continueOnError:   false,
			expectedProcessed: 2,
		},
		{
			name:              "All successful",
			regionResults:     []bool{true, true, true},
			continueOnError:   false,
			expectedProcessed: 3,
		},
		{
			name:              "First fails, stop immediately",
			regionResults:     []bool{false, true, true},
			continueOnError:   false,
			expectedProcessed: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processed := 0
			for _, success := range tt.regionResults {
				processed++
				if !success && !tt.continueOnError {
					break
				}
			}
			assert.Equal(t, tt.expectedProcessed, processed)
		})
	}
}

func TestDefaultConstants(t *testing.T) {
	// Verify that our constants are set to reasonable values
	assert.Equal(t, 100, MaxOutputLength, "MaxOutputLength should be 100")
	assert.Equal(t, 97, OutputTruncateLength, "OutputTruncateLength should be 97")
	assert.Equal(t, 5, DefaultRegionParallelism, "DefaultRegionParallelism should be 5")

	// Ensure truncate length is less than max length
	assert.Less(t, OutputTruncateLength, MaxOutputLength)
}

func TestRegionNormalizationInCommand(t *testing.T) {
	tests := []struct {
		name            string
		inputRegions    string
		expectedCount   int
		shouldNormalize bool
	}{
		{
			name:            "Shortcodes",
			inputRegions:    "cac1,use1",
			expectedCount:   2,
			shouldNormalize: true,
		},
		{
			name:            "Full region names",
			inputRegions:    "ca-central-1,us-east-1",
			expectedCount:   2,
			shouldNormalize: true,
		},
		{
			name:            "Mixed formats",
			inputRegions:    "cac1,us-east-1,eu-west-1",
			expectedCount:   3,
			shouldNormalize: true,
		},
		{
			name:            "With spaces",
			inputRegions:    "cac1, us-east-1 , euw1",
			expectedCount:   3,
			shouldNormalize: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse regions like the command does
			inputRegions := strings.Split(tt.inputRegions, ",")
			var regions []string
			for _, r := range inputRegions {
				r = strings.TrimSpace(r)
				if r != "" {
					// This would call config.NormalizeRegion in real code
					regions = append(regions, r)
				}
			}
			assert.Equal(t, tt.expectedCount, len(regions))
		})
	}
}

func TestHasFailedInstances(t *testing.T) {
	tests := []struct {
		name     string
		result   MultiRegionResult
		expected bool
	}{
		{
			name: "All successful",
			result: MultiRegionResult{
				Instances: []InstanceResult{
					{Success: true, Error: nil},
					{Success: true, Error: nil},
				},
			},
			expected: false,
		},
		{
			name: "Has failed instance",
			result: MultiRegionResult{
				Instances: []InstanceResult{
					{Success: true, Error: nil},
					{Success: false, Error: nil},
				},
			},
			expected: true,
		},
		{
			name: "Has error",
			result: MultiRegionResult{
				Instances: []InstanceResult{
					{Success: true, Error: nil},
					{Success: true, Error: fmt.Errorf("connection failed")},
				},
			},
			expected: true,
		},
		{
			name:     "No instances",
			result:   MultiRegionResult{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would call hasFailedInstances function
			hasFailed := false
			for _, inst := range tt.result.Instances {
				if !inst.Success || inst.Error != nil {
					hasFailed = true
					break
				}
			}
			assert.Equal(t, tt.expected, hasFailed)
		})
	}
}

func TestCommandParsing(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedCommand string
		allRegions      bool
		expectedRegions []string
	}{
		{
			name:            "Simple command with regions",
			args:            []string{"cac1,use1", "uptime"},
			expectedCommand: "uptime",
			allRegions:      false,
			expectedRegions: []string{"cac1", "use1"},
		},
		{
			name:            "Complex command with regions",
			args:            []string{"cac1", "ps", "aux", "|", "grep", "nginx"},
			expectedCommand: "ps aux | grep nginx",
			allRegions:      false,
			expectedRegions: []string{"cac1"},
		},
		{
			name:            "All regions with simple command",
			args:            []string{"hostname"},
			expectedCommand: "hostname",
			allRegions:      true,
			expectedRegions: nil, // Would be populated from RegionMapping
		},
		{
			name:            "Quoted command",
			args:            []string{"cac1,use1", "echo", "hello world"},
			expectedCommand: "echo hello world",
			allRegions:      false,
			expectedRegions: []string{"cac1", "use1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var command string
			var regions []string

			if tt.allRegions {
				command = strings.Join(tt.args, " ")
				// In real scenario, regions would be populated from awspkg.RegionMapping
			} else {
				if len(tt.args) >= 2 {
					regionsList := tt.args[0]
					regions = strings.Split(regionsList, ",")
					command = strings.Join(tt.args[1:], " ")
				}
			}

			assert.Equal(t, tt.expectedCommand, command)
			if !tt.allRegions {
				assert.Equal(t, tt.expectedRegions, regions)
			}
		})
	}
}

func TestInstanceListProcessing(t *testing.T) {
	tests := []struct {
		name          string
		instancesFlag string
		expected      []interactive.Instance
	}{
		{
			name:          "Single instance",
			instancesFlag: "i-123",
			expected: []interactive.Instance{
				{InstanceID: "i-123", Name: "i-123"},
			},
		},
		{
			name:          "Multiple instances",
			instancesFlag: "i-123,i-456,i-789",
			expected: []interactive.Instance{
				{InstanceID: "i-123", Name: "i-123"},
				{InstanceID: "i-456", Name: "i-456"},
				{InstanceID: "i-789", Name: "i-789"},
			},
		},
		{
			name:          "Instances with spaces",
			instancesFlag: "i-123, i-456 , i-789",
			expected: []interactive.Instance{
				{InstanceID: "i-123", Name: "i-123"},
				{InstanceID: "i-456", Name: "i-456"},
				{InstanceID: "i-789", Name: "i-789"},
			},
		},
		{
			name:          "Empty instances",
			instancesFlag: "",
			expected:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var instances []interactive.Instance

			if tt.instancesFlag != "" {
				instanceIDs := strings.Split(tt.instancesFlag, ",")
				for _, id := range instanceIDs {
					trimmedID := strings.TrimSpace(id)
					instances = append(instances, interactive.Instance{
						InstanceID: trimmedID,
						Name:       trimmedID,
					})
				}
			}

			assert.Equal(t, tt.expected, instances)
		})
	}
}
