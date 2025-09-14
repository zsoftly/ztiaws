package platform

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWindowsBuilder_GetSSMDocument(t *testing.T) {
	builder := NewWindowsBuilder()
	assert.Equal(t, "AWS-RunPowerShellScript", builder.GetSSMDocument())
}

func TestWindowsBuilder_BuildExecCommand(t *testing.T) {
	builder := NewWindowsBuilder()

	tests := []struct {
		name     string
		command  string
		contains []string
	}{
		{
			name:    "Simple command",
			command: "Get-Process",
			contains: []string{
				"Get-Process",
				"$LASTEXITCODE",
				"Write-Output \"EXIT_CODE:$exitCode\"",
			},
		},
		{
			name:    "Command with pipes",
			command: "Get-Service | Where-Object {$_.Status -eq 'Running'}",
			contains: []string{
				"Get-Service | Where-Object {$_.Status -eq 'Running'}",
				"$LASTEXITCODE",
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

func TestWindowsBuilder_BuildFileExistsCommand(t *testing.T) {
	builder := NewWindowsBuilder()

	tests := []struct {
		name     string
		path     string
		contains string
	}{
		{
			name:     "Simple path",
			path:     "C:\\temp\\test.txt",
			contains: "Test-Path 'C:\\temp\\test.txt'",
		},
		{
			name:     "Path with spaces",
			path:     "C:\\Program Files\\test.txt",
			contains: "Test-Path 'C:\\Program Files\\test.txt'",
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

func TestWindowsBuilder_BuildFileSizeCommand(t *testing.T) {
	builder := NewWindowsBuilder()

	tests := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "Simple path",
			path: "C:\\temp\\test.txt",
			contains: []string{
				"Get-Item",
				"C:\\temp\\test.txt",
				".Length",
			},
		},
		{
			name: "UNC path",
			path: "\\\\server\\share\\file.txt",
			contains: []string{
				"Get-Item",
				"\\\\server\\share\\file.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildFileSizeCommand(tt.path)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestWindowsBuilder_BuildDirectoryCreateCommand(t *testing.T) {
	builder := NewWindowsBuilder()

	tests := []struct {
		name     string
		path     string
		contains []string
	}{
		{
			name: "Simple directory",
			path: "C:\\temp\\mydir",
			contains: []string{
				"New-Item",
				"-ItemType Directory",
				"-Force",
				"C:\\temp\\mydir",
			},
		},
		{
			name: "Nested directory",
			path: "C:\\temp\\a\\b\\c",
			contains: []string{
				"New-Item",
				"C:\\temp\\a\\b\\c",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildDirectoryCreateCommand(tt.path)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestWindowsBuilder_BuildFileReadCommand(t *testing.T) {
	builder := NewWindowsBuilder()

	result := builder.BuildFileReadCommand("C:\\temp\\test.txt")
	assert.Contains(t, result, "[Convert]::ToBase64String")
	assert.Contains(t, result, "[System.IO.File]::ReadAllBytes")
	assert.Contains(t, result, "C:\\temp\\test.txt")
}

func TestWindowsBuilder_BuildFileWriteCommand(t *testing.T) {
	builder := NewWindowsBuilder()

	tests := []struct {
		name        string
		path        string
		base64Data  string
		checkFor    []string
		expectError bool
	}{
		{
			name:       "Small data",
			path:       "C:\\temp\\test.txt",
			base64Data: "SGVsbG8gV29ybGQ=", // "Hello World"
			checkFor: []string{
				"[Convert]::FromBase64String('SGVsbG8gV29ybGQ=')",
				"[System.IO.File]::WriteAllBytes",
				"C:\\temp\\test.txt",
			},
			expectError: false,
		},
		{
			name:       "Large data",
			path:       "C:\\temp\\test.txt",
			base64Data: strings.Repeat("A", 5000),
			checkFor: []string{
				"$base64 = @'",
				"'@",
				"[Convert]::FromBase64String($base64)",
				"[System.IO.File]::WriteAllBytes",
			},
			expectError: false,
		},
		{
			name:        "Invalid here-string terminator",
			path:        "C:\\temp\\test.txt",
			base64Data:  "SGVsbG8'@V29ybGQ=",
			checkFor:    []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builder.BuildFileWriteCommand(tt.path, tt.base64Data)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "here-string")
			} else {
				assert.NoError(t, err)
				for _, check := range tt.checkFor {
					assert.Contains(t, result, check)
				}
			}
		})
	}
}

func TestWindowsBuilder_BuildFileAppendCommand(t *testing.T) {
	builder := NewWindowsBuilder()

	tests := []struct {
		name        string
		path        string
		base64Data  string
		checkFor    []string
		expectError bool
	}{
		{
			name:       "Small data append",
			path:       "C:\\temp\\test.txt",
			base64Data: "SGVsbG8gV29ybGQ=", // "Hello World"
			checkFor: []string{
				"[Convert]::FromBase64String('SGVsbG8gV29ybGQ=')",
				"[System.IO.File]::Open",
				"[System.IO.FileMode]::Append",
				"C:\\temp\\test.txt",
			},
			expectError: false,
		},
		{
			name:       "Large data append",
			path:       "C:\\temp\\test.txt",
			base64Data: strings.Repeat("B", 5000),
			checkFor: []string{
				"$base64 = @'",
				"'@",
				"[Convert]::FromBase64String($base64)",
				"[System.IO.File]::Open",
				"[System.IO.FileMode]::Append",
			},
			expectError: false,
		},
		{
			name:        "Invalid here-string terminator in append",
			path:        "C:\\temp\\test.txt",
			base64Data:  "SGVsbG8'@V29ybGQ=",
			checkFor:    []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builder.BuildFileAppendCommand(tt.path, tt.base64Data)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "here-string")
			} else {
				assert.NoError(t, err)
				for _, check := range tt.checkFor {
					assert.Contains(t, result, check)
				}
			}
		})
	}
}

func TestWindowsBuilder_NormalizePath(t *testing.T) {
	builder := NewWindowsBuilder()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "Windows path unchanged",
			input:       "C:\\Users\\file.txt",
			expected:    "C:\\Users\\file.txt",
			expectError: false,
		},
		{
			name:        "Unix path converted",
			input:       "C:/Users/file.txt",
			expected:    "C:\\Users\\file.txt",
			expectError: false,
		},
		{
			name:        "UNC path valid",
			input:       "\\\\server\\share\\file.txt",
			expected:    "\\\\server\\share\\file.txt",
			expectError: false,
		},
		{
			name:        "Invalid UNC path - too short",
			input:       `\\s`,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid UNC path - no share",
			input:       `\\server`,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid UNC server - dots",
			input:       `\\..server\share`,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid UNC share - special chars",
			input:       `\\server\share:bad`,
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid drive letter",
			input:       "9:\\Users\\file.txt",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Path traversal attempt",
			input:       "C:\\\\Users\\\\..\\\\..\\\\Windows",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Mixed separators",
			input:       "C:/Users\\Documents/file.txt",
			expected:    "C:\\Users\\Documents\\file.txt",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builder.NormalizePath(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestWindowsBuilder_ParseExitCode(t *testing.T) {
	builder := NewWindowsBuilder()

	tests := []struct {
		name         string
		output       string
		expectedCode int
		expectError  bool
	}{
		{
			name:         "Successful command",
			output:       "Command output\r\nEXIT_CODE:0\r\n",
			expectedCode: 0,
			expectError:  false,
		},
		{
			name:         "Failed command",
			output:       "Error message\r\nEXIT_CODE:1\r\n",
			expectedCode: 1,
			expectError:  false,
		},
		{
			name:         "Exit code with spaces",
			output:       "Some output\r\n  EXIT_CODE:127  \r\n",
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

func TestWindowsBuilder_ParseFileSize(t *testing.T) {
	builder := NewWindowsBuilder()

	tests := []struct {
		name         string
		output       string
		expectedSize int64
		expectError  bool
	}{
		{
			name:         "Simple size",
			output:       "1024\r\n",
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
			output:       "2048\r\nsome error message",
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

func TestWindowsBuilder_ParseFileExists(t *testing.T) {
	builder := NewWindowsBuilder()

	tests := []struct {
		name        string
		output      string
		exitCode    int
		expected    bool
		expectError bool
	}{
		{
			name:        "File exists",
			output:      "EXISTS\r\n",
			exitCode:    0,
			expected:    true,
			expectError: false,
		},
		{
			name:        "File does not exist",
			output:      "NOT_EXISTS\r\n",
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

func TestWindowsBuilder_EscapePowerShellArg(t *testing.T) {
	builder := NewWindowsBuilder()

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
			expected: "'it''s'",
		},
		{
			name:     "String with multiple quotes",
			input:    "can't won't",
			expected: "'can''t won''t'",
		},
		{
			name:     "Path with backslashes",
			input:    "C:\\Program Files\\app.exe",
			expected: "'C:\\Program Files\\app.exe'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.EscapePowerShellArg(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
