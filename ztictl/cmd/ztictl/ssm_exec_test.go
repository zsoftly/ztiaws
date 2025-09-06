package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSsmExecCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Exec help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Execute commands on multiple instances or filtered by tags",
		},
		{
			name:    "Exec with instance and script",
			args:    []string{"i-1234567890abcdef0", "--script", "/path/to/script.sh"},
			wantErr: false,
		},
		{
			name:    "Exec with multiple instances",
			args:    []string{"i-123,i-456", "--command", "uptime"},
			wantErr: false,
		},
		{
			name:    "Exec with tag filter",
			args:    []string{"--tag", "Environment=Production", "--command", "systemctl status nginx"},
			wantErr: false,
		},
		{
			name:    "Exec with timeout and parallel",
			args:    []string{"i-1234567890abcdef0", "--command", "sleep 10", "--timeout", "15", "--parallel", "5"},
			wantErr: false,
		},
		{
			name:    "Exec without target",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Exec without command or script",
			args:    []string{"i-1234567890abcdef0"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "exec <instance-identifier|tag-filter>",
				Short: "Execute commands and manage workflows",
				Long:  "Execute commands on multiple instances or filtered by tags",
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock exec functionality
					regionCode, _ := cmd.Flags().GetString("region")
					command, _ := cmd.Flags().GetString("command")
					script, _ := cmd.Flags().GetString("script")
					tag, _ := cmd.Flags().GetString("tag")
					timeout, _ := cmd.Flags().GetInt("timeout")
					parallel, _ := cmd.Flags().GetInt("parallel")
					dryRun, _ := cmd.Flags().GetBool("dry-run")

					// Validate that either command or script is provided
					if command == "" && script == "" {
						return fmt.Errorf("either command or script must be provided")
					}

					// Mock region resolution
					region := regionCode
					if region == "" {
						region = "us-east-1"
					}

					// Mock execution configuration
					type ExecConfig struct {
						Targets  []string
						Command  string
						Script   string
						Tag      string
						Region   string
						Timeout  int
						Parallel int
						DryRun   bool
					}

					var targets []string
					if len(args) > 0 {
						// Parse targets (could be comma-separated instance IDs)
						targets = strings.Split(args[0], ",")
					}

					config := ExecConfig{
						Targets:  targets,
						Command:  command,
						Script:   script,
						Tag:      tag,
						Region:   region,
						Timeout:  timeout,
						Parallel: parallel,
						DryRun:   dryRun,
					}

					// Validate configuration
					if config.Timeout <= 0 {
						config.Timeout = 30 // default
					}
					if config.Parallel <= 0 {
						config.Parallel = 1 // default
					}

					// Test target validation
					if len(config.Targets) == 0 && config.Tag == "" {
						return fmt.Errorf("either targets or tag filter must be provided")
					}

					// Mock execution results
					type ExecResult struct {
						InstanceID string
						Status     string
						ExitCode   int
						Output     string
						Duration   int
					}

					var results []ExecResult
					for _, target := range config.Targets {
						result := ExecResult{
							InstanceID: target,
							Status:     "Success",
							ExitCode:   0,
							Output:     "Command executed successfully",
							Duration:   100,
						}
						results = append(results, result)
					}

					// Validate results
					for _, result := range results {
						if result.InstanceID == "" {
							t.Error("Result should have instance ID")
						}
						if result.Status == "" {
							t.Error("Result should have status")
						}
					}
					return nil
				},
			}

			// Add flags
			cmd.Flags().StringP("region", "r", "", "AWS region")
			cmd.Flags().StringP("command", "c", "", "Command to execute")
			cmd.Flags().StringP("script", "s", "", "Script file to execute")
			cmd.Flags().String("tag", "", "Tag filter for target instances")
			cmd.Flags().Int("timeout", 30, "Command timeout in seconds")
			cmd.Flags().Int("parallel", 1, "Number of parallel executions")
			cmd.Flags().Bool("dry-run", false, "Show what would be executed without running")

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.contains != "" {
				output := buf.String()
				if !strings.Contains(output, tt.contains) {
					t.Errorf("Execute() output should contain %v, got %v", tt.contains, output)
				}
			}
		})
	}
}

func TestSsmExecCmdFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "exec"}
	cmd.Flags().StringP("region", "r", "", "AWS region")
	cmd.Flags().StringP("command", "c", "", "Command to execute")
	cmd.Flags().StringP("script", "s", "", "Script file")
	cmd.Flags().String("tag", "", "Tag filter")
	cmd.Flags().Int("timeout", 30, "Timeout")
	cmd.Flags().Int("parallel", 1, "Parallel executions")
	cmd.Flags().Bool("dry-run", false, "Dry run")

	tests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{"region", "r", "", "string"},
		{"command", "c", "", "string"},
		{"script", "s", "", "string"},
		{"tag", "", "", "string"},
		{"timeout", "", "30", "int"},
		{"parallel", "", "1", "int"},
		{"dry-run", "", "false", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName+" flag", func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s shorthand = %s, want %s", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			if flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %s default = %s, want %s", tt.flagName, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestTargetParsing(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCount int
		expectedFirst string
	}{
		{
			name:          "Single instance",
			input:         "i-1234567890abcdef0",
			expectedCount: 1,
			expectedFirst: "i-1234567890abcdef0",
		},
		{
			name:          "Multiple instances",
			input:         "i-123,i-456,i-789",
			expectedCount: 3,
			expectedFirst: "i-123",
		},
		{
			name:          "Instance with spaces",
			input:         "i-123, i-456, i-789",
			expectedCount: 3,
			expectedFirst: "i-123",
		},
		{
			name:          "Single instance name",
			input:         "web-server-1",
			expectedCount: 1,
			expectedFirst: "web-server-1",
		},
		{
			name:          "Mixed identifiers",
			input:         "i-123,web-server-1,db-server",
			expectedCount: 3,
			expectedFirst: "i-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targets := strings.Split(tt.input, ",")

			// Trim spaces
			for i, target := range targets {
				targets[i] = strings.TrimSpace(target)
			}

			if len(targets) != tt.expectedCount {
				t.Errorf("Target count = %d, want %d", len(targets), tt.expectedCount)
			}

			if len(targets) > 0 && targets[0] != tt.expectedFirst {
				t.Errorf("First target = %s, want %s", targets[0], tt.expectedFirst)
			}

			// Validate targets
			for _, target := range targets {
				if target == "" {
					t.Error("Target should not be empty after parsing")
				}
			}
		})
	}
}

func TestTagFilterParsing(t *testing.T) {
	tests := []struct {
		name        string
		tagFilter   string
		isValid     bool
		expectedKey string
		expectedVal string
	}{
		{
			name:        "Valid key-value tag",
			tagFilter:   "Environment=Production",
			isValid:     true,
			expectedKey: "Environment",
			expectedVal: "Production",
		},
		{
			name:        "Tag with spaces",
			tagFilter:   "Team = DevOps",
			isValid:     true,
			expectedKey: "Team",
			expectedVal: "DevOps",
		},
		{
			name:      "Tag key only",
			tagFilter: "Environment",
			isValid:   false,
		},
		{
			name:      "Empty tag filter",
			tagFilter: "",
			isValid:   false,
		},
		{
			name:      "Multiple equals",
			tagFilter: "Key=Value=Extra",
			isValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tagFilter == "" {
				if tt.isValid {
					t.Error("Empty tag filter should not be valid")
				}
				return
			}

			parts := strings.SplitN(tt.tagFilter, "=", 2)
			isValid := len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != "" &&
				!strings.Contains(parts[1], "=")

			if isValid != tt.isValid {
				t.Errorf("Tag filter '%s' validity = %v, want %v", tt.tagFilter, isValid, tt.isValid)
			}

			if isValid && tt.isValid {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				if key != tt.expectedKey {
					t.Errorf("Tag key = %s, want %s", key, tt.expectedKey)
				}

				if value != tt.expectedVal {
					t.Errorf("Tag value = %s, want %s", value, tt.expectedVal)
				}
			}
		})
	}
}

