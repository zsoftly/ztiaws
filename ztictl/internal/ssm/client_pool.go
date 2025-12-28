package ssm

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type clientKey struct {
	region string
}

type clientSet struct {
	Config    aws.Config
	SSMClient *ssm.Client
	EC2Client *ec2.Client
	IAMClient *iam.Client
	S3Client  *s3.Client
	STSClient *sts.Client
	RDSClient *rds.Client
}

type ClientPool struct {
	clients map[clientKey]*clientSet
	mu      sync.RWMutex
}

func NewClientPool() *ClientPool {
	return &ClientPool{
		clients: make(map[clientKey]*clientSet),
	}
}

func (p *ClientPool) GetClients(ctx context.Context, region string) (*clientSet, error) {
	key := clientKey{region: region}

	p.mu.RLock()
	if clients, exists := p.clients[key]; exists {
		p.mu.RUnlock()
		return clients, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	if clients, exists := p.clients[key]; exists {
		return clients, nil
	}

	clients, err := p.createClientSet(ctx, region)
	if err != nil {
		return nil, err
	}

	p.clients[key] = clients

	return clients, nil
}

func (p *ClientPool) createClientSet(ctx context.Context, region string) (*clientSet, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for region %s: %w", region, err)
	}

	clients := &clientSet{
		Config:    cfg,
		SSMClient: ssm.NewFromConfig(cfg),
		EC2Client: ec2.NewFromConfig(cfg),
		IAMClient: iam.NewFromConfig(cfg),
		S3Client:  s3.NewFromConfig(cfg),
		STSClient: sts.NewFromConfig(cfg),
		RDSClient: rds.NewFromConfig(cfg),
	}

	return clients, nil
}

func (p *ClientPool) GetSSMClient(ctx context.Context, region string) (*ssm.Client, error) {
	clients, err := p.GetClients(ctx, region)
	if err != nil {
		return nil, err
	}
	return clients.SSMClient, nil
}

func (p *ClientPool) GetEC2Client(ctx context.Context, region string) (*ec2.Client, error) {
	clients, err := p.GetClients(ctx, region)
	if err != nil {
		return nil, err
	}
	return clients.EC2Client, nil
}

func (p *ClientPool) GetIAMClient(ctx context.Context, region string) (*iam.Client, error) {
	clients, err := p.GetClients(ctx, region)
	if err != nil {
		return nil, err
	}
	return clients.IAMClient, nil
}

func (p *ClientPool) GetS3Client(ctx context.Context, region string) (*s3.Client, error) {
	clients, err := p.GetClients(ctx, region)
	if err != nil {
		return nil, err
	}
	return clients.S3Client, nil
}

func (p *ClientPool) GetSTSClient(ctx context.Context, region string) (*sts.Client, error) {
	clients, err := p.GetClients(ctx, region)
	if err != nil {
		return nil, err
	}
	return clients.STSClient, nil
}

func (p *ClientPool) GetRDSClient(ctx context.Context, region string) (*rds.Client, error) {
	clients, err := p.GetClients(ctx, region)
	if err != nil {
		return nil, err
	}
	return clients.RDSClient, nil
}

func (p *ClientPool) GetPlatformClients(ctx context.Context, region string) (*ssm.Client, *ec2.Client, error) {
	clients, err := p.GetClients(ctx, region)
	if err != nil {
		return nil, nil, err
	}
	return clients.SSMClient, clients.EC2Client, nil
}

func (p *ClientPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clients = make(map[clientKey]*clientSet)
}

func (p *ClientPool) RegionCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients)
}
