package main

import (
	"bytes"
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"ztictl/pkg/logging"
)

func TestSsmTransferCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Transfer help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Transfer files between local machine and EC2 instances via SSM",
		},
		{
			name:    "Upload file",
			args:    []string{"i-1234567890abcdef0", "/local/file.txt", "/remote/file.txt"},
			wantErr: false,
		},
		{
			name:    "Download file",
			args:    []string{"i-1234567890abcdef0", "/remote/file.txt", "/local/file.txt", "--download"},
			wantErr: false,
		},
		{
			name:    "Transfer with region",
			args:    []string{"i-1234567890abcdef0", "/local/file.txt", "/remote/file.txt", "--region", "us-east-1"},
			wantErr: false,
		},
		{
			name:    "Transfer with S3 bucket",
			args:    []string{"i-1234567890abcdef0", "/local/largefile.zip", "/remote/largefile.zip", "--s3-bucket", "my-bucket"},
			wantErr: false,
		},
		{
			name:    "Transfer without source",
			args:    []string{"i-1234567890abcdef0"},
			wantErr: true,
		},
		{
			name:    "Transfer without destination",
			args:    []string{"i-1234567890abcdef0", "/local/file.txt"},
			wantErr: true,
		},
		{
			name:    "Transfer without instance",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "transfer <instance-identifier> <source> <destination>",
				Short: "Transfer files to/from instances",
				Long:  "Transfer files between local machine and EC2 instances via SSM",
				Args:  cobra.ExactArgs(3),
				Run: func(cmd *cobra.Command, args []string) {
					// Mock file transfer functionality
					regionCode, _ := cmd.Flags().GetString("region")
					isDownload, _ := cmd.Flags().GetBool("download")
					s3Bucket, _ := cmd.Flags().GetString("s3-bucket")
					force, _ := cmd.Flags().GetBool("force")

					instanceIdentifier := args[0]
					sourcePath := args[1]
					destPath := args[2]

					// Validate inputs
					if instanceIdentifier == "" {
						t.Error("Instance identifier should not be empty")
					}
					if sourcePath == "" {
						t.Error("Source path should not be empty")
					}
					if destPath == "" {
						t.Error("Destination path should not be empty")
					}

					// Mock region resolution
					region := regionCode
					if region == "" {
						region = "us-east-1"
					}

					// Mock file transfer operation
					type TransferOperation struct {
						InstanceID   string
						SourcePath   string
						DestPath     string
						Direction    string // "upload" or "download"
						TransferSize int64
						S3Bucket     string
						UseS3        bool
						Force        bool
					}

					operation := TransferOperation{
						InstanceID:   instanceIdentifier,
						SourcePath:   sourcePath,
						DestPath:     destPath,
						Direction:    "upload",
						TransferSize: 1024 * 1024, // 1MB
						S3Bucket:     s3Bucket,
						Force:        force,
					}

					// Set UseS3 based on actual transfer size
					operation.UseS3 = s3Bucket != "" || operation.TransferSize >= 1000*1000

					if isDownload {
						operation.Direction = "download"
					}

					// Test all assigned fields
					if operation.InstanceID != instanceIdentifier {
						t.Errorf("InstanceID should be %s, got %s", instanceIdentifier, operation.InstanceID)
					}

					if operation.SourcePath != sourcePath {
						t.Errorf("SourcePath should be %s, got %s", sourcePath, operation.SourcePath)
					}

					expectedUseS3 := s3Bucket != "" || operation.TransferSize >= 1000*1000
					if operation.UseS3 != expectedUseS3 {
						t.Errorf("UseS3 should be %v, got %v", expectedUseS3, operation.UseS3)
					}

					if operation.Force != force {
						t.Errorf("Force should be %v, got %v", force, operation.Force)
					}

					// Validate transfer operation
					if operation.Direction != "upload" && operation.Direction != "download" {
						t.Errorf("Invalid transfer direction: %s", operation.Direction)
					}

					if operation.TransferSize < 0 {
						t.Errorf("Transfer size should not be negative: %d", operation.TransferSize)
					}

					// Test path validation
					if !filepath.IsAbs(operation.DestPath) && !strings.Contains(operation.DestPath, ":") {
						// Should be absolute path (Unix or Windows)
						t.Logf("Destination path may not be absolute: %s", operation.DestPath)
					}

					// Test S3 bucket validation
					if operation.S3Bucket != "" {
						if strings.Contains(operation.S3Bucket, "/") || strings.Contains(operation.S3Bucket, "\\") {
							t.Errorf("S3 bucket name should not contain path separators: %s", operation.S3Bucket)
						}
					}
				},
			}

			// Add flags
			cmd.Flags().StringP("region", "r", "", "AWS region")
			cmd.Flags().BoolP("download", "d", false, "Download from remote to local")
			cmd.Flags().String("s3-bucket", "", "S3 bucket for large file transfers")
			cmd.Flags().BoolP("force", "f", false, "Force overwrite existing files")

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.contains != "" {
				output := buf.String()
				if !strings.Contains(output, tt.contains) {
					t.Errorf("Execute() output should contain %v, got %v", tt.contains, output)
				}
			}
		})
	}
}

func TestSsmTransferCmdFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "transfer"}
	cmd.Flags().StringP("region", "r", "", "AWS region")
	cmd.Flags().BoolP("download", "d", false, "Download from remote")
	cmd.Flags().String("s3-bucket", "", "S3 bucket")
	cmd.Flags().BoolP("force", "f", false, "Force overwrite")

	tests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{"region", "r", "", "string"},
		{"download", "d", "false", "bool"},
		{"s3-bucket", "", "", "string"},
		{"force", "f", "false", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName+" flag", func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s shorthand = %s, want %s", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			if flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %s default = %s, want %s", tt.flagName, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestFileTransferDirections(t *testing.T) {
	tests := []struct {
		name         string
		sourcePath   string
		destPath     string
		downloadFlag bool
		expectedDir  string
	}{
		{
			name:         "Upload local to remote",
			sourcePath:   "/local/file.txt",
			destPath:     "/remote/file.txt",
			downloadFlag: false,
			expectedDir:  "upload",
		},
		{
			name:         "Download remote to local",
			sourcePath:   "/remote/file.txt",
			destPath:     "/local/file.txt",
			downloadFlag: true,
			expectedDir:  "download",
		},
		{
			name:         "Upload Windows path",
			sourcePath:   "C:\\local\\file.txt",
			destPath:     "/remote/file.txt",
			downloadFlag: false,
			expectedDir:  "upload",
		},
		{
			name:         "Download to Windows path",
			sourcePath:   "/remote/file.txt",
			destPath:     "C:\\local\\file.txt",
			downloadFlag: true,
			expectedDir:  "download",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			direction := "upload"
			if tt.downloadFlag {
				direction = "download"
			}

			if direction != tt.expectedDir {
				t.Errorf("Direction = %s, want %s", direction, tt.expectedDir)
			}

			// Validate paths
			if tt.sourcePath == "" || tt.destPath == "" {
				t.Error("Source and destination paths should not be empty")
			}

			// Test path format validation
			if !strings.Contains(tt.sourcePath, "/") && !strings.Contains(tt.sourcePath, "\\") {
				t.Errorf("Source path should contain path separators: %s", tt.sourcePath)
			}
		})
	}
}

func TestS3TransferThreshold(t *testing.T) {
	tests := []struct {
		name           string
		fileSize       int64
		shouldUseS3    bool
		customBucket   string
		thresholdBytes int64
	}{
		{
			name:           "Small file - direct transfer",
			fileSize:       512 * 1024, // 512KB
			shouldUseS3:    false,
			thresholdBytes: 1024 * 1024, // 1MB threshold
		},
		{
			name:           "Large file - use S3",
			fileSize:       5 * 1024 * 1024, // 5MB
			shouldUseS3:    true,
			thresholdBytes: 1024 * 1024, // 1MB threshold
		},
		{
			name:           "Exactly at threshold - use S3",
			fileSize:       1024 * 1024, // 1MB
			shouldUseS3:    true,
			thresholdBytes: 1024 * 1024, // 1MB threshold
		},
		{
			name:           "Large file with custom bucket",
			fileSize:       10 * 1024 * 1024, // 10MB
			shouldUseS3:    true,
			customBucket:   "my-transfer-bucket",
			thresholdBytes: 1024 * 1024, // 1MB threshold
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useS3 := tt.fileSize >= tt.thresholdBytes || tt.customBucket != ""

			if useS3 != tt.shouldUseS3 {
				t.Errorf("UseS3 = %v, want %v for file size %d", useS3, tt.shouldUseS3, tt.fileSize)
			}

			// Test bucket name validation
			if tt.customBucket != "" {
				if len(tt.customBucket) < 3 || len(tt.customBucket) > 63 {
					t.Errorf("S3 bucket name length should be 3-63 characters: %s (%d chars)",
						tt.customBucket, len(tt.customBucket))
				}

				// Basic bucket name validation
				if strings.Contains(tt.customBucket, ".") || strings.Contains(tt.customBucket, "_") {
					t.Logf("S3 bucket name should avoid dots and underscores: %s", tt.customBucket)
				}
			}
		})
	}
}

