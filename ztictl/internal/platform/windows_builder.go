package platform

import (
	"fmt"
	"strconv"
	"strings"
)

// WindowsBuilder implements CommandBuilder for Windows systems
type WindowsBuilder struct {
	BaseBuilder
}

// NewWindowsBuilder creates a new WindowsBuilder
func NewWindowsBuilder() *WindowsBuilder {
	return &WindowsBuilder{}
}

// GetSSMDocument returns the SSM document for Windows
func (b *WindowsBuilder) GetSSMDocument() string {
	return "AWS-RunPowerShellScript"
}

// BuildExecCommand wraps a command for execution with error handling
func (b *WindowsBuilder) BuildExecCommand(command string) string {
	// PowerShell command with exit code capture
	return fmt.Sprintf(`
$ErrorActionPreference = 'Continue'
try {
    %s
    $exitCode = $LASTEXITCODE
    if ($exitCode -eq $null) { $exitCode = 0 }
} catch {
    Write-Error $_.Exception.Message
    $exitCode = 1
}
Write-Output "EXIT_CODE:$exitCode"
exit $exitCode`, command)
}

// BuildFileExistsCommand creates a command to check if a file exists
func (b *WindowsBuilder) BuildFileExistsCommand(path string) string {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))
	return fmt.Sprintf(`if (Test-Path %s) { Write-Output 'EXISTS' } else { Write-Output 'NOT_EXISTS' }`, safePath)
}

// BuildFileSizeCommand creates a command to get the size of a file
func (b *WindowsBuilder) BuildFileSizeCommand(path string) string {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))
	return fmt.Sprintf(`(Get-Item %s -ErrorAction SilentlyContinue).Length`, safePath)
}

// BuildDirectoryCreateCommand creates a command to create a directory
func (b *WindowsBuilder) BuildDirectoryCreateCommand(path string) string {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))
	return fmt.Sprintf(`New-Item -ItemType Directory -Force -Path %s | Out-Null`, safePath)
}

// BuildFileReadCommand creates a command to read a file (base64 encoded)
func (b *WindowsBuilder) BuildFileReadCommand(path string) string {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))
	// Read file as bytes and convert to base64
	return fmt.Sprintf(`[Convert]::ToBase64String([System.IO.File]::ReadAllBytes(%s))`, safePath)
}

// BuildFileWriteCommand creates a command to write base64 data to a file
func (b *WindowsBuilder) BuildFileWriteCommand(path string, base64Data string) string {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))

	// For large data, split into chunks to avoid command line limits
	const chunkSize = 4096
	if len(base64Data) > chunkSize {
		// Use a here-string for large data
		return fmt.Sprintf(`
$base64 = @'
%s
'@
$bytes = [Convert]::FromBase64String($base64)
[System.IO.File]::WriteAllBytes(%s, $bytes)`, base64Data, safePath)
	}

	// For small data, use inline
	return fmt.Sprintf(`
$bytes = [Convert]::FromBase64String('%s')
[System.IO.File]::WriteAllBytes(%s, $bytes)`, base64Data, safePath)
}

// BuildFileAppendCommand creates a command to append base64 data to a file
func (b *WindowsBuilder) BuildFileAppendCommand(path string, base64Data string) string {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))

	// For large data, split into chunks to avoid command line limits
	const chunkSize = 4096
	if len(base64Data) > chunkSize {
		// Use a here-string for large data
		return fmt.Sprintf(`
$base64 = @'
%s
'@
$bytes = [Convert]::FromBase64String($base64)
$stream = [System.IO.File]::Open(%s, [System.IO.FileMode]::Append)
$stream.Write($bytes, 0, $bytes.Length)
$stream.Close()`, base64Data, safePath)
	}

	// For small data, use inline
	return fmt.Sprintf(`
$bytes = [Convert]::FromBase64String('%s')
$stream = [System.IO.File]::Open(%s, [System.IO.FileMode]::Append)
$stream.Write($bytes, 0, $bytes.Length)
$stream.Close()`, base64Data, safePath)
}

// NormalizePath converts a path to Windows format
func (b *WindowsBuilder) NormalizePath(path string) string {
	// Windows can use both forward and backslashes, but backslash is standard
	// PowerShell handles both, so we'll normalize to backslash
	normalized := strings.ReplaceAll(path, "/", "\\")

	// Handle UNC paths
	if strings.HasPrefix(normalized, "\\\\") {
		return normalized
	}

	// Handle drive letters
	if len(normalized) >= 2 && normalized[1] == ':' {
		return normalized
	}

	// If path doesn't start with drive letter or UNC, assume relative
	return normalized
}

// ParseExitCode extracts the exit code from command output
func (b *WindowsBuilder) ParseExitCode(output string) (int, error) {
	// Look for our EXIT_CODE marker
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
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
func (b *WindowsBuilder) ParseFileSize(output string) (int64, error) {
	// Clean the output
	sizeStr := strings.TrimSpace(output)

	// PowerShell might include newlines or carriage returns
	sizeStr = strings.ReplaceAll(sizeStr, "\r", "")
	lines := strings.Split(sizeStr, "\n")
	if len(lines) > 0 {
		sizeStr = strings.TrimSpace(lines[0])
	}

	// Parse the size
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file size '%s': %w", sizeStr, err)
	}

	return size, nil
}

// ParseFileExists interprets command output to determine if file exists
func (b *WindowsBuilder) ParseFileExists(output string, exitCode int) (bool, error) {
	output = strings.TrimSpace(output)
	output = strings.ReplaceAll(output, "\r", "")

	// Check for NOT_EXISTS first since it contains "EXISTS"
	if strings.Contains(output, "NOT_EXISTS") {
		return false, nil
	} else if strings.Contains(output, "EXISTS") {
		return true, nil
	}

	// Fallback to exit code
	if exitCode == 0 {
		return true, nil
	}

	return false, nil
}

// EscapePowerShellArg escapes a string for use as a PowerShell argument
func (b *WindowsBuilder) EscapePowerShellArg(arg string) string {
	// For PowerShell, we need to handle special characters
	// Single quotes are the safest for literal strings

	// If the argument contains single quotes, we need to escape them
	if strings.Contains(arg, "'") {
		// In PowerShell, double single quotes escape a single quote
		arg = strings.ReplaceAll(arg, "'", "''")
	}

	// Wrap in single quotes
	return fmt.Sprintf("'%s'", arg)
}
