package ssm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"ztictl/pkg/logging"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// IAMManager handles IAM operations for S3 file transfers
type IAMManager struct {
	logger    *logging.Logger
	iamClient *iam.Client
	ec2Client *ec2.Client
}

// PolicyCleanupFunc represents a function that cleans up IAM resources
type PolicyCleanupFunc func() error

// S3PolicyDocument represents the IAM policy for S3 access
type S3PolicyDocument struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

type Statement struct {
	Effect   string   `json:"Effect"`
	Action   []string `json:"Action"`
	Resource string   `json:"Resource"`
}

// Configuration constants
const (
	IAMPropagationDelay = 5 * time.Second
	PolicyNamePrefix    = "ZTIaws-SSM-S3-Access"
)

// NewIAMManager creates a new IAM manager
func NewIAMManager(logger *logging.Logger, iamClient *iam.Client, ec2Client *ec2.Client) (*IAMManager, error) {
	return &IAMManager{
		logger:    logger,
		iamClient: iamClient,
		ec2Client: ec2Client,
	}, nil
}

// generateUniqueID creates a unique identifier for policy names
func (m *IAMManager) generateUniqueID() string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Generate 8 random bytes (16 hex characters)
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to pseudo-random bytes based on timestamp and nanoseconds
		m.logger.Warn("Failed to generate random bytes, using timestamp-based fallback", "error", err)
		nano := time.Now().UnixNano()
		// Generate pseudo-random bytes from timestamp and nanoseconds
		for i := 0; i < 8; i++ {
			randomBytes[i] = byte((nano >> (i * 8)) ^ (nano >> (i * 4)))
		}
	}

	randomHex := hex.EncodeToString(randomBytes)
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	return fmt.Sprintf("%s-%s-%s", timestamp, hostname, randomHex)
}

// getInstanceProfileRole gets the IAM role name for an EC2 instance
func (m *IAMManager) getInstanceProfileRole(ctx context.Context, instanceID string) (string, error) {
	m.logger.Debug("Getting instance profile role for instance", "instanceID", instanceID)

	// Get instance profile name
	describeResult, err := m.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(describeResult.Reservations) == 0 || len(describeResult.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("instance not found: %s", instanceID)
	}

	instance := describeResult.Reservations[0].Instances[0]
	if instance.IamInstanceProfile == nil || instance.IamInstanceProfile.Arn == nil {
		return "", fmt.Errorf("no IAM instance profile found for instance %s", instanceID)
	}

	// Extract instance profile name from ARN
	arnParts := strings.Split(*instance.IamInstanceProfile.Arn, "/")
	if len(arnParts) < 2 {
		return "", fmt.Errorf("invalid instance profile ARN format")
	}
	instanceProfileName := arnParts[len(arnParts)-1]

	// Get role name from instance profile
	getProfileResult, err := m.iamClient.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get instance profile: %w", err)
	}

	if len(getProfileResult.InstanceProfile.Roles) == 0 {
		return "", fmt.Errorf("no role found in instance profile %s", instanceProfileName)
	}

	roleName := *getProfileResult.InstanceProfile.Roles[0].RoleName
	return roleName, nil
}

// createS3PolicyDocument creates the IAM policy document for S3 access
func (m *IAMManager) createS3PolicyDocument(bucketName string) (string, error) {
	policy := S3PolicyDocument{
		Version: "2012-10-17",
		Statement: []Statement{
			{
				Effect: "Allow",
				Action: []string{
					"s3:GetObject",
					"s3:PutObject",
					"s3:DeleteObject",
				},
				Resource: fmt.Sprintf("arn:aws:s3:::%s/*", bucketName),
			},
			{
				Effect: "Allow",
				Action: []string{
					"s3:ListBucket",
				},
				Resource: fmt.Sprintf("arn:aws:s3:::%s", bucketName),
			},
		},
	}

	policyDoc, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal policy document: %w", err)
	}

	return string(policyDoc), nil
}