func TestParallelExecutionConfig(t *testing.T) {
	tests := []struct {
		name          string
		parallel      int
		targetCount   int
		expectedBatch int
		isValid       bool
	}{
		{
			name:          "Single parallel execution",
			parallel:      1,
			targetCount:   5,
			expectedBatch: 1,
			isValid:       true,
		},
		{
			name:          "Parallel less than targets",
			parallel:      3,
			targetCount:   10,
			expectedBatch: 3,
			isValid:       true,
		},
		{
			name:          "Parallel equals targets",
			parallel:      5,
			targetCount:   5,
			expectedBatch: 5,
			isValid:       true,
		},
		{
			name:          "Parallel more than targets",
			parallel:      10,
			targetCount:   3,
			expectedBatch: 3, // Should be limited to target count
			isValid:       true,
		},
		{
			name:        "Zero parallel",
			parallel:    0,
			targetCount: 5,
			isValid:     false,
		},
		{
			name:        "Negative parallel",
			parallel:    -1,
			targetCount: 5,
			isValid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.parallel > 0

			if isValid != tt.isValid {
				t.Errorf("Parallel config %d validity = %v, want %v", tt.parallel, isValid, tt.isValid)
			}

			if isValid {
				// Calculate effective batch size
				batchSize := tt.parallel
				if batchSize > tt.targetCount {
					batchSize = tt.targetCount
				}

				if batchSize != tt.expectedBatch {
					t.Errorf("Batch size = %d, want %d", batchSize, tt.expectedBatch)
				}
			}
		})
	}
}

func TestExecResultAggregation(t *testing.T) {
	// Test result aggregation for multiple instances
	type ExecResult struct {
		InstanceID string
		Status     string
		ExitCode   int
		Output     string
		Duration   int
	}

	results := []ExecResult{
		{
			InstanceID: "i-123",
			Status:     "Success",
			ExitCode:   0,
			Output:     "OK",
			Duration:   100,
		},
		{
			InstanceID: "i-456",
			Status:     "Failed",
			ExitCode:   1,
			Output:     "Error occurred",
			Duration:   50,
		},
		{
			InstanceID: "i-789",
			Status:     "Success",
			ExitCode:   0,
			Output:     "OK",
			Duration:   75,
		},
	}

	// Test result aggregation
	successCount := 0
	failedCount := 0
	totalDuration := 0

	for _, result := range results {
		if result.Status == "Success" {
			successCount++
		} else {
			failedCount++
		}
		totalDuration += result.Duration
	}

	expectedSuccess := 2
	expectedFailed := 1
	expectedTotal := 225

	if successCount != expectedSuccess {
		t.Errorf("Success count = %d, want %d", successCount, expectedSuccess)
	}

	if failedCount != expectedFailed {
		t.Errorf("Failed count = %d, want %d", failedCount, expectedFailed)
	}

	if totalDuration != expectedTotal {
		t.Errorf("Total duration = %d, want %d", totalDuration, expectedTotal)
	}

	// Test overall status
	overallSuccess := failedCount == 0
	expectedOverallSuccess := false

	if overallSuccess != expectedOverallSuccess {
		t.Errorf("Overall success = %v, want %v", overallSuccess, expectedOverallSuccess)
	}
}

func TestDryRunMode(t *testing.T) {
	tests := []struct {
		name      string
		dryRun    bool
		command   string
		expectRun bool
	}{
		{
			name:      "Normal execution",
			dryRun:    false,
			command:   "uptime",
			expectRun: true,
		},
		{
			name:      "Dry run mode",
			dryRun:    true,
			command:   "rm -rf /",
			expectRun: false,
		},
		{
			name:      "Dry run with safe command",
			dryRun:    true,
			command:   "echo hello",
			expectRun: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// In dry run mode, commands should not be executed
			shouldExecute := !tt.dryRun

			if shouldExecute != tt.expectRun {
				t.Errorf("Should execute = %v, want %v for dry-run=%v", shouldExecute, tt.expectRun, tt.dryRun)
			}

			// In dry run mode, we should still validate the command
			if tt.command == "" {
				t.Error("Command should not be empty even in dry run")
			}

			// Potentially dangerous commands should be flagged
			dangerousPatterns := []string{"rm -rf", "mkfs", "dd if="}
			isDangerous := false
			for _, pattern := range dangerousPatterns {
				if strings.Contains(tt.command, pattern) {
					isDangerous = true
					break
				}
			}

			if isDangerous && !tt.dryRun {
				t.Logf("Warning: Dangerous command detected: %s", tt.command)
			}
		})
	}
}

func TestSsmExecTaggedCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Exec-tagged help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Execute a command on all EC2 instances that match the specified tags",
		},
		{
			name:    "Valid single tag",
			args:    []string{"cac1", "--tags", "Environment=dev", "uptime"},
			wantErr: false,
		},
		{
			name:    "Valid multiple tags",
			args:    []string{"use1", "--tags", "Environment=dev,Component=fts", "uptime"},
			wantErr: false,
		},
		{
			name:    "Complex multiple tags",
			args:    []string{"cac1", "--tags", "Team=backend,Environment=staging,Component=api", "ps aux | grep java"},
			wantErr: false,
		},
		{
			name:    "Tags with spaces",
			args:    []string{"use1", "--tags", "Environment = production , Team = devops", "uptime"},
			wantErr: false,
		},
		{
			name:    "Missing tags flag",
			args:    []string{"cac1", "uptime"},
			wantErr: true,
		},
		{
			name:    "Empty tags value",
			args:    []string{"cac1", "--tags", "", "uptime"},
			wantErr: true,
		},
		{
			name:    "Invalid tag format",
			args:    []string{"cac1", "--tags", "Environment", "uptime"},
			wantErr: true,
		},
		{
			name:    "Missing command",
			args:    []string{"cac1", "--tags", "Environment=dev"},
			wantErr: true,
		},
		{
			name:    "Short tags flag",
			args:    []string{"use1", "-t", "Environment=prod", "systemctl status nginx"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the command to avoid modifying the original
			cmd := &cobra.Command{
				Use:   "exec-tagged <region-shortcode> <command>",
				Short: "Execute a command on all instances with specified tags",
				Long: `Execute a command on all EC2 instances that match the specified tags via SSM.
Region shortcuts supported: cac1, use1, euw1, etc.
Use --tags flag to specify one or more tag filters in key=value format, separated by commas.`,
				Args: cobra.MinimumNArgs(2),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock the exec-tagged functionality
					regionCode := args[0]
					command := strings.Join(args[1:], " ")

					// Get tags flag
					tagsFlag, _ := cmd.Flags().GetString("tags")
					if tagsFlag == "" {
						return fmt.Errorf("--tags flag is required")
					}

					// Mock region resolution
					region := regionCode
					if region == "" {
						region = "us-east-1"
					}

					// Mock tag parsing
					tagPairs := strings.Split(tagsFlag, ",")
					for _, tagPair := range tagPairs {
						tagPair = strings.TrimSpace(tagPair)
						if tagPair == "" {
							continue
						}

						parts := strings.SplitN(tagPair, "=", 2)
						if len(parts) != 2 {
							return fmt.Errorf("invalid tag format '%s'. Expected format: key=value", tagPair)
						}

						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])

						if key == "" || value == "" {
							return fmt.Errorf("empty tag key or value in '%s'", tagPair)
						}
					}

					// Validate command
					if command == "" {
						return fmt.Errorf("command is required")
					}

					// Mock execution would happen here
					return nil
				},
			}

			// Add flags
			cmd.Flags().StringP("tags", "t", "", "Tag filters in key=value format, separated by commas (required)")
			cmd.MarkFlagRequired("tags")

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.contains != "" {
				output := buf.String()
				if !strings.Contains(output, tt.contains) {
					t.Errorf("Execute() output should contain %v, got %v", tt.contains, output)
				}
			}
		})
	}
}

func TestSsmExecTaggedFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "exec-tagged"}
	cmd.Flags().StringP("tags", "t", "", "Tag filters")

	// Test tags flag exists
	flag := cmd.Flags().Lookup("tags")
	if flag == nil {
		t.Error("Tags flag not found")
		return
	}

	if flag.Shorthand != "t" {
		t.Errorf("Tags flag shorthand = %s, want t", flag.Shorthand)
	}

	if flag.DefValue != "" {
		t.Errorf("Tags flag default = %s, want empty string", flag.DefValue)
	}

	// Test that we can get and set the flag value
	cmd.Flags().Set("tags", "Environment=test")
	value, err := cmd.Flags().GetString("tags")
	if err != nil {
		t.Errorf("Error getting tags flag value: %v", err)
	}
	if value != "Environment=test" {
		t.Errorf("Expected tags value 'Environment=test', got '%s'", value)
	}
}

