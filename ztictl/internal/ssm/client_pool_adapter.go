package ssm

import (
	"context"

	"ztictl/pkg/aws"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// ClientPoolAdapter adapts the SSM ClientPool to implement the aws.ClientPoolInterface
type ClientPoolAdapter struct {
	pool *ClientPool
}

// NewClientPoolAdapter creates a new adapter for the SSM ClientPool
func NewClientPoolAdapter(pool *ClientPool) *ClientPoolAdapter {
	return &ClientPoolAdapter{pool: pool}
}

// GetSSMClient returns an SSM client for the specified region
func (a *ClientPoolAdapter) GetSSMClient(ctx context.Context, region string) (*ssm.Client, error) {
	return a.pool.GetSSMClient(ctx, region)
}

// GetEC2Client returns an EC2 client for the specified region
func (a *ClientPoolAdapter) GetEC2Client(ctx context.Context, region string) (*ec2.Client, error) {
	return a.pool.GetEC2Client(ctx, region)
}

// Ensure ClientPoolAdapter implements the interface
var _ aws.ClientPoolInterface = (*ClientPoolAdapter)(nil)
