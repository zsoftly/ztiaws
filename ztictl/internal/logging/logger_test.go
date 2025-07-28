package logging

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	// Test that creating a new logger doesn't panic
	logger := NewLogger(false)
	if logger == nil {
		t.Error("NewLogger returned nil")
	}
}

func TestNewLoggerWithDebug(t *testing.T) {
	// Test debug mode logger
	logger := NewLogger(true)
	if logger == nil {
		t.Error("NewLogger with debug returned nil")
	}
}

func TestLoggerSetLevel(t *testing.T) {
	logger := NewLogger(false)

	// Test that SetLevel doesn't panic
	logger.SetLevel(DebugLevel)
	logger.SetLevel(InfoLevel)
	logger.SetLevel(WarnLevel)
	logger.SetLevel(ErrorLevel)
}

func TestLoggerBasicLogging(t *testing.T) {
	logger := NewLogger(false)

	// Test that basic logging doesn't panic
	logger.Info("test info message")
	logger.Warn("test warn message", "key", "value")
	logger.Error("test error message", "error", "test")
	logger.Debug("test debug message")
}

func TestLoggerWithFields(t *testing.T) {
	logger := NewLogger(false)

	// Test logging with various field types
	logger.Info("test with fields",
		"string", "value",
		"int", 42,
		"bool", true,
	)
}
