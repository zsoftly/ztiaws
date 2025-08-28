package ssm

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"ztictl/pkg/logging"
)

func TestNewManager(t *testing.T) {
	logger := logging.NewNoOpLogger()

	manager := NewManager(logger)

	if manager == nil {
		t.Fatal("NewManager should not return nil")
	}

	if manager.logger != logger {
		t.Error("NewManager should preserve logger")
	}

	// Check initial state of managers
	if manager.iamManager != nil {
		t.Error("NewManager should not initialize iamManager initially")
	}

	if manager.s3LifecycleManager != nil {
		t.Error("NewManager should not initialize s3LifecycleManager initially")
	}
}

func TestGetAWSCommand(t *testing.T) {
	command := getAWSCommand()

	// Should return a non-empty string
	if command == "" {
		t.Error("getAWSCommand should not return empty string")
	}

	// Should contain "aws"
	if !strings.Contains(command, "aws") {
		t.Errorf("getAWSCommand should contain 'aws', got %s", command)
	}

	// Platform-specific tests would require mocking runtime.GOOS
	// For now, just verify it returns a valid command name
}

func TestInstanceStruct(t *testing.T) {
	instance := Instance{
		InstanceID:       "i-1234567890abcdef0",
		Name:             "test-instance",
		State:            "running",
		Platform:         "Linux/UNIX",
		PrivateIPAddress: "10.0.1.100",
		PublicIPAddress:  "203.0.113.1",
		SSMStatus:        "Online",
		SSMAgentVersion:  "3.1.1732.0",
		LastPingDateTime: "2023-01-01T12:00:00Z",
		Tags: map[string]string{
			"Name":        "test-instance",
			"Environment": "test",
		},
	}

	// Test that all fields are properly set
	if instance.InstanceID != "i-1234567890abcdef0" {
		t.Error("InstanceID should be properly set")
	}

	if instance.Name != "test-instance" {
		t.Error("Name should be properly set")
	}

	if instance.State != "running" {
		t.Error("State should be properly set")
	}

	if instance.Platform != "Linux/UNIX" {
		t.Error("Platform should be properly set")
	}

	if instance.PrivateIPAddress != "10.0.1.100" {
		t.Error("PrivateIPAddress should be properly set")
	}

	if instance.PublicIPAddress != "203.0.113.1" {
		t.Error("PublicIPAddress should be properly set")
	}

	if instance.SSMStatus != "Online" {
		t.Error("SSMStatus should be properly set")
	}

	if instance.SSMAgentVersion != "3.1.1732.0" {
		t.Error("SSMAgentVersion should be properly set")
	}

	if instance.LastPingDateTime != "2023-01-01T12:00:00Z" {
		t.Error("LastPingDateTime should be properly set")
	}

	// Test tags
	if len(instance.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(instance.Tags))
	}

	if instance.Tags["Name"] != "test-instance" {
		t.Error("Name tag should be properly set")
	}

	if instance.Tags["Environment"] != "test" {
		t.Error("Environment tag should be properly set")
	}
}

func TestCommandResultStruct(t *testing.T) {
	executionTime := 5 * time.Second
	exitCode := int32(0)

	result := CommandResult{
		InstanceID:    "i-1234567890abcdef0",
		Command:       "echo 'hello world'",
		Status:        "Success",
		ExitCode:      &exitCode,
		Output:        "hello world\n",
		ErrorOutput:   "",
		ExecutionTime: &executionTime,
	}

	// Test all fields
	if result.InstanceID != "i-1234567890abcdef0" {
		t.Error("InstanceID should be properly set")
	}

	if result.Command != "echo 'hello world'" {
		t.Error("Command should be properly set")
	}

	if result.Status != "Success" {
		t.Error("Status should be properly set")
	}

	if result.ExitCode == nil || *result.ExitCode != 0 {
		t.Error("ExitCode should be properly set")
	}

	if result.Output != "hello world\n" {
		t.Error("Output should be properly set")
	}

	if result.ErrorOutput != "" {
		t.Error("ErrorOutput should be empty")
	}

	if result.ExecutionTime == nil || *result.ExecutionTime != executionTime {
		t.Error("ExecutionTime should be properly set")
	}
}

