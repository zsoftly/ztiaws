package testutil

import "os"

// Mock AWS credentials for testing
// These are well-known example credentials from AWS documentation
// See: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html
const (
	MockAWSAccessKeyID     = "AKIAIOSFODNN7EXAMPLE"                     // #nosec G101
	MockAWSSecretAccessKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" // #nosec G101
	MockAWSSessionToken    = "test-session-token"
	MockAWSRegion          = "ca-central-1"
)

// SetupAWSTestEnvironment configures environment variables for AWS SDK testing.
// This prevents the SDK from attempting to fetch real credentials from IMDS,
// credential files, or other sources during tests.
func SetupAWSTestEnvironment() {
	// Disable EC2 IMDS to prevent timeout attempts
	os.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	// Set mock credentials that allow AWS SDK to initialize
	os.Setenv("AWS_ACCESS_KEY_ID", MockAWSAccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", MockAWSSecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", MockAWSSessionToken)
	os.Setenv("AWS_REGION", MockAWSRegion)
}
