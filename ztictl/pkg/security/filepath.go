package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateFilePath ensures the path is within the allowed base directory
// This prevents directory traversal attacks (CWE-22)
func ValidateFilePath(targetPath, baseDir string) error {
	// Clean and resolve paths
	cleanTarget, err := filepath.Abs(filepath.Clean(targetPath))
	if err != nil {
		return fmt.Errorf("failed to resolve target path: %w", err)
	}

	cleanBase, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}

	// Check if target is within base directory
	relPath, err := filepath.Rel(cleanBase, cleanTarget)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}

	// Robust directory traversal validation
	// Check for absolute paths (indicates escape) or any .. path segments
	// Handle both Unix (/) and Windows (\) path separators for security
	if containsDirectoryTraversal(relPath) {
		return fmt.Errorf("path escapes base directory: %s", targetPath)
	}

	return nil
}

// ValidateFilePathWithWorkingDir validates file path against current working directory
// This is a convenience wrapper for the common case of validating against CWD
func ValidateFilePathWithWorkingDir(targetPath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	return ValidateFilePath(targetPath, cwd)
}

// containsDirectoryTraversal checks if a relative path contains directory traversal patterns
// that could be used to escape the base directory. This function handles both Unix (/)
// and Windows (\) path separators for comprehensive cross-platform security.
func containsDirectoryTraversal(relPath string) bool {
	// Check if path is absolute (indicates escape from relative base)
	isAbsolutePath := filepath.IsAbs(relPath)

	// Use the shared traversal pattern detection logic
	return isAbsolutePath || hasTraversalPatterns(relPath)
}

// hasTraversalPatterns detects directory traversal patterns in a given path.
// This is a reusable helper function that checks for all common traversal patterns
// including parent directory references, embedded traversals, and direct parent access.
func hasTraversalPatterns(path string) bool {
	// Check for parent directory reference at start of path
	hasUnixParentPrefix := strings.HasPrefix(path, "../")
	hasWindowsParentPrefix := strings.HasPrefix(path, "..\\")
	hasParentPrefix := hasUnixParentPrefix || hasWindowsParentPrefix

	// Check for embedded parent directory references within path
	hasUnixEmbeddedParent := strings.Contains(path, "/../")
	hasWindowsEmbeddedParent := strings.Contains(path, "\\..\\")
	hasEmbeddedParent := hasUnixEmbeddedParent || hasWindowsEmbeddedParent

	// Check for parent directory reference at end of path
	hasUnixParentSuffix := strings.HasSuffix(path, "/..")
	hasWindowsParentSuffix := strings.HasSuffix(path, "\\..")
	hasParentSuffix := hasUnixParentSuffix || hasWindowsParentSuffix

	// Check for direct parent directory reference
	isDirectParent := path == ".."

	// Return true if any traversal pattern is detected
	return hasParentPrefix || hasEmbeddedParent || hasParentSuffix || isDirectParent
}

// ContainsUnsafePath checks for dangerous path patterns including directory traversal,
// null bytes, and suspicious path separators. This is a comprehensive security check
// suitable for validating user-provided paths that may legitimately be outside the
// current working directory (such as log directories).
//
// SECURITY NOTE: This function checks for dangerous patterns in the original path
// BEFORE calling filepath.Clean(), because filepath.Clean() can normalize away
// attack patterns that we need to detect (e.g., "safe/./../../etc" becomes "../etc").
func ContainsUnsafePath(path string) bool {
	// Check for null bytes (can be used in path traversal attacks)
	if strings.Contains(path, "\x00") {
		return true
	}

	// Check for suspicious repeated separators
	if strings.Contains(path, "//") || strings.Contains(path, "\\\\") {
		return true
	}

	// CRITICAL SECURITY: Check for dangerous patterns in the ORIGINAL path
	// before filepath.Clean() can normalize them away
	return hasTraversalPatterns(path)
}
