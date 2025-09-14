package platform

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"ztictl/pkg/security"
)

type LinuxBuilder struct {
	BaseBuilder
}

func NewLinuxBuilder() *LinuxBuilder {
	return &LinuxBuilder{}
}

func (b *LinuxBuilder) GetSSMDocument() string {
	return "AWS-RunShellScript"
}

func (b *LinuxBuilder) BuildExecCommand(command string) string {
	return fmt.Sprintf(`
set -e
%s
EXIT_CODE=$?
echo "EXIT_CODE:$EXIT_CODE"
exit $EXIT_CODE`, command)
}

func (b *LinuxBuilder) BuildFileExistsCommand(path string) string {
	sanitized := b.SanitizePath(path)
	// Ensure Unix-style paths regardless of host OS
	sanitized = strings.ReplaceAll(sanitized, "\\", "/")
	safePath := b.EscapeShellArg(sanitized)
	return fmt.Sprintf("test -e %s && echo 'EXISTS' || echo 'NOT_EXISTS'", safePath)
}

func (b *LinuxBuilder) BuildFileSizeCommand(path string) string {
	sanitized := b.SanitizePath(path)
	// Ensure Unix-style paths regardless of host OS
	sanitized = strings.ReplaceAll(sanitized, "\\", "/")
	safePath := b.EscapeShellArg(sanitized)
	return fmt.Sprintf("stat -c '%%s' %s 2>/dev/null || stat -f '%%z' %s 2>/dev/null", safePath, safePath)
}

func (b *LinuxBuilder) BuildDirectoryCreateCommand(path string) string {
	sanitized := b.SanitizePath(path)
	// Ensure Unix-style paths regardless of host OS
	sanitized = strings.ReplaceAll(sanitized, "\\", "/")
	safePath := b.EscapeShellArg(sanitized)
	return fmt.Sprintf("mkdir -p %s", safePath)
}

func (b *LinuxBuilder) BuildFileReadCommand(path string) string {
	sanitized := b.SanitizePath(path)
	// Ensure Unix-style paths regardless of host OS
	sanitized = strings.ReplaceAll(sanitized, "\\", "/")
	safePath := b.EscapeShellArg(sanitized)
	return fmt.Sprintf("base64 -w 0 %s 2>/dev/null || base64 %s", safePath, safePath)
}

func (b *LinuxBuilder) BuildFileWriteCommand(path string, base64Data string) (string, error) {
	sanitized := b.SanitizePath(path)
	// Ensure Unix-style paths regardless of host OS
	sanitized = strings.ReplaceAll(sanitized, "\\", "/")
	safePath := b.EscapeShellArg(sanitized)

	// Get directory using Unix-style path operations
	dirSanitized := b.SanitizePath(filepath.Dir(path))
	dirSanitized = strings.ReplaceAll(dirSanitized, "\\", "/")
	dirPath := b.EscapeShellArg(dirSanitized)

	const chunkSize = 4096
	if len(base64Data) > chunkSize {
		return fmt.Sprintf(`
mkdir -p %s
cat << 'EOF_BASE64' | base64 -d > %s
%s
EOF_BASE64`, dirPath, safePath, base64Data), nil
	}

	return fmt.Sprintf(`
mkdir -p %s
echo '%s' | base64 -d > %s`, dirPath, base64Data, safePath), nil
}

func (b *LinuxBuilder) BuildFileAppendCommand(path string, base64Data string) (string, error) {
	sanitized := b.SanitizePath(path)
	// Ensure Unix-style paths regardless of host OS
	sanitized = strings.ReplaceAll(sanitized, "\\", "/")
	safePath := b.EscapeShellArg(sanitized)

	// Get directory using Unix-style path operations
	dirSanitized := b.SanitizePath(filepath.Dir(path))
	dirSanitized = strings.ReplaceAll(dirSanitized, "\\", "/")
	dirPath := b.EscapeShellArg(dirSanitized)

	const chunkSize = 4096
	if len(base64Data) > chunkSize {
		return fmt.Sprintf(`
mkdir -p %s
cat << 'EOF_BASE64' | base64 -d >> %s
%s
EOF_BASE64`, dirPath, safePath, base64Data), nil
	}

	return fmt.Sprintf(`
mkdir -p %s
echo '%s' | base64 -d >> %s`, dirPath, base64Data, safePath), nil
}

func (b *LinuxBuilder) NormalizePath(path string) (string, error) {
	normalized := strings.ReplaceAll(path, "\\", "/")

	if security.ContainsUnsafePath(normalized) {
		return "", fmt.Errorf("unsafe path detected: %s", path)
	}

	return normalized, nil
}

func (b *LinuxBuilder) ParseExitCode(output string) (int, error) {
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

	if output != "" {
		return 0, nil
	}

	return -1, fmt.Errorf("no exit code found in output")
}

func (b *LinuxBuilder) ParseFileSize(output string) (int64, error) {
	sizeStr := strings.TrimSpace(output)

	lines := strings.Split(sizeStr, "\n")
	if len(lines) > 0 {
		sizeStr = lines[0]
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file size '%s': %w", sizeStr, err)
	}

	return size, nil
}

func (b *LinuxBuilder) ParseFileExists(output string, exitCode int) (bool, error) {
	output = strings.TrimSpace(output)

	switch output {
	case "EXISTS":
		return true, nil
	case "NOT_EXISTS":
		return false, nil
	default:
		if exitCode == 0 {
			return true, nil
		}
		return false, nil
	}
}
