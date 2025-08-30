package aws

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestInstanceStruct(t *testing.T) {
	now := time.Now()
	instance := Instance{
		ID:               "i-1234567890abcdef0",
		Name:             "test-instance",
		State:            "running",
		InstanceType:     "t3.micro",
		PrivateIP:        "10.0.1.100",
		PublicIP:         "203.0.113.1",
		Platform:         "Linux/UNIX",
		SSMStatus:        "Online",
		Tags:             map[string]string{"Name": "test-instance", "Environment": "test"},
		LaunchTime:       &now,
		AvailabilityZone: "us-east-1a",
		Region:           "us-east-1",
	}

	// Test all fields are properly set
	if instance.ID != "i-1234567890abcdef0" {
		t.Error("ID should be properly set")
	}

	if instance.Name != "test-instance" {
		t.Error("Name should be properly set")
	}

	if instance.State != "running" {
		t.Error("State should be properly set")
	}

	if instance.InstanceType != "t3.micro" {
		t.Error("InstanceType should be properly set")
	}

	if instance.PrivateIP != "10.0.1.100" {
		t.Error("PrivateIP should be properly set")
	}

	if instance.PublicIP != "203.0.113.1" {
		t.Error("PublicIP should be properly set")
	}

	if instance.Platform != "Linux/UNIX" {
		t.Error("Platform should be properly set")
	}

	if instance.SSMStatus != "Online" {
		t.Error("SSMStatus should be properly set")
	}

	if len(instance.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(instance.Tags))
	}

	if instance.Tags["Name"] != "test-instance" {
		t.Error("Name tag should be properly set")
	}

	if instance.Tags["Environment"] != "test" {
		t.Error("Environment tag should be properly set")
	}

	if instance.LaunchTime == nil || !instance.LaunchTime.Equal(now) {
		t.Error("LaunchTime should be properly set")
	}

	if instance.AvailabilityZone != "us-east-1a" {
		t.Error("AvailabilityZone should be properly set")
	}

	if instance.Region != "us-east-1" {
		t.Error("Region should be properly set")
	}
}

func TestInstanceJSON(t *testing.T) {
	now := time.Now()
	instance := Instance{
		ID:               "i-1234567890abcdef0",
		Name:             "test-instance",
		State:            "running",
		InstanceType:     "t3.micro",
		PrivateIP:        "10.0.1.100",
		PublicIP:         "203.0.113.1",
		Platform:         "Linux/UNIX",
		SSMStatus:        "Online",
		Tags:             map[string]string{"Name": "test-instance"},
		LaunchTime:       &now,
		AvailabilityZone: "us-east-1a",
		Region:           "us-east-1",
	}

	// Test JSON serialization
	jsonBytes, err := json.Marshal(instance)
	if err != nil {
		t.Fatalf("Failed to marshal Instance to JSON: %v", err)
	}

	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, "i-1234567890abcdef0") {
		t.Error("JSON should contain instance ID")
	}

	if !strings.Contains(jsonStr, "test-instance") {
		t.Error("JSON should contain instance name")
	}

	if !strings.Contains(jsonStr, "running") {
		t.Error("JSON should contain instance state")
	}

	// Test JSON deserialization
	var deserializedInstance Instance
	if err := json.Unmarshal(jsonBytes, &deserializedInstance); err != nil {
		t.Fatalf("Failed to unmarshal Instance from JSON: %v", err)
	}

	if deserializedInstance.ID != instance.ID {
		t.Error("Deserialized instance ID should match")
	}

	if deserializedInstance.Name != instance.Name {
		t.Error("Deserialized instance name should match")
	}

	if deserializedInstance.State != instance.State {
		t.Error("Deserialized instance state should match")
	}
}

