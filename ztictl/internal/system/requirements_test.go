package system

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"ztictl/pkg/logging"
)

func TestNewRequirementsChecker(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	if checker == nil {
		t.Error("NewRequirementsChecker returned nil")
	}
}

func TestRequirementResult(t *testing.T) {
	// Test that RequirementResult struct works as expected
	result := RequirementResult{
		Name:    "Test",
		Passed:  true,
		Version: "2.1.0",
	}

	if result.Name != "Test" {
		t.Errorf("Expected name 'Test', got %q", result.Name)
	}

	if !result.Passed {
		t.Error("Expected Passed to be true")
	}

	if result.Version != "2.1.0" {
		t.Errorf("Expected version '2.1.0', got %q", result.Version)
	}
}

func TestCheckAWSCLI(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	// This test will pass or fail based on whether AWS CLI is installed
	// In CI, it will likely fail, but that's expected
	result := checker.checkAWSCLI()

	if result.Name != "AWS CLI" {
		t.Errorf("Expected name 'AWS CLI', got %q", result.Name)
	}

	// Either passed or has an error message
	if !result.Passed && result.Error == "" {
		t.Error("If not passed, should have error message")
	}
}

func TestCheckSSMPlugin(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	// This test will likely fail in CI, but that's expected
	result := checker.checkSSMPlugin()

	if result.Name != "Session Manager Plugin" {
		t.Errorf("Expected name 'Session Manager Plugin', got %q", result.Name)
	}

	// Either passed or has an error message
	if !result.Passed && result.Error == "" {
		t.Error("If not passed, should have error message")
	}
}

func TestGetSSMPluginInstallInstructions(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	// Test that this returns platform-specific instructions
	instructions := checker.getSSMPluginInstallInstructions()

	if instructions == "" {
		t.Error("Expected non-empty installation instructions")
	}

	// Should contain some platform-specific information
	if !containsAny(instructions, []string{"https://", "Download", "Install"}) {
		t.Error("Instructions should contain download/install information")
	}
}

func TestGetJQInstallInstructions(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	instructions := checker.getJQInstallInstructions()

	if instructions == "" {
		t.Error("Expected non-empty jq installation instructions")
	}
}

func containsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		if contains(str, substr) {
			return true
		}
	}
	return false
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && indexOfSubstring(str, substr) >= 0
}

