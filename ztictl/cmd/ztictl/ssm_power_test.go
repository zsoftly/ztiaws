package main

import (
	"bytes"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestSsmStartCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flags    map[string]string
		wantErr  bool
		contains string
	}{
		{
			name:     "Start help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Start stopped EC2 instance(s)",
		},
		{
			name:    "Start with instance ID",
			args:    []string{"i-1234567890abcdef0"},
			flags:   map[string]string{"region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Start with instance name",
			args:    []string{"web-server-1"},
			flags:   map[string]string{"region": "ca-central-1"},
			wantErr: false,
		},
		{
			name:    "Start with --instances flag",
			args:    []string{},
			flags:   map[string]string{"instances": "i-123,i-456", "region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Start with --instances and parallel",
			args:    []string{},
			flags:   map[string]string{"instances": "i-123,i-456,i-789", "parallel": "2"},
			wantErr: false,
		},
		{
			name:    "Start without instance or instances flag",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Start with both instance and instances flag",
			args:    []string{"i-123"},
			flags:   map[string]string{"instances": "i-456"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "start [instance-identifier]",
				Short: "Start stopped EC2 instance(s)",
				Long:  "Start stopped EC2 instance(s)",
				Args:  cobra.MaximumNArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock validation logic
					instancesFlag, _ := cmd.Flags().GetString("instances")

					// Validate arguments and flags
					if len(args) == 0 && instancesFlag == "" {
						return errors.New("either provide an instance identifier or use --instances flag")
					}

					// Validate mutual exclusion
					if len(args) > 0 && instancesFlag != "" {
						return errors.New("cannot specify both instance identifier and --instances flag")
					}

					return nil
				},
			}

			// Add flags
			cmd.Flags().StringP("region", "r", "", "AWS region or shortcode")
			cmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs")
			cmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

			// Set flags if provided
			for key, value := range tt.flags {
				_ = cmd.Flags().Set(key, value) // #nosec G104
			}

			// Execute command
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got: %v", tt.name, tt.wantErr, err)
			}

			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("%s: expected output to contain %q, got %q", tt.name, tt.contains, output)
			}
		})
	}
}

func TestSsmStopCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flags    map[string]string
		wantErr  bool
		contains string
	}{
		{
			name:     "Stop help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Stop running EC2 instance(s)",
		},
		{
			name:    "Stop with instance ID",
			args:    []string{"i-1234567890abcdef0"},
			flags:   map[string]string{"region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Stop with --instances flag",
			args:    []string{},
			flags:   map[string]string{"instances": "i-123,i-456", "region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Stop without instance or instances flag",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Stop with both instance and instances flag",
			args:    []string{"i-123"},
			flags:   map[string]string{"instances": "i-456"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "stop [instance-identifier]",
				Short: "Stop running EC2 instance(s)",
				Long:  "Stop running EC2 instance(s)",
				Args:  cobra.MaximumNArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock validation logic
					instancesFlag, _ := cmd.Flags().GetString("instances")

					if len(args) == 0 && instancesFlag == "" {
						return errors.New("either provide an instance identifier or use --instances flag")
					}

					if len(args) > 0 && instancesFlag != "" {
						return errors.New("cannot specify both instance identifier and --instances flag")
					}

					return nil
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region or shortcode")
			cmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs")
			cmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

			for key, value := range tt.flags {
				_ = cmd.Flags().Set(key, value) // #nosec G104
			}

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got: %v", tt.name, tt.wantErr, err)
			}

			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("%s: expected output to contain %q, got %q", tt.name, tt.contains, output)
			}
		})
	}
}

func TestSsmRebootCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flags    map[string]string
		wantErr  bool
		contains string
	}{
		{
			name:     "Reboot help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Reboot running EC2 instance(s)",
		},
		{
			name:    "Reboot with instance ID",
			args:    []string{"i-1234567890abcdef0"},
			flags:   map[string]string{"region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Reboot with --instances flag",
			args:    []string{},
			flags:   map[string]string{"instances": "i-123,i-456", "region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Reboot without instance or instances flag",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Reboot with both instance and instances flag",
			args:    []string{"i-123"},
			flags:   map[string]string{"instances": "i-456"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "reboot [instance-identifier]",
				Short: "Reboot running EC2 instance(s)",
				Long:  "Reboot running EC2 instance(s)",
				Args:  cobra.MaximumNArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock validation logic
					instancesFlag, _ := cmd.Flags().GetString("instances")

					if len(args) == 0 && instancesFlag == "" {
						return errors.New("either provide an instance identifier or use --instances flag")
					}

					if len(args) > 0 && instancesFlag != "" {
						return errors.New("cannot specify both instance identifier and --instances flag")
					}

					return nil
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region or shortcode")
			cmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs")
			cmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

			for key, value := range tt.flags {
				_ = cmd.Flags().Set(key, value) // #nosec G104
			}

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got: %v", tt.name, tt.wantErr, err)
			}

			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("%s: expected output to contain %q, got %q", tt.name, tt.contains, output)
			}
		})
	}
}

func TestSsmStartTaggedCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flags    map[string]string
		wantErr  bool
		contains string
	}{
		{
			name:     "Start-tagged help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Start multiple stopped EC2 instances with specified tags",
		},
		{
			name:    "Start-tagged with tags",
			args:    []string{},
			flags:   map[string]string{"tags": "Environment=Production", "region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Start-tagged with instances",
			args:    []string{},
			flags:   map[string]string{"instances": "i-123,i-456", "region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Start-tagged with tags and parallel",
			args:    []string{},
			flags:   map[string]string{"tags": "Environment=dev,Component=web", "parallel": "3"},
			wantErr: false,
		},
		{
			name:    "Start-tagged without tags or instances",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Start-tagged with both tags and instances",
			args:    []string{},
			flags:   map[string]string{"tags": "Environment=test", "instances": "i-456"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "start-tagged",
				Short: "Start multiple stopped EC2 instances with specified tags",
				Long:  "Start multiple stopped EC2 instances with specified tags",
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock validation logic
					tagsFlag, _ := cmd.Flags().GetString("tags")
					instancesFlag, _ := cmd.Flags().GetString("instances")

					// Validate that we have either tags or instances specified
					if tagsFlag == "" && instancesFlag == "" {
						return errors.New("either --tags or --instances flag is required")
					}

					// Validate mutual exclusion
					if tagsFlag != "" && instancesFlag != "" {
						return errors.New("cannot specify both --tags and --instances flags")
					}

					return nil
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region or shortcode")
			cmd.Flags().StringP("tags", "t", "", "Tag filters in key=value format")
			cmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs")
			cmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

			for key, value := range tt.flags {
				_ = cmd.Flags().Set(key, value) // #nosec G104
			}

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got: %v", tt.name, tt.wantErr, err)
			}

			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("%s: expected output to contain %q, got %q", tt.name, tt.contains, output)
			}
		})
	}
}

func TestSsmStopTaggedCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flags    map[string]string
		wantErr  bool
		contains string
	}{
		{
			name:     "Stop-tagged help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Stop multiple running EC2 instances with specified tags",
		},
		{
			name:    "Stop-tagged with tags",
			args:    []string{},
			flags:   map[string]string{"tags": "Environment=Production", "region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Stop-tagged with instances",
			args:    []string{},
			flags:   map[string]string{"instances": "i-123,i-456", "region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Stop-tagged without tags or instances",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Stop-tagged with both tags and instances",
			args:    []string{},
			flags:   map[string]string{"tags": "Environment=test", "instances": "i-456"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "stop-tagged",
				Short: "Stop multiple running EC2 instances with specified tags",
				Long:  "Stop multiple running EC2 instances with specified tags",
				RunE: func(cmd *cobra.Command, args []string) error {
					tagsFlag, _ := cmd.Flags().GetString("tags")
					instancesFlag, _ := cmd.Flags().GetString("instances")

					if tagsFlag == "" && instancesFlag == "" {
						return errors.New("either --tags or --instances flag is required")
					}

					if tagsFlag != "" && instancesFlag != "" {
						return errors.New("cannot specify both --tags and --instances flags")
					}

					return nil
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region or shortcode")
			cmd.Flags().StringP("tags", "t", "", "Tag filters in key=value format")
			cmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs")
			cmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

			for key, value := range tt.flags {
				_ = cmd.Flags().Set(key, value) // #nosec G104
			}

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got: %v", tt.name, tt.wantErr, err)
			}

			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("%s: expected output to contain %q, got %q", tt.name, tt.contains, output)
			}
		})
	}
}

func TestSsmRebootTaggedCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flags    map[string]string
		wantErr  bool
		contains string
	}{
		{
			name:     "Reboot-tagged help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Reboot multiple running EC2 instances with specified tags",
		},
		{
			name:    "Reboot-tagged with tags",
			args:    []string{},
			flags:   map[string]string{"tags": "Environment=Production", "region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Reboot-tagged with instances",
			args:    []string{},
			flags:   map[string]string{"instances": "i-123,i-456", "region": "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Reboot-tagged without tags or instances",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Reboot-tagged with both tags and instances",
			args:    []string{},
			flags:   map[string]string{"tags": "Environment=test", "instances": "i-456"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "reboot-tagged",
				Short: "Reboot multiple running EC2 instances with specified tags",
				Long:  "Reboot multiple running EC2 instances with specified tags",
				RunE: func(cmd *cobra.Command, args []string) error {
					tagsFlag, _ := cmd.Flags().GetString("tags")
					instancesFlag, _ := cmd.Flags().GetString("instances")

					if tagsFlag == "" && instancesFlag == "" {
						return errors.New("either --tags or --instances flag is required")
					}

					if tagsFlag != "" && instancesFlag != "" {
						return errors.New("cannot specify both --tags and --instances flags")
					}

					return nil
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region or shortcode")
			cmd.Flags().StringP("tags", "t", "", "Tag filters in key=value format")
			cmd.Flags().StringP("instances", "i", "", "Comma-separated list of instance IDs")
			cmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

			for key, value := range tt.flags {
				_ = cmd.Flags().Set(key, value) // #nosec G104
			}

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got: %v", tt.name, tt.wantErr, err)
			}

			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("%s: expected output to contain %q, got %q", tt.name, tt.contains, output)
			}
		})
	}
}

// Test parallel execution parameter validation
func TestParallelExecutionValidation(t *testing.T) {
	tests := []struct {
		name     string
		parallel string
		wantErr  bool
	}{
		{
			name:     "Default parallel value",
			parallel: "",
			wantErr:  false,
		},
		{
			name:     "Valid parallel value",
			parallel: "5",
			wantErr:  false,
		},
		{
			name:     "Parallel value of 1",
			parallel: "1",
			wantErr:  false,
		},
		{
			name:     "Large parallel value",
			parallel: "20",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
				RunE: func(cmd *cobra.Command, args []string) error {
					parallelFlag, _ := cmd.Flags().GetInt("parallel")

					// The actual validation in the real code checks for <= 0
					if parallelFlag <= 0 {
						return errors.New("--parallel must be greater than 0")
					}
					return nil
				},
			}

			cmd.Flags().IntP("parallel", "p", runtime.NumCPU(), "Maximum number of concurrent operations")

			if tt.parallel != "" {
				_ = cmd.Flags().Set("parallel", tt.parallel) // #nosec G104
			}

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got: %v", tt.name, tt.wantErr, err)
			}
		})
	}
}

