package platform

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// LinuxBuilder implements CommandBuilder for Linux/Unix systems
type LinuxBuilder struct {
	BaseBuilder
}

// NewLinuxBuilder creates a new LinuxBuilder
func NewLinuxBuilder() *LinuxBuilder {
	return &LinuxBuilder{}
}

// GetSSMDocument returns the SSM document for Linux
func (b *LinuxBuilder) GetSSMDocument() string {
	return "AWS-RunShellScript"
}

// BuildExecCommand wraps a command for execution with error handling
func (b *LinuxBuilder) BuildExecCommand(command string) string {
	// Execute command and capture exit code
	return fmt.Sprintf(`
set -e
%s
EXIT_CODE=$?
echo "EXIT_CODE:$EXIT_CODE"
exit $EXIT_CODE`, command)
}

// BuildFileExistsCommand creates a command to check if a file exists
func (b *LinuxBuilder) BuildFileExistsCommand(path string) string {
	safePath := b.EscapeShellArg(b.SanitizePath(path))
	return fmt.Sprintf("test -e %s && echo 'EXISTS' || echo 'NOT_EXISTS'", safePath)
}

// BuildFileSizeCommand creates a command to get the size of a file
func (b *LinuxBuilder) BuildFileSizeCommand(path string) string {
	safePath := b.EscapeShellArg(b.SanitizePath(path))
	// Use stat with portable format across different Unix systems
	return fmt.Sprintf("stat -c '%%s' %s 2>/dev/null || stat -f '%%z' %s 2>/dev/null", safePath, safePath)
}

// BuildDirectoryCreateCommand creates a command to create a directory
func (b *LinuxBuilder) BuildDirectoryCreateCommand(path string) string {
	safePath := b.EscapeShellArg(b.SanitizePath(path))
	return fmt.Sprintf("mkdir -p %s", safePath)
}

// BuildFileReadCommand creates a command to read a file (base64 encoded)
func (b *LinuxBuilder) BuildFileReadCommand(path string) string {
	safePath := b.EscapeShellArg(b.SanitizePath(path))
	// Use base64 without line wrapping for easier parsing
	return fmt.Sprintf("base64 -w 0 %s 2>/dev/null || base64 %s", safePath, safePath)
}

// BuildFileWriteCommand creates a command to write base64 data to a file
func (b *LinuxBuilder) BuildFileWriteCommand(path string, base64Data string) string {
	safePath := b.EscapeShellArg(b.SanitizePath(path))

	// Create parent directory if it doesn't exist
	dirPath := b.EscapeShellArg(b.SanitizePath(strings.TrimSuffix(path, "/"+filepath.Base(path))))

	// Split base64 data into chunks to avoid command line length limits
	const chunkSize = 4096
	if len(base64Data) > chunkSize {
		// For large data, use a here-document
		return fmt.Sprintf(`
mkdir -p %s
cat << 'EOF_BASE64' | base64 -d > %s
%s
EOF_BASE64`, dirPath, safePath, base64Data)
	}

	// For small data, use a simple echo
	return fmt.Sprintf(`
mkdir -p %s
echo '%s' | base64 -d > %s`, dirPath, base64Data, safePath)
}

// BuildFileAppendCommand creates a command to append base64 data to a file
func (b *LinuxBuilder) BuildFileAppendCommand(path string, base64Data string) string {
	safePath := b.EscapeShellArg(b.SanitizePath(path))

	// Create parent directory if it doesn't exist
	dirPath := b.EscapeShellArg(b.SanitizePath(strings.TrimSuffix(path, "/"+filepath.Base(path))))

	// Split base64 data into chunks to avoid command line length limits
	const chunkSize = 4096
	if len(base64Data) > chunkSize {
		// For large data, use a here-document
		return fmt.Sprintf(`
mkdir -p %s
cat << 'EOF_BASE64' | base64 -d >> %s
%s
EOF_BASE64`, dirPath, safePath, base64Data)
	}

	// For small data, use a simple echo
	return fmt.Sprintf(`
mkdir -p %s
echo '%s' | base64 -d >> %s`, dirPath, base64Data, safePath)
}

// NormalizePath converts a path to Unix format
func (b *LinuxBuilder) NormalizePath(path string) string {
	// Linux uses forward slashes
	return strings.ReplaceAll(path, "\\", "/")
}

// ParseExitCode extracts the exit code from command output
func (b *LinuxBuilder) ParseExitCode(output string) (int, error) {
	// Look for our EXIT_CODE marker
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "EXIT_CODE:") {
			codeStr := strings.TrimPrefix(line, "EXIT_CODE:")
			code, err := strconv.Atoi(strings.TrimSpace(codeStr))
			if err != nil {
				return -1, fmt.Errorf("failed to parse exit code: %w", err)
			}
			return code, nil
		}
	}

	// If no explicit exit code found, assume success if we got output
	if output != "" {
		return 0, nil
	}

	return -1, fmt.Errorf("no exit code found in output")
}

// ParseFileSize extracts file size from command output
func (b *LinuxBuilder) ParseFileSize(output string) (int64, error) {
	// Clean the output
	sizeStr := strings.TrimSpace(output)

	// Remove any error messages
	lines := strings.Split(sizeStr, "\n")
	if len(lines) > 0 {
		sizeStr = lines[0]
	}

	// Parse the size
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file size '%s': %w", sizeStr, err)
	}

	return size, nil
}

// ParseFileExists interprets command output to determine if file exists
func (b *LinuxBuilder) ParseFileExists(output string, exitCode int) (bool, error) {
	output = strings.TrimSpace(output)

	switch output {
	case "EXISTS":
		return true, nil
	case "NOT_EXISTS":
		return false, nil
	default:
		// Fallback to exit code
		if exitCode == 0 {
			return true, nil
		}
		return false, nil
	}
}
