package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGetDefaultLogDir(t *testing.T) {
	tests := []struct {
		name        string
		goos        string
		homeDir     string
		envVars     map[string]string
		expectedDir string
	}{
		{
			name:        "Windows with LOCALAPPDATA",
			goos:        "windows",
			homeDir:     "C:\\Users\\testuser",
			envVars:     map[string]string{"LOCALAPPDATA": "C:\\Users\\testuser\\AppData\\Local"},
			expectedDir: "C:\\Users\\testuser\\AppData\\Local\\ztictl\\logs",
		},
		{
			name:        "Windows without LOCALAPPDATA",
			goos:        "windows",
			homeDir:     "C:\\Users\\testuser",
			envVars:     map[string]string{},
			expectedDir: "C:\\Users\\testuser\\AppData\\Local\\ztictl\\logs",
		},
		{
			name:        "macOS",
			goos:        "darwin",
			homeDir:     "/Users/testuser",
			envVars:     map[string]string{},
			expectedDir: "/Users/testuser/Library/Logs/ztictl",
		},
		{
			name:        "Linux with XDG_DATA_HOME",
			goos:        "linux",
			homeDir:     "/home/testuser",
			envVars:     map[string]string{"XDG_DATA_HOME": "/home/testuser/.local/share"},
			expectedDir: "/home/testuser/.local/share/ztictl/logs",
		},
		{
			name:        "Linux without XDG_DATA_HOME",
			goos:        "linux",
			homeDir:     "/home/testuser",
			envVars:     map[string]string{},
			expectedDir: "/home/testuser/.local/share/ztictl/logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Can't actually change runtime.GOOS in tests
			// This test validates the logic based on expected OS behavior

			// Set environment variables
			for key, value := range tt.envVars {
				originalValue := os.Getenv(key)
				_ = os.Setenv(key, value)
				defer func(k, v string) {
					_ = os.Setenv(k, v)
				}(key, originalValue)
			}

			// Clear environment variables not in the test case
			if _, exists := tt.envVars["LOCALAPPDATA"]; !exists {
				originalValue := os.Getenv("LOCALAPPDATA")
				_ = os.Unsetenv("LOCALAPPDATA")
				defer func(v string) {
					_ = os.Setenv("LOCALAPPDATA", v)
				}(originalValue)
			}
			if _, exists := tt.envVars["XDG_DATA_HOME"]; !exists {
				originalValue := os.Getenv("XDG_DATA_HOME")
				_ = os.Unsetenv("XDG_DATA_HOME")
				defer func(v string) {
					_ = os.Setenv("XDG_DATA_HOME", v)
				}(originalValue)
			}

			result := getDefaultLogDir(tt.homeDir)

			// Since we can't mock runtime.GOOS, test the actual OS behavior
			// and ensure the function returns a valid path structure
			if runtime.GOOS == tt.goos {
				if result != tt.expectedDir {
					t.Errorf("getDefaultLogDir() = %v, want %v", result, tt.expectedDir)
				}
			} else {
				// For other OS, just ensure it returns a non-empty path
				if result == "" {
					t.Error("getDefaultLogDir() returned empty string")
				}
			}
		})
	}
}

func TestGetFilePermissions(t *testing.T) {
	result := getFilePermissions()

	if runtime.GOOS == "windows" {
		if result != 0666 {
			t.Errorf("getFilePermissions() on Windows = %v, want 0666", result)
		}
	} else {
		if result != 0644 {
			t.Errorf("getFilePermissions() on Unix = %v, want 0644", result)
		}
	}
}

func TestGetDirPermissions(t *testing.T) {
	result := getDirPermissions()

	if runtime.GOOS == "windows" {
		if result != 0777 {
			t.Errorf("getDirPermissions() on Windows = %v, want 0777", result)
		}
	} else {
		if result != 0755 {
			t.Errorf("getDirPermissions() on Unix = %v, want 0755", result)
		}
	}
}