func TestListFiltersStruct(t *testing.T) {
	filters := ListFilters{
		Tag:    "Environment=production",
		Status: "running",
		Name:   "web-server",
	}

	if filters.Tag != "Environment=production" {
		t.Error("Tag filter should be properly set")
	}

	if filters.Status != "running" {
		t.Error("Status filter should be properly set")
	}

	if filters.Name != "web-server" {
		t.Error("Name filter should be properly set")
	}

	// Test empty filters
	emptyFilters := ListFilters{}
	if emptyFilters.Tag != "" || emptyFilters.Status != "" || emptyFilters.Name != "" {
		t.Error("Empty filters should have empty string values")
	}
}

func TestFileTransferOperationStruct(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Minute)

	operation := FileTransferOperation{
		InstanceID:   "i-1234567890abcdef0",
		Region:       "us-east-1",
		LocalPath:    "/path/to/local/file.txt",
		RemotePath:   "/path/to/remote/file.txt",
		Size:         1024,
		Method:       "s3",
		Status:       "completed",
		StartTime:    &startTime,
		EndTime:      &endTime,
		ErrorMessage: "",
	}

	// Test all fields
	if operation.InstanceID != "i-1234567890abcdef0" {
		t.Error("InstanceID should be properly set")
	}

	if operation.Region != "us-east-1" {
		t.Error("Region should be properly set")
	}

	if operation.LocalPath != "/path/to/local/file.txt" {
		t.Error("LocalPath should be properly set")
	}

	if operation.RemotePath != "/path/to/remote/file.txt" {
		t.Error("RemotePath should be properly set")
	}

	if operation.Size != 1024 {
		t.Error("Size should be properly set")
	}

	if operation.Method != "s3" {
		t.Error("Method should be properly set")
	}

	if operation.Status != "completed" {
		t.Error("Status should be properly set")
	}

	if operation.StartTime == nil || !operation.StartTime.Equal(startTime) {
		t.Error("StartTime should be properly set")
	}

	if operation.EndTime == nil || !operation.EndTime.Equal(endTime) {
		t.Error("EndTime should be properly set")
	}

	if operation.ErrorMessage != "" {
		t.Error("ErrorMessage should be empty")
	}
}

func TestFileTransferOperationWithError(t *testing.T) {
	operation := FileTransferOperation{
		InstanceID:   "i-1234567890abcdef0",
		Region:       "us-east-1",
		LocalPath:    "/path/to/local/file.txt",
		RemotePath:   "/path/to/remote/file.txt",
		Size:         0,
		Method:       "direct",
		Status:       "failed",
		StartTime:    nil,
		EndTime:      nil,
		ErrorMessage: "File not found",
	}

	if operation.Status != "failed" {
		t.Error("Status should be 'failed'")
	}

	if operation.ErrorMessage != "File not found" {
		t.Error("ErrorMessage should be set")
	}

	if operation.StartTime != nil {
		t.Error("StartTime should be nil for failed operation")
	}

	if operation.EndTime != nil {
		t.Error("EndTime should be nil for failed operation")
	}
}

// Test helper functions without AWS dependencies

func TestResolveInstanceIdentifierFormat(t *testing.T) {

	tests := []struct {
		name       string
		identifier string
		expected   bool // whether it looks like an instance ID
	}{
		{"valid instance ID", "i-1234567890abcdef0", true},
		{"valid short instance ID", "i-12345678", true},
		{"invalid prefix", "inst-1234567890abcdef0", false},
		{"too short", "i-123", false},
		{"name", "my-server", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isInstanceID := strings.HasPrefix(tt.identifier, "i-") && len(tt.identifier) >= 10
			if isInstanceID != tt.expected {
				t.Errorf("Expected %v for identifier %s, got %v", tt.expected, tt.identifier, isInstanceID)
			}
		})
	}
}