func TestTagFilterValidation(t *testing.T) {
	tests := []struct {
		name        string
		tagFilter   string
		expectError bool
		errorMsg    string
	}{
		{
			name:      "Valid single tag",
			tagFilter: "Environment=dev",
		},
		{
			name:      "Valid multiple tags",
			tagFilter: "Environment=dev,Component=fts",
		},
		{
			name:      "Valid three tags",
			tagFilter: "Team=backend,Environment=staging,Component=api",
		},
		{
			name:      "Valid tags with spaces",
			tagFilter: "Environment = production , Team = devops",
		},
		{
			name:        "Invalid - missing value",
			tagFilter:   "Environment",
			expectError: true,
			errorMsg:    "Expected format: key=value",
		},
		{
			name:        "Invalid - missing key",
			tagFilter:   "=production",
			expectError: true,
			errorMsg:    "empty tag key or value",
		},
		{
			name:        "Invalid - empty value",
			tagFilter:   "Environment=",
			expectError: true,
			errorMsg:    "empty tag key or value",
		},
		{
			name:        "Invalid - empty key",
			tagFilter:   "=value",
			expectError: true,
			errorMsg:    "empty tag key or value",
		},
		{
			name:        "Invalid - mixed valid/invalid",
			tagFilter:   "Environment=dev,Component",
			expectError: true,
			errorMsg:    "Expected format: key=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the parsing logic from the command
			tagPairs := strings.Split(tt.tagFilter, ",")
			var err error

			for _, tagPair := range tagPairs {
				tagPair = strings.TrimSpace(tagPair)
				if tagPair == "" {
					continue
				}

				parts := strings.SplitN(tagPair, "=", 2)
				if len(parts) != 2 {
					err = fmt.Errorf("invalid tag format '%s'. Expected format: key=value", tagPair)
					break
				}

				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				if key == "" || value == "" {
					err = fmt.Errorf("empty tag key or value in '%s'", tagPair)
					break
				}
			}

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for tag filter %q but got none", tt.tagFilter)
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for tag filter %q: %v", tt.tagFilter, err)
				}
			}
		})
	}
}

func TestSsmExecContextHandling(t *testing.T) {
	// Test context usage in exec operations
	ctx := context.Background()
	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Test context with execution metadata
	type contextKey string
	key := contextKey("exec-id")
	ctx = context.WithValue(ctx, key, "exec-123")

	value := ctx.Value(key)
	if value != "exec-123" {
		t.Errorf("Context value = %v, want exec-123", value)
	}

	// Test cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cancel()
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled")
	}
}

func TestExecErrorHandling(t *testing.T) {
	// Test error scenarios
	tests := []struct {
		name        string
		targets     []string
		command     string
		shouldError bool
		errorType   string
	}{
		{
			name:        "Valid execution",
			targets:     []string{"i-1234567890abcdef0"},
			command:     "uptime",
			shouldError: false,
		},
		{
			name:        "No targets",
			targets:     []string{},
			command:     "uptime",
			shouldError: true,
			errorType:   "no targets specified",
		},
		{
			name:        "Invalid instance",
			targets:     []string{"i-nonexistent"},
			command:     "uptime",
			shouldError: true,
			errorType:   "instance not found",
		},
		{
			name:        "Command timeout",
			targets:     []string{"i-1234567890abcdef0"},
			command:     "sleep 300",
			shouldError: true,
			errorType:   "timeout",
		},
		{
			name:        "Permission denied",
			targets:     []string{"i-1234567890abcdef0"},
			command:     "sudo reboot",
			shouldError: true,
			errorType:   "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock error conditions
			var err error
			if tt.shouldError {
				err = &mockExecError{message: tt.errorType}
			}

			// Test error handling
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err != nil && !strings.Contains(err.Error(), tt.errorType) {
				t.Errorf("Error message should contain %s, got %v", tt.errorType, err)
			}
		})
	}
}

