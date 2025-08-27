package ssm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"ztictl/pkg/logging"
)

func TestNewIAMManager(t *testing.T) {
	logger := logging.NewNoOpLogger()

	manager, err := NewIAMManager(logger, nil, nil)

	if err != nil {
		t.Fatalf("NewIAMManager should not return error: %v", err)
	}

	if manager == nil {
		t.Fatal("NewIAMManager should not return nil")
	}

	if manager.logger != logger {
		t.Error("NewIAMManager should preserve logger")
	}
}

func TestGenerateUniqueID(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := &IAMManager{logger: logger}

	// Generate multiple IDs to test basic functionality
	id1 := manager.generateUniqueID()
	id2 := manager.generateUniqueID()

	// IDs should not be empty
	if id1 == "" {
		t.Error("generateUniqueID should not return empty string")
	}

	// IDs should be different
	if id1 == id2 {
		t.Error("generateUniqueID should generate unique IDs")
	}

	// ID should contain timestamp and hostname
	parts := strings.Split(id1, "-")
	if len(parts) < 3 {
		t.Errorf("generateUniqueID should contain at least 3 parts, got %d", len(parts))
	}

	// First part should be timestamp (numeric)
	if len(parts[0]) < 10 { // Unix timestamp should be at least 10 digits
		t.Error("generateUniqueID first part should be a valid timestamp")
	}

	// Second part should be hostname (from os.Hostname() or "unknown")
	hostname := parts[1]
	if hostname == "" {
		t.Error("generateUniqueID should contain hostname")
	}

	// Third part should be hex string (16 characters)
	hexPart := parts[len(parts)-1]
	if len(hexPart) != 16 {
		t.Errorf("Hex part should be 16 characters, got %d", len(hexPart))
	}

	// Verify hex characters
	for _, char := range hexPart {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Errorf("Invalid hex character %c in ID: %s", char, id1)
		}
	}
}

func TestCreateS3PolicyDocument(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := &IAMManager{logger: logger}

	tests := []struct {
		name        string
		bucketName  string
		expectError bool
	}{
		{"valid bucket name", "test-bucket", false},
		{"bucket with numbers", "test-bucket-123", false},
		{"bucket with dashes", "my-test-bucket", false},
		{"empty bucket name", "", false}, // Should still create policy, just with empty resource
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policyDoc, err := manager.createS3PolicyDocument(tt.bucketName)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// Verify policy document is valid JSON
				var policy S3PolicyDocument
				if err := json.Unmarshal([]byte(policyDoc), &policy); err != nil {
					t.Errorf("Policy document is not valid JSON: %v", err)
				}

				// Verify policy structure
				if policy.Version != "2012-10-17" {
					t.Error("Policy should have correct version")
				}

				if len(policy.Statement) != 2 {
					t.Errorf("Policy should have 2 statements, got %d", len(policy.Statement))
				}

				// Check first statement (object operations)
				stmt1 := policy.Statement[0]
				if stmt1.Effect != "Allow" {
					t.Error("First statement should have Allow effect")
				}
				if len(stmt1.Action) != 3 {
					t.Errorf("First statement should have 3 actions, got %d", len(stmt1.Action))
				}
				expectedActions := []string{"s3:GetObject", "s3:PutObject", "s3:DeleteObject"}
				for _, expectedAction := range expectedActions {
					found := false
					for _, action := range stmt1.Action {
						if action == expectedAction {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("First statement missing action %s", expectedAction)
					}
				}

				if tt.bucketName != "" {
					expectedResource := fmt.Sprintf("arn:aws:s3:::%s/*", tt.bucketName)
					if stmt1.Resource != expectedResource {
						t.Errorf("First statement resource should be %s, got %s", expectedResource, stmt1.Resource)
					}
				}

				// Check second statement (bucket operations)
				stmt2 := policy.Statement[1]
				if stmt2.Effect != "Allow" {
					t.Error("Second statement should have Allow effect")
				}
				if len(stmt2.Action) != 1 || stmt2.Action[0] != "s3:ListBucket" {
					t.Error("Second statement should have s3:ListBucket action")
				}

				if tt.bucketName != "" {
					expectedResource := fmt.Sprintf("arn:aws:s3:::%s", tt.bucketName)
					if stmt2.Resource != expectedResource {
						t.Errorf("Second statement resource should be %s, got %s", expectedResource, stmt2.Resource)
					}
				}
			}
		})
	}
}

