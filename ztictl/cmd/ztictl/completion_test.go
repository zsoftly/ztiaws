package main

import (
	"os"
	"strings"
	"testing"
)

// TestDetectShell tests the shell detection functionality
func TestDetectShell(t *testing.T) {
	tests := []struct {
		name        string
		envShell    string
		psModule    string
		goos        string
		wantShell   string
		wantPresent bool
	}{
		{
			name:        "detect bash from SHELL env",
			envShell:    "/bin/bash",
			psModule:    "",
			goos:        "linux",
			wantShell:   "bash",
			wantPresent: true,
		},
		{
			name:        "detect zsh from SHELL env",
			envShell:    "/usr/bin/zsh",
			psModule:    "",
			goos:        "linux",
			wantShell:   "zsh",
			wantPresent: true,
		},
		{
			name:        "detect fish from SHELL env",
			envShell:    "/usr/local/bin/fish",
			psModule:    "",
			goos:        "linux",
			wantShell:   "fish",
			wantPresent: true,
		},
		{
			name:        "detect PowerShell on Windows",
			envShell:    "",
			psModule:    "C:\\Program Files\\PowerShell",
			goos:        "windows",
			wantShell:   "powershell",
			wantPresent: true,
		},
		{
			name:        "no shell detected",
			envShell:    "",
			psModule:    "",
			goos:        "linux",
			wantShell:   "",
			wantPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			origShell := os.Getenv("SHELL")
			origPSModule := os.Getenv("PSModulePath")

			// Set test environment
			os.Setenv("SHELL", tt.envShell)
			os.Setenv("PSModulePath", tt.psModule)
			// Note: We can't actually change runtime.GOOS, so we'll skip Windows-specific tests on non-Windows

			// Run detection
			shell := detectShell()

			// Restore environment
			os.Setenv("SHELL", origShell)
			os.Setenv("PSModulePath", origPSModule)

			// Check result
			if tt.wantPresent && shell != tt.wantShell {
				t.Errorf("detectShell() = %q, want %q", shell, tt.wantShell)
			}
			if !tt.wantPresent && shell != "" {
				t.Errorf("detectShell() = %q, want empty string", shell)
			}
		})
	}
}

// TestInstallCompletion tests the installCompletion function routing
func TestInstallCompletion(t *testing.T) {
	tests := []struct {
		name      string
		shell     string
		wantError bool
	}{
		{
			name:      "unsupported shell",
			shell:     "tcsh",
			wantError: true,
		},
		{
			name:      "empty shell",
			shell:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := installCompletion(tt.shell)
			if (err != nil) != tt.wantError {
				t.Errorf("installCompletion(%q) error = %v, wantError %v", tt.shell, err, tt.wantError)
			}
		})
	}
}

// TestCompletionCommand tests the completion command initialization
func TestCompletionCommand(t *testing.T) {
	// Test that the command is properly configured
	if completionCmd == nil {
		t.Fatal("completionCmd is nil")
	}

	if completionCmd.Use != "completion [bash|zsh|fish|powershell]" {
		t.Errorf("unexpected Use field: %s", completionCmd.Use)
	}

	// Check that valid args are set
	validArgs := completionCmd.ValidArgs
	expectedArgs := []string{"bash", "zsh", "fish", "powershell"}

	if len(validArgs) != len(expectedArgs) {
		t.Errorf("ValidArgs length = %d, want %d", len(validArgs), len(expectedArgs))
	}

	for _, expected := range expectedArgs {
		found := false
		for _, arg := range validArgs {
			if arg == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ValidArgs missing %q", expected)
		}
	}

	// Check that install flag exists
	installFlag := completionCmd.Flags().Lookup("install")
	if installFlag == nil {
		t.Error("install flag not found")
	}
	if installFlag.Shorthand != "i" {
		t.Errorf("install flag shorthand = %q, want 'i'", installFlag.Shorthand)
	}
}

// TestShowCompletionInstructions tests that instructions are generated without panics
func TestShowCompletionInstructions(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell", "unknown"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			// This should not panic
			showCompletionInstructions(shell)
		})
	}
}

// TestCompletionFunctionHelpers verifies helper functions don't panic
func TestCompletionFunctionHelpers(t *testing.T) {
	// These should not panic even if called directly
	t.Run("showBashInstructions", func(t *testing.T) {
		showBashInstructions()
	})

	t.Run("showZshInstructions", func(t *testing.T) {
		showZshInstructions()
	})

	t.Run("showFishInstructions", func(t *testing.T) {
		showFishInstructions()
	})

	t.Run("showPowerShellInstructions", func(t *testing.T) {
		showPowerShellInstructions()
	})
}

// TestCompletionInstructionsContainKeyInfo tests that instructions include essential information
func TestCompletionInstructionsContainKeyInfo(t *testing.T) {
	// Capture output for bash instructions
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showBashInstructions()

	w.Close()
	os.Stdout = oldStdout

	// Read the output
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Check for key content
	expectedStrings := []string{
		"BASH COMPLETION SETUP",
		"--install",
		"source <(ztictl completion bash)",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Bash instructions missing expected content: %q", expected)
		}
	}
}

// TestNoHardcodedCompletions verifies we removed environment-specific completions
func TestNoHardcodedCompletions(t *testing.T) {
	// Check that ssm exec command doesn't have ValidArgsFunction
	if ssmExecCmd.ValidArgsFunction != nil {
		t.Error("ssmExecCmd still has ValidArgsFunction - hardcoded completions should be removed")
	}
}
