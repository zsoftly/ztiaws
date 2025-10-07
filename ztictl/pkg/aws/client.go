package aws

import (
	"context"

	"ztictl/pkg/errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Client wraps AWS service clients with common configuration
type Client struct {
	Config  aws.Config
	EC2     *ec2.Client
	SSM     *ssm.Client
	STS     *sts.Client
	SSO     *sso.Client
	SSOOIDC *ssooidc.Client
}

// ClientOptions configures the AWS client
type ClientOptions struct {
	Region  string
	Profile string
}

// NewClient creates a new AWS client with the specified options
func NewClient(ctx context.Context, opts ClientOptions) (*Client, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(opts.Region),
		config.WithSharedConfigProfile(opts.Profile),
	)
	if err != nil {
		return nil, errors.NewAWSError("failed to load AWS configuration", err)
	}

	return &Client{
		Config:  cfg,
		EC2:     ec2.NewFromConfig(cfg),
		SSM:     ssm.NewFromConfig(cfg),
		STS:     sts.NewFromConfig(cfg),
		SSO:     sso.NewFromConfig(cfg),
		SSOOIDC: ssooidc.NewFromConfig(cfg),
	}, nil
}

// NewClientWithRegion creates a new client for a specific region
func (c *Client) NewClientWithRegion(region string) *Client {
	newConfig := c.Config.Copy()
	newConfig.Region = region

	return &Client{
		Config:  newConfig,
		EC2:     ec2.NewFromConfig(newConfig),
		SSM:     ssm.NewFromConfig(newConfig),
		STS:     sts.NewFromConfig(newConfig),
		SSO:     sso.NewFromConfig(newConfig),
		SSOOIDC: ssooidc.NewFromConfig(newConfig),
	}
}

// GetCallerIdentity returns information about the current AWS credentials
func (c *Client) GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	output, err := c.STS.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, errors.NewAWSError("failed to get caller identity", err)
	}
	return output, nil
}

// ValidateCredentials checks if the current AWS credentials are valid
func (c *Client) ValidateCredentials(ctx context.Context) error {
	_, err := c.GetCallerIdentity(ctx)
	return err
}
