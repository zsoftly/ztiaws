package security

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "security_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		targetPath  string
		baseDir     string
		shouldError bool
		description string
	}{
		{
			name:        "Valid path within base directory",
			targetPath:  filepath.Join(tempDir, "valid", "file.txt"),
			baseDir:     tempDir,
			shouldError: false,
			description: "Path within base directory should be allowed",
		},
		{
			name:        "Simple parent directory traversal",
			targetPath:  filepath.Join(tempDir, "..", "malicious.txt"),
			baseDir:     tempDir,
			shouldError: true,
			description: "Simple ../ traversal should be blocked",
		},
		{
			name:        "Nested directory traversal",
			targetPath:  filepath.Join(tempDir, "subdir", "..", "..", "malicious.txt"),
			baseDir:     tempDir,
			shouldError: true,
			description: "Nested ../ traversal should be blocked",
		},
		{
			name:        "Embedded directory traversal",
			targetPath:  filepath.Join(tempDir, "foo", "..", "..", "malicious.txt"),
			baseDir:     tempDir,
			shouldError: true,
			description: "Embedded ../ in path should be blocked",
		},
		{
			name:        "Path ending with dotdot that escapes",
			targetPath:  filepath.Join(tempDir, "..", "outside.txt"),
			baseDir:     tempDir,
			shouldError: true,
			description: "Path ending with .. that escapes should be blocked",
		},
		{
			name:        "Complex traversal ending with dotdot",
			targetPath:  filepath.Join(tempDir, "subdir", "..", ".."),
			baseDir:     tempDir,
			shouldError: true,
			description: "Complex path ending with .. should be blocked if it escapes",
		},
		{
			name:        "Direct parent reference",
			targetPath:  "..",
			baseDir:     tempDir,
			shouldError: true,
			description: "Direct .. reference should be blocked",
		},
		{
			name:        "Absolute path outside base",
			targetPath:  "/etc/passwd",
			baseDir:     tempDir,
			shouldError: true,
			description: "Absolute path outside base should be blocked",
		},
		{
			name:        "Current directory reference",
			targetPath:  filepath.Join(tempDir, ".", "file.txt"),
			baseDir:     tempDir,
			shouldError: false,
			description: "Current directory reference should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.targetPath, tt.baseDir)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none for %s: %s", tt.name, tt.description)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got %v for %s: %s", err, tt.name, tt.description)
			}
		})
	}
}

func TestValidateFilePathWithWorkingDir(t *testing.T) {
	// Test the working directory convenience function

	tests := []struct {
		name        string
		targetPath  string
		shouldError bool
	}{
		{
			name:        "Valid relative path",
			targetPath:  "testfile.txt",
			shouldError: false,
		},
		{
			name:        "Parent directory traversal",
			targetPath:  "../../../etc/passwd",
			shouldError: true,
		},
		{
			name:        "Valid subdirectory path",
			targetPath:  "subdir/file.txt",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePathWithWorkingDir(tt.targetPath)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none for %s", tt.name)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got %v for %s", err, tt.name)
			}
		})
	}
}

func TestValidateFilePathEdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "security_edge_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test edge cases that could potentially bypass validation
	edgeCases := []struct {
		name        string
		targetPath  string
		shouldBlock bool
	}{
		{
			name:        "Empty string",
			targetPath:  "",
			shouldBlock: false, // Empty string resolves to current directory
		},
		{
			name:        "Dot slash prefix",
			targetPath:  "./subdir/file.txt",
			shouldBlock: false,
		},
		{
			name:        "Multiple dots",
			targetPath:  "...txt",
			shouldBlock: false, // Not a traversal, just a filename with dots
		},
		{
			name:        "Traversal escaping base",
			targetPath:  filepath.Join("..", "outside.txt"),
			shouldBlock: true,
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			fullPath := filepath.Join(tempDir, tc.targetPath)
			err := ValidateFilePath(fullPath, tempDir)

			if tc.shouldBlock && err == nil {
				t.Errorf("Expected %s to be blocked but it was allowed", tc.name)
			}

			if !tc.shouldBlock && err != nil {
				t.Errorf("Expected %s to be allowed but got error: %v", tc.name, err)
			}
		})
	}
}

