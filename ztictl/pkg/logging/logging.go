package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"ztictl/pkg/colors"
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
	return 0644
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

	if err := os.MkdirAll(logDirPath, getDirPermissions()); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create log directory %s, file logging disabled: %v\n", logDirPath, err)
		return
	}

	logFilePath := filepath.Join(logDirPath, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, getFilePermissions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create/open log file %s, file logging disabled: %v\n", logFilePath, err)
		return
	}

	loggerMutex.Lock()
	// Close previous file if it exists
	if logFile != nil {
		logFile.Close()
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
		logFile.Close()
		logFile = nil
		fileLogger = nil
	}
}

// getTimestamp returns a formatted timestamp like in the bash utils
func getTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// logToFile writes a timestamped message to the log file (thread-safe)
func logToFile(level string, format string, args ...interface{}) {
	loggerMutex.RLock()
	logger := fileLogger
	loggerMutex.RUnlock()

	if logger != nil {
		timestamp := getTimestamp()
		message := fmt.Sprintf(format, args...)
		logger.Printf("%s [%s] %s", timestamp, level, message)
	}
}

// LogInfo logs an info message - colored to console, timestamped to file
func LogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	colors.Success.Printf("[INFO] %s\n", message)
	logToFile("INFO", format, args...)
}

// LogWarn logs a warning message - colored to console, timestamped to file
func LogWarn(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	colors.Warning.Printf("[WARN] %s\n", message)
	logToFile("WARN", format, args...)
}

// LogError logs an error message - colored to console, timestamped to file
func LogError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	colors.Error.Printf("[ERROR] %s\n", message)
	logToFile("ERROR", format, args...)
}

// LogDebug logs a debug message - colored to console, timestamped to file
func LogDebug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	colors.Data.Printf("[DEBUG] %s\n", message)
	logToFile("DEBUG", format, args...)
}

// LogSuccess logs a success message - colored to console, timestamped to file
func LogSuccess(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	colors.Success.Printf("[SUCCESS] %s\n", message)
	logToFile("SUCCESS", format, args...)
}
