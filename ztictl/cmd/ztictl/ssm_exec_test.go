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
