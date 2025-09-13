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

	// Track failed requirements for summary
	var criticalFailures []system.RequirementResult
	var optionalFailures []system.RequirementResult
	var warningFailures []system.RequirementResult

	// Display results
	allPassed := true
	for _, result := range results {
		if result.Passed {
			logger.Info("✅", "check", result.Name)
			if result.Version != "" {
				logger.Info("   Version", "version", result.Version)
			}
		} else {
			// AWS Credentials is a warning, not an error
			if result.Name == "AWS Credentials" {
				logger.Warn("⚠️ ", "check", result.Name, "warning", result.Error)
				if result.Suggestion != "" {
					logger.Info("   💡", "fix", result.Suggestion)
				}
				warningFailures = append(warningFailures, result)
				// Don't set allPassed to false for warnings
			} else {
				logger.Error("❌", "check", result.Name, "error", result.Error)
				if result.Suggestion != "" {
					logger.Info("   💡", "fix", result.Suggestion)
				}
				allPassed = false

				// Categorize failures
				if result.Name == "Go Version" {
					optionalFailures = append(optionalFailures, result)
				} else {
					criticalFailures = append(criticalFailures, result)
				}
			}
		}
	}

	fmt.Println()

	// Display configuration status
	cfg := config.Get()
	configExists := false
	configValid := true

	// First, try to load config with validation to check for errors
	if err := config.Load(); err != nil {
		// Check if this is a validation error vs a system/environment error
		if strings.Contains(err.Error(), "invalid configuration") || strings.Contains(err.Error(), "placeholder") {
			logger.Error("❌ Configuration", "status", "Invalid configuration detected")
			logger.Info("   💡", "fix", "Run 'ztictl config repair' to fix configuration issues")
			configValid = false
			configExists = true // Config file exists but has errors
		} else {
			// System/environment error (e.g., CI environment, missing home dir)
			// Treat as no config found and continue gracefully
			configValid = false
			configExists = false
		}
	} else if cfg != nil && cfg.SSO.StartURL != "" {
		// Check if it's a placeholder URL
		if aws.IsPlaceholderSSOURL(cfg.SSO.StartURL) {
			logger.Error("❌ Configuration", "status", "Placeholder URL detected", "sso_url", cfg.SSO.StartURL)
			logger.Info("   💡", "fix", "Run 'ztictl config init' to set up your actual SSO URL")
			configValid = false
		} else {
			logger.Info("✅ Configuration", "status", "Loaded", "sso_url", cfg.SSO.StartURL)
		}
		configExists = true
	} else {
		logger.Error("❌ Configuration", "status", "Not found or incomplete")
		logger.Info("   💡", "fix", "Run 'ztictl config init' to set up SSO configuration")
		configValid = false
	}

	fmt.Println()

	if allPassed && configExists && configValid {
		logger.Info("🎉 All requirements met! You're ready to use ztictl")
		fmt.Println("\n📝 Next steps:")
		fmt.Println("  1. Authenticate: ztictl auth login")
		fmt.Println("  2. List instances: ztictl ssm list")
		fmt.Println("  3. Connect to instance: ztictl ssm connect <instance-id>")
		return nil
	}

	// Provide detailed action plan
	if len(criticalFailures) > 0 || !configExists || !configValid {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("📋 ACTION REQUIRED - Follow these steps:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		stepNumber := 1

		// Configuration setup first if needed
		if !configExists {
			fmt.Printf("\n%d. SET UP CONFIGURATION:\n", stepNumber)
			fmt.Println("   Run: ztictl config init --interactive")
			fmt.Println("   This will guide you through setting up AWS SSO")
			stepNumber++
		}

		// Critical requirements
		for _, failure := range criticalFailures {
			fmt.Printf("\n%d. INSTALL %s:\n", stepNumber, strings.ToUpper(failure.Name))

			switch failure.Name {
			case "AWS CLI":
				fmt.Println("   Option A: Use official installer")
				fmt.Println("   Visit: https://aws.amazon.com/cli/")
				fmt.Println("   ")
				fmt.Println("   Option B: Platform-specific install:")
				fmt.Println("   - macOS: brew install awscli")
				fmt.Println("   - Ubuntu/Debian: sudo apt install awscli")
				fmt.Println("   - Windows: Use MSI installer from AWS website")

			case "Session Manager Plugin":
				fmt.Println("   Option A: Quick install")
				if fix {
					fmt.Println("   Run: ztictl config check --fix")
				} else {
					fmt.Printf("   %s\n", failure.Suggestion)
				}
				fmt.Println("   ")
				fmt.Println("   Option B: Manual install")
				fmt.Println("   Visit: https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html")

			case "jq":
				fmt.Println("   Option A: Quick install")
				if fix {
					fmt.Println("   Run: ztictl config check --fix")
				} else {
					fmt.Printf("   %s\n", failure.Suggestion)
				}
				fmt.Println("   ")
				fmt.Println("   Option B: Download binary")
				fmt.Println("   Visit: https://stedolan.github.io/jq/download/")

				// AWS Credentials is now handled as a warning, not critical
			}
			stepNumber++
		}

		// Warnings (like AWS credentials)
		if len(warningFailures) > 0 {
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("⚠️  WARNINGS (recommended but not required):")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			for _, failure := range warningFailures {
				if failure.Name == "AWS Credentials" {
					fmt.Println("\n• AWS Credentials: Your AWS credentials are missing or expired")
					fmt.Println("  ")
					fmt.Println("  Option A: Use ztictl (recommended)")
					fmt.Println("  Run: ztictl auth login")
					fmt.Println("  ")
					fmt.Println("  Option B: Use AWS CLI")
					fmt.Println("  Run: aws configure sso")
				} else {
					fmt.Printf("\n• %s: %s\n", failure.Name, failure.Suggestion)
				}
			}
		}

		// Optional requirements
		if len(optionalFailures) > 0 {
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("📦 OPTIONAL (for development only):")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			for _, failure := range optionalFailures {
				fmt.Printf("\n• %s: %s\n", failure.Name, failure.Suggestion)
			}
		}

		// Automatic fix option
		if fix {
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			logger.Info("Attempting automatic fixes for supported components...")
			if err := checker.FixIssues(results); err != nil {
				logger.Error("Some automatic fixes failed", "error", err)
				fmt.Println("\n⚠️  Please complete the manual steps above")
			} else {
				fmt.Println("\n✅ Automatic fixes applied. Run 'ztictl config check' again to verify")
			}
			return nil
		} else if len(criticalFailures) > 0 {
			// Check if any can be auto-fixed
			canAutoFix := false
			for _, failure := range criticalFailures {
				if failure.Name == "Session Manager Plugin" || failure.Name == "jq" {
					canAutoFix = true
					break
				}
			}

			if canAutoFix {
				fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
				fmt.Println("💡 TIP: Some issues can be fixed automatically")
				fmt.Println("   Run: ztictl config check --fix")
				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			}
		}

		// Final verification step
		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("✔️  VERIFY: After completing the steps above")
		fmt.Println("   Run: ztictl config check")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		return fmt.Errorf("prerequisites not met - see action plan above")
	}

	// If we only have warnings (no critical failures), still show them but don't fail
	if len(warningFailures) > 0 {
		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("⚠️  WARNINGS (recommended but not required):")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for _, failure := range warningFailures {
			if failure.Name == "AWS Credentials" {
				fmt.Println("\n• AWS Credentials: Your AWS credentials are missing or expired")
				fmt.Println("  ")
				fmt.Println("  Option A: Use ztictl (recommended)")
				fmt.Println("  Run: ztictl auth login")
				fmt.Println("  ")
				fmt.Println("  Option B: Use AWS CLI")
				fmt.Println("  Run: aws configure sso")
			} else {
				fmt.Printf("\n• %s: %s\n", failure.Name, failure.Suggestion)
			}
		}
		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("✅ All required checks passed! You can use ztictl")
		fmt.Println("   Note: Address the warnings above for full functionality")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}

	// All checks passed
	return nil
}