func TestNewInstanceFromEC2(t *testing.T) {
	now := time.Now()

	ec2Instance := types.Instance{
		InstanceId:       aws.String("i-1234567890abcdef0"),
		State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
		InstanceType:     types.InstanceTypeT3Micro,
		PrivateIpAddress: aws.String("10.0.1.100"),
		PublicIpAddress:  aws.String("203.0.113.1"),
		LaunchTime:       &now,
		Placement:        &types.Placement{AvailabilityZone: aws.String("us-east-1a")},
		Tags: []types.Tag{
			{Key: aws.String("Name"), Value: aws.String("test-instance")},
			{Key: aws.String("Environment"), Value: aws.String("production")},
		},
	}

	instance := NewInstanceFromEC2(ec2Instance, "us-east-1")

	// Test basic fields
	if instance.ID != "i-1234567890abcdef0" {
		t.Errorf("Expected ID i-1234567890abcdef0, got %s", instance.ID)
	}

	if instance.State != "running" {
		t.Errorf("Expected state running, got %s", instance.State)
	}

	if instance.InstanceType != "t3.micro" {
		t.Errorf("Expected type t3.micro, got %s", instance.InstanceType)
	}

	if instance.PrivateIP != "10.0.1.100" {
		t.Errorf("Expected private IP 10.0.1.100, got %s", instance.PrivateIP)
	}

	if instance.PublicIP != "203.0.113.1" {
		t.Errorf("Expected public IP 203.0.113.1, got %s", instance.PublicIP)
	}

	if instance.Platform != "Linux/UNIX" {
		t.Errorf("Expected platform Linux/UNIX, got %s", instance.Platform)
	}

	if instance.Region != "us-east-1" {
		t.Errorf("Expected region us-east-1, got %s", instance.Region)
	}

	if instance.AvailabilityZone != "us-east-1a" {
		t.Errorf("Expected AZ us-east-1a, got %s", instance.AvailabilityZone)
	}

	// Test tags
	if len(instance.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(instance.Tags))
	}

	if instance.Tags["Name"] != "test-instance" {
		t.Error("Name tag should be set correctly")
	}

	if instance.Tags["Environment"] != "production" {
		t.Error("Environment tag should be set correctly")
	}

	// Test that name is extracted from tags
	if instance.Name != "test-instance" {
		t.Errorf("Expected name test-instance, got %s", instance.Name)
	}

	// Test launch time
	if instance.LaunchTime == nil || !instance.LaunchTime.Equal(now) {
		t.Error("Launch time should be set correctly")
	}
}

func TestNewInstanceFromEC2WithoutPublicIP(t *testing.T) {
	ec2Instance := types.Instance{
		InstanceId:       aws.String("i-1234567890abcdef0"),
		State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
		InstanceType:     types.InstanceTypeT3Micro,
		PrivateIpAddress: aws.String("10.0.1.100"),
		// PublicIpAddress is nil
		Placement: &types.Placement{AvailabilityZone: aws.String("us-east-1a")},
	}

	instance := NewInstanceFromEC2(ec2Instance, "us-east-1")

	if instance.PublicIP != "" {
		t.Errorf("Expected empty public IP, got %s", instance.PublicIP)
	}

	if instance.PrivateIP != "10.0.1.100" {
		t.Error("Private IP should still be set")
	}
}

func TestNewInstanceFromEC2WithoutName(t *testing.T) {
	ec2Instance := types.Instance{
		InstanceId:   aws.String("i-1234567890abcdef0"),
		State:        &types.InstanceState{Name: types.InstanceStateNameRunning},
		InstanceType: types.InstanceTypeT3Micro,
		Placement:    &types.Placement{AvailabilityZone: aws.String("us-east-1a")},
		// No Name tag
		Tags: []types.Tag{
			{Key: aws.String("Environment"), Value: aws.String("test")},
		},
	}

	instance := NewInstanceFromEC2(ec2Instance, "us-east-1")

	// Should fallback to instance ID as name
	if instance.Name != "i-1234567890abcdef0" {
		t.Errorf("Expected name to be instance ID, got %s", instance.Name)
	}
}

