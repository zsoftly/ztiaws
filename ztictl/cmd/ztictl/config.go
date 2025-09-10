package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"ztictl/internal/config"
	"ztictl/internal/system"
	"ztictl/pkg/aws"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long: `Manage ztictl configuration including initialization, validation, and display.

Examples:
  ztictl config init                    # Initialize configuration
  ztictl config check                   # Check requirements
  ztictl config show                    # Show current config`,
}

// configInitCmd represents the config init command
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long: `Initialize ztictl configuration by creating a configuration file.
This will guide you through an interactive setup process to configure AWS SSO settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		interactive, _ := cmd.Flags().GetBool("interactive")

		if err := initializeConfigFile(force, interactive); err != nil {
			logger.Error("Configuration initialization failed", "error", err)
			os.Exit(1)
		}
	},
}

// configCheckCmd represents the config check command
var configCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check system requirements and configuration",
	Long: `Check that all required dependencies are installed and configured properly.
This includes AWS CLI, Session Manager plugin, and other system requirements.`,
	Run: func(cmd *cobra.Command, args []string) {
		fix, _ := cmd.Flags().GetBool("fix")

		if err := checkRequirements(fix); err != nil {
			logger.Error("Requirements check failed", "error", err)
			os.Exit(1)
		}
	},
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  `Display the current ztictl configuration settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get configuration
		cfg := config.Get()

		if cfg == nil {
			fmt.Println("No configuration loaded")
			return
		}

		// Display configuration
		fmt.Println("\n=== Current Configuration ===")
		fmt.Printf("\nSSO Configuration:\n")
		fmt.Printf("  Start URL: %s\n", cfg.SSO.StartURL)
		fmt.Printf("  Region: %s\n", cfg.SSO.Region)

		fmt.Printf("\nDefaults:\n")
		fmt.Printf("  Default Region: %s\n", cfg.DefaultRegion)

		fmt.Printf("\nLogging:\n")
		fmt.Printf("  Directory: %s\n", cfg.Logging.Directory)
		fmt.Printf("  File Logging: %v\n", cfg.Logging.FileLogging)
		fmt.Printf("  Level: %s\n", cfg.Logging.Level)

		fmt.Printf("\nSystem:\n")
		fmt.Printf("  IAM Propagation Delay: %d seconds\n", cfg.System.IAMPropagationDelay)
		fmt.Printf("  File Size Threshold: %d bytes\n", cfg.System.FileSizeThreshold)
		fmt.Printf("  S3 Bucket Prefix: %s\n", cfg.System.S3BucketPrefix)
		fmt.Printf("  Temp Directory: %s\n", cfg.System.TempDirectory)

		// Display file path
		home, err := os.UserHomeDir()
		if err == nil {
			defaultConfig := filepath.Join(home, ".ztictl.yaml")
			if _, err := os.Stat(defaultConfig); err == nil {
				fmt.Printf("\nConfig File: %s\n", defaultConfig)
			} else {
				fmt.Printf("\nConfig File: Not found (using defaults)\n")
				fmt.Printf("Run 'ztictl config init' to create configuration file\n")
			}
		}
	},
}

// configValidateCmd represents the config validate command
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  `Validate the ztictl configuration file for syntax and required fields.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateConfiguration(); err != nil {
			logger.Error("Configuration validation failed", "error", err)
			os.Exit(1)
		}
	},
}

// configRepairCmd represents the config repair command
var configRepairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repair configuration issues interactively",
	Long: `Detect configuration issues and guide you through fixing them interactively.
This command will identify invalid values and prompt you for correct replacements.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := repairConfiguration(); err != nil {
			logger.Error("Configuration repair failed", "error", err)
			os.Exit(1)
		}
	},
}

