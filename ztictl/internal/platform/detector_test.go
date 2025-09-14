package platform

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"ztictl/pkg/logging"
)

// MockSSMClient for testing SSM operations
type MockSSMClient struct {
	mock.Mock
}

func (m *MockSSMClient) DescribeInstanceInformation(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ssm.DescribeInstanceInformationOutput), args.Error(1)
}

// MockEC2Client for testing EC2 operations
type MockEC2Client struct {
	mock.Mock
}

func (m *MockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ec2.DescribeInstancesOutput), args.Error(1)
}

func TestPlatformNormalization(t *testing.T) {
	detector := &Detector{}

	tests := []struct {
		name     string
		input    string
		expected Platform
	}{
		{"Windows Server", "Windows", PlatformWindows},
		{"Windows lowercase", "windows", PlatformWindows},
		{"Windows Server 2019", "Windows Server 2019", PlatformWindows},
		{"Linux explicit", "Linux", PlatformLinux},
		{"Linux lowercase", "linux", PlatformLinux},
		{"Unix", "Unix", PlatformLinux},
		{"Ubuntu", "Ubuntu", PlatformLinux},
		{"Amazon Linux", "Amazon Linux", PlatformLinux},
		{"CentOS", "centos", PlatformLinux},
		{"RHEL", "rhel", PlatformLinux},
		{"Debian", "debian", PlatformLinux},
		{"Empty string defaults to Linux", "", PlatformLinux},
		{"Unknown OS", "FreeBSD", PlatformUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.normalizePlatform(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectPlatformFromSSM(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"

	t.Run("Windows detection from SSM", func(t *testing.T) {
		mockSSM := &MockSSMClient{}
		mockEC2 := &MockEC2Client{}
		detector := NewDetector(mockSSM, mockEC2, logging.NewNoOpLogger())

		mockSSM.On("DescribeInstanceInformation", ctx, mock.Anything).Return(
			&ssm.DescribeInstanceInformationOutput{
				InstanceInformationList: []ssmtypes.InstanceInformation{
					{
						InstanceId:      aws.String(instanceID),
						PlatformType:    ssmtypes.PlatformTypeWindows,
						PlatformName:    aws.String("Microsoft Windows Server 2019 Datacenter"),
						PlatformVersion: aws.String("10.0.17763"),
					},
				},
			},
			nil,
		)

		result, err := detector.DetectPlatform(ctx, instanceID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, PlatformWindows, result.Platform)
		assert.Equal(t, ConfidenceHigh, result.Confidence)
		assert.Equal(t, "SSM", result.Source)
		assert.Equal(t, "Microsoft Windows Server 2019 Datacenter", result.PlatformName)

		mockSSM.AssertExpectations(t)
	})

	t.Run("Linux detection from SSM", func(t *testing.T) {
		mockSSM := &MockSSMClient{}
		mockEC2 := &MockEC2Client{}
		detector := NewDetector(mockSSM, mockEC2, logging.NewNoOpLogger())

		mockSSM.On("DescribeInstanceInformation", ctx, mock.Anything).Return(
			&ssm.DescribeInstanceInformationOutput{
				InstanceInformationList: []ssmtypes.InstanceInformation{
					{
						InstanceId:      aws.String(instanceID),
						PlatformType:    ssmtypes.PlatformTypeLinux,
						PlatformName:    aws.String("Amazon Linux"),
						PlatformVersion: aws.String("2"),
					},
				},
			},
			nil,
		)

		result, err := detector.DetectPlatform(ctx, instanceID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, PlatformLinux, result.Platform)
		assert.Equal(t, ConfidenceHigh, result.Confidence)
		assert.Equal(t, "SSM", result.Source)

		mockSSM.AssertExpectations(t)
	})
}

func TestDetectPlatformFromEC2(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"

	t.Run("Windows detection from EC2", func(t *testing.T) {
		mockSSM := &MockSSMClient{}
		mockEC2 := &MockEC2Client{}
		detector := NewDetector(mockSSM, mockEC2, logging.NewNoOpLogger())

		// SSM fails
		mockSSM.On("DescribeInstanceInformation", ctx, mock.Anything).Return(
			nil,
			assert.AnError,
		)

		// EC2 succeeds
		windowsPlatform := ec2types.PlatformValuesWindows
		mockEC2.On("DescribeInstances", ctx, mock.Anything).Return(
			&ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{
					{
						Instances: []ec2types.Instance{
							{
								InstanceId: aws.String(instanceID),
								Platform:   windowsPlatform,
							},
						},
					},
				},
			},
			nil,
		)

		result, err := detector.DetectPlatform(ctx, instanceID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, PlatformWindows, result.Platform)
		assert.Equal(t, ConfidenceMedium, result.Confidence)
		assert.Equal(t, "EC2", result.Source)

		mockSSM.AssertExpectations(t)
		mockEC2.AssertExpectations(t)
	})

	t.Run("Linux detection from EC2 (empty platform)", func(t *testing.T) {
		mockSSM := &MockSSMClient{}
		mockEC2 := &MockEC2Client{}
		detector := NewDetector(mockSSM, mockEC2, logging.NewNoOpLogger())

		// SSM fails
		mockSSM.On("DescribeInstanceInformation", ctx, mock.Anything).Return(
			nil,
			assert.AnError,
		)

		// EC2 succeeds with empty platform (Linux)
		mockEC2.On("DescribeInstances", ctx, mock.Anything).Return(
			&ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{
					{
						Instances: []ec2types.Instance{
							{
								InstanceId: aws.String(instanceID),
								Platform:   "", // Empty platform typically means Linux
							},
						},
					},
				},
			},
			nil,
		)

		result, err := detector.DetectPlatform(ctx, instanceID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, PlatformLinux, result.Platform)
		assert.Equal(t, ConfidenceMedium, result.Confidence)
		assert.Equal(t, "EC2", result.Source)

		mockSSM.AssertExpectations(t)
		mockEC2.AssertExpectations(t)
	})
}

func TestDetectPlatformFallback(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"

	t.Run("Fallback to default when all APIs fail", func(t *testing.T) {
		mockSSM := &MockSSMClient{}
		mockEC2 := &MockEC2Client{}
		detector := NewDetector(mockSSM, mockEC2, logging.NewNoOpLogger())

		// Both SSM and EC2 fail
		mockSSM.On("DescribeInstanceInformation", ctx, mock.Anything).Return(
			nil,
			assert.AnError,
		)
		mockEC2.On("DescribeInstances", ctx, mock.Anything).Return(
			nil,
			assert.AnError,
		)

		result, err := detector.DetectPlatform(ctx, instanceID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, PlatformLinux, result.Platform)
		assert.Equal(t, ConfidenceLow, result.Confidence)
		assert.Equal(t, "default", result.Source)

		mockSSM.AssertExpectations(t)
		mockEC2.AssertExpectations(t)
	})
}

func TestCaching(t *testing.T) {
	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"

	t.Run("Cache hit prevents API calls", func(t *testing.T) {
		mockSSM := &MockSSMClient{}
		mockEC2 := &MockEC2Client{}
		detector := NewDetector(mockSSM, mockEC2, logging.NewNoOpLogger())

		// First call - APIs are invoked
		mockSSM.On("DescribeInstanceInformation", ctx, mock.Anything).Return(
			&ssm.DescribeInstanceInformationOutput{
				InstanceInformationList: []ssmtypes.InstanceInformation{
					{
						InstanceId:   aws.String(instanceID),
						PlatformType: ssmtypes.PlatformTypeWindows,
					},
				},
			},
			nil,
		).Once() // Important: only expect one call

		// First detection
		result1, err := detector.DetectPlatform(ctx, instanceID)
		assert.NoError(t, err)
		assert.Equal(t, PlatformWindows, result1.Platform)

		// Second detection (should use cache)
		result2, err := detector.DetectPlatform(ctx, instanceID)
		assert.NoError(t, err)
		assert.Equal(t, PlatformWindows, result2.Platform)

		// Verify SSM was only called once
		mockSSM.AssertExpectations(t)
	})

	t.Run("Cache expiry causes new API call", func(t *testing.T) {
		mockSSM := &MockSSMClient{}
		mockEC2 := &MockEC2Client{}
		detector := NewDetector(mockSSM, mockEC2, logging.NewNoOpLogger())
		detector.SetCacheTTL(100 * time.Millisecond) // Short TTL for testing

		// Expect two calls to SSM
		mockSSM.On("DescribeInstanceInformation", ctx, mock.Anything).Return(
			&ssm.DescribeInstanceInformationOutput{
				InstanceInformationList: []ssmtypes.InstanceInformation{
					{
						InstanceId:   aws.String(instanceID),
						PlatformType: ssmtypes.PlatformTypeLinux,
					},
				},
			},
			nil,
		).Twice()

		// First detection
		result1, err := detector.DetectPlatform(ctx, instanceID)
		assert.NoError(t, err)
		assert.Equal(t, PlatformLinux, result1.Platform)

		// Wait for cache to expire
		time.Sleep(150 * time.Millisecond)

		// Second detection (cache expired, should call API again)
		result2, err := detector.DetectPlatform(ctx, instanceID)
		assert.NoError(t, err)
		assert.Equal(t, PlatformLinux, result2.Platform)

		mockSSM.AssertExpectations(t)
	})

	t.Run("ClearCache removes all cached entries", func(t *testing.T) {
		mockSSM := &MockSSMClient{}
		mockEC2 := &MockEC2Client{}
		detector := NewDetector(mockSSM, mockEC2, logging.NewNoOpLogger())

		// Expect two calls to SSM
		mockSSM.On("DescribeInstanceInformation", ctx, mock.Anything).Return(
			&ssm.DescribeInstanceInformationOutput{
				InstanceInformationList: []ssmtypes.InstanceInformation{
					{
						InstanceId:   aws.String(instanceID),
						PlatformType: ssmtypes.PlatformTypeWindows,
					},
				},
			},
			nil,
		).Twice()

		// First detection
		result1, err := detector.DetectPlatform(ctx, instanceID)
		assert.NoError(t, err)
		assert.Equal(t, PlatformWindows, result1.Platform)

		// Clear cache
		detector.ClearCache()

		// Second detection (cache cleared, should call API again)
		result2, err := detector.DetectPlatform(ctx, instanceID)
		assert.NoError(t, err)
		assert.Equal(t, PlatformWindows, result2.Platform)

		mockSSM.AssertExpectations(t)
	})
}