func TestS3PolicyDocument(t *testing.T) {
	tests := []struct {
		name         string
		bucketName   string
		expectedJSON string
	}{
		{
			name:       "valid bucket name",
			bucketName: "test-bucket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := S3PolicyDocument{
				Version: "2012-10-17",
				Statement: []Statement{
					{
						Effect: "Allow",
						Action: []string{
							"s3:GetObject",
							"s3:PutObject",
							"s3:DeleteObject",
						},
						Resource: fmt.Sprintf("arn:aws:s3:::%s/*", tt.bucketName),
					},
					{
						Effect:   "Allow",
						Action:   []string{"s3:ListBucket"},
						Resource: fmt.Sprintf("arn:aws:s3:::%s", tt.bucketName),
					},
				},
			}

			// Test that the policy can be serialized to JSON
			jsonBytes, err := json.Marshal(policy)
			if err != nil {
				t.Fatalf("Failed to marshal policy to JSON: %v", err)
			}

			// Verify it contains expected elements
			jsonStr := string(jsonBytes)
			if !strings.Contains(jsonStr, "2012-10-17") {
				t.Error("Policy JSON should contain version")
			}
			if !strings.Contains(jsonStr, tt.bucketName) {
				t.Error("Policy JSON should contain bucket name")
			}
			if !strings.Contains(jsonStr, "s3:GetObject") {
				t.Error("Policy JSON should contain s3:GetObject action")
			}

			// Test deserialization
			var deserializedPolicy S3PolicyDocument
			if err := json.Unmarshal(jsonBytes, &deserializedPolicy); err != nil {
				t.Fatalf("Failed to unmarshal policy from JSON: %v", err)
			}

			if deserializedPolicy.Version != policy.Version {
				t.Error("Deserialized policy should have same version")
			}

			if len(deserializedPolicy.Statement) != len(policy.Statement) {
				t.Error("Deserialized policy should have same number of statements")
			}
		})
	}
}

func TestPolicyCleanupFunc(t *testing.T) {
	// Test that PolicyCleanupFunc type works as expected
	var cleanupFunc PolicyCleanupFunc = func() error {
		return nil
	}

	err := cleanupFunc()
	if err != nil {
		t.Errorf("PolicyCleanupFunc should execute without error: %v", err)
	}

	// Test cleanup func that returns error
	cleanupFuncWithError := PolicyCleanupFunc(func() error {
		return fmt.Errorf("cleanup failed")
	})

	err = cleanupFuncWithError()
	if err == nil {
		t.Error("PolicyCleanupFunc should return error when configured to do so")
	}
	if err.Error() != "cleanup failed" {
		t.Errorf("Expected 'cleanup failed' error, got %v", err)
	}
}

func TestConstants(t *testing.T) {
	// Test IAM propagation delay constant
	if IAMPropagationDelay != 5*time.Second {
		t.Errorf("IAMPropagationDelay should be 5 seconds, got %v", IAMPropagationDelay)
	}

	// Test policy name prefix constant
	if PolicyNamePrefix != "ZTIaws-SSM-S3-Access" {
		t.Errorf("PolicyNamePrefix should be 'ZTIaws-SSM-S3-Access', got %s", PolicyNamePrefix)
	}
}

