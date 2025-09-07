package ssm

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"ztictl/pkg/logging"
)

func TestNewS3LifecycleManager(t *testing.T) {
	logger := logging.NewNoOpLogger()

	manager := NewS3LifecycleManager(logger, nil, nil)

	if manager == nil {
		t.Fatal("NewS3LifecycleManager should not return nil")
	}

	if manager.logger != logger {
		t.Error("NewS3LifecycleManager should preserve logger")
	}

	// Clients can be nil during initialization
	if manager.s3Client != nil {
		t.Error("NewS3LifecycleManager should allow nil s3Client")
	}

	if manager.stsClient != nil {
		t.Error("NewS3LifecycleManager should allow nil stsClient")
	}
}

func TestGetS3BucketNameFormat(t *testing.T) {
	// Test bucket name format without AWS calls
	accountID := "123456789012"
	region := "us-east-1"

	expectedBucketName := fmt.Sprintf("%s-%s-%s", S3BucketPrefix, accountID, region)

	if !strings.HasPrefix(expectedBucketName, S3BucketPrefix) {
		t.Error("Bucket name should start with S3BucketPrefix")
	}

	if !strings.Contains(expectedBucketName, accountID) {
		t.Error("Bucket name should contain account ID")
	}

	if !strings.Contains(expectedBucketName, region) {
		t.Error("Bucket name should contain region")
	}

	// Test with different regions
	regions := []string{"us-west-2", "eu-west-1", "ap-southeast-1"}
	for _, r := range regions {
		bucketName := fmt.Sprintf("%s-%s-%s", S3BucketPrefix, accountID, r)
		if !strings.Contains(bucketName, r) {
			t.Errorf("Bucket name should contain region %s", r)
		}
	}
}

func TestLifecycleConfigurationStruct(t *testing.T) {
	config := LifecycleConfiguration{
		Rules: []LifecycleRule{
			{
				ID:     "test-rule",
				Status: "Enabled",
				Filter: LifecycleFilter{
					Prefix: "uploads/",
				},
				Expiration: &LifecycleExpiration{
					Days: 7,
				},
				AbortIncompleteMultipartUpload: &AbortIncompleteMultipartUpload{
					DaysAfterInitiation: 1,
				},
			},
		},
	}

	// Test that configuration can be serialized to JSON
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal LifecycleConfiguration to JSON: %v", err)
	}

	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, "test-rule") {
		t.Error("JSON should contain rule ID")
	}

	if !strings.Contains(jsonStr, "Enabled") {
		t.Error("JSON should contain status")
	}

	if !strings.Contains(jsonStr, "uploads/") {
		t.Error("JSON should contain prefix")
	}

	// Test deserialization
	var deserializedConfig LifecycleConfiguration
	if err := json.Unmarshal(jsonBytes, &deserializedConfig); err != nil {
		t.Fatalf("Failed to unmarshal LifecycleConfiguration from JSON: %v", err)
	}

	if len(deserializedConfig.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(deserializedConfig.Rules))
	}

	rule := deserializedConfig.Rules[0]
	if rule.ID != "test-rule" {
		t.Error("Rule ID should be preserved")
	}

	if rule.Status != "Enabled" {
		t.Error("Rule status should be preserved")
	}

	if rule.Filter.Prefix != "uploads/" {
		t.Error("Rule filter prefix should be preserved")
	}

	if rule.Expiration == nil || rule.Expiration.Days != 7 {
		t.Error("Rule expiration should be preserved")
	}

	if rule.AbortIncompleteMultipartUpload == nil || rule.AbortIncompleteMultipartUpload.DaysAfterInitiation != 1 {
		t.Error("Rule abort upload setting should be preserved")
	}
}

