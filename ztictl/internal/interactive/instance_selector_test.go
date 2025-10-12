package interactive

import (
	"testing"
)

func TestInstanceStruct(t *testing.T) {
	instance := Instance{
		InstanceID:       "i-1234567890abcdef0",
		Name:             "test-instance",
		State:            "running",
		Platform:         "Linux",
		PrivateIPAddress: "10.0.1.100",
		PublicIPAddress:  "54.1.2.3",
		SSMStatus:        "Online",
		SSMAgentVersion:  "3.1.1732.0",
		LastPingDateTime: "2024-01-15T10:30:00Z",
		Tags: map[string]string{
			"Environment": "prod",
			"Team":        "platform",
		},
	}

	if instance.InstanceID != "i-1234567890abcdef0" {
		t.Errorf("InstanceID = %s, want i-1234567890abcdef0", instance.InstanceID)
	}

	if instance.Name != "test-instance" {
		t.Errorf("Name = %s, want test-instance", instance.Name)
	}

	if instance.State != "running" {
		t.Errorf("State = %s, want running", instance.State)
	}

	if instance.Platform != "Linux" {
		t.Errorf("Platform = %s, want Linux", instance.Platform)
	}

	if instance.PrivateIPAddress != "10.0.1.100" {
		t.Errorf("PrivateIPAddress = %s, want 10.0.1.100", instance.PrivateIPAddress)
	}

	if instance.PublicIPAddress != "54.1.2.3" {
		t.Errorf("PublicIPAddress = %s, want 54.1.2.3", instance.PublicIPAddress)
	}

	if instance.SSMStatus != "Online" {
		t.Errorf("SSMStatus = %s, want Online", instance.SSMStatus)
	}

	if instance.SSMAgentVersion != "3.1.1732.0" {
		t.Errorf("SSMAgentVersion = %s, want 3.1.1732.0", instance.SSMAgentVersion)
	}

	if instance.LastPingDateTime != "2024-01-15T10:30:00Z" {
		t.Errorf("LastPingDateTime = %s, want 2024-01-15T10:30:00Z", instance.LastPingDateTime)
	}

	if len(instance.Tags) != 2 {
		t.Errorf("Tags length = %d, want 2", len(instance.Tags))
	}

	if instance.Tags["Environment"] != "prod" {
		t.Errorf("Tags[Environment] = %s, want prod", instance.Tags["Environment"])
	}
}

func TestInstanceWithEmptyFields(t *testing.T) {
	instance := Instance{
		InstanceID: "i-1234567890abcdef0",
		State:      "running",
	}

	if instance.InstanceID != "i-1234567890abcdef0" {
		t.Errorf("InstanceID = %s, want i-1234567890abcdef0", instance.InstanceID)
	}

	if instance.State != "running" {
		t.Errorf("State = %s, want running", instance.State)
	}

	if instance.Name != "" {
		t.Errorf("Name should be empty, got %s", instance.Name)
	}

	if instance.PublicIPAddress != "" {
		t.Errorf("PublicIPAddress should be empty, got %s", instance.PublicIPAddress)
	}

	if instance.SSMStatus != "" {
		t.Errorf("SSMStatus should be empty, got %s", instance.SSMStatus)
	}
}

func TestInstanceWithoutPublicIP(t *testing.T) {
	instance := Instance{
		InstanceID:       "i-1234567890abcdef0",
		Name:             "private-instance",
		State:            "running",
		Platform:         "Linux",
		PrivateIPAddress: "10.0.1.100",
		SSMStatus:        "Online",
	}

	if instance.InstanceID != "i-1234567890abcdef0" {
		t.Errorf("InstanceID = %s, want i-1234567890abcdef0", instance.InstanceID)
	}

	if instance.Name != "private-instance" {
		t.Errorf("Name = %s, want private-instance", instance.Name)
	}

	if instance.State != "running" {
		t.Errorf("State = %s, want running", instance.State)
	}

	if instance.Platform != "Linux" {
		t.Errorf("Platform = %s, want Linux", instance.Platform)
	}

	if instance.SSMStatus != "Online" {
		t.Errorf("SSMStatus = %s, want Online", instance.SSMStatus)
	}

	if instance.PublicIPAddress != "" {
		t.Errorf("PublicIPAddress should be empty for private instance, got %s", instance.PublicIPAddress)
	}

	if instance.PrivateIPAddress != "10.0.1.100" {
		t.Errorf("PrivateIPAddress = %s, want 10.0.1.100", instance.PrivateIPAddress)
	}
}