// validateConfiguration handles the config validation logic and returns errors instead of calling os.Exit
func validateConfiguration() error {
	logger.Info("Validating configuration...")

	// Check if configuration file exists
	if !config.Exists() {
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".ztictl.yaml")
		return fmt.Errorf("configuration file not found at %s - run 'ztictl config init' to create it", configPath)
	}

	// Re-load configuration to check for errors
	if err := config.Load(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	cfg := config.Get()

	// Report configuration file location
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".ztictl.yaml")
	logger.Info("Configuration source", "file", configPath)

	// Check if SSO configuration is properly set
	if cfg.SSO.StartURL == "" {
		return fmt.Errorf("SSO start URL is not configured - run 'ztictl config init' to set it up")
	}

	// Perform additional validation
	var errors []string

	if cfg.SSO.StartURL != "" {
		if cfg.SSO.Region == "" {
			errors = append(errors, "SSO region is required when SSO start URL is provided")
		}
		// Check if it's a placeholder URL
		if aws.IsPlaceholderSSOURL(cfg.SSO.StartURL) {
			errors = append(errors, fmt.Sprintf("SSO start URL is a placeholder value: %s", cfg.SSO.StartURL))
		}
	}

	// Report loaded configuration
	logger.Info("SSO configuration", "url", cfg.SSO.StartURL, "region", cfg.SSO.Region)

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
			fmt.Printf("\nEnter your AWS SSO domain ID (e.g., d-1234567890 or company-name): ")
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
			// Validate the constructed URL using the new validation function
			if err := aws.ValidateSSOURL(newValue); err != nil {
				fmt.Printf("Invalid SSO URL: %s\n", err)
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
