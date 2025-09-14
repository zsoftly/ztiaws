package platform

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinuxBuilder_GetSSMDocument(t *testing.T) {
	builder := NewLinuxBuilder()
	assert.Equal(t, "AWS-RunShellScript", builder.GetSSMDocument())
}

func TestLinuxBuilder_BuildExecCommand(t *testing.T) {
	builder := NewLinuxBuilder()

	tests := []struct {
		name     string
		command  string
		contains []string
	}{
		{
			name:    "Simple command",
			command: "echo hello",
			contains: []string{
				"echo hello",
				"EXIT_CODE=$?",
				"exit $EXIT_CODE",
			},
		},
		{
			name:    "Complex command with pipes",
			command: "ps aux | grep nginx | awk '{print $2}'",
			contains: []string{
				"ps aux | grep nginx | awk '{print $2}'",
				"EXIT_CODE=$?",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildExecCommand(tt.command)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestLinuxBuilder_BuildFileExistsCommand(t *testing.T) {
	builder := NewLinuxBuilder()

	tests := []struct {
		name     string
		path     string
		contains string
	}{
		{
			name:     "Simple path",
			path:     "/tmp/test.txt",
			contains: "test -e '/tmp/test.txt'",
		},
		{
			name:     "Path with spaces",
			path:     "/tmp/my file.txt",
			contains: "test -e '/tmp/my file.txt'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildFileExistsCommand(tt.path)
			assert.Contains(t, result, tt.contains)
			assert.Contains(t, result, "EXISTS")
			assert.Contains(t, result, "NOT_EXISTS")
		})
	}
}

func TestLinuxBuilder_BuildFileSizeCommand(t *testing.T) {
	builder := NewLinuxBuilder()

	tests := []struct {
		name string
		path string
	}{
		{
			name: "Simple path",
			path: "/tmp/test.txt",
		},
		{
			name: "Path with special characters",
			path: "/tmp/test file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildFileSizeCommand(tt.path)
			assert.Contains(t, result, "stat")
			assert.Contains(t, result, "%s") // Linux format
			assert.Contains(t, result, "%z") // macOS format
		})
	}
}

func TestLinuxBuilder_BuildDirectoryCreateCommand(t *testing.T) {
	builder := NewLinuxBuilder()

	tests := []struct {
		name     string
		path     string
		contains string
	}{
		{
			name:     "Simple directory",
			path:     "/tmp/mydir",
			contains: "mkdir -p '/tmp/mydir'",
		},
		{
			name:     "Nested directory",
			path:     "/tmp/a/b/c",
			contains: "mkdir -p '/tmp/a/b/c'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildDirectoryCreateCommand(tt.path)
			assert.Equal(t, tt.contains, result)
		})
	}
}

func TestLinuxBuilder_BuildFileReadCommand(t *testing.T) {
	builder := NewLinuxBuilder()

	result := builder.BuildFileReadCommand("/tmp/test.txt")
	assert.Contains(t, result, "base64")
	assert.Contains(t, result, "/tmp/test.txt")
	// Should try without line wrapping first
	assert.Contains(t, result, "-w 0")
}

func TestLinuxBuilder_BuildFileWriteCommand(t *testing.T) {
	builder := NewLinuxBuilder()

	tests := []struct {
		name       string
		path       string
		base64Data string
		checkFor   []string
	}{
		{
			name:       "Small data",
			path:       "/tmp/test.txt",
			base64Data: "SGVsbG8gV29ybGQ=", // "Hello World"
			checkFor: []string{
				"echo 'SGVsbG8gV29ybGQ='",
				"base64 -d",
				"> '/tmp/test.txt'",
			},
		},
		{
			name:       "Large data",
			path:       "/tmp/test.txt",
			base64Data: strings.Repeat("A", 5000),
			checkFor: []string{
				"cat << 'EOF_BASE64'",
				"base64 -d",
				"> '/tmp/test.txt'",
				"EOF_BASE64",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildFileWriteCommand(tt.path, tt.base64Data)
			for _, check := range tt.checkFor {
				assert.Contains(t, result, check)
			}
		})
	}
}

func TestLinuxBuilder_NormalizePath(t *testing.T) {
	builder := NewLinuxBuilder()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Unix path unchanged",
			input:    "/home/user/file.txt",
			expected: "/home/user/file.txt",
		},
		{
			name:     "Windows path converted",
			input:    "C:\\Users\\file.txt",
			expected: "C:/Users/file.txt",
		},
		{
			name:     "Mixed separators",
			input:    "/home\\user/file.txt",
			expected: "/home/user/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.NormalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLinuxBuilder_ParseExitCode(t *testing.T) {
	builder := NewLinuxBuilder()

	tests := []struct {
		name         string
		output       string
		expectedCode int
		expectError  bool
	}{
		{
			name:         "Successful command",
			output:       "Command output\nEXIT_CODE:0\n",
			expectedCode: 0,
			expectError:  false,
		},
		{
			name:         "Failed command",
			output:       "Error message\nEXIT_CODE:1\n",
			expectedCode: 1,
			expectError:  false,
		},
		{
			name:         "Exit code 127",
			output:       "Command not found\nEXIT_CODE:127\n",
			expectedCode: 127,
			expectError:  false,
		},
		{
			name:         "No exit code with output",
			output:       "Some output",
			expectedCode: 0,
			expectError:  false,
		},
		{
			name:         "No exit code no output",
			output:       "",
			expectedCode: -1,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := builder.ParseExitCode(tt.output)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}

func TestLinuxBuilder_ParseFileSize(t *testing.T) {
	builder := NewLinuxBuilder()

	tests := []struct {
		name         string
		output       string
		expectedSize int64
		expectError  bool
	}{
		{
			name:         "Simple size",
			output:       "1024\n",
			expectedSize: 1024,
			expectError:  false,
		},
		{
			name:         "Large file",
			output:       "1073741824",
			expectedSize: 1073741824,
			expectError:  false,
		},
		{
			name:         "Size with extra output",
			output:       "2048\nsome error message",
			expectedSize: 2048,
			expectError:  false,
		},
		{
			name:         "Invalid size",
			output:       "not a number",
			expectedSize: 0,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := builder.ParseFileSize(tt.output)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedSize, size)
		})
	}
}

func TestLinuxBuilder_ParseFileExists(t *testing.T) {
	builder := NewLinuxBuilder()

	tests := []struct {
		name        string
		output      string
		exitCode    int
		expected    bool
		expectError bool
	}{
		{
			name:        "File exists",
			output:      "EXISTS",
			exitCode:    0,
			expected:    true,
			expectError: false,
		},
		{
			name:        "File does not exist",
			output:      "NOT_EXISTS",
			exitCode:    1,
			expected:    false,
			expectError: false,
		},
		{
			name:        "Exit code 0 fallback",
			output:      "",
			exitCode:    0,
			expected:    true,
			expectError: false,
		},
		{
			name:        "Exit code 1 fallback",
			output:      "",
			exitCode:    1,
			expected:    false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := builder.ParseFileExists(tt.output, tt.exitCode)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, exists)
		})
	}
}

func TestLinuxBuilder_EscapeShellArg(t *testing.T) {
	builder := NewLinuxBuilder()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple string",
			input:    "hello",
			expected: "'hello'",
		},
		{
			name:     "String with spaces",
			input:    "hello world",
			expected: "'hello world'",
		},
		{
			name:     "String with single quote",
			input:    "it's",
			expected: `"it's"`,
		},
		{
			name:     "String with dollar sign and single quote",
			input:    "$test's",
			expected: `"\$test's"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.EscapeShellArg(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