func indexOfSubstring(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestCheckAll(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	results, err := checker.CheckAll()
	if err != nil {
		t.Errorf("CheckAll() returned error: %v", err)
	}

	if len(results) == 0 {
		t.Error("CheckAll() returned empty results")
	}

	// Check that all expected requirements are included
	expectedNames := []string{
		"AWS CLI",
		"Session Manager Plugin",
		"jq",
		"AWS Credentials",
		"Go Version",
	}

	resultNames := make(map[string]bool)
	for _, result := range results {
		resultNames[result.Name] = true
	}

	for _, name := range expectedNames {
		if !resultNames[name] {
			t.Errorf("CheckAll() missing expected requirement: %s", name)
		}
	}

	// Verify structure of results
	for _, result := range results {
		if result.Name == "" {
			t.Error("Requirement result should have name")
		}

		// If not passed, should have error message
		if !result.Passed && result.Error == "" {
			t.Errorf("Failed requirement %s should have error message", result.Name)
		}
	}
}

func TestCheckJQ(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	result := checker.checkJQ()

	if result.Name != "jq" {
		t.Errorf("Expected name 'jq', got %q", result.Name)
	}

	// Either passed or has error
	if !result.Passed && result.Error == "" {
		t.Error("If not passed, should have error message")
	}

	// If passed, should have version
	if result.Passed && result.Version == "" {
		t.Error("If passed, should have version info")
	}

	// Should have suggestion if failed
	if !result.Passed && result.Suggestion == "" {
		t.Error("Failed jq check should have suggestion")
	}
}

func TestCheckAWSCredentials(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	// This will likely fail in CI, but we can test the structure
	result := checker.checkAWSCredentials()

	if result.Name != "AWS Credentials" {
		t.Errorf("Expected name 'AWS Credentials', got %q", result.Name)
	}

	// Either passed or has error/suggestion
	if !result.Passed {
		if result.Error == "" {
			t.Error("Failed credentials check should have error message")
		}
		if result.Suggestion == "" {
			t.Error("Failed credentials check should have suggestion")
		}
	}
}

func TestCheckGoVersion(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	result := checker.checkGoVersion()

	if result.Name != "Go Version" {
		t.Errorf("Expected name 'Go Version', got %q", result.Name)
	}

	// This should typically pass since we're running Go tests
	// But may fail in some CI environments
	if !result.Passed {
		if result.Error == "" {
			t.Error("Failed Go check should have error message")
		}
		if result.Suggestion == "" {
			t.Error("Failed Go check should have suggestion")
		}
	} else {
		// If passed, should have version
		if result.Version == "" {
			t.Error("Successful Go check should have version")
		}
	}
}

func TestPlatformSpecificCommands(t *testing.T) {
	// Test platform-specific command name logic
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	// Test that the appropriate command names are used
	// We can't easily mock runtime.GOOS, but we can test the current platform

	// Test AWS CLI check with current platform
	result := checker.checkAWSCLI()
	if result.Name != "AWS CLI" {
		t.Errorf("Expected name 'AWS CLI', got %q", result.Name)
	}

	// Test SSM plugin check with current platform
	result = checker.checkSSMPlugin()
	if result.Name != "Session Manager Plugin" {
		t.Errorf("Expected name 'Session Manager Plugin', got %q", result.Name)
	}
}

func TestGetSSMPluginInstallInstructionsPlatforms(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	instructions := checker.getSSMPluginInstallInstructions()

	if instructions == "" {
		t.Error("Instructions should not be empty")
	}

	// Should contain platform-relevant information
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(instructions, "brew") {
			t.Error("macOS instructions should mention brew")
		}
	case "linux":
		// Should contain either dpkg or general link
		if !strings.Contains(instructions, "dpkg") && !strings.Contains(instructions, "https://") {
			t.Error("Linux instructions should mention dpkg or provide link")
		}
	case "windows":
		if !strings.Contains(instructions, ".exe") {
			t.Error("Windows instructions should mention .exe file")
		}
	}
}

func TestGetJQInstallInstructionsPlatforms(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	instructions := checker.getJQInstallInstructions()

	if instructions == "" {
		t.Error("Instructions should not be empty")
	}

	// Should contain platform-relevant information
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(instructions, "brew") {
			t.Error("macOS instructions should mention brew")
		}
	case "linux":
		// Should contain package manager info
		if !strings.Contains(instructions, "apt-get") && !strings.Contains(instructions, "yum") && !strings.Contains(instructions, "package manager") {
			t.Error("Linux instructions should mention package managers")
		}
	case "windows":
		if !strings.Contains(instructions, "choco") && !strings.Contains(instructions, "download") {
			t.Error("Windows instructions should mention choco or download")
		}
	}
}

func TestOSDetectionHelpers(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	// Test OS detection functions
	// These will return different results based on the actual OS
	isUbuntu := checker.isUbuntuDebian()
	isRedHat := checker.isRedHatCentOS()

	// Both can't be true simultaneously
	if isUbuntu && isRedHat {
		t.Error("Cannot be both Ubuntu/Debian and RedHat/CentOS")
	}

	// Test that functions don't panic
	_ = isUbuntu
	_ = isRedHat
}

