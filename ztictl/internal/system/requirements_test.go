package system

import (
	"testing"
	"ztictl/internal/logging"
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
		Version: "1.0.0",
	}
	
	if result.Name != "Test" {
		t.Errorf("Expected name 'Test', got %q", result.Name)
	}
	
	if !result.Passed {
		t.Error("Expected Passed to be true")
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