func TestInstanceSSMStatusVariations(t *testing.T) {
	tests := []struct {
		name      string
		ssmStatus string
	}{
		{"Online status", "Online"},
		{"Connection lost", "ConnectionLost"},
		{"No agent", "No Agent"},
		{"Empty status", ""},
		{"Unknown status", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := Instance{
				InstanceID: "i-1234567890abcdef0",
				SSMStatus:  tt.ssmStatus,
			}

			if instance.InstanceID != "i-1234567890abcdef0" {
				t.Errorf("InstanceID = %s, want i-1234567890abcdef0", instance.InstanceID)
			}

			if instance.SSMStatus != tt.ssmStatus {
				t.Errorf("SSMStatus = %s, want %s", instance.SSMStatus, tt.ssmStatus)
			}
		})
	}
}

func TestInstancePlatformTypes(t *testing.T) {
	platforms := []string{"Linux", "Windows", "Windows Server 2022", "Red Hat Enterprise Linux", "Ubuntu"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			instance := Instance{
				InstanceID: "i-1234567890abcdef0",
				Platform:   platform,
			}

			if instance.InstanceID != "i-1234567890abcdef0" {
				t.Errorf("InstanceID = %s, want i-1234567890abcdef0", instance.InstanceID)
			}

			if instance.Platform != platform {
				t.Errorf("Platform = %s, want %s", instance.Platform, platform)
			}
		})
	}
}

func TestSelectInstanceEmptyList(t *testing.T) {
	instances := []Instance{}

	_, err := SelectInstance(instances, "Select an instance")
	if err == nil {
		t.Error("Expected error for empty instance list, got nil")
	}

	if err.Error() != "no instances available" {
		t.Errorf("Error message = %s, want 'no instances available'", err.Error())
	}
}

func TestFuzzyInstanceSelectorInterface(t *testing.T) {
	var _ InstanceSelector = (*FuzzyInstanceSelector)(nil)
}

func TestInstanceTagsMap(t *testing.T) {
	instance := Instance{
		InstanceID: "i-1234567890abcdef0",
		Tags: map[string]string{
			"Name":        "test-instance",
			"Environment": "prod",
			"Team":        "platform",
			"CostCenter":  "engineering",
		},
	}

	if instance.InstanceID != "i-1234567890abcdef0" {
		t.Errorf("InstanceID = %s, want i-1234567890abcdef0", instance.InstanceID)
	}

	if len(instance.Tags) != 4 {
		t.Errorf("Tags length = %d, want 4", len(instance.Tags))
	}

	expectedTags := map[string]string{
		"Name":        "test-instance",
		"Environment": "prod",
		"Team":        "platform",
		"CostCenter":  "engineering",
	}

	for key, expectedValue := range expectedTags {
		if value, exists := instance.Tags[key]; !exists {
			t.Errorf("Tag %s not found in instance tags", key)
		} else if value != expectedValue {
			t.Errorf("Tag %s = %s, want %s", key, value, expectedValue)
		}
	}
}

func TestInstanceWithNilTags(t *testing.T) {
	instance := Instance{
		InstanceID: "i-1234567890abcdef0",
		Tags:       nil,
	}

	if instance.InstanceID != "i-1234567890abcdef0" {
		t.Errorf("InstanceID = %s, want i-1234567890abcdef0", instance.InstanceID)
	}

	if instance.Tags != nil {
		t.Errorf("Tags should be nil, got %v", instance.Tags)
	}

	// Test that accessing nil map doesn't panic
	_, exists := instance.Tags["Name"]
	if exists {
		t.Error("Expected false for non-existent key in nil map")
	}
}
