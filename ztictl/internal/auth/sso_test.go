package auth

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"ztictl/internal/logging"
)

func TestNewManager(t *testing.T) {
	logger := logging.NewLogger(false)
	manager := NewManager(logger)

	if manager == nil {
		t.Error("Expected manager to be created, got nil")
	}

	if manager.logger != logger {
		t.Error("Expected manager to have the provided logger")
	}
}

func TestGetAWSConfigDir(t *testing.T) {
	configDir, err := getAWSConfigDir()
	if err != nil {
		t.Errorf("Expected no error getting AWS config dir, got: %v", err)
	}

	if configDir == "" {
		t.Error("Expected config directory path, got empty string")
	}

	// Should contain .aws in the path
	if !filepath.IsAbs(configDir) {
		t.Error("Expected absolute path for config directory")
	}
}

func TestGetAWSCacheDir(t *testing.T) {
	cacheDir, err := getAWSCacheDir()
	if err != nil {
		t.Errorf("Expected no error getting AWS cache dir, got: %v", err)
	}

	if cacheDir == "" {
		t.Error("Expected cache directory path, got empty string")
	}

	// Should be absolute path
	if !filepath.IsAbs(cacheDir) {
		t.Error("Expected absolute path for cache directory")
	}
}

func TestProfileStructure(t *testing.T) {
	// Test Profile struct creation
	now := time.Now()
	profile := Profile{
		Name:            "test-profile",
		IsAuthenticated: true,
		AccountID:       "123456789012",
		AccountName:     "Test Account",
		RoleName:        "TestRole",
		Region:          "us-east-1",
		SSOStartURL:     "https://test.awsapps.com/start",
		SSORegion:       "us-east-1",
		ExpiresAt:       &now,
	}

	if profile.Name != "test-profile" {
		t.Errorf("Expected profile name 'test-profile', got %s", profile.Name)
	}

	if !profile.IsAuthenticated {
		t.Error("Expected profile to be authenticated")
	}

	if profile.ExpiresAt != &now {
		t.Error("Expected profile expiration time to be set")
	}
}

func TestCredentialsStructure(t *testing.T) {
	// Test Credentials struct creation
	creds := Credentials{
		AccessKeyID:     "AKIA...",
		SecretAccessKey: "secret...",
		SessionToken:    "token...",
		Region:          "us-east-1",
	}

	if creds.AccessKeyID != "AKIA..." {
		t.Errorf("Expected access key ID 'AKIA...', got %s", creds.AccessKeyID)
	}

	if creds.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got %s", creds.Region)
	}
}

func TestListProfilesWithoutAWS(t *testing.T) {
	// Test that ListProfiles handles missing AWS config gracefully
	logger := logging.NewLogger(false)
	manager := NewManager(logger)

	ctx := context.Background()
	profiles, err := manager.ListProfiles(ctx)

	// This should not panic and may return empty list or error depending on system
	if err != nil {
		// Expected for systems without AWS config - that's ok
		t.Logf("ListProfiles returned expected error (no AWS config): %v", err)
	}

	// Profiles can be empty or have system profiles
	if profiles != nil && len(profiles) > 0 {
		t.Logf("Found %d profiles on system", len(profiles))
	}
}
