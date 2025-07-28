package splash

import (
	"os"
	"testing"
)

func TestShowSplash(t *testing.T) {
	// Use temporary directory for testing
	tempDir := t.TempDir()

	// Temporarily change home directory
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}

	// Set temp dir as home for this test
	if originalHome != "" {
		defer func() {
			if runtime := os.Getenv("HOME"); runtime == "" {
				os.Setenv("USERPROFILE", originalHome)
			} else {
				os.Setenv("HOME", originalHome)
			}
		}()
		os.Setenv("HOME", tempDir)
		os.Setenv("USERPROFILE", tempDir)
	}

	// Test first run - should return true (splash shown)
	shown, err := ShowSplash("2.0.0-test")
	if err != nil {
		t.Fatalf("ShowSplash failed: %v", err)
	}

	if !shown {
		t.Error("Expected splash to be shown on first run")
	}

	// Test subsequent run with same version - should return false
	shown, err = ShowSplash("2.0.0-test")
	if err != nil {
		t.Fatalf("ShowSplash failed on second call: %v", err)
	}

	if shown {
		t.Error("Expected splash NOT to be shown on subsequent run with same version")
	}

	// Test with new version - should return true
	shown, err = ShowSplash("2.1.0-test")
	if err != nil {
		t.Fatalf("ShowSplash failed with new version: %v", err)
	}

	if !shown {
		t.Error("Expected splash to be shown with new version")
	}
}

func TestShowBriefWelcome(t *testing.T) {
	// Test that ShowBriefWelcome doesn't panic
	ShowBriefWelcome("2.0.0-test")
}

func TestAnimateMessage(t *testing.T) {
	// Test that animateMessage doesn't panic with short message
	animateMessage("test")
}