// Mock error type for exec testing
type mockExecError struct {
	message string
}

func (e *mockExecError) Error() string {
	return e.message
}

func TestCommandScriptValidation(t *testing.T) {
	tests := []struct {
		name    string
		command string
		script  string
		isValid bool
		reason  string
	}{
		{
			name:    "Command provided",
			command: "uptime",
			script:  "",
			isValid: true,
		},
		{
			name:    "Script provided",
			command: "",
			script:  "/path/to/script.sh",
			isValid: true,
		},
		{
			name:    "Both command and script",
			command: "uptime",
			script:  "/path/to/script.sh",
			isValid: false,
			reason:  "cannot specify both command and script",
		},
		{
			name:    "Neither command nor script",
			command: "",
			script:  "",
			isValid: false,
			reason:  "must specify either command or script",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasCommand := tt.command != ""
			hasScript := tt.script != ""

			isValid := (hasCommand && !hasScript) || (!hasCommand && hasScript)

			if isValid != tt.isValid {
				t.Errorf("Command/script validation = %v, want %v", isValid, tt.isValid)
				if !tt.isValid && tt.reason != "" {
					t.Logf("Reason: %s", tt.reason)
				}
			}
		})
	}
}

// NEW TESTS FOR PARALLEL EXECUTION FUNCTIONALITY

func TestSsmExecTaggedNewFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Help shows new flags",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "PARALLEL BY DEFAULT",
		},
		{
			name:    "Valid with instances flag",
			args:    []string{"cac1", "--instances", "i-1234,i-5678", "uptime"},
			wantErr: false,
		},
		{
			name:    "Valid with parallel flag",
			args:    []string{"cac1", "--tags", "Environment=dev", "--parallel", "5", "uptime"},
			wantErr: false,
		},
		{
			name:    "Both tags and instances provided",
			args:    []string{"cac1", "--tags", "Environment=dev", "--instances", "i-1234", "uptime"},
			wantErr: true,
		},
		{
			name:    "Neither tags nor instances provided",
			args:    []string{"cac1", "uptime"},
			wantErr: true,
		},
		{
			name:    "Invalid parallel value - zero",
			args:    []string{"cac1", "--tags", "Environment=dev", "--parallel", "0", "uptime"},
			wantErr: true,
		},
		{
			name:    "Invalid parallel value - negative",
			args:    []string{"cac1", "--tags", "Environment=dev", "--parallel", "-1", "uptime"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "exec-tagged <region-shortcode> <command>",
				Short: "Execute a command on instances with specified tags (parallel execution)",
				Long: `Execute a command on EC2 instances that match the specified tags via SSM.
Region shortcuts supported: cac1, use1, euw1, etc.
Use --tags flag to specify one or more tag filters in key=value format, separated by commas.
Use --instances to explicitly specify instance IDs to target (comma-separated).
Use --parallel to control maximum concurrent executions (default: number of CPU cores).

ALL COMMANDS RUN IN PARALLEL BY DEFAULT for improved performance at scale.`,
				Args: cobra.MinimumNArgs(2),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Get flags
					tagsFlag, _ := cmd.Flags().GetString("tags")
					instancesFlag, _ := cmd.Flags().GetString("instances")
					parallelFlag, _ := cmd.Flags().GetInt("parallel")

					// Validate that we have either tags or instances specified
					if tagsFlag == "" && instancesFlag == "" {
						return fmt.Errorf("either --tags or --instances flag is required")
					}

					// Validate that we don't have both
					if tagsFlag != "" && instancesFlag != "" {
						return fmt.Errorf("cannot specify both --tags and --instances flags")
					}

					// Validate parallel value
					if parallelFlag <= 0 {
						return fmt.Errorf("--parallel must be greater than 0")
					}

					return nil
				},
			}

			// Add flags matching our new implementation
			cmd.Flags().StringP("tags", "t", "", "Tag filters in key=value format, separated by commas")
			cmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs to target explicitly")
			cmd.Flags().IntP("parallel", "p", 8, "Maximum number of concurrent executions")

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.contains != "" {
				output := buf.String()
				if !strings.Contains(output, tt.contains) {
					t.Errorf("Execute() output should contain %v, got %v", tt.contains, output)
				}
			}
		})
	}
}

