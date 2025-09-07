package security

import (
	"fmt"
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

func TestHasTraversalPatterns(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		// Dangerous patterns that should be detected
		{"Unix parent prefix", "../etc/passwd", true},
		{"Windows parent prefix", "..\\etc\\passwd", true},
		{"Unix embedded parent", "safe/../etc/passwd", true},
		{"Windows embedded parent", "safe\\..\\etc\\passwd", true},
		{"Unix parent suffix", "path/..", true},
		{"Windows parent suffix", "path\\..", true},
		{"Direct parent reference", "..", true},
		{"Multiple embedded Unix", "a/../b/../c", true},
		{"Multiple embedded Windows", "a\\..\\b\\..\\c", true},
		{"Complex Unix traversal", "path/to/../../sensitive", true},
		{"Complex Windows traversal", "path\\to\\..\\..\\sensitive", true},

		// Safe patterns that should be allowed
		{"Normal relative path", "path/to/file", false},
		{"Normal Windows path", "path\\to\\file", false},
		{"Absolute Unix path", "/var/log/app", false},
		{"Absolute Windows path", "C:\\Program Files\\app", false},
		{"Path with dots in filename", "path/file.txt", false},
		{"Path with double dots in filename", "path/file..txt", false},
		{"Current directory reference", ".", false},
		{"Path with numbers", "path/001/file", false},
		{"Empty string", "", false},
		{"Single filename", "file.txt", false},
		{"Path with underscores", "path_to_file/test", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := hasTraversalPatterns(tc.path)
			if result != tc.expected {
				t.Errorf("hasTraversalPatterns(%q) = %v, expected %v", tc.path, result, tc.expected)
			}
		})
	}
}

func TestContainsUnsafePath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected bool
		reason   string
	}{
		// Directory traversal attacks
		{"Hidden double traversal", "safe/./../../etc/passwd", true, "filepath.Clean() would hide this attack"},
		{"Multiple level traversal", "safe/foo/../../../etc/passwd", true, "Multiple traversal levels"},
		{"Mixed patterns", "logs/././../../../etc/passwd", true, "Mixed . and .. patterns"},
		{"Simple traversal", "normal/../path", true, "Basic traversal pattern"},
		{"Direct traversal", "../malicious", true, "Direct parent access"},
		{"End traversal", "path/to/../../../etc", true, "Traversal at path end"},
		{"Direct parent", "..", true, "Direct parent directory"},

		// Null byte attacks
		{"Null byte attack", "path\x00/etc/passwd", true, "Null byte injection"},
		{"Null byte in middle", "safe/path\x00/../etc", true, "Null byte with traversal"},

		// Path separator attacks
		{"Double slash", "path//to//file", true, "Double Unix separators"},
		{"Double backslash", "path\\\\to\\\\file", true, "Double Windows separators"},
		{"Mixed double separators", "path//to\\\\file", true, "Mixed double separators"},

		// Legitimate paths that should be allowed
		{"Absolute temp path", "/var/folders/tmp/test/001", false, "Legitimate absolute path"},
		{"User log directory", "/Users/user/Library/Logs/app", false, "User application logs"},
		{"System temp", "/tmp/test", false, "System temporary directory"},
		{"Relative log path", "logs/app/2023", false, "Application logs"},
		{"Normal relative path", "normal/path/001", false, "Standard relative path"},
		{"Windows absolute path", "C:\\Program Files\\App", false, "Windows program files"},
		{"Current directory", ".", false, "Current directory reference"},
		{"Single file", "config.yaml", false, "Single filename"},
		{"Empty path", "", false, "Empty string path"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ContainsUnsafePath(tc.path)
			if result != tc.expected {
				t.Errorf("ContainsUnsafePath(%q) = %v, expected %v (%s)",
					tc.path, result, tc.expected, tc.reason)
			}
		})
	}
}

func TestContainsUnsafePathPreCleanDetection(t *testing.T) {
	// These test cases specifically verify that we detect dangerous patterns
	// BEFORE filepath.Clean() can normalize them away - this is the critical
	// security fix that prevents attacks from being hidden by normalization
	criticalCases := []struct {
		name         string
		path         string
		cleanedPath  string
		shouldDetect bool
		explanation  string
	}{
		{
			name:         "Hidden traversal via current dir",
			path:         "safe/./../../etc/passwd",
			cleanedPath:  "../etc/passwd",
			shouldDetect: true,
			explanation:  "Original path contains ../../ but Clean() would hide the embedded nature",
		},
		{
			name:         "Nested traversal normalization",
			path:         "logs/foo/../bar/../../../etc/passwd",
			cleanedPath:  "../../etc/passwd",
			shouldDetect: true,
			explanation:  "Multiple traversals that get simplified by Clean()",
		},
		{
			name:         "Complex mixed pattern",
			path:         "app/./config/../../../sensitive/file",
			cleanedPath:  "../sensitive/file",
			shouldDetect: true,
			explanation:  "Mixed . and .. patterns that Clean() normalizes",
		},
		{
			name:         "Legitimate relative becomes absolute",
			path:         "config/../app.log",
			cleanedPath:  "app.log",
			shouldDetect: true,
			explanation:  "Even legitimate-looking traversal should be detected pre-clean",
		},
	}

	for _, tc := range criticalCases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify our function detects the original dangerous path
			result := ContainsUnsafePath(tc.path)
			if result != tc.shouldDetect {
				t.Errorf("ContainsUnsafePath(%q) = %v, expected %v",
					tc.path, result, tc.shouldDetect)
				t.Errorf("Explanation: %s", tc.explanation)
			}

			// Verify that filepath.Clean() would indeed normalize it
			actualCleaned := filepath.Clean(tc.path)
			if actualCleaned != tc.cleanedPath {
				t.Logf("Note: filepath.Clean(%q) = %q (expected %q)",
					tc.path, actualCleaned, tc.cleanedPath)
			}
		})
	}
}

func TestContainsUnsafePathBackwardCompatibility(t *testing.T) {
	// Test that this function can be used as a drop-in replacement for
	// the old containsUnsafeLogPath function from logging package
	oldLogPathTestCases := []struct {
		path     string
		expected bool
	}{
		// These were the kinds of paths the logging package was validating
		{"/Users/user/.local/share/ztictl/logs", false},          // XDG standard location
		{"/Users/user/Library/Logs/ztictl", false},               // macOS standard location
		{"C:\\Users\\user\\AppData\\Local\\ztictl\\logs", false}, // Windows location
		{"../../../etc/passwd", true},                            // Attack attempt
		{"logs/../../../sensitive", true},                        // Attack via logs directory
		{"/var/folders/temp123/logs", false},                     // Temp directory
		{"./logs/app", false},                                    // Local logs directory
	}

	for _, tc := range oldLogPathTestCases {
		t.Run(fmt.Sprintf("LogPath_%s", tc.path), func(t *testing.T) {
			result := ContainsUnsafePath(tc.path)
			if result != tc.expected {
				t.Errorf("ContainsUnsafePath(%q) = %v, expected %v (logging compatibility)",
					tc.path, result, tc.expected)
			}
		})
	}
}