func TestValidateFilePathCrossPlatform(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "security_cross_platform_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Cross-platform tests for directory traversal patterns
	crossPlatformTests := []struct {
		name        string
		targetPath  string
		shouldError bool
		description string
	}{
		{
			name:        "Unix style parent directory traversal",
			targetPath:  filepath.Join(tempDir, "..", "malicious.txt"),
			shouldError: true,
			description: "Unix-style ../ traversal should be blocked",
		},
		{
			name:        "Windows style parent directory traversal",
			targetPath:  filepath.Join(tempDir, "..", "malicious.txt"),
			shouldError: true,
			description: "Windows-style ..\\ traversal should be blocked",
		},
		{
			name:        "Mixed separator traversal",
			targetPath:  filepath.Join(tempDir, "..", "subdir", "..", "malicious.txt"),
			shouldError: true,
			description: "Mixed separator traversal should be blocked",
		},
		{
			name:        "Windows embedded traversal",
			targetPath:  filepath.Join(tempDir, "subdir", "..", "..", "malicious.txt"),
			shouldError: true,
			description: "Windows embedded ..\\..\\ traversal should be blocked",
		},
		{
			name:        "Windows path ending with dotdot",
			targetPath:  filepath.Join(tempDir, "subdir", "..", ".."),
			shouldError: true,
			description: "Windows path ending with ..\\ should be blocked if it escapes",
		},
	}

	for _, tt := range crossPlatformTests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.targetPath, tempDir)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none for %s: %s", tt.name, tt.description)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got %v for %s: %s", err, tt.name, tt.description)
			}
		})
	}
}

func TestValidateFilePathPlatformSpecific(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "security_platform_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if runtime.GOOS == "windows" {
		// Windows-specific tests
		windowsTests := []struct {
			name        string
			targetPath  string
			shouldError bool
			description string
		}{
			{
				name:        "Windows drive letter absolute path",
				targetPath:  "C:\\Windows\\System32\\config\\SAM",
				shouldError: true,
				description: "Windows absolute path outside base should be blocked",
			},
			{
				name:        "Windows UNC path",
				targetPath:  "\\\\server\\share\\file.txt",
				shouldError: true,
				description: "Windows UNC path should be blocked",
			},
			{
				name:        "Windows backslash only traversal",
				targetPath:  filepath.Join(tempDir, "..", "..", "sensitive.txt"),
				shouldError: true,
				description: "Windows backslash traversal should be blocked",
			},
		}

		for _, tt := range windowsTests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateFilePath(tt.targetPath, tempDir)

				if tt.shouldError && err == nil {
					t.Errorf("Expected error but got none for %s: %s", tt.name, tt.description)
				}

				if !tt.shouldError && err != nil {
					t.Errorf("Expected no error but got %v for %s: %s", err, tt.name, tt.description)
				}
			})
		}
	} else {
		// Unix-specific tests
		unixTests := []struct {
			name        string
			targetPath  string
			shouldError bool
			description string
		}{
			{
				name:        "Unix absolute system path",
				targetPath:  "/etc/shadow",
				shouldError: true,
				description: "Unix absolute system path should be blocked",
			},
			{
				name:        "Unix home directory traversal",
				targetPath:  filepath.Join(tempDir, "..", "..", "..", "home", "user", ".ssh", "id_rsa"),
				shouldError: true,
				description: "Unix home directory traversal should be blocked",
			},
			{
				name:        "Unix proc filesystem access",
				targetPath:  "/proc/self/environ",
				shouldError: true,
				description: "Unix proc filesystem access should be blocked",
			},
		}

		for _, tt := range unixTests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateFilePath(tt.targetPath, tempDir)

				if tt.shouldError && err == nil {
					t.Errorf("Expected error but got none for %s: %s", tt.name, tt.description)
				}

				if !tt.shouldError && err != nil {
					t.Errorf("Expected no error but got %v for %s: %s", err, tt.name, tt.description)
				}
			})
		}
	}
}