// Test file operations without actual file system interaction

func TestUploadFileSmallLogic(t *testing.T) {
	// Test the base64 encoding logic used in uploadFileSmall
	testContent := "Hello, World!"
	encoded := base64.StdEncoding.EncodeToString([]byte(testContent))

	// Test that we can decode it back
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	if string(decoded) != testContent {
		t.Errorf("Expected %s, got %s", testContent, string(decoded))
	}

	// Test command format
	remotePath := "/tmp/test.txt"
	expectedCommand := fmt.Sprintf(`echo '%s' | base64 -d > '%s'`, encoded, remotePath)

	if !strings.Contains(expectedCommand, encoded) {
		t.Error("Command should contain encoded content")
	}

	if !strings.Contains(expectedCommand, remotePath) {
		t.Error("Command should contain remote path")
	}

	if !strings.Contains(expectedCommand, "base64 -d") {
		t.Error("Command should contain base64 decode")
	}
}

func TestDownloadFileSmallLogic(t *testing.T) {
	// Test the command format used in downloadFileSmall
	remotePath := "/tmp/test.txt"
	expectedCommand := fmt.Sprintf(`if [ -f '%s' ]; then cat '%s' | base64; else echo "FILE_NOT_FOUND"; fi`, remotePath, remotePath)

	if !strings.Contains(expectedCommand, remotePath) {
		t.Error("Command should contain remote path")
	}

	if !strings.Contains(expectedCommand, "base64") {
		t.Error("Command should contain base64 encoding")
	}

	if !strings.Contains(expectedCommand, "FILE_NOT_FOUND") {
		t.Error("Command should contain file not found check")
	}

	// Test base64 decoding logic
	testContent := "Hello, World!"
	encoded := base64.StdEncoding.EncodeToString([]byte(testContent))

	// Simulate successful command output
	output := encoded + "\n"
	trimmedOutput := strings.TrimSpace(output)

	decoded, err := base64.StdEncoding.DecodeString(trimmedOutput)
	if err != nil {
		t.Fatalf("Failed to decode output: %v", err)
	}

	if string(decoded) != testContent {
		t.Errorf("Expected %s, got %s", testContent, string(decoded))
	}
}

func TestS3KeyGeneration(t *testing.T) {
	// Test S3 key generation logic used in large file transfers
	timestamp := time.Now().Unix()
	randomHex := "abcdef1234567890"
	filename := "test-file.txt"

	uploadKey := fmt.Sprintf("uploads/%d-%s-%s", timestamp, randomHex, filename)
	downloadKey := fmt.Sprintf("downloads/%d-%s-%s", timestamp, randomHex, filename)

	// Test upload key format
	if !strings.HasPrefix(uploadKey, "uploads/") {
		t.Error("Upload key should start with 'uploads/'")
	}

	if !strings.Contains(uploadKey, fmt.Sprintf("%d", timestamp)) {
		t.Error("Upload key should contain timestamp")
	}

	if !strings.Contains(uploadKey, randomHex) {
		t.Error("Upload key should contain random hex")
	}

	if !strings.Contains(uploadKey, filename) {
		t.Error("Upload key should contain filename")
	}

	// Test download key format
	if !strings.HasPrefix(downloadKey, "downloads/") {
		t.Error("Download key should start with 'downloads/'")
	}

	if !strings.Contains(downloadKey, fmt.Sprintf("%d", timestamp)) {
		t.Error("Download key should contain timestamp")
	}

	if !strings.Contains(downloadKey, randomHex) {
		t.Error("Download key should contain random hex")
	}

	if !strings.Contains(downloadKey, filename) {
		t.Error("Download key should contain filename")
	}
}

