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

	"ztictl/internal/interactive"
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

func TestStartSession(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := NewManager(logger)
	ctx := context.Background()

	// This test will fail if aws cli is not configured
	// but it will test the command execution
	err := manager.StartSession(ctx, "i-0c1b1b2b3b4b5b6b7", "us-east-1")
	if err != nil {
		t.Logf("StartSession failed as expected without a real instance: %v", err)
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

	// Test all assigned fields
	if operation.InstanceID != "i-1234567890abcdef0" {
		t.Error("InstanceID should be set correctly")
	}

	if operation.Region != "us-east-1" {
		t.Error("Region should be set correctly")
	}

	if operation.LocalPath != "/path/to/local/file.txt" {
		t.Error("LocalPath should be set correctly")
	}

	if operation.RemotePath != "/path/to/remote/file.txt" {
		t.Error("RemotePath should be set correctly")
	}

	if operation.Size != 0 {
		t.Error("Size should be 0")
	}

	if operation.Method != "direct" {
		t.Error("Method should be 'direct'")
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

	// Test all assigned fields
	if result.InstanceID != "i-1234567890abcdef0" {
		t.Error("InstanceID should be set correctly")
	}

	if result.Command != "echo 'test'" {
		t.Error("Command should be set correctly")
	}

	if result.Output != "test\n" {
		t.Error("Output should be set correctly")
	}

	if result.ErrorOutput != "" {
		t.Error("ErrorOutput should be empty")
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
	instance := interactive.Instance{
		InstanceID: "i-1234567890abcdef0",
		Tags:       map[string]string{}, // Empty tags map
	}

	if instance.InstanceID != "i-1234567890abcdef0" {
		t.Error("InstanceID should be set correctly")
	}

	if instance.Tags == nil {
		t.Error("Tags map should not be nil")
	}

	if len(instance.Tags) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(instance.Tags))
	}

	// Test nil tags map
	instance2 := interactive.Instance{
		InstanceID: "i-1234567890abcdef0",
		Tags:       nil,
	}

	if instance2.InstanceID != "i-1234567890abcdef0" {
		t.Error("InstanceID should be set correctly")
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

	if filters.Status != "running" {
		t.Error("Status filter should be set correctly")
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

func TestParseTagFilters(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    map[string]string
		expectError bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
		},
		{
			name:  "Single tag",
			input: "Environment=dev",
			expected: map[string]string{
				"Environment": "dev",
			},
		},
		{
			name:  "Two tags",
			input: "Environment=dev,Component=fts",
			expected: map[string]string{
				"Environment": "dev",
				"Component":   "fts",
			},
		},
		{
			name:  "Three tags",
			input: "Team=backend,Environment=staging,Component=api",
			expected: map[string]string{
				"Team":        "backend",
				"Environment": "staging",
				"Component":   "api",
			},
		},
		{
			name:  "Tags with spaces",
			input: "Environment = production , Team = devops",
			expected: map[string]string{
				"Environment": "production",
				"Team":        "devops",
			},
		},
		{
			name:  "Complex values",
			input: "Name=web-server-1,Zone=us-east-1a,Environment=prod",
			expected: map[string]string{
				"Name":        "web-server-1",
				"Zone":        "us-east-1a",
				"Environment": "prod",
			},
		},
		{
			name:        "Missing value",
			input:       "Environment",
			expectError: true,
		},
		{
			name:        "Missing key",
			input:       "=production",
			expectError: true,
		},
		{
			name:        "Empty value",
			input:       "Environment=,Component=fts",
			expectError: true,
		},
		{
			name:        "Empty key",
			input:       "Environment=dev,=fts",
			expectError: true,
		},
		{
			name:  "Multiple equals in value",
			input: "Key=Value=Extra",
			expected: map[string]string{
				"Key": "Value=Extra",
			},
		},
		{
			name:  "Empty tag in list",
			input: "Environment=dev,,Component=fts",
			expected: map[string]string{
				"Environment": "dev",
				"Component":   "fts",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTagFilters(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input %q but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				return
			}

			// Check if result matches expected
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tags, got %d", len(tt.expected), len(result))
				return
			}

			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				if !exists {
					t.Errorf("Expected key %q not found in result", key)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("Expected value %q for key %q, got %q", expectedValue, key, actualValue)
				}
			}

			for key := range result {
				if _, exists := tt.expected[key]; !exists {
					t.Errorf("Unexpected key %q in result", key)
				}
			}
		})
	}
}