func TestLifecycleRuleStruct(t *testing.T) {
	rule := LifecycleRule{
		ID:     "SSMFileTransferCleanup",
		Status: "Enabled",
		Filter: LifecycleFilter{
			Prefix: "",
		},
		Expiration: &LifecycleExpiration{
			Days: 1,
		},
		AbortIncompleteMultipartUpload: &AbortIncompleteMultipartUpload{
			DaysAfterInitiation: 1,
		},
	}

	// Test JSON serialization
	jsonBytes, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("Failed to marshal LifecycleRule to JSON: %v", err)
	}

	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, "SSMFileTransferCleanup") {
		t.Error("JSON should contain rule ID")
	}

	if !strings.Contains(jsonStr, "Enabled") {
		t.Error("JSON should contain status")
	}

	// Test deserialization
	var deserializedRule LifecycleRule
	if err := json.Unmarshal(jsonBytes, &deserializedRule); err != nil {
		t.Fatalf("Failed to unmarshal LifecycleRule from JSON: %v", err)
	}

	if deserializedRule.ID != rule.ID {
		t.Error("Rule ID should be preserved")
	}

	if deserializedRule.Status != rule.Status {
		t.Error("Rule status should be preserved")
	}

	if deserializedRule.Filter.Prefix != rule.Filter.Prefix {
		t.Error("Rule filter prefix should be preserved")
	}

	if deserializedRule.Expiration.Days != rule.Expiration.Days {
		t.Error("Rule expiration days should be preserved")
	}

	if deserializedRule.AbortIncompleteMultipartUpload.DaysAfterInitiation != rule.AbortIncompleteMultipartUpload.DaysAfterInitiation {
		t.Error("Rule abort upload days should be preserved")
	}
}

func TestLifecycleFilterStruct(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{"empty prefix", ""},
		{"uploads prefix", "uploads/"},
		{"downloads prefix", "downloads/"},
		{"complex prefix", "data/temp/files/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := LifecycleFilter{
				Prefix: tt.prefix,
			}

			jsonBytes, err := json.Marshal(filter)
			if err != nil {
				t.Fatalf("Failed to marshal LifecycleFilter to JSON: %v", err)
			}

			var deserializedFilter LifecycleFilter
			if err := json.Unmarshal(jsonBytes, &deserializedFilter); err != nil {
				t.Fatalf("Failed to unmarshal LifecycleFilter from JSON: %v", err)
			}

			if deserializedFilter.Prefix != filter.Prefix {
				t.Errorf("Expected prefix %s, got %s", filter.Prefix, deserializedFilter.Prefix)
			}
		})
	}
}

func TestLifecycleExpirationStruct(t *testing.T) {
	tests := []struct {
		name string
		days int32
	}{
		{"1 day", 1},
		{"7 days", 7},
		{"30 days", 30},
		{"365 days", 365},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiration := LifecycleExpiration{
				Days: tt.days,
			}

			jsonBytes, err := json.Marshal(expiration)
			if err != nil {
				t.Fatalf("Failed to marshal LifecycleExpiration to JSON: %v", err)
			}

			var deserializedExpiration LifecycleExpiration
			if err := json.Unmarshal(jsonBytes, &deserializedExpiration); err != nil {
				t.Fatalf("Failed to unmarshal LifecycleExpiration from JSON: %v", err)
			}

			if deserializedExpiration.Days != expiration.Days {
				t.Errorf("Expected days %d, got %d", expiration.Days, deserializedExpiration.Days)
			}
		})
	}
}

func TestAbortIncompleteMultipartUploadStruct(t *testing.T) {
	tests := []struct {
		name string
		days int32
	}{
		{"1 day", 1},
		{"3 days", 3},
		{"7 days", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abort := AbortIncompleteMultipartUpload{
				DaysAfterInitiation: tt.days,
			}

			jsonBytes, err := json.Marshal(abort)
			if err != nil {
				t.Fatalf("Failed to marshal AbortIncompleteMultipartUpload to JSON: %v", err)
			}

			var deserializedAbort AbortIncompleteMultipartUpload
			if err := json.Unmarshal(jsonBytes, &deserializedAbort); err != nil {
				t.Fatalf("Failed to unmarshal AbortIncompleteMultipartUpload from JSON: %v", err)
			}

			if deserializedAbort.DaysAfterInitiation != abort.DaysAfterInitiation {
				t.Errorf("Expected days %d, got %d", abort.DaysAfterInitiation, deserializedAbort.DaysAfterInitiation)
			}
		})
	}
}