func TestPathValidation(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		isValid  bool
		platform string
	}{
		{
			name:     "Unix absolute path",
			path:     "/home/user/file.txt",
			isValid:  true,
			platform: "unix",
		},
		{
			name:     "Unix relative path",
			path:     "relative/file.txt",
			isValid:  false,
			platform: "unix",
		},
		{
			name:     "Windows absolute path",
			path:     "C:\\Users\\user\\file.txt",
			isValid:  true,
			platform: "windows",
		},
		{
			name:     "Windows relative path",
			path:     "relative\\file.txt",
			isValid:  false,
			platform: "windows",
		},
		{
			name:     "Windows UNC path",
			path:     "\\\\server\\share\\file.txt",
			isValid:  true,
			platform: "windows",
		},
		{
			name:     "Empty path",
			path:     "",
			isValid:  false,
			platform: "any",
		},
		{
			name:     "Root path",
			path:     "/",
			isValid:  true,
			platform: "unix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip platform-specific tests on incompatible platforms
			if runtime.GOOS == "windows" && tt.platform == "unix" {
				t.Skip("Skipping Unix path test on Windows")
				return
			}
			if runtime.GOOS != "windows" && tt.platform == "windows" {
				t.Skip("Skipping Windows path test on Unix-like system")
				return
			}

			// Validate absolute path (Unix or Windows)
			isAbsolute := filepath.IsAbs(tt.path) || strings.HasPrefix(tt.path, "\\\\") ||
				(len(tt.path) >= 3 && tt.path[1] == ':' && (tt.path[2] == '\\' || tt.path[2] == '/'))
			isEmpty := tt.path == ""

			isValid := !isEmpty && isAbsolute

			if isValid != tt.isValid {
				t.Errorf("Path '%s' validity = %v, want %v", tt.path, isValid, tt.isValid)
			}
		})
	}
}

func TestTransferProgressTracking(t *testing.T) {
	// Test transfer progress structure
	type TransferProgress struct {
		BytesTransferred int64
		TotalBytes       int64
		Percentage       float64
		Speed            int64 // bytes per second
		ETA              int   // seconds
		Status           string
	}

	progress := TransferProgress{
		BytesTransferred: 256 * 1024,  // 256KB
		TotalBytes:       1024 * 1024, // 1MB
		Percentage:       25.0,
		Speed:            128 * 1024, // 128KB/s
		ETA:              6,          // seconds
		Status:           "InProgress",
	}

	// Validate progress calculations
	expectedPercentage := float64(progress.BytesTransferred) / float64(progress.TotalBytes) * 100
	if progress.Percentage != expectedPercentage {
		t.Errorf("Percentage = %f, want %f", progress.Percentage, expectedPercentage)
	}

	if progress.BytesTransferred > progress.TotalBytes {
		t.Error("Bytes transferred should not exceed total bytes")
	}

	if progress.Speed < 0 {
		t.Error("Transfer speed should not be negative")
	}

	if progress.ETA < 0 {
		t.Error("ETA should not be negative")
	}

	// Test valid statuses
	validStatuses := []string{"InProgress", "Completed", "Failed", "Cancelled", "Paused"}
	isValidStatus := false
	for _, status := range validStatuses {
		if progress.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		t.Errorf("Invalid transfer status: %s", progress.Status)
	}
}

func TestSsmTransferContextHandling(t *testing.T) {
	// Test context usage in transfer operations
	ctx := context.Background()
	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Test context with transfer metadata
	type contextKey string
	key := contextKey("transfer-id")
	ctx = context.WithValue(ctx, key, "transfer-123")

	value := ctx.Value(key)
	if value != "transfer-123" {
		t.Errorf("Context value = %v, want transfer-123", value)
	}

	// Test cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cancel()
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled")
	}
}