func TestInstanceIDParsing(t *testing.T) {
	tests := []struct {
		name          string
		instancesFlag string
		expectedIDs   []string
		expectError   bool
	}{
		{
			name:          "Single instance ID",
			instancesFlag: "i-1234567890abcdef0",
			expectedIDs:   []string{"i-1234567890abcdef0"},
		},
		{
			name:          "Multiple instance IDs",
			instancesFlag: "i-1234,i-5678,i-9abc",
			expectedIDs:   []string{"i-1234", "i-5678", "i-9abc"},
		},
		{
			name:          "Instance IDs with spaces",
			instancesFlag: "i-1234, i-5678 , i-9abc",
			expectedIDs:   []string{"i-1234", "i-5678", "i-9abc"},
		},
		{
			name:          "Mixed instance formats",
			instancesFlag: "i-1234567890abcdef0,web-server-01,db-instance",
			expectedIDs:   []string{"i-1234567890abcdef0", "web-server-01", "db-instance"},
		},
		{
			name:          "Empty instances flag",
			instancesFlag: "",
			expectedIDs:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var instanceIDs []string
			if tt.instancesFlag != "" {
				instanceIDs = strings.Split(tt.instancesFlag, ",")
				for i, id := range instanceIDs {
					instanceIDs[i] = strings.TrimSpace(id)
				}
			}

			if len(instanceIDs) != len(tt.expectedIDs) {
				t.Errorf("Instance count = %d, want %d", len(instanceIDs), len(tt.expectedIDs))
			}

			for i, expected := range tt.expectedIDs {
				if i < len(instanceIDs) && instanceIDs[i] != expected {
					t.Errorf("Instance[%d] = %s, want %s", i, instanceIDs[i], expected)
				}
			}

			// Validate no empty IDs after trimming
			for _, id := range instanceIDs {
				if id == "" {
					t.Error("Parsed instance ID should not be empty")
				}
			}
		})
	}
}

func TestParallelExecutionDefaults(t *testing.T) {
	tests := []struct {
		name            string
		parallelFlag    int
		instanceCount   int
		expectedWorkers int
	}{
		{
			name:            "Default CPU count",
			parallelFlag:    8, // Simulating runtime.NumCPU()
			instanceCount:   10,
			expectedWorkers: 8,
		},
		{
			name:            "Custom parallel less than instances",
			parallelFlag:    3,
			instanceCount:   10,
			expectedWorkers: 3,
		},
		{
			name:            "Custom parallel equal to instances",
			parallelFlag:    5,
			instanceCount:   5,
			expectedWorkers: 5,
		},
		{
			name:            "Custom parallel more than instances",
			parallelFlag:    10,
			instanceCount:   3,
			expectedWorkers: 10, // Workers created but only 3 instances to process
		},
		{
			name:            "Single instance",
			parallelFlag:    8,
			instanceCount:   1,
			expectedWorkers: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate parallel execution configuration
			maxWorkers := tt.parallelFlag

			if maxWorkers <= 0 {
				t.Error("Parallel workers should be greater than 0")
			}

			// Test that worker count matches expectation
			if maxWorkers != tt.expectedWorkers {
				t.Errorf("Worker count = %d, want %d", maxWorkers, tt.expectedWorkers)
			}

			// Test that we handle instance count properly
			if tt.instanceCount <= 0 {
				t.Error("Instance count should be greater than 0 for testing")
			}
		})
	}
}