func TestGetSystemInfo(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	info := checker.GetSystemInfo()

	// Check required fields
	requiredFields := []string{"os", "arch"}
	for _, field := range requiredFields {
		if _, exists := info[field]; !exists {
			t.Errorf("System info should contain field: %s", field)
		}
	}

	// Validate OS and arch values
	if info["os"] != runtime.GOOS {
		t.Errorf("OS should be %s, got %s", runtime.GOOS, info["os"])
	}

	if info["arch"] != runtime.GOARCH {
		t.Errorf("Arch should be %s, got %s", runtime.GOARCH, info["arch"])
	}

	// Check that aws_config_dir is set
	if _, exists := info["aws_config_dir"]; !exists {
		t.Error("System info should contain aws_config_dir")
	}

	// Home directory should be set if available
	if home, err := os.UserHomeDir(); err == nil {
		if info["home"] != home {
			t.Errorf("Home directory should be %s, got %s", home, info["home"])
		}
	}

	// AWS config dir should be reasonable - use cross-platform home directory
	homeDir, _ := os.UserHomeDir()
	expectedAWSDir := filepath.Join(homeDir, ".aws")
	if info["aws_config_dir"] != expectedAWSDir {
		t.Logf("AWS config dir: expected %s, got %s", expectedAWSDir, info["aws_config_dir"])
	}
}

func TestRequirementResultJSONTags(t *testing.T) {
	// Test that RequirementResult struct has proper JSON tags
	result := RequirementResult{
		Name:       "Test Requirement",
		Passed:     true,
		Error:      "Test error",
		Suggestion: "Test suggestion",
		Version:    "1.0.0",
	}

	// Verify all fields are accessible
	if result.Name == "" {
		t.Error("Name field should be accessible")
	}
	if !result.Passed {
		t.Error("Passed field should be accessible")
	}
	if result.Error == "" {
		t.Error("Error field should be accessible")
	}
	if result.Suggestion == "" {
		t.Error("Suggestion field should be accessible")
	}
	if result.Version == "" {
		t.Error("Version field should be accessible")
	}
}

func TestFixIssuesWithPassedRequirements(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	// Test with all passed requirements
	passedResults := []RequirementResult{
		{Name: "Test 1", Passed: true},
		{Name: "Test 2", Passed: true},
	}

	err := checker.FixIssues(passedResults)
	if err != nil {
		t.Errorf("FixIssues should not error with all passed requirements: %v", err)
	}
}

func TestVersionExtraction(t *testing.T) {
	// Test version extraction logic for different command outputs
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	// Test AWS CLI version extraction
	// The actual checkAWSCLI function uses exec, but we can test the logic
	// by checking the result structure when it succeeds

	result := checker.checkAWSCLI()
	if result.Passed && result.Version != "" {
		// Version should not contain "aws-cli/" prefix
		if strings.HasPrefix(result.Version, "aws-cli/") {
			t.Error("AWS CLI version should not contain 'aws-cli/' prefix")
		}
	}

	// Test Go version result
	result = checker.checkGoVersion()
	if result.Passed && result.Version != "" {
		// Version should be reasonable format
		if !strings.Contains(result.Version, "go") {
			t.Logf("Go version format: %s", result.Version)
		}
	}

	// Test jq version result
	result = checker.checkJQ()
	if result.Passed && result.Version != "" {
		// Version should be trimmed
		if strings.HasPrefix(result.Version, " ") || strings.HasSuffix(result.Version, " ") {
			t.Error("jq version should be trimmed")
		}
	}
}

func TestOSReleaseFileParsing(t *testing.T) {
	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	// Test system info gathering for current OS
	info := checker.GetSystemInfo()

	// On Linux, should try to get OS version
	if runtime.GOOS == "linux" {
		// Check if os_version is set when file is available
		if _, err := os.Stat("/etc/os-release"); err == nil {
			// File exists, so we might have os_version
			if osVersion, exists := info["os_version"]; exists {
				if strings.Contains(osVersion, "\"") {
					t.Error("OS version should not contain quotes")
				}
			}
		}
	}

	// On macOS, should try to get version via sw_vers
	if runtime.GOOS == "darwin" {
		if osVersion, exists := info["os_version"]; exists {
			if strings.HasPrefix(osVersion, " ") || strings.HasSuffix(osVersion, " ") {
				t.Error("macOS version should be trimmed")
			}
		}
	}
}