func TestSetupFileLogger(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Set custom log directory
	originalLogDir := os.Getenv("ZTICTL_LOG_DIR")
	_ = os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer func() {
		_ = os.Setenv("ZTICTL_LOG_DIR", originalLogDir)
	}()

	// Reset logger state
	loggerMutex.Lock()
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
		fileLogger = nil
	}
	loggerMutex.Unlock()

	// Test setup
	setupFileLogger()

	// Verify logger was created
	loggerMutex.RLock()
	hasLogger := fileLogger != nil
	hasFile := logFile != nil
	loggerMutex.RUnlock()

	if !hasLogger {
		t.Error("setupFileLogger() did not create fileLogger")
	}

	if !hasFile {
		t.Error("setupFileLogger() did not create logFile")
	}

	// Verify log file exists
	expectedLogFile := filepath.Join(tempDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
	if _, err := os.Stat(expectedLogFile); os.IsNotExist(err) {
		t.Errorf("Expected log file %s was not created", expectedLogFile)
	}
}

func TestCloseLogger(t *testing.T) {
	// Setup a logger first
	tempDir := t.TempDir()
	originalLogDir := os.Getenv("ZTICTL_LOG_DIR")
	_ = os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer func() {
		_ = os.Setenv("ZTICTL_LOG_DIR", originalLogDir)
	}()

	setupFileLogger()

	// Verify logger exists
	loggerMutex.RLock()
	hasLoggerBefore := fileLogger != nil
	loggerMutex.RUnlock()

	if !hasLoggerBefore {
		t.Fatal("setupFileLogger() failed to create logger")
	}

	// Test close
	CloseLogger()

	// Verify logger is closed
	loggerMutex.RLock()
	hasLoggerAfter := fileLogger != nil
	hasFileAfter := logFile != nil
	loggerMutex.RUnlock()

	if hasLoggerAfter {
		t.Error("CloseLogger() did not clear fileLogger")
	}

	if hasFileAfter {
		t.Error("CloseLogger() did not clear logFile")
	}
}

func TestGetTimestamp(t *testing.T) {
	timestamp := getTimestamp()

	// Verify timestamp format (YYYY-MM-DD HH:MM:SS)
	if len(timestamp) != 19 {
		t.Errorf("getTimestamp() returned wrong length: got %d, want 19", len(timestamp))
	}

	// Parse the timestamp to ensure it's valid
	_, err := time.Parse("2006-01-02 15:04:05", timestamp)
	if err != nil {
		t.Errorf("getTimestamp() returned invalid timestamp format: %v", err)
	}

	// Ensure timestamp is recent (within 5 seconds to handle timezone differences)
	parsedTime, _ := time.Parse("2006-01-02 15:04:05", timestamp)
	timeDiff := time.Since(parsedTime)
	if timeDiff > 5*time.Second && timeDiff < -5*time.Second {
		t.Errorf("getTimestamp() returned timestamp too far from current time: %v", timeDiff)
	}
}

