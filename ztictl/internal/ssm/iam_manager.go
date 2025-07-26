package ssm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"ztictl/internal/logging"
)

// IAMManager handles IAM operations for S3 file transfers
type IAMManager struct {
	logger       *logging.Logger
	iamClient    *iam.Client
	ec2Client    *ec2.Client
	tempDir      string
	registryFile string
	lockDir      string
	mu           sync.RWMutex
}

// PolicyRegistryEntry represents a policy registry entry
type PolicyRegistryEntry struct {
	InstanceID string    `json:"instance_id"`
	Region     string    `json:"region"`
	PolicyARN  string    `json:"policy_arn"`
	PolicyFile string    `json:"policy_file"`
	Timestamp  time.Time `json:"timestamp"`
}

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
	IAMPropagationDelay         = 5 * time.Second
	IAMLockAcquisitionTimeout   = 30 * time.Second
	StaleLockTimeoutSeconds     = 120 * time.Second
	RegistryCleanupAgeThreshold = 24 * time.Hour
	PolicyNamePrefix            = "ZTIaws-SSM-S3-Access"
)

// NewIAMManager creates a new IAM manager
func NewIAMManager(logger *logging.Logger, iamClient *iam.Client, ec2Client *ec2.Client) (*IAMManager, error) {
	tempDir := filepath.Join(os.TempDir(), "ztiaws")
	if err := os.MkdirAll(tempDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	registryFile := filepath.Join(tempDir, "policy-registry.json")
	lockDir := filepath.Join(tempDir, "locks")

	if err := os.MkdirAll(lockDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create lock directory: %w", err)
	}

	return &IAMManager{
		logger:       logger,
		iamClient:    iamClient,
		ec2Client:    ec2Client,
		tempDir:      tempDir,
		registryFile: registryFile,
		lockDir:      lockDir,
	}, nil
}

// generateUniqueID creates a unique identifier for policy names
func (m *IAMManager) generateUniqueID() string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Generate 8 random bytes (16 hex characters)
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp + process ID if random fails
		return fmt.Sprintf("%s-%d", timestamp, os.Getpid())
	}

	randomHex := hex.EncodeToString(randomBytes)
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	return fmt.Sprintf("%s-%s-%s", timestamp, hostname, randomHex)
}

// acquireInstanceLock acquires a filesystem lock for IAM operations on a specific instance
func (m *IAMManager) acquireInstanceLock(instanceID string) (string, error) {
	lockFile := filepath.Join(m.lockDir, fmt.Sprintf("iam-%s.lock", instanceID))
	elapsed := time.Duration(0)
	waitInterval := 1 * time.Second

	m.logger.Debug("Attempting to acquire IAM lock for instance: %s", instanceID)

	for elapsed < IAMLockAcquisitionTimeout {
		// Use mkdir as an atomic operation for locking
		if err := os.Mkdir(lockFile, 0700); err == nil {
			m.logger.Debug("Acquired IAM lock for instance: %s", instanceID)
			return lockFile, nil
		}

		// Check if lock is stale
		if info, err := os.Stat(lockFile); err == nil {
			lockAge := time.Since(info.ModTime())
			if lockAge > StaleLockTimeoutSeconds {
				m.logger.Debug("Removing stale lock for instance: %s (age: %v)", instanceID, lockAge)
				os.RemoveAll(lockFile)
			}
		}

		time.Sleep(waitInterval)
		elapsed += waitInterval
	}

	return "", fmt.Errorf("failed to acquire IAM lock for instance %s within %v", instanceID, IAMLockAcquisitionTimeout)
}

// releaseInstanceLock releases a filesystem lock
func (m *IAMManager) releaseInstanceLock(lockFile string) {
	if lockFile != "" {
		if err := os.RemoveAll(lockFile); err != nil {
			m.logger.Warn("Failed to release IAM lock: %s", lockFile)
		} else {
			m.logger.Debug("Released IAM lock: %s", filepath.Base(lockFile))
		}
	}
}

