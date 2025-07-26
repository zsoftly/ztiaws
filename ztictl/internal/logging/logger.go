package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

// Level represents logging levels
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

// Logger is a structured logger with colored output
type Logger struct {
	*logrus.Logger
	colorEnabled bool
}

// ColorFormatter provides colored console output
type ColorFormatter struct {
	DisableColors bool
}

// Format formats the log entry with colors
func (f *ColorFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor *color.Color

	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = color.New(color.FgCyan)
	case logrus.InfoLevel:
		levelColor = color.New(color.FgGreen)
	case logrus.WarnLevel:
		levelColor = color.New(color.FgYellow)
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = color.New(color.FgRed)
	default:
		levelColor = color.New(color.FgWhite)
	}

	var formattedLevel string
	if f.DisableColors {
		formattedLevel = fmt.Sprintf("[%s]", entry.Level.String()[:4])
	} else {
		formattedLevel = levelColor.Sprintf("[%s]", entry.Level.String()[:4])
	}

	// Format: [LEVEL] message
	output := fmt.Sprintf("%s %s", formattedLevel, entry.Message)

	// Add fields if present
	if len(entry.Data) > 0 {
		output += " |"
		for key, value := range entry.Data {
			output += fmt.Sprintf(" %s=%v", key, value)
		}
	}

	return []byte(output + "\n"), nil
}

// NewLogger creates a new logger instance
func NewLogger(debug bool) *Logger {
	log := logrus.New()

	// Set level
	if debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	// Use custom formatter for console output
	log.SetFormatter(&ColorFormatter{
		DisableColors: !isTerminal(),
	})

	logger := &Logger{
		Logger:       log,
		colorEnabled: isTerminal(),
	}

	// Set up file logging if LOG_DIR is set
	if logDir := os.Getenv("LOG_DIR"); logDir != "" {
		logger.setupFileLogging(logDir)
	} else if homeDir, err := os.UserHomeDir(); err == nil {
		defaultLogDir := filepath.Join(homeDir, "logs")
		logger.setupFileLogging(defaultLogDir)
	}

	return logger
}

// setupFileLogging configures file logging
func (l *Logger) setupFileLogging(logDir string) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		l.Warn("Failed to create log directory", "dir", logDir, "error", err)
		return
	}

	// Create log file with date
	logFile := filepath.Join(logDir, fmt.Sprintf("ztictl-%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		l.Warn("Failed to open log file", "file", logFile, "error", err)
		return
	}

	// Use MultiWriter to write to both console and file
	multiWriter := io.MultiWriter(os.Stdout, file)
	l.SetOutput(multiWriter)

	l.Debug("File logging enabled", "file", logFile)
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level Level) {
	switch level {
	case DebugLevel:
		l.Logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		l.Logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		l.Logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		l.Logger.SetLevel(logrus.ErrorLevel)
	}
}

// Info logs an info message with optional fields
func (l *Logger) Info(msg string, fields ...interface{}) {
	entry := l.Logger.WithFields(l.fieldsToLogrusFields(fields))
	entry.Info(msg)
}

// Debug logs a debug message with optional fields
func (l *Logger) Debug(msg string, fields ...interface{}) {
	entry := l.Logger.WithFields(l.fieldsToLogrusFields(fields))
	entry.Debug(msg)
}

// Warn logs a warning message with optional fields
func (l *Logger) Warn(msg string, fields ...interface{}) {
	entry := l.Logger.WithFields(l.fieldsToLogrusFields(fields))
	entry.Warn(msg)
}

// Error logs an error message with optional fields
func (l *Logger) Error(msg string, fields ...interface{}) {
	entry := l.Logger.WithFields(l.fieldsToLogrusFields(fields))
	entry.Error(msg)
}

// fieldsToLogrusFields converts alternating key-value pairs to logrus.Fields
func (l *Logger) fieldsToLogrusFields(fields []interface{}) logrus.Fields {
	logrusFields := make(logrus.Fields)

	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			logrusFields[key] = fields[i+1]
		}
	}

	return logrusFields
}

// isTerminal returns true if the output is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
