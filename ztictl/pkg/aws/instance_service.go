package aws

import (
	"context"
	"fmt"
	"strings"

	"ztictl/internal/interactive"
	"ztictl/pkg/logging"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// InstanceService provides common instance operations across commands
type InstanceService struct {
	clientPool ClientPoolInterface
	logger     *logging.Logger
}

// ClientPoolInterface defines the interface for AWS client pools
type ClientPoolInterface interface {
	GetSSMClient(ctx context.Context, region string) (*ssm.Client, error)
	GetEC2Client(ctx context.Context, region string) (*ec2.Client, error)
}

// ListFilters represents filters for listing instances
type ListFilters struct {
	Tag    string `json:"tag,omitempty"`    // Format: key=value (deprecated, use Tags)
	Tags   string `json:"tags,omitempty"`   // Format: key1=value1,key2=value2
	Status string `json:"status,omitempty"` // Instance state
	Name   string `json:"name,omitempty"`   // Name pattern
}

// NewInstanceService creates a new instance service
func NewInstanceService(clientPool ClientPoolInterface, logger *logging.Logger) *InstanceService {
	return &InstanceService{
		clientPool: clientPool,
		logger:     logger,
	}
}

// ListInstances retrieves instances with SSM status - shared between auth and ssm commands
func (s *InstanceService) ListInstances(ctx context.Context, region string, filters *ListFilters) ([]interactive.Instance, error) {
	s.logger.Debug("Listing all EC2 instances with SSM status in region", "region", region)

	// Get clients from pool
	ec2Client, err := s.clientPool.GetEC2Client(ctx, region)
	if err != nil {
		return nil, fmt.Errorf("failed to get EC2 client for region %s: %w", region, err)
	}

	// Get all EC2 instances first
	allInstances, err := s.getAllEC2Instances(ctx, ec2Client, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get EC2 instances: %w", err)
	}

	if len(allInstances) == 0 {
		return []interactive.Instance{}, nil
	}

	// Get SSM status information
	ssmClient, err := s.clientPool.GetSSMClient(ctx, region)
	if err != nil {
		s.logger.Warn("Failed to get SSM client, continuing without SSM status", "error", err)
	}

	var ssmStatusMap map[string]ssmtypes.InstanceInformation
	if ssmClient != nil {
		ssmStatusMap, err = s.getSSMStatusMap(ctx, ssmClient)
		if err != nil {
			s.logger.Warn("Failed to get SSM status information, marking all as 'No Agent'", "error", err)
			// Continue without SSM status - we'll mark all as "No Agent"
		}
	}

	// Combine EC2 data with SSM status
	instances := make([]interactive.Instance, 0, len(allInstances))
	for _, ec2Instance := range allInstances {
		instanceID := *ec2Instance.InstanceId

		// Extract instance name from tags
		var instanceName string
		for _, tag := range ec2Instance.Tags {
			if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
				instanceName = *tag.Value
				break
			}
		}

		// Convert EC2 tags to map
		tagMap := make(map[string]string)
		for _, tag := range ec2Instance.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagMap[*tag.Key] = *tag.Value
			}
		}

		// Get SSM status if available
		var ssmStatus, ssmAgentVersion, lastPingDateTime string
		if ssmStatusMap != nil {
			if ssmInfo, exists := ssmStatusMap[instanceID]; exists {
				if ssmInfo.PingStatus != "" {
					ssmStatus = string(ssmInfo.PingStatus)
				}
				if ssmInfo.AgentVersion != nil {
					ssmAgentVersion = *ssmInfo.AgentVersion
				}
				if ssmInfo.LastPingDateTime != nil {
					lastPingDateTime = ssmInfo.LastPingDateTime.String()
				}
			} else {
				ssmStatus = "No Agent"
			}
		} else {
			ssmStatus = "No Agent"
		}

		// Extract IP addresses
		var privateIP, publicIP string
		if ec2Instance.PrivateIpAddress != nil {
			privateIP = *ec2Instance.PrivateIpAddress
		}
		if ec2Instance.PublicIpAddress != nil {
			publicIP = *ec2Instance.PublicIpAddress
		}

		instance := interactive.Instance{
			InstanceID:       instanceID,
			Name:             instanceName,
			State:            string(ec2Instance.State.Name),
			Platform:         getPlatformFromInstance(ec2Instance),
			PrivateIPAddress: privateIP,
			PublicIPAddress:  publicIP,
			SSMStatus:        ssmStatus,
			SSMAgentVersion:  ssmAgentVersion,
			LastPingDateTime: lastPingDateTime,
			Tags:             tagMap,
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// ResolveInstanceIdentifier resolves an instance name or ID to an instance ID
func (s *InstanceService) ResolveInstanceIdentifier(ctx context.Context, identifier, region string) (string, error) {
	// If it's already an instance ID, validate and return it
	if isInstanceID(identifier) {
		err := s.validateInstanceID(ctx, identifier, region)
		if err != nil {
			return "", fmt.Errorf("instance ID validation failed: %w", err)
		}
		return identifier, nil
	}

	// Search by name tag
	return s.findInstanceByName(ctx, identifier, region)
}

// SelectInstanceWithFallback handles the common pattern of "if no ID provided, show fuzzy finder"
func (s *InstanceService) SelectInstanceWithFallback(ctx context.Context, identifier, region string, filters *ListFilters) (string, error) {
	if identifier != "" {
		// Resolve provided identifier
		return s.ResolveInstanceIdentifier(ctx, identifier, region)
	}

	// No identifier provided, show fuzzy finder
	s.logger.Info("No instance specified, fetching instances from region", "region", region)
	instances, err := s.ListInstances(ctx, region, filters)
	if err != nil {
		return "", fmt.Errorf("failed to list instances: %w", err)
	}

	if len(instances) == 0 {
		return "", fmt.Errorf("no instances found in region: %s", region)
	}

	s.logger.Info("Found instances, launching interactive selector", "count", len(instances), "region", region)
	selected, err := interactive.SelectInstance(instances, fmt.Sprintf("Select instance (%d available)", len(instances)))
	if err != nil {
		return "", err
	}

	return selected.InstanceID, nil
}

// Helper methods

// getAllEC2Instances retrieves all EC2 instances in a region with optional filtering
func (s *InstanceService) getAllEC2Instances(ctx context.Context, ec2Client *ec2.Client, filters *ListFilters) ([]types.Instance, error) {
	input := &ec2.DescribeInstancesInput{}

	// Apply filters
	var ec2Filters []types.Filter
	if filters != nil {
		// Handle tag filters (both old single tag and new multiple tags)
		tagFilters := make(map[string]string)

		// Parse legacy single tag filter
		if filters.Tag != "" {
			parsed, err := parseTagFilter(filters.Tag)
			if err != nil {
				return nil, fmt.Errorf("invalid tag filter format: %w", err)
			}
			for k, v := range parsed {
				tagFilters[k] = v
			}
		}

		// Parse new multiple tags filter
		if filters.Tags != "" {
			parsed, err := parseTagFilters(filters.Tags)
			if err != nil {
				return nil, fmt.Errorf("invalid tags filter format: %w", err)
			}
			for k, v := range parsed {
				tagFilters[k] = v
			}
		}

		// Apply tag filters
		for key, value := range tagFilters {
			ec2Filters = append(ec2Filters, types.Filter{
				Name:   aws.String("tag:" + key),
				Values: []string{value},
			})
		}

		// Apply status filter
		if filters.Status != "" {
			ec2Filters = append(ec2Filters, types.Filter{
				Name:   aws.String("instance-state-name"),
				Values: []string{filters.Status},
			})
		}

		// Apply name filter
		if filters.Name != "" {
			ec2Filters = append(ec2Filters, types.Filter{
				Name:   aws.String("tag:Name"),
				Values: []string{"*" + filters.Name + "*"},
			})
		}
	}

	if len(ec2Filters) > 0 {
		input.Filters = ec2Filters
	}

	var allInstances []types.Instance
	paginator := ec2.NewDescribeInstancesPaginator(ec2Client, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %w", err)
		}

		for _, reservation := range output.Reservations {
			allInstances = append(allInstances, reservation.Instances...)
		}
	}

	return allInstances, nil
}

