package main

import "ztictl/internal/testutil"

func init() {
	// Configure AWS test environment with mock credentials
	testutil.SetupAWSTestEnvironment()
}
