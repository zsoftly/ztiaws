package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"ztictl/pkg/colors"
	"ztictl/pkg/security"
)

var (
	fileLogger  *log.Logger
	logFile     *os.File // Store file handle for proper cleanup
	loggerMutex sync.RWMutex
)

func init() {
	// Initialize file logger
	setupFileLogger()
}

// getDefaultLogDir returns platform-appropriate default log directory
func getDefaultLogDir(homeDir string) string {
	switch runtime.GOOS {
	case "windows":
		// Windows: Use %LOCALAPPDATA%\ztictl\logs or fallback to home
		if appData := os.Getenv("LOCALAPPDATA"); appData != "" {
			return filepath.Join(appData, "ztictl", "logs")
		}
		return filepath.Join(homeDir, "AppData", "Local", "ztictl", "logs")
	case "darwin":
		// macOS: Use ~/Library/Logs/ztictl
		return filepath.Join(homeDir, "Library", "Logs", "ztictl")
	default:
		// Linux and others: Use ~/.local/share/ztictl/logs (XDG Base Directory)
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			return filepath.Join(xdgData, "ztictl", "logs")
		}
		return filepath.Join(homeDir, ".local", "share", "ztictl", "logs")
	}
}

// getFilePermissions returns platform-appropriate file permissions
func getFilePermissions() os.FileMode {
	if runtime.GOOS == "windows" {
		// Windows doesn't use Unix-style permissions
		return 0666
	}
	return 0600
}

// getDirPermissions returns platform-appropriate directory permissions
func getDirPermissions() os.FileMode {
	if runtime.GOOS == "windows" {
		// Windows doesn't use Unix-style permissions
		return 0777
	}
	return 0755
}

// setupFileLogger creates a timestamped log file with configurable directory
func setupFileLogger() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not determine home directory, file logging disabled: %v\n", err)
		return
	}

	// Allow configurable log directory via environment variable
	logDirPath := os.Getenv("ZTICTL_LOG_DIR")
	if logDirPath == "" {
		// Use platform-appropriate default location
		logDirPath = getDefaultLogDir(homeDir)
	}

	// Validate log directory path to prevent directory traversal attacks
	if security.ContainsUnsafePath(logDirPath) {
		fmt.Fprintf(os.Stderr, "Warning: Invalid log directory path %s, using default location\n", logDirPath)
		logDirPath = getDefaultLogDir(homeDir)
	}

	if err := os.MkdirAll(logDirPath, getDirPermissions()); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create log directory %s, file logging disabled: %v\n", logDirPath, err)
		return
	}

	logFilePath := filepath.Join(logDirPath, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
	// #nosec G304 - logDirPath is validated above and log filename is controlled by application
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, getFilePermissions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create/open log file %s, file logging disabled: %v\n", logFilePath, err)
		return
	}

	loggerMutex.Lock()
	// Close previous file if it exists
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			// Log the error to stderr since we're setting up a new file logger
			fmt.Fprintf(os.Stderr, "Warning: Error closing previous log file: %v\n", err)
		}
	}
	logFile = file
	fileLogger = log.New(file, "", 0) // No prefix, we'll add our own timestamp
	loggerMutex.Unlock()
}

// CloseLogger properly closes the log file to prevent resource leaks
// Should be called during application shutdown
func CloseLogger() {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	if logFile != nil {
		if err := logFile.Close(); err != nil {
			// Log the error to stderr since the file logger is being closed
			fmt.Fprintf(os.Stderr, "Warning: Error closing log file: %v\n", err)
		}
		logFile = nil
		fileLogger = nil
	}
}

// getTimestamp returns a formatted timestamp like in the bash utils
func getTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// logToFile writes a timestamped message to the log file (thread-safe)
func logToFile(level string, message string) {
	loggerMutex.RLock()
	logger := fileLogger
	loggerMutex.RUnlock()

	if logger != nil {
		timestamp := getTimestamp()
		logger.Printf("%s [%s] %s", timestamp, level, message)
	}
}

// LogInfo logs an info message - colored to console, timestamped to file
func LogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	_, _ = colors.Success.Printf("[INFO] %s\n", message)
	logToFile("INFO", message)
}

// LogWarn logs a warning message - colored to console, timestamped to file
func LogWarn(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	_, _ = colors.Warning.Printf("[WARN] %s\n", message)
	logToFile("WARN", message)
}

// LogError logs an error message - colored to console, timestamped to file
func LogError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	_, _ = colors.Error.Printf("[ERROR] %s\n", message)
	logToFile("ERROR", message)
}

// LogDebug logs a debug message - colored to console, timestamped to file
func LogDebug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	_, _ = colors.Data.Printf("[DEBUG] %s\n", message)
	logToFile("DEBUG", message)
}

// LogSuccess logs a success message - colored to console, timestamped to file
func LogSuccess(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	_, _ = colors.Success.Printf("[SUCCESS] %s\n", message)
	logToFile("SUCCESS", message)
}

// Compatibility layer for legacy Logger pattern
// This allows existing code to continue working while providing consistent uppercase logging

// Logger provides compatibility for code that expects the old logger interface
type Logger struct {
	debugEnabled bool
	noOp         bool
}

// Level represents logging levels (for compatibility)
type Level int

const (
	// DebugLevel for debug messages
	DebugLevel Level = iota
	// InfoLevel for info messages
	InfoLevel
	// WarnLevel for warning messages
	WarnLevel
	// ErrorLevel for error messages
	ErrorLevel
)

// NewLogger creates a new logger instance with debug level control
func NewLogger(debug bool) *Logger {
	return &Logger{
		debugEnabled: debug,
	}
}

// NewNoOpLogger creates a logger that discards all output
func NewNoOpLogger() *Logger {
	return &Logger{
		debugEnabled: false,
		noOp:         true,
	}
}

// SetLevel is a compatibility method that does nothing.
// It exists only to support legacy code that expects a SetLevel method.
// Logging level is handled globally in the new logging system.
func (l *Logger) SetLevel(level Level) {
	_ = level // Explicitly acknowledge unused parameter to avoid linter warnings
	// No-op since our centralized logging handles this globally
}

// formatFields converts key-value pairs to a formatted string
func (l *Logger) formatFields(fields ...interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	var parts []string
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			parts = append(parts, fmt.Sprintf("%v=%v", fields[i], fields[i+1]))
		} else {
			parts = append(parts, fmt.Sprintf("%v=<no_value>", fields[i]))
		}
	}

	if len(parts) > 0 {
		return " | " + strings.Join(parts, " ")
	}
	return ""
}

// Info logs an info message using centralized logging
func (l *Logger) Info(msg string, fields ...interface{}) {
	if l.noOp {
		return
	}
	fieldsStr := l.formatFields(fields...)
	LogInfo("%s%s", msg, fieldsStr)
}

// Debug logs a debug message using centralized logging (respects debug flag)
func (l *Logger) Debug(msg string, fields ...interface{}) {
	if l.noOp || !l.debugEnabled {
		return // Skip debug messages when no-op or debug is disabled
	}
	fieldsStr := l.formatFields(fields...)
	LogDebug("%s%s", msg, fieldsStr)
}

// Warn logs a warning message using centralized logging
func (l *Logger) Warn(msg string, fields ...interface{}) {
	if l.noOp {
		return
	}
	fieldsStr := l.formatFields(fields...)
	LogWarn("%s%s", msg, fieldsStr)
}

// Error logs an error message using centralized logging
func (l *Logger) Error(msg string, fields ...interface{}) {
	if l.noOp {
		return
	}
	fieldsStr := l.formatFields(fields...)
	LogError("%s%s", msg, fieldsStr)
}