// AttachS3Permissions attaches S3 permissions to an instance's IAM role and returns a cleanup function
func (m *IAMManager) AttachS3Permissions(ctx context.Context, instanceID, region, bucketName string) (PolicyCleanupFunc, error) {
	m.logger.Debug("Attaching S3 permissions", "bucketName", bucketName, "instanceID", instanceID)

	// Get the role name
	roleName, err := m.getInstanceProfileRole(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	m.logger.Debug("Found role", "roleName", roleName)

	// Create unique policy name with timestamp
	uniqueID := m.generateUniqueID()
	policyName := fmt.Sprintf("%s-%s", PolicyNamePrefix, uniqueID)

	// Create policy document
	policyDoc, err := m.createS3PolicyDocument(bucketName)
	if err != nil {
		return nil, err
	}

	// Create and attach the policy
	createPolicyResult, err := m.iamClient.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDoc),
		Description:    aws.String("Temporary S3 access for ztiaws SSM file transfer"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 policy: %w", err)
	}

	policyARN := *createPolicyResult.Policy.Arn
	m.logger.Debug("Created policy", "policyARN", policyARN)

	// Attach policy to role
	if _, err := m.iamClient.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyARN),
	}); err != nil {
		// Clean up policy if attachment fails
		_, _ = m.iamClient.DeletePolicy(ctx, &iam.DeletePolicyInput{PolicyArn: aws.String(policyARN)}) // #nosec G104 - cleanup operation
		return nil, fmt.Errorf("failed to attach policy to role: %w", err)
	}

	m.logger.Debug("Attached policy to role", "roleName", roleName)

	// Wait for IAM propagation
	m.logger.Debug("Waiting for IAM changes to propagate", "delay", IAMPropagationDelay)
	time.Sleep(IAMPropagationDelay)

	// Return cleanup function
	cleanupFunc := func() error {
		m.logger.Debug("Cleaning up IAM policy", "policyARN", policyARN, "roleName", roleName)

		// Detach policy from role
		if _, err := m.iamClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: aws.String(policyARN),
		}); err != nil {
			m.logger.Warn("Failed to detach policy from role (may already be detached)", "error", err)
		} else {
			m.logger.Debug("Detached policy from role", "roleName", roleName)
		}

		// Delete the policy
		if _, err := m.iamClient.DeletePolicy(ctx, &iam.DeletePolicyInput{
			PolicyArn: aws.String(policyARN),
		}); err != nil {
			m.logger.Warn("Failed to delete policy", "policyARN", policyARN, "error", err)
			return err
		} else {
			m.logger.Debug("Deleted policy", "policyARN", policyARN)
		}

		return nil
	}

	return cleanupFunc, nil
}

// RemoveS3Permissions is deprecated - use the cleanup function returned by AttachS3Permissions instead
func (m *IAMManager) RemoveS3Permissions(ctx context.Context, instanceID, region string) error {
	m.logger.Warn("RemoveS3Permissions called - this method is deprecated, use cleanup function instead", "instanceID", instanceID)
	return nil
}

// ValidateInstanceIAMSetup validates that an instance has the required IAM setup
func (m *IAMManager) ValidateInstanceIAMSetup(ctx context.Context, instanceID, region string) error {
	m.logger.Debug("Validating IAM setup for instance", "instanceID", instanceID)

	// Check if instance has IAM instance profile
	describeResult, err := m.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(describeResult.Reservations) == 0 || len(describeResult.Reservations[0].Instances) == 0 {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	instance := describeResult.Reservations[0].Instances[0]
	if instance.IamInstanceProfile == nil || instance.IamInstanceProfile.Arn == nil {
		return fmt.Errorf("instance %s does not have an IAM instance profile attached", instanceID)
	}

	m.logger.Debug("Instance has IAM instance profile", "arn", *instance.IamInstanceProfile.Arn)

	// Get and validate role exists
	roleName, err := m.getInstanceProfileRole(ctx, instanceID)
	if err != nil {
		return err
	}

	m.logger.Debug("IAM validation successful", "instanceID", instanceID, "roleName", roleName)
	return nil
}

// EmergencyCleanup performs emergency cleanup - simplified to just log a warning
func (m *IAMManager) EmergencyCleanup(ctx context.Context, region string) error {
	m.logger.Warn("EmergencyCleanup called - with new design, cleanup should be handled by cleanup functions returned from AttachS3Permissions")
	m.logger.Info("Emergency cleanup completed - no action taken with simplified design")
	return nil
}