// getSSMStatusMap retrieves SSM status information for all instances and returns as a map
func (s *InstanceService) getSSMStatusMap(ctx context.Context, ssmClient *ssm.Client) (map[string]ssmtypes.InstanceInformation, error) {
	statusMap := make(map[string]ssmtypes.InstanceInformation)

	input := &ssm.DescribeInstanceInformationInput{}
	paginator := ssm.NewDescribeInstanceInformationPaginator(ssmClient, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe SSM instance information: %w", err)
		}

		for _, instance := range output.InstanceInformationList {
			if instance.InstanceId != nil {
				statusMap[*instance.InstanceId] = instance
			}
		}
	}

	return statusMap, nil
}

// validateInstanceID validates that an instance ID exists
func (s *InstanceService) validateInstanceID(ctx context.Context, instanceID, region string) error {
	ec2Client, err := s.clientPool.GetEC2Client(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to get EC2 client for region %s: %w", region, err)
	}

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	_, err = ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("instance %s not found: %w", instanceID, err)
	}

	return nil
}

// findInstanceByName finds an instance by its Name tag
func (s *InstanceService) findInstanceByName(ctx context.Context, name, region string) (string, error) {
	ec2Client, err := s.clientPool.GetEC2Client(ctx, region)
	if err != nil {
		return "", fmt.Errorf("failed to get EC2 client for region %s: %w", region, err)
	}

	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{name},
			},
		},
	}

	result, err := ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to search for instance by name '%s': %w", name, err)
	}

	var foundInstances []types.Instance
	for _, reservation := range result.Reservations {
		foundInstances = append(foundInstances, reservation.Instances...)
	}

	if len(foundInstances) == 0 {
		return "", fmt.Errorf("no instance found with name '%s'", name)
	}

	if len(foundInstances) > 1 {
		return "", fmt.Errorf("multiple instances found with name '%s', use instance ID instead", name)
	}

	return *foundInstances[0].InstanceId, nil
}

