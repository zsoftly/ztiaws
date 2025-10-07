package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewValidationError(t *testing.T) {
	msg := "test validation error"
	err := NewValidationError(msg)

	if err == nil {
		t.Error("NewValidationError returned nil")
	}

	if err.Error() != "validation error: "+msg {
		t.Errorf("Expected error message to contain %q, got %q", msg, err.Error())
	}

	// Test that it's the correct type
	var ztiErr *ZtiError
	if !errors.As(err, &ztiErr) {
		t.Error("Error is not of type ZtiError")
	}

	if ztiErr.Type != ErrTypeValidation {
		t.Errorf("Expected error type %q, got %q", ErrTypeValidation, ztiErr.Type)
	}
}

func TestNewConfigError(t *testing.T) {
	msg := "test config error"
	wrappedErr := errors.New("wrapped error")
	err := NewConfigError(msg, wrappedErr)

	if err == nil {
		t.Error("NewConfigError returned nil")
	}

	if !contains(err.Error(), msg) {
		t.Errorf("Error message should contain %q, got %q", msg, err.Error())
	}

	// Test that it's the correct type
	var ztiErr *ZtiError
	if !errors.As(err, &ztiErr) {
		t.Error("Error is not of type ZtiError")
	}

	if ztiErr.Type != ErrTypeConfig {
		t.Errorf("Expected error type %q, got %q", ErrTypeConfig, ztiErr.Type)
	}

	// Test unwrapping
	if !errors.Is(err, wrappedErr) {
		t.Error("Error should wrap the original error")
	}
}

func TestNewAuthError(t *testing.T) {
	msg := "test auth error"
	wrappedErr := errors.New("wrapped error")
	err := NewAuthError(msg, wrappedErr)

	if err == nil {
		t.Error("NewAuthError returned nil")
	}

	var ztiErr *ZtiError
	if !errors.As(err, &ztiErr) {
		t.Error("Error is not of type ZtiError")
	}

	if ztiErr.Type != ErrTypeAuth {
		t.Errorf("Expected error type %q, got %q", ErrTypeAuth, ztiErr.Type)
	}
}

func TestNewSSMError(t *testing.T) {
	msg := "test ssm error"
	wrappedErr := errors.New("wrapped error")
	err := NewSSMError(msg, wrappedErr)

	if err == nil {
		t.Error("NewSSMError returned nil")
	}

	var ztiErr *ZtiError
	if !errors.As(err, &ztiErr) {
		t.Error("Error is not of type ZtiError")
	}

	if ztiErr.Type != ErrTypeSSM {
		t.Errorf("Expected error type %q, got %q", ErrTypeSSM, ztiErr.Type)
	}
}

func TestZtiErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrType
		expected string
	}{
		{"auth error", ErrTypeAuth, "auth"},
		{"ssm error", ErrTypeSSM, "ssm"},
		{"config error", ErrTypeConfig, "config"},
		{"aws error", ErrTypeAWS, "aws"},
		{"validation error", ErrTypeValidation, "validation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.errType) != tt.expected {
				t.Errorf("Error type %v = %q, want %q", tt.errType, string(tt.errType), tt.expected)
			}
		})
	}
}

func TestErrorChaining(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := NewConfigError("config failed", originalErr)

	// Test that we can check for specific error types
	var ztiErr *ZtiError
	if !errors.As(wrappedErr, &ztiErr) {
		t.Error("Should be able to extract ZtiError")
	}

	// Test that original error is preserved
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("Should preserve original error in chain")
	}

	// Test error type
	if ztiErr.Type != ErrTypeConfig {
		t.Errorf("Expected error type %q, got %q", ErrTypeConfig, ztiErr.Type)
	}
}

