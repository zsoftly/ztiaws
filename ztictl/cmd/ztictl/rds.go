package main

import (
	"github.com/spf13/cobra"
)

// RDS command - Main orchestrator for RDS operations
var rdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "RDS operations",
	Long: `Manage AWS RDS (Relational Database Service) operations including
starting, stopping, rebooting, and listing database instances.

Examples:
  ztictl rds list                       # List RDS instances
  ztictl rds start <db-identifier>      # Start a stopped RDS instance
  ztictl rds stop <db-identifier>       # Stop a running RDS instance
  ztictl rds reboot <db-identifier>     # Reboot an RDS instance`,
}

func init() {
	rootCmd.AddCommand(rdsCmd)

	// Add subcommands
	rdsCmd.AddCommand(rdsListCmd)   // rds_power.go
	rdsCmd.AddCommand(rdsStartCmd)  // rds_power.go
	rdsCmd.AddCommand(rdsStopCmd)   // rds_power.go
	rdsCmd.AddCommand(rdsRebootCmd) // rds_power.go
}
