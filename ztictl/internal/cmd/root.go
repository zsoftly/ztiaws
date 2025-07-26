package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"ztictl/internal/config"
	"ztictl/internal/logging"
	"ztictl/internal/splash"
)

const (
	// Version represents the current version of ztictl
	Version = "1.0.0"
)

var (
	configFile string
	debug      bool
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
		// Handle splash screen and first-run configuration
		handleStartup(cmd)
	},
}

// handleStartup manages splash screen display and first-run configuration
func handleStartup(cmd *cobra.Command) {
	// Skip splash for help commands and completion
	if cmd.Name() == "help" || cmd.Name() == "completion" {
		return
	}
	
	// Check if we should show splash screen
	showedSplash, err := splash.ShowSplash(Version)
	if err != nil {
		// Don't fail if splash fails, just log and continue
		if logger != nil {
			logger.Warn("Failed to show splash screen", "error", err)
		}
		return
	}
	
	// If this is first run, trigger interactive configuration
	if showedSplash {
		// Check if config file exists
		configExists := config.Exists()
		
		if !configExists {
			// Show interactive configuration
			if err := config.InteractiveInit(); err != nil {
				if logger != nil {
					logger.Error("Failed to initialize configuration", "error", err)
				}
				os.Exit(1)
			}
			
			// Reload configuration after interactive setup
			if err := config.Load(); err != nil {
				if logger != nil {
					logger.Error("Failed to load configuration after setup", "error", err)
				}
				os.Exit(1)
			}
		}
	}
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

	// Bind flags to viper
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Initialize logger early
	logger = logging.NewLogger(debug)

	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Error("Unable to find home directory", "error", err)
			os.Exit(1)
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
		logger.Debug("Using config file", "file", viper.ConfigFileUsed())
	}

	// Load configuration
	if err := config.Load(); err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Update logger with configuration
	if viper.GetBool("debug") {
		logger.SetLevel(logging.DebugLevel)
	}
}

// GetLogger returns the global logger instance
func GetLogger() *logging.Logger {
	return logger
}
