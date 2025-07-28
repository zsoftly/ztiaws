package ssm

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	appconfig "ztictl/internal/config"
	"ztictl/internal/logging"
	"ztictl/pkg/errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Manager handles AWS Systems Manager operations
type Manager struct {
	logger             *logging.Logger
	iamManager         *IAMManager
	s3LifecycleManager *S3LifecycleManager
}

// Instance represents an EC2 instance with SSM information
type Instance struct {
	InstanceID       string            `json:"instance_id"`
	Name             string            `json:"name"`
	State            string            `json:"state"`
	Platform         string            `json:"platform"`
	PrivateIPAddress string            `json:"private_ip_address"`
	PublicIPAddress  string            `json:"public_ip_address,omitempty"`
	SSMStatus        string            `json:"ssm_status"`
	SSMAgentVersion  string            `json:"ssm_agent_version,omitempty"`
	LastPingDateTime string            `json:"last_ping_date_time,omitempty"`
	Tags             map[string]string `json:"tags,omitempty"`
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	InstanceID    string         `json:"instance_id"`
	Command       string         `json:"command"`
	Status        string         `json:"status"`
	ExitCode      *int32         `json:"exit_code,omitempty"`
	Output        string         `json:"output"`
	ErrorOutput   string         `json:"error_output,omitempty"`
	ExecutionTime *time.Duration `json:"execution_time,omitempty"`
}

// ListFilters represents filters for listing instances
type ListFilters struct {
	Tag    string `json:"tag,omitempty"`    // Format: key=value
	Status string `json:"status,omitempty"` // Instance state
	Name   string `json:"name,omitempty"`   // Name pattern
}

