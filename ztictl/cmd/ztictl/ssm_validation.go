package main

import (
	"context"
	"fmt"

	"ztictl/internal/interactive"
	"ztictl/internal/ssm"
	awsservice "ztictl/pkg/aws"
	"ztictl/pkg/colors"
)

// InstanceValidationRequirements defines the validation criteria for an instance
type InstanceValidationRequirements struct {
	// AllowedStates defines which EC2 states are acceptable for the operation
	// Common values: ["running"], ["stopped"], ["running", "stopped"]
	AllowedStates []string

	// RequireSSMOnline indicates whether the SSM agent must be online
	// Set to true for connect/exec operations, false for power operations
	RequireSSMOnline bool

	// Operation is the name of the operation being performed (used in error messages)
	// Examples: "connect", "execute commands", "start", "stop", "reboot"
	Operation string
}

// ValidateInstanceState validates that an instance meets the requirements for an operation
// It fetches instance details and checks both EC2 state and SSM agent status
func ValidateInstanceState(ctx context.Context, ssmManager *ssm.Manager, instanceID, region string, requirements InstanceValidationRequirements) error {
	// Get instance details
	instances, err := ssmManager.GetInstanceService().ListInstances(ctx, region, &awsservice.ListFilters{})
	if err != nil {
		return fmt.Errorf("failed to fetch instance details: %w", err)
	}

	// Find the specific instance
	var targetInstance *interactive.Instance
	for i := range instances {
		if instances[i].InstanceID == instanceID {
			targetInstance = &instances[i]
			break
		}
	}

	if targetInstance == nil {
		return fmt.Errorf("instance %s not found in region %s", instanceID, region)
	}

	// Validate EC2 instance state
	if err := validateEC2State(targetInstance, region, requirements); err != nil {
		return err
	}

	// Validate SSM agent status if required
	if requirements.RequireSSMOnline {
		if err := validateSSMStatus(targetInstance, requirements.Operation); err != nil {
			return err
		}
	}

	return nil
}

// validateEC2State checks if the instance is in an allowed EC2 state
func validateEC2State(instance *interactive.Instance, region string, requirements InstanceValidationRequirements) error {
	// Check if instance is in an acceptable state
	validState := false
	for _, allowedState := range requirements.AllowedStates {
		if instance.State == allowedState {
			validState = true
			break
		}
	}

	if !validState {
		// Display error header
		colors.PrintError("\n✗ Cannot %s - Instance is not in required state\n", requirements.Operation)
		fmt.Printf("\n")

		// Display instance details
		colors.PrintWarning("Instance Details:\n")
		fmt.Printf("  Instance ID: %s\n", colors.ColorData("%s", instance.InstanceID))
		fmt.Printf("  Name:        %s\n", colors.ColorData("%s", instance.Name))
		fmt.Printf("  State:       %s\n", getInstanceStateColor(instance.State))
		fmt.Printf("  Required:    %v\n", requirements.AllowedStates)
		fmt.Printf("\n")

		// Provide helpful suggestions based on current state
		displayStateSuggestion(instance.State, region, instance.InstanceID, requirements)
		fmt.Printf("\n")

		return fmt.Errorf("instance is in '%s' state, expected one of: %v", instance.State, requirements.AllowedStates)
	}

	return nil
}

// validateSSMStatus checks if the SSM agent is online
func validateSSMStatus(instance *interactive.Instance, operation string) error {
	if instance.SSMStatus != "Online" {
		colors.PrintWarning("\n⚠ Warning: SSM Agent is not online\n")
		fmt.Printf("\n")
		fmt.Printf("  Instance ID:  %s\n", colors.ColorData("%s", instance.InstanceID))
		fmt.Printf("  Name:         %s\n", colors.ColorData("%s", instance.Name))
		fmt.Printf("  State:        %s\n", getInstanceStateColor(instance.State))
		fmt.Printf("  SSM Status:   %s\n", getSSMStatusColor(instance.SSMStatus))
		fmt.Printf("\n")

		colors.PrintData("Possible reasons:\n")
		fmt.Printf("  1. SSM Agent not installed or not running\n")
		fmt.Printf("  2. Instance doesn't have required IAM role (AmazonSSMManagedInstanceCore)\n")
		fmt.Printf("  3. Network connectivity issues to SSM endpoints\n")
		fmt.Printf("  4. Instance recently started (agent may still be initializing)\n")
		fmt.Printf("\n")

		if instance.SSMStatus == "ConnectionLost" {
			colors.PrintWarning("The agent was previously online but connection was lost.\n")
			fmt.Printf("This may be temporary. You can try to proceed, but connection may fail.\n\n")
		} else {
			colors.PrintError("Cannot %s - SSM Agent must be 'Online' to establish connection.\n\n", operation)
			return fmt.Errorf("SSM agent is '%s', expected 'Online'", instance.SSMStatus)
		}
	}

	return nil
}

// displayStateSuggestion provides helpful suggestions based on the instance's current state
func displayStateSuggestion(state, region, instanceID string, requirements InstanceValidationRequirements) {
	switch state {
	case "stopped":
		if contains(requirements.AllowedStates, "running") {
			colors.PrintData("💡 Tip: Start the instance first:\n")
			fmt.Printf("   ztictl ssm start %s --region %s\n", instanceID, region)
		}
	case "running":
		if contains(requirements.AllowedStates, "stopped") {
			colors.PrintData("💡 Tip: Instance is already running. Use 'reboot' to restart it:\n")
			fmt.Printf("   ztictl ssm reboot %s --region %s\n", instanceID, region)
		}
	case "stopping":
		colors.PrintWarning("⏳ Instance is currently stopping. Wait for it to stop, then start it.\n")
	case "pending":
		colors.PrintWarning("⏳ Instance is starting. Please wait a moment and try again.\n")
	case "terminated":
		colors.PrintError("✗ Instance has been terminated. Cannot perform operations on terminated instances.\n")
	case "shutting-down":
		colors.PrintWarning("⏳ Instance is shutting down. It will soon be terminated.\n")
	default:
		colors.PrintWarning("⚠ Instance is in '%s' state.\n", state)
	}
}

// getInstanceStateColor returns a colored state string with appropriate visual indicator
func getInstanceStateColor(state string) string {
	switch state {
	case "running":
		return colors.ColorSuccess("● %s", "running")
	case "stopped":
		return colors.ColorError("○ %s", "stopped")
	case "stopping":
		return colors.ColorWarning("◑ %s", "stopping")
	case "pending":
		return colors.ColorWarning("◐ %s", "pending")
	case "terminated":
		return colors.ColorError("✗ %s", "terminated")
	case "shutting-down":
		return colors.ColorWarning("◑ %s", "shutting-down")
	default:
		return colors.ColorData("%s", state)
	}
}

// getSSMStatusColor returns a colored SSM status string with appropriate visual indicator
func getSSMStatusColor(status string) string {
	switch status {
	case "Online":
		return colors.ColorSuccess("✓ %s", "Online")
	case "ConnectionLost":
		return colors.ColorWarning("⚠ %s", "Connection Lost")
	case "No Agent":
		return colors.ColorError("✗ %s", "No Agent")
	default:
		if status == "" {
			return colors.ColorError("✗ %s", "No Agent")
		}
		return colors.ColorWarning("? %s", status)
	}
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