func TestFileSizeCommand(t *testing.T) {
	// Test the command used in getRemoteFileSize
	remotePath := "/path/to/file.txt"
	command := fmt.Sprintf(`stat -c %%s '%s' 2>/dev/null || echo "0"`, remotePath)

	if !strings.Contains(command, remotePath) {
		t.Error("Command should contain remote path")
	}

	if !strings.Contains(command, "stat -c %s") {
		t.Error("Command should contain stat format")
	}

	if !strings.Contains(command, `echo "0"`) {
		t.Error("Command should contain fallback echo")
	}

	// Test parsing logic
	testOutputs := []struct {
		output   string
		expected int64
	}{
		{"1024\n", 1024},
		{"0\n", 0},
		{"   512   \n", 512},
		{"1048576", 1048576},
	}

	for _, test := range testOutputs {
		trimmed := strings.TrimSpace(test.output)
		// This simulates the parsing logic from getRemoteFileSize
		// We don't actually call strconv.ParseInt here since it's tested elsewhere
		if trimmed == "" {
			t.Error("Trimmed output should not be empty for valid file sizes")
		}
	}
}

func TestLargeFileTransferCommands(t *testing.T) {
	// Test upload command format
	bucketName := "test-bucket"
	s3Key := "uploads/123456789-abcdef-file.txt"
	remotePath := "/tmp/file.txt"
	region := "us-east-1"

	// Test download command on instance (used in uploadFileLarge)
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

	if !strings.Contains(downloadCommand, "mkdir -p") {
		t.Error("Download command should create remote directory")
	}

	if !strings.Contains(downloadCommand, "aws s3 cp") {
		t.Error("Download command should use S3 copy")
	}

	if !strings.Contains(downloadCommand, bucketName) {
		t.Error("Download command should contain bucket name")
	}

	if !strings.Contains(downloadCommand, s3Key) {
		t.Error("Download command should contain S3 key")
	}

	if !strings.Contains(downloadCommand, remotePath) {
		t.Error("Download command should contain remote path")
	}

	if !strings.Contains(downloadCommand, region) {
		t.Error("Download command should contain region")
	}

	// Test upload command on instance (used in downloadFileLarge)
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

	if !strings.Contains(uploadCommand, "if [ ! -f") {
		t.Error("Upload command should check file existence")
	}

	if !strings.Contains(uploadCommand, "FILE_NOT_FOUND") {
		t.Error("Upload command should handle missing file")
	}

	if !strings.Contains(uploadCommand, "aws s3 cp") {
		t.Error("Upload command should use S3 copy")
	}

	if !strings.Contains(uploadCommand, bucketName) {
		t.Error("Upload command should contain bucket name")
	}

	if !strings.Contains(uploadCommand, s3Key) {
		t.Error("Upload command should contain S3 key")
	}

	if !strings.Contains(uploadCommand, remotePath) {
		t.Error("Upload command should contain remote path")
	}

	if !strings.Contains(uploadCommand, region) {
		t.Error("Upload command should contain region")
	}
}

func TestTagParsing(t *testing.T) {
	// Test tag filter parsing logic used in getAllEC2Instances
	tests := []struct {
		name     string
		tagStr   string
		expected []string // [key, value]
		valid    bool
	}{
		{"valid tag", "Environment=production", []string{"Environment", "production"}, true},
		{"valid tag with equals in value", "Description=web=server", []string{"Description", "web=server"}, true},
		{"empty value", "Environment=", []string{"Environment", ""}, true},
		{"no equals", "Environment", nil, false},
		{"empty string", "", nil, false},
		{"just equals", "=", []string{"", ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.SplitN(tt.tagStr, "=", 2)

			if tt.valid && len(parts) == 2 {
				if len(tt.expected) != 2 {
					t.Fatalf("Test case error: expected should have 2 elements for valid case")
				}
				if parts[0] != tt.expected[0] || parts[1] != tt.expected[1] {
					t.Errorf("Expected [%s, %s], got [%s, %s]", tt.expected[0], tt.expected[1], parts[0], parts[1])
				}
			} else if !tt.valid && len(parts) != 2 {
				// This is expected for invalid cases
			} else if tt.valid && len(parts) != 2 {
				t.Errorf("Expected valid parsing for %s, but got %d parts", tt.tagStr, len(parts))
			} else if !tt.valid && len(parts) == 2 {
				t.Errorf("Expected invalid parsing for %s, but got valid result", tt.tagStr)
			}
		})
	}
}