func TestGetPlatform(t *testing.T) {
	tests := []struct {
		name     string
		instance types.Instance
		expected string
	}{
		{
			name: "Windows platform",
			instance: types.Instance{
				Platform: types.PlatformValuesWindows,
			},
			expected: "Windows",
		},
		{
			name:     "No platform specified",
			instance: types.Instance{
				// Platform is empty
			},
			expected: "Linux/UNIX",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPlatform(tt.instance)
			if result != tt.expected {
				t.Errorf("Expected platform %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestExtractTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []types.Tag
		expected map[string]string
	}{
		{
			name: "valid tags",
			tags: []types.Tag{
				{Key: aws.String("Name"), Value: aws.String("test-instance")},
				{Key: aws.String("Environment"), Value: aws.String("production")},
			},
			expected: map[string]string{
				"Name":        "test-instance",
				"Environment": "production",
			},
		},
		{
			name:     "empty tags",
			tags:     []types.Tag{},
			expected: map[string]string{},
		},
		{
			name: "tags with nil key or value",
			tags: []types.Tag{
				{Key: aws.String("Name"), Value: aws.String("test")},
				{Key: nil, Value: aws.String("value")},
				{Key: aws.String("key"), Value: nil},
				{Key: nil, Value: nil},
			},
			expected: map[string]string{
				"Name": "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTags(tt.tags)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tags, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists || actualValue != expectedValue {
					t.Errorf("Expected tag %s=%s, got %s=%s", key, expectedValue, key, actualValue)
				}
			}
		})
	}
}

func TestSSOAccountStruct(t *testing.T) {
	account := SSOAccount{
		ID:           "123456789012",
		Name:         "Production Account",
		EmailAddress: "admin@example.com",
	}

	if account.ID != "123456789012" {
		t.Error("ID should be properly set")
	}

	if account.Name != "Production Account" {
		t.Error("Name should be properly set")
	}

	if account.EmailAddress != "admin@example.com" {
		t.Error("EmailAddress should be properly set")
	}

	// Test JSON serialization
	jsonBytes, err := json.Marshal(account)
	if err != nil {
		t.Fatalf("Failed to marshal SSOAccount to JSON: %v", err)
	}

	var deserializedAccount SSOAccount
	if err := json.Unmarshal(jsonBytes, &deserializedAccount); err != nil {
		t.Fatalf("Failed to unmarshal SSOAccount from JSON: %v", err)
	}

	if deserializedAccount.ID != account.ID {
		t.Error("Deserialized account ID should match")
	}
}

func TestSSORole(t *testing.T) {
	role := SSORole{
		Name:      "AdminRole",
		AccountID: "123456789012",
	}

	if role.Name != "AdminRole" {
		t.Error("Name should be properly set")
	}

	if role.AccountID != "123456789012" {
		t.Error("AccountID should be properly set")
	}

	// Test JSON serialization
	jsonBytes, err := json.Marshal(role)
	if err != nil {
		t.Fatalf("Failed to marshal SSORole to JSON: %v", err)
	}

	var deserializedRole SSORole
	if err := json.Unmarshal(jsonBytes, &deserializedRole); err != nil {
		t.Fatalf("Failed to unmarshal SSORole from JSON: %v", err)
	}

	if deserializedRole.Name != role.Name {
		t.Error("Deserialized role name should match")
	}
}

func TestCommandExecution(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Second)

	execution := CommandExecution{
		CommandID:  "12345678-1234-1234-1234-123456789012",
		InstanceID: "i-1234567890abcdef0",
		Status:     "Success",
		Output:     "Hello, World!\n",
		Error:      "",
		ExitCode:   0,
		StartTime:  &startTime,
		EndTime:    &endTime,
	}

	// Test all fields
	if execution.CommandID != "12345678-1234-1234-1234-123456789012" {
		t.Error("CommandID should be properly set")
	}

	if execution.InstanceID != "i-1234567890abcdef0" {
		t.Error("InstanceID should be properly set")
	}

	if execution.Status != "Success" {
		t.Error("Status should be properly set")
	}

	if execution.Output != "Hello, World!\n" {
		t.Error("Output should be properly set")
	}

	if execution.Error != "" {
		t.Error("Error should be empty")
	}

	if execution.ExitCode != 0 {
		t.Error("ExitCode should be 0")
	}

	if execution.StartTime == nil || !execution.StartTime.Equal(startTime) {
		t.Error("StartTime should be properly set")
	}

	if execution.EndTime == nil || !execution.EndTime.Equal(endTime) {
		t.Error("EndTime should be properly set")
	}

	// Test JSON serialization
	jsonBytes, err := json.Marshal(execution)
	if err != nil {
		t.Fatalf("Failed to marshal CommandExecution to JSON: %v", err)
	}

	var deserializedExecution CommandExecution
	if err := json.Unmarshal(jsonBytes, &deserializedExecution); err != nil {
		t.Fatalf("Failed to unmarshal CommandExecution from JSON: %v", err)
	}

	if deserializedExecution.CommandID != execution.CommandID {
		t.Error("Deserialized command ID should match")
	}
}

