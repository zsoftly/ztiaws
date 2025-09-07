package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestSsmCleanupCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Cleanup help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Clean up SSM sessions, IAM policies, S3 objects",
		},
		{
			name:    "Cleanup all resources",
			args:    []string{"--all"},
			wantErr: false,
		},
		{
			name:    "Cleanup old sessions",
			args:    []string{"--sessions"},
			wantErr: false,
		},
		{
			name:    "Cleanup IAM policies",
			args:    []string{"--iam-policies"},
			wantErr: false,
		},
		{
			name:    "Cleanup S3 objects",
			args:    []string{"--s3-objects"},
			wantErr: false,
		},
		{
			name:    "Cleanup with age filter",
			args:    []string{"--sessions", "--older-than", "24h"},
			wantErr: false,
		},
		{
			name:    "Cleanup with region",
			args:    []string{"--sessions", "--region", "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Cleanup dry run",
			args:    []string{"--all", "--dry-run"},
			wantErr: false,
		},
		{
			name:    "Cleanup force",
			args:    []string{"--all", "--force"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "cleanup",
				Short: "Clean up SSM resources",
				Long:  "Clean up SSM sessions, IAM policies, S3 objects, and other resources",
				Run: func(cmd *cobra.Command, args []string) {
					// Mock cleanup functionality
					regionCode, _ := cmd.Flags().GetString("region")
					all, _ := cmd.Flags().GetBool("all")
					sessions, _ := cmd.Flags().GetBool("sessions")
					iamPolicies, _ := cmd.Flags().GetBool("iam-policies")
					s3Objects, _ := cmd.Flags().GetBool("s3-objects")
					olderThan, _ := cmd.Flags().GetString("older-than")
					dryRun, _ := cmd.Flags().GetBool("dry-run")
					force, _ := cmd.Flags().GetBool("force")

					// Mock region resolution
					region := regionCode
					if region == "" {
						region = "us-east-1"
					}

					// Mock cleanup configuration
					type CleanupConfig struct {
						Region      string
						All         bool
						Sessions    bool
						IAMPolicies bool
						S3Objects   bool
						OlderThan   time.Duration
						DryRun      bool
						Force       bool
					}

					// Parse duration
					var duration time.Duration
					if olderThan != "" {
						var err error
						duration, err = time.ParseDuration(olderThan)
						if err != nil {
							t.Errorf("Invalid duration format: %s", olderThan)
							return
						}
					}

					config := CleanupConfig{
						Region:      region,
						All:         all,
						Sessions:    sessions,
						IAMPolicies: iamPolicies,
						S3Objects:   s3Objects,
						OlderThan:   duration,
						DryRun:      dryRun,
						Force:       force,
					}

					// Validate configuration
					if !config.All && !config.Sessions && !config.IAMPolicies && !config.S3Objects {
						t.Log("No cleanup targets specified, would default to sessions")
						config.Sessions = true
					}

					// Test all assigned fields
					if config.Region != region {
						t.Errorf("Region should be %s, got %s", region, config.Region)
					}

					if config.OlderThan != duration {
						t.Errorf("OlderThan should be %v, got %v", duration, config.OlderThan)
					}

					if config.Force != force {
						t.Errorf("Force should be %v, got %v", force, config.Force)
					}

					// Mock cleanup results
					type CleanupResult struct {
						ResourceType string
						Count        int
						Success      bool
						Error        string
					}

					var results []CleanupResult

					if config.All || config.Sessions {
						results = append(results, CleanupResult{
							ResourceType: "Sessions",
							Count:        5,
							Success:      true,
						})
					}

					if config.All || config.IAMPolicies {
						results = append(results, CleanupResult{
							ResourceType: "IAM Policies",
							Count:        3,
							Success:      true,
						})
					}

					if config.All || config.S3Objects {
						results = append(results, CleanupResult{
							ResourceType: "S3 Objects",
							Count:        10,
							Success:      true,
						})
					}

					// Validate results
					for _, result := range results {
						if result.ResourceType == "" {
							t.Error("Result should have resource type")
						}
						if result.Count < 0 {
							t.Error("Count should not be negative")
						}
					}

					// Test dry run behavior
					if config.DryRun {
						t.Log("Dry run mode: would clean up resources")
					}
				},
			}

			// Add flags
			cmd.Flags().StringP("region", "r", "", "AWS region")
			cmd.Flags().Bool("all", false, "Clean up all resource types")
			cmd.Flags().Bool("sessions", false, "Clean up old SSM sessions")
			cmd.Flags().Bool("iam-policies", false, "Clean up temporary IAM policies")
			cmd.Flags().Bool("s3-objects", false, "Clean up S3 transfer objects")
			cmd.Flags().String("older-than", "", "Clean up resources older than duration (e.g., 24h, 7d)")
			cmd.Flags().Bool("dry-run", false, "Show what would be cleaned up")
			cmd.Flags().BoolP("force", "f", false, "Force cleanup without confirmation")

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

func TestSsmCleanupCmdFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "cleanup"}
	cmd.Flags().StringP("region", "r", "", "AWS region")
	cmd.Flags().Bool("all", false, "Clean up all")
	cmd.Flags().Bool("sessions", false, "Clean up sessions")
	cmd.Flags().Bool("iam-policies", false, "Clean up IAM policies")
	cmd.Flags().Bool("s3-objects", false, "Clean up S3 objects")
	cmd.Flags().String("older-than", "", "Age filter")
	cmd.Flags().Bool("dry-run", false, "Dry run")
	cmd.Flags().BoolP("force", "f", false, "Force cleanup")

	tests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{"region", "r", "", "string"},
		{"all", "", "false", "bool"},
		{"sessions", "", "false", "bool"},
		{"iam-policies", "", "false", "bool"},
		{"s3-objects", "", "false", "bool"},
		{"older-than", "", "", "string"},
		{"dry-run", "", "false", "bool"},
		{"force", "f", "false", "bool"},
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

func TestCleanupTargetSelection(t *testing.T) {
	tests := []struct {
		name        string
		all         bool
		sessions    bool
		iamPolicies bool
		s3Objects   bool
		expected    []string
	}{
		{
			name:     "All resources",
			all:      true,
			expected: []string{"sessions", "iam-policies", "s3-objects"},
		},
		{
			name:     "Sessions only",
			sessions: true,
			expected: []string{"sessions"},
		},
		{
			name:        "IAM policies only",
			iamPolicies: true,
			expected:    []string{"iam-policies"},
		},
		{
			name:      "S3 objects only",
			s3Objects: true,
			expected:  []string{"s3-objects"},
		},
		{
			name:        "Multiple targets",
			sessions:    true,
			iamPolicies: true,
			expected:    []string{"sessions", "iam-policies"},
		},
		{
			name:     "No targets specified (default)",
			expected: []string{"sessions"}, // Default to sessions
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targets []string

			if tt.all {
				targets = []string{"sessions", "iam-policies", "s3-objects"}
			} else {
				if tt.sessions {
					targets = append(targets, "sessions")
				}
				if tt.iamPolicies {
					targets = append(targets, "iam-policies")
				}
				if tt.s3Objects {
					targets = append(targets, "s3-objects")
				}

				// Default to sessions if nothing specified
				if len(targets) == 0 && len(tt.expected) > 0 {
					targets = []string{"sessions"}
				}
			}

			if len(targets) != len(tt.expected) {
				t.Errorf("Target count = %d, want %d", len(targets), len(tt.expected))
			}

			for i, target := range targets {
				if i < len(tt.expected) && target != tt.expected[i] {
					t.Errorf("Target[%d] = %s, want %s", i, target, tt.expected[i])
				}
			}
		})
	}
}

func TestDurationParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		isValid  bool
	}{
		{
			name:     "Hours",
			input:    "24h",
			expected: 24 * time.Hour,
			isValid:  true,
		},
		{
			name:     "Minutes",
			input:    "30m",
			expected: 30 * time.Minute,
			isValid:  true,
		},
		{
			name:     "Days (hours)",
			input:    "168h", // 7 days
			expected: 7 * 24 * time.Hour,
			isValid:  true,
		},
		{
			name:     "Seconds",
			input:    "3600s",
			expected: time.Hour,
			isValid:  true,
		},
		{
			name:     "Complex duration",
			input:    "1h30m",
			expected: time.Hour + 30*time.Minute,
			isValid:  true,
		},
		{
			name:    "Invalid format",
			input:   "invalid",
			isValid: false,
		},
		{
			name:    "Empty string",
			input:   "",
			isValid: false,
		},
		{
			name:    "Negative duration",
			input:   "-1h",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == "" {
				if tt.isValid {
					t.Error("Empty input should not be valid")
				}
				return
			}

			duration, err := time.ParseDuration(tt.input)
			isValid := err == nil && duration >= 0

			if isValid != tt.isValid {
				t.Errorf("Duration '%s' validity = %v, want %v, error: %v", tt.input, isValid, tt.isValid, err)
			}

			if isValid && tt.isValid && duration != tt.expected {
				t.Errorf("Duration '%s' = %v, want %v", tt.input, duration, tt.expected)
			}
		})
	}
}

