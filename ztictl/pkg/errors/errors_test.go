package errors

import (
	"errors"
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