// initializeConfigFile handles the config initialization logic and returns errors instead of calling os.Exit
func initializeConfigFile(force, interactive bool) error {
	// Determine config file path
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to find home directory: %w", err)
	}

	configPath := filepath.Join(home, ".ztictl.yaml")

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		// File exists - check if we should overwrite
		if !force {
			fmt.Printf("\n⚠️  Configuration file already exists at %s\n", configPath)
			fmt.Println("Would you like to overwrite it? (yes/no)")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "yes" && response != "y" {
				fmt.Println("Configuration initialization cancelled.")
				return nil
			}
			// User confirmed, proceed as if --force was provided
			force = true
		}

		// Run interactive setup if force is now true (either from flag or user confirmation)
		if interactive || force {
			if err := runInteractiveConfig(configPath); err != nil {
				return fmt.Errorf("interactive configuration failed: %w", err)
			}
			return nil
		}
	}

	// Run interactive setup if requested (for new files)
	if interactive {
		if err := runInteractiveConfig(configPath); err != nil {
			return fmt.Errorf("interactive configuration failed: %w", err)
		}
		return nil
	}

	// Create sample configuration (non-interactive)
	if err := config.CreateSampleConfig(configPath); err != nil {
		return fmt.Errorf("failed to create configuration file: %w", err)
	}

	fmt.Printf("Sample configuration created at %s\n", configPath)
	fmt.Println("Please edit the file with your AWS SSO settings and run 'ztictl auth login' to authenticate.")

	return nil
}

// checkRequirements handles the requirements check logic and returns errors instead of calling os.Exit
func checkRequirements(fix bool) error {
	logger.Info("Checking system requirements...")

	// Create system requirements checker
	checker := system.NewRequirementsChecker(logger)

	// Run all checks
	results, _ := checker.CheckAll()

	// Display results
	allPassed := true
	for _, result := range results {
		if result.Passed {
			logger.Info("✅", "check", result.Name)
		} else {
			logger.Error("❌", "check", result.Name, "error", result.Error)
			if result.Suggestion != "" {
				logger.Info("💡 Fix", "suggestion", result.Suggestion)
			}
			allPassed = false
		}
	}

	fmt.Println()

	// Display configuration status
	cfg := config.Get()
	if cfg != nil && cfg.SSO.StartURL != "" {
		logger.Info("Configuration loaded", "sso_start_url", cfg.SSO.StartURL)
	} else {
		logger.Warn("No SSO configuration found. Run 'ztictl config init' to set up.")
	}

	if allPassed {
		logger.Info("All requirements met! ✅")
		return nil
	} else {
		logger.Error("Some requirements are not met")

		if fix {
			logger.Info("Attempting to fix issues...")
			if err := checker.FixIssues(results); err != nil {
				return fmt.Errorf("failed to fix some issues: %w", err)
			}
			logger.Info("Issues fixed successfully. Please run check again to verify.")
			return nil
		} else {
			logger.Info("Run with --fix to attempt automatic fixes")
			return fmt.Errorf("some requirements are not met")
		}
	}
}

