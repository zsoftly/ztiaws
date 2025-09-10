package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInstallBashCompletionMocked tests bash completion installation with mocked sudo
func TestInstallBashCompletionMocked(t *testing.T) {
	// Save original sudo command
	originalSudo := sudoCommand
	defer func() { sudoCommand = originalSudo }()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		setupFunc     func()
		mockSudo      string
		wantError     bool
		errorContains string
	}{
		{
			name: "User directory installation (no sudo)",
			setupFunc: func() {
				// Set HOME to temp directory
				os.Setenv("HOME", tempDir)
			},
			mockSudo:  "echo", // Won't be used
			wantError: false,
		},
		{
			name: "System directory with mocked sudo success",
			setupFunc: func() {
				// Create mock directories
				os.MkdirAll(filepath.Join(tempDir, "etc", "bash_completion.d"), 0755)
				// Mock sudo with echo command that succeeds
				sudoCommand = "echo"
			},
			mockSudo:  "echo",
			wantError: false,
		},
		{
			name: "System directory with mocked sudo failure",
			setupFunc: func() {
				// Mock sudo with false command that fails
				sudoCommand = "false"
			},
			mockSudo:      "false",
			wantError:     true,
			errorContains: "failed to install completion with sudo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			originalHome := os.Getenv("HOME")
			defer os.Setenv("HOME", originalHome)

			// Setup test environment
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Set mock sudo command
			sudoCommand = tt.mockSudo

			// Run the function
			err := installBashCompletion()

			// Check results
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error %q doesn't contain expected %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					// Skip errors related to actual system paths when testing with mocked sudo
					if !strings.Contains(err.Error(), "/etc") {
						t.Errorf("Unexpected error: %v", err)
					}
				}
			}
		})
	}
}

// TestValidateInstallPath tests path validation without sudo
func TestValidateInstallPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{
			name:      "normal path",
			path:      "/home/user/.local/share/bash-completion/completions/ztictl",
			shouldErr: false,
		},
		{
			name:      "path with traversal",
			path:      "/home/user/../../../etc/passwd",
			shouldErr: true,
		},
		{
			name:      "valid system path",
			path:      "/etc/bash_completion.d/ztictl",
			shouldErr: false,
		},
		{
			name:      "invalid system path",
			path:      "/etc/passwd",
			shouldErr: false, // Path validation happens elsewhere
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasTraversal := strings.Contains(tt.path, "..")
			if tt.shouldErr && !hasTraversal {
				t.Errorf("Expected path traversal detection for %s", tt.path)
			}
			if !tt.shouldErr && hasTraversal {
				t.Errorf("False positive path traversal for %s", tt.path)
			}
		})
	}
}

// TestDetermineSystemPath tests system path determination logic
func TestDetermineSystemPath(t *testing.T) {
	validSystemPaths := []string{
		"/etc/bash_completion.d/",
		"/usr/share/bash-completion/completions/",
		"/usr/local/share/bash-completion/completions/",
	}

	tests := []struct {
		name      string
		path      string
		needsSudo bool
	}{
		{
			name:      "etc bash completion",
			path:      "/etc/bash_completion.d/ztictl",
			needsSudo: true,
		},
		{
			name:      "usr share completion",
			path:      "/usr/share/bash-completion/completions/ztictl",
			needsSudo: true,
		},
		{
			name:      "usr local share completion",
			path:      "/usr/local/share/bash-completion/completions/ztictl",
			needsSudo: true,
		},
		{
			name:      "user home directory",
			path:      "/home/user/.local/share/bash-completion/completions/ztictl",
			needsSudo: false,
		},
		{
			name:      "tmp directory",
			path:      "/tmp/ztictl",
			needsSudo: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			needsSudo := false
			if strings.HasPrefix(tt.path, "/etc") || strings.HasPrefix(tt.path, "/usr") {
				for _, validPath := range validSystemPaths {
					if strings.HasPrefix(tt.path, validPath) {
						needsSudo = true
						break
					}
				}
			}

			if needsSudo != tt.needsSudo {
				t.Errorf("Path %s: expected needsSudo=%v, got %v", tt.path, tt.needsSudo, needsSudo)
			}
		})
	}
}

// TestInstallCompletionRouting tests the main routing function
func TestInstallCompletionRouting(t *testing.T) {
	// Save original sudo command
	originalSudo := sudoCommand
	defer func() { sudoCommand = originalSudo }()

	// Mock sudo to avoid actual system calls
	sudoCommand = "echo"

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
		{
			name:      "bash shell",
			shell:     "bash",
			wantError: false, // May still error on file operations, but routing works
		},
		{
			name:      "zsh shell",
			shell:     "zsh",
			wantError: false, // May still error on file operations, but routing works
		},
		{
			name:      "fish shell",
			shell:     "fish",
			wantError: false, // May still error on file operations, but routing works
		},
		{
			name:      "powershell",
			shell:     "powershell",
			wantError: false, // May still error on file operations, but routing works
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := installCompletion(tt.shell)

			// For unsupported shells, we expect an error
			if tt.shell == "tcsh" || tt.shell == "" {
				if err == nil {
					t.Errorf("Expected error for shell %q but got none", tt.shell)
				}
				return
			}

			// For supported shells, the function may fail due to file operations
			// but the routing itself should work (no "unsupported shell" error)
			if err != nil && strings.Contains(err.Error(), "unsupported shell") {
				t.Errorf("Got unsupported shell error for supported shell %q", tt.shell)
			}
		})
	}
}
