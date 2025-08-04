package ssm

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"ztictl/pkg/logging"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// S3LifecycleManager handles S3 bucket lifecycle management
type S3LifecycleManager struct {
	logger    *logging.Logger
	s3Client  *s3.Client
	stsClient *sts.Client
}

// LifecycleConfiguration represents the S3 lifecycle configuration
type LifecycleConfiguration struct {
	Rules []LifecycleRule `json:"Rules"`
}

type LifecycleRule struct {
	ID                             string                          `json:"ID"`
	Status                         string                          `json:"Status"`
	Filter                         LifecycleFilter                 `json:"Filter"`
	Expiration                     *LifecycleExpiration            `json:"Expiration,omitempty"`
	AbortIncompleteMultipartUpload *AbortIncompleteMultipartUpload `json:"AbortIncompleteMultipartUpload,omitempty"`
}

type LifecycleFilter struct {
	Prefix string `json:"Prefix"`
}

type LifecycleExpiration struct {
	Days int32 `json:"Days"`
}

type AbortIncompleteMultipartUpload struct {
	DaysAfterInitiation int32 `json:"DaysAfterInitiation"`
}

const (
	S3BucketPrefix         = "ztiaws-ssm-transfer"
	LifecycleRuleID        = "SSMFileTransferCleanup"
	DefaultExpirationDays  = 1
	DefaultAbortUploadDays = 1
)

// NewS3LifecycleManager creates a new S3 lifecycle manager
func NewS3LifecycleManager(logger *logging.Logger, s3Client *s3.Client, stsClient *sts.Client) *S3LifecycleManager {
	return &S3LifecycleManager{
		logger:    logger,
		s3Client:  s3Client,
		stsClient: stsClient,
	}
}

// GetS3BucketName generates a unique S3 bucket name for this AWS account
func (m *S3LifecycleManager) GetS3BucketName(ctx context.Context, region string) (string, error) {
	// Get AWS account ID
	result, err := m.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("failed to get AWS account ID: %w", err)
	}

	accountID := *result.Account
	bucketName := fmt.Sprintf("%s-%s-%s", S3BucketPrefix, accountID, region)

	return bucketName, nil
}

// createLifecycleConfiguration creates the lifecycle configuration for auto-cleanup
func (m *S3LifecycleManager) createLifecycleConfiguration() *s3types.BucketLifecycleConfiguration {
	return &s3types.BucketLifecycleConfiguration{
		Rules: []s3types.LifecycleRule{
			{
				ID:     aws.String(LifecycleRuleID),
				Status: s3types.ExpirationStatusEnabled,
				Filter: &s3types.LifecycleRuleFilter{
					Prefix: aws.String(""), // Apply to all objects
				},
				Expiration: &s3types.LifecycleExpiration{
					Days: aws.Int32(DefaultExpirationDays),
				},
				AbortIncompleteMultipartUpload: &s3types.AbortIncompleteMultipartUpload{
					DaysAfterInitiation: aws.Int32(DefaultAbortUploadDays),
				},
			},
		},
	}
}

// ApplyLifecycleConfig applies lifecycle configuration to ensure auto-cleanup
func (m *S3LifecycleManager) ApplyLifecycleConfig(ctx context.Context, bucketName, region string) error {
	m.logger.Debug("Applying lifecycle configuration to bucket: %s", bucketName)

	lifecycleConfig := m.createLifecycleConfiguration()

	_, err := m.s3Client.PutBucketLifecycleConfiguration(ctx, &s3.PutBucketLifecycleConfigurationInput{
		Bucket:                 aws.String(bucketName),
		LifecycleConfiguration: lifecycleConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to apply lifecycle configuration to bucket %s: %w", bucketName, err)
	}

	m.logger.Debug("Lifecycle configuration applied successfully to bucket: %s", bucketName)
	return nil
}

// VerifyLifecycleConfig verifies that lifecycle configuration is active
func (m *S3LifecycleManager) VerifyLifecycleConfig(ctx context.Context, bucketName, region string) error {
	m.logger.Debug("Verifying lifecycle configuration for bucket: %s", bucketName)

	result, err := m.s3Client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// Check if it's a NoSuchLifecycleConfiguration error
		if strings.Contains(err.Error(), "NoSuchLifecycleConfiguration") {
			m.logger.Debug("No lifecycle configuration found for bucket: %s", bucketName)
			return fmt.Errorf("no lifecycle configuration found")
		}
		return fmt.Errorf("failed to get lifecycle configuration: %w", err)
	}

	// Check if our specific rule exists and is enabled
	for _, rule := range result.Rules {
		if rule.ID != nil && *rule.ID == LifecycleRuleID {
			if rule.Status == s3types.ExpirationStatusEnabled {
				m.logger.Debug("Lifecycle configuration verified as enabled for bucket: %s", bucketName)
				return nil
			} else {
				return fmt.Errorf("lifecycle configuration exists but is not enabled for bucket: %s", bucketName)
			}
		}
	}

	return fmt.Errorf("lifecycle rule %s not found for bucket: %s", LifecycleRuleID, bucketName)
}

