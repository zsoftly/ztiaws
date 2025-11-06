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
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	appconfig "ztictl/internal/config"
	"ztictl/internal/interactive"
	"ztictl/internal/platform"
	awsservice "ztictl/pkg/aws"
	"ztictl/pkg/errors"
	"ztictl/pkg/logging"
	"ztictl/pkg/security"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// Compiled regex patterns for input validation (performance optimization)
var (
	// AWS instance IDs follow the pattern: i-[0-9a-f]{8,17}
	instanceIDRegex = regexp.MustCompile(`^i-[0-9a-f]{8,17}$`)

	// AWS regions follow patterns like: us-east-1, eu-west-2, ap-southeast-1, etc.
	awsRegionRegex = regexp.MustCompile(`^[a-z]{2,3}-[a-z]+-[0-9]+$`)
)

// Manager handles AWS Systems Manager operations
type Manager struct {
	mu                 sync.Mutex
	logger             *logging.Logger
	instanceService    *awsservice.InstanceService
	iamManager         *IAMManager
	s3LifecycleManager *S3LifecycleManager
	platformDetector   *platform.Detector
	builderManager     *platform.BuilderManager
	clientPool         *ClientPool
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
	Tag    string `json:"tag,omitempty"`    // Format: key=value (deprecated, use Tags)
	Tags   string `json:"tags,omitempty"`   // Format: key1=value1,key2=value2
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
	// Note: Platform detector and builder manager are not initialized here.
	// They will be initialized on first use via initializePlatformComponents()
	clientPool := NewClientPool()
	clientPoolAdapter := NewClientPoolAdapter(clientPool)

	return &Manager{
		logger:          logger,
		clientPool:      clientPool,
		instanceService: awsservice.NewInstanceService(clientPoolAdapter, logger),
	}
}

// StartSession starts an SSM session to an instance
func (m *Manager) StartSession(ctx context.Context, instanceIdentifier, region string) error {
	// Resolve instance identifier to instance ID
	instanceID, err := m.resolveInstanceIdentifier(ctx, instanceIdentifier, region)
	if err != nil {
		return fmt.Errorf("failed to resolve instance: %w", err)
	}

	m.logger.Info("Starting SSM session for instance", "instanceID", instanceID, "region", region)

	// Validate parameters to prevent command injection
	if err := validateInstanceID(instanceID); err != nil {
		return fmt.Errorf("invalid instance ID: %w", err)
	}
	if err := validateAWSRegion(region); err != nil {
		return fmt.Errorf("invalid region: %w", err)
	}

	// Use AWS CLI for session manager (Go SDK doesn't support interactive sessions)
	// Build command with validated and sanitized parameters
	awsCmd := getAWSCommand()

	// Parameters are already validated above, but create explicit parameter strings for clarity
	regionParam := region
	targetParam := instanceID

	// #nosec G204 - Parameters are validated above using strict regex patterns for AWS instance ID and region format
	cmd := exec.CommandContext(ctx, awsCmd,
		"ssm", "start-session",
		"--region", regionParam,
		"--target", targetParam)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.NewSSMError("failed to start session", err)
	}

	return nil
}

// initializePlatformComponents initializes platform detection components if not already done
func (m *Manager) initializePlatformComponents(ctx context.Context, region string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.platformDetector != nil && m.builderManager != nil {
		return nil
	}

	ssmClient, ec2Client, err := m.clientPool.GetPlatformClients(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to get AWS clients from pool: %w", err)
	}

	m.platformDetector = platform.NewDetector(ssmClient, ec2Client, m.logger)
	m.builderManager = platform.NewBuilderManager(m.platformDetector)

	return nil
}

// GetInstanceService exposes the shared instance service
func (m *Manager) GetInstanceService() *awsservice.InstanceService {
	return m.instanceService
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
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.iamManager != nil && m.s3LifecycleManager != nil {
		return nil
	}

	clients, err := m.clientPool.GetClients(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to get AWS clients from pool: %w", err)
	}

	if m.iamManager == nil {
		iamManager, err := NewIAMManager(m.logger, clients.IAMClient, clients.EC2Client)
		if err != nil {
			return fmt.Errorf("failed to create IAM manager: %w", err)
		}
		m.iamManager = iamManager
	}

	if m.s3LifecycleManager == nil {
		m.s3LifecycleManager = NewS3LifecycleManager(m.logger, clients.S3Client, clients.STSClient)
	}

	return nil
}