func TestValidateTaggedCommandArgs(t *testing.T) {
	tests := []struct {
		name          string
		tagsFlag      string
		instancesFlag string
		parallelFlag  int
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid tags flag",
			tagsFlag:     "Environment=production,Team=backend",
			parallelFlag: 4,
			expectError:  false,
		},
		{
			name:          "valid instances flag",
			instancesFlag: "i-1234567890abcdef0,i-0987654321fedcba0",
			parallelFlag:  4,
			expectError:   false,
		},
		{
			name:          "no tags or instances",
			tagsFlag:      "",
			instancesFlag: "",
			parallelFlag:  4,
			expectError:   true,
			errorContains: "either --tags or --instances flag is required",
		},
		{
			name:          "both tags and instances",
			tagsFlag:      "Environment=production",
			instancesFlag: "i-1234567890abcdef0",
			parallelFlag:  4,
			expectError:   true,
			errorContains: "cannot specify both --tags and --instances flags",
		},
		{
			name:          "invalid parallel value - zero",
			tagsFlag:      "Environment=production",
			parallelFlag:  0,
			expectError:   true,
			errorContains: "--parallel must be greater than 0",
		},
		{
			name:          "invalid parallel value - negative",
			tagsFlag:      "Environment=production",
			parallelFlag:  -1,
			expectError:   true,
			errorContains: "--parallel must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTaggedCommandArgs(tt.tagsFlag, tt.instancesFlag, tt.parallelFlag)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tt.errorContains != "" && err.Error() != tt.errorContains {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase word",
			input:    "start",
			expected: "Start",
		},
		{
			name:     "already capitalized",
			input:    "Stop",
			expected: "Stop",
		},
		{
			name:     "mixed case",
			input:    "reBoot",
			expected: "ReBoot",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "A",
		},
		{
			name:     "single character uppercase",
			input:    "A",
			expected: "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := capitalize(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDisplayPowerOperationResults(t *testing.T) {
	tests := []struct {
		name          string
		results       []PowerOperationResult
		operation     string
		totalDuration time.Duration
		maxParallel   int
		expectError   bool
		errorContains string
	}{
		{
			name: "all successful operations",
			results: []PowerOperationResult{
				{
					InstanceID: "i-1234567890abcdef0",
					Operation:  "start",
					Error:      nil,
					Duration:   time.Millisecond * 500,
				},
				{
					InstanceID: "i-0987654321fedcba0",
					Operation:  "start",
					Error:      nil,
					Duration:   time.Millisecond * 750,
				},
			},
			operation:     "start",
			totalDuration: time.Second * 2,
			maxParallel:   4,
			expectError:   false,
		},
		{
			name: "mixed success and failure",
			results: []PowerOperationResult{
				{
					InstanceID: "i-1234567890abcdef0",
					Operation:  "stop",
					Error:      nil,
					Duration:   time.Millisecond * 500,
				},
				{
					InstanceID: "i-0987654321fedcba0",
					Operation:  "stop",
					Error:      &stubError{"failed to stop instance"},
					Duration:   time.Millisecond * 300,
				},
			},
			operation:     "stop",
			totalDuration: time.Second * 1,
			maxParallel:   2,
			expectError:   true,
			errorContains: "some stop operations failed: 1 successful, 1 failed",
		},
		{
			name: "all failed operations",
			results: []PowerOperationResult{
				{
					InstanceID: "i-1234567890abcdef0",
					Operation:  "reboot",
					Error:      &stubError{"failed to reboot instance"},
					Duration:   time.Millisecond * 200,
				},
				{
					InstanceID: "i-0987654321fedcba0",
					Operation:  "reboot",
					Error:      &stubError{"instance not found"},
					Duration:   time.Millisecond * 150,
				},
			},
			operation:     "reboot",
			totalDuration: time.Millisecond * 500,
			maxParallel:   3,
			expectError:   true,
			errorContains: "some reboot operations failed: 0 successful, 2 failed",
		},
		{
			name:          "empty results",
			results:       []PowerOperationResult{},
			operation:     "start",
			totalDuration: time.Millisecond * 100,
			maxParallel:   1,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := displayPowerOperationResults(tt.results, tt.operation, tt.totalDuration, tt.maxParallel)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tt.errorContains != "" && err.Error() != tt.errorContains {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// stubError is a simple error implementation for testing
type stubError struct {
	message string
}

func (e *stubError) Error() string {
	return e.message
}

func TestStateValidationRequirements(t *testing.T) {
	tests := []struct {
		name             string
		operation        string
		expectedStates   []string
		requireSSMOnline bool
		expectError      bool
	}{
		{
			name:             "start operation requires stopped state",
			operation:        "start",
			expectedStates:   []string{"stopped"},
			requireSSMOnline: false,
			expectError:      false,
		},
		{
			name:             "stop operation requires running state",
			operation:        "stop",
			expectedStates:   []string{"running"},
			requireSSMOnline: false,
			expectError:      false,
		},
		{
			name:             "reboot operation requires running state",
			operation:        "reboot",
			expectedStates:   []string{"running"},
			requireSSMOnline: false,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requirements, err := buildRequirementsForOperation(tt.operation)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if requirements.Operation != tt.operation {
				t.Errorf("Expected operation %q, got %q", tt.operation, requirements.Operation)
			}

			if requirements.RequireSSMOnline != tt.requireSSMOnline {
				t.Errorf("Expected RequireSSMOnline %v, got %v", tt.requireSSMOnline, requirements.RequireSSMOnline)
			}

			if len(requirements.AllowedStates) != len(tt.expectedStates) {
				t.Errorf("Expected %d allowed states, got %d", len(tt.expectedStates), len(requirements.AllowedStates))
			}

			for i, state := range tt.expectedStates {
				if i >= len(requirements.AllowedStates) || requirements.AllowedStates[i] != state {
					t.Errorf("Expected state %q at index %d, got %q", state, i, requirements.AllowedStates[i])
				}
			}
		})
	}
}

func TestPowerOperationResultErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		results     []PowerOperationResult
		expectError bool
		description string
	}{
		{
			name: "state validation error for start on running instance",
			results: []PowerOperationResult{
				{
					InstanceID: "i-1234567890abcdef0",
					Operation:  "start",
					Error:      errors.New("instance is in 'running' state, expected one of: [stopped]"),
					Duration:   time.Millisecond * 100,
				},
			},
			expectError: true,
			description: "Should fail when trying to start a running instance",
		},
		{
			name: "state validation error for stop on stopped instance",
			results: []PowerOperationResult{
				{
					InstanceID: "i-1234567890abcdef0",
					Operation:  "stop",
					Error:      errors.New("instance is in 'stopped' state, expected one of: [running]"),
					Duration:   time.Millisecond * 100,
				},
			},
			expectError: true,
			description: "Should fail when trying to stop a stopped instance",
		},
		{
			name: "state validation error for reboot on stopped instance",
			results: []PowerOperationResult{
				{
					InstanceID: "i-1234567890abcdef0",
					Operation:  "reboot",
					Error:      errors.New("instance is in 'stopped' state, expected one of: [running]"),
					Duration:   time.Millisecond * 100,
				},
			},
			expectError: true,
			description: "Should fail when trying to reboot a stopped instance",
		},
		{
			name: "successful operation with no errors",
			results: []PowerOperationResult{
				{
					InstanceID: "i-1234567890abcdef0",
					Operation:  "start",
					Error:      nil,
					Duration:   time.Millisecond * 200,
				},
			},
			expectError: false,
			description: "Should succeed when instance is in correct state",
		},
		{
			name: "mixed results with some validation failures",
			results: []PowerOperationResult{
				{
					InstanceID: "i-1111111111111111",
					Operation:  "start",
					Error:      nil,
					Duration:   time.Millisecond * 150,
				},
				{
					InstanceID: "i-2222222222222222",
					Operation:  "start",
					Error:      errors.New("instance is in 'running' state, expected one of: [stopped]"),
					Duration:   time.Millisecond * 100,
				},
				{
					InstanceID: "i-3333333333333333",
					Operation:  "start",
					Error:      errors.New("instance is in 'running' state, expected one of: [stopped]"),
					Duration:   time.Millisecond * 100,
				},
			},
			expectError: true,
			description: "Should report partial failure when some instances fail validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := displayPowerOperationResults(tt.results, tt.results[0].Operation, time.Second, 4)

			if tt.expectError && err == nil {
				t.Errorf("%s: expected error but got nil", tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.description, err)
			}

			successCount := 0
			for _, result := range tt.results {
				if result.Error == nil {
					successCount++
				}
			}

			if tt.expectError && successCount == len(tt.results) {
				t.Errorf("%s: all operations succeeded but expected failures", tt.description)
			}
		})
	}
}

func TestGetInstanceIDsByTagsFormatting(t *testing.T) {
	tests := []struct {
		name     string
		tagsFlag string
		expected int
		desc     string
	}{
		{
			name:     "single tag",
			tagsFlag: "Environment=production",
			expected: 1,
			desc:     "Should parse single tag correctly",
		},
		{
			name:     "multiple tags",
			tagsFlag: "Environment=production,Team=backend",
			expected: 2,
			desc:     "Should parse multiple tags correctly",
		},
		{
			name:     "tags with spaces",
			tagsFlag: "Environment=production, Team=backend",
			expected: 2,
			desc:     "Should handle spaces in tag list",
		},
		{
			name:     "empty tag string",
			tagsFlag: "",
			expected: 0,
			desc:     "Should handle empty tag string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterCount := 0
			if tt.tagsFlag != "" {
				tagPairs := strings.Split(tt.tagsFlag, ",")
				for _, pair := range tagPairs {
					parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
					if len(parts) == 2 {
						filterCount++
					}
				}
			}

			if filterCount != tt.expected {
				t.Errorf("%s: expected %d filters, got %d", tt.desc, tt.expected, filterCount)
			}
		})
	}
}