// EnsureS3Bucket creates S3 bucket if it doesn't exist and ensures lifecycle configuration
func (m *S3LifecycleManager) EnsureS3Bucket(ctx context.Context, bucketName, region string) error {
	m.logger.Info("Checking S3 bucket: %s", bucketName)

	bucketCreated := false

	// Check if bucket exists
	_, err := m.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		m.logger.Info("Creating S3 bucket: %s", bucketName)

		// Create bucket with appropriate configuration
		createBucketInput := &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		}

		// For regions other than us-east-1, we need to specify the location constraint
		if region != "us-east-1" {
			createBucketInput.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
				LocationConstraint: s3types.BucketLocationConstraint(region),
			}
		}

		_, err = m.s3Client.CreateBucket(ctx, createBucketInput)
		if err != nil {
			return fmt.Errorf("failed to create S3 bucket %s: %w", bucketName, err)
		}

		bucketCreated = true
	} else {
		m.logger.Info("S3 bucket already exists: %s", bucketName)
	}

	// Ensure lifecycle configuration is applied (for both existing and new buckets)
	if err := m.VerifyLifecycleConfig(ctx, bucketName, region); err != nil {
		m.logger.Info("Applying lifecycle configuration for auto-cleanup...")
		if err := m.ApplyLifecycleConfig(ctx, bucketName, region); err != nil {
			if bucketCreated {
				// Clean up the bucket we just created since it's not properly configured
				m.logger.Error("Failed to apply lifecycle configuration to newly created bucket")
				m.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
					Bucket: aws.String(bucketName),
				})
				return fmt.Errorf("failed to apply lifecycle configuration to newly created bucket: %w", err)
			} else {
				// For existing buckets, we'll continue even if lifecycle config fails
				// as the bucket may have other lifecycle rules or permissions issues
				m.logger.Warn("Failed to apply lifecycle configuration to existing bucket (continuing anyway)", "error", err)
			}
		}
	} else {
		m.logger.Debug("Lifecycle configuration already properly configured for bucket: %s", bucketName)
	}

	if bucketCreated {
		m.logger.Info("S3 bucket created successfully with lifecycle configuration: %s", bucketName)
	} else {
		m.logger.Info("S3 bucket verified and configured: %s", bucketName)
	}

	return nil
}

// CleanupS3Object removes an object from S3 bucket
func (m *S3LifecycleManager) CleanupS3Object(ctx context.Context, bucketName, objectKey, region string) error {
	m.logger.Debug("Cleaning up S3 object: s3://%s/%s", bucketName, objectKey)

	_, err := m.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		// Don't fail the entire operation for cleanup errors
		m.logger.Warn("Failed to cleanup S3 object", "bucket", bucketName, "key", objectKey, "error", err)
		return nil
	}

	m.logger.Debug("Successfully cleaned up S3 object: s3://%s/%s", bucketName, objectKey)
	return nil
}

// UploadToS3 uploads a file to S3
func (m *S3LifecycleManager) UploadToS3(ctx context.Context, bucketName, objectKey, filePath, region string) error {
	m.logger.Info("Uploading to S3: s3://%s/%s", bucketName, objectKey)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = m.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	m.logger.Info("Successfully uploaded to S3: s3://%s/%s", bucketName, objectKey)
	return nil
}

// DownloadFromS3 downloads a file from S3
func (m *S3LifecycleManager) DownloadFromS3(ctx context.Context, bucketName, objectKey, filePath, region string) error {
	m.logger.Info("Downloading from S3: s3://%s/%s", bucketName, objectKey)

	result, err := m.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("failed to download file from S3: %w", err)
	}
	defer result.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = io.Copy(file, result.Body)
	if err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	m.logger.Info("Successfully downloaded from S3: s3://%s/%s", bucketName, objectKey)
	return nil
}