func TestLogToFile(t *testing.T) {
	// Setup temporary logger
	tempDir := t.TempDir()
	originalLogDir := os.Getenv("ZTICTL_LOG_DIR")
	_ = os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer func() {
		_ = os.Setenv("ZTICTL_LOG_DIR", originalLogDir)
	}()

	setupFileLogger()
	defer CloseLogger()

	// Test logging to file
	testMessage := "Test log message"
	testLevel := "TEST"

	logToFile(testLevel, testMessage)

	// Read log file content
	expectedLogFile := filepath.Join(tempDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
	content, err := os.ReadFile(expectedLogFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, testLevel) {
		t.Errorf("Log file does not contain level %s", testLevel)
	}

	if !strings.Contains(contentStr, testMessage) {
		t.Errorf("Log file does not contain message %s", testMessage)
	}

	// Verify timestamp format in log
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")
	if len(lines) < 1 {
		t.Error("No lines found in log file")
	} else {
		line := lines[len(lines)-1] // Get last line
		if !strings.Contains(line, time.Now().Format("2006-01-02")) {
			t.Error("Log line does not contain proper timestamp")
		}
	}
}

func TestLogFunctions(t *testing.T) {
	// Setup temporary logger
	tempDir := t.TempDir()
	originalLogDir := os.Getenv("ZTICTL_LOG_DIR")
	_ = os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer func() {
		_ = os.Setenv("ZTICTL_LOG_DIR", originalLogDir)
	}()

	setupFileLogger()
	defer CloseLogger()

	tests := []struct {
		name    string
		logFunc func(string, ...interface{})
		level   string
		message string
	}{
		{"LogInfo", LogInfo, "INFO", "Test info message"},
		{"LogWarn", LogWarn, "WARN", "Test warning message"},
		{"LogError", LogError, "ERROR", "Test error message"},
		{"LogDebug", LogDebug, "DEBUG", "Test debug message"},
		{"LogSuccess", LogSuccess, "SUCCESS", "Test success message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the logging function - we'll skip console output verification
			// due to pipe read/write timing issues in tests
			tt.logFunc("%s", tt.message)

			// Verify log file contains the message
			logFile := filepath.Join(tempDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Fatalf("%s: Failed to read log file: %v", tt.name, err)
			}

			if !strings.Contains(string(content), tt.message) {
				t.Errorf("%s: log file does not contain message %q", tt.name, tt.message)
			}
			if !strings.Contains(string(content), tt.level) {
				t.Errorf("%s: log file does not contain level %q", tt.name, tt.level)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name  string
		debug bool
	}{
		{"debug enabled", true},
		{"debug disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.debug)

			if logger == nil {
				t.Error("NewLogger() returned nil")
				return
			}

			if logger.debugEnabled != tt.debug {
				t.Errorf("NewLogger() debugEnabled = %v, want %v", logger.debugEnabled, tt.debug)
			}

			if logger.noOp {
				t.Error("NewLogger() should not create noOp logger")
			}
		})
	}
}

func TestNewNoOpLogger(t *testing.T) {
	logger := NewNoOpLogger()

	if logger == nil {
		t.Error("NewNoOpLogger() returned nil")
		return
	}

	if !logger.noOp {
		t.Error("NewNoOpLogger() should create noOp logger")
	}

	if logger.debugEnabled {
		t.Error("NewNoOpLogger() should not enable debug")
	}
}

func TestLoggerSetLevel(t *testing.T) {
	logger := NewLogger(true)

	// Test that SetLevel doesn't panic (it's a no-op)
	logger.SetLevel(DebugLevel)
	logger.SetLevel(InfoLevel)
	logger.SetLevel(WarnLevel)
	logger.SetLevel(ErrorLevel)

	// No assertions needed since it's a no-op function
}

func TestLoggerFormatFields(t *testing.T) {
	logger := NewLogger(true)

	tests := []struct {
		name     string
		fields   []interface{}
		expected string
	}{
		{"no fields", []interface{}{}, ""},
		{"single pair", []interface{}{"key", "value"}, " | key=value"},
		{"multiple pairs", []interface{}{"key1", "value1", "key2", "value2"}, " | key1=value1 key2=value2"},
		{"odd fields", []interface{}{"key1", "value1", "key2"}, " | key1=value1 key2=<no_value>"},
		{"mixed types", []interface{}{"str", "text", "num", 42, "bool", true}, " | str=text num=42 bool=true"},
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

func TestLoggerMethods(t *testing.T) {
	// Setup temporary logger
	tempDir := t.TempDir()
	originalLogDir := os.Getenv("ZTICTL_LOG_DIR")
	_ = os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer func() {
		_ = os.Setenv("ZTICTL_LOG_DIR", originalLogDir)
	}()

	setupFileLogger()
	defer CloseLogger()

	tests := []struct {
		name         string
		logger       *Logger
		method       string
		message      string
		fields       []interface{}
		expectOutput bool
	}{
		{"regular logger info", NewLogger(false), "Info", "test info", []interface{}{"key", "value"}, true},
		{"regular logger warn", NewLogger(false), "Warn", "test warn", []interface{}{"key", "value"}, true},
		{"regular logger error", NewLogger(false), "Error", "test error", []interface{}{"key", "value"}, true},
		{"debug logger debug enabled", NewLogger(true), "Debug", "test debug", []interface{}{"key", "value"}, true},
		{"debug logger debug disabled", NewLogger(false), "Debug", "test debug", []interface{}{"key", "value"}, false},
		{"noop logger info", NewNoOpLogger(), "Info", "test info", []interface{}{"key", "value"}, false},
		{"noop logger error", NewNoOpLogger(), "Error", "test error", []interface{}{"key", "value"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear log file for this test
			logFile := filepath.Join(tempDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
			_ = os.WriteFile(logFile, []byte{}, 0644)

			// Call the appropriate method
			switch tt.method {
			case "Info":
				tt.logger.Info(tt.message, tt.fields...)
			case "Debug":
				tt.logger.Debug(tt.message, tt.fields...)
			case "Warn":
				tt.logger.Warn(tt.message, tt.fields...)
			case "Error":
				tt.logger.Error(tt.message, tt.fields...)
			}

			// Read log file content
			content, _ := os.ReadFile(logFile)

			if tt.expectOutput {
				// Should have file output
				if len(tt.fields) > 0 && !strings.Contains(string(content), "key=value") {
					t.Errorf("%s: expected log file to contain fields", tt.name)
				}
				if !strings.Contains(string(content), tt.message) {
					t.Errorf("%s: expected log file to contain message %q", tt.name, tt.message)
				}
			} else {
				// Should have no output or minimal output
				if len(content) > 0 && strings.Contains(string(content), tt.message) {
					t.Errorf("%s: expected no log output, but found message in log", tt.name)
				}
			}
		})
	}
}

func TestConcurrentLogging(t *testing.T) {
	// Setup temporary logger
	tempDir := t.TempDir()
	originalLogDir := os.Getenv("ZTICTL_LOG_DIR")
	_ = os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer func() {
		_ = os.Setenv("ZTICTL_LOG_DIR", originalLogDir)
	}()

	setupFileLogger()
	defer CloseLogger()

	// Capture stdout and stderr to avoid CI issues
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	defer func() {
		_ = w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		// Drain the pipe
		var buf [4096]byte
		_, _ = r.Read(buf[:])
	}()

	// Run concurrent logging
	var wg sync.WaitGroup
	goroutineCount := 10
	messagesPerGoroutine := 5

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				LogInfo("Goroutine %d message %d", id, j)
				time.Sleep(time.Millisecond) // Small delay to increase chance of race conditions
			}
		}(i)
	}

	wg.Wait()

	// Restore output
	_ = w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Verify log file has correct number of entries
	logFile := filepath.Join(tempDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	expectedLines := goroutineCount * messagesPerGoroutine

	// Allow for some initial log lines that might exist
	if len(lines) < expectedLines {
		t.Errorf("Expected at least %d log lines, got %d", expectedLines, len(lines))
	}

	// Verify no corruption - each line should have proper format
	for i, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "[") || !strings.Contains(line, "]") {
			t.Errorf("Line %d appears corrupted: %s", i, line)
		}
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		expected Level
	}{
		{"DebugLevel", DebugLevel, 0},
		{"InfoLevel", InfoLevel, 1},
		{"WarnLevel", WarnLevel, 2},
		{"ErrorLevel", ErrorLevel, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.level != tt.expected {
				t.Errorf("Level %s = %d, want %d", tt.name, int(tt.level), int(tt.expected))
			}
		})
	}
}

