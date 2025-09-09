package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"ztictl/pkg/errors"
	"ztictl/pkg/security"
)

// Config represents the application configuration
type Config struct {
	// AWS SSO Configuration
	SSO SSOConfig `mapstructure:"sso"`

	// Default AWS region
	DefaultRegion string `mapstructure:"default_region"`

	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging"`

	// System configuration
	System SystemConfig `mapstructure:"system"`

	// Region configuration for multi-region operations
	Regions RegionConfig `mapstructure:"regions"`
}

// SSOConfig represents SSO-specific configuration
type SSOConfig struct {
	// SSO start URL
	StartURL string `mapstructure:"start_url"`

	// SSO region
	Region string `mapstructure:"region"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	// Log directory path
	Directory string `mapstructure:"directory"`

	// Enable file logging
	FileLogging bool `mapstructure:"file_logging"`

	// Log level (debug, info, warn, error)
	Level string `mapstructure:"level"`
}

// SystemConfig represents system-specific configuration
type SystemConfig struct {
	// IAM propagation delay in seconds
	IAMPropagationDelay int `mapstructure:"iam_propagation_delay"`

	// File size threshold for S3 transfer (in bytes)
	FileSizeThreshold int64 `mapstructure:"file_size_threshold"`

	// S3 bucket prefix for file transfers
	S3BucketPrefix string `mapstructure:"s3_bucket_prefix"`

	// Temporary directory for file operations
	TempDirectory string `mapstructure:"temp_directory"`
}

// RegionConfig represents region configuration for multi-region operations
type RegionConfig struct {
	// Groups of regions (e.g., production, development, all)
	Groups map[string][]string `mapstructure:"groups"`

	// Enabled regions for the account
	Enabled []string `mapstructure:"enabled"`
}

var (
	// Global configuration instance
	cfg *Config
)

// Load loads the configuration from file and environment variables
func Load() error {
	cfg = &Config{}

	// Set defaults first
	setDefaults()

	// Check if this is a first run (no config file exists)
	isFirstRun := !Exists()

	// Debug: check what viper has loaded
	if viper.GetBool("debug") {
		fmt.Printf("[DEBUG] Config file exists: %v\n", Exists())
		fmt.Printf("[DEBUG] Config file path: %s\n", getConfigPath())
		fmt.Printf("[DEBUG] Viper config file used: %s\n", viper.ConfigFileUsed())
		fmt.Printf("[DEBUG] Viper has SSO start_url: %q\n", viper.GetString("sso.start_url"))
	}

	// Try to load from legacy .env file first (from the parent directory where authaws is)
	envFilePath := filepath.Join("..", ".env")
	envFileExists := false
	if _, err := os.Stat(envFilePath); err == nil {
		envFileExists = true
		if err := LoadLegacyEnvFile(envFilePath); err != nil {
			return err
		}
	}

	// If config file exists but viper didn't load it, try to read it manually
	if !isFirstRun && viper.ConfigFileUsed() == "" {
		configPath := getConfigPath()
		if viper.GetBool("debug") {
			fmt.Printf("[DEBUG] Viper didn't load config, trying manual read from: %s\n", configPath)
		}

		// Set the config file explicitly
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			if viper.GetBool("debug") {
				fmt.Printf("[DEBUG] Manual config read failed: %v\n", err)
			}
			// Continue anyway, will use defaults
		} else {
			if viper.GetBool("debug") {
				fmt.Printf("[DEBUG] Manual config read successful\n")
			}
		}
	}

	// If this is first run and no .env file exists, we need user configuration
	if isFirstRun && !envFileExists {
		// For first run without existing .env, create a minimal valid config with defaults
		// The user can run 'ztictl config init' later to configure properly
		cfg = &Config{
			SSO: SSOConfig{
				StartURL: "", // Will be empty, user needs to configure
				Region:   viper.GetString("sso.region"),
			},
			DefaultRegion: viper.GetString("default_region"),
			Logging: LoggingConfig{
				Directory:   expandPath(viper.GetString("logging.directory")),
				FileLogging: viper.GetBool("logging.file_logging"),
				Level:       viper.GetString("logging.level"),
			},
			System: SystemConfig{
				IAMPropagationDelay: viper.GetInt("system.iam_propagation_delay"),
				FileSizeThreshold:   viper.GetInt64("system.file_size_threshold"),
				S3BucketPrefix:      viper.GetString("system.s3_bucket_prefix"),
				TempDirectory:       viper.GetString("system.temp_directory"),
			},
		}
	} else {
		// Try to load from config file (normal operation)
		if err := viper.Unmarshal(cfg); err != nil {
			return errors.NewConfigError("failed to unmarshal configuration", err)
		}

		// Log raw values for debugging
		if viper.GetBool("debug") {
			fmt.Printf("[DEBUG] Raw SSO Start URL from viper: %q\n", viper.GetString("sso.start_url"))
			fmt.Printf("[DEBUG] Raw SSO Region from viper: %q\n", viper.GetString("sso.region"))
		}

		// Expand paths with tilde support
		cfg.Logging.Directory = expandPath(cfg.Logging.Directory)

		// Validate loaded values
		if err := validateLoadedConfig(cfg); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}
	}

	// Validate configuration (but allow empty SSO config for first run)
	if err := validate(cfg); err != nil {
		// If validation fails and it's first run, provide helpful guidance
		if isFirstRun && !envFileExists {
			return errors.NewValidationError("Configuration needed. Please run 'ztictl config init' to set up your AWS SSO settings")
		}
		return err
	}

	return nil
}

// Get returns the global configuration instance
func Get() *Config {
	if cfg == nil {
		cfg = &Config{}
		setDefaults()
	}
	return cfg
}

// setDefaults sets default configuration values
func setDefaults() {
	// AWS defaults
	viper.SetDefault("default_region", "ca-central-1")

	// SSO defaults - these should be overridden by user config or .env file
	viper.SetDefault("sso.region", "us-east-1")

	// Logging defaults
	home, _ := os.UserHomeDir()
	viper.SetDefault("logging.directory", filepath.Join(home, "logs"))
	viper.SetDefault("logging.file_logging", true)
	viper.SetDefault("logging.level", "info")

	// System defaults
	viper.SetDefault("system.iam_propagation_delay", 5)
	viper.SetDefault("system.file_size_threshold", 1048576) // 1MB
	viper.SetDefault("system.s3_bucket_prefix", "ztictl-ssm-file-transfer")
	viper.SetDefault("system.temp_directory", os.TempDir()) // Platform-appropriate temp directory
}

// validate validates the configuration
func validate(cfg *Config) error {
	// If SSO start URL is empty, this might be a first run - allow it but warn
	if cfg.SSO.StartURL == "" {
		// This is okay for first run, but commands that need SSO will fail gracefully
		return nil
	}

	// Validate SSO configuration if provided
	if cfg.SSO.StartURL != "" {
		if cfg.SSO.Region == "" {
			return errors.NewValidationError("SSO region must be specified when SSO start URL is provided")
		}
	}

	return nil
}

// LoadLegacyEnvFile loads configuration from the legacy .env file if present
func LoadLegacyEnvFile(envFilePath string) error {
	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		return nil // No .env file, not an error
	}

	// Validate file path to prevent directory traversal (G304 fix)
	// Allow absolute paths and specific relative paths needed by the application
	cleanPath, err := filepath.Abs(filepath.Clean(envFilePath))
	if err != nil {
		return errors.NewConfigError("failed to resolve .env file path", err)
	}

	// Only validate relative paths that could be dangerous (but allow ../.env which is used by design)
	if !filepath.IsAbs(envFilePath) && envFilePath != "../.env" && !filepath.IsAbs(cleanPath) {
		if err := security.ValidateFilePathWithWorkingDir(envFilePath); err != nil {
			return errors.NewConfigError("invalid .env file path", err)
		}
	}
	envFilePath = cleanPath

	// Read .env file manually since viper doesn't handle bash-style env files well
	envFile, err := os.Open(envFilePath)
	if err != nil {
		return errors.NewConfigError("failed to open .env file", err)
	}
	defer envFile.Close()

	// Parse .env file and set viper values
	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes if present
			if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
				(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
				value = value[1 : len(value)-1]
			}

			// Map legacy .env variables to new config structure
			switch key {
			case "SSO_START_URL":
				viper.Set("sso.start_url", value)
			case "SSO_REGION":
				viper.Set("sso.region", value)
			case "LOG_DIR":
				viper.Set("logging.directory", value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return errors.NewConfigError("failed to parse .env file", err)
	}

	return nil
}

// CreateSampleConfig creates a sample configuration file
func CreateSampleConfig(configPath string) error {
	// Get home directory for platform-compatible paths
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to get home directory: %w", err)
	}

	// Use absolute path for log directory with forward slashes for YAML
	logDir := filepath.Join(home, "logs")
	logDir = filepath.ToSlash(logDir) // Convert to forward slashes for YAML compatibility

	// Use platform-appropriate temp directory with forward slashes for YAML
	tempDir := os.TempDir()
	tempDir = filepath.ToSlash(tempDir) // Convert to forward slashes for YAML compatibility

	sampleConfig := fmt.Sprintf(`# ztictl Configuration File
# This file configures ztictl with your AWS SSO and system settings

# AWS SSO Configuration
sso:
  # Your AWS SSO portal URL (required for SSO authentication)
  start_url: "https://d-xxxxxxxxxx.awsapps.com/start"
  
  # The AWS region where your SSO is configured
  region: "us-east-1"

# Default AWS region for operations
default_region: "ca-central-1"

# Logging configuration
logging:
  # Directory for log files (absolute path, Windows compatible)
  directory: "%s"
  
  # Enable file logging (in addition to console)
  file_logging: true
  
  # Log level: debug, info, warn, error
  level: "info"

# System configuration
system:
  # IAM propagation delay in seconds (how long to wait for IAM changes)
  iam_propagation_delay: 5
  
  # File size threshold for S3 transfer (bytes) - files larger than this use S3
  file_size_threshold: 1048576  # 1MB
  
  # S3 bucket prefix for file transfers
  s3_bucket_prefix: "ztictl-ssm-file-transfer"
  
  # Temporary directory for file operations (platform-appropriate)
  temp_directory: "%s"
`, logDir, tempDir)

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return errors.NewConfigError("failed to create config directory", err)
	}

	// Write sample config
	if err := os.WriteFile(configPath, []byte(sampleConfig), 0600); err != nil {
		return errors.NewConfigError("failed to write sample config", err)
	}

	return nil
}

// getConfigPath returns the default configuration file path
func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ztictl.yaml")
}

// Exists checks if a configuration file exists
func Exists() bool {
	configPath := getConfigPath()
	_, err := os.Stat(configPath)
	return err == nil
}

// InteractiveInit performs an interactive configuration setup
func InteractiveInit() error {
	fmt.Println("\n🚀 Welcome to ztictl! Let's set up your configuration.")
	fmt.Println("========================================================")

	reader := bufio.NewReader(os.Stdin)
	config := &Config{}

	// AWS SSO Configuration
	fmt.Println("\n📋 AWS SSO Configuration")
	fmt.Println("-------------------------")

	// Get SSO Start URL with validation
	for {
		fmt.Print("Enter your AWS SSO Start URL (e.g., https://d-xxxxxxxxxx.awsapps.com/start): ")
		startURL, _ := reader.ReadString('\n')
		startURL = strings.TrimSpace(startURL)

		if err := validateInput(startURL, "url"); err != nil {
			fmt.Printf("❌ %s\n", err)
			continue
		}
		config.SSO.StartURL = startURL
		break
	}

	// Get SSO Region with validation
	for {
		fmt.Print("Enter your SSO region [us-east-1]: ")
		ssoRegion, _ := reader.ReadString('\n')
		ssoRegion = strings.TrimSpace(ssoRegion)
		if ssoRegion == "" {
			ssoRegion = "us-east-1"
		}

		if err := validateInput(ssoRegion, "region"); err != nil {
			fmt.Printf("❌ %s\n", err)
			continue
		}
		config.SSO.Region = ssoRegion
		break
	}

	// Default AWS Region
	fmt.Println("\n🌍 Default AWS Region")
	fmt.Println("--------------------")
	// Get Default Region with validation
	for {
		fmt.Print("Enter your default AWS region [ca-central-1]: ")
		defaultRegion, _ := reader.ReadString('\n')
		defaultRegion = strings.TrimSpace(defaultRegion)
		if defaultRegion == "" {
			defaultRegion = "ca-central-1"
		}

		if err := validateInput(defaultRegion, "region"); err != nil {
			fmt.Printf("❌ %s\n", err)
			continue
		}
		config.DefaultRegion = defaultRegion
		break
	}

	// Logging Configuration
	fmt.Println("\n📝 Logging Configuration")
	fmt.Println("------------------------")
	fmt.Print("Enable file logging? [y/N]: ")
	fileLogging, _ := reader.ReadString('\n')
	config.Logging.FileLogging = strings.ToLower(strings.TrimSpace(fileLogging)) == "y"

	fmt.Print("Log directory [~/logs]: ")
	logDir, _ := reader.ReadString('\n')
	logDir = strings.TrimSpace(logDir)
	if logDir == "" {
		logDir = "~/logs"
	}
	// Expand and convert to forward slashes for YAML compatibility
	logDir = expandPath(logDir)
	logDir = filepath.ToSlash(logDir)
	config.Logging.Directory = logDir

	fmt.Print("Log level (debug/info/warn/error) [info]: ")
	logLevel, _ := reader.ReadString('\n')
	logLevel = strings.TrimSpace(logLevel)
	if logLevel == "" {
		logLevel = "info"
	}
	config.Logging.Level = logLevel

	// System Configuration
	fmt.Println("\n⚙️  System Configuration")
	fmt.Println("-----------------------")
	fmt.Print("IAM propagation delay in seconds [5]: ")
	iamDelay, _ := reader.ReadString('\n')
	iamDelay = strings.TrimSpace(iamDelay)
	if iamDelay == "" {
		config.System.IAMPropagationDelay = 5
	} else {
		_, _ = fmt.Sscanf(iamDelay, "%d", &config.System.IAMPropagationDelay) // Ignore error, will use default on parse failure
	}

	fmt.Print("S3 bucket prefix for file transfers [ztictl-ssm-file-transfer]: ")
	s3Prefix, _ := reader.ReadString('\n')
	s3Prefix = strings.TrimSpace(s3Prefix)
	if s3Prefix == "" {
		s3Prefix = "ztictl-ssm-file-transfer"
	}
	config.System.S3BucketPrefix = s3Prefix

	// Set default file size threshold
	config.System.FileSizeThreshold = 1048576 // 1MB

	// Set temp directory with forward slashes for YAML
	config.System.TempDirectory = filepath.ToSlash(os.TempDir())

	// Write the configuration
	if err := writeInteractiveConfig(config); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	fmt.Println("\n✅ Configuration saved successfully!")
	fmt.Printf("📁 Config file: %s\n", getConfigPath())
	fmt.Println("\nNext steps:")
	fmt.Println("1. Run 'ztictl config check' to verify requirements")
	fmt.Println("2. Run 'ztictl auth login' to authenticate")

	return nil
}

// writeInteractiveConfig writes the interactively created config
func writeInteractiveConfig(config *Config) error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Generate YAML content
	yamlContent := fmt.Sprintf(`# ztictl Configuration File
# Generated interactively on first run

# AWS SSO Configuration
sso:
  start_url: "%s"
  region: "%s"

# Default AWS region for operations
default_region: "%s"

# Logging configuration
logging:
  directory: "%s"
  file_logging: %t
  level: "%s"

# System configuration
system:
  iam_propagation_delay: %d
  file_size_threshold: %d
  s3_bucket_prefix: "%s"
`,
		config.SSO.StartURL,
		config.SSO.Region,
		config.DefaultRegion,
		config.Logging.Directory,
		config.Logging.FileLogging,
		config.Logging.Level,
		config.System.IAMPropagationDelay,
		config.System.FileSizeThreshold,
		config.System.S3BucketPrefix,
	)

	// Write to file
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// expandPath expands paths with tilde (~) to the user's home directory
func expandPath(path string) string {
	if path == "" {
		return path
	}

	// Handle tilde expansion
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path // Return original path if we can't get home dir
		}
		return filepath.Join(home, path[2:])
	}

	// Handle bare tilde
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}

	return path
}

// validateLoadedConfig validates configuration values after loading
func validateLoadedConfig(cfg *Config) error {
	// Validate SSO Start URL if provided
	if cfg.SSO.StartURL != "" {
		// Check for common mistakes first
		if strings.Contains(cfg.SSO.StartURL, `""`) || strings.HasPrefix(cfg.SSO.StartURL, `"`) {
			return errors.NewValidationError("SSO start URL contains invalid quotes")
		}
		if !isValidURL(cfg.SSO.StartURL) {
			return errors.NewValidationError(fmt.Sprintf("SSO start URL is not a valid URL: %s", cfg.SSO.StartURL))
		}
	}

	// Validate regions
	if cfg.SSO.Region != "" && !isValidRegion(cfg.SSO.Region) {
		return errors.NewValidationError(fmt.Sprintf("SSO region is not valid: %s", cfg.SSO.Region))
	}
	if cfg.DefaultRegion != "" && !isValidRegion(cfg.DefaultRegion) {
		return errors.NewValidationError(fmt.Sprintf("Default region is not valid: %s", cfg.DefaultRegion))
	}

	// Paths are already converted to forward slashes for YAML storage, no validation needed

	return nil
}

// isValidURL checks if a string is a valid URL
func isValidURL(s string) bool {
	if s == "" {
		return false
	}
	// Basic URL validation - must start with http:// or https:// and have something after
	if strings.HasPrefix(s, "http://") {
		return len(s) > 7 // More than just "http://"
	}
	if strings.HasPrefix(s, "https://") {
		return len(s) > 8 // More than just "https://"
	}
	return false
}

// isValidRegion checks if a region string follows AWS region format
func isValidRegion(region string) bool {
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

// validateInput validates user input during interactive configuration
func validateInput(input string, inputType string) error {
	input = strings.TrimSpace(input)

	switch inputType {
	case "url":
		if input != "" && !isValidURL(input) {
			return fmt.Errorf("invalid URL format. Must start with http:// or https://")
		}
	case "region":
		if input != "" && !isValidRegion(input) {
			return fmt.Errorf("invalid AWS region: %s", input)
		}
	case "path":
		// No validation needed - paths will be automatically converted to forward slashes for YAML
	}

	return nil
}
