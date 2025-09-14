package platform

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"ztictl/pkg/logging"
)

type Platform string

const (
	PlatformLinux   Platform = "Linux"
	PlatformWindows Platform = "Windows"
	PlatformUnknown Platform = "Unknown"
)

type DetectionConfidence int

const (
	ConfidenceNone DetectionConfidence = iota
	ConfidenceLow
	ConfidenceMedium
	ConfidenceHigh
)

type DetectionResult struct {
	Platform        Platform
	Confidence      DetectionConfidence
	Source          string
	DetectedAt      time.Time
	PlatformName    string
	PlatformVersion string
}

type Detector struct {
	ssmClient SSMClient
	ec2Client EC2Client
	cache     map[string]*DetectionResult
	cacheMu   sync.RWMutex
	cacheTTL  time.Duration
	logger    *logging.Logger
}

// NewDetector creates a new platform detector
// BREAKING CHANGE: v2.1.0 - Logger is now a required parameter for better observability.
// This change ensures consistent logging across all components. All callers have been
// updated to provide a logger instance (use logging.NewNoOpLogger() for tests).
func NewDetector(ssmClient SSMClient, ec2Client EC2Client, logger *logging.Logger) *Detector {
	return &Detector{
		ssmClient: ssmClient,
		ec2Client: ec2Client,
		cache:     make(map[string]*DetectionResult),
		cacheTTL:  15 * time.Minute,
		logger:    logger,
	}
}

func (d *Detector) DetectPlatform(ctx context.Context, instanceID string) (*DetectionResult, error) {
	if result := d.getCachedResult(instanceID); result != nil {
		d.logger.Debug("Using cached platform for instance", "instanceID", instanceID, "platform", result.Platform)
		return result, nil
	}

	if result, err := d.detectFromSSM(ctx, instanceID); err == nil && result.Confidence == ConfidenceHigh {
		d.cacheResult(instanceID, result)
		return result, nil
	} else if err != nil {
		d.logger.Debug("SSM detection failed", "instanceID", instanceID, "error", err)
	}

	if result, err := d.detectFromEC2(ctx, instanceID); err == nil && result.Confidence >= ConfidenceMedium {
		d.cacheResult(instanceID, result)
		return result, nil
	} else if err != nil {
		d.logger.Debug("EC2 detection failed", "instanceID", instanceID, "error", err)
	}

	result := &DetectionResult{
		Platform:   PlatformLinux,
		Confidence: ConfidenceLow,
		Source:     "default",
		DetectedAt: time.Now(),
	}
	d.cacheResult(instanceID, result)
	d.logger.Warn("Using default platform for instance", "instanceID", instanceID, "platform", "Linux")
	return result, nil
}

func (d *Detector) detectFromSSM(ctx context.Context, instanceID string) (*DetectionResult, error) {
	if d.ssmClient == nil {
		return nil, fmt.Errorf("SSM client not available")
	}

	input := &ssm.DescribeInstanceInformationInput{
		Filters: []ssmtypes.InstanceInformationStringFilter{
			{
				Key:    aws.String("InstanceIds"),
				Values: []string{instanceID},
			},
		},
	}

	resp, err := d.ssmClient.DescribeInstanceInformation(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance information: %w", err)
	}

	if len(resp.InstanceInformationList) == 0 {
		return nil, fmt.Errorf("no SSM information found for instance %s", instanceID)
	}

	info := resp.InstanceInformationList[0]
	platform := d.normalizePlatform(string(info.PlatformType))

	result := &DetectionResult{
		Platform:        platform,
		Confidence:      ConfidenceHigh,
		Source:          "SSM",
		DetectedAt:      time.Now(),
		PlatformName:    aws.ToString(info.PlatformName),
		PlatformVersion: aws.ToString(info.PlatformVersion),
	}

	d.logger.Debug("Detected platform from SSM",
		"instanceID", instanceID,
		"platform", platform,
		"platformName", result.PlatformName,
		"platformVersion", result.PlatformVersion)

	return result, nil
}

func (d *Detector) detectFromEC2(ctx context.Context, instanceID string) (*DetectionResult, error) {
	if d.ec2Client == nil {
		return nil, fmt.Errorf("EC2 client not available")
	}

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	resp, err := d.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("no EC2 information found for instance %s", instanceID)
	}

	instance := resp.Reservations[0].Instances[0]
	platformStr := string(instance.Platform)

	if platformStr == "" {
		platformStr = "Linux"
	}

	platform := d.normalizePlatform(platformStr)

	result := &DetectionResult{
		Platform:   platform,
		Confidence: ConfidenceMedium,
		Source:     "EC2",
		DetectedAt: time.Now(),
	}

	d.logger.Debug("Detected platform from EC2", "instanceID", instanceID, "platform", platform)

	return result, nil
}

func (d *Detector) normalizePlatform(platform string) Platform {
	normalized := strings.ToLower(platform)

	if strings.Contains(normalized, "windows") {
		return PlatformWindows
	}

	if strings.Contains(normalized, "linux") ||
		strings.Contains(normalized, "unix") ||
		strings.Contains(normalized, "ubuntu") ||
		strings.Contains(normalized, "amazon") ||
		strings.Contains(normalized, "centos") ||
		strings.Contains(normalized, "rhel") ||
		strings.Contains(normalized, "debian") {
		return PlatformLinux
	}

	if platform == "" {
		return PlatformLinux
	}

	return PlatformUnknown
}

func (d *Detector) getCachedResult(instanceID string) *DetectionResult {
	d.cacheMu.RLock()
	defer d.cacheMu.RUnlock()

	result, exists := d.cache[instanceID]
	if !exists {
		return nil
	}

	if time.Since(result.DetectedAt) > d.cacheTTL {
		return nil
	}

	return result
}

func (d *Detector) cacheResult(instanceID string, result *DetectionResult) {
	d.cacheMu.Lock()
	defer d.cacheMu.Unlock()
	d.cache[instanceID] = result
}

func (d *Detector) ClearCache() {
	d.cacheMu.Lock()
	defer d.cacheMu.Unlock()
	d.cache = make(map[string]*DetectionResult)
}

func (d *Detector) SetCacheTTL(ttl time.Duration) {
	d.cacheTTL = ttl
}