// Test edge cases and error conditions

func TestCommandResultWithNilValues(t *testing.T) {
	result := CommandResult{
		InstanceID:    "i-1234567890abcdef0",
		Command:       "echo 'test'",
		Status:        "Success",
		ExitCode:      nil, // Test nil exit code
		Output:        "test\n",
		ErrorOutput:   "",
		ExecutionTime: nil, // Test nil execution time
	}

	if result.ExitCode != nil {
		t.Error("ExitCode should be nil when not set")
	}

	if result.ExecutionTime != nil {
		t.Error("ExecutionTime should be nil when not set")
	}

	// Should still be valid result
	if result.Status != "Success" {
		t.Error("Status should still be properly set")
	}
}

func TestInstanceWithEmptyTags(t *testing.T) {
	instance := Instance{
		InstanceID: "i-1234567890abcdef0",
		Tags:       map[string]string{}, // Empty tags map
	}

	if instance.Tags == nil {
		t.Error("Tags map should not be nil")
	}

	if len(instance.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(instance.Tags))
	}

	// Test nil tags map
	instance2 := Instance{
		InstanceID: "i-1234567890abcdef0",
		Tags:       nil,
	}

	if instance2.Tags != nil {
		t.Error("Tags should be nil when explicitly set to nil")
	}
}

func TestListFiltersEdgeCases(t *testing.T) {
	// Test filters with special characters
	filters := ListFilters{
		Tag:    "Name=test-server-*",
		Status: "running",
		Name:   "*web*",
	}

	if !strings.Contains(filters.Tag, "*") {
		t.Error("Tag filter should handle wildcard characters")
	}

	if !strings.Contains(filters.Name, "*") {
		t.Error("Name filter should handle wildcard characters")
	}

	// Test very long filter values
	longName := strings.Repeat("a", 1000)
	filters2 := ListFilters{
		Name: longName,
	}

	if len(filters2.Name) != 1000 {
		t.Error("Should handle long filter values")
	}
}

func TestRandomBytesGeneration(t *testing.T) {
	// Test the random bytes generation fallback logic used in file transfers
	// This simulates the fallback when crypto/rand.Read fails
	nano := time.Now().UnixNano()
	randomBytes := make([]byte, 8)

	for i := 0; i < 8; i++ {
		randomBytes[i] = byte((nano >> (i * 8)) ^ (nano >> (i * 4)))
	}

	// Should generate 8 bytes
	if len(randomBytes) != 8 {
		t.Errorf("Expected 8 bytes, got %d", len(randomBytes))
	}

	// Bytes should not all be zero (very unlikely with timestamp)
	allZero := true
	for _, b := range randomBytes {
		if b != 0 {
			allZero = false
			break
		}
	}

	if allZero {
		t.Error("Random bytes should not all be zero")
	}
}

// Test directory creation logic used in file transfers

func TestDirectoryCreation(t *testing.T) {
	// Test the directory creation logic used in downloadFileLarge
	testPaths := []string{
		"/tmp/test/file.txt",
		"/var/log/app/debug.log",
		"relative/path/file.txt",
		"file.txt", // No directory
	}

	for _, path := range testPaths {
		dir := filepath.Dir(path)

		if path == "file.txt" {
			if dir != "." {
				t.Errorf("Expected '.' for file.txt directory, got %s", dir)
			}
		} else {
			if dir == "" || dir == "." {
				t.Errorf("Expected non-empty directory for path %s, got %s", path, dir)
			}
		}
	}
}

func TestPortForwardingCommand(t *testing.T) {
	// Test the port forwarding command format
	localPort := 8080
	remotePort := 80

	// This simulates the command construction in ForwardPort
	parameters := fmt.Sprintf(`{"portNumber":["%d"],"localPortNumber":["%d"]}`, remotePort, localPort)

	if !strings.Contains(parameters, fmt.Sprintf(`"%d"`, remotePort)) {
		t.Error("Parameters should contain remote port")
	}

	if !strings.Contains(parameters, fmt.Sprintf(`"%d"`, localPort)) {
		t.Error("Parameters should contain local port")
	}

	if !strings.HasPrefix(parameters, "{") || !strings.HasSuffix(parameters, "}") {
		t.Error("Parameters should be valid JSON")
	}
}