func TestStatementStruct(t *testing.T) {
	statement := Statement{
		Effect:   "Allow",
		Action:   []string{"s3:GetObject", "s3:PutObject"},
		Resource: "arn:aws:s3:::test-bucket/*",
	}

	// Test JSON serialization
	jsonBytes, err := json.Marshal(statement)
	if err != nil {
		t.Fatalf("Failed to marshal Statement to JSON: %v", err)
	}

	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, "Allow") {
		t.Error("Statement JSON should contain Effect")
	}
	if !strings.Contains(jsonStr, "s3:GetObject") {
		t.Error("Statement JSON should contain actions")
	}
	if !strings.Contains(jsonStr, "test-bucket") {
		t.Error("Statement JSON should contain resource")
	}

	// Test JSON deserialization
	var deserializedStatement Statement
	if err := json.Unmarshal(jsonBytes, &deserializedStatement); err != nil {
		t.Fatalf("Failed to unmarshal Statement from JSON: %v", err)
	}

	if deserializedStatement.Effect != statement.Effect {
		t.Error("Deserialized statement should have same Effect")
	}

	if len(deserializedStatement.Action) != len(statement.Action) {
		t.Error("Deserialized statement should have same number of actions")
	}

	if deserializedStatement.Resource != statement.Resource {
		t.Error("Deserialized statement should have same Resource")
	}
}

func TestIAMManagerStructFields(t *testing.T) {
	logger := logging.NewNoOpLogger()

	manager := &IAMManager{
		logger: logger,
	}

	// Test that all fields are properly accessible
	if manager.logger == nil {
		t.Error("IAMManager logger field should not be nil")
	}

	// Test setting clients to nil (should be acceptable)
	manager.iamClient = nil
	manager.ec2Client = nil

	if manager.iamClient != nil {
		t.Error("IAMManager iamClient should be nil when set to nil")
	}
	if manager.ec2Client != nil {
		t.Error("IAMManager ec2Client should be nil when set to nil")
	}
}

func TestRemoveS3Permissions(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := &IAMManager{logger: logger}

	// Test deprecated method
	ctx := context.Background()
	err := manager.RemoveS3Permissions(ctx, "i-1234567890abcdef0", "us-east-1")

	// Should not return error (deprecated method does nothing)
	if err != nil {
		t.Errorf("RemoveS3Permissions should not return error: %v", err)
	}
}

func TestEmergencyCleanup(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := &IAMManager{logger: logger}

	// Test emergency cleanup
	ctx := context.Background()
	err := manager.EmergencyCleanup(ctx, "us-east-1")

	// Should not return error
	if err != nil {
		t.Errorf("EmergencyCleanup should not return error: %v", err)
	}
}

func TestS3PolicyDocumentEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		policy  S3PolicyDocument
		wantErr bool
	}{
		{
			name: "empty policy",
			policy: S3PolicyDocument{
				Version:   "",
				Statement: []Statement{},
			},
			wantErr: false,
		},
		{
			name: "policy with nil statement",
			policy: S3PolicyDocument{
				Version:   "2012-10-17",
				Statement: nil,
			},
			wantErr: false,
		},
		{
			name: "policy with empty statement fields",
			policy: S3PolicyDocument{
				Version: "2012-10-17",
				Statement: []Statement{
					{
						Effect:   "",
						Action:   []string{},
						Resource: "",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.policy)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantErr {
				// Should be able to unmarshal back
				var unmarshaled S3PolicyDocument
				if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
					t.Errorf("Failed to unmarshal: %v", err)
				}
			}
		})
	}
}

func TestGenerateUniqueIDRandomness(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := &IAMManager{logger: logger}

	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)
	numIDs := 100

	for i := 0; i < numIDs; i++ {
		id := manager.generateUniqueID()
		if ids[id] {
			t.Errorf("Generated duplicate ID: %s", id)
		}
		ids[id] = true

		// Verify ID format
		parts := strings.Split(id, "-")
		if len(parts) < 3 {
			t.Errorf("ID should have at least 3 parts: %s", id)
		}

		// Last part should be hex string (16 characters)
		hexPart := parts[len(parts)-1]
		if len(hexPart) != 16 {
			t.Errorf("Hex part should be 16 characters, got %d for ID: %s", len(hexPart), id)
		}

		// Verify hex characters
		for _, char := range hexPart {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
				t.Errorf("Invalid hex character %c in ID: %s", char, id)
			}
		}
	}

	if len(ids) != numIDs {
		t.Errorf("Expected %d unique IDs, got %d", numIDs, len(ids))
	}
}