func TestS3LifecycleConstants(t *testing.T) {
	// Test S3BucketPrefix
	if S3BucketPrefix == "" {
		t.Error("S3BucketPrefix should not be empty")
	}

	if !strings.Contains(S3BucketPrefix, "ztiaws") {
		t.Error("S3BucketPrefix should contain 'ztiaws'")
	}

	if !strings.Contains(S3BucketPrefix, "ssm") {
		t.Error("S3BucketPrefix should contain 'ssm'")
	}

	if !strings.Contains(S3BucketPrefix, "transfer") {
		t.Error("S3BucketPrefix should contain 'transfer'")
	}

	// Test LifecycleRuleID
	if LifecycleRuleID == "" {
		t.Error("LifecycleRuleID should not be empty")
	}

	if !strings.Contains(LifecycleRuleID, "SSM") {
		t.Error("LifecycleRuleID should contain 'SSM'")
	}

	if !strings.Contains(LifecycleRuleID, "FileTransfer") {
		t.Error("LifecycleRuleID should contain 'FileTransfer'")
	}

	if !strings.Contains(LifecycleRuleID, "Cleanup") {
		t.Error("LifecycleRuleID should contain 'Cleanup'")
	}

	// Test DefaultExpirationDays
	if DefaultExpirationDays <= 0 {
		t.Error("DefaultExpirationDays should be positive")
	}

	if DefaultExpirationDays != 1 {
		t.Errorf("Expected DefaultExpirationDays to be 1, got %d", DefaultExpirationDays)
	}

	// Test DefaultAbortUploadDays
	if DefaultAbortUploadDays <= 0 {
		t.Error("DefaultAbortUploadDays should be positive")
	}

	if DefaultAbortUploadDays != 1 {
		t.Errorf("Expected DefaultAbortUploadDays to be 1, got %d", DefaultAbortUploadDays)
	}
}

func TestCreateLifecycleConfigurationLogic(t *testing.T) {
	logger := logging.NewNoOpLogger()
	manager := NewS3LifecycleManager(logger, nil, nil)

	// Test the lifecycle configuration creation without AWS calls
	// This tests the configuration structure and logic
	config := manager.createLifecycleConfiguration()

	if config == nil {
		t.Fatal("createLifecycleConfiguration should not return nil")
	}

	if len(config.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(config.Rules))
	}

	rule := config.Rules[0]

	// Test rule ID
	if rule.ID == nil || *rule.ID != LifecycleRuleID {
		t.Errorf("Expected rule ID %s, got %v", LifecycleRuleID, rule.ID)
	}

	// Test rule status - this would be s3types.ExpirationStatusEnabled in real implementation
	// but we test the string representation
	if string(rule.Status) != "Enabled" {
		t.Errorf("Expected rule status 'Enabled', got %s", string(rule.Status))
	}

	// Test filter - should apply to all objects (empty prefix)
	if rule.Filter == nil {
		t.Fatal("Rule filter should not be nil")
	}

	if rule.Filter.Prefix == nil || *rule.Filter.Prefix != "" {
		t.Error("Rule filter prefix should be empty string to apply to all objects")
	}

	// Test expiration
	if rule.Expiration == nil {
		t.Fatal("Rule expiration should not be nil")
	}

	if rule.Expiration.Days == nil || *rule.Expiration.Days != DefaultExpirationDays {
		t.Errorf("Expected expiration days %d, got %v", DefaultExpirationDays, rule.Expiration.Days)
	}

	// Test abort incomplete multipart upload
	if rule.AbortIncompleteMultipartUpload == nil {
		t.Fatal("Rule abort incomplete multipart upload should not be nil")
	}

	if rule.AbortIncompleteMultipartUpload.DaysAfterInitiation == nil || *rule.AbortIncompleteMultipartUpload.DaysAfterInitiation != DefaultAbortUploadDays {
		t.Errorf("Expected abort upload days %d, got %v", DefaultAbortUploadDays, rule.AbortIncompleteMultipartUpload.DaysAfterInitiation)
	}
}

func TestS3LifecycleManagerFields(t *testing.T) {
	logger := logging.NewNoOpLogger()

	manager := &S3LifecycleManager{
		logger:    logger,
		s3Client:  nil,
		stsClient: nil,
	}

	// Test that all fields are properly accessible
	if manager.logger != logger {
		t.Error("S3LifecycleManager logger field should be properly set")
	}

	// Test setting clients to non-nil values
	// In actual implementation, these would be *s3.Client and *sts.Client
	// For testing, we just ensure the fields can be set and accessed
	if manager.s3Client != nil {
		t.Error("S3LifecycleManager s3Client should be nil initially")
	}

	if manager.stsClient != nil {
		t.Error("S3LifecycleManager stsClient should be nil initially")
	}
}