// FileTransferOperation represents a file transfer operation
type FileTransferOperation struct {
	InstanceID   string     `json:"instance_id"`
	Region       string     `json:"region"`
	LocalPath    string     `json:"local_path"`
	RemotePath   string     `json:"remote_path"`
	Size         int64      `json:"size"`
	Method       string     `json:"method"` // "direct" or "s3"
	Status       string     `json:"status"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

// NewManager creates a new SSM manager
func NewManager(logger *logging.Logger) *Manager {
	return &Manager{
		logger: logger,
	}
}

// getAWSCommand returns the platform-appropriate AWS CLI command name
func getAWSCommand() string {
	if runtime.GOOS == "windows" {
		return "aws.exe"
	}
	return "aws"
}

// initializeManagers initializes the IAM and S3 lifecycle managers
func (m *Manager) initializeManagers(ctx context.Context, region string) error {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %w", err)
	}

	// Create AWS service clients
	iamClient := iam.NewFromConfig(cfg)
	ec2Client := ec2.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg)
	stsClient := sts.NewFromConfig(cfg)

	// Initialize IAM manager
	iamManager, err := NewIAMManager(m.logger, iamClient, ec2Client)
	if err != nil {
		return fmt.Errorf("failed to create IAM manager: %w", err)
	}
	m.iamManager = iamManager

	// Initialize S3 lifecycle manager
	m.s3LifecycleManager = NewS3LifecycleManager(m.logger, s3Client, stsClient)

	return nil
}

// StartSession starts an SSM session to an instance
func (m *Manager) StartSession(ctx context.Context, instanceIdentifier, region string) error {
	// Resolve instance identifier to instance ID
	instanceID, err := m.resolveInstanceIdentifier(ctx, instanceIdentifier, region)
	if err != nil {
		return fmt.Errorf("failed to resolve instance: %w", err)
	}

	m.logger.Info("Starting SSM session", "instance", instanceID, "region", region)

	// Use AWS CLI for session manager (Go SDK doesn't support interactive sessions)
	cmd := exec.CommandContext(ctx, getAWSCommand(), "ssm", "start-session",
		"--region", region, "--target", instanceID)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.NewSSMError("failed to start session", err)
	}

	return nil
}

// ListInstances lists SSM-enabled instances in a region
func (m *Manager) ListInstances(ctx context.Context, region string, filters *ListFilters) ([]Instance, error) {
	m.logger.Debug("Listing SSM instances", "region", region)

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, errors.NewAWSError("failed to load AWS config", err)
	}

	// Get SSM instance information
	ssmClient := ssm.NewFromConfig(awsCfg)
	ssmInstances, err := m.getSSMInstances(ctx, ssmClient, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSM instances: %w", err)
	}

	if len(ssmInstances) == 0 {
		return []Instance{}, nil
	}

	// Get EC2 instance details
	ec2Client := ec2.NewFromConfig(awsCfg)
	instances, err := m.enrichWithEC2Data(ctx, ec2Client, ssmInstances)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich with EC2 data: %w", err)
	}

	return instances, nil
}

// ExecuteCommand executes a command on an instance via SSM
func (m *Manager) ExecuteCommand(ctx context.Context, instanceIdentifier, region, command, comment string) (*CommandResult, error) {
	// Resolve instance identifier
	instanceID, err := m.resolveInstanceIdentifier(ctx, instanceIdentifier, region)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve instance: %w", err)
	}

	m.logger.Info("Executing command", "instance", instanceID, "command", command)

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, errors.NewAWSError("failed to load AWS config", err)
	}

	ssmClient := ssm.NewFromConfig(awsCfg)

	// Send command
	if comment == "" {
		comment = "Command executed via ztictl"
	}

	startTime := time.Now()

	sendResp, err := ssmClient.SendCommand(ctx, &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []string{instanceID},
		Parameters: map[string][]string{
			"commands": {command},
		},
		Comment: aws.String(comment),
	})
	if err != nil {
		return nil, errors.NewSSMError("failed to send command", err)
	}

	commandID := aws.ToString(sendResp.Command.CommandId)
	m.logger.Debug("Command sent", "command_id", commandID)

	// Wait for command completion
	result, err := m.waitForCommandCompletion(ctx, ssmClient, commandID, instanceID, region)
	if err != nil {
		return nil, err
	}

	executionTime := time.Since(startTime)
	result.ExecutionTime = &executionTime
	result.Command = command

	return result, nil
}

// UploadFile uploads a file to an instance via SSM
func (m *Manager) UploadFile(ctx context.Context, instanceIdentifier, region, localPath, remotePath string) error {
	// Resolve instance identifier
	instanceID, err := m.resolveInstanceIdentifier(ctx, instanceIdentifier, region)
	if err != nil {
		return fmt.Errorf("failed to resolve instance: %w", err)
	}

	// Check if local file exists
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("local file not found: %w", err)
	}

	cfg := appconfig.Get()

	m.logger.Info("Uploading file", "instance", instanceID, "local", localPath, "remote", remotePath, "size", fileInfo.Size())

	// Choose transfer method based on file size
	if fileInfo.Size() < cfg.System.FileSizeThreshold {
		return m.uploadFileSmall(ctx, instanceID, region, localPath, remotePath)
	} else {
		return m.uploadFileLarge(ctx, instanceID, region, localPath, remotePath)
	}
}

// DownloadFile downloads a file from an instance via SSM
func (m *Manager) DownloadFile(ctx context.Context, instanceIdentifier, region, remotePath, localPath string) error {
	// Resolve instance identifier
	instanceID, err := m.resolveInstanceIdentifier(ctx, instanceIdentifier, region)
	if err != nil {
		return fmt.Errorf("failed to resolve instance: %w", err)
	}

	m.logger.Info("Downloading file", "instance", instanceID, "remote", remotePath, "local", localPath)

	// First, get file size to determine transfer method
	fileSize, err := m.getRemoteFileSize(ctx, instanceID, region, remotePath)
	if err != nil {
		return fmt.Errorf("failed to get remote file size: %w", err)
	}

	cfg := appconfig.Get()

	// Choose transfer method based on file size
	if fileSize < cfg.System.FileSizeThreshold {
		return m.downloadFileSmall(ctx, instanceID, region, remotePath, localPath)
	} else {
		return m.downloadFileLarge(ctx, instanceID, region, remotePath, localPath)
	}
}

// ForwardPort sets up port forwarding through SSM
func (m *Manager) ForwardPort(ctx context.Context, instanceIdentifier, region string, localPort, remotePort int) error {
	// Resolve instance identifier
	instanceID, err := m.resolveInstanceIdentifier(ctx, instanceIdentifier, region)
	if err != nil {
		return fmt.Errorf("failed to resolve instance: %w", err)
	}

	m.logger.Info("Starting port forwarding", "instance", instanceID, "local_port", localPort, "remote_port", remotePort)

	// Use AWS CLI for port forwarding (Go SDK doesn't support this directly)
	cmd := exec.CommandContext(ctx, getAWSCommand(), "ssm", "start-session",
		"--region", region,
		"--target", instanceID,
		"--document-name", "AWS-StartPortForwardingSession",
		"--parameters", fmt.Sprintf(`{"portNumber":["%d"],"localPortNumber":["%d"]}`, remotePort, localPort))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Port forwarding: localhost:%d -> %s:%d\n", localPort, instanceID, remotePort)
	fmt.Printf("Press Ctrl+C to stop port forwarding\n\n")

	if err := cmd.Run(); err != nil {
		return errors.NewSSMError("failed to start port forwarding", err)
	}

	return nil
}

// GetInstanceStatus gets SSM status for a specific instance
func (m *Manager) GetInstanceStatus(ctx context.Context, instanceIdentifier, region string) (*Instance, error) {
	// Resolve instance identifier
	instanceID, err := m.resolveInstanceIdentifier(ctx, instanceIdentifier, region)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve instance: %w", err)
	}

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, errors.NewAWSError("failed to load AWS config", err)
	}

	ssmClient := ssm.NewFromConfig(awsCfg)

	// Get SSM instance information
	resp, err := ssmClient.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{
		Filters: []ssmtypes.InstanceInformationStringFilter{
			{
				Key:    aws.String("InstanceIds"),
				Values: []string{instanceID},
			},
		},
	})
	if err != nil {
		return nil, errors.NewSSMError("failed to describe instance information", err)
	}

	if len(resp.InstanceInformationList) == 0 {
		return nil, fmt.Errorf("instance %s not found in SSM", instanceID)
	}

	info := resp.InstanceInformationList[0]

	return &Instance{
		InstanceID:       aws.ToString(info.InstanceId),
		SSMStatus:        string(info.PingStatus),
		SSMAgentVersion:  aws.ToString(info.AgentVersion),
		LastPingDateTime: info.LastPingDateTime.Format(time.RFC3339),
		Platform:         aws.ToString(info.PlatformName),
	}, nil
}

// ListInstanceStatuses lists SSM status for all instances in a region
func (m *Manager) ListInstanceStatuses(ctx context.Context, region string) ([]Instance, error) {
	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, errors.NewAWSError("failed to load AWS config", err)
	}

	ssmClient := ssm.NewFromConfig(awsCfg)

	// Get all SSM instances
	resp, err := ssmClient.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{})
	if err != nil {
		return nil, errors.NewSSMError("failed to describe instance information", err)
	}

	instances := make([]Instance, len(resp.InstanceInformationList))
	for i, info := range resp.InstanceInformationList {
		instances[i] = Instance{
			InstanceID:       aws.ToString(info.InstanceId),
			SSMStatus:        string(info.PingStatus),
			SSMAgentVersion:  aws.ToString(info.AgentVersion),
			LastPingDateTime: info.LastPingDateTime.Format(time.RFC3339),
			Platform:         aws.ToString(info.PlatformName),
		}
	}

	return instances, nil
}

// Helper methods

// resolveInstanceIdentifier resolves an instance name or ID to an instance ID
func (m *Manager) resolveInstanceIdentifier(ctx context.Context, identifier, region string) (string, error) {
	// If it's already an instance ID, validate and return it
	if strings.HasPrefix(identifier, "i-") && len(identifier) >= 10 {
		// Validate the instance exists
		if err := m.validateInstanceID(ctx, identifier, region); err != nil {
			return "", err
		}
		return identifier, nil
	}

	// Search by name tag
	instanceID, err := m.findInstanceByName(ctx, identifier, region)
	if err != nil {
		return "", err
	}

	m.logger.Info("Resolved instance", "name", identifier, "instance_id", instanceID)
	return instanceID, nil
}

// validateInstanceID validates that an instance ID exists
func (m *Manager) validateInstanceID(ctx context.Context, instanceID, region string) error {
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return err
	}

	ec2Client := ec2.NewFromConfig(awsCfg)

	_, err = ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})

	if err != nil {
		return fmt.Errorf("instance ID '%s' not found in region '%s'", instanceID, region)
	}

	return nil
}

// findInstanceByName finds an instance by its Name tag
func (m *Manager) findInstanceByName(ctx context.Context, name, region string) (string, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return "", err
	}

	ec2Client := ec2.NewFromConfig(awsCfg)

	resp, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{name},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running", "stopped"},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to search for instance: %w", err)
	}

	var instances []string
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, aws.ToString(instance.InstanceId))
		}
	}

	if len(instances) == 0 {
		return "", fmt.Errorf("no instance found with name '%s'", name)
	}

	if len(instances) > 1 {
		return "", fmt.Errorf("multiple instances found with name '%s'. Please use instance ID", name)
	}

	return instances[0], nil
}

// getSSMInstances retrieves SSM-enabled instances
func (m *Manager) getSSMInstances(ctx context.Context, ssmClient *ssm.Client, filters *ListFilters) ([]ssmtypes.InstanceInformation, error) {
	input := &ssm.DescribeInstanceInformationInput{}

	// Apply filters
	if filters != nil && filters.Status != "" {
		input.Filters = append(input.Filters, ssmtypes.InstanceInformationStringFilter{
			Key:    aws.String("PingStatus"),
			Values: []string{filters.Status},
		})
	}

	resp, err := ssmClient.DescribeInstanceInformation(ctx, input)
	if err != nil {
		return nil, err
	}

	return resp.InstanceInformationList, nil
}

// enrichWithEC2Data adds EC2 instance data to SSM instances
func (m *Manager) enrichWithEC2Data(ctx context.Context, ec2Client *ec2.Client, ssmInstances []ssmtypes.InstanceInformation) ([]Instance, error) {
	if len(ssmInstances) == 0 {
		return []Instance{}, nil
	}

	// Extract instance IDs
	instanceIDs := make([]string, len(ssmInstances))
	for i, inst := range ssmInstances {
		instanceIDs[i] = aws.ToString(inst.InstanceId)
	}

	// Get EC2 data
	resp, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	})
	if err != nil {
		return nil, err
	}

	// Create mapping
	ec2Data := make(map[string]types.Instance)
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			ec2Data[aws.ToString(instance.InstanceId)] = instance
		}
	}

	// Combine data
	instances := make([]Instance, 0, len(ssmInstances))
	for _, ssmInst := range ssmInstances {
		instanceID := aws.ToString(ssmInst.InstanceId)
		instance := Instance{
			InstanceID:       instanceID,
			SSMStatus:        string(ssmInst.PingStatus),
			SSMAgentVersion:  aws.ToString(ssmInst.AgentVersion),
			LastPingDateTime: ssmInst.LastPingDateTime.Format(time.RFC3339),
			Platform:         aws.ToString(ssmInst.PlatformName),
		}

		// Add EC2 data if available
		if ec2Inst, exists := ec2Data[instanceID]; exists {
			instance.State = string(ec2Inst.State.Name)
			instance.PrivateIPAddress = aws.ToString(ec2Inst.PrivateIpAddress)
			if ec2Inst.PublicIpAddress != nil {
				instance.PublicIPAddress = aws.ToString(ec2Inst.PublicIpAddress)
			}

			// Extract name from tags
			for _, tag := range ec2Inst.Tags {
				if aws.ToString(tag.Key) == "Name" {
					instance.Name = aws.ToString(tag.Value)
					break
				}
			}

			// Store all tags
			instance.Tags = make(map[string]string)
			for _, tag := range ec2Inst.Tags {
				instance.Tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// waitForCommandCompletion waits for a command to complete and returns the result
func (m *Manager) waitForCommandCompletion(ctx context.Context, ssmClient *ssm.Client, commandID, instanceID, region string) (*CommandResult, error) {
	maxWait := 5 * time.Minute
	pollInterval := 2 * time.Second
	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		// Check command status
		listResp, err := ssmClient.ListCommandInvocations(ctx, &ssm.ListCommandInvocationsInput{
			CommandId:  aws.String(commandID),
			InstanceId: aws.String(instanceID),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to check command status: %w", err)
		}

		if len(listResp.CommandInvocations) == 0 {
			time.Sleep(pollInterval)
			continue
		}

		invocation := listResp.CommandInvocations[0]
		status := string(invocation.Status)

		// If still in progress, continue waiting
		if status == "InProgress" || status == "Pending" || status == "Delayed" {
			time.Sleep(pollInterval)
			continue
		}

		// Command completed, get detailed results
		detailResp, err := ssmClient.GetCommandInvocation(ctx, &ssm.GetCommandInvocationInput{
			CommandId:  aws.String(commandID),
			InstanceId: aws.String(instanceID),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get command result: %w", err)
		}

		result := &CommandResult{
			InstanceID:  instanceID,
			Status:      status,
			Output:      aws.ToString(detailResp.StandardOutputContent),
			ErrorOutput: aws.ToString(detailResp.StandardErrorContent),
		}

		if detailResp.ResponseCode != 0 {
			result.ExitCode = &detailResp.ResponseCode
		}

		return result, nil
	}

	return nil, fmt.Errorf("command execution timed out after %v", maxWait)
}

// File transfer helper methods

func (m *Manager) uploadFileSmall(ctx context.Context, instanceID, region, localPath, remotePath string) error {
	// Read file content
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	// Encode content as base64
	encoded := base64.StdEncoding.EncodeToString(content)

	// Create upload command
	command := fmt.Sprintf(`echo '%s' | base64 -d > '%s'`, encoded, remotePath)

	// Execute via SSM
	result, err := m.ExecuteCommand(ctx, instanceID, region, command, "File upload via ztictl")
	if err != nil {
		return fmt.Errorf("upload command failed: %w", err)
	}

	if result.Status != "Success" {
		return fmt.Errorf("upload failed: %s", result.ErrorOutput)
	}

	return nil
}

func (m *Manager) downloadFileSmall(ctx context.Context, instanceID, region, remotePath, localPath string) error {
	// Create download command
	command := fmt.Sprintf(`if [ -f '%s' ]; then cat '%s' | base64; else echo "FILE_NOT_FOUND"; fi`, remotePath, remotePath)

	// Execute via SSM
	result, err := m.ExecuteCommand(ctx, instanceID, region, command, "File download via ztictl")
	if err != nil {
		return fmt.Errorf("download command failed: %w", err)
	}

	if result.Status != "Success" {
		return fmt.Errorf("download failed: %s", result.ErrorOutput)
	}

	if strings.TrimSpace(result.Output) == "FILE_NOT_FOUND" {
		return fmt.Errorf("remote file not found: %s", remotePath)
	}

	// Decode content
	content, err := base64.StdEncoding.DecodeString(strings.TrimSpace(result.Output))
	if err != nil {
		return fmt.Errorf("failed to decode file content: %w", err)
	}

	// Write to local file
	if err := os.WriteFile(localPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	return nil
}

func (m *Manager) uploadFileLarge(ctx context.Context, instanceID, region, localPath, remotePath string) error {
	m.logger.Info("Starting large file upload via S3", "instance", instanceID, "file", localPath, "size", "large")

	// Initialize managers if not already done
	if m.iamManager == nil || m.s3LifecycleManager == nil {
		if err := m.initializeManagers(ctx, region); err != nil {
			return fmt.Errorf("failed to initialize managers: %w", err)
		}
	}

	// Validate instance IAM setup
	if err := m.iamManager.ValidateInstanceIAMSetup(ctx, instanceID, region); err != nil {
		return fmt.Errorf("instance IAM validation failed: %w", err)
	}

	// Get S3 bucket name
	bucketName, err := m.s3LifecycleManager.GetS3BucketName(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to get S3 bucket name: %w", err)
	}

	// Ensure bucket exists with lifecycle configuration
	if err := m.s3LifecycleManager.EnsureS3Bucket(ctx, bucketName, region); err != nil {
		return fmt.Errorf("failed to ensure S3 bucket exists: %w", err)
	}

	// Attach S3 permissions to instance IAM role
	m.logger.Info("Attaching temporary S3 permissions to instance", "instance", instanceID)
	cleanup, err := m.iamManager.AttachS3Permissions(ctx, instanceID, region, bucketName)
	if err != nil {
		return fmt.Errorf("failed to attach S3 permissions: %w", err)
	}

	// Defer cleanup of IAM permissions
	defer func() {
		m.logger.Info("Cleaning up temporary IAM permissions", "instance", instanceID)
		if err := cleanup(); err != nil {
			m.logger.Warn("Failed to clean up IAM permissions", "error", err)
		}
	}()

	// Generate unique S3 key for this transfer
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to pseudo-random bytes based on timestamp and nanoseconds
		m.logger.Warn("Failed to generate random bytes for S3 key, using timestamp-based fallback", "error", err)
		nano := time.Now().UnixNano()
		// Generate pseudo-random bytes from timestamp and nanoseconds
		for i := 0; i < 8; i++ {
			randomBytes[i] = byte((nano >> (i * 8)) ^ (nano >> (i * 4)))
		}
	}
	timestamp := time.Now().Unix()
	s3Key := fmt.Sprintf("uploads/%d-%s-%s", timestamp, hex.EncodeToString(randomBytes), filepath.Base(localPath))

	// Defer cleanup of S3 object
	defer func() {
		m.s3LifecycleManager.CleanupS3Object(ctx, bucketName, s3Key, region)
	}()

	// Upload to S3
	if err := m.s3LifecycleManager.UploadToS3(ctx, bucketName, s3Key, localPath, region); err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	m.logger.Info("File uploaded to S3, now downloading on instance", "instance", instanceID)

	// Create the remote directory if it doesn't exist
	remoteDir := filepath.Dir(remotePath)
	downloadCommand := fmt.Sprintf(`
		# Create remote directory if needed
		mkdir -p '%s'
		
		# Download file from S3 to the instance
		aws s3 cp 's3://%s/%s' '%s' --region '%s'
		
		# Verify download and cleanup S3 object
		if [ $? -eq 0 ]; then
			echo "File downloaded successfully to %s"
			aws s3 rm 's3://%s/%s' --region '%s'
		else
			echo "Failed to download file from S3"
			exit 1
		fi
	`, remoteDir, bucketName, s3Key, remotePath, region, remotePath, bucketName, s3Key, region)

	// Execute download command on instance
	result, err := m.ExecuteCommand(ctx, instanceID, region, downloadCommand, "Large file download from S3 via ztictl")
	if err != nil {
		return fmt.Errorf("failed to download file on instance: %w", err)
	}

	if result.Status != "Success" {
		return fmt.Errorf("file download failed on instance: %s", result.ErrorOutput)
	}

	m.logger.Info("Large file upload completed successfully", "instance", instanceID, "remote_path", remotePath)
	return nil
}

func (m *Manager) downloadFileLarge(ctx context.Context, instanceID, region, remotePath, localPath string) error {
	m.logger.Info("Starting large file download via S3", "instance", instanceID, "file", remotePath, "size", "large")

	// Initialize managers if not already done
	if m.iamManager == nil || m.s3LifecycleManager == nil {
		if err := m.initializeManagers(ctx, region); err != nil {
			return fmt.Errorf("failed to initialize managers: %w", err)
		}
	}

	// Validate instance IAM setup
	if err := m.iamManager.ValidateInstanceIAMSetup(ctx, instanceID, region); err != nil {
		return fmt.Errorf("instance IAM validation failed: %w", err)
	}

	// Get S3 bucket name
	bucketName, err := m.s3LifecycleManager.GetS3BucketName(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to get S3 bucket name: %w", err)
	}

	// Ensure bucket exists with lifecycle configuration
	if err := m.s3LifecycleManager.EnsureS3Bucket(ctx, bucketName, region); err != nil {
		return fmt.Errorf("failed to ensure S3 bucket exists: %w", err)
	}

	// Attach S3 permissions to instance IAM role
	m.logger.Info("Attaching temporary S3 permissions to instance", "instance", instanceID)
	cleanup, err := m.iamManager.AttachS3Permissions(ctx, instanceID, region, bucketName)
	if err != nil {
		return fmt.Errorf("failed to attach S3 permissions: %w", err)
	}

	// Defer cleanup of IAM permissions
	defer func() {
		m.logger.Info("Cleaning up temporary IAM permissions", "instance", instanceID)
		if err := cleanup(); err != nil {
			m.logger.Warn("Failed to clean up IAM permissions", "error", err)
		}
	}()

	// Generate unique S3 key for this transfer
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to pseudo-random bytes based on timestamp and nanoseconds
		m.logger.Warn("Failed to generate random bytes for S3 key, using timestamp-based fallback", "error", err)
		nano := time.Now().UnixNano()
		// Generate pseudo-random bytes from timestamp and nanoseconds
		for i := 0; i < 8; i++ {
			randomBytes[i] = byte((nano >> (i * 8)) ^ (nano >> (i * 4)))
		}
	}
	timestamp := time.Now().Unix()
	s3Key := fmt.Sprintf("downloads/%d-%s-%s", timestamp, hex.EncodeToString(randomBytes), filepath.Base(remotePath))

	// Defer cleanup of S3 object
	defer func() {
		m.s3LifecycleManager.CleanupS3Object(ctx, bucketName, s3Key, region)
	}()

	m.logger.Info("Uploading file from instance to S3", "bucket", bucketName, "key", s3Key)

	// Create command to upload to S3 from the instance and then clean up
	uploadCommand := fmt.Sprintf(`
		# Check if file exists
		if [ ! -f '%s' ]; then
			echo "FILE_NOT_FOUND"
			exit 1
		fi
		
		# Upload file from instance to S3
		aws s3 cp '%s' 's3://%s/%s' --region '%s'
		
		# Verify upload
		if [ $? -eq 0 ]; then
			echo "File uploaded successfully to S3"
		else
			echo "Failed to upload file to S3"
			exit 1
		fi
	`, remotePath, remotePath, bucketName, s3Key, region)

	// Execute upload command on instance
	result, err := m.ExecuteCommand(ctx, instanceID, region, uploadCommand, "Large file upload to S3 via ztictl")
	if err != nil {
		return fmt.Errorf("failed to upload file from instance: %w", err)
	}

	if result.Status != "Success" {
		if strings.Contains(result.Output, "FILE_NOT_FOUND") {
			return fmt.Errorf("remote file not found: %s", remotePath)
		}
		return fmt.Errorf("file upload failed on instance: %s", result.ErrorOutput)
	}

	m.logger.Info("File uploaded from instance, now downloading locally")

	// Create local directory if needed
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Download from S3 to local file
	if err := m.s3LifecycleManager.DownloadFromS3(ctx, bucketName, s3Key, localPath, region); err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}

	m.logger.Info("Large file download completed successfully", "local_path", localPath)
	return nil
}

// EmergencyCleanup performs emergency cleanup of IAM policies and resources
func (m *Manager) EmergencyCleanup(ctx context.Context, region string) error {
	m.logger.Info("Performing emergency cleanup", "region", region)

	// Initialize managers if not already done
	if m.iamManager == nil || m.s3LifecycleManager == nil {
		if err := m.initializeManagers(ctx, region); err != nil {
			m.logger.Warn("Failed to initialize managers for emergency cleanup", "error", err)
			return err
		}
	}

	// Perform IAM emergency cleanup
	if err := m.iamManager.EmergencyCleanup(ctx, region); err != nil {
		m.logger.Warn("Failed to perform IAM emergency cleanup", "error", err)
	}

	m.logger.Info("Emergency cleanup completed")
	return nil
}

// Cleanup performs routine cleanup operations
func (m *Manager) Cleanup(ctx context.Context, region string) error {
	// Initialize managers if not already done
	if m.iamManager == nil || m.s3LifecycleManager == nil {
		if err := m.initializeManagers(ctx, region); err != nil {
			return fmt.Errorf("failed to initialize managers: %w", err)
		}
	}

	return nil
}

func (m *Manager) getRemoteFileSize(ctx context.Context, instanceID, region, remotePath string) (int64, error) {
	// Get file size using stat command
	command := fmt.Sprintf(`stat -c %%s '%s' 2>/dev/null || echo "0"`, remotePath)

	result, err := m.ExecuteCommand(ctx, instanceID, region, command, "Get file size via ztictl")
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	if result.Status != "Success" {
		return 0, fmt.Errorf("failed to get file size: %s", result.ErrorOutput)
	}

	size, err := strconv.ParseInt(strings.TrimSpace(result.Output), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file size: %w", err)
	}

	return size, nil
}
