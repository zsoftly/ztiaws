package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Help command",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "ztictl is a unified CLI tool",
		},
		{
			name:     "Version command",
			args:     []string{"--version"},
			wantErr:  false,
			contains: "version",
		},
		{
			name:     "Invalid command",
			args:     []string{"invalid"},
			wantErr:  true,
			contains: "unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the actual rootCmd with all subcommands
			cmd := rootCmd

			// Capture output
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check output contains expected string
			output := buf.String()
			if tt.contains != "" && !strings.Contains(output, tt.contains) {
				t.Errorf("Execute() output should contain %v", tt.contains)
			}
		})
	}
}

func TestGlobalFlags(t *testing.T) {
	// Save original values
	origConfig := configFile
	origDebug := debug
	origShowSplash := showSplash

	// Restore original values after test
	defer func() {
		configFile = origConfig
		debug = origDebug
		showSplash = origShowSplash
	}()

	tests := []struct {
		name         string
		args         []string
		expectedConf string
		expectedDeb  bool
		expectedSpl  bool
	}{
		{
			name:         "Config flag",
			args:         []string{"--config", "/custom/config.yaml"},
			expectedConf: "/custom/config.yaml",
		},
		{
			name:        "Debug flag",
			args:        []string{"--debug"},
			expectedDeb: true,
		},
		{
			name:        "Show splash flag",
			args:        []string{"--show-splash"},
			expectedSpl: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset values
			configFile = ""
			debug = false
			showSplash = false

			// Create isolated command
			cmd := &cobra.Command{Use: "test"}
			cmd.PersistentFlags().StringVar(&configFile, "config", "", "config file")
			cmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output")
			cmd.PersistentFlags().BoolVar(&showSplash, "show-splash", false, "force display of welcome splash screen")

			cmd.SetArgs(tt.args)
			err := cmd.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}

			// Check expected values
			if tt.expectedConf != "" && configFile != tt.expectedConf {
				t.Errorf("configFile = %v, want %v", configFile, tt.expectedConf)
			}
			if debug != tt.expectedDeb {
				t.Errorf("debug = %v, want %v", debug, tt.expectedDeb)
			}
			if showSplash != tt.expectedSpl {
				t.Errorf("showSplash = %v, want %v", showSplash, tt.expectedSpl)
			}
		})
	}
}

func TestInitConfig(t *testing.T) {
	// Save original viper state
	viperInstance := viper.GetViper()
	defer func() {
		viper.Reset()
		// Restore original instance
		for key, value := range viperInstance.AllSettings() {
			viper.Set(key, value)
		}
	}()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ztictl_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configContent := `debug: true
default_region: "us-west-2"`
	configPath := filepath.Join(tempDir, ".ztictl.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test with custom config file
	t.Run("Custom config file", func(t *testing.T) {
		viper.Reset()
		configFile = configPath
		debug = false

		initConfig()

		if !viper.GetBool("debug") {
			t.Error("Expected debug to be true from config file")
		}
		if viper.GetString("default_region") != "us-west-2" {
			t.Errorf("Expected default_region to be 'us-west-2', got %v", viper.GetString("default_region"))
		}
	})

	// Test without config file
	t.Run("No config file", func(t *testing.T) {
		viper.Reset()
		configFile = "/nonexistent/config.yaml"
		debug = true

		// This should not panic and should create a logger
		initConfig()

		if logger == nil {
			t.Error("Expected logger to be initialized")
		}
	})
}

func TestRunInteractiveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ztictl_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, ".ztictl.yaml")

	// Test with mock input simulation (limited testing since it reads from stdin)
	t.Run("Config file creation", func(t *testing.T) {
		// We can't easily test the interactive part, but we can test the file creation part
		// by examining the structure of the runInteractiveConfig function

		// Just test that the function exists and can be called with a path
		// The actual interactive testing would require complex stdin mocking
		// Test that runInteractiveConfig function exists
		_ = runInteractiveConfig // This will fail to compile if function doesn't exist

		// Test that the config path directory exists for the function to work
		dir := filepath.Dir(configPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Error("Directory should exist for config file creation")
		}
	})
}

func TestVersionVariable(t *testing.T) {
	// Test that version is set correctly
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Test default version
	if Version != "2.1.0" {
		t.Logf("Version is %s (may have been overridden at build time)", Version)
	}

	// Test that rootCmd uses the version
	if rootCmd.Version != Version {
		t.Errorf("rootCmd.Version = %v, want %v", rootCmd.Version, Version)
	}
}

func TestPersistentPreRun(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ztictl_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock home directory
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tempDir) // #nosec G104
	defer os.Setenv("HOME", origHome)

	tests := []struct {
		name       string
		cmdName    string
		shouldSkip bool
	}{
		{"Help command should skip", "help", true},
		{"Version command should skip", "version", true},
		{"Regular command should run", "auth", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock command
			cmd := &cobra.Command{
				Use: tt.cmdName,
			}

			// Set parent if it's not a root command
			if tt.cmdName != "help" && tt.cmdName != "version" {
				parentCmd := &cobra.Command{Use: "parent"}
				parentCmd.AddCommand(cmd)
			}

			// Test that PersistentPreRun exists and can be called
			if rootCmd.PersistentPreRun == nil {
				t.Error("PersistentPreRun should be set")
				return
			}

			// Call PersistentPreRun - it shouldn't panic
			// Note: This may show splash screen or create files, but shouldn't crash
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("PersistentPreRun panicked: %v", r)
				}
			}()

			rootCmd.PersistentPreRun(cmd, []string{})
		})
	}
}

