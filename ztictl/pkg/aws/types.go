package aws

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Instance represents an EC2 instance with relevant metadata
type Instance struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	State            string            `json:"state"`
	InstanceType     string            `json:"instance_type"`
	PrivateIP        string            `json:"private_ip"`
	PublicIP         string            `json:"public_ip,omitempty"`
	Platform         string            `json:"platform"`
	SSMStatus        string            `json:"ssm_status"`
	Tags             map[string]string `json:"tags"`
	LaunchTime       *time.Time        `json:"launch_time,omitempty"`
	AvailabilityZone string            `json:"availability_zone"`
	Region           string            `json:"region"`
}

// NewInstanceFromEC2 creates an Instance from an EC2 instance
func NewInstanceFromEC2(ec2Instance types.Instance, region string) *Instance {
	instance := &Instance{
		ID:               *ec2Instance.InstanceId,
		State:            string(ec2Instance.State.Name),
		InstanceType:     string(ec2Instance.InstanceType),
		Platform:         getPlatform(ec2Instance),
		Tags:             extractTags(ec2Instance.Tags),
		LaunchTime:       ec2Instance.LaunchTime,
		AvailabilityZone: *ec2Instance.Placement.AvailabilityZone,
		Region:           region,
	}

	// Set IP addresses
	if ec2Instance.PrivateIpAddress != nil {
		instance.PrivateIP = *ec2Instance.PrivateIpAddress
	}
	if ec2Instance.PublicIpAddress != nil {
		instance.PublicIP = *ec2Instance.PublicIpAddress
	}

	// Get name from tags
	if name, exists := instance.Tags["Name"]; exists {
		instance.Name = name
	} else {
		instance.Name = instance.ID
	}

	return instance
}

// SSOAccount represents an AWS SSO account
type SSOAccount struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	EmailAddress string `json:"email_address,omitempty"`
}

// SSORole represents an AWS SSO role
type SSORole struct {
	Name      string `json:"name"`
	AccountID string `json:"account_id"`
}

// CommandExecution represents the execution of an SSM command
type CommandExecution struct {
	CommandID  string                 `json:"command_id"`
	InstanceID string                 `json:"instance_id"`
	Status     string                 `json:"status"`
	Output     string                 `json:"output,omitempty"`
	Error      string                 `json:"error,omitempty"`
	ExitCode   int32                  `json:"exit_code,omitempty"`
	StartTime  *time.Time             `json:"start_time,omitempty"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
}

// FileTransfer represents a file transfer operation
type FileTransfer struct {
	ID           string     `json:"id"`
	SourcePath   string     `json:"source_path"`
	TargetPath   string     `json:"target_path"`
	InstanceID   string     `json:"instance_id"`
	Region       string     `json:"region"`
	Size         int64      `json:"size"`
	Method       string     `json:"method"` // "direct" or "s3"
	Status       string     `json:"status"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

// SSMInstanceInfo represents SSM-specific instance information
type SSMInstanceInfo struct {
	InstanceID       string `json:"instance_id"`
	PingStatus       string `json:"ping_status"`
	LastPingDateTime string `json:"last_ping_date_time"`
	AgentVersion     string `json:"agent_version"`
	IsLatestVersion  bool   `json:"is_latest_version"`
	PlatformType     string `json:"platform_type"`
	PlatformName     string `json:"platform_name"`
	PlatformVersion  string `json:"platform_version"`
	ResourceType     string `json:"resource_type"`
}

// getPlatform determines the platform from EC2 instance data
func getPlatform(instance types.Instance) string {
	if instance.Platform != "" {
		return string(instance.Platform)
	}
	
	// Default to Linux if platform is not specified
	return "Linux/UNIX"
}

// extractTags converts EC2 tags to a map
func extractTags(tags []types.Tag) map[string]string {
	tagMap := make(map[string]string)
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			tagMap[*tag.Key] = *tag.Value
		}
	}
	return tagMap
}

// SSMCommandStatus represents the status of an SSM command
type SSMCommandStatus string

const (
	SSMCommandStatusPending    SSMCommandStatus = "Pending"
	SSMCommandStatusInProgress SSMCommandStatus = "InProgress"
	SSMCommandStatusSuccess    SSMCommandStatus = "Success"
	SSMCommandStatusCancelled  SSMCommandStatus = "Cancelled"
	SSMCommandStatusFailed     SSMCommandStatus = "Failed"
	SSMCommandStatusTimedOut   SSMCommandStatus = "TimedOut"
	SSMCommandStatusCancelling SSMCommandStatus = "Cancelling"
)

// IsComplete returns true if the command has completed (successfully or not)
func (s SSMCommandStatus) IsComplete() bool {
	return s == SSMCommandStatusSuccess ||
		s == SSMCommandStatusFailed ||
		s == SSMCommandStatusCancelled ||
		s == SSMCommandStatusTimedOut
}

// IsRunning returns true if the command is currently running
func (s SSMCommandStatus) IsRunning() bool {
	return s == SSMCommandStatusPending || s == SSMCommandStatusInProgress
}
