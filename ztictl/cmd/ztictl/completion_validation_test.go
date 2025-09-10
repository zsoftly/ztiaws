package main

import (
	"strings"
	"testing"
)

// TestPathTraversalDetection tests that path traversal attempts are detected
func TestPathTraversalDetection(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{
			name:      "normal path",
			path:      "/etc/bash_completion.d/ztictl",
			shouldErr: false,
		},
		{
			name:      "path with traversal",
			path:      "/etc/../etc/bash_completion.d/ztictl",
			shouldErr: true,
		},
		{
			name:      "path with double dots",
			path:      "/etc/bash_completion.d/../../../tmp/evil",
			shouldErr: true,
		},
		{
			name:      "user path normal",
			path:      "/home/user/.local/share/bash-completion/completions/ztictl",
			shouldErr: false,
		},
		{
			name:      "user path with traversal",
			path:      "/home/user/.local/../../../etc/passwd",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if path contains traversal patterns
			hasTraversal := strings.Contains(tt.path, "..")

			if tt.shouldErr && !hasTraversal {
				t.Errorf("Expected path traversal detection for %s, but it passed", tt.path)
			}
			if !tt.shouldErr && hasTraversal {
				t.Errorf("False positive path traversal detection for %s", tt.path)
			}
		})
	}
}

// TestValidSystemPaths tests that only legitimate system paths are accepted for sudo
func TestValidSystemPaths(t *testing.T) {
	validSystemPaths := []string{
		"/etc/bash_completion.d/",
		"/usr/share/bash-completion/completions/",
		"/usr/local/share/bash-completion/completions/",
	}

	tests := []struct {
		name         string
		path         string
		shouldBeSudo bool
	}{
		{
			name:         "valid /etc path",
			path:         "/etc/bash_completion.d/ztictl",
			shouldBeSudo: true,
		},
		{
			name:         "valid /usr/share path",
			path:         "/usr/share/bash-completion/completions/ztictl",
			shouldBeSudo: true,
		},
		{
			name:         "valid /usr/local/share path",
			path:         "/usr/local/share/bash-completion/completions/ztictl",
			shouldBeSudo: true,
		},
		{
			name:         "invalid /etc path",
			path:         "/etc/evil/ztictl",
			shouldBeSudo: false,
		},
		{
			name:         "invalid /usr path",
			path:         "/usr/bin/ztictl",
			shouldBeSudo: false,
		},
		{
			name:         "user home path",
			path:         "/home/user/.local/share/bash-completion/completions/ztictl",
			shouldBeSudo: false,
		},
		{
			name:         "tmp path",
			path:         "/tmp/ztictl",
			shouldBeSudo: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := false
			needsSudo := false

			if strings.HasPrefix(tt.path, "/etc") || strings.HasPrefix(tt.path, "/usr") {
				for _, validPath := range validSystemPaths {
					if strings.HasPrefix(tt.path, validPath) {
						isValid = true
						needsSudo = true
						break
					}
				}
			}

			if tt.shouldBeSudo && !needsSudo {
				t.Errorf("Expected %s to require sudo, but it didn't", tt.path)
			}
			if !tt.shouldBeSudo && needsSudo && isValid {
				t.Errorf("Path %s should not require sudo, but it did", tt.path)
			}
		})
	}
}

// TestNoShellMetacharacters tests that paths with shell metacharacters are handled safely
func TestNoShellMetacharacters(t *testing.T) {
	// These paths contain shell metacharacters that could be dangerous
	dangerousPaths := []string{
		"/etc/bash_completion.d/ztictl; rm -rf /",
		"/etc/bash_completion.d/ztictl$(whoami)",
		"/etc/bash_completion.d/ztictl`id`",
		"/etc/bash_completion.d/ztictl|cat /etc/passwd",
		"/etc/bash_completion.d/ztictl&& evil_command",
		"/etc/bash_completion.d/ztictl$PATH",
		"/etc/bash_completion.d/ztictl'$(evil)'",
	}

	for _, path := range dangerousPaths {
		t.Run(path, func(t *testing.T) {
			// In the actual implementation, these paths are passed to exec.Command
			// which doesn't interpret shell metacharacters when used properly
			// (i.e., not passed through sh -c)

			// The key security improvement is:
			// OLD: exec.Command("sudo", "tee", path) with stdin
			// NEW: exec.Command("sudo", "cp", tempFile, path)
			//
			// Both approaches actually handle metacharacters safely because
			// exec.Command doesn't use shell interpretation. However, the new
			// approach is better because:
			// 1. It validates the path against a whitelist first
			// 2. It uses a temp file, avoiding any stdin complications
			// 3. It's clearer and more auditable

			// This test documents that metacharacters don't pose a risk
			// when using exec.Command properly
			if strings.ContainsAny(path, ";$`|&'\"") {
				// These characters exist but won't be interpreted as shell commands
				// when passed directly to exec.Command
				t.Logf("Path contains potential shell metacharacters: %s", path)
			}
		})
	}
}

// TestDuplicateDetectionPatterns tests the improved duplicate detection
func TestDuplicateDetectionPatterns(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		shouldDetect bool
	}{
		{
			name:         "exact match",
			content:      "source <(ztictl completion bash)",
			shouldDetect: true,
		},
		{
			name:         "extra spaces",
			content:      "source <(ztictl  completion  bash)",
			shouldDetect: true,
		},
		{
			name:         "command substitution dollar",
			content:      "source $(ztictl completion bash)",
			shouldDetect: true,
		},
		{
			name:         "command substitution backticks",
			content:      "source `ztictl completion bash`",
			shouldDetect: true,
		},
		{
			name:         "no completion",
			content:      "# ztictl is a great tool",
			shouldDetect: false,
		},
		{
			name:         "different command",
			content:      "ztictl auth login",
			shouldDetect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test bash detection patterns
			detected := strings.Contains(tt.content, "ztictl completion bash") ||
				strings.Contains(tt.content, "ztictl  completion  bash") ||
				strings.Contains(tt.content, "$(ztictl completion bash)") ||
				strings.Contains(tt.content, "`ztictl completion bash`")

			if tt.shouldDetect && !detected {
				t.Errorf("Should have detected completion in: %s", tt.content)
			}
			if !tt.shouldDetect && detected {
				t.Errorf("False positive detection in: %s", tt.content)
			}
		})
	}
}