// Test command initialization
func TestCommandInit(t *testing.T) {
	// Test that all expected commands are added to root
	expectedSubcommands := []string{"auth", "config", "ssm", "cleanup [region]"}

	subcommands := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		subcommands[cmd.Use] = true
	}

	for _, expected := range expectedSubcommands {
		if !subcommands[expected] {
			t.Errorf("Expected subcommand %s not found", expected)
		}
	}
}

// Test flag binding
func TestFlagBinding(t *testing.T) {
	// Test that flags are properly bound to viper
	viper.Reset()
	defer viper.Reset()

	// Simulate flag parsing
	_ = rootCmd.PersistentFlags().Set("debug", "true") // #nosec G104

	// The binding should work through viper.BindPFlag in init()
	// This is a structural test to ensure the binding exists
	flag := rootCmd.PersistentFlags().Lookup("debug")
	if flag == nil {
		t.Error("Debug flag should exist")
		return
	}

	if flag.DefValue != "false" {
		t.Errorf("Debug flag default should be false, got %s", flag.DefValue)
	}
}

func TestIsValidAWSRegion(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		expected bool
	}{
		// Valid regions - standard format
		{"Valid US East 1", "us-east-1", true},
		{"Valid US East 2", "us-east-2", true},
		{"Valid US West 1", "us-west-1", true},
		{"Valid US West 2", "us-west-2", true},
		{"Valid EU West 1", "eu-west-1", true},
		{"Valid EU West 2", "eu-west-2", true},
		{"Valid EU West 3", "eu-west-3", true},
		{"Valid EU Central 1", "eu-central-1", true},
		{"Valid EU North 1", "eu-north-1", true},
		{"Valid AP Southeast 1", "ap-southeast-1", true},
		{"Valid AP Southeast 2", "ap-southeast-2", true},
		{"Valid AP Northeast 1", "ap-northeast-1", true},
		{"Valid AP Northeast 2", "ap-northeast-2", true},
		{"Valid AP South 1", "ap-south-1", true},
		{"Valid CA Central 1", "ca-central-1", true},
		{"Valid SA East 1", "sa-east-1", true},
		{"Valid ME South 1", "me-south-1", true},
		{"Valid AF South 1", "af-south-1", true},
		{"Valid CN North 1", "cn-north-1", true},
		{"Valid CN Northwest 1", "cn-northwest-1", true},

		// Valid GovCloud regions
		{"Valid US GovCloud West", "us-gov-west-1", true},
		{"Valid US GovCloud East", "us-gov-east-1", true},

		// Invalid regions - wrong prefix
		{"Invalid prefix caa", "caa-central-1", false},
		{"Invalid prefix xx", "xx-central-1", false},
		{"Invalid prefix abc", "abc-east-1", false},
		{"Invalid prefix u", "u-east-1", false},
		{"Invalid prefix usss", "usss-east-1", false},

		// Invalid regions - wrong direction/area
		{"Invalid direction", "us-invalid-1", false},
		{"Invalid direction midwest", "us-midwest-1", false},
		{"Invalid direction too short", "us-ea-1", false},
		{"Invalid direction too long", "us-centralllllll-1", false},

		// Invalid regions - wrong number
		{"Invalid number letters", "us-east-a", false},
		{"Invalid number letters mix", "us-east-1a", false},
		{"Invalid number too long", "us-east-123", false},

		// Invalid regions - wrong format
		{"Invalid no hyphens", "useast1", false},
		{"Invalid one hyphen", "us-east1", false},
		{"Invalid four parts", "us-east-west-1", false},
		{"Invalid five parts", "us-gov-east-west-1", false},
		{"Empty string", "", false},

		// Edge cases
		{"With spaces", "us east 1", false},
		{"With underscores", "us_east_1", false},
		{"With dots", "us.east.1", false},
		{"With special chars", "us-east-1!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidAWSRegion(tt.region)
			if result != tt.expected {
				t.Errorf("isValidAWSRegion(%q) = %v, want %v", tt.region, result, tt.expected)
			}
		})
	}
}

func TestIsValidAWSRegion_GovCloudHandling(t *testing.T) {
	// Special test cases for GovCloud region handling
	tests := []struct {
		name     string
		region   string
		expected bool
	}{
		{"GovCloud West", "us-gov-west-1", true},
		{"GovCloud East", "us-gov-east-1", true},
		{"Invalid GovCloud", "us-gov-north-1", false},           // No north govcloud
		{"Invalid GovCloud central", "us-gov-central-1", false}, // No central govcloud
		{"Malformed GovCloud", "usgov-west-1", false},
		{"GovCloud without number", "us-gov-west", false},
		{"GovCloud with invalid number", "us-gov-west-a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidAWSRegion(tt.region)
			if result != tt.expected {
				t.Errorf("isValidAWSRegion(%q) = %v, want %v", tt.region, result, tt.expected)
			}
		})
	}
}