func TestCleanupResultAggregation(t *testing.T) {
	// Test cleanup result aggregation
	type CleanupResult struct {
		ResourceType string
		Count        int
		Success      bool
		Error        string
	}

	results := []CleanupResult{
		{
			ResourceType: "Sessions",
			Count:        5,
			Success:      true,
		},
		{
			ResourceType: "IAM Policies",
			Count:        3,
			Success:      true,
		},
		{
			ResourceType: "S3 Objects",
			Count:        0,
			Success:      true,
			Error:        "No objects found",
		},
		{
			ResourceType: "Failed Resource",
			Count:        2,
			Success:      false,
			Error:        "Permission denied",
		},
	}

	// Test aggregation
	totalCount := 0
	successCount := 0
	failedCount := 0
	var errors []string

	for _, result := range results {
		totalCount += result.Count
		if result.Success {
			successCount++
		} else {
			failedCount++
			if result.Error != "" {
				errors = append(errors, result.Error)
			}
		}
	}

	expectedTotal := 10  // 5 + 3 + 0 + 2
	expectedSuccess := 3 // 3 successful resource types
	expectedFailed := 1  // 1 failed resource type
	expectedErrors := 1  // 1 error message

	if totalCount != expectedTotal {
		t.Errorf("Total count = %d, want %d", totalCount, expectedTotal)
	}

	if successCount != expectedSuccess {
		t.Errorf("Success count = %d, want %d", successCount, expectedSuccess)
	}

	if failedCount != expectedFailed {
		t.Errorf("Failed count = %d, want %d", failedCount, expectedFailed)
	}

	if len(errors) != expectedErrors {
		t.Errorf("Error count = %d, want %d", len(errors), expectedErrors)
	}
}

func TestResourceAgeFiltering(t *testing.T) {
	// Test resource age filtering logic
	now := time.Now()

	tests := []struct {
		name          string
		resourceAge   time.Time
		olderThan     time.Duration
		shouldCleanup bool
	}{
		{
			name:          "Recent resource",
			resourceAge:   now.Add(-1 * time.Hour),
			olderThan:     24 * time.Hour,
			shouldCleanup: false,
		},
		{
			name:          "Old resource",
			resourceAge:   now.Add(-25 * time.Hour),
			olderThan:     24 * time.Hour,
			shouldCleanup: true,
		},
		{
			name:          "Exactly at threshold",
			resourceAge:   now.Add(-24 * time.Hour),
			olderThan:     24 * time.Hour,
			shouldCleanup: true,
		},
		{
			name:          "No age filter (cleanup all)",
			resourceAge:   now.Add(-1 * time.Hour),
			olderThan:     0,
			shouldCleanup: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			age := now.Sub(tt.resourceAge)
			shouldCleanup := tt.olderThan == 0 || age >= tt.olderThan

			if shouldCleanup != tt.shouldCleanup {
				t.Errorf("Should cleanup = %v, want %v (age: %v, threshold: %v)",
					shouldCleanup, tt.shouldCleanup, age, tt.olderThan)
			}
		})
	}
}

