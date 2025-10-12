package aws

import (
	"context"
	"fmt"
	"testing"

	"ztictl/pkg/logging"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// MockClientPool implements ClientPoolInterface for testing
type MockClientPool struct{}

func (m *MockClientPool) GetSSMClient(ctx context.Context, region string) (*ssm.Client, error) {
	// Return error to simulate client unavailable
	return nil, fmt.Errorf("mock client pool: SSM client not available")
}

func (m *MockClientPool) GetEC2Client(ctx context.Context, region string) (*ec2.Client, error) {
	// Return error to simulate client unavailable
	return nil, fmt.Errorf("mock client pool: EC2 client not available")
}

func TestInstanceServiceCreation(t *testing.T) {
	logger := logging.NewNoOpLogger()
	mockPool := &MockClientPool{}

	service := NewInstanceService(mockPool, logger)

	if service == nil {
		t.Error("NewInstanceService should not return nil")
	}

	if service.clientPool != mockPool {
		t.Error("Instance service should store the provided client pool")
	}

	if service.logger != logger {
		t.Error("Instance service should store the provided logger")
	}
}

func TestInstanceServiceInterfaces(t *testing.T) {
	// Test that MockClientPool implements the interface
	var _ ClientPoolInterface = (*MockClientPool)(nil)
}

func TestSelectInstanceWithFallbackLogic(t *testing.T) {
	logger := logging.NewNoOpLogger()
	mockPool := &MockClientPool{}
	service := NewInstanceService(mockPool, logger)

	ctx := context.Background()

	// Test with provided identifier (should call ResolveInstanceIdentifier)
	// This will fail due to no real AWS clients, but tests the code path
	_, err := service.SelectInstanceWithFallback(ctx, "i-12345678", "us-east-1", nil)
	if err == nil {
		t.Error("Expected error when no AWS clients are available")
	}

	// Test with empty identifier (should try to list instances for fuzzy finder)
	// This will also fail due to no real AWS clients, but tests the code path
	_, err = service.SelectInstanceWithFallback(ctx, "", "us-east-1", nil)
	if err == nil {
		t.Error("Expected error when no AWS clients are available")
	}
}

func TestListFiltersNilHandling(t *testing.T) {
	logger := logging.NewNoOpLogger()
	mockPool := &MockClientPool{}
	service := NewInstanceService(mockPool, logger)

	ctx := context.Background()

	// Test with nil filters (should not panic)
	_, err := service.ListInstances(ctx, "us-east-1", nil)
	if err == nil {
		t.Error("Expected error when no AWS clients are available")
	}
	// The important thing is that it didn't panic due to nil filters
}

func TestIsInstanceIDValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid instance ID", "i-1234567890abcdef", true},
		{"valid short instance ID", "i-12345678", true},
		{"invalid prefix", "e-1234567890abcdef", false},
		{"too short", "i-123", false},
		{"uppercase letters", "i-123456789ABCDEF", false},
		{"special characters", "i-1234567890abcd@f", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInstanceID(tt.input)
			if result != tt.expected {
				t.Errorf("isInstanceID(%q) = %v; expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTagFilterParsing(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLen int
		expectError bool
	}{
		{"single tag", "Environment=prod", 1, false},
		{"multiple tags", "Environment=prod,Team=backend", 2, false},
		{"tag with equals in value", "Config=key=value", 1, false},
		{"empty string", "", 0, false},
		{"invalid format", "Environment", 0, true},
		{"empty key", "=value", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTagFilters(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("parseTagFilters(%q) expected error but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("parseTagFilters(%q) unexpected error: %v", tt.input, err)
				return
			}

			if len(result) != tt.expectedLen {
				t.Errorf("parseTagFilters(%q) returned %d items; expected %d", tt.input, len(result), tt.expectedLen)
			}
		})
	}
}
