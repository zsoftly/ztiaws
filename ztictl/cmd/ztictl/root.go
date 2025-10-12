package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"ztictl/internal/config"
	"ztictl/internal/splash"
	"ztictl/pkg/logging"
	"ztictl/pkg/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Version represents the current version of ztictl
	// This can be set at build time using -ldflags "-X main.version=X.Y.Z"
	// Default version is "2.5.0"; override at build time with -ldflags "-X main.Version=X.Y.Z"
	Version    = "2.10.0"
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
	Run: func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			version.PrintVersionWithCheck(Version)
			return
		}
		_ = cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip splash for help, version, and completion commands
		if cmd.Name() == "help" || cmd.Name() == "version" || cmd.Name() == cobra.ShellCompRequestCmd || cmd.Parent() == nil {
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
				_ = os.Rename(versionFile, tempFile) // Ignore error as backup is optional
			}

			// Show splash as first run
			showedSplash, err = splash.ShowSplash(Version)

			// Restore version file
			if _, err := os.Stat(tempFile); err == nil {
				_ = os.Rename(tempFile, versionFile) // Ignore error as restore is optional
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
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")) // #nosec G104

	// Disable Cobra's default completion command in favor of our custom one
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Initialize logger with our adapter
	logger = logging.NewLogger(debug)

	// Perform configuration setup and handle any errors
	if err := setupConfiguration(); err != nil {
		// Check if we're running a config command
		if len(os.Args) > 1 {
			// Check for config command or its subcommands
			if os.Args[1] == "config" ||
				(len(os.Args) > 2 && os.Args[1] == "config" &&
					(os.Args[2] == "init" || os.Args[2] == "repair" || os.Args[2] == "check")) {
				logger.Warn("Configuration has errors, but allowing config command to run", "error", err)
				// Load config with invalid values allowed for config commands
				_, _ = config.LoadWithOptions(true) // Ignore error, we're in repair mode
				return
			}
		}

		// For other commands, show helpful error message
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "validation") {
			logger.Error("Configuration validation failed", "error", err)
			logger.Info("Run 'ztictl config init --interactive' to fix configuration")
		} else {
			logger.Error("Configuration setup failed", "error", err)
		}
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
		home, err := getUserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to find home directory: %w", err)
		}

		logger.Debug("Home directory", "path", home)
		logger.Debug("Looking for config file", "name", ".ztictl.yaml", "paths", []string{home, "."})

		// Search config in home directory with name ".ztictl" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ztictl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// Log the error but don't fail - config.Load() will handle missing config
		logger.Debug("Could not read config file", "error", err)
		// Try to be more specific about what went wrong
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			logger.Debug("Config file not found in expected locations")
		} else {
			logger.Error("Error reading config file", "error", err)
		}
	} else {
		logger.Debug("Using config file", "file", viper.ConfigFileUsed())
	}

	// Load configuration
	if err := config.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	return nil
}

// getUserHomeDir returns the user home directory, respecting environment variables for testing
func getUserHomeDir() (string, error) {
	// Check environment variables first (for test isolation)
	if runtime.GOOS == "windows" {
		if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
			return userProfile, nil
		}
	} else {
		if home := os.Getenv("HOME"); home != "" {
			return home, nil
		}
	}
	// Fall back to os.UserHomeDir() for normal operation
	return os.UserHomeDir()
}

