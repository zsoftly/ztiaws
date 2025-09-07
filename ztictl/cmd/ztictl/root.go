package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ztictl/internal/config"
	"ztictl/internal/splash"
	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version represents the current version of ztictl
	// This can be set at build time using -ldflags "-X main.version=X.Y.Z"
	// Default version is "2.5.0"; override at build time with -ldflags "-X main.Version=X.Y.Z"
	Version    = "2.5.1"
	configFile string
	debug      bool
	showSplash bool
	logger     *logging.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ztictl",
	Short: "Unified AWS SSO authentication and Systems Manager CLI tool",
	Long: `ztictl is a unified CLI tool that replaces the bash-based authaws and ssm scripts.
It provides seamless AWS SSO authentication and comprehensive AWS Systems Manager operations
through a single, cross-platform binary.

Features:
- AWS SSO authentication with interactive account/role selection
- SSM-based instance discovery and management
- Session Manager connections
- Remote command execution via SSM
- File transfer through SSM (with S3 for large files)
- Port forwarding through SSM tunnels
- Multi-region support`,
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip splash for help and version commands
		if cmd.Name() == "help" || cmd.Name() == "version" || cmd.Parent() == nil {
			return
		}

		// Force splash screen if --show-splash flag is used
		var showedSplash bool
		var err error

		if showSplash {
			// Force splash screen by temporarily removing version file
			homeDir, _ := os.UserHomeDir()
			versionFile := filepath.Join(homeDir, ".ztictl_version")
			tempFile := versionFile + ".backup"

			// Backup existing version file if it exists
			if _, err := os.Stat(versionFile); err == nil {
				os.Rename(versionFile, tempFile)
			}

			// Show splash as first run
			showedSplash, err = splash.ShowSplash(Version)

			// Restore version file
			if _, err := os.Stat(tempFile); err == nil {
				os.Rename(tempFile, versionFile)
			}
		} else {
			// Normal splash behavior
			showedSplash, err = splash.ShowSplash(Version)
		}

		if err != nil {
			// Don't exit, just continue - splash errors are not critical
			return
		}

		// If this is the first run, show helpful message instead of automatic setup
		if showedSplash {
			cfg := config.Get()
			if cfg != nil && cfg.SSO.StartURL == "" {
				fmt.Println("\n🚀 Welcome to ztictl!")
				fmt.Println("To get started, run: ztictl config init")
				fmt.Println("Then authenticate with: ztictl auth login")
				fmt.Println()
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.ztictl.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output")
	rootCmd.PersistentFlags().BoolVar(&showSplash, "show-splash", false, "force display of welcome splash screen")

	// Bind flags to viper
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Initialize logger with our adapter
	logger = logging.NewLogger(debug)

	// Perform configuration setup and handle any errors
	if err := setupConfiguration(); err != nil {
		logger.Error("Configuration setup failed", "error", err)
		os.Exit(1)
	}
}

// setupConfiguration handles the actual configuration logic and returns errors
// instead of calling os.Exit directly, improving testability and separation of concerns
func setupConfiguration() error {
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to find home directory: %w", err)
		}

		// Search config in home directory with name ".ztictl" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ztictl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if debug {
			logger.Debug("Using config file", "file", viper.ConfigFileUsed())
		}
	}

	// Load configuration
	if err := config.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	return nil
}

// GetLogger returns a compatibility logger instance
func GetLogger() *logging.Logger {
	if logger == nil {
		logger = logging.NewLogger(debug)
	}
	return logger
}

// runInteractiveConfig prompts the user for configuration values
func runInteractiveConfig(configPath string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n🔧 Interactive Configuration Setup")
	fmt.Println("==================================")
	fmt.Println("Let's configure ztictl with your AWS SSO settings.")

	// Get SSO Start URL
	fmt.Print("Enter your AWS SSO start URL (e.g., https://yourcompany.awsapps.com/start): ")
	startURL, _ := reader.ReadString('\n')
	startURL = strings.TrimSpace(startURL)

	if startURL == "" {
		return fmt.Errorf("SSO start URL cannot be empty")
	}

	// Get SSO Region
	fmt.Print("Enter your AWS SSO region (e.g., us-east-1) [us-east-1]: ")
	ssoRegion, _ := reader.ReadString('\n')
	ssoRegion = strings.TrimSpace(ssoRegion)
	if ssoRegion == "" {
		ssoRegion = "us-east-1"
	}

	// Get Default Region
	fmt.Print("Enter your default AWS region (e.g., us-east-1) [us-east-1]: ")
	defaultRegion, _ := reader.ReadString('\n')
	defaultRegion = strings.TrimSpace(defaultRegion)
	if defaultRegion == "" {
		defaultRegion = "us-east-1"
	}

	// Get Profile Name
	fmt.Print("Enter a profile name [default-sso-profile]: ")
	profileName, _ := reader.ReadString('\n')
	profileName = strings.TrimSpace(profileName)
	if profileName == "" {
		profileName = "default-sso-profile"
	}

	// Get home directory for Windows-compatible paths
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to get home directory: %w", err)
	}

	// Use platform-appropriate log directory
	logDir := filepath.Join(home, "logs")

	// Use platform-appropriate temp directory
	tempDir := os.TempDir()

	// Create configuration
	configContent := fmt.Sprintf(`# ztictl Configuration File
# This file contains configuration for AWS SSO and ztictl behavior

# AWS SSO Configuration
sso:
  start_url: "%s"
  region: "%s"
  default_profile: "%s"

# Default AWS region for operations
default_region: "%s"

# Logging configuration
logging:
  directory: "%s"
  file_logging: true
  level: "info"

# System configuration
system:
  session_manager_plugin_path: ""  # Auto-detected if empty
  temp_directory: "%s"
  
# Region shortcuts for convenience
region_shortcuts:
  cac1: "ca-central-1"
  use1: "us-east-1" 
  use2: "us-east-2"
  usw1: "us-west-1"
  usw2: "us-west-2"
  euw1: "eu-west-1"
  euw2: "eu-west-2"
  euc1: "eu-central-1"
  apne1: "ap-northeast-1"
  apne2: "ap-northeast-2"
  apse1: "ap-southeast-1"
  apse2: "ap-southeast-2"
  aps1: "ap-south-1"
`, startURL, ssoRegion, profileName, defaultRegion, logDir, tempDir)

	// Write configuration file
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Printf("\n✅ Configuration saved to: %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'ztictl config check' to verify system requirements")
	fmt.Println("2. Run 'ztictl auth login' to authenticate with AWS SSO")
	fmt.Println("3. Run 'ztictl ssm list' to see your EC2 instances")

	return nil
}