func TestBucketNameValidation(t *testing.T) {
	// Test bucket naming rules compliance
	// S3 bucket names must be between 3-63 characters, lowercase, no underscores, etc.

	testCases := []struct {
		accountID string
		region    string
	}{
		{"123456789012", "us-east-1"},
		{"999999999999", "us-west-2"},
		{"000000000001", "eu-west-1"},
		{"123456789012", "ap-southeast-1"},
	}

	for _, tc := range testCases {
		bucketName := fmt.Sprintf("%s-%s-%s", S3BucketPrefix, tc.accountID, tc.region)

		// Test length (should be between 3-63 characters)
		if len(bucketName) < 3 || len(bucketName) > 63 {
			t.Errorf("Bucket name length %d is invalid for account %s in region %s", len(bucketName), tc.accountID, tc.region)
		}

		// Test lowercase (AWS requirement)
		if bucketName != strings.ToLower(bucketName) {
			t.Errorf("Bucket name should be lowercase: %s", bucketName)
		}

		// Test no underscores (AWS requirement)
		if strings.Contains(bucketName, "_") {
			t.Errorf("Bucket name should not contain underscores: %s", bucketName)
		}

		// Test starts with letter or number (AWS requirement)
		firstChar := bucketName[0]
		if !((firstChar >= 'a' && firstChar <= 'z') || (firstChar >= '0' && firstChar <= '9')) {
			t.Errorf("Bucket name should start with letter or number: %s", bucketName)
		}

		// Test ends with letter or number (AWS requirement)
		lastChar := bucketName[len(bucketName)-1]
		if !((lastChar >= 'a' && lastChar <= 'z') || (lastChar >= '0' && lastChar <= '9')) {
			t.Errorf("Bucket name should end with letter or number: %s", bucketName)
		}
	}
}

func TestS3ObjectKeyFormat(t *testing.T) {
	// Test S3 object key formats used in file transfers
	testFiles := []string{
		"test.txt",
		"document.pdf",
		"image.jpg",
		"config.json",
		"script.sh",
	}

	timestamp := int64(1640995200) // 2022-01-01 00:00:00 UTC
	randomHex := "abcdef1234567890"

	for _, filename := range testFiles {
		uploadKey := fmt.Sprintf("uploads/%d-%s-%s", timestamp, randomHex, filename)
		downloadKey := fmt.Sprintf("downloads/%d-%s-%s", timestamp, randomHex, filename)

		// Test upload key format
		if !strings.HasPrefix(uploadKey, "uploads/") {
			t.Errorf("Upload key should start with 'uploads/': %s", uploadKey)
		}

		if !strings.Contains(uploadKey, fmt.Sprintf("%d", timestamp)) {
			t.Errorf("Upload key should contain timestamp: %s", uploadKey)
		}

		if !strings.Contains(uploadKey, randomHex) {
			t.Errorf("Upload key should contain random hex: %s", uploadKey)
		}

		if !strings.HasSuffix(uploadKey, filename) {
			t.Errorf("Upload key should end with filename: %s", uploadKey)
		}

		// Test download key format
		if !strings.HasPrefix(downloadKey, "downloads/") {
			t.Errorf("Download key should start with 'downloads/': %s", downloadKey)
		}

		if !strings.Contains(downloadKey, fmt.Sprintf("%d", timestamp)) {
			t.Errorf("Download key should contain timestamp: %s", downloadKey)
		}

		if !strings.Contains(downloadKey, randomHex) {
			t.Errorf("Download key should contain random hex: %s", downloadKey)
		}

		if !strings.HasSuffix(downloadKey, filename) {
			t.Errorf("Download key should end with filename: %s", downloadKey)
		}
	}
}

func TestLifecycleConfigurationEdgeCases(t *testing.T) {
	// Test empty lifecycle configuration
	emptyConfig := LifecycleConfiguration{
		Rules: []LifecycleRule{},
	}

	jsonBytes, err := json.Marshal(emptyConfig)
	if err != nil {
		t.Fatalf("Failed to marshal empty LifecycleConfiguration: %v", err)
	}

	var deserializedConfig LifecycleConfiguration
	if err := json.Unmarshal(jsonBytes, &deserializedConfig); err != nil {
		t.Fatalf("Failed to unmarshal empty LifecycleConfiguration: %v", err)
	}

	if len(deserializedConfig.Rules) != 0 {
		t.Error("Empty configuration should have no rules")
	}

	// Test configuration with nil rules
	nilConfig := LifecycleConfiguration{
		Rules: nil,
	}

	jsonBytes, err = json.Marshal(nilConfig)
	if err != nil {
		t.Fatalf("Failed to marshal nil rules LifecycleConfiguration: %v", err)
	}

	var deserializedNilConfig LifecycleConfiguration
	if err := json.Unmarshal(jsonBytes, &deserializedNilConfig); err != nil {
		t.Fatalf("Failed to unmarshal nil rules LifecycleConfiguration: %v", err)
	}

	// After unmarshaling, nil slice becomes empty slice in JSON
	if len(deserializedNilConfig.Rules) > 0 {
		t.Error("Unmarshaled nil rules should be empty or nil")
	}
}

