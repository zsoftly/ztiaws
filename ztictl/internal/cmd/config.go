package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"ztictl/internal/config"
	"ztictl/internal/system"
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
	Long: `Initialize ztictl configuration by creating a sample configuration file.
This will create a .ztictl.yaml file in your home directory with default settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		interactive, _ := cmd.Flags().GetBool("interactive")
		
		if interactive {
			// Check if config file already exists for interactive mode
			if config.Exists() && !force {
				logger.Error("Configuration file already exists")
				logger.Info("Use --force to overwrite existing configuration")
				os.Exit(1)
			}
			
			// Use interactive configuration
			if err := config.InteractiveInit(); err != nil {
				logger.Error("Failed to create interactive configuration", "error", err)
				os.Exit(1)
			}
		} else {
			// Check if config file already exists for non-interactive mode
			if config.Exists() && !force {
				logger.Error("Configuration file already exists")
				logger.Info("Use --force to overwrite existing configuration or --interactive for guided setup")
				os.Exit(1)
			}
			
			// Determine config file path
			home, err := os.UserHomeDir()
			if err != nil {
				logger.Error("Unable to find home directory", "error", err)
				os.Exit(1)
			}

			configPath := filepath.Join(home, ".ztictl.yaml")

			// Create sample configuration
			if err := config.CreateSampleConfig(configPath); err != nil {
				logger.Error("Failed to create configuration file", "error", err)
				os.Exit(1)
			}

			logger.Info("Configuration file created successfully", "path", configPath)
			logger.Info("Please edit the configuration file with your AWS SSO settings")
			
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("1. Edit %s with your AWS SSO settings\n", configPath)
			fmt.Printf("2. Run 'ztictl config check' to verify requirements\n")
			fmt.Printf("3. Run 'ztictl auth login' to authenticate\n")
			fmt.Printf("\nTip: Use 'ztictl config init --interactive' for guided setup\n")
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
		
		logger.Info("Checking system requirements...")

		checker := system.NewRequirementsChecker(logger)
		
		// Check all requirements
		results, err := checker.CheckAll()
		if err != nil {
			logger.Error("Failed to check requirements", "error", err)
			os.Exit(1)
		}

		// Display results
		allPassed := true
		fmt.Println("\nSystem Requirements Check:")
		fmt.Println("==========================")

		for _, result := range results {
			status := "❌ FAIL"
			if result.Passed {
				status = "✅ PASS"
			} else {
				allPassed = false
			}

			fmt.Printf("%-30s %s\n", result.Name, status)
			
			if !result.Passed {
				fmt.Printf("  Issue: %s\n", result.Error)
				if result.Suggestion != "" {
					fmt.Printf("  Fix: %s\n", result.Suggestion)
				}
				fmt.Println()
			}
		}

		if allPassed {
			logger.Info("All requirements met! ✅")
		} else {
			logger.Error("Some requirements are not met")
			
			if fix {
				logger.Info("Attempting to fix issues...")
				if err := checker.FixIssues(results); err != nil {
					logger.Error("Failed to fix some issues", "error", err)
					os.Exit(1)
				}
				logger.Info("Issues fixed successfully. Please run check again to verify.")
			} else {
				logger.Info("Run with --fix to attempt automatic fixes")
				os.Exit(1)
			}
		}
	},
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current ztictl configuration including all settings and their sources.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		fmt.Println("\nztictl Configuration:")
		fmt.Println("=====================")
		
		fmt.Printf("Default Region: %s\n", cfg.DefaultRegion)
		fmt.Println()
		
		fmt.Println("AWS SSO Configuration:")
		fmt.Printf("  Start URL: %s\n", cfg.SSO.StartURL)
		fmt.Printf("  Region: %s\n", cfg.SSO.Region) 
		fmt.Printf("  Default Profile: %s\n", cfg.SSO.DefaultProfile)
		fmt.Println()
		
		fmt.Println("Logging Configuration:")
		fmt.Printf("  Directory: %s\n", cfg.Logging.Directory)
		fmt.Printf("  File Logging: %t\n", cfg.Logging.FileLogging)
		fmt.Printf("  Level: %s\n", cfg.Logging.Level)
		fmt.Println()
		
		fmt.Println("System Configuration:")
		fmt.Printf("  IAM Propagation Delay: %d seconds\n", cfg.System.IAMPropagationDelay)
		fmt.Printf("  File Size Threshold: %d bytes (%.1f MB)\n", 
			cfg.System.FileSizeThreshold, float64(cfg.System.FileSizeThreshold)/1024/1024)
		fmt.Printf("  S3 Bucket Prefix: %s\n", cfg.System.S3BucketPrefix)

		// Show config file location if available
		if configFile := os.Getenv("ZTICTL_CONFIG"); configFile != "" {
			fmt.Printf("\nConfig File: %s\n", configFile)
		} else {
			home, _ := os.UserHomeDir()
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
		logger.Info("Validating configuration...")

		// Re-load configuration to check for errors
		if err := config.Load(); err != nil {
			logger.Error("Configuration validation failed", "error", err)
			os.Exit(1)
		}

		cfg := config.Get()

		// Perform additional validation
		var errors []string

		if cfg.SSO.StartURL != "" {
			if cfg.SSO.Region == "" {
				errors = append(errors, "SSO region is required when SSO start URL is provided")
			}
			if cfg.SSO.DefaultProfile == "" {
				errors = append(errors, "SSO default profile is required when SSO start URL is provided")
			}
		}

		if len(errors) > 0 {
			logger.Error("Configuration validation failed:")
			for _, err := range errors {
				fmt.Printf("  - %s\n", err)
			}
			os.Exit(1)
		}

		logger.Info("Configuration validation passed ✅")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configCheckCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)

	// Add flags
	configInitCmd.Flags().BoolP("force", "f", false, "Overwrite existing configuration file")
	configInitCmd.Flags().BoolP("interactive", "i", false, "Interactive configuration setup")
	configCheckCmd.Flags().BoolP("fix", "", false, "Attempt to automatically fix issues")
}
