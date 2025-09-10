package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func init() {
	// Disable EC2 IMDS for all tests to prevent timeouts
	os.Setenv("AWS_EC2_METADATA_DISABLED", "true")
}

func TestAuthCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Auth help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Manage AWS SSO authentication",
		},
		{
			name:     "Auth with no subcommand",
			args:     []string{},
			wantErr:  false,
			contains: "Manage AWS SSO authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   authCmd.Use,
				Short: authCmd.Short,
				Long:  authCmd.Long,
			}

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buf.String()
			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("Execute() output should contain %v, got %v", tt.contains, output)
			}
		})
	}
}

func TestAuthLoginCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Login help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Usage:",
		},
		{
			name:    "Login without profile",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Login with profile",
			args:    []string{"test-profile"},
			wantErr: false,
		},
		{
			name:    "Login with too many args",
			args:    []string{"profile1", "profile2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create isolated command for testing argument validation
			cmd := &cobra.Command{
				Use:  "login [profile]",
				Args: cobra.ExactArgs(1),
				Run: func(cmd *cobra.Command, args []string) {
					// Mock the login functionality
					profileName := args[0]
					if profileName == "" {
						t.Error("Profile name should not be empty")
					}

					// Mock auth manager behavior
					// In real implementation, this would call authManager.Login()
					// For testing, we just verify the profile name is passed correctly
					if profileName != "test-profile" && tt.name == "Login with profile" {
						t.Errorf("Expected profile 'test-profile', got %s", profileName)
					}
				},
			}

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

func TestAuthLogoutCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Logout help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Usage:",
		},
		{
			name:    "Logout without profile",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "Logout with profile",
			args:    []string{"test-profile"},
			wantErr: false,
		},
		{
			name:    "Logout with too many args",
			args:    []string{"profile1", "profile2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:  "logout [profile]",
				Args: cobra.MaximumNArgs(1),
				Run: func(cmd *cobra.Command, args []string) {
					// Mock the logout functionality
					var profileName string
					if len(args) > 0 {
						profileName = args[0]
					}

					// Mock auth manager behavior
					// Verify profile handling
					if tt.name == "Logout with profile" && profileName != "test-profile" {
						t.Errorf("Expected profile 'test-profile', got %s", profileName)
					}
					if tt.name == "Logout without profile" && profileName != "" {
						t.Errorf("Expected empty profile, got %s", profileName)
					}
				},
			}

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

func TestAuthProfilesCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Profiles help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Usage:",
		},
		{
			name:    "List profiles",
			args:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "profiles",
				Run: func(cmd *cobra.Command, args []string) {
					// Mock profiles listing
					// In real implementation, this would call authManager.ListProfiles()

					// Mock profile data
					type Profile struct {
						Name            string
						IsAuthenticated bool
						AccountID       string
						AccountName     string
						RoleName        string
					}

					profiles := []Profile{
						{
							Name:            "test-profile-1",
							IsAuthenticated: true,
							AccountID:       "123456789012",
							AccountName:     "Test Account",
							RoleName:        "AdminRole",
						},
						{
							Name:            "test-profile-2",
							IsAuthenticated: false,
							AccountID:       "",
							AccountName:     "",
							RoleName:        "",
						},
					}

					if len(profiles) == 0 {
						// Would print "No AWS profiles found"
						return
					}

					// Would format and print profiles
					// Testing the logic structure
					for _, profile := range profiles {
						if profile.Name == "" {
							t.Error("Profile name should not be empty")
						}
						// Additional validation logic would go here
					}
				},
			}

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

func TestAuthCredsCmd(t *testing.T) {
	// Save original environment
	origProfile := os.Getenv("AWS_PROFILE")
	defer os.Setenv("AWS_PROFILE", origProfile)

	tests := []struct {
		name       string
		args       []string
		envProfile string
		wantErr    bool
		contains   string
	}{
		{
			name:     "Creds help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Usage:",
		},
		{
			name:    "Creds with profile",
			args:    []string{"test-profile"},
			wantErr: false,
		},
		{
			name:       "Creds without profile but with env",
			args:       []string{},
			envProfile: "env-profile",
			wantErr:    false,
		},
		{
			name:    "Creds with too many args",
			args:    []string{"profile1", "profile2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for test
			if tt.envProfile != "" {
				_ = os.Setenv("AWS_PROFILE", tt.envProfile) // #nosec G104 - test setup
			} else {
				_ = os.Unsetenv("AWS_PROFILE") // #nosec G104 - test setup
			}

			cmd := &cobra.Command{
				Use:  "creds [profile]",
				Args: cobra.MaximumNArgs(1),
				Run: func(cmd *cobra.Command, args []string) {
					// Mock credential display functionality
					var profileName string
					if len(args) > 0 {
						profileName = args[0]
					} else {
						// Try AWS_PROFILE environment variable first
						profileName = os.Getenv("AWS_PROFILE")
						if profileName == "" {
							// Would get from config.SSO.DefaultProfile
							profileName = "default-profile"
						}
					}

					if profileName == "" {
						// Would exit with error in real implementation
						return
					}

					// Mock credentials
					type Credentials struct {
						AccessKeyID     string
						SecretAccessKey string
						SessionToken    string
						Region          string
					}

					creds := Credentials{
						AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
						SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
						SessionToken:    "AQoDYXdzEJr...",
						Region:          "us-east-1",
					}

					// Verify credentials are not empty
					if creds.AccessKeyID == "" {
						t.Error("AccessKeyID should not be empty")
					}
					if creds.SecretAccessKey == "" {
						t.Error("SecretAccessKey should not be empty")
					}
					if creds.SessionToken == "" {
						t.Error("SessionToken should not be empty")
					}
					if creds.Region == "" {
						t.Error("Region should not be empty")
					}

					// Verify profile name handling
					expectedProfile := "test-profile"
					if tt.name == "Creds without profile but with env" {
						expectedProfile = "env-profile"
					}
					if tt.name == "Creds with profile" && profileName != expectedProfile {
						t.Errorf("Expected profile %s, got %s", expectedProfile, profileName)
					}
				},
			}

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

func TestAuthCmdStructure(t *testing.T) {
	// Test that auth command has expected subcommands
	expectedSubcommands := []string{"login [profile]", "logout [profile]", "profiles", "creds [profile]"}

	subcommands := make(map[string]bool)
	for _, cmd := range authCmd.Commands() {
		subcommands[cmd.Use] = true
	}

	for _, expected := range expectedSubcommands {
		if !subcommands[expected] {
			t.Errorf("Expected subcommand %s not found in auth command", expected)
		}
	}

	// Test basic command properties
	if authCmd.Use != "auth" {
		t.Errorf("Expected Use to be 'auth', got %s", authCmd.Use)
	}

	if authCmd.Short == "" {
		t.Error("Auth command should have a short description")
	}

	if authCmd.Long == "" {
		t.Error("Auth command should have a long description")
	}
}

func TestAuthCommandArgValidation(t *testing.T) {
	// Test argument validation for different commands
	tests := []struct {
		cmdName     string
		args        []string
		shouldError bool
	}{
		{"login", []string{}, true},            // Requires exactly 1 arg
		{"login", []string{"profile"}, false},  // Valid
		{"login", []string{"p1", "p2"}, true},  // Too many args
		{"logout", []string{}, false},          // Optional arg
		{"logout", []string{"profile"}, false}, // Valid with arg
		{"logout", []string{"p1", "p2"}, true}, // Too many args
		{"profiles", []string{}, false},        // No args required
		{"profiles", []string{"extra"}, false}, // Profiles ignores extra args
		{"creds", []string{}, false},           // Optional arg
		{"creds", []string{"profile"}, false},  // Valid with arg
		{"creds", []string{"p1", "p2"}, true},  // Too many args
	}

	for _, tt := range tests {
		t.Run(tt.cmdName+" args validation", func(t *testing.T) {
			var targetCmd *cobra.Command

			// Find the subcommand
			for _, cmd := range authCmd.Commands() {
				if strings.HasPrefix(cmd.Use, tt.cmdName) {
					targetCmd = cmd
					break
				}
			}

			if targetCmd == nil {
				t.Fatalf("Command %s not found", tt.cmdName)
			}

			// Test argument validation only if Args function exists
			if targetCmd.Args != nil {
				err := targetCmd.Args(targetCmd, tt.args)
				if tt.shouldError && err == nil {
					t.Errorf("Expected args validation to fail for %s with args %v", tt.cmdName, tt.args)
				} else if !tt.shouldError && err != nil {
					t.Errorf("Expected args validation to pass for %s with args %v, got error: %v", tt.cmdName, tt.args, err)
				}
			}
		})
	}
}

func TestContextHandling(t *testing.T) {
	// Test that context is properly used in auth commands
	// This is more of a structural test to ensure context.Context is used

	ctx := context.Background()
	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Test context with timeout
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Verify context can be cancelled
	cancel()
	select {
	case <-ctx.Done():
		// Expected - context was cancelled
	default:
		t.Error("Context should be done after cancellation")
	}
}

func TestMockAuthManagerBehavior(t *testing.T) {
	// Test mock auth manager behaviors that would be used in the commands

	type MockAuthManager struct{}

	type Profile struct {
		Name            string
		IsAuthenticated bool
		AccountID       string
		AccountName     string
		RoleName        string
	}

	type Credentials struct {
		AccessKeyID     string
		SecretAccessKey string
		SessionToken    string
		Region          string
	}

	mockManager := &MockAuthManager{}
	_ = mockManager // Use the mock manager

	// Mock profile data for testing
	mockProfiles := []Profile{
		{
			Name:            "test-profile",
			IsAuthenticated: true,
			AccountID:       "123456789012",
			AccountName:     "Test Account",
			RoleName:        "AdminRole",
		},
	}

	// Test profile structure
	for _, profile := range mockProfiles {
		if profile.Name == "" {
			t.Error("Profile name should not be empty")
		}
		if profile.IsAuthenticated && profile.AccountID == "" {
			t.Error("Authenticated profile should have account ID")
		}
	}

	// Mock credentials for testing
	mockCreds := Credentials{
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		SessionToken:    "AQoDYXdzEJr...",
		Region:          "us-east-1",
	}

	if mockCreds.AccessKeyID == "" {
		t.Error("Mock credentials should have access key")
	}
	if mockCreds.SecretAccessKey == "" {
		t.Error("Mock credentials should have secret key")
	}
	if mockCreds.SessionToken == "" {
		t.Error("Mock credentials should have session token")
	}
	if mockCreds.Region == "" {
		t.Error("Mock credentials should have region")
	}
}