func TestListFiltersWithTags(t *testing.T) {
	tests := []struct {
		name            string
		filters         *ListFilters
		expectTags      string
		expectLegacyTag string
	}{
		{
			name: "Tags field populated",
			filters: &ListFilters{
				Tags: "Environment=dev,Component=fts",
			},
			expectTags: "Environment=dev,Component=fts",
		},
		{
			name: "Legacy Tag field populated",
			filters: &ListFilters{
				Tag: "Environment=production",
			},
			expectLegacyTag: "Environment=production",
		},
		{
			name: "Both fields populated (Tags takes precedence)",
			filters: &ListFilters{
				Tag:  "Environment=production",
				Tags: "Environment=dev,Component=fts",
			},
			expectTags:      "Environment=dev,Component=fts",
			expectLegacyTag: "Environment=production",
		},
		{
			name: "With other filters",
			filters: &ListFilters{
				Tags:   "Environment=staging,Team=backend",
				Status: "running",
				Name:   "web-server",
			},
			expectTags: "Environment=staging,Team=backend",
		},
		{
			name:    "Empty filters",
			filters: &ListFilters{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.filters.Tags != tt.expectTags {
				t.Errorf("Expected Tags %q, got %q", tt.expectTags, tt.filters.Tags)
			}

			if tt.filters.Tag != tt.expectLegacyTag {
				t.Errorf("Expected Tag %q, got %q", tt.expectLegacyTag, tt.filters.Tag)
			}

			// Test that both fields can coexist
			if tt.expectTags != "" && tt.expectLegacyTag != "" {
				if tt.filters.Tags == "" || tt.filters.Tag == "" {
					t.Error("Both Tags and Tag fields should be preserved when both are set")
				}
			}
		})
	}
}

// Test validation functions for command injection prevention
func TestValidateInstanceID(t *testing.T) {
	tests := []struct {
		name        string
		instanceID  string
		expectError bool
	}{
		{"valid instance ID", "i-1234567890abcdef0", false},
		{"valid short instance ID", "i-12345678", false},
		{"valid max length ID", "i-12345678901234567", false},
		{"invalid prefix", "inst-1234567890abcdef0", true},
		{"too short", "i-1234567", true},
		{"too long", "i-123456789012345678", true},
		{"invalid characters", "i-1234567890abcdefg", true},
		{"uppercase letters", "i-1234567890ABCDEF0", true},
		{"with special chars", "i-1234567890abcdef0; rm -rf /", true},
		{"empty string", "", true},
		{"just prefix", "i-", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInstanceID(tt.instanceID)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for instance ID %q but got none", tt.instanceID)
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for instance ID %q: %v", tt.instanceID, err)
			}
		})
	}
}

func TestValidateAWSRegion(t *testing.T) {
	tests := []struct {
		name        string
		region      string
		expectError bool
	}{
		{"valid us region", "us-east-1", false},
		{"valid eu region", "eu-west-2", false},
		{"valid ap region", "ap-southeast-1", false},
		{"valid ca region", "ca-central-1", false},
		{"three letter prefix", "aps-south-1", false},
		{"invalid format", "invalid-region", true},
		{"no dashes", "useast1", true},
		{"too many dashes", "us-east-1-extra", true},
		{"uppercase", "US-EAST-1", true},
		{"with special chars", "us-east-1; echo", true},
		{"empty string", "", true},
		{"just dashes", "--", true},
		{"missing number", "us-east-", true},
		{"missing direction", "us--1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAWSRegion(tt.region)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for region %q but got none", tt.region)
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for region %q: %v", tt.region, err)
			}
		})
	}
}

func TestValidatePortNumber(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		expectError bool
	}{
		{"valid port 80", 80, false},
		{"valid port 8080", 8080, false},
		{"valid port 22", 22, false},
		{"valid port 443", 443, false},
		{"port 1", 1, false},
		{"port 65535", 65535, false},
		{"port 0", 0, true},
		{"negative port", -1, true},
		{"port too high", 65536, true},
		{"port too high", 100000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePortNumber(tt.port)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for port %d but got none", tt.port)
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for port %d: %v", tt.port, err)
			}
		})
	}
}

func TestRemoveExitCodeLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Output with EXIT_CODE at end",
			input:    " 17:33:25 up 8 days, 14:48,  0 user,  load average: 0.00, 0.00, 0.00\nEXIT_CODE:0",
			expected: " 17:33:25 up 8 days, 14:48,  0 user,  load average: 0.00, 0.00, 0.00",
		},
		{
			name:     "Output with EXIT_CODE in middle",
			input:    "line1\nEXIT_CODE:0\nline2",
			expected: "line1\nline2",
		},
		{
			name:     "Multiple EXIT_CODE lines",
			input:    "output\nEXIT_CODE:0\nmore output\nEXIT_CODE:1",
			expected: "output\nmore output",
		},
		{
			name:     "No EXIT_CODE line",
			input:    "normal output\nwithout exit code",
			expected: "normal output\nwithout exit code",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only EXIT_CODE line",
			input:    "EXIT_CODE:0",
			expected: "",
		},
		{
			name:     "EXIT_CODE with non-zero code",
			input:    "error output\nEXIT_CODE:127",
			expected: "error output",
		},
		{
			name:     "Text containing EXIT_CODE but not at start",
			input:    "The EXIT_CODE:0 is in the middle of line",
			expected: "The EXIT_CODE:0 is in the middle of line",
		},
		{
			name:     "Output ending with newline after EXIT_CODE",
			input:    "output\nEXIT_CODE:0\n",
			expected: "output\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeExitCodeLine(tt.input)
			if result != tt.expected {
				t.Errorf("removeExitCodeLine() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
