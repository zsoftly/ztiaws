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

	// Check for parent directory reference at start of path
	hasUnixParentPrefix := strings.HasPrefix(relPath, "../")
	hasWindowsParentPrefix := strings.HasPrefix(relPath, "..\\")
	hasParentPrefix := hasUnixParentPrefix || hasWindowsParentPrefix

	// Check for embedded parent directory references within path
	hasUnixEmbeddedParent := strings.Contains(relPath, "/../")
	hasWindowsEmbeddedParent := strings.Contains(relPath, "\\..\\")
	hasEmbeddedParent := hasUnixEmbeddedParent || hasWindowsEmbeddedParent

	// Check for parent directory reference at end of path
	hasUnixParentSuffix := strings.HasSuffix(relPath, "/..")
	hasWindowsParentSuffix := strings.HasSuffix(relPath, "\\..")
	hasParentSuffix := hasUnixParentSuffix || hasWindowsParentSuffix

	// Check for direct parent directory reference
	isDirectParent := relPath == ".."

	// Return true if any traversal pattern is detected
	return isAbsolutePath || hasParentPrefix || hasEmbeddedParent || hasParentSuffix || isDirectParent
}