func TestFileTransfer(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Minute)

	transfer := FileTransfer{
		ID:           "transfer-123",
		SourcePath:   "/local/file.txt",
		TargetPath:   "/remote/file.txt",
		InstanceID:   "i-1234567890abcdef0",
		Region:       "us-east-1",
		Size:         1024,
		Method:       "s3",
		Status:       "completed",
		StartTime:    &startTime,
		EndTime:      &endTime,
		ErrorMessage: "",
	}

	// Test all fields
	if transfer.ID != "transfer-123" {
		t.Error("ID should be properly set")
	}

	if transfer.SourcePath != "/local/file.txt" {
		t.Error("SourcePath should be properly set")
	}

	if transfer.TargetPath != "/remote/file.txt" {
		t.Error("TargetPath should be properly set")
	}

	if transfer.InstanceID != "i-1234567890abcdef0" {
		t.Error("InstanceID should be properly set")
	}

	if transfer.Region != "us-east-1" {
		t.Error("Region should be properly set")
	}

	if transfer.Size != 1024 {
		t.Error("Size should be properly set")
	}

	if transfer.Method != "s3" {
		t.Error("Method should be properly set")
	}

	if transfer.Status != "completed" {
		t.Error("Status should be properly set")
	}

	if transfer.StartTime == nil || !transfer.StartTime.Equal(startTime) {
		t.Error("StartTime should be properly set")
	}

	if transfer.EndTime == nil || !transfer.EndTime.Equal(endTime) {
		t.Error("EndTime should be properly set")
	}

	if transfer.ErrorMessage != "" {
		t.Error("ErrorMessage should be empty")
	}

	// Test JSON serialization
	jsonBytes, err := json.Marshal(transfer)
	if err != nil {
		t.Fatalf("Failed to marshal FileTransfer to JSON: %v", err)
	}

	var deserializedTransfer FileTransfer
	if err := json.Unmarshal(jsonBytes, &deserializedTransfer); err != nil {
		t.Fatalf("Failed to unmarshal FileTransfer from JSON: %v", err)
	}

	if deserializedTransfer.ID != transfer.ID {
		t.Error("Deserialized transfer ID should match")
	}
}

func TestSSMInstanceInfo(t *testing.T) {
	info := SSMInstanceInfo{
		InstanceID:       "i-1234567890abcdef0",
		PingStatus:       "Online",
		LastPingDateTime: "2023-01-01T12:00:00Z",
		AgentVersion:     "3.1.1732.0",
		IsLatestVersion:  true,
		PlatformType:     "Linux",
		PlatformName:     "Amazon Linux",
		PlatformVersion:  "2",
		ResourceType:     "EC2Instance",
	}

	// Test all fields
	if info.InstanceID != "i-1234567890abcdef0" {
		t.Error("InstanceID should be properly set")
	}

	if info.PingStatus != "Online" {
		t.Error("PingStatus should be properly set")
	}

	if info.LastPingDateTime != "2023-01-01T12:00:00Z" {
		t.Error("LastPingDateTime should be properly set")
	}

	if info.AgentVersion != "3.1.1732.0" {
		t.Error("AgentVersion should be properly set")
	}

	if !info.IsLatestVersion {
		t.Error("IsLatestVersion should be true")
	}

	if info.PlatformType != "Linux" {
		t.Error("PlatformType should be properly set")
	}

	if info.PlatformName != "Amazon Linux" {
		t.Error("PlatformName should be properly set")
	}

	if info.PlatformVersion != "2" {
		t.Error("PlatformVersion should be properly set")
	}

	if info.ResourceType != "EC2Instance" {
		t.Error("ResourceType should be properly set")
	}

	// Test JSON serialization
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal SSMInstanceInfo to JSON: %v", err)
	}

	var deserializedInfo SSMInstanceInfo
	if err := json.Unmarshal(jsonBytes, &deserializedInfo); err != nil {
		t.Fatalf("Failed to unmarshal SSMInstanceInfo from JSON: %v", err)
	}

	if deserializedInfo.InstanceID != info.InstanceID {
		t.Error("Deserialized instance ID should match")
	}
}