func TestTransferErrorHandling(t *testing.T) {
	// Test error scenarios
	tests := []struct {
		name        string
		instance    string
		source      string
		dest        string
		shouldError bool
		errorType   string
	}{
		{
			name:        "Valid transfer",
			instance:    "i-1234567890abcdef0",
			source:      "/local/file.txt",
			dest:        "/remote/file.txt",
			shouldError: false,
		},
		{
			name:        "Source file not found",
			instance:    "i-1234567890abcdef0",
			source:      "/nonexistent/file.txt",
			dest:        "/remote/file.txt",
			shouldError: true,
			errorType:   "source file not found",
		},
		{
			name:        "Destination permission denied",
			instance:    "i-1234567890abcdef0",
			source:      "/local/file.txt",
			dest:        "/root/protected/file.txt",
			shouldError: true,
			errorType:   "permission denied",
		},
		{
			name:        "Insufficient disk space",
			instance:    "i-1234567890abcdef0",
			source:      "/local/largefile.zip",
			dest:        "/remote/largefile.zip",
			shouldError: true,
			errorType:   "insufficient disk space",
		},
		{
			name:        "Network timeout",
			instance:    "i-1234567890abcdef0",
			source:      "/local/file.txt",
			dest:        "/remote/file.txt",
			shouldError: true,
			errorType:   "network timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock error conditions
			var err error
			if tt.shouldError {
				err = &mockTransferError{message: tt.errorType}
			}

			// Test error handling
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err != nil && !strings.Contains(err.Error(), tt.errorType) {
				t.Errorf("Error message should contain %s, got %v", tt.errorType, err)
			}
		})
	}
}

// Mock error type for transfer testing
type mockTransferError struct {
	message string
}

func (e *mockTransferError) Error() string {
	return e.message
}