// loadPolicyRegistry loads the policy registry from disk
func (m *IAMManager) loadPolicyRegistry() ([]PolicyRegistryEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, err := os.Stat(m.registryFile); os.IsNotExist(err) {
		return []PolicyRegistryEntry{}, nil
	}

	data, err := os.ReadFile(m.registryFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry file: %w", err)
	}

	var entries []PolicyRegistryEntry
	if len(data) > 0 {
		if err := json.Unmarshal(data, &entries); err != nil {
			return nil, fmt.Errorf("failed to unmarshal registry: %w", err)
		}
	}

	return entries, nil
}

// savePolicyRegistry saves the policy registry to disk
func (m *IAMManager) savePolicyRegistry(entries []PolicyRegistryEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := os.WriteFile(m.registryFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write registry file: %w", err)
	}

	return nil
}

// addPolicyToRegistry adds a policy to the registry
func (m *IAMManager) addPolicyToRegistry(instanceID, region, policyARN, policyFile string) error {
	entries, err := m.loadPolicyRegistry()
	if err != nil {
		return err
	}

	entry := PolicyRegistryEntry{
		InstanceID: instanceID,
		Region:     region,
		PolicyARN:  policyARN,
		PolicyFile: policyFile,
		Timestamp:  time.Now(),
	}

	entries = append(entries, entry)

	if err := m.savePolicyRegistry(entries); err != nil {
		return err
	}

	m.logger.Debug("Added policy to registry: %s for instance %s", policyARN, instanceID)
	return nil
}

// removePoliciesFromRegistry removes policies for a specific instance from registry
func (m *IAMManager) removePoliciesFromRegistry(instanceID string) error {
	entries, err := m.loadPolicyRegistry()
	if err != nil {
		return err
	}

	var filteredEntries []PolicyRegistryEntry
	for _, entry := range entries {
		if entry.InstanceID != instanceID {
			filteredEntries = append(filteredEntries, entry)
		} else {
			// Clean up associated policy file if it exists
			if entry.PolicyFile != "" {
				os.Remove(entry.PolicyFile)
			}
		}
	}

	if err := m.savePolicyRegistry(filteredEntries); err != nil {
		return err
	}

	m.logger.Debug("Removed policies from registry for instance: %s", instanceID)
	return nil
}

// getPoliciesForInstance gets policies for a specific instance from registry
func (m *IAMManager) getPoliciesForInstance(instanceID string) ([]PolicyRegistryEntry, error) {
	entries, err := m.loadPolicyRegistry()
	if err != nil {
		return nil, err
	}

	var instancePolicies []PolicyRegistryEntry
	for _, entry := range entries {
		if entry.InstanceID == instanceID {
			instancePolicies = append(instancePolicies, entry)
		}
	}

	return instancePolicies, nil
}

// cleanupStaleRegistryEntries removes old entries from registry
func (m *IAMManager) cleanupStaleRegistryEntries() error {
	entries, err := m.loadPolicyRegistry()
	if err != nil {
		return err
	}

	currentTime := time.Now()
	var validEntries []PolicyRegistryEntry
	cleanedCount := 0

	for _, entry := range entries {
		age := currentTime.Sub(entry.Timestamp)
		if age > RegistryCleanupAgeThreshold {
			m.logger.Debug("Removing stale registry entry for instance %s (age: %v)", entry.InstanceID, age)
			// Clean up associated policy file if it exists
			if entry.PolicyFile != "" {
				os.Remove(entry.PolicyFile)
			}
			cleanedCount++
		} else {
			validEntries = append(validEntries, entry)
		}
	}

	if cleanedCount > 0 {
		if err := m.savePolicyRegistry(validEntries); err != nil {
			return err
		}
		m.logger.Debug("Cleaned up %d stale registry entries", cleanedCount)
	}

	return nil
}