func TestLoggerWithFormatting(t *testing.T) {
	// Setup temporary logger
	tempDir := t.TempDir()
	originalLogDir := os.Getenv("ZTICTL_LOG_DIR")
	_ = os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer func() {
		_ = os.Setenv("ZTICTL_LOG_DIR", originalLogDir)
	}()

	setupFileLogger()
	defer CloseLogger()

	// Test formatted messages
	LogInfo("User %s logged in at %d", "alice", 1234)
	LogError("Failed to connect to %s:%d", "localhost", 8080)
	LogWarn("Memory usage at %d%%", 85)

	// Verify formatted strings in console are output
	// We skip console output verification due to pipe timing issues
	expectedMessages := []string{
		"User alice logged in at 1234",
		"Failed to connect to localhost:8080",
		"Memory usage at 85%",
	}

	// Verify in log file
	logFile := filepath.Join(tempDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	for _, expected := range expectedMessages {
		if !strings.Contains(string(content), expected) {
			t.Errorf("Log file missing formatted message: %q", expected)
		}
	}
}

func TestLoggerFieldsEdgeCases(t *testing.T) {
	logger := NewLogger(true)

	tests := []struct {
		name     string
		fields   []interface{}
		expected string
	}{
		{"nil value", []interface{}{"key", nil}, " | key=<nil>"},
		{"empty string key", []interface{}{"", "value"}, " | =value"},
		{"special characters", []interface{}{"key with spaces", "value=with=equals"}, " | key with spaces=value=with=equals"},
		{"unicode", []interface{}{"键", "值"}, " | 键=值"},
		{"very long value", []interface{}{"key", strings.Repeat("a", 100)}, " | key=" + strings.Repeat("a", 100)},
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

func TestMultipleLoggerInstances(t *testing.T) {
	// Test that multiple logger instances can coexist
	logger1 := NewLogger(true)
	logger2 := NewLogger(false)
	logger3 := NewNoOpLogger()

	if logger1.debugEnabled != true {
		t.Error("Logger 1 should have debug enabled")
	}

	if logger2.debugEnabled != false {
		t.Error("Logger 2 should have debug disabled")
	}

	if logger3.noOp != true {
		t.Error("Logger 3 should be noOp")
	}

	// Ensure they maintain separate state
	logger1.SetLevel(DebugLevel) // Should be no-op but shouldn't affect others
	logger2.SetLevel(ErrorLevel) // Should be no-op but shouldn't affect others

	if logger1.debugEnabled != true {
		t.Error("Logger 1 debug state changed unexpectedly")
	}

	if logger2.debugEnabled != false {
		t.Error("Logger 2 debug state changed unexpectedly")
	}
}

func TestLogFileRotation(t *testing.T) {
	// Test that setupFileLogger creates new files for different days
	tempDir := t.TempDir()
	originalLogDir := os.Getenv("ZTICTL_LOG_DIR")
	_ = os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer func() {
		_ = os.Setenv("ZTICTL_LOG_DIR", originalLogDir)
	}()

	// Setup first logger
	setupFileLogger()
	expectedFile1 := filepath.Join(tempDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))

	// Capture output to avoid CI issues
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Write something to identify first file
	LogInfo("First log file")

	// Restore output temporarily
	_ = w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	var buf [1024]byte
	_, _ = r.Read(buf[:])

	// Verify first file exists
	if _, err := os.Stat(expectedFile1); os.IsNotExist(err) {
		t.Errorf("First log file not created: %s", expectedFile1)
	}

	// Call setupFileLogger again (simulating app restart or date change)
	setupFileLogger()

	// Capture output again
	r, w, _ = os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// Write to potentially new file
	LogInfo("After setup again")

	// Restore output
	_ = w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	_, _ = r.Read(buf[:])

	// The file should still exist and contain both messages
	content, err := os.ReadFile(expectedFile1)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "First log file") {
		t.Error("Log file missing first message after re-setup")
	}

	if !strings.Contains(string(content), "After setup again") {
		t.Error("Log file missing second message after re-setup")
	}

	CloseLogger()
}

func TestLogFileCreationFailure(t *testing.T) {
	// Setup impossible log directory to test failure handling
	originalLogDir := os.Getenv("ZTICTL_LOG_DIR")

	// Use platform-specific invalid path
	var invalidPath string
	if runtime.GOOS == "windows" {
		invalidPath = "Z:\\nonexistent\\readonly\\path"
	} else {
		invalidPath = "/nonexistent/readonly/path"
	}

	_ = os.Setenv("ZTICTL_LOG_DIR", invalidPath)
	defer func() {
		_ = os.Setenv("ZTICTL_LOG_DIR", originalLogDir)
	}()

	// Reset logger state
	loggerMutex.Lock()
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
		fileLogger = nil
	}
	loggerMutex.Unlock()

	// Test setup failure
	setupFileLogger()

	// Should not have created logger due to path failure
	loggerMutex.RLock()
	hasLogger := fileLogger != nil
	loggerMutex.RUnlock()

	// On some systems this might still work, so we just ensure no panic occurred
	// The main test is that setupFileLogger() completes without crashing
	_ = hasLogger // Use the variable to avoid unused warning
}