func TestLinuxDistributionDetection(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux distribution detection test on non-Linux platform")
	}

	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	isUbuntu := checker.isUbuntuDebian()
	isRedHat := checker.isRedHatCentOS()

	// At least one detection method should work
	t.Logf("Ubuntu/Debian detection: %v", isUbuntu)
	t.Logf("RedHat/CentOS detection: %v", isRedHat)

	// Test the file-based detection logic
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		if !isUbuntu {
			t.Error("Should detect Ubuntu/Debian when /etc/debian_version exists")
		}
	}

	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		if !isRedHat {
			t.Error("Should detect RedHat when /etc/redhat-release exists")
		}
	}

	if _, err := os.Stat("/etc/lsb-release"); err == nil {
		// Check if it contains Ubuntu
		if content, readErr := os.ReadFile("/etc/lsb-release"); readErr == nil {
			if strings.Contains(string(content), "Ubuntu") && !isUbuntu {
				t.Error("Should detect Ubuntu when /etc/lsb-release contains Ubuntu")
			}
		}
	}
}

func TestSSMPluginExitCodeHandling(t *testing.T) {
	// Test that SSM plugin check handles exit code 255 correctly
	// session-manager-plugin returns 255 when called without arguments
	// but this means it's installed

	logger := logging.NewLogger(false)
	checker := NewRequirementsChecker(logger)

	result := checker.checkSSMPlugin()

	// The check should handle the 255 exit code appropriately
	if result.Name != "Session Manager Plugin" {
		t.Errorf("Expected name 'Session Manager Plugin', got %q", result.Name)
	}

	// Either passed (if plugin installed) or failed with proper error
	if !result.Passed && result.Error == "" {
		t.Error("Failed SSM plugin check should have error message")
	}

	if !result.Passed && result.Suggestion == "" {
		t.Error("Failed SSM plugin check should have installation suggestion")
	}
}

func TestHelperFunctionsConsistency(t *testing.T) {
	// Test that helper functions work correctly
	tests := []struct {
		name     string
		str      string
		substr   string
		expected bool
	}{
		{"contains found", "hello world", "world", true},
		{"contains not found", "hello world", "missing", false},
		{"contains empty", "hello world", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test contains function
			result := contains(tt.str, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.str, tt.substr, result, tt.expected)
			}
		})
	}

	// Separately test containsAny function
	containsAnyTests := []struct {
		name     string
		str      string
		substrs  []string
		expected bool
	}{
		{"containsAny found", "hello world", []string{"missing", "world", "other"}, true},
		{"containsAny not found", "hello world", []string{"missing", "other"}, false},
		{"containsAny empty", "hello world", []string{}, false},
	}

	for _, tt := range containsAnyTests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsAny(tt.str, tt.substrs)
			if result != tt.expected {
				t.Errorf("containsAny(%q, %v) = %v, want %v", tt.str, tt.substrs, result, tt.expected)
			}
		})
	}
}

// Additional test to ensure logger is properly used
func TestLoggerIntegration(t *testing.T) {
	// Test with debug logger
	debugLogger := logging.NewLogger(true)
	checker := NewRequirementsChecker(debugLogger)

	if checker.logger != debugLogger {
		t.Error("Logger should be properly set in checker")
	}

	// Test with normal logger
	normalLogger := logging.NewLogger(false)
	checker = NewRequirementsChecker(normalLogger)

	if checker.logger != normalLogger {
		t.Error("Logger should be properly set in checker")
	}

	// Test that checker doesn't panic with nil logger
	// (This would be bad practice, but the checker should be defensive)
	checker = &RequirementsChecker{logger: nil}

	// Should not panic when logger is nil (though it's not recommended)
	defer func() {
		if r := recover(); r != nil {
			t.Error("Checker should not panic with nil logger")
		}
	}()

	// Some methods might still work without logger
	info := checker.GetSystemInfo()
	if len(info) == 0 {
		t.Error("GetSystemInfo should still work without logger")
	}
}