// getInstanceProfileRole gets the IAM role name for an EC2 instance
func (m *IAMManager) getInstanceProfileRole(ctx context.Context, instanceID, region string) (string, error) {
	m.logger.Debug("Getting instance profile role for instance: %s", instanceID)

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

// createSecureTempFile creates a temporary file with restricted permissions
func (m *IAMManager) createSecureTempFile(prefix string) (string, error) {
	tempFile := filepath.Join(m.tempDir, fmt.Sprintf("%s-%s", prefix, m.generateUniqueID()))

	file, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create secure temp file: %w", err)
	}
	file.Close()

	return tempFile, nil
}

// AttachS3Permissions attaches S3 permissions to an instance's IAM role
func (m *IAMManager) AttachS3Permissions(ctx context.Context, instanceID, region, bucketName string) error {
	m.logger.Debug("Attaching S3 permissions for bucket: %s", bucketName)

	// Acquire lock for this instance
	lockFile, err := m.acquireInstanceLock(instanceID)
	if err != nil {
		return fmt.Errorf("failed to acquire lock for IAM operations on instance %s: %w", instanceID, err)
	}
	defer m.releaseInstanceLock(lockFile)

	// Get the role name
	roleName, err := m.getInstanceProfileRole(ctx, instanceID, region)
	if err != nil {
		return err
	}

	m.logger.Debug("Found role: %s", roleName)

	// Create unique policy name
	uniqueID := m.generateUniqueID()
	policyName := fmt.Sprintf("%s-%s", PolicyNamePrefix, uniqueID)

	// Create policy document
	policyDoc, err := m.createS3PolicyDocument(bucketName)
	if err != nil {
		return err
	}

	// Create policy tracking file
	policyTrackingFile, err := m.createSecureTempFile(fmt.Sprintf("ztiaws-s3-policy-%s-%s", instanceID, uniqueID))
	if err != nil {
		return err
	}

	// Create and attach the policy
	createPolicyResult, err := m.iamClient.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(policyDoc),
		Description:    aws.String("Temporary S3 access for ztiaws SSM file transfer"),
	})
	if err != nil {
		os.Remove(policyTrackingFile)
		return fmt.Errorf("failed to create S3 policy: %w", err)
	}

	policyARN := *createPolicyResult.Policy.Arn
	m.logger.Debug("Created policy: %s", policyARN)

	// Write policy ARN to tracking file
	if err := os.WriteFile(policyTrackingFile, []byte(policyARN), 0600); err != nil {
		// Clean up policy if we can't track it
		m.iamClient.DeletePolicy(ctx, &iam.DeletePolicyInput{PolicyArn: aws.String(policyARN)})
		os.Remove(policyTrackingFile)
		return fmt.Errorf("failed to write policy tracking file: %w", err)
	}

	// Attach policy to role
	if _, err := m.iamClient.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyARN),
	}); err != nil {
		// Clean up policy if attachment fails
		m.iamClient.DeletePolicy(ctx, &iam.DeletePolicyInput{PolicyArn: aws.String(policyARN)})
		os.Remove(policyTrackingFile)
		return fmt.Errorf("failed to attach policy to role: %w", err)
	}

	m.logger.Debug("Attached policy to role: %s", roleName)

	// Add to registry for efficient cleanup
	if err := m.addPolicyToRegistry(instanceID, region, policyARN, policyTrackingFile); err != nil {
		m.logger.Warn("Failed to add policy to registry: %v", err)
		// Continue anyway as the policy is attached
	}

	// Wait for IAM propagation
	m.logger.Debug("Waiting for IAM changes to propagate (%v)", IAMPropagationDelay)
	time.Sleep(IAMPropagationDelay)

	return nil
}