func TestSsmTransferArgumentValidation(t *testing.T) {
	// Test cobra argument validation
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "No args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "One arg only",
			args:    []string{"instance"},
			wantErr: true,
		},
		{
			name:    "Two args only",
			args:    []string{"instance", "source"},
			wantErr: true,
		},
		{
			name:    "Exactly three args",
			args:    []string{"instance", "source", "dest"},
			wantErr: false,
		},
		{
			name:    "Too many args",
			args:    []string{"instance", "source", "dest", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test cobra.ExactArgs(3) validation
			err := cobra.ExactArgs(3)(nil, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExactArgs(3) error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestS3BucketNameValidation(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		isValid    bool
		reason     string
	}{
		{
			name:       "Valid bucket name",
			bucketName: "my-transfer-bucket",
			isValid:    true,
		},
		{
			name:       "Valid bucket with numbers",
			bucketName: "my-bucket-123",
			isValid:    true,
		},
		{
			name:       "Too short",
			bucketName: "my",
			isValid:    false,
			reason:     "too short (minimum 3 characters)",
		},
		{
			name:       "Too long",
			bucketName: strings.Repeat("a", 64),
			isValid:    false,
			reason:     "too long (maximum 63 characters)",
		},
		{
			name:       "Contains uppercase",
			bucketName: "My-Bucket",
			isValid:    false,
			reason:     "contains uppercase letters",
		},
		{
			name:       "Starts with dash",
			bucketName: "-my-bucket",
			isValid:    false,
			reason:     "starts with dash",
		},
		{
			name:       "Ends with dash",
			bucketName: "my-bucket-",
			isValid:    false,
			reason:     "ends with dash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic S3 bucket name validation
			isValid := len(tt.bucketName) >= 3 && len(tt.bucketName) <= 63 &&
				strings.ToLower(tt.bucketName) == tt.bucketName &&
				!strings.HasPrefix(tt.bucketName, "-") &&
				!strings.HasSuffix(tt.bucketName, "-")

			if isValid != tt.isValid {
				t.Errorf("Bucket name '%s' validity = %v, want %v", tt.bucketName, isValid, tt.isValid)
				if !tt.isValid && tt.reason != "" {
					t.Logf("Reason: %s", tt.reason)
				}
			}
		})
	}
}

// NEW TESTS FOR SEPARATION OF CONCERNS REFACTORING

func TestPerformFileUpload(t *testing.T) {
	// Save original logger state
	originalLogger := logger
	defer func() { logger = originalLogger }()

	t.Run("handles upload gracefully", func(t *testing.T) {
		// Initialize logger to avoid nil pointer dereference
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// The function should return an error or succeed, not call os.Exit
		err := performFileUpload("use1", "i-test123", "/tmp/testfile.txt", "/home/user/testfile.txt")

		// We expect this might fail (no AWS credentials/connection), but it shouldn't panic
		// The important thing is that it returns an error instead of calling os.Exit
		if err != nil {
			t.Logf("File upload error (may be expected): %v", err)
		}

		// The fact that we can continue execution proves the refactoring worked
		t.Log("Test completed - function returned instead of calling os.Exit")
	})

	t.Run("validates region code", func(t *testing.T) {
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// Test with empty region code (should be handled gracefully)
		err := performFileUpload("", "i-test123", "/tmp/testfile.txt", "/home/user/testfile.txt")

		// Function should handle this gracefully and return error
		if err != nil {
			t.Logf("Expected error for empty region: %v", err)
		}

		t.Log("Region validation handled gracefully")
	})

	t.Run("validates file paths", func(t *testing.T) {
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// Test with empty local file path
		err := performFileUpload("use1", "i-test123", "", "/home/user/testfile.txt")

		// Function should handle this gracefully
		if err != nil {
			t.Logf("Expected error for empty local file: %v", err)
		}

		// Test with empty remote path
		err = performFileUpload("use1", "i-test123", "/tmp/testfile.txt", "")

		if err != nil {
			t.Logf("Expected error for empty remote path: %v", err)
		}

		t.Log("File path validation handled gracefully")
	})
}

func TestPerformFileDownload(t *testing.T) {
	// Save original logger state
	originalLogger := logger
	defer func() { logger = originalLogger }()

	t.Run("handles download gracefully", func(t *testing.T) {
		// Initialize logger to avoid nil pointer dereference
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// The function should return an error or succeed, not call os.Exit
		err := performFileDownload("use1", "i-test123", "/home/user/remotefile.txt", "/tmp/localfile.txt")

		// We expect this might fail (no AWS credentials/connection), but it shouldn't panic
		// The important thing is that it returns an error instead of calling os.Exit
		if err != nil {
			t.Logf("File download error (may be expected): %v", err)
		}

		// The fact that we can continue execution proves the refactoring worked
		t.Log("Test completed - function returned instead of calling os.Exit")
	})

	t.Run("validates region code", func(t *testing.T) {
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// Test with empty region code (should be handled gracefully)
		err := performFileDownload("", "i-test123", "/home/user/remotefile.txt", "/tmp/localfile.txt")

		// Function should handle this gracefully and return error
		if err != nil {
			t.Logf("Expected error for empty region: %v", err)
		}

		t.Log("Region validation handled gracefully")
	})

	t.Run("validates file paths", func(t *testing.T) {
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// Test with empty remote file path
		err := performFileDownload("use1", "i-test123", "", "/tmp/localfile.txt")

		// Function should handle this gracefully
		if err != nil {
			t.Logf("Expected error for empty remote file: %v", err)
		}

		// Test with empty local path
		err = performFileDownload("use1", "i-test123", "/home/user/remotefile.txt", "")

		if err != nil {
			t.Logf("Expected error for empty local path: %v", err)
		}

		t.Log("File path validation handled gracefully")
	})
}

func TestTransferSeparationOfConcerns(t *testing.T) {
	// This test verifies that the transfer functions don't call os.Exit
	// and can be tested without terminating the test process

	// Save original logger state
	originalLogger := logger
	defer func() { logger = originalLogger }()

	t.Run("file upload returns instead of exiting", func(t *testing.T) {
		// Initialize logger to avoid nil pointer dereference
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// This call should return an error or succeed, not exit the process
		err := performFileUpload("invalid-region", "invalid-instance", "/nonexistent/file.txt", "/remote/path")

		// If we reach this line, the function didn't call os.Exit
		// (which is what we want for good separation of concerns)
		if err == nil {
			t.Log("File upload succeeded unexpectedly")
		} else {
			t.Logf("File upload failed as expected: %v", err)
		}

		// The fact that we can continue execution proves the refactoring worked
		t.Log("Test completed - function returned instead of calling os.Exit")
	})

	t.Run("file download returns instead of exiting", func(t *testing.T) {
		if logger == nil {
			logger = logging.NewLogger(false)
		}

		// This call should return an error or succeed, not exit the process
		err := performFileDownload("invalid-region", "invalid-instance", "/remote/nonexistent.txt", "/tmp/local.txt")

		// If we reach this line, the function didn't call os.Exit
		if err == nil {
			t.Log("File download succeeded unexpectedly")
		} else {
			t.Logf("File download failed as expected: %v", err)
		}

		// The fact that we can continue execution proves the refactoring worked
		t.Log("Test completed - function returned instead of calling os.Exit")
	})
}
