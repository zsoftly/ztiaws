package errors

import (
	"fmt"
)

// ErrType represents different types of errors
type ErrType string

const (
	// ErrTypeAuth represents authentication-related errors
	ErrTypeAuth ErrType = "auth"
	// ErrTypeSSM represents SSM-related errors
	ErrTypeSSM ErrType = "ssm"
	// ErrTypeConfig represents configuration errors
	ErrTypeConfig ErrType = "config"
	// ErrTypeAWS represents AWS service errors
	ErrTypeAWS ErrType = "aws"
	// ErrTypeValidation represents validation errors
	ErrTypeValidation ErrType = "validation"
)

// ZtiError represents a custom error with context
type ZtiError struct {
	Type       ErrType
	Message    string
	Underlying error
	Context    map[string]interface{}
}

// Error implements the error interface
func (e *ZtiError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s error: %s (caused by: %v)", e.Type, e.Message, e.Underlying)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *ZtiError) Unwrap() error {
	return e.Underlying
}

// New creates a new ZtiError
func New(errType ErrType, message string) *ZtiError {
	return &ZtiError{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with a ZtiError
func Wrap(errType ErrType, message string, err error) *ZtiError {
	return &ZtiError{
		Type:       errType,
		Message:    message,
		Underlying: err,
		Context:    make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *ZtiError) WithContext(key string, value interface{}) *ZtiError {
	e.Context[key] = value
	return e
}

// GetContext returns context value
func (e *ZtiError) GetContext(key string) (interface{}, bool) {
	val, exists := e.Context[key]
	return val, exists
}

// Common error constructors
func NewAuthError(message string, err error) *ZtiError {
	if err != nil {
		return Wrap(ErrTypeAuth, message, err)
	}
	return New(ErrTypeAuth, message)
}

func NewSSMError(message string, err error) *ZtiError {
	if err != nil {
		return Wrap(ErrTypeSSM, message, err)
	}
	return New(ErrTypeSSM, message)
}

func NewConfigError(message string, err error) *ZtiError {
	if err != nil {
		return Wrap(ErrTypeConfig, message, err)
	}
	return New(ErrTypeConfig, message)
}

func NewAWSError(message string, err error) *ZtiError {
	if err != nil {
		return Wrap(ErrTypeAWS, message, err)
	}
	return New(ErrTypeAWS, message)
}

func NewValidationError(message string) *ZtiError {
	return New(ErrTypeValidation, message)
}
