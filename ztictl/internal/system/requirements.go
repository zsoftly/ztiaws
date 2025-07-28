package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"ztictl/internal/logging"
)

// RequirementsChecker checks system requirements and dependencies
type RequirementsChecker struct {
	logger *logging.Logger
}

// RequirementResult represents the result of a requirement check
type RequirementResult struct {
	Name       string `json:"name"`
	Passed     bool   `json:"passed"`
	Error      string `json:"error,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	Version    string `json:"version,omitempty"`
}

// NewRequirementsChecker creates a new requirements checker
func NewRequirementsChecker(logger *logging.Logger) *RequirementsChecker {
	return &RequirementsChecker{
		logger: logger,
	}
}

// CheckAll checks all system requirements
func (c *RequirementsChecker) CheckAll() ([]RequirementResult, error) {
	results := []RequirementResult{}

	// Check AWS CLI
	awsResult := c.checkAWSCLI()
	results = append(results, awsResult)

	// Check Session Manager Plugin
	ssmResult := c.checkSSMPlugin()
	results = append(results, ssmResult)

	// Check jq
	jqResult := c.checkJQ()
	results = append(results, jqResult)

	// Check AWS credentials
	credsResult := c.checkAWSCredentials()
	results = append(results, credsResult)

	// Check Go version (for development)
	goResult := c.checkGoVersion()
	results = append(results, goResult)

	return results, nil
}

// FixIssues attempts to automatically fix failed requirements
func (c *RequirementsChecker) FixIssues(results []RequirementResult) error {
	for _, result := range results {
		if !result.Passed {
			switch result.Name {
			case "Session Manager Plugin":
				if err := c.installSSMPlugin(); err != nil {
					return fmt.Errorf("failed to install SSM plugin: %w", err)
				}
			case "jq":
				if err := c.installJQ(); err != nil {
					return fmt.Errorf("failed to install jq: %w", err)
				}
			default:
				c.logger.Warn("Cannot automatically fix requirement", "name", result.Name)
			}
		}
	}
	return nil
}

// checkAWSCLI checks if AWS CLI is installed and accessible
func (c *RequirementsChecker) checkAWSCLI() RequirementResult {
	result := RequirementResult{Name: "AWS CLI"}

	// Use platform-appropriate command name
	cmdName := "aws"
	if runtime.GOOS == "windows" {
		cmdName = "aws.exe"
	}

	// Check if aws command exists
	cmd := exec.Command(cmdName, "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Error = "AWS CLI not found"
		result.Suggestion = "Install AWS CLI from https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
		return result
	}

	// Extract version
	versionLine := strings.TrimSpace(string(output))
	if strings.Contains(versionLine, "aws-cli/") {
		parts := strings.Split(versionLine, " ")
		if len(parts) > 0 {
			result.Version = strings.TrimPrefix(parts[0], "aws-cli/")
		}
	}

	result.Passed = true
	return result
}

// checkSSMPlugin checks if Session Manager plugin is installed
func (c *RequirementsChecker) checkSSMPlugin() RequirementResult {
	result := RequirementResult{Name: "Session Manager Plugin"}

	// Use platform-appropriate command name
	cmdName := "session-manager-plugin"
	if runtime.GOOS == "windows" {
		cmdName = "session-manager-plugin.exe"
	}

	// Check if session-manager-plugin exists
	cmd := exec.Command(cmdName)
	err := cmd.Run()

	// session-manager-plugin returns exit code 255 when called without arguments, but that means it's installed
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 255 {
			result.Passed = true
			return result
		}

		result.Error = "Session Manager plugin not found"
		result.Suggestion = c.getSSMPluginInstallInstructions()
		return result
	}

	result.Passed = true
	return result
}

// checkJQ checks if jq is installed
func (c *RequirementsChecker) checkJQ() RequirementResult {
	result := RequirementResult{Name: "jq"}

	cmd := exec.Command("jq", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Error = "jq not found"
		result.Suggestion = c.getJQInstallInstructions()
		return result
	}

	result.Version = strings.TrimSpace(string(output))
	result.Passed = true
	return result
}

// checkAWSCredentials checks if AWS credentials are configured and valid
func (c *RequirementsChecker) checkAWSCredentials() RequirementResult {
	result := RequirementResult{Name: "AWS Credentials"}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		result.Error = "Failed to load AWS configuration"
		result.Suggestion = "Configure AWS credentials using 'aws configure' or 'ztictl auth login'"
		return result
	}

	// Test credentials by calling STS GetCallerIdentity
	stsClient := sts.NewFromConfig(cfg)
	_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		result.Error = "AWS credentials are not valid or have expired"
		result.Suggestion = "Authenticate using 'ztictl auth login' or update your AWS credentials"
		return result
	}

	result.Passed = true
	return result
}

// checkGoVersion checks Go version (for development environments)
func (c *RequirementsChecker) checkGoVersion() RequirementResult {
	result := RequirementResult{Name: "Go Version"}

	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		result.Error = "Go not found"
		result.Suggestion = "Install Go from https://golang.org/dl/ (required for development only)"
		return result
	}

	versionLine := strings.TrimSpace(string(output))
	if strings.Contains(versionLine, "go version") {
		parts := strings.Split(versionLine, " ")
		if len(parts) >= 3 {
			result.Version = parts[2]
		}
	}

	result.Passed = true
	return result
}

// getSSMPluginInstallInstructions returns OS-specific installation instructions for SSM plugin
func (c *RequirementsChecker) getSSMPluginInstallInstructions() string {
	switch runtime.GOOS {
	case "linux":
		// Detect distribution
		if c.isUbuntuDebian() {
			return "Install with: curl \"https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb\" -o \"session-manager-plugin.deb\" && sudo dpkg -i session-manager-plugin.deb"
		}
		return "Install from: https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html"
	case "darwin":
		return "Install with: brew install --cask session-manager-plugin"
	case "windows":
		return "Download from: https://s3.amazonaws.com/session-manager-downloads/plugin/latest/windows/SessionManagerPluginSetup.exe"
	default:
		return "Visit: https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html"
	}
}

// getJQInstallInstructions returns OS-specific installation instructions for jq
func (c *RequirementsChecker) getJQInstallInstructions() string {
	switch runtime.GOOS {
	case "linux":
		if c.isUbuntuDebian() {
			return "Install with: sudo apt-get install jq"
		} else if c.isRedHatCentOS() {
			return "Install with: sudo yum install jq"
		}
		return "Install jq using your distribution's package manager"
	case "darwin":
		return "Install with: brew install jq"
	case "windows":
		return "Install with: choco install jq or download from https://stedolan.github.io/jq/download/"
	default:
		return "Visit: https://stedolan.github.io/jq/download/"
	}
}

// installSSMPlugin attempts to automatically install the SSM plugin
func (c *RequirementsChecker) installSSMPlugin() error {
	c.logger.Info("Attempting to install Session Manager plugin...")

	switch runtime.GOOS {
	case "linux":
		if c.isUbuntuDebian() {
			return c.installSSMPluginUbuntu()
		}
		return fmt.Errorf("automatic installation not supported for this Linux distribution")
	case "darwin":
		return c.installSSMPluginMacOS()
	default:
		return fmt.Errorf("automatic installation not supported for %s", runtime.GOOS)
	}
}

// installJQ attempts to automatically install jq
func (c *RequirementsChecker) installJQ() error {
	c.logger.Info("Attempting to install jq...")

	switch runtime.GOOS {
	case "linux":
		if c.isUbuntuDebian() {
			return c.installJQUbuntu()
		} else if c.isRedHatCentOS() {
			return c.installJQRedHat()
		}
		return fmt.Errorf("automatic installation not supported for this Linux distribution")
	case "darwin":
		return c.installJQMacOS()
	default:
		return fmt.Errorf("automatic installation not supported for %s", runtime.GOOS)
	}
}

// OS detection helpers

func (c *RequirementsChecker) isUbuntuDebian() bool {
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return true
	}
	if _, err := os.Stat("/etc/lsb-release"); err == nil {
		content, err := os.ReadFile("/etc/lsb-release")
		if err == nil && strings.Contains(string(content), "Ubuntu") {
			return true
		}
	}
	return false
}

func (c *RequirementsChecker) isRedHatCentOS() bool {
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return true
	}
	if _, err := os.Stat("/etc/centos-release"); err == nil {
		return true
	}
	return false
}

// Installation helpers

func (c *RequirementsChecker) installSSMPluginUbuntu() error {
	// Download the package
	downloadCmd := exec.Command("curl",
		"https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb",
		"-o", "session-manager-plugin.deb")

	if err := downloadCmd.Run(); err != nil {
		return fmt.Errorf("failed to download SSM plugin: %w", err)
	}

	// Install the package
	installCmd := exec.Command("sudo", "dpkg", "-i", "session-manager-plugin.deb")
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install SSM plugin: %w", err)
	}

	// Clean up
	os.Remove("session-manager-plugin.deb")

	return nil
}

func (c *RequirementsChecker) installSSMPluginMacOS() error {
	// Check if Homebrew is available
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew not found. Please install manually or install Homebrew first")
	}

	cmd := exec.Command("brew", "install", "--cask", "session-manager-plugin")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install SSM plugin via Homebrew: %w", err)
	}

	return nil
}

func (c *RequirementsChecker) installJQUbuntu() error {
	cmd := exec.Command("sudo", "apt-get", "update")
	if err := cmd.Run(); err != nil {
		c.logger.Warn("Failed to update package list", "error", err)
	}

	cmd = exec.Command("sudo", "apt-get", "install", "-y", "jq")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install jq: %w", err)
	}

	return nil
}

func (c *RequirementsChecker) installJQRedHat() error {
	cmd := exec.Command("sudo", "yum", "install", "-y", "jq")
	if err := cmd.Run(); err != nil {
		// Try with dnf for newer systems
		cmd = exec.Command("sudo", "dnf", "install", "-y", "jq")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install jq: %w", err)
		}
	}

	return nil
}

func (c *RequirementsChecker) installJQMacOS() error {
	// Check if Homebrew is available
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew not found. Please install manually or install Homebrew first")
	}

	cmd := exec.Command("brew", "install", "jq")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install jq via Homebrew: %w", err)
	}

	return nil
}

// GetSystemInfo returns information about the current system
func (c *RequirementsChecker) GetSystemInfo() map[string]string {
	info := map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	}

	// Get OS version
	switch runtime.GOOS {
	case "linux":
		if content, err := os.ReadFile("/etc/os-release"); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					info["os_version"] = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
					break
				}
			}
		}
	case "darwin":
		if cmd := exec.Command("sw_vers", "-productVersion"); cmd != nil {
			if output, err := cmd.Output(); err == nil {
				info["os_version"] = strings.TrimSpace(string(output))
			}
		}
	}

	// Get Home directory
	if home, err := os.UserHomeDir(); err == nil {
		info["home"] = home
	}

	// Get AWS config directory
	info["aws_config_dir"] = filepath.Join(os.Getenv("HOME"), ".aws")

	return info
}