// validateConfiguration handles the config validation logic and returns errors instead of calling os.Exit
func validateConfiguration() error {
	logger.Info("Validating configuration...")

	// Re-load configuration to check for errors
	if err := config.Load(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	cfg := config.Get()

	// Perform additional validation
	var errors []string

	if cfg.SSO.StartURL != "" {
		if cfg.SSO.Region == "" {
			errors = append(errors, "SSO region is required when SSO start URL is provided")
		}
	}

	if len(errors) > 0 {
		logger.Error("Configuration validation failed:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("configuration validation failed with %d errors", len(errors))
	}

	logger.Info("Configuration validation passed ✅")
	return nil
}

// repairConfiguration guides user through fixing configuration issues
func repairConfiguration() error {
	logger.Info("Checking configuration for issues...")

	// Try to load config with validation errors allowed
	valErr, err := config.LoadWithOptions(true)
	if err != nil && valErr == nil {
		// Fatal error that's not a validation error
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if valErr == nil {
		logger.Info("No configuration issues found ✅")
		return nil
	}

	// We have a validation error
	logger.Error("Found configuration issue", "error", valErr.Error())
	fmt.Println()
	fmt.Printf("The %s has an invalid value: '%s'\n", valErr.Field, valErr.Value)
	fmt.Printf("Error: %s\n\n", valErr.Message)

	// Prompt for interactive fix
	fmt.Println("Would you like to fix this interactively? (yes/no)")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "yes" && response != "y" {
		fmt.Println("\nYou can manually edit the configuration file at ~/.ztictl.yaml")
		fmt.Printf("Fix the %s field which currently has value: %s\n", valErr.Field, valErr.Value)
		return fmt.Errorf("repair cancelled by user")
	}

	// Get the correct value from user
	var newValue string
	for {
		switch valErr.Field {
		case "SSO region", "Default region":
			fmt.Printf("\nEnter a valid AWS region (e.g., us-east-1, ca-central-1): ")
		case "SSO start URL":
			// First check if current value looks like it already has a domain ID we can extract
			if strings.Contains(valErr.Value, ".awsapps.com") {
				// Extract domain ID from existing URL
				parts := strings.Split(valErr.Value, "//")
				if len(parts) > 1 {
					domainPart := strings.Split(parts[1], ".awsapps.com")[0]
					fmt.Printf("\nDetected domain ID: %s\n", domainPart)
				}
			}
			fmt.Printf("\nEnter your AWS SSO domain ID (e.g., d-1234567890 or zsoftly): ")
		}

		newValue, _ = reader.ReadString('\n')
		newValue = strings.TrimSpace(newValue)

		if newValue == "" {
			fmt.Println("Value cannot be empty. Please try again.")
			continue
		}

		// Validate the new value
		switch valErr.Field {
		case "SSO region", "Default region":
			if !aws.IsValidAWSRegion(newValue) {
				fmt.Printf("'%s' is not a valid AWS region format. Expected format: xx-xxxx-n\n", newValue)
				continue
			}
		case "SSO start URL":
			// Build full URL from domain ID
			if !strings.HasPrefix(newValue, "https://") {
				newValue = fmt.Sprintf("https://%s.awsapps.com/start", newValue)
			}
			// Validate the constructed URL
			if !strings.HasPrefix(newValue, "https://") {
				fmt.Println("Invalid SSO URL format")
				continue
			}
		}

		break
	}

	// Read and parse the config file using YAML parser
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to find home directory: %w", err)
	}
	configPath := filepath.Join(home, ".ztictl.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into a map to preserve structure and comments
	var configData map[string]interface{}
	if err := yaml.Unmarshal(data, &configData); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply the fix based on the field
	switch valErr.Field {
	case "SSO region":
		// Update SSO region in the parsed data
		if ssoData, ok := configData["sso"].(map[string]interface{}); ok {
			ssoData["region"] = newValue
		} else {
			return fmt.Errorf("SSO configuration not found in config file")
		}
	case "Default region":
		// Update default region in the parsed data
		configData["default_region"] = newValue
	case "SSO start URL":
		// Update SSO start URL in the parsed data
		if ssoData, ok := configData["sso"].(map[string]interface{}); ok {
			ssoData["start_url"] = newValue
		} else {
			return fmt.Errorf("SSO configuration not found in config file")
		}
	default:
		return fmt.Errorf("unknown field: %s", valErr.Field)
	}

	// Marshal back to YAML
	updatedData, err := yaml.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	// Write the fixed config back
	if err := os.WriteFile(configPath, updatedData, 0600); err != nil {
		return fmt.Errorf("failed to write fixed config: %w", err)
	}

	logger.Info("Configuration updated", "field", valErr.Field, "old", valErr.Value, "new", newValue)
	logger.Info("Verifying fixed configuration...")

	// Reload and validate the fixed config
	if err := config.Load(); err != nil {
		logger.Error("Configuration still has issues after repair", "error", err)
		fmt.Println("\nThere are more configuration issues. Please run 'ztictl config repair' again.")
		return nil
	}

	logger.Info("Configuration is now valid! ✅")
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configCheckCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configRepairCmd)

	// Add flags
	configInitCmd.Flags().BoolP("force", "f", false, "Overwrite existing configuration file")
	configInitCmd.Flags().BoolP("interactive", "i", false, "Interactive configuration setup")
	configCheckCmd.Flags().BoolP("fix", "", false, "Attempt to automatically fix issues")
}