// ListInstances lists all EC2 instances in a region with their SSM status
func (m *Manager) ListInstances(ctx context.Context, region string, filters *ListFilters) ([]interactive.Instance, error) {
	// Convert SSM ListFilters to AWS ListFilters
	var awsFilters *awsservice.ListFilters
	if filters != nil {
		awsFilters = &awsservice.ListFilters{
			Tag:    filters.Tag,
			Tags:   filters.Tags,
			Status: filters.Status,
			Name:   filters.Name,
		}
	}

	return m.instanceService.ListInstances(ctx, region, awsFilters)
}

// ExecuteCommand executes a command on an instance via SSM
func (m *Manager) ExecuteCommand(ctx context.Context, instanceIdentifier, region, command, comment string) (*CommandResult, error) {
	// Resolve instance identifier
	instanceID, err := m.resolveInstanceIdentifier(ctx, instanceIdentifier, region)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve instance: %w", err)
	}

	m.logger.Info("Executing command on instance", "instanceID", instanceID, "command", command)

	// Initialize platform components if needed
	if err := m.initializePlatformComponents(ctx, region); err != nil {
		return nil, fmt.Errorf("failed to initialize platform components: %w", err)
	}

	// Get command builder for the instance platform
	builder, err := m.builderManager.GetBuilder(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get command builder: %w", err)
	}

	// Get SSM client from pool
	ssmClient, err := m.clientPool.GetSSMClient(ctx, region)
	if err != nil {
		return nil, errors.NewAWSError("failed to get SSM client", err)
	}

	// Send command
	if comment == "" {
		comment = "Command executed via ztictl"
	}

	startTime := time.Now()

	// Get the appropriate SSM document for the platform
	documentName := builder.GetSSMDocument()

	// Build the command with platform-specific wrapper
	wrappedCommand := builder.BuildExecCommand(command)

	sendResp, err := ssmClient.SendCommand(ctx, &ssm.SendCommandInput{
		DocumentName: aws.String(documentName),
		InstanceIds:  []string{instanceID},
		Parameters: map[string][]string{
			"commands": {wrappedCommand},
		},
		Comment: aws.String(comment),
	})
	if err != nil {
		return nil, errors.NewSSMError("failed to send command", err)
	}

	commandID := aws.ToString(sendResp.Command.CommandId)
	m.logger.Debug("Command sent with ID", "commandID", commandID)

	// Wait for command completion
	result, err := m.waitForCommandCompletion(ctx, ssmClient, commandID, instanceID)
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

	// Validate that the local path is within safe boundaries
	if err := security.ValidateFilePathWithWorkingDir(localPath); err != nil {
		return fmt.Errorf("unsafe file path: %w", err)
	}

	// Check if local file exists
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("local file not found: %w", err)
	}

	cfg := appconfig.Get()

	m.logger.Info("Uploading file to instance", "instanceID", instanceID, "localPath", localPath, "remotePath", remotePath, "size", fileInfo.Size())

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

	// Validate that the local path is within safe boundaries
	if err := security.ValidateFilePathWithWorkingDir(localPath); err != nil {
		return fmt.Errorf("unsafe file path: %w", err)
	}

	m.logger.Info("Downloading file from instance", "instanceID", instanceID, "remotePath", remotePath, "localPath", localPath)

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

	m.logger.Info("Starting port forwarding for instance", "instanceID", instanceID, "localPort", localPort, "remotePort", remotePort)

	// Validate parameters to prevent command injection
	if err := validateInstanceID(instanceID); err != nil {
		return fmt.Errorf("invalid instance ID: %w", err)
	}
	if err := validateAWSRegion(region); err != nil {
		return fmt.Errorf("invalid region: %w", err)
	}
	if err := validatePortNumber(localPort); err != nil {
		return fmt.Errorf("invalid local port: %w", err)
	}
	if err := validatePortNumber(remotePort); err != nil {
		return fmt.Errorf("invalid remote port: %w", err)
	}

	// Use AWS CLI for port forwarding (Go SDK doesn't support this directly)
	// Build command with validated and sanitized parameters
	awsCmd := getAWSCommand()

	// Parameters are already validated above, but create explicit parameter strings for clarity
	regionParam := region
	targetParam := instanceID
	parametersJSON := fmt.Sprintf(`{"portNumber":["%d"],"localPortNumber":["%d"]}`, remotePort, localPort)

	// #nosec G204 - Parameters are validated above using strict regex patterns for AWS instance ID, region format, and port ranges
	cmd := exec.CommandContext(ctx, awsCmd,
		"ssm", "start-session",
		"--region", regionParam,
		"--target", targetParam,
		"--document-name", "AWS-StartPortForwardingSession",
		"--parameters", parametersJSON)

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
func (m *Manager) GetInstanceStatus(ctx context.Context, instanceIdentifier, region string) (*interactive.Instance, error) {
	// Resolve instance identifier
	instanceID, err := m.resolveInstanceIdentifier(ctx, instanceIdentifier, region)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve instance: %w", err)
	}

	// Get SSM client from pool
	ssmClient, err := m.clientPool.GetSSMClient(ctx, region)
	if err != nil {
		return nil, errors.NewAWSError("failed to get SSM client", err)
	}

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

	return &interactive.Instance{
		InstanceID:       aws.ToString(info.InstanceId),
		SSMStatus:        string(info.PingStatus),
		SSMAgentVersion:  aws.ToString(info.AgentVersion),
		LastPingDateTime: info.LastPingDateTime.Format(time.RFC3339),
		Platform:         aws.ToString(info.PlatformName),
	}, nil
}