// RemoveS3Permissions removes S3 permissions from an instance's IAM role
func (m *IAMManager) RemoveS3Permissions(ctx context.Context, instanceID, region string) error {
	m.logger.Debug("Removing S3 permissions for instance: %s", instanceID)

	// Acquire lock for this instance
	lockFile, err := m.acquireInstanceLock(instanceID)
	if err != nil {
		m.logger.Warn("Failed to acquire lock for IAM cleanup on instance %s - skipping cleanup: %v", instanceID, err)
		return nil // Don't fail the entire operation
	}
	defer m.releaseInstanceLock(lockFile)

	// Get policies for this instance from registry
	policyEntries, err := m.getPoliciesForInstance(instanceID)
	if err != nil {
		return fmt.Errorf("failed to get policies from registry: %w", err)
	}

	if len(policyEntries) == 0 {
		m.logger.Debug("No policy entries found in registry for instance: %s", instanceID)
		return nil
	}

	// Get role name
	roleName, err := m.getInstanceProfileRole(ctx, instanceID, region)
	if err != nil {
		m.logger.Debug("Could not get role name for cleanup, but continuing with policy deletion: %v", err)
		roleName = "" // We'll still try to delete policies
	}

	// Process each policy entry
	for _, entry := range policyEntries {
		m.logger.Debug("Processing policy from registry: %s", entry.PolicyARN)

		// Clean up the policy file
		if entry.PolicyFile != "" {
			os.Remove(entry.PolicyFile)
		}

		if entry.PolicyARN == "" {
			m.logger.Debug("No policy ARN found in registry entry")
			continue
		}

		// Detach policy from role if we have role name
		if roleName != "" {
			if _, err := m.iamClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
				RoleName:  aws.String(roleName),
				PolicyArn: aws.String(entry.PolicyARN),
			}); err != nil {
				m.logger.Warn("Failed to detach policy from role (may already be detached): %v", err)
			} else {
				m.logger.Debug("Detached policy from role: %s", roleName)
			}
		}

		// Delete the policy
		if _, err := m.iamClient.DeletePolicy(ctx, &iam.DeletePolicyInput{
			PolicyArn: aws.String(entry.PolicyARN),
		}); err != nil {
			m.logger.Warn("Failed to delete policy %s: %v", entry.PolicyARN, err)
		} else {
			m.logger.Debug("Deleted policy: %s", entry.PolicyARN)
		}
	}

	// Remove entries from registry
	if err := m.removePoliciesFromRegistry(instanceID); err != nil {
		m.logger.Warn("Failed to remove policies from registry: %v", err)
	}

	return nil
}

// ValidateInstanceIAMSetup validates that an instance has the required IAM setup
func (m *IAMManager) ValidateInstanceIAMSetup(ctx context.Context, instanceID, region string) error {
	m.logger.Debug("Validating IAM setup for instance: %s", instanceID)

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
		return fmt.Errorf("instance %s does not have an IAM instance profile attached. Please attach an IAM instance profile with appropriate permissions to the instance", instanceID)
	}

	m.logger.Debug("Instance has IAM instance profile: %s", *instance.IamInstanceProfile.Arn)

	// Get and validate role exists
	roleName, err := m.getInstanceProfileRole(ctx, instanceID, region)
	if err != nil {
		return fmt.Errorf("failed to get IAM role for instance %s: %w", instanceID, err)
	}

	m.logger.Debug("IAM validation successful for instance: %s (role: %s)", instanceID, roleName)
	return nil
}