// runInteractiveConfig prompts the user for configuration values
func runInteractiveConfig(configPath string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n🔧 Interactive Configuration Setup")
	fmt.Println("==================================")
	fmt.Println("Let's configure ztictl with your AWS SSO settings.")

	// Get SSO Domain ID and build the URL
	var startURL string
	for {
		fmt.Print("Enter your AWS SSO domain ID (e.g., d-1234567890 or company-name): ")
		input, _ := reader.ReadString('\n')
		domainID := strings.TrimSpace(input)

		if domainID == "" {
			fmt.Println("❌ Domain ID cannot be empty")
			continue
		}

		// Build the full SSO URL from the domain ID
		startURL = fmt.Sprintf("https://%s.awsapps.com/start", domainID)

		// Validate the constructed URL
		if strings.Contains(domainID, "//") || strings.Contains(domainID, ".") {
			fmt.Println("❌ Invalid domain ID. Please enter only the domain identifier, not a full URL")
			fmt.Println("   Example: 'd-1234567890' or 'mycompany', not 'https://mycompany.awsapps.com/start'")
			continue
		}

		fmt.Printf("✅ SSO URL will be: %s\n", startURL)
		break
	}

	// Get SSO Region with validation
	var ssoRegion string
	for {
		fmt.Print("Enter your AWS SSO region [ca-central-1]: ")
		input, _ := reader.ReadString('\n')
		ssoRegion = strings.TrimSpace(input)
		if ssoRegion == "" {
			ssoRegion = "ca-central-1"
		}

		// Validate region format (xx-xxxx-n)
		if !isValidAWSRegion(ssoRegion) {
			fmt.Printf("❌ Invalid AWS region format: %s\n", ssoRegion)
			fmt.Println("Region format should be like: us-east-1, ca-central-1, eu-west-2, ap-southeast-1")
			continue
		}
		break
	}

	// Get Default Region with validation
	var defaultRegion string
	for {
		fmt.Print("Enter your default AWS region [ca-central-1]: ")
		input, _ := reader.ReadString('\n')
		defaultRegion = strings.TrimSpace(input)
		if defaultRegion == "" {
			defaultRegion = "ca-central-1"
		}

		// Validate region format (xx-xxxx-n)
		if !isValidAWSRegion(defaultRegion) {
			fmt.Printf("❌ Invalid AWS region format: %s\n", defaultRegion)
			fmt.Println("Region format should be like: us-east-1, ca-central-1, eu-west-2, ap-southeast-1")
			continue
		}
		break
	}

	// Note: Profile name removed - it will be specified during auth login

	// Get home directory for Windows-compatible paths
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to get home directory: %w", err)
	}

	// Use platform-appropriate log directory with forward slashes for YAML compatibility
	logDir := filepath.Join(home, "logs")
	logDir = filepath.ToSlash(logDir) // Convert to forward slashes for YAML

	// Use platform-appropriate temp directory with forward slashes for YAML compatibility
	tempDir := os.TempDir()
	tempDir = filepath.ToSlash(tempDir) // Convert to forward slashes for YAML

	// Create configuration
	configContent := fmt.Sprintf(`# ztictl Configuration File
# This file contains configuration for AWS SSO and ztictl behavior

# AWS SSO Configuration
sso:
  start_url: "%s"
  region: "%s"

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
`, startURL, ssoRegion, defaultRegion, logDir, tempDir)

	// Write configuration file
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Printf("\n✅ Configuration saved to: %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'ztictl config check' to verify system requirements")
	fmt.Println("2. Run 'ztictl auth login <profile-name>' to authenticate with AWS SSO")
	fmt.Println("3. Run 'ztictl ssm list' to see your EC2 instances")

	return nil
}

// isValidAWSRegion checks if a region string follows AWS region format
func isValidAWSRegion(region string) bool {
	// AWS region format: {area}-{subarea}-{number}
	// Examples: us-east-1, eu-west-2, ap-southeast-1, ca-central-1
	if region == "" {
		return false
	}

	// Must contain at least two hyphens
	parts := strings.Split(region, "-")

	// Handle special case for us-gov regions (4 parts)
	if len(parts) == 4 && parts[0] == "us" && parts[1] == "gov" {
		// us-gov-east-1, us-gov-west-1
		parts = []string{"us-gov", parts[2], parts[3]}
	}

	// Now must have exactly 3 parts
	if len(parts) != 3 {
		return false
	}

	// First part: valid region codes only
	validPrefixes := map[string]bool{
		"us": true, "eu": true, "ap": true, "ca": true,
		"sa": true, "me": true, "af": true, "cn": true,
		"us-gov": true, // Special case for GovCloud
	}

	if !validPrefixes[parts[0]] {
		return false
	}

	// Second part: valid direction/area names
	// Special case for GovCloud - only east and west are valid
	if parts[0] == "us-gov" {
		if parts[1] != "east" && parts[1] != "west" {
			return false
		}
	} else {
		validDirections := map[string]bool{
			"east": true, "west": true, "north": true, "south": true,
			"central": true, "northeast": true, "southeast": true, "northwest": true, "southwest": true,
		}

		if !validDirections[parts[1]] {
			return false
		}
	}

	// Third part: must be a number (1-99)
	if len(parts[2]) < 1 || len(parts[2]) > 2 {
		return false
	}

	// Check if third part is a number
	for _, char := range parts[2] {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