func TestNew(t *testing.T) {
	message := "test error message"
	err := New(ErrTypeAuth, message)

	if err == nil {
		t.Error("New should not return nil")
	}

	if err.Type != ErrTypeAuth {
		t.Errorf("Expected error type %s, got %s", ErrTypeAuth, err.Type)
	}

	if err.Message != message {
		t.Errorf("Expected message %s, got %s", message, err.Message)
	}

	if err.Underlying != nil {
		t.Error("New should not have underlying error")
	}

	if err.Context == nil {
		t.Error("New should initialize Context map")
	}

	if len(err.Context) != 0 {
		t.Error("New should initialize empty Context map")
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	message := "wrapped error message"
	err := Wrap(ErrTypeSSM, message, originalErr)

	if err == nil {
		t.Error("Wrap should not return nil")
	}

	if err.Type != ErrTypeSSM {
		t.Errorf("Expected error type %s, got %s", ErrTypeSSM, err.Type)
	}

	if err.Message != message {
		t.Errorf("Expected message %s, got %s", message, err.Message)
	}

	if !errors.Is(err.Underlying, originalErr) {
		t.Error("Wrap should preserve underlying error")
	}

	if err.Context == nil {
		t.Error("Wrap should initialize Context map")
	}
}

func TestZtiErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ZtiError
		expected string
	}{
		{
			name: "error without underlying",
			err: &ZtiError{
				Type:    ErrTypeAuth,
				Message: "auth failed",
			},
			expected: "auth error: auth failed",
		},
		{
			name: "error with underlying",
			err: &ZtiError{
				Type:       ErrTypeSSM,
				Message:    "ssm operation failed",
				Underlying: errors.New("network error"),
			},
			expected: "ssm error: ssm operation failed (caused by: network error)",
		},
		{
			name: "error with empty message",
			err: &ZtiError{
				Type:    ErrTypeConfig,
				Message: "",
			},
			expected: "config error: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestZtiErrorUnwrap(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name     string
		err      *ZtiError
		expected error
	}{
		{
			name: "error with underlying",
			err: &ZtiError{
				Underlying: originalErr,
			},
			expected: originalErr,
		},
		{
			name: "error without underlying",
			err: &ZtiError{
				Underlying: nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Unwrap()
			if !errors.Is(result, tt.expected) {
				t.Errorf("Unwrap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestZtiErrorWithContext(t *testing.T) {
	err := New(ErrTypeValidation, "validation failed")

	// Test adding context
	result := err.WithContext("field", "username")

	// Should return the same error instance
	if result != err {
		t.Error("WithContext should return the same error instance")
	}

	// Verify context was added
	val, exists := err.GetContext("field")
	if !exists {
		t.Error("Context value should exist")
	}

	if val != "username" {
		t.Errorf("Context value = %v, want 'username'", val)
	}

	// Test chaining context
	err.WithContext("operation", "login").WithContext("retry", 3)

	if len(err.Context) != 3 {
		t.Errorf("Expected 3 context entries, got %d", len(err.Context))
	}

	// Verify all context values
	expectedContext := map[string]interface{}{
		"field":     "username",
		"operation": "login",
		"retry":     3,
	}

	for key, expectedVal := range expectedContext {
		val, exists := err.GetContext(key)
		if !exists {
			t.Errorf("Context key %s should exist", key)
		}
		if val != expectedVal {
			t.Errorf("Context[%s] = %v, want %v", key, val, expectedVal)
		}
	}
}

func TestZtiErrorGetContext(t *testing.T) {
	err := New(ErrTypeAWS, "aws error")

	// Test getting non-existent key
	val, exists := err.GetContext("nonexistent")
	if exists {
		t.Error("Non-existent key should not exist")
	}
	if val != nil {
		t.Error("Non-existent key should return nil value")
	}

	// Add context and test retrieval
	err.WithContext("region", "us-east-1")
	err.WithContext("service", "ec2")
	err.WithContext("count", 42)
	err.WithContext("enabled", true)
	err.WithContext("data", nil) // Test nil value

	tests := []struct {
		key      string
		expected interface{}
		exists   bool
	}{
		{"region", "us-east-1", true},
		{"service", "ec2", true},
		{"count", 42, true},
		{"enabled", true, true},
		{"data", nil, true},
		{"missing", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			val, exists := err.GetContext(tt.key)
			if exists != tt.exists {
				t.Errorf("GetContext(%s) exists = %v, want %v", tt.key, exists, tt.exists)
			}
			if val != tt.expected {
				t.Errorf("GetContext(%s) = %v, want %v", tt.key, val, tt.expected)
			}
		})
	}
}

func TestNewAWSError(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		underlying error
		shouldWrap bool
	}{
		{"without underlying", "aws service error", nil, false},
		{"with underlying", "aws service failed", errors.New("api error"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAWSError(tt.message, tt.underlying)

			if err == nil {
				t.Error("NewAWSError should not return nil")
			}

			if err.Type != ErrTypeAWS {
				t.Errorf("Expected error type %s, got %s", ErrTypeAWS, err.Type)
			}

			if err.Message != tt.message {
				t.Errorf("Expected message %s, got %s", tt.message, err.Message)
			}

			if tt.shouldWrap {
				if !errors.Is(err.Underlying, tt.underlying) {
					t.Error("Should wrap underlying error")
				}
			} else {
				if err.Underlying != nil {
					t.Error("Should not have underlying error")
				}
			}
		})
	}
}

func TestErrorTypeStringConversion(t *testing.T) {
	tests := []struct {
		errType  ErrType
		expected string
	}{
		{ErrTypeAuth, "auth"},
		{ErrTypeSSM, "ssm"},
		{ErrTypeConfig, "config"},
		{ErrTypeAWS, "aws"},
		{ErrTypeValidation, "validation"},
	}

	for _, tt := range tests {
		t.Run(string(tt.errType), func(t *testing.T) {
			if string(tt.errType) != tt.expected {
				t.Errorf("ErrType string conversion = %s, want %s", string(tt.errType), tt.expected)
			}
		})
	}
}

func TestComplexErrorChaining(t *testing.T) {
	// Create a chain of errors
	originalErr := errors.New("network timeout")
	wrappedErr := NewAWSError("EC2 API call failed", originalErr)
	doubleWrappedErr := NewSSMError("SSM operation failed", wrappedErr)

	// Test error chain
	if !errors.Is(doubleWrappedErr, originalErr) {
		t.Error("Should be able to find original error in chain")
	}

	if !errors.Is(doubleWrappedErr, wrappedErr) {
		t.Error("Should be able to find wrapped error in chain")
	}

	// Test As functionality
	var ztiErr *ZtiError
	if !errors.As(doubleWrappedErr, &ztiErr) {
		t.Error("Should be able to extract ZtiError from chain")
	}

	if ztiErr.Type != ErrTypeSSM {
		t.Errorf("Expected extracted error type %s, got %s", ErrTypeSSM, ztiErr.Type)
	}
}

func TestContextTypes(t *testing.T) {
	err := New(ErrTypeValidation, "validation error")

	// Test various context value types
	testCases := map[string]interface{}{
		"string": "test string",
		"int":    42,
		"float":  3.14,
		"bool":   true,
		"nil":    nil,
		"slice":  []string{"a", "b", "c"},
		"map":    map[string]int{"count": 5},
		"struct": struct{ Name string }{Name: "test"},
	}

	// Add all context values
	for key, value := range testCases {
		err.WithContext(key, value)
	}

	// Verify all values
	for key, expectedValue := range testCases {
		t.Run(key, func(t *testing.T) {
			val, exists := err.GetContext(key)
			if !exists {
				t.Errorf("Context key %s should exist", key)
			}

			// Use reflect.DeepEqual for complex types
			if !deepEqual(val, expectedValue) {
				t.Errorf("Context[%s] = %#v, want %#v", key, val, expectedValue)
			}
		})
	}
}

func TestErrorInterfaceCompliance(t *testing.T) {
	var err error = New(ErrTypeAuth, "test error")

	// Should implement error interface
	if err.Error() == "" {
		t.Error("Error should implement error interface with non-empty string")
	}

	// Test type assertion
	ztiErr := &ZtiError{}
	ok := errors.As(err, &ztiErr)
	if !ok {
		t.Error("Should be able to type assert to *ZtiError")
	}

	if ztiErr.Type != ErrTypeAuth {
		t.Error("Type assertion should preserve error details")
	}
}

func TestErrorWithEmptyContext(t *testing.T) {
	err := &ZtiError{
		Type:    ErrTypeConfig,
		Message: "config error",
		Context: nil, // Explicitly nil context
	}

	// Should handle nil context gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetContext should not panic with nil context map: %v", r)
		}
	}()

	val, exists := err.GetContext("test")
	if exists {
		t.Error("Should not find key in nil context")
	}
	if val != nil {
		t.Error("Should return nil value for nil context")
	}
}

func TestAllErrorConstructors(t *testing.T) {
	testErr := errors.New("underlying error")

	constructors := []struct {
		name         string
		constructor  func() *ZtiError
		expectedType ErrType
	}{
		{"NewAuthError nil", func() *ZtiError { return NewAuthError("auth failed", nil) }, ErrTypeAuth},
		{"NewAuthError with underlying", func() *ZtiError { return NewAuthError("auth failed", testErr) }, ErrTypeAuth},
		{"NewSSMError nil", func() *ZtiError { return NewSSMError("ssm failed", nil) }, ErrTypeSSM},
		{"NewSSMError with underlying", func() *ZtiError { return NewSSMError("ssm failed", testErr) }, ErrTypeSSM},
		{"NewConfigError nil", func() *ZtiError { return NewConfigError("config failed", nil) }, ErrTypeConfig},
		{"NewConfigError with underlying", func() *ZtiError { return NewConfigError("config failed", testErr) }, ErrTypeConfig},
		{"NewAWSError nil", func() *ZtiError { return NewAWSError("aws failed", nil) }, ErrTypeAWS},
		{"NewAWSError with underlying", func() *ZtiError { return NewAWSError("aws failed", testErr) }, ErrTypeAWS},
		{"NewValidationError", func() *ZtiError { return NewValidationError("validation failed") }, ErrTypeValidation},
	}

	for _, tt := range constructors {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()

			if err == nil {
				t.Error("Constructor should not return nil")
			}

			if err.Type != tt.expectedType {
				t.Errorf("Expected error type %s, got %s", tt.expectedType, err.Type)
			}

			if err.Context == nil {
				t.Error("Constructor should initialize context map")
			}

			// Verify error string format
			errStr := err.Error()
			if !contains(errStr, string(tt.expectedType)) {
				t.Errorf("Error string should contain type %s: %s", tt.expectedType, errStr)
			}
		})
	}
}

// Helper functions
func contains(str, substr string) bool {
	return len(str) >= len(substr) && indexOfSubstring(str, substr) >= 0
}

func indexOfSubstring(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func deepEqual(a, b interface{}) bool {
	// Simple equality check - for more complex types in a real project,
	// you might want to use reflect.DeepEqual
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%#v", a) == fmt.Sprintf("%#v", b)
}