// ListInstanceStatuses lists SSM status for all instances in a region
func (m *Manager) ListInstanceStatuses(ctx context.Context, region string) ([]interactive.Instance, error) {
	// Get SSM client from pool
	ssmClient, err := m.clientPool.GetSSMClient(ctx, region)
	if err != nil {
		return nil, errors.NewAWSError("failed to get SSM client", err)
	}

	// Get all SSM instances
	resp, err := ssmClient.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{})
	if err != nil {
		return nil, errors.NewSSMError("failed to describe instance information", err)
	}

	instances := make([]interactive.Instance, len(resp.InstanceInformationList))
	for i, info := range resp.InstanceInformationList {
		instances[i] = interactive.Instance{
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

// parseTagFilters parses comma-separated tag filters into individual key=value pairs
func parseTagFilters(tagsStr string) (map[string]string, error) {
	if tagsStr == "" {
		return nil, nil
	}

	tags := make(map[string]string)
	tagPairs := strings.Split(tagsStr, ",")

	for _, tagPair := range tagPairs {
		tagPair = strings.TrimSpace(tagPair)
		if tagPair == "" {
			continue
		}

		parts := strings.SplitN(tagPair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tag format '%s'. Expected format: key=value", tagPair)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" || value == "" {
			return nil, fmt.Errorf("empty tag key or value in '%s'", tagPair)
		}

		tags[key] = value
	}

	return tags, nil
}

// resolveInstanceIdentifier resolves an instance name or ID to an instance ID
func (m *Manager) resolveInstanceIdentifier(ctx context.Context, identifier, region string) (string, error) {
	return m.instanceService.ResolveInstanceIdentifier(ctx, identifier, region)
}

// waitForCommandCompletion waits for a command to complete and returns the result
func (m *Manager) waitForCommandCompletion(ctx context.Context, ssmClient *ssm.Client, commandID, instanceID string) (*CommandResult, error) {
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

		// Clean the output to remove the EXIT_CODE line that was added by the wrapper script
		cleanOutput := removeExitCodeLine(aws.ToString(detailResp.StandardOutputContent))

		result := &CommandResult{
			InstanceID:  instanceID,
			Status:      status,
			Output:      cleanOutput,
			ErrorOutput: aws.ToString(detailResp.StandardErrorContent),
		}

		if detailResp.ResponseCode != 0 {
			result.ExitCode = &detailResp.ResponseCode
		}

		return result, nil
	}

	return nil, fmt.Errorf("command execution timed out after %v", maxWait)
}

// removeExitCodeLine removes the EXIT_CODE line from command output
// The platform builders add this line to capture exit codes, but it shouldn't be shown to users
func removeExitCodeLine(output string) string {
	if output == "" {
		return output
	}

	lines := strings.Split(output, "\n")
	filteredLines := make([]string, 0, len(lines))

	for _, line := range lines {
		// Skip lines that start with EXIT_CODE:
		if strings.HasPrefix(line, "EXIT_CODE:") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}

	// Preserve trailing newline if original output had one
	result := strings.Join(filteredLines, "\n")
	if strings.HasSuffix(output, "\n") && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result
}

// File transfer helper methods

func (m *Manager) uploadFileSmall(ctx context.Context, instanceID, region, localPath, remotePath string) error {
	// Note: File path validation is performed in UploadFile() caller
	// Clean the path for consistent handling
	cleanPath, err := filepath.Abs(filepath.Clean(localPath))
	if err != nil {
		return fmt.Errorf("invalid local file path: %w", err)
	}

	// Read file content
	// #nosec G304 - cleanPath is derived from localPath which is validated in UploadFile() caller using security.ValidateFilePathWithWorkingDir()
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	// Initialize platform components if needed
	if err := m.initializePlatformComponents(ctx, region); err != nil {
		return fmt.Errorf("failed to initialize platform components: %w", err)
	}

	// Get command builder for the instance platform
	builder, err := m.builderManager.GetBuilder(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get command builder: %w", err)
	}

	// Encode content as base64
	encoded := base64.StdEncoding.EncodeToString(content)

	// Create platform-specific upload command
	command, err := builder.BuildFileWriteCommand(remotePath, encoded)
	if err != nil {
		return fmt.Errorf("failed to build file write command: %w", err)
	}

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
	// Note: File path validation is performed in DownloadFile() caller
	// Initialize platform components if needed
	if err := m.initializePlatformComponents(ctx, region); err != nil {
		return fmt.Errorf("failed to initialize platform components: %w", err)
	}

	// Get command builder for the instance platform
	builder, err := m.builderManager.GetBuilder(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to get command builder: %w", err)
	}

	// First check if file exists
	checkCommand := builder.BuildFileExistsCommand(remotePath)
	checkResult, err := m.ExecuteCommand(ctx, instanceID, region, checkCommand, "Check file existence via ztictl")
	if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	// Use actual exit code from the command result
	actualExitCode := 0
	if checkResult.ExitCode != nil {
		actualExitCode = int(*checkResult.ExitCode)
	}

	exists, err := builder.ParseFileExists(checkResult.Output, actualExitCode)
	if err != nil {
		return fmt.Errorf("failed to parse file existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("remote file not found: %s", remotePath)
	}

	// Create platform-specific download command
	command := builder.BuildFileReadCommand(remotePath)

	// Execute via SSM
	result, err := m.ExecuteCommand(ctx, instanceID, region, command, "File download via ztictl")
	if err != nil {
		return fmt.Errorf("download command failed: %w", err)
	}

	if result.Status != "Success" {
		return fmt.Errorf("download failed: %s", result.ErrorOutput)
	}

	// Decode content
	content, err := base64.StdEncoding.DecodeString(strings.TrimSpace(result.Output))
	if err != nil {
		return fmt.Errorf("failed to decode file content: %w", err)
	}

	// Write to local file
	if err := os.WriteFile(localPath, content, 0600); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	return nil
}

func (m *Manager) uploadFileLarge(ctx context.Context, instanceID, region, localPath, remotePath string) error {
	// Note: File path validation is performed in UploadFile() caller
	m.logger.Info("Starting large file upload via S3 for instance", "instanceID", instanceID, "localPath", localPath)

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
	m.logger.Info("Attaching temporary S3 permissions to instance", "instanceID", instanceID)
	cleanup, err := m.iamManager.AttachS3Permissions(ctx, instanceID, region, bucketName)
	if err != nil {
		return fmt.Errorf("failed to attach S3 permissions: %w", err)
	}

	// Defer cleanup of IAM permissions
	defer func() {
		m.logger.Info("Cleaning up temporary IAM permissions for instance", "instanceID", instanceID)
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
		if err := m.s3LifecycleManager.CleanupS3Object(ctx, bucketName, s3Key, region); err != nil {
			m.logger.Warn("Failed to cleanup S3 object", "bucketName", bucketName, "s3Key", s3Key, "error", err)
		}
	}()

	// Upload to S3
	if err := m.s3LifecycleManager.UploadToS3(ctx, bucketName, s3Key, localPath, region); err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	m.logger.Info("File uploaded to S3, now downloading on instance", "instanceID", instanceID)

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

	m.logger.Info("Large file upload completed successfully for instance", "instanceID", instanceID, "remotePath", remotePath)
	return nil
}

func (m *Manager) downloadFileLarge(ctx context.Context, instanceID, region, remotePath, localPath string) error {
	// Note: File path validation is performed in DownloadFile() caller
	m.logger.Info("Starting large file download via S3 for instance", "instanceID", instanceID, "remotePath", remotePath)

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
	m.logger.Info("Attaching temporary S3 permissions to instance", "instanceID", instanceID)
	cleanup, err := m.iamManager.AttachS3Permissions(ctx, instanceID, region, bucketName)
	if err != nil {
		return fmt.Errorf("failed to attach S3 permissions: %w", err)
	}

	// Defer cleanup of IAM permissions
	defer func() {
		m.logger.Info("Cleaning up temporary IAM permissions for instance", "instanceID", instanceID)
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
		if err := m.s3LifecycleManager.CleanupS3Object(ctx, bucketName, s3Key, region); err != nil {
			m.logger.Warn("Failed to cleanup S3 object", "bucketName", bucketName, "s3Key", s3Key, "error", err)
		}
	}()

	m.logger.Info("Uploading file from instance to S3 bucket", "bucketName", bucketName, "s3Key", s3Key)

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
	if err := os.MkdirAll(localDir, 0750); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Download from S3 to local file
	if err := m.s3LifecycleManager.DownloadFromS3(ctx, bucketName, s3Key, localPath, region); err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}

	m.logger.Info("Large file download completed successfully", "localPath", localPath)
	return nil
}

// EmergencyCleanup performs emergency cleanup of IAM policies and resources
func (m *Manager) EmergencyCleanup(ctx context.Context, region string) error {
	m.logger.Info("Performing emergency cleanup in region", "region", region)

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
	// Initialize platform components if needed
	if err := m.initializePlatformComponents(ctx, region); err != nil {
		return 0, fmt.Errorf("failed to initialize platform components: %w", err)
	}

	// Get command builder for the instance platform
	builder, err := m.builderManager.GetBuilder(ctx, instanceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get command builder: %w", err)
	}

	// Build platform-specific command to get file size
	command := builder.BuildFileSizeCommand(remotePath)

	result, err := m.ExecuteCommand(ctx, instanceID, region, command, "Get file size via ztictl")
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	if result.Status != "Success" {
		return 0, fmt.Errorf("failed to get file size: %s", result.ErrorOutput)
	}

	// Parse the file size using the platform-specific parser
	size, err := builder.ParseFileSize(result.Output)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file size: %w", err)
	}

	return size, nil
}

// validateInstanceID validates AWS EC2 instance ID format
func validateInstanceID(instanceID string) error {
	if !instanceIDRegex.MatchString(instanceID) {
		return fmt.Errorf("instance ID must match pattern i-[0-9a-f]{8,17}, got: %s", instanceID)
	}
	return nil
}

// validateAWSRegion validates AWS region format
func validateAWSRegion(region string) error {
	if !awsRegionRegex.MatchString(region) {
		return fmt.Errorf("region must match AWS format (e.g., us-east-1), got: %s", region)
	}
	return nil
}

// validatePortNumber validates port number range
func validatePortNumber(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got: %d", port)
	}
	return nil
}
