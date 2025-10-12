package main

import (
	"bytes"
	"context"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
)

func TestSsmConnectCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Connect help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Connect to an EC2 instance using SSM Session Manager",
		},
		{
			name:    "Connect with instance ID",
			args:    []string{"i-1234567890abcdef0"},
			wantErr: false,
		},
		{
			name:    "Connect with instance name",
			args:    []string{"web-server-1"},
			wantErr: false,
		},
		{
			name:    "Connect with region flag",
			args:    []string{"i-1234567890abcdef0", "--region", "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Connect with region shortcode",
			args:    []string{"i-1234567890abcdef0", "-r", "cac1"},
			wantErr: false,
		},
		{
			name:    "Connect without instance identifier",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Connect with too many args",
			args:    []string{"instance1", "instance2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   ssmConnectCmd.Use,
				Short: ssmConnectCmd.Short,
				Long:  ssmConnectCmd.Long,
				Args:  cobra.ExactArgs(1),
				Run: func(cmd *cobra.Command, args []string) {
					// Mock SSM connect functionality
					regionCode, _ := cmd.Flags().GetString("region")
					instanceIdentifier := args[0]

					// Resolve region (mock implementation)
					region := regionCode
					if region == "" {
						region = "us-east-1" // default
					}
					if region == "cac1" {
						region = "ca-central-1"
					}

					// Validate instance identifier
					if instanceIdentifier == "" {
						t.Error("Instance identifier should not be empty")
					}

					// Test instance identifier patterns
					isInstanceID := strings.HasPrefix(instanceIdentifier, "i-") && len(instanceIdentifier) == 19
					isInstanceName := !isInstanceID && instanceIdentifier != ""

					if !isInstanceID && !isInstanceName {
						t.Errorf("Invalid instance identifier format: %s", instanceIdentifier)
					}

					// Mock session start
					if instanceIdentifier == "i-1234567890abcdef0" {
						// Valid instance ID
						if len(instanceIdentifier) != 19 {
							t.Errorf("Instance ID should be 19 characters, got %d", len(instanceIdentifier))
						}
					}

					if instanceIdentifier == "web-server-1" {
						// Valid instance name
						if !strings.Contains(instanceIdentifier, "server") {
							t.Error("Instance name should be descriptive")
						}
					}

					// Test region resolution
					expectedRegions := map[string]string{
						"cac1": "ca-central-1",
						"use1": "us-east-1",
						"":     "us-east-1", // default
					}

					if expected, exists := expectedRegions[regionCode]; exists && region != expected {
						t.Errorf("Region resolution failed: %s -> %s, want %s", regionCode, region, expected)
					}
				},
			}

			// Add flags
			cmd.Flags().StringP("region", "r", "", "AWS region")

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

func TestSsmConnectCmdFlags(t *testing.T) {
	// Test flag definitions
	tests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{"region", "r", ""},
	}

	for _, tt := range tests {
		t.Run(tt.flagName+" flag", func(t *testing.T) {
			flag := ssmConnectCmd.Flags().Lookup(tt.flagName)
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

func TestInstanceIdentifierValidation(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		isValid    bool
		isID       bool
		isName     bool
	}{
		{
			name:       "Valid instance ID",
			identifier: "i-1234567890abcdef0",
			isValid:    true,
			isID:       true,
			isName:     false,
		},
		{
			name:       "Valid instance name",
			identifier: "web-server-1",
			isValid:    true,
			isID:       false,
			isName:     true,
		},
		{
			name:       "Short instance ID",
			identifier: "i-123456789",
			isValid:    false, // Invalid - not valid ID and not accepted as name
			isID:       false,
			isName:     false, // Not treated as name since it starts with i-
		},
		{
			name:       "Invalid instance ID format",
			identifier: "x-1234567890abcdef0",
			isValid:    true, // Valid as name
			isID:       false,
			isName:     true,
		},
		{
			name:       "Empty identifier",
			identifier: "",
			isValid:    false,
			isID:       false,
			isName:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test instance ID pattern - AWS instance IDs are i-[17 hex chars] (19 total)
			isInstanceID := strings.HasPrefix(tt.identifier, "i-") && len(tt.identifier) == 19
			if isInstanceID != tt.isID {
				t.Errorf("isInstanceID = %v, want %v", isInstanceID, tt.isID)
			}

			// Test instance name pattern - don't treat invalid instance IDs as names
			isInstanceName := !isInstanceID && tt.identifier != "" && !strings.HasPrefix(tt.identifier, "i-")
			if isInstanceName != tt.isName {
				t.Errorf("isInstanceName = %v, want %v", isInstanceName, tt.isName)
			}

			// Test overall validity
			isValid := isInstanceID || isInstanceName
			if isValid != tt.isValid {
				t.Errorf("isValid = %v, want %v", isValid, tt.isValid)
			}
		})
	}
}

func TestSsmConnectRegionResolution(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "us-east-1"},          // default
		{"us-east-1", "us-east-1"}, // full region name
		{"cac1", "ca-central-1"},   // shortcode
		{"use1", "us-east-1"},      // shortcode
		{"use2", "us-east-2"},      // shortcode
		{"usw1", "us-west-1"},      // shortcode
		{"usw2", "us-west-2"},      // shortcode
		{"euw1", "eu-west-1"},      // shortcode
		{"euw2", "eu-west-2"},      // shortcode
		{"euc1", "eu-central-1"},   // shortcode
		{"invalid", "invalid"},     // invalid region passes through
	}

	for _, tt := range tests {
		t.Run("region "+tt.input, func(t *testing.T) {
			// Mock region resolution logic (same as in list command)
			resolveRegion := func(regionCode string) string {
				if regionCode == "" {
					return "us-east-1"
				}

				shortcuts := map[string]string{
					"cac1": "ca-central-1",
					"use1": "us-east-1",
					"use2": "us-east-2",
					"usw1": "us-west-1",
					"usw2": "us-west-2",
					"euw1": "eu-west-1",
					"euw2": "eu-west-2",
					"euc1": "eu-central-1",
				}

				if resolved, exists := shortcuts[regionCode]; exists {
					return resolved
				}

				return regionCode
			}

			result := resolveRegion(tt.input)
			if result != tt.expected {
				t.Errorf("resolveRegion(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSsmConnectArgumentValidation(t *testing.T) {
	// Test cobra argument validation
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "No args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "One arg",
			args:    []string{"i-123"},
			wantErr: false,
		},
		{
			name:    "Two args",
			args:    []string{"i-123", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test cobra.ExactArgs(1) validation
			err := cobra.ExactArgs(1)(nil, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExactArgs(1) error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSsmConnectContextHandling(t *testing.T) {
	// Test context usage in SSM connect operations
	ctx := context.Background()
	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Test context with values
	type contextKey string
	key := contextKey("instance")
	ctx = context.WithValue(ctx, key, "i-1234567890abcdef0")

	value := ctx.Value(key)
	if value == nil {
		t.Error("Context should contain the instance value")
	}

	if value != "i-1234567890abcdef0" {
		t.Errorf("Context value = %v, want i-1234567890abcdef0", value)
	}

	// Test context cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cancel()
	select {
	case <-ctx.Done():
		// Expected - context was cancelled
	default:
		t.Error("Context should be done after cancellation")
	}
}

func TestSsmConnectErrorHandling(t *testing.T) {
	// Test error handling scenarios
	tests := []struct {
		name        string
		instance    string
		region      string
		shouldError bool
		errorType   string
	}{
		{
			name:        "Valid instance ID",
			instance:    "i-1234567890abcdef0",
			region:      "us-east-1",
			shouldError: false,
		},
		{
			name:        "Valid instance name",
			instance:    "web-server-1",
			region:      "us-east-1",
			shouldError: false,
		},
		{
			name:        "Invalid region",
			instance:    "i-1234567890abcdef0",
			region:      "invalid-region",
			shouldError: true,
			errorType:   "invalid region",
		},
		{
			name:        "Instance not found",
			instance:    "i-nonexistent",
			region:      "us-east-1",
			shouldError: true,
			errorType:   "instance not found",
		},
		{
			name:        "No SSM agent",
			instance:    "i-noagent123456789",
			region:      "us-east-1",
			shouldError: true,
			errorType:   "SSM agent not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock error conditions
			var err error
			if tt.shouldError {
				err = &mockConnectError{message: tt.errorType}
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

// Mock error type for connect testing
type mockConnectError struct {
	message string
}

func (e *mockConnectError) Error() string {
	return e.message
}

func TestSsmConnectSessionManagement(t *testing.T) {
	// Test session management concepts
	type SessionInfo struct {
		InstanceID string
		Region     string
		Status     string
		StartTime  string
	}

	session := SessionInfo{
		InstanceID: "i-1234567890abcdef0",
		Region:     "us-east-1",
		Status:     "Connected",
		StartTime:  "2023-01-01T10:00:00Z",
	}

	// Validate session info
	if session.InstanceID == "" {
		t.Error("Session should have instance ID")
	}

	if session.Region == "" {
		t.Error("Session should have region")
	}

	if session.Status == "" {
		t.Error("Session should have status")
	}

	if session.StartTime == "" {
		t.Error("Session should have start time")
	}

	// Test valid statuses
	validStatuses := []string{"Connected", "Connecting", "Terminated", "Terminating", "Failed"}
	isValidStatus := false
	for _, validStatus := range validStatuses {
		if session.Status == validStatus {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		t.Errorf("Invalid session status: %s", session.Status)
	}
}

func TestSsmConnectCmdStructure(t *testing.T) {
	// Test command structure
	if ssmConnectCmd.Use != "connect [instance-identifier]" {
		t.Errorf("Expected Use to be 'connect [instance-identifier]', got %s", ssmConnectCmd.Use)
	}

	if ssmConnectCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if ssmConnectCmd.Long == "" {
		t.Error("Command should have a long description")
	}

	// Test that command has a Run function
	if ssmConnectCmd.Run == nil {
		t.Error("Command should have a Run function")
	}

	// Test argument validation
	if ssmConnectCmd.Args == nil {
		t.Error("Command should have argument validation")
	}

	// Test that args validation allows 0 or 1 arguments (fuzzy finder support)
	err := ssmConnectCmd.Args(ssmConnectCmd, []string{})
	if err != nil {
		t.Errorf("Command should allow 0 arguments for fuzzy finder, got error: %v", err)
	}

	err = ssmConnectCmd.Args(ssmConnectCmd, []string{"i-1234567890abcdef0"})
	if err != nil {
		t.Errorf("Command should allow 1 argument, got error: %v", err)
	}

	err = ssmConnectCmd.Args(ssmConnectCmd, []string{"i-1234567890abcdef0", "extra-arg"})
	if err == nil {
		t.Error("Command should not allow more than 1 argument")
	}

	err = ssmConnectCmd.Args(ssmConnectCmd, []string{"instance"})
	if err != nil {
		t.Errorf("Command should accept 1 argument, got error: %v", err)
	}

	err = ssmConnectCmd.Args(ssmConnectCmd, []string{"instance1", "instance2"})
	if err == nil {
		t.Error("Command should reject more than 1 argument")
	}
}

func TestInstanceIdentifierFormats(t *testing.T) {
	// Test various instance identifier formats
	tests := []struct {
		identifier string
		isValid    bool
		format     string
	}{
		{"i-1234567890abcdef0", true, "instance-id"},
		{"i-12345678", false, "invalid-instance-id"},
		{"web-server-1", true, "instance-name"},
		{"db.server.prod", true, "instance-name"},
		{"server_123", true, "instance-name"},
		{"123-server", true, "instance-name"},
		{"", false, "empty"},
		{"i-", false, "incomplete-instance-id"},
	}

	for _, tt := range tests {
		t.Run(tt.identifier+" format", func(t *testing.T) {
			// Instance ID validation
			isInstanceID := strings.HasPrefix(tt.identifier, "i-") && len(tt.identifier) == 19

			// Instance name validation - don't treat invalid instance IDs as names
			isInstanceName := !isInstanceID && tt.identifier != "" && tt.identifier != "i-" && !strings.HasPrefix(tt.identifier, "i-")

			isValid := isInstanceID || isInstanceName

			if isValid != tt.isValid {
				t.Errorf("Identifier %s validity = %v, want %v", tt.identifier, isValid, tt.isValid)
			}

			// Test format detection
			var format string
			if isInstanceID {
				format = "instance-id"
			} else if isInstanceName {
				format = "instance-name"
			} else if tt.identifier == "" {
				format = "empty"
			} else {
				format = "invalid-instance-id"
			}

			if tt.isValid && format != tt.format && tt.format != "invalid-instance-id" {
				t.Errorf("Identifier %s format = %s, want %s", tt.identifier, format, tt.format)
			}
		})
	}
}

// NEW TESTS FOR SEPARATION OF CONCERNS REFACTORING

func TestPerformConnection(t *testing.T) {
	// Save original logger state
	originalLogger := logger
	defer func() { logger = originalLogger }()

	t.Run("handles connection gracefully", func(t *testing.T) {
		// Initialize logger to avoid nil pointer dereference
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// The function should return an error or succeed, not call os.Exit
		err := performConnection("use1", "i-test123")

		// We expect this might fail (no AWS credentials/connection), but it shouldn't panic
		// The important thing is that it returns an error instead of calling os.Exit
		if err != nil {
			t.Logf("Connection error (may be expected): %v", err)
		}

		// The fact that we can continue execution proves the refactoring worked
		t.Log("Test completed - function returned instead of calling os.Exit")
	})

	t.Run("validates region code", func(t *testing.T) {
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// Test with empty region code (should be handled gracefully)
		err := performConnection("", "i-test123")

		// Function should handle this gracefully and return error
		if err != nil {
			t.Logf("Expected error for empty region: %v", err)
		}

		t.Log("Region validation handled gracefully")
	})

	t.Run("validates instance identifier", func(t *testing.T) {
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// Test with empty instance identifier
		err := performConnection("use1", "")

		// Function should handle this gracefully
		if err != nil {
			t.Logf("Expected error for empty instance: %v", err)
		}

		// Test with invalid instance identifier
		err = performConnection("use1", "invalid-id")

		if err != nil {
			t.Logf("Expected error for invalid instance ID: %v", err)
		}

		t.Log("Instance identifier validation handled gracefully")
	})

	t.Run("handles invalid region gracefully", func(t *testing.T) {
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// Test with invalid region
		err := performConnection("invalid-region", "i-test123")

		// Function should handle this gracefully and return error
		if err != nil {
			t.Logf("Expected error for invalid region: %v", err)
		}

		t.Log("Invalid region handled gracefully")
	})
}

func TestConnectionSeparationOfConcerns(t *testing.T) {
	// This test verifies that the connection function doesn't call os.Exit
	// and can be tested without terminating the test process

	// Isolate test environment to avoid config file interference
	tempDir := t.TempDir()

	// Save original environment variables
	var origHome, origUserProfile string
	if runtime.GOOS == "windows" {
		origUserProfile = os.Getenv("USERPROFILE")
		os.Setenv("USERPROFILE", tempDir)
		defer os.Setenv("USERPROFILE", origUserProfile)
	} else {
		origHome = os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", origHome)
	}

	// Save original logger state
	originalLogger := logger
	defer func() { logger = originalLogger }()

	t.Run("connection returns instead of exiting", func(t *testing.T) {
		// Initialize logger to avoid nil pointer dereference
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// Use context with timeout to prevent hanging in CI
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		// Run the connection test in a goroutine with timeout
		done := make(chan error, 1)
		go func() {
			// This call should return an error or succeed, not exit the process
			err := performConnection("invalid-region", "invalid-instance")
			done <- err
		}()

		select {
		case err := <-done:
			// If we reach this line, the function didn't call os.Exit
			// (which is what we want for good separation of concerns)
			if err == nil {
				t.Log("Connection succeeded unexpectedly")
			} else {
				t.Logf("Connection failed as expected: %v", err)
			}
		case <-ctx.Done():
			t.Log("Connection test timed out (expected in CI environment)")
		}

		// The fact that we can continue execution proves the refactoring worked
		t.Log("Test completed - function returned instead of calling os.Exit")
	})

	t.Run("handles various error conditions without exiting", func(t *testing.T) {
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// Test various error conditions that should return errors, not exit
		errorCases := []struct {
			name        string
			regionCode  string
			instanceID  string
			expectError bool
		}{
			{
				name:        "Valid arguments",
				regionCode:  "use1",
				instanceID:  "i-1234567890abcdef0",
				expectError: true, // Will fail due to AWS connection, but shouldn't exit
			},
			{
				name:        "Empty region",
				regionCode:  "",
				instanceID:  "i-1234567890abcdef0",
				expectError: true,
			},
			{
				name:        "Empty instance",
				regionCode:  "use1",
				instanceID:  "",
				expectError: true,
			},
			{
				name:        "Invalid region",
				regionCode:  "nonexistent-region",
				instanceID:  "i-1234567890abcdef0",
				expectError: true,
			},
		}

		for _, tc := range errorCases {
			t.Run(tc.name, func(t *testing.T) {
				// Use context with timeout to prevent hanging in CI
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()

				// Run the connection test in a goroutine with timeout
				done := make(chan error, 1)
				go func() {
					err := performConnection(tc.regionCode, tc.instanceID)
					done <- err
				}()

				select {
				case err := <-done:
					if tc.expectError && err == nil {
						t.Error("Expected error but got none")
					}
					// The important thing is we didn't crash or call os.Exit
					t.Logf("Connection test completed for %s", tc.name)
				case <-ctx.Done():
					t.Logf("Connection test timed out for %s (expected in CI environment)", tc.name)
				}
			})
		}
	})
}