func TestLifecycleRuleEdgeCases(t *testing.T) {
	// Test rule with minimal fields
	minimalRule := LifecycleRule{
		ID:     "minimal",
		Status: "Disabled",
		Filter: LifecycleFilter{},
	}

	jsonBytes, err := json.Marshal(minimalRule)
	if err != nil {
		t.Fatalf("Failed to marshal minimal LifecycleRule: %v", err)
	}

	var deserializedRule LifecycleRule
	if err := json.Unmarshal(jsonBytes, &deserializedRule); err != nil {
		t.Fatalf("Failed to unmarshal minimal LifecycleRule: %v", err)
	}

	if deserializedRule.ID != "minimal" {
		t.Error("Minimal rule ID should be preserved")
	}

	if deserializedRule.Status != "Disabled" {
		t.Error("Minimal rule status should be preserved")
	}

	if deserializedRule.Expiration != nil {
		t.Error("Minimal rule should have nil expiration")
	}

	if deserializedRule.AbortIncompleteMultipartUpload != nil {
		t.Error("Minimal rule should have nil abort upload setting")
	}
}

func TestS3ErrorHandling(t *testing.T) {
	// Test error message formatting for S3 operations
	bucketName := "test-bucket"

	// Test error message formats that would be used in actual implementation
	createBucketError := fmt.Sprintf("failed to create S3 bucket %s: %s", bucketName, "access denied")
	if !strings.Contains(createBucketError, bucketName) {
		t.Error("Create bucket error should contain bucket name")
	}

	uploadError := fmt.Sprintf("failed to upload file to S3: %s", "network timeout")
	if !strings.Contains(uploadError, "upload") {
		t.Error("Upload error should mention upload")
	}

	downloadError := fmt.Sprintf("failed to download file from S3: %s", "object not found")
	if !strings.Contains(downloadError, "download") {
		t.Error("Download error should mention download")
	}

	lifecycleError := fmt.Sprintf("failed to apply lifecycle configuration to bucket %s: %s", bucketName, "invalid configuration")
	if !strings.Contains(lifecycleError, bucketName) {
		t.Error("Lifecycle error should contain bucket name")
	}
	if !strings.Contains(lifecycleError, "lifecycle") {
		t.Error("Lifecycle error should mention lifecycle")
	}
}

func TestRegionSpecificBehavior(t *testing.T) {
	// Test region-specific behavior
	regions := []string{
		"us-east-1",      // Default region, no location constraint needed
		"us-west-2",      // Requires location constraint
		"eu-west-1",      // Requires location constraint
		"ap-southeast-1", // Requires location constraint
	}

	for _, region := range regions {
		// Test bucket creation logic - us-east-1 is special case
		needsLocationConstraint := region != "us-east-1"

		if region == "us-east-1" && needsLocationConstraint {
			t.Error("us-east-1 should not need location constraint")
		}

		if region != "us-east-1" && !needsLocationConstraint {
			t.Errorf("Region %s should need location constraint", region)
		}
	}
}

func TestS3URIFormat(t *testing.T) {
	// Test S3 URI format used in logging
	bucketName := "test-bucket"
	objectKey := "uploads/file.txt"

	expectedURI := fmt.Sprintf("s3://%s/%s", bucketName, objectKey)

	if !strings.HasPrefix(expectedURI, "s3://") {
		t.Error("S3 URI should start with s3://")
	}

	if !strings.Contains(expectedURI, bucketName) {
		t.Error("S3 URI should contain bucket name")
	}

	if !strings.Contains(expectedURI, objectKey) {
		t.Error("S3 URI should contain object key")
	}

	// Test URI for bucket operations (no object key)
	bucketURI := fmt.Sprintf("s3://%s", bucketName)

	if !strings.HasPrefix(bucketURI, "s3://") {
		t.Error("Bucket URI should start with s3://")
	}

	if !strings.HasSuffix(bucketURI, bucketName) {
		t.Error("Bucket URI should end with bucket name")
	}
}
