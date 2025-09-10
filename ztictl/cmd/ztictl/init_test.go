package main

import "os"

func init() {
	// Disable EC2 IMDS for all tests to prevent AWS credential timeouts
	os.Setenv("AWS_EC2_METADATA_DISABLED", "true")
}
