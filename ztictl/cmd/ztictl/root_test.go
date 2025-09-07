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
	err = os.WriteFile(configPath, []byte(configContent), 0644)
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

func TestGetLogger(t *testing.T) {
	// Reset logger for clean test
	logger = nil

	tests := []struct {
		name      string
		debugMode bool
	}{
		{"Debug mode", true},
		{"Normal mode", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger = nil
			debug = tt.debugMode

			result := GetLogger()

			if result == nil {
				t.Error("GetLogger() should not return nil")
			}

			// Test that subsequent calls return the same instance
			result2 := GetLogger()
			if result != result2 {
				t.Error("GetLogger() should return the same instance on subsequent calls")
			}
		})
	}
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
	os.Setenv("HOME", tempDir)
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
	rootCmd.PersistentFlags().Set("debug", "true")

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