// EmergencyCleanup performs emergency cleanup of policies
func (m *IAMManager) EmergencyCleanup(ctx context.Context, region string) error {
	m.logger.Debug("Performing emergency IAM cleanup")

	entries, err := m.loadPolicyRegistry()
	if err != nil {
		m.logger.Debug("No registry file found, attempting fallback cleanup")
		return m.fallbackCleanup(ctx, region)
	}

	cleanupCount := 0
	for _, entry := range entries {
		m.logger.Debug("Emergency cleanup of policy: %s", entry.PolicyARN)

		// Clean up policy file
		if entry.PolicyFile != "" {
			os.Remove(entry.PolicyFile)
		}

		// Try to get role name and detach policy
		if entry.InstanceID != "" && region != "" {
			if roleName, err := m.getInstanceProfileRole(ctx, entry.InstanceID, region); err == nil {
				m.iamClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
					RoleName:  aws.String(roleName),
					PolicyArn: aws.String(entry.PolicyARN),
				})
			}
		}

		// Delete the policy
		if entry.PolicyARN != "" {
			if _, err := m.iamClient.DeletePolicy(ctx, &iam.DeletePolicyInput{
				PolicyArn: aws.String(entry.PolicyARN),
			}); err == nil {
				cleanupCount++
			}
		}
	}

	// Clear the registry after cleanup
	if err := m.savePolicyRegistry([]PolicyRegistryEntry{}); err != nil {
		m.logger.Warn("Failed to clear registry after emergency cleanup: %v", err)
	}

	if cleanupCount > 0 {
		m.logger.Debug("Emergency cleanup completed for %d policies from registry", cleanupCount)
	} else {
		m.logger.Debug("No policies found in registry for cleanup")
	}

	// Clean up stale locks
	m.cleanupStaleLocks()

	// Clean up stale registry entries
	m.cleanupStaleRegistryEntries()

	return nil
}

// fallbackCleanup performs cleanup by scanning temp directory
func (m *IAMManager) fallbackCleanup(ctx context.Context, region string) error {
	m.logger.Debug("Performing fallback cleanup by scanning temp directory")

	// Scan our dedicated temp directory for policy files
	files, err := filepath.Glob(filepath.Join(m.tempDir, "ztiaws-s3-policy-*"))
	if err != nil {
		return err
	}

	cleanupCount := 0
	for _, policyFile := range files {
		policyARNBytes, err := os.ReadFile(policyFile)
		if err != nil {
			continue
		}

		policyARN := string(policyARNBytes)
		if policyARN == "" {
			continue
		}

		// Extract instance ID from filename
		filename := filepath.Base(policyFile)
		if strings.HasPrefix(filename, "ztiaws-s3-policy-i-") {
			parts := strings.Split(filename, "-")
			if len(parts) >= 4 {
				instanceID := parts[3] // Should be like "i-1234567890abcdef0"
				if strings.HasPrefix(instanceID, "i-") {
					m.logger.Debug("Extracted instance ID: %s from policy file", instanceID)

					// Try to get role name and detach policy
					if region != "" {
						if roleName, err := m.getInstanceProfileRole(ctx, instanceID, region); err == nil {
							m.iamClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
								RoleName:  aws.String(roleName),
								PolicyArn: aws.String(policyARN),
							})
						}
					}
				}
			}
		}

		// Delete the policy regardless
		m.logger.Debug("Attempting direct policy cleanup: %s", policyARN)
		if _, err := m.iamClient.DeletePolicy(ctx, &iam.DeletePolicyInput{
			PolicyArn: aws.String(policyARN),
		}); err == nil {
			cleanupCount++
		}

		os.Remove(policyFile)
	}

	if cleanupCount > 0 {
		m.logger.Debug("Fallback cleanup completed for %d policy files", cleanupCount)
	} else {
		m.logger.Debug("No policies required cleanup in fallback mode")
	}

	return nil
}

// cleanupStaleLocks removes stale lock files
func (m *IAMManager) cleanupStaleLocks() {
	m.logger.Debug("Cleaning up stale lock files...")

	files, err := filepath.Glob(filepath.Join(m.lockDir, "iam-*.lock"))
	if err != nil {
		return
	}

	lockCleanupCount := 0
	for _, lockFile := range files {
		if info, err := os.Stat(lockFile); err == nil {
			lockAge := time.Since(info.ModTime())
			if lockAge > StaleLockTimeoutSeconds {
				m.logger.Debug("Removing stale lock: %s (age: %v)", filepath.Base(lockFile), lockAge)
				if err := os.RemoveAll(lockFile); err == nil {
					lockCleanupCount++
				}
			}
		}
	}

	if lockCleanupCount > 0 {
		m.logger.Debug("Cleaned up %d stale lock files", lockCleanupCount)
	}

	// Try to remove lock directory if empty
	os.Remove(m.lockDir)
}