func TestSSMCommandStatus(t *testing.T) {
	// Test all status constants
	if SSMCommandStatusPending != "Pending" {
		t.Error("SSMCommandStatusPending should be 'Pending'")
	}

	if SSMCommandStatusInProgress != "InProgress" {
		t.Error("SSMCommandStatusInProgress should be 'InProgress'")
	}

	if SSMCommandStatusSuccess != "Success" {
		t.Error("SSMCommandStatusSuccess should be 'Success'")
	}

	if SSMCommandStatusCancelled != "Cancelled" {
		t.Error("SSMCommandStatusCancelled should be 'Cancelled'")
	}

	if SSMCommandStatusFailed != "Failed" {
		t.Error("SSMCommandStatusFailed should be 'Failed'")
	}

	if SSMCommandStatusTimedOut != "TimedOut" {
		t.Error("SSMCommandStatusTimedOut should be 'TimedOut'")
	}

	if SSMCommandStatusCancelling != "Cancelling" {
		t.Error("SSMCommandStatusCancelling should be 'Cancelling'")
	}
}

func TestSSMCommandStatusIsComplete(t *testing.T) {
	tests := []struct {
		status   SSMCommandStatus
		complete bool
	}{
		{SSMCommandStatusSuccess, true},
		{SSMCommandStatusFailed, true},
		{SSMCommandStatusCancelled, true},
		{SSMCommandStatusTimedOut, true},
		{SSMCommandStatusPending, false},
		{SSMCommandStatusInProgress, false},
		{SSMCommandStatusCancelling, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsComplete() != tt.complete {
				t.Errorf("Expected IsComplete() to return %v for status %s", tt.complete, tt.status)
			}
		})
	}
}

func TestSSMCommandStatusIsRunning(t *testing.T) {
	tests := []struct {
		status  SSMCommandStatus
		running bool
	}{
		{SSMCommandStatusPending, true},
		{SSMCommandStatusInProgress, true},
		{SSMCommandStatusSuccess, false},
		{SSMCommandStatusFailed, false},
		{SSMCommandStatusCancelled, false},
		{SSMCommandStatusTimedOut, false},
		{SSMCommandStatusCancelling, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsRunning() != tt.running {
				t.Errorf("Expected IsRunning() to return %v for status %s", tt.running, tt.status)
			}
		})
	}
}

func TestInstanceWithEmptyValues(t *testing.T) {
	// Test instance with minimal/empty values
	instance := Instance{
		ID: "i-1234567890abcdef0",
		// All other fields empty/nil
	}

	if instance.ID != "i-1234567890abcdef0" {
		t.Error("ID should be set even when other fields are empty")
	}

	if instance.Name != "" {
		t.Error("Name should be empty")
	}

	if instance.Tags != nil {
		t.Error("Tags should be nil (not initialized)")
	}

	if instance.LaunchTime != nil {
		t.Error("LaunchTime should be nil")
	}
}

func TestStructsWithNilPointers(t *testing.T) {
	// Test that structs handle nil pointers gracefully

	// CommandExecution with nil times
	execution := CommandExecution{
		CommandID:  "test-123",
		InstanceID: "i-1234567890abcdef0",
		Status:     "Success",
		StartTime:  nil,
		EndTime:    nil,
	}

	if execution.StartTime != nil {
		t.Error("StartTime should be nil")
	}

	if execution.EndTime != nil {
		t.Error("EndTime should be nil")
	}

	// FileTransfer with nil times
	transfer := FileTransfer{
		ID:        "transfer-123",
		StartTime: nil,
		EndTime:   nil,
	}

	if transfer.StartTime != nil {
		t.Error("StartTime should be nil")
	}

	if transfer.EndTime != nil {
		t.Error("EndTime should be nil")
	}

	// Instance with nil launch time
	instance := Instance{
		ID:         "i-1234567890abcdef0",
		LaunchTime: nil,
	}

	if instance.LaunchTime != nil {
		t.Error("LaunchTime should be nil")
	}
}
