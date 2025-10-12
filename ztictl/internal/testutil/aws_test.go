package testutil

import (
	"os"
	"testing"
)

// TestMockAWSCredentials validates that mock credentials are properly defined
func TestMockAWSCredentials(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		minLen   int
		expected string
	}{
		{
			name:     "MockAWSAccessKeyID format",
			value:    MockAWSAccessKeyID,
			minLen:   20,
			expected: "AKIAIOSFODNN7EXAMPLE",
		},
		{
			name:     "MockAWSSecretAccessKey format",
			value:    MockAWSSecretAccessKey,
			minLen:   40,
			expected: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		{
			name:     "MockAWSSessionToken",
			value:    MockAWSSessionToken,
			minLen:   1,
			expected: "test-session-token",
		},
		{
			name:     "MockAWSRegion",
			value:    MockAWSRegion,
			minLen:   1,
			expected: "ca-central-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %s, want %s", tt.name, tt.value, tt.expected)
			}

			if len(tt.value) < tt.minLen {
				t.Errorf("%s length = %d, want at least %d", tt.name, len(tt.value), tt.minLen)
			}
		})
	}
}

// TestSetupAWSTestEnvironment validates that the setup function sets environment variables
func TestSetupAWSTestEnvironment(t *testing.T) {
	// Clean up environment before test
	envVars := []string{
		"AWS_EC2_METADATA_DISABLED",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_SESSION_TOKEN",
		"AWS_REGION",
	}

	// Store original values
	originals := make(map[string]string)
	for _, key := range envVars {
		originals[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	// Restore original values after test
	defer func() {
		for _, key := range envVars {
			if val, ok := originals[key]; ok && val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Run setup
	SetupAWSTestEnvironment()

	// Verify environment variables are set
	tests := []struct {
		name     string
		envVar   string
		expected string
	}{
		{
			name:     "EC2 metadata disabled",
			envVar:   "AWS_EC2_METADATA_DISABLED",
			expected: "true",
		},
		{
			name:     "Access key ID set",
			envVar:   "AWS_ACCESS_KEY_ID",
			expected: MockAWSAccessKeyID,
		},
		{
			name:     "Secret access key set",
			envVar:   "AWS_SECRET_ACCESS_KEY",
			expected: MockAWSSecretAccessKey,
		},
		{
			name:     "Session token set",
			envVar:   "AWS_SESSION_TOKEN",
			expected: MockAWSSessionToken,
		},
		{
			name:     "Region set",
			envVar:   "AWS_REGION",
			expected: MockAWSRegion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := os.Getenv(tt.envVar)
			if actual != tt.expected {
				t.Errorf("%s = %s, want %s", tt.envVar, actual, tt.expected)
			}
		})
	}
}

// TestMockCredentialsAreExamples ensures we're using AWS documentation examples
func TestMockCredentialsAreExamples(t *testing.T) {
	// These should match AWS documentation examples to ensure they're safe
	// Reference: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html

	if MockAWSAccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Error("MockAWSAccessKeyID should use AWS documentation example")
	}

	if MockAWSSecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Error("MockAWSSecretAccessKey should use AWS documentation example")
	}
}

// TestConstantsAreNotEmpty ensures all constants are defined
func TestConstantsAreNotEmpty(t *testing.T) {
	constants := map[string]string{
		"MockAWSAccessKeyID":     MockAWSAccessKeyID,
		"MockAWSSecretAccessKey": MockAWSSecretAccessKey,
		"MockAWSSessionToken":    MockAWSSessionToken,
		"MockAWSRegion":          MockAWSRegion,
	}

	for name, value := range constants {
		if value == "" {
			t.Errorf("%s is empty", name)
		}
	}
}
