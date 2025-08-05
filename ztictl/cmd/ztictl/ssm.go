package main

import (
	"github.com/spf13/cobra"
)

// SSM command - Main orchestrator (equivalent to main ssm bash script)
var ssmCmd = &cobra.Command{
	Use:   "ssm",
	Short: "SSM operations",
	Long: `Manage AWS Systems Manager operations including instance connections, command execution, 
file transfers, and port forwarding through SSM.

Examples:
  ztictl ssm connect <instance>         # Connect to instance via SSM
  ztictl ssm list [filters]             # List SSM-enabled instances  
  ztictl ssm forward <instance> <ports> # Port forwarding via SSM
  ztictl ssm transfer <src> <dst>       # File transfer via SSM
  ztictl ssm command <instance> <cmd>   # Execute command via SSM
  ztictl ssm exec <region> <instance> <cmd>    # Quick exec with region shortcode
  ztictl ssm exec-tagged <region> <tag> <cmd> # Execute on tagged instances
  ztictl ssm exec-multi [flags] <cmd>   # Execute across multiple regions
  ztictl ssm status [instance]          # Check SSM agent status`,
}

func init() {
	rootCmd.AddCommand(ssmCmd)

	// Add subcommands - each defined in separate files following bash modular pattern
	// Equivalent to sourcing individual .sh files in bash version
        ssmCmd.AddCommand(ssmExecMultiCmd)          // ssm_exec_multi.go
	ssmCmd.AddCommand(ssmConnectCmd)          // ssm_connect.go
	ssmCmd.AddCommand(ssmListCmd)             // ssm_list.go
	ssmCmd.AddCommand(ssmCommandCmd)          // ssm_command.go
	ssmCmd.AddCommand(ssmTransferCmd)         // ssm_transfer.go
	ssmCmd.AddCommand(ssmForwardCmd)          // ssm_management.go
	ssmCmd.AddCommand(ssmStatusCmd)           // ssm_management.go
	ssmCmd.AddCommand(ssmExecCmd)             // ssm_exec.go
	ssmCmd.AddCommand(ssmExecTaggedCmd)       // ssm_exec.go
	ssmCmd.AddCommand(ssmCleanupCmd)          // ssm_cleanup.go
	ssmCmd.AddCommand(ssmEmergencyCleanupCmd) // ssm_cleanup.go
}
