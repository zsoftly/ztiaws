//go:build !integration
// +build !integration

package ssm

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// testConfig returns a test AWS config that skips credential checks
func testConfig(t *testing.T) aws.Config {
	t.Helper()

	// Set environment to disable IMDS for tests
	os.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	// Create a config with dummy credentials for unit tests
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "test-access-key",
				SecretAccessKey: "test-secret-key",
				SessionToken:    "test-session-token",
				Source:          "test",
			}, nil
		})),
	)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	return cfg
}