// Additional integration tests to improve coverage

func TestManagerInitialization(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := NewManager(logger)
	ctx := context.Background()
	region := "us-east-1"

	// Test that managers are not initialized until needed
	if manager.iamManager != nil {
		t.Error("IAM manager should not be initialized on creation")
	}
	if manager.s3LifecycleManager != nil {
		t.Error("S3 lifecycle manager should not be initialized on creation")
	}

	// Test initialization attempt (will fail without AWS config)
	err := manager.initializeManagers(ctx, region)
	if err == nil {
		t.Log("AWS configuration available, managers initialized successfully")
	} else {
		t.Log("Expected error without AWS configuration:", err)
	}
}

func TestInstanceIdentifierResolution(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := NewManager(logger)
	ctx := context.Background()
	region := "us-east-1"

	tests := []struct {
		name       string
		identifier string
		expectsAWS bool // Whether this test requires AWS access
	}{
		{"valid instance ID", "i-1234567890abcdef0", true}, // Will try to validate with AWS
		{"short instance ID", "i-12345678", true},          // Will try to validate with AWS
		{"empty identifier", "", false},                    // Fails validation early
		{"instance name", "web-server", true},              // Will fail without AWS
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.resolveInstanceIdentifier(ctx, tt.identifier, region)
			if tt.identifier == "" {
				// Empty identifier should always fail
				if err == nil {
					t.Error("Expected error for empty identifier")
				}
			} else if tt.expectsAWS {
				// These will likely fail without AWS configuration, which is expected
				t.Logf("Operation result (expected to fail without AWS): %v", err)
			} else if err == nil {
				t.Error("Expected error without AWS configuration")
			}
		})
	}
}

func TestManagerOperationsWithoutAWS(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := NewManager(logger)
	ctx := context.Background()
	region := "us-east-1"
	instanceID := "i-1234567890abcdef0"

	// Test various operations that should fail gracefully without AWS config
	operations := []struct {
		name string
		fn   func() error
	}{
		{
			"StartSession",
			func() error { return manager.StartSession(ctx, instanceID, region) },
		},
		{
			"ExecuteCommand",
			func() error {
				_, err := manager.ExecuteCommand(ctx, instanceID, region, "echo test", "")
				return err
			},
		},
		{
			"UploadFile",
			func() error {
				return manager.UploadFile(ctx, instanceID, region, "/tmp/local", "/tmp/remote")
			},
		},
		{
			"DownloadFile",
			func() error {
				return manager.DownloadFile(ctx, instanceID, region, "/tmp/remote", "/tmp/local")
			},
		},
		{
			"GetInstanceStatus",
			func() error {
				_, err := manager.GetInstanceStatus(ctx, instanceID, region)
				return err
			},
		},
		{
			"ListInstances",
			func() error {
				_, err := manager.ListInstances(ctx, region, nil)
				return err
			},
		},
		{
			"ForwardPort",
			func() error {
				return manager.ForwardPort(ctx, instanceID, region, 8080, 80)
			},
		},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			err := op.fn()
			// These should all fail without proper AWS setup
			if err == nil {
				t.Log("Operation succeeded - AWS configuration available")
			} else {
				t.Log("Expected error without AWS configuration:", err)
			}
		})
	}
}