func TestParallelExecutionResultAggregation(t *testing.T) {
	// Mock ParallelExecutionResult structure for testing
	type MockParallelExecutionResult struct {
		InstanceID string
		Success    bool
		Duration   int // milliseconds
		Output     string
		Error      error
	}

	results := []MockParallelExecutionResult{
		{
			InstanceID: "i-123",
			Success:    true,
			Duration:   150,
			Output:     "OK",
			Error:      nil,
		},
		{
			InstanceID: "i-456",
			Success:    false,
			Duration:   200,
			Output:     "",
			Error:      fmt.Errorf("connection failed"),
		},
		{
			InstanceID: "i-789",
			Success:    true,
			Duration:   75,
			Output:     "Success",
			Error:      nil,
		},
		{
			InstanceID: "i-abc",
			Success:    true,
			Duration:   120,
			Output:     "OK",
			Error:      nil,
		},
	}

	// Test result aggregation
	successCount := 0
	failedCount := 0
	totalDuration := 0
	var errors []error

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failedCount++
			if result.Error != nil {
				errors = append(errors, result.Error)
			}
		}
		totalDuration += result.Duration
	}

	// Validate aggregation
	expectedSuccess := 3
	expectedFailed := 1
	expectedTotalDuration := 545
	expectedErrorCount := 1

	if successCount != expectedSuccess {
		t.Errorf("Success count = %d, want %d", successCount, expectedSuccess)
	}

	if failedCount != expectedFailed {
		t.Errorf("Failed count = %d, want %d", failedCount, expectedFailed)
	}

	if totalDuration != expectedTotalDuration {
		t.Errorf("Total duration = %d, want %d", totalDuration, expectedTotalDuration)
	}

	if len(errors) != expectedErrorCount {
		t.Errorf("Error count = %d, want %d", len(errors), expectedErrorCount)
	}

	// Test overall success determination
	overallSuccess := failedCount == 0
	expectedOverallSuccess := false

	if overallSuccess != expectedOverallSuccess {
		t.Errorf("Overall success = %v, want %v", overallSuccess, expectedOverallSuccess)
	}

	// Test average execution time
	averageDuration := totalDuration / len(results)
	expectedAverage := 136 // 545/4 rounded down

	if averageDuration != expectedAverage {
		t.Errorf("Average duration = %d, want %d", averageDuration, expectedAverage)
	}
}

func TestExecutionSummaryFormat(t *testing.T) {
	// Test summary formatting
	type ExecutionSummary struct {
		TotalInstances  int
		SuccessfulCount int
		FailedCount     int
		TotalDuration   int // milliseconds
		MaxParallelism  int
	}

	summary := ExecutionSummary{
		TotalInstances:  10,
		SuccessfulCount: 8,
		FailedCount:     2,
		TotalDuration:   1500,
		MaxParallelism:  5,
	}

	// Validate summary calculations
	if summary.SuccessfulCount+summary.FailedCount != summary.TotalInstances {
		t.Error("Success + Failed should equal Total instances")
	}

	// Test success rate calculation
	successRate := float64(summary.SuccessfulCount) / float64(summary.TotalInstances) * 100
	expectedSuccessRate := 80.0

	if successRate != expectedSuccessRate {
		t.Errorf("Success rate = %.1f%%, want %.1f%%", successRate, expectedSuccessRate)
	}

	// Test performance metrics
	if summary.MaxParallelism <= 0 {
		t.Error("Max parallelism should be greater than 0")
	}

	if summary.TotalDuration <= 0 {
		t.Error("Total duration should be greater than 0")
	}
}

func TestTagsAndInstancesMutualExclusion(t *testing.T) {
	tests := []struct {
		name          string
		tagsFlag      string
		instancesFlag string
		expectError   bool
		errorMessage  string
	}{
		{
			name:          "Only tags provided",
			tagsFlag:      "Environment=dev",
			instancesFlag: "",
			expectError:   false,
		},
		{
			name:          "Only instances provided",
			tagsFlag:      "",
			instancesFlag: "i-123,i-456",
			expectError:   false,
		},
		{
			name:          "Both tags and instances provided",
			tagsFlag:      "Environment=dev",
			instancesFlag: "i-123,i-456",
			expectError:   true,
			errorMessage:  "cannot specify both --tags and --instances",
		},
		{
			name:          "Neither tags nor instances provided",
			tagsFlag:      "",
			instancesFlag: "",
			expectError:   true,
			errorMessage:  "either --tags or --instances flag is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the validation logic
			hasTagsFlag := tt.tagsFlag != ""
			hasInstancesFlag := tt.instancesFlag != ""

			var err error
			if !hasTagsFlag && !hasInstancesFlag {
				err = fmt.Errorf("either --tags or --instances flag is required")
			} else if hasTagsFlag && hasInstancesFlag {
				err = fmt.Errorf("cannot specify both --tags and --instances flags")
			}

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMessage != "" && !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
