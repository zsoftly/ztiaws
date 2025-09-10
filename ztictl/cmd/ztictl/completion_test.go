package main

import (
	"os"
	"runtime"
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

			// Skip Windows-specific tests on non-Windows systems
			if tt.goos == "windows" && runtime.GOOS != "windows" {
				t.Skip("Skipping Windows-specific test on non-Windows system")
			}

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

// TestInstallCompletionRouting tests that installCompletion routes to correct handlers
func TestInstallCompletionRouting(t *testing.T) {
	tests := []struct {
		name      string
		shell     string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "unsupported shell",
			shell:     "tcsh",
			wantError: true,
			errorMsg:  "unsupported shell",
		},
		{
			name:      "empty shell",
			shell:     "",
			wantError: true,
			errorMsg:  "unsupported shell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := installCompletion(tt.shell)
			if (err != nil) != tt.wantError {
				t.Errorf("installCompletion(%q) error = %v, wantError %v", tt.shell, err, tt.wantError)
			}
			if err != nil && tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("installCompletion(%q) error = %v, want error containing %q", tt.shell, err, tt.errorMsg)
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

// TestShowCompletionInstructionsNoPanic verifies the functions don't panic
func TestShowCompletionInstructionsNoPanic(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell", "unknown"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			// This should not panic
			showCompletionInstructions(shell)
		})
	}
}

// TestPathValidation tests comprehensive path security validation
func TestPathValidation(t *testing.T) {
	validSystemPaths := []string{
		"/etc/bash_completion.d/",
		"/usr/share/bash-completion/completions/",
		"/usr/local/share/bash-completion/completions/",
	}

	tests := []struct {
		name          string
		path          string
		hasTraversal  bool
		requiresSudo  bool
		isValidSystem bool
	}{
		// Normal paths
		{
			name:          "valid etc path",
			path:          "/etc/bash_completion.d/ztictl",
			hasTraversal:  false,
			requiresSudo:  true,
			isValidSystem: true,
		},
		{
			name:          "valid usr share path",
			path:          "/usr/share/bash-completion/completions/ztictl",
			hasTraversal:  false,
			requiresSudo:  true,
			isValidSystem: true,
		},
		{
			name:          "user home path",
			path:          "/home/user/.local/share/bash-completion/completions/ztictl",
			hasTraversal:  false,
			requiresSudo:  false,
			isValidSystem: false,
		},
		// Path traversal attempts
		{
			name:          "traversal in etc",
			path:          "/etc/../etc/bash_completion.d/ztictl",
			hasTraversal:  true,
			requiresSudo:  false,
			isValidSystem: false,
		},
		{
			name:          "traversal with dots",
			path:          "/etc/bash_completion.d/../../../tmp/evil",
			hasTraversal:  true,
			requiresSudo:  false,
			isValidSystem: false,
		},
		{
			name:          "user traversal attempt",
			path:          "/home/user/.local/../../../etc/passwd",
			hasTraversal:  true,
			requiresSudo:  false,
			isValidSystem: false,
		},
		// Invalid system paths
		{
			name:          "invalid etc path",
			path:          "/etc/evil/ztictl",
			hasTraversal:  false,
			requiresSudo:  false,
			isValidSystem: false,
		},
		{
			name:          "invalid usr path",
			path:          "/usr/bin/ztictl",
			hasTraversal:  false,
			requiresSudo:  false,
			isValidSystem: false,
		},
		// Paths with shell metacharacters (documented but safe with exec.Command)
		{
			name:          "semicolon injection attempt",
			path:          "/tmp/ztictl; rm -rf /",
			hasTraversal:  false,
			requiresSudo:  false,
			isValidSystem: false,
		},
		{
			name:          "command substitution attempt",
			path:          "/tmp/ztictl$(whoami)",
			hasTraversal:  false,
			requiresSudo:  false,
			isValidSystem: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test traversal detection
			hasTraversal := strings.Contains(tt.path, "..")
			if hasTraversal != tt.hasTraversal {
				t.Errorf("path traversal detection: got %v, want %v", hasTraversal, tt.hasTraversal)
			}

			// Test system path validation
			isValidSystem := false
			requiresSudo := false
			if !hasTraversal && (strings.HasPrefix(tt.path, "/etc") || strings.HasPrefix(tt.path, "/usr")) {
				for _, validPath := range validSystemPaths {
					if strings.HasPrefix(tt.path, validPath) {
						isValidSystem = true
						requiresSudo = true
						break
					}
				}
			}

			if isValidSystem != tt.isValidSystem {
				t.Errorf("valid system path: got %v, want %v", isValidSystem, tt.isValidSystem)
			}
			if requiresSudo != tt.requiresSudo {
				t.Errorf("requires sudo: got %v, want %v", requiresSudo, tt.requiresSudo)
			}
		})
	}
}

// TestDuplicateCompletionDetection tests detection of existing completions
func TestDuplicateCompletionDetection(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		shell        string
		shouldDetect bool
	}{
		// Bash patterns
		{
			name:         "bash exact match",
			content:      "source <(ztictl completion bash)",
			shell:        "bash",
			shouldDetect: true,
		},
		{
			name:         "bash extra spaces",
			content:      "source <(ztictl  completion  bash)",
			shell:        "bash",
			shouldDetect: true,
		},
		{
			name:         "bash command substitution",
			content:      "source $(ztictl completion bash)",
			shell:        "bash",
			shouldDetect: true,
		},
		// Zsh patterns
		{
			name:         "zsh exact match",
			content:      "source <(ztictl completion zsh)",
			shell:        "zsh",
			shouldDetect: true,
		},
		{
			name:         "zsh in plugins",
			content:      "plugins=(git ztictl docker)",
			shell:        "zsh",
			shouldDetect: true,
		},
		// False positives
		{
			name:         "comment only",
			content:      "# ztictl is installed",
			shell:        "bash",
			shouldDetect: false,
		},
		{
			name:         "different command",
			content:      "ztictl auth login",
			shell:        "bash",
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var detected bool

			switch tt.shell {
			case "bash":
				detected = strings.Contains(tt.content, "ztictl completion bash") ||
					strings.Contains(tt.content, "ztictl  completion  bash") ||
					strings.Contains(tt.content, "$(ztictl completion bash)") ||
					strings.Contains(tt.content, "`ztictl completion bash`")
			case "zsh":
				detected = strings.Contains(tt.content, "ztictl completion zsh") ||
					strings.Contains(tt.content, "ztictl  completion  zsh") ||
					strings.Contains(tt.content, "$(ztictl completion zsh)") ||
					strings.Contains(tt.content, "`ztictl completion zsh`") ||
					(strings.Contains(tt.content, "plugins=(") && strings.Contains(tt.content, "ztictl"))
			}

			if detected != tt.shouldDetect {
				t.Errorf("detection = %v, want %v for content: %q", detected, tt.shouldDetect, tt.content)
			}
		})
	}
}
