package platform

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// CommandBuilder defines the interface for platform-specific command construction
type CommandBuilder interface {
	// GetSSMDocument returns the appropriate SSM document name for the platform
	GetSSMDocument() string

	// BuildExecCommand wraps a command for execution with error handling
	BuildExecCommand(command string) string

	// BuildFileExistsCommand creates a command to check if a file exists
	BuildFileExistsCommand(path string) string

	// BuildFileSizeCommand creates a command to get the size of a file
	BuildFileSizeCommand(path string) string

	// BuildDirectoryCreateCommand creates a command to create a directory
	BuildDirectoryCreateCommand(path string) string

	// BuildFileReadCommand creates a command to read a file (base64 encoded)
	BuildFileReadCommand(path string) string

	// BuildFileWriteCommand creates a command to write base64 data to a file
	// BREAKING CHANGE: v2.1.0 - Returns (string, error) instead of string to handle validation errors.
	// This is necessary for security validation of PowerShell here-strings on Windows.
	// All callers have been updated to handle the error return value.
	BuildFileWriteCommand(path string, base64Data string) (string, error)

	// BuildFileAppendCommand creates a command to append base64 data to a file
	// BREAKING CHANGE: v2.1.0 - Returns (string, error) instead of string to handle validation errors.
	// This is necessary for security validation of PowerShell here-strings on Windows.
	// All callers have been updated to handle the error return value.
	BuildFileAppendCommand(path string, base64Data string) (string, error)

	// NormalizePath converts a path to the platform's format with validation
	NormalizePath(path string) (string, error)

	// ParseExitCode extracts the exit code from command output
	ParseExitCode(output string) (int, error)

	// ParseFileSize extracts file size from command output
	ParseFileSize(output string) (int64, error)

	// ParseFileExists interprets command output to determine if file exists
	ParseFileExists(output string, exitCode int) (bool, error)
}

// BuilderFactory creates the appropriate CommandBuilder for a platform
type BuilderFactory struct{}

// NewBuilderFactory creates a new BuilderFactory
func NewBuilderFactory() *BuilderFactory {
	return &BuilderFactory{}
}

// GetBuilder returns the appropriate CommandBuilder for the given platform
func (f *BuilderFactory) GetBuilder(platform Platform) (CommandBuilder, error) {
	switch platform {
	case PlatformLinux:
		return NewLinuxBuilder(), nil
	case PlatformWindows:
		return NewWindowsBuilder(), nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

// CommandContext provides context for command execution
type CommandContext struct {
	Platform    Platform
	InstanceID  string
	Region      string
	WorkingDir  string
	Environment map[string]string
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	Output      string
	ErrorOutput string
	ExitCode    int
	Success     bool
}

// BaseBuilder provides common functionality for all platform builders
type BaseBuilder struct{}

// SanitizePath removes potentially dangerous characters from paths
func (b *BaseBuilder) SanitizePath(path string) string {
	// Remove null bytes and other dangerous characters
	path = strings.ReplaceAll(path, "\x00", "")
	path = strings.ReplaceAll(path, "\n", "")
	path = strings.ReplaceAll(path, "\r", "")

	// Normalize path separators
	path = filepath.Clean(path)

	// Note: Path validation for security (traversal, etc.) is handled
	// in the platform-specific NormalizePath methods which understand
	// platform-specific conventions like Windows UNC paths

	return path
}

// EscapeShellArg escapes a string for use as a shell argument
func (b *BaseBuilder) EscapeShellArg(arg string) string {
	// For Linux/Unix shells
	if strings.Contains(arg, "'") {
		// If the argument contains single quotes, use double quotes and escape
		arg = strings.ReplaceAll(arg, "\\", "\\\\")
		arg = strings.ReplaceAll(arg, "\"", "\\\"")
		arg = strings.ReplaceAll(arg, "$", "\\$")
		arg = strings.ReplaceAll(arg, "`", "\\`")
		return fmt.Sprintf("\"%s\"", arg)
	}
	// Otherwise, use single quotes for simplicity
	return fmt.Sprintf("'%s'", arg)
}

// BuilderManager manages command builders and platform detection
type BuilderManager struct {
	detector *Detector
	factory  *BuilderFactory
	builders map[string]CommandBuilder // Cache builders by instance ID
	mu       sync.RWMutex              // Protects builders map for concurrent access
}

// NewBuilderManager creates a new BuilderManager
func NewBuilderManager(detector *Detector) *BuilderManager {
	return &BuilderManager{
		detector: detector,
		factory:  NewBuilderFactory(),
		builders: make(map[string]CommandBuilder),
	}
}

// GetBuilder gets or creates a CommandBuilder for an instance
func (m *BuilderManager) GetBuilder(ctx context.Context, instanceID string) (CommandBuilder, error) {
	// Check cache with read lock
	m.mu.RLock()
	if builder, exists := m.builders[instanceID]; exists {
		m.mu.RUnlock()
		return builder, nil
	}
	m.mu.RUnlock()

	// Detect platform
	result, err := m.detector.DetectPlatform(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to detect platform: %w", err)
	}

	// Create builder
	builder, err := m.factory.GetBuilder(result.Platform)
	if err != nil {
		return nil, fmt.Errorf("failed to create builder: %w", err)
	}

	// Cache for future use with write lock
	m.mu.Lock()
	m.builders[instanceID] = builder
	m.mu.Unlock()

	return builder, nil
}

// ClearCache clears the builder cache
func (m *BuilderManager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.builders = make(map[string]CommandBuilder)
}