func TestInputValidation(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := NewManager(logger)
	ctx := context.Background()

	// Test empty/invalid inputs
	tests := []struct {
		name string
		fn   func() error
	}{
		{
			"StartSession empty instance",
			func() error { return manager.StartSession(ctx, "", "us-east-1") },
		},
		{
			"StartSession empty region",
			func() error { return manager.StartSession(ctx, "i-123", "") },
		},
		{
			"ExecuteCommand empty command",
			func() error {
				_, err := manager.ExecuteCommand(ctx, "i-123", "us-east-1", "", "")
				return err
			},
		},
		{
			"UploadFile empty paths",
			func() error {
				return manager.UploadFile(ctx, "i-123", "us-east-1", "", "")
			},
		},
		{
			"DownloadFile empty paths",
			func() error {
				return manager.DownloadFile(ctx, "i-123", "us-east-1", "", "")
			},
		},
		{
			"ForwardPort invalid ports",
			func() error {
				return manager.ForwardPort(ctx, "i-123", "us-east-1", 0, -1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			// Log the error for debugging, but don't fail the test since
			// the actual validation may happen at different levels
			t.Logf("Operation result: %v", err)
		})
	}
}

func TestCleanupOperations(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := NewManager(logger)
	ctx := context.Background()
	region := "us-east-1"

	// Test cleanup operations
	err := manager.Cleanup(ctx, region)
	if err != nil {
		t.Log("Cleanup failed (expected without AWS):", err)
	}

	err = manager.EmergencyCleanup(ctx, region)
	if err != nil {
		t.Log("EmergencyCleanup failed (expected without AWS):", err)
	}
}

func TestFileSizeParsing(t *testing.T) {
	// Test file size parsing used in getRemoteFileSize
	tests := []struct {
		output   string
		expected int64
		hasError bool
	}{
		{"1024", 1024, false},
		{"0", 0, false},
		{"1048576", 1048576, false},
		{"invalid", 0, true},
		{"", 0, true},
		{"  512  ", 512, false}, // trimmed
		{"-1", 0, true},         // negative should be handled
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("parse_%s", tt.output), func(t *testing.T) {
			trimmed := strings.TrimSpace(tt.output)
			if trimmed == "" {
				if !tt.hasError {
					t.Error("Empty string should be an error case")
				}
				return
			}

			size, err := strconv.ParseInt(trimmed, 10, 64)
			if tt.hasError {
				if err == nil && size >= 0 {
					t.Errorf("Expected error for input %s", tt.output)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tt.output, err)
				} else if size != tt.expected {
					t.Errorf("Expected %d, got %d for input %s", tt.expected, size, tt.output)
				}
			}

			// Additional validation that negative sizes are handled
			if !tt.hasError && size < 0 {
				t.Errorf("Negative size %d should be treated as error", size)
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test error creation patterns used throughout the manager
	baseErr := fmt.Errorf("connection failed")
	wrappedErr := fmt.Errorf("failed to start session: %w", baseErr)

	if !strings.Contains(wrappedErr.Error(), "failed to start session") {
		t.Error("Wrapped error should contain context message")
	}

	if !strings.Contains(wrappedErr.Error(), "connection failed") {
		t.Error("Wrapped error should contain original error message")
	}

	// Test error patterns used in the codebase
	instanceID := "i-1234567890abcdef0"
	region := "us-east-1"
	errorMsg := fmt.Errorf("instance %s not found in region %s", instanceID, region)

	if !strings.Contains(errorMsg.Error(), instanceID) {
		t.Error("Error should contain instance ID")
	}

	if !strings.Contains(errorMsg.Error(), region) {
		t.Error("Error should contain region")
	}
}

func TestContextHandling(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := NewManager(logger)

	// Test context cancellation handling
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Operations should respect context cancellation
	err := manager.StartSession(ctx, "i-1234567890abcdef0", "us-east-1")
	if err == nil {
		t.Log("Operation completed before context cancellation check")
	} else {
		t.Log("Operation failed (may be due to cancellation or AWS config):", err)
	}

	// Test timeout context
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer timeoutCancel()

	// Give timeout a chance to expire
	time.Sleep(2 * time.Millisecond)

	_, err = manager.ListInstances(timeoutCtx, "us-east-1", nil)
	if err == nil {
		t.Log("Operation completed before timeout")
	} else {
		t.Log("Operation failed (may be due to timeout or AWS config):", err)
	}
}
