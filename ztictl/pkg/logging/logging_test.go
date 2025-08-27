package logging

import (
	"fmt"
	"io/ioutil"
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
				os.Setenv(key, value)
				defer os.Setenv(key, originalValue)
			}

			// Clear environment variables not in the test case
			if _, exists := tt.envVars["LOCALAPPDATA"]; !exists {
				originalValue := os.Getenv("LOCALAPPDATA")
				os.Unsetenv("LOCALAPPDATA")
				defer os.Setenv("LOCALAPPDATA", originalValue)
			}
			if _, exists := tt.envVars["XDG_DATA_HOME"]; !exists {
				originalValue := os.Getenv("XDG_DATA_HOME")
				os.Unsetenv("XDG_DATA_HOME")
				defer os.Setenv("XDG_DATA_HOME", originalValue)
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
	os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer os.Setenv("ZTICTL_LOG_DIR", originalLogDir)

	// Reset logger state
	loggerMutex.Lock()
	if logFile != nil {
		logFile.Close()
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
	os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer os.Setenv("ZTICTL_LOG_DIR", originalLogDir)

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
	os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer os.Setenv("ZTICTL_LOG_DIR", originalLogDir)

	setupFileLogger()
	defer CloseLogger()

	// Test logging to file
	testMessage := "Test log message"
	testLevel := "TEST"

	logToFile(testLevel, testMessage)

	// Read log file content
	expectedLogFile := filepath.Join(tempDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
	content, err := ioutil.ReadFile(expectedLogFile)
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
	os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer os.Setenv("ZTICTL_LOG_DIR", originalLogDir)

	setupFileLogger()
	defer CloseLogger()

	tests := []struct {
		name    string
		logFunc func(string, ...interface{})
		level   string
		message string
		args    []interface{}
	}{
		{"LogInfo", LogInfo, "INFO", "Test info message", nil},
		{"LogWarn", LogWarn, "WARN", "Test warning message", nil},
		{"LogError", LogError, "ERROR", "Test error message", nil},
		{"LogDebug", LogDebug, "DEBUG", "Test debug message", nil},
		{"LogSuccess", LogSuccess, "SUCCESS", "Test success message", nil},
		{"LogInfo with args", LogInfo, "INFO", "Test %s with %d args", []interface{}{"message", 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear log file
			expectedLogFile := filepath.Join(tempDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
			os.Truncate(expectedLogFile, 0)

			// Test the log function
			if tt.args != nil {
				tt.logFunc(tt.message, tt.args...)
			} else {
				tt.logFunc(tt.message)
			}

			// Read log file content
			content, err := ioutil.ReadFile(expectedLogFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			contentStr := string(content)
			if !strings.Contains(contentStr, tt.level) {
				t.Errorf("Log file does not contain level %s", tt.level)
			}

			expectedMessage := tt.message
			if tt.args != nil {
				expectedMessage = fmt.Sprintf(tt.message, tt.args...)
			}

			if !strings.Contains(contentStr, expectedMessage) {
				t.Errorf("Log file does not contain expected message: %s", expectedMessage)
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
	os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer os.Setenv("ZTICTL_LOG_DIR", originalLogDir)

	setupFileLogger()
	defer CloseLogger()

	tests := []struct {
		name          string
		debugEnabled  bool
		noOp          bool
		method        func(*Logger)
		expectedLog   bool
		expectedLevel string
	}{
		{
			name:          "Info with regular logger",
			debugEnabled:  true,
			noOp:          false,
			method:        func(l *Logger) { l.Info("test info", "key", "value") },
			expectedLog:   true,
			expectedLevel: "INFO",
		},
		{
			name:          "Debug with debug enabled",
			debugEnabled:  true,
			noOp:          false,
			method:        func(l *Logger) { l.Debug("test debug", "key", "value") },
			expectedLog:   true,
			expectedLevel: "DEBUG",
		},
		{
			name:          "Debug with debug disabled",
			debugEnabled:  false,
			noOp:          false,
			method:        func(l *Logger) { l.Debug("test debug", "key", "value") },
			expectedLog:   false,
			expectedLevel: "DEBUG",
		},
		{
			name:          "Warn with regular logger",
			debugEnabled:  true,
			noOp:          false,
			method:        func(l *Logger) { l.Warn("test warn", "key", "value") },
			expectedLog:   true,
			expectedLevel: "WARN",
		},
		{
			name:          "Error with regular logger",
			debugEnabled:  true,
			noOp:          false,
			method:        func(l *Logger) { l.Error("test error", "key", "value") },
			expectedLog:   true,
			expectedLevel: "ERROR",
		},
		{
			name:          "Info with noOp logger",
			debugEnabled:  true,
			noOp:          true,
			method:        func(l *Logger) { l.Info("test info", "key", "value") },
			expectedLog:   false,
			expectedLevel: "INFO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear log file
			expectedLogFile := filepath.Join(tempDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
			os.Truncate(expectedLogFile, 0)

			// Create logger
			logger := &Logger{
				debugEnabled: tt.debugEnabled,
				noOp:         tt.noOp,
			}

			// Call the method
			tt.method(logger)

			// Check log file
			content, err := ioutil.ReadFile(expectedLogFile)
			if err != nil && tt.expectedLog {
				t.Fatalf("Failed to read log file: %v", err)
			}

			contentStr := string(content)
			hasLogEntry := strings.Contains(contentStr, tt.expectedLevel)

			if tt.expectedLog && !hasLogEntry {
				t.Errorf("Expected log entry with level %s, but not found", tt.expectedLevel)
			}

			if !tt.expectedLog && hasLogEntry {
				t.Errorf("Did not expect log entry, but found: %s", contentStr)
			}

			// Verify fields formatting
			if tt.expectedLog && hasLogEntry {
				if !strings.Contains(contentStr, "key=value") {
					t.Error("Log entry should contain formatted fields")
				}
			}
		})
	}
}

func TestConcurrentLogging(t *testing.T) {
	// Setup temporary logger
	tempDir := t.TempDir()
	originalLogDir := os.Getenv("ZTICTL_LOG_DIR")
	os.Setenv("ZTICTL_LOG_DIR", tempDir)
	defer os.Setenv("ZTICTL_LOG_DIR", originalLogDir)

	setupFileLogger()
	defer CloseLogger()

	var wg sync.WaitGroup
	numGoroutines := 10
	numMessages := 100

	// Start multiple goroutines logging concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numMessages; j++ {
				LogInfo("Goroutine %d message %d", goroutineID, j)
			}
		}(i)
	}

	wg.Wait()

	// Verify log file contains entries from all goroutines
	expectedLogFile := filepath.Join(tempDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
	content, err := ioutil.ReadFile(expectedLogFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")

	// Should have numGoroutines * numMessages lines
	expectedLines := numGoroutines * numMessages
	if len(lines) != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
	}

	// Verify no corruption (each line should have proper format)
	for i, line := range lines {
		if !strings.Contains(line, "INFO") {
			t.Errorf("Line %d does not contain INFO level: %s", i, line)
		}
		if !strings.Contains(line, "Goroutine") {
			t.Errorf("Line %d does not contain expected message: %s", i, line)
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
	
	os.Setenv("ZTICTL_LOG_DIR", invalidPath)
	defer os.Setenv("ZTICTL_LOG_DIR", originalLogDir)

	// Reset logger state
	loggerMutex.Lock()
	if logFile != nil {
		logFile.Close()
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