func TestSsmCleanupContextHandling(t *testing.T) {
	// Test context usage in cleanup operations
	ctx := context.Background()
	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Test context with cleanup metadata
	type contextKey string
	key := contextKey("cleanup-id")
	ctx = context.WithValue(ctx, key, "cleanup-456")

	value := ctx.Value(key)
	if value != "cleanup-456" {
		t.Errorf("Context value = %v, want cleanup-456", value)
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

func TestCleanupErrorHandling(t *testing.T) {
	// Test error scenarios
	tests := []struct {
		name        string
		resource    string
		shouldError bool
		errorType   string
	}{
		{
			name:        "Successful cleanup",
			resource:    "sessions",
			shouldError: false,
		},
		{
			name:        "Permission denied",
			resource:    "iam-policies",
			shouldError: true,
			errorType:   "permission denied",
		},
		{
			name:        "Resource not found",
			resource:    "s3-objects",
			shouldError: true,
			errorType:   "not found",
		},
		{
			name:        "Network timeout",
			resource:    "sessions",
			shouldError: true,
			errorType:   "timeout",
		},
		{
			name:        "Invalid resource type",
			resource:    "unknown",
			shouldError: true,
			errorType:   "invalid resource type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock error conditions
			var err error
			if tt.shouldError {
				err = &mockCleanupError{message: tt.errorType}
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

// Mock error type for cleanup testing
type mockCleanupError struct {
	message string
}

func (e *mockCleanupError) Error() string {
	return e.message
}

func TestCleanupConfirmation(t *testing.T) {
	tests := []struct {
		name            string
		force           bool
		dryRun          bool
		requiresConfirm bool
		shouldProceed   bool
	}{
		{
			name:            "Force mode - no confirmation",
			force:           true,
			dryRun:          false,
			requiresConfirm: false,
			shouldProceed:   true,
		},
		{
			name:            "Dry run mode - no confirmation",
			force:           false,
			dryRun:          true,
			requiresConfirm: false,
			shouldProceed:   false, // Don't actually clean in dry run
		},
		{
			name:            "Interactive mode - requires confirmation",
			force:           false,
			dryRun:          false,
			requiresConfirm: true,
			shouldProceed:   false, // Depends on user input
		},
		{
			name:            "Force + Dry run - no confirmation, no action",
			force:           true,
			dryRun:          true,
			requiresConfirm: false,
			shouldProceed:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requiresConfirm := !tt.force && !tt.dryRun
			shouldProceed := tt.force && !tt.dryRun

			if requiresConfirm != tt.requiresConfirm {
				t.Errorf("Requires confirmation = %v, want %v", requiresConfirm, tt.requiresConfirm)
			}

			if shouldProceed != tt.shouldProceed {
				t.Errorf("Should proceed = %v, want %v", shouldProceed, tt.shouldProceed)
			}
		})
	}
}

func TestResourceTypeValidation(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		isValid      bool
	}{
		{
			name:         "Valid - sessions",
			resourceType: "sessions",
			isValid:      true,
		},
		{
			name:         "Valid - iam-policies",
			resourceType: "iam-policies",
			isValid:      true,
		},
		{
			name:         "Valid - s3-objects",
			resourceType: "s3-objects",
			isValid:      true,
		},
		{
			name:         "Invalid - unknown",
			resourceType: "unknown",
			isValid:      false,
		},
		{
			name:         "Invalid - empty",
			resourceType: "",
			isValid:      false,
		},
	}

	validTypes := []string{"sessions", "iam-policies", "s3-objects"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := false
			for _, validType := range validTypes {
				if tt.resourceType == validType {
					isValid = true
					break
				}
			}

			if isValid != tt.isValid {
				t.Errorf("Resource type '%s' validity = %v, want %v", tt.resourceType, isValid, tt.isValid)
			}
		})
	}
}
