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

// Platform represents the operating system platform
type Platform string

const (
	// PlatformLinux represents Linux/Unix systems
	PlatformLinux Platform = "Linux"
	// PlatformWindows represents Windows systems
	PlatformWindows Platform = "Windows"
	// PlatformUnknown represents unknown platform
	PlatformUnknown Platform = "Unknown"
)

// DetectionConfidence represents the confidence level of platform detection
type DetectionConfidence int

const (
	// ConfidenceNone indicates no confidence in detection
	ConfidenceNone DetectionConfidence = iota
	// ConfidenceLow indicates low confidence (default fallback)
	ConfidenceLow
	// ConfidenceMedium indicates medium confidence (EC2 metadata)
	ConfidenceMedium
	// ConfidenceHigh indicates high confidence (SSM data)
	ConfidenceHigh
)

// DetectionResult contains platform detection information
type DetectionResult struct {
	Platform        Platform
	Confidence      DetectionConfidence
	Source          string
	DetectedAt      time.Time
	PlatformName    string
	PlatformVersion string
}

// Detector handles platform detection with caching
type Detector struct {
	ssmClient SSMClient
	ec2Client EC2Client
	cache     map[string]*DetectionResult
	cacheMu   sync.RWMutex
	cacheTTL  time.Duration
}

// NewDetector creates a new platform detector
func NewDetector(ssmClient SSMClient, ec2Client EC2Client) *Detector {
	return &Detector{
		ssmClient: ssmClient,
		ec2Client: ec2Client,
		cache:     make(map[string]*DetectionResult),
		cacheTTL:  15 * time.Minute,
	}
}

// DetectPlatform detects the platform of an instance with hierarchical fallback
func (d *Detector) DetectPlatform(ctx context.Context, instanceID string) (*DetectionResult, error) {
	// Check cache first
	if result := d.getCachedResult(instanceID); result != nil {
		logging.LogDebug("Using cached platform for instance %s: %s", instanceID, result.Platform)
		return result, nil
	}

	// Try SSM first (highest confidence)
	if result, err := d.detectFromSSM(ctx, instanceID); err == nil && result.Confidence == ConfidenceHigh {
		d.cacheResult(instanceID, result)
		return result, nil
	}

	// Try EC2 (medium confidence)
	if result, err := d.detectFromEC2(ctx, instanceID); err == nil && result.Confidence >= ConfidenceMedium {
		d.cacheResult(instanceID, result)
		return result, nil
	}

	// Default fallback
	result := &DetectionResult{
		Platform:   PlatformLinux,
		Confidence: ConfidenceLow,
		Source:     "default",
		DetectedAt: time.Now(),
	}
	d.cacheResult(instanceID, result)
	logging.LogWarn("Using default platform (Linux) for instance %s", instanceID)
	return result, nil
}

// detectFromSSM uses SSM to detect platform
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

	logging.LogDebug("Detected platform from SSM for %s: %s (%s %s)",
		instanceID, platform, result.PlatformName, result.PlatformVersion)

	return result, nil
}

// detectFromEC2 uses EC2 metadata to detect platform
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
		// If platform is empty, it's likely Linux
		platformStr = "Linux"
	}

	platform := d.normalizePlatform(platformStr)

	result := &DetectionResult{
		Platform:   platform,
		Confidence: ConfidenceMedium,
		Source:     "EC2",
		DetectedAt: time.Now(),
	}

	logging.LogDebug("Detected platform from EC2 for %s: %s", instanceID, platform)

	return result, nil
}

// normalizePlatform converts various platform strings to standardized Platform type
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
		return PlatformLinux // Default to Linux for empty platform
	}

	return PlatformUnknown
}

// getCachedResult retrieves a cached result if still valid
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

// cacheResult stores a detection result in cache
func (d *Detector) cacheResult(instanceID string, result *DetectionResult) {
	d.cacheMu.Lock()
	defer d.cacheMu.Unlock()
	d.cache[instanceID] = result
}

// ClearCache clears the platform detection cache
func (d *Detector) ClearCache() {
	d.cacheMu.Lock()
	defer d.cacheMu.Unlock()
	d.cache = make(map[string]*DetectionResult)
}

// SetCacheTTL sets the cache time-to-live duration
func (d *Detector) SetCacheTTL(ttl time.Duration) {
	d.cacheTTL = ttl
}
