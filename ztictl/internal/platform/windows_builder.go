package platform

import (
	"fmt"
	"strconv"
	"strings"
	"ztictl/pkg/security"
)

type WindowsBuilder struct {
	BaseBuilder
}

func NewWindowsBuilder() *WindowsBuilder {
	return &WindowsBuilder{}
}

func (b *WindowsBuilder) GetSSMDocument() string {
	return "AWS-RunPowerShellScript"
}

func (b *WindowsBuilder) BuildExecCommand(command string) string {
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

func (b *WindowsBuilder) BuildFileExistsCommand(path string) string {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))
	return fmt.Sprintf(`if (Test-Path %s) { Write-Output 'EXISTS' } else { Write-Output 'NOT_EXISTS' }`, safePath)
}

func (b *WindowsBuilder) BuildFileSizeCommand(path string) string {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))
	return fmt.Sprintf(`(Get-Item %s -ErrorAction SilentlyContinue).Length`, safePath)
}

func (b *WindowsBuilder) BuildDirectoryCreateCommand(path string) string {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))
	return fmt.Sprintf(`New-Item -ItemType Directory -Force -Path %s | Out-Null`, safePath)
}

func (b *WindowsBuilder) BuildFileReadCommand(path string) string {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))
	return fmt.Sprintf(`[Convert]::ToBase64String([System.IO.File]::ReadAllBytes(%s))`, safePath)
}

func (b *WindowsBuilder) validateBase64ForHereString(base64Data string) error {
	// PowerShell here-strings are terminated by "'@" on a new line.
	// This validation ensures that the base64 data does not contain the "'@" sequence,
	// which could prematurely terminate the here-string and break the script.
	if strings.Contains(base64Data, "'@") {
		return fmt.Errorf("base64 data contains invalid sequence \"'@\" which could break here-string")
	}
	return nil
}
func (b *WindowsBuilder) BuildFileWriteCommand(path string, base64Data string) (string, error) {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))

	if err := b.validateBase64ForHereString(base64Data); err != nil {
		return "", err
	}

	const chunkSize = 4096
	if len(base64Data) > chunkSize {
		return fmt.Sprintf(`
$base64 = @'
%s
'@
$bytes = [Convert]::FromBase64String($base64)
[System.IO.File]::WriteAllBytes(%s, $bytes)`, base64Data, safePath), nil
	}

	return fmt.Sprintf(`
$bytes = [Convert]::FromBase64String('%s')
[System.IO.File]::WriteAllBytes(%s, $bytes)`, base64Data, safePath), nil
}

func (b *WindowsBuilder) BuildFileAppendCommand(path string, base64Data string) (string, error) {
	safePath := b.EscapePowerShellArg(b.SanitizePath(path))

	if err := b.validateBase64ForHereString(base64Data); err != nil {
		return "", err
	}

	const chunkSize = 4096
	if len(base64Data) > chunkSize {
		return fmt.Sprintf(`
$base64 = @'
%s
'@
$bytes = [Convert]::FromBase64String($base64)
$stream = [System.IO.File]::Open(%s, [System.IO.FileMode]::Append)
$stream.Write($bytes, 0, $bytes.Length)
$stream.Close()`, base64Data, safePath), nil
	}

	return fmt.Sprintf(`
$bytes = [Convert]::FromBase64String('%s')
$stream = [System.IO.File]::Open(%s, [System.IO.FileMode]::Append)
$stream.Write($bytes, 0, $bytes.Length)
$stream.Close()`, base64Data, safePath), nil
}

func (b *WindowsBuilder) NormalizePath(path string) (string, error) {
	// Path validation occurs at multiple levels:
	// 1. SanitizePath() removes null bytes and control characters before normalization.
	// 2. security.ContainsUnsafePath() checks both original and normalized forms for unsafe patterns.
	// 3. UNC paths receive additional validation for server/share components.
	// The validation order is: sanitize, normalize, then validate.

	normalized := strings.ReplaceAll(path, "/", "\\")
	isUNC := strings.HasPrefix(normalized, "\\\\")
	if isUNC {
		// Validate UNC path structure
		if len(normalized) < 5 || !strings.Contains(normalized[2:], "\\") {
			return "", fmt.Errorf("invalid UNC path: %s", path)
		}

		parts := strings.SplitN(normalized[2:], "\\", 3)
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid UNC path format: %s", path)
		}

		server := parts[0]
		share := parts[1]

		// Server and share validation checks for:
		// Validate UNC server and share names according to Windows naming rules:
		// - Server and share names must not be empty.
		// - Server name must not start or end with a dot, or contain double dots (traversal).
		// - Server and share names must not contain invalid Windows filename characters.
		// - Additional security checks (e.g., command injection) are handled elsewhere.
		// This validation focuses on Windows naming rules.
		if server == "" || strings.HasPrefix(server, ".") || strings.HasSuffix(server, ".") ||
			strings.Contains(server, "..") || strings.ContainsAny(server, "/:<>|\"?*") {
			return "", fmt.Errorf("invalid UNC server name: %s", server)
		}

		if share == "" || strings.ContainsAny(share, "/:<>|\"?*") ||
			strings.Contains(share, "..") {
			return "", fmt.Errorf("invalid UNC share name: %s", share)
		}

		if len(parts) >= 3 && parts[2] != "" {
			if security.ContainsUnsafePath(parts[2]) {
				return "", fmt.Errorf("unsafe path in UNC: %s", path)
			}
		}

		return normalized, nil
	}

	if security.ContainsUnsafePath(normalized) {
		return "", fmt.Errorf("unsafe path detected: %s", path)
	}

	if len(normalized) >= 2 && normalized[1] == ':' {
		drive := normalized[0]
		if !((drive >= 'A' && drive <= 'Z') || (drive >= 'a' && drive <= 'z')) {
			return "", fmt.Errorf("invalid drive letter: %s", string(drive))
		}
		return normalized, nil
	}

	return normalized, nil
}

func (b *WindowsBuilder) ParseExitCode(output string) (int, error) {
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

	if output != "" {
		return 0, nil
	}

	return -1, fmt.Errorf("no exit code found in output")
}

func (b *WindowsBuilder) ParseFileSize(output string) (int64, error) {
	sizeStr := strings.TrimSpace(output)

	sizeStr = strings.ReplaceAll(sizeStr, "\r", "")
	lines := strings.Split(sizeStr, "\n")
	if len(lines) > 0 {
		sizeStr = strings.TrimSpace(lines[0])
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file size '%s': %w", sizeStr, err)
	}

	return size, nil
}

func (b *WindowsBuilder) ParseFileExists(output string, exitCode int) (bool, error) {
	output = strings.TrimSpace(output)
	output = strings.ReplaceAll(output, "\r", "")

	if strings.Contains(output, "NOT_EXISTS") {
		return false, nil
	} else if strings.Contains(output, "EXISTS") {
		return true, nil
	}

	if exitCode == 0 {
		return true, nil
	}

	return false, nil
}

func (b *WindowsBuilder) EscapePowerShellArg(arg string) string {
	if strings.Contains(arg, "'") {
		arg = strings.ReplaceAll(arg, "'", "''")
	}

	return fmt.Sprintf("'%s'", arg)
}
