package logging

import (
	"strings"
	"testing"
)

// TestLoggerFormatFields tests that the formatFields method properly formats structured logging fields
func TestLoggerFormatFields(t *testing.T) {
	logger := NewLogger(false)

	tests := []struct {
		name     string
		fields   []interface{}
		expected string
	}{
		{
			name:     "no fields",
			fields:   []interface{}{},
			expected: "",
		},
		{
			name:     "single key-value pair",
			fields:   []interface{}{"key", "value"},
			expected: " - key=value",
		},
		{
			name:     "multiple key-value pairs",
			fields:   []interface{}{"key1", "value1", "key2", "value2"},
			expected: " - key1=value1 key2=value2",
		},
		{
			name:     "odd number of fields",
			fields:   []interface{}{"key1", "value1", "key2"},
			expected: " - key1=value1 key2=<no_value>",
		},
		{
			name:     "file path value - no pipe character",
			fields:   []interface{}{"file", "C:\\Users\\ditah\\.ztictl.yaml"},
			expected: " - file=C:\\Users\\ditah\\.ztictl.yaml",
		},
		{
			name:     "multiple fields including file path",
			fields:   []interface{}{"action", "loading", "file", "/home/user/.config"},
			expected: " - action=loading file=/home/user/.config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.formatFields(tt.fields...)
			if result != tt.expected {
				t.Errorf("formatFields() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestLoggerNoOpMode tests that NoOp logger is properly configured
func TestLoggerNoOpMode(t *testing.T) {
	logger := NewNoOpLogger()

	// Check that the logger is in no-op mode
	if !logger.noOp {
		t.Error("NoOp logger should have noOp flag set to true")
	}

	// These calls should not panic (they just return early)
	logger.Info("test message", "key", "value")
	logger.Debug("debug message", "key", "value")
	logger.Warn("warning message", "key", "value")
	logger.Error("error message", "key", "value")
}

// TestLoggerDebugFlag tests that debug messages are only shown when debug is enabled
func TestLoggerDebugFlag(t *testing.T) {
	// Test with debug disabled - debug messages should not appear
	logger := NewLogger(false)

	// The logger.Debug method checks the debug flag internally,
	// so when debug is false, it returns early without output
	// We can verify this by checking the logger's debugEnabled field
	if logger.debugEnabled {
		t.Error("Logger should have debug disabled")
	}

	// Test with debug enabled
	logger = NewLogger(true)

	if !logger.debugEnabled {
		t.Error("Logger should have debug enabled")
	}
}

// TestLoggerInfoMessage tests that Info method formats fields correctly
func TestLoggerInfoMessage(t *testing.T) {
	logger := NewLogger(false)

	// Test that formatFields produces the expected format
	formatted := logger.formatFields("status", "ok", "count", 42)

	// Check the formatted string
	if !strings.Contains(formatted, "status=ok") {
		t.Errorf("Formatted string missing status=ok: %s", formatted)
	}
	if !strings.Contains(formatted, "count=42") {
		t.Errorf("Formatted string missing count=42: %s", formatted)
	}
	// Ensure we're using dash separator, not pipe
	if strings.Contains(formatted, " | ") {
		t.Errorf("Formatted string contains pipe separator instead of dash: %s", formatted)
	}
	if !strings.Contains(formatted, " - ") {
		t.Errorf("Formatted string missing dash separator: %s", formatted)
	}
}

// TestLoggerSetLevel tests that SetLevel is a no-op (for compatibility)
func TestLoggerSetLevel(t *testing.T) {
	logger := NewLogger(false)

	// This should not panic or cause issues
	logger.SetLevel(DebugLevel)
	logger.SetLevel(InfoLevel)
	logger.SetLevel(WarnLevel)
	logger.SetLevel(ErrorLevel)

	// The logger should still be functional after SetLevel calls
	// We can verify by checking that it's not nil and not in no-op mode
	if logger == nil {
		t.Error("Logger is nil after SetLevel calls")
	}
	if logger.noOp {
		t.Error("Logger is in no-op mode after SetLevel calls")
	}
}

// TestFormatFieldsNoPipeCharacter ensures no pipe characters in output for PowerShell compatibility
func TestFormatFieldsNoPipeCharacter(t *testing.T) {
	logger := NewLogger(false)

	// Test various field combinations to ensure no pipe character
	testCases := [][]interface{}{
		{"file", "C:\\Program Files\\app.exe"},
		{"path", "/usr/local/bin", "user", "admin"},
		{"url", "https://example.com", "status", "200"},
	}

	for _, fields := range testCases {
		result := logger.formatFields(fields...)
		if strings.Contains(result, "|") {
			t.Errorf("formatFields contains pipe character: %q with fields %v", result, fields)
		}
	}
}