// Helper functions

// isInstanceID checks if a string matches the AWS instance ID pattern
func isInstanceID(identifier string) bool {
	// AWS instance IDs follow the pattern: i-[0-9a-f]{8,17}
	if len(identifier) < 10 || len(identifier) > 19 {
		return false
	}

	if !strings.HasPrefix(identifier, "i-") {
		return false
	}

	suffix := identifier[2:]
	for _, char := range suffix {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}

	return true
}

// parseTagFilter parses a single tag filter in the format key=value
func parseTagFilter(tagStr string) (map[string]string, error) {
	result := make(map[string]string)

	parts := strings.SplitN(tagStr, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("tag filter must be in format key=value, got: %s", tagStr)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if key == "" {
		return nil, fmt.Errorf("tag key cannot be empty")
	}

	result[key] = value
	return result, nil
}

// parseTagFilters parses comma-separated tag filters into individual key=value pairs
func parseTagFilters(tagsStr string) (map[string]string, error) {
	result := make(map[string]string)

	if tagsStr == "" {
		return result, nil
	}

	// Split by comma and process each tag
	tagPairs := strings.Split(tagsStr, ",")
	for _, tagPair := range tagPairs {
		tagPair = strings.TrimSpace(tagPair)
		if tagPair == "" {
			continue
		}

		parsed, err := parseTagFilter(tagPair)
		if err != nil {
			return nil, err
		}

		// Merge the single tag into the result
		for k, v := range parsed {
			result[k] = v
		}
	}

	return result, nil
}

// getPlatformFromInstance determines the platform from EC2 instance information
func getPlatformFromInstance(instance types.Instance) string {
	// Check platform details first (most reliable)
	if instance.PlatformDetails != nil {
		platform := *instance.PlatformDetails
		if strings.Contains(strings.ToLower(platform), "windows") {
			return "Windows"
		}
		if strings.Contains(strings.ToLower(platform), "linux") {
			return "Linux"
		}
	}

	// Check platform field (legacy)
	if instance.Platform != "" {
		return string(instance.Platform)
	}

	// Default assumption for instances without explicit platform info
	return "Linux"
}
