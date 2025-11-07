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
				_ = os.Setenv("USERPROFILE", originalHome) // #nosec G104
			} else {
				_ = os.Setenv("HOME", originalHome) // #nosec G104
			}
		}()
		_ = os.Setenv("HOME", tempDir)        // #nosec G104
		_ = os.Setenv("USERPROFILE", tempDir) // #nosec G104
	}

	// When running tests, there's no terminal attached, so splash should not be shown
	// This tests that the function correctly detects non-terminal environments
	shown, err := ShowSplash("2.1.0-test")
	if err != nil {
		t.Fatalf("ShowSplash failed: %v", err)
	}

	if shown {
		t.Error("Expected splash NOT to be shown in non-terminal environment (test)")
	}

	// Verify it consistently returns false in non-terminal environments
	shown, err = ShowSplash("2.1.0-test")
	if err != nil {
		t.Fatalf("ShowSplash failed on second call: %v", err)
	}

	if shown {
		t.Error("Expected splash NOT to be shown on subsequent run in non-terminal environment")
	}

	// Even with a new version, splash should not be shown in non-terminal environments
	shown, err = ShowSplash("2.2.0-test")
	if err != nil {
		t.Fatalf("ShowSplash failed with new version: %v", err)
	}

	if shown {
		t.Error("Expected splash NOT to be shown with new version in non-terminal environment")
	}
}

func TestShowBriefWelcome(t *testing.T) {
	// Test that ShowBriefWelcome doesn't panic
	ShowBriefWelcome("2.1.0-test")
}

func TestAnimateMessage(t *testing.T) {
	// Test that animateMessage doesn't panic with short message
	animateMessage("test")
}
