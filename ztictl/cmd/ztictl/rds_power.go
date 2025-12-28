package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"ztictl/internal/ssm"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/spf13/cobra"
)

// rdsClientPool is reused across RDS commands for efficiency
var rdsClientPool = ssm.NewClientPool()

// rdsListCmd represents the rds list command
var rdsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List RDS instances",
	Long: `List all RDS database instances in the specified region.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Examples:
  ztictl rds list --region cac1
  ztictl rds list -r ca-central-1`,
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")

		if err := performRDSList(regionCode); err != nil {
			logging.LogError("RDS list failed: %v", err)
			os.Exit(1)
		}
	},
}

// rdsStartCmd represents the rds start command
var rdsStartCmd = &cobra.Command{
	Use:   "start <db-identifier>",
	Short: "Start a stopped RDS instance",
	Long: `Start a stopped RDS database instance.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Note: Starting an RDS instance can take several minutes.

Examples:
  ztictl rds start my-database --region cac1
  ztictl rds start prod-db -r ca-central-1
  ztictl rds start my-database --region cac1 --wait`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		wait, _ := cmd.Flags().GetBool("wait")
		dbIdentifier := args[0]

		if err := performRDSStart(regionCode, dbIdentifier, wait); err != nil {
			logging.LogError("RDS start failed: %v", err)
			os.Exit(1)
		}
	},
}

// rdsStopCmd represents the rds stop command
var rdsStopCmd = &cobra.Command{
	Use:   "stop <db-identifier>",
	Short: "Stop a running RDS instance",
	Long: `Stop a running RDS database instance.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Note:
  - Stopping an RDS instance can take several minutes.
  - RDS instances can only be stopped for up to 7 days before they are automatically restarted.
  - Some instance types (e.g., Multi-AZ deployments with SQL Server) cannot be stopped.

Examples:
  ztictl rds stop my-database --region cac1
  ztictl rds stop dev-db -r ca-central-1
  ztictl rds stop my-database --region cac1 --wait`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		wait, _ := cmd.Flags().GetBool("wait")
		dbIdentifier := args[0]

		if err := performRDSStop(regionCode, dbIdentifier, wait); err != nil {
			logging.LogError("RDS stop failed: %v", err)
			os.Exit(1)
		}
	},
}

// rdsRebootCmd represents the rds reboot command
var rdsRebootCmd = &cobra.Command{
	Use:   "reboot <db-identifier>",
	Short: "Reboot an RDS instance",
	Long: `Reboot an RDS database instance.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Note: Rebooting an RDS instance will cause a brief outage.

Examples:
  ztictl rds reboot my-database --region cac1
  ztictl rds reboot prod-db -r ca-central-1
  ztictl rds reboot my-database --region cac1 --force-failover`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		forceFailover, _ := cmd.Flags().GetBool("force-failover")
		wait, _ := cmd.Flags().GetBool("wait")
		dbIdentifier := args[0]

		if err := performRDSReboot(regionCode, dbIdentifier, forceFailover, wait); err != nil {
			logging.LogError("RDS reboot failed: %v", err)
			os.Exit(1)
		}
	},
}

// performRDSList lists all RDS instances in the region
func performRDSList(regionCode string) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()

	rdsClient, err := rdsClientPool.GetRDSClient(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to create RDS client: %w", err)
	}

	logging.LogInfo("Listing RDS instances in region: %s", region)

	resp, err := rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return fmt.Errorf("failed to describe RDS instances: %w", err)
	}

	if len(resp.DBInstances) == 0 {
		fmt.Printf("\nNo RDS instances found in region: %s\n", region)
		return nil
	}

	fmt.Printf("\n")
	colors.PrintHeader("RDS Instances in %s:\n", region)
	colors.PrintHeader("==============================\n")
	colors.PrintHeader("%-30s %-15s %-15s %-20s %s\n", "DB Identifier", "Status", "Engine", "Class", "Endpoint")
	colors.PrintHeader("%s\n", strings.Repeat("-", 110))

	for _, db := range resp.DBInstances {
		endpoint := "-"
		if db.Endpoint != nil && db.Endpoint.Address != nil {
			endpoint = fmt.Sprintf("%s:%d", aws.ToString(db.Endpoint.Address), aws.ToInt32(db.Endpoint.Port))
		}

		status := aws.ToString(db.DBInstanceStatus)
		coloredStatus := colorizeRDSStatus(status)

		fmt.Printf("%-30s %-15s %-15s %-20s %s\n",
			aws.ToString(db.DBInstanceIdentifier),
			coloredStatus,
			aws.ToString(db.Engine),
			aws.ToString(db.DBInstanceClass),
			endpoint,
		)
	}

	fmt.Printf("\nTotal: %d instance(s)\n", len(resp.DBInstances))
	return nil
}

// performRDSStart starts a stopped RDS instance
func performRDSStart(regionCode, dbIdentifier string, wait bool) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()

	rdsClient, err := rdsClientPool.GetRDSClient(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to create RDS client: %w", err)
	}

	logging.LogInfo("Starting RDS instance %s in region: %s", dbIdentifier, region)
	fmt.Printf("Starting RDS instance: %s\n", dbIdentifier)

	_, err = rdsClient.StartDBInstance(ctx, &rds.StartDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbIdentifier),
	})
	if err != nil {
		return fmt.Errorf("failed to start RDS instance: %w", err)
	}

	colors.PrintSuccess("✓ Start command sent for RDS instance: %s\n", dbIdentifier)

	if wait {
		fmt.Printf("Waiting for instance to become available...\n")
		if err := waitForRDSStatus(ctx, rdsClient, dbIdentifier, "available"); err != nil {
			return err
		}
		colors.PrintSuccess("✓ RDS instance %s is now available\n", dbIdentifier)
	} else {
		fmt.Printf("Note: Instance startup may take several minutes. Use --wait to wait for completion.\n")
	}

	return nil
}

// performRDSStop stops a running RDS instance
func performRDSStop(regionCode, dbIdentifier string, wait bool) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()

	rdsClient, err := rdsClientPool.GetRDSClient(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to create RDS client: %w", err)
	}

	logging.LogInfo("Stopping RDS instance %s in region: %s", dbIdentifier, region)
	fmt.Printf("Stopping RDS instance: %s\n", dbIdentifier)

	_, err = rdsClient.StopDBInstance(ctx, &rds.StopDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbIdentifier),
	})
	if err != nil {
		return fmt.Errorf("failed to stop RDS instance: %w", err)
	}

	colors.PrintSuccess("✓ Stop command sent for RDS instance: %s\n", dbIdentifier)

	if wait {
		fmt.Printf("Waiting for instance to stop...\n")
		if err := waitForRDSStatus(ctx, rdsClient, dbIdentifier, "stopped"); err != nil {
			return err
		}
		colors.PrintSuccess("✓ RDS instance %s is now stopped\n", dbIdentifier)
	} else {
		fmt.Printf("Note: Instance shutdown may take several minutes. Use --wait to wait for completion.\n")
		fmt.Printf("Warning: RDS instances auto-restart after 7 days of being stopped.\n")
	}

	return nil
}

// performRDSReboot reboots an RDS instance
func performRDSReboot(regionCode, dbIdentifier string, forceFailover, wait bool) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()

	rdsClient, err := rdsClientPool.GetRDSClient(ctx, region)
	if err != nil {
		return fmt.Errorf("failed to create RDS client: %w", err)
	}

	logging.LogInfo("Rebooting RDS instance %s in region: %s (force-failover: %v)", dbIdentifier, region, forceFailover)
	fmt.Printf("Rebooting RDS instance: %s\n", dbIdentifier)

	if forceFailover {
		fmt.Printf("Force failover enabled (Multi-AZ only)\n")
	}

	_, err = rdsClient.RebootDBInstance(ctx, &rds.RebootDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbIdentifier),
		ForceFailover:        aws.Bool(forceFailover),
	})
	if err != nil {
		return fmt.Errorf("failed to reboot RDS instance: %w", err)
	}

	colors.PrintSuccess("✓ Reboot command sent for RDS instance: %s\n", dbIdentifier)

	if wait {
		fmt.Printf("Waiting for instance to become available...\n")
		// First wait for rebooting status, then available
		if err := waitForRDSStatus(ctx, rdsClient, dbIdentifier, "available"); err != nil {
			return err
		}
		colors.PrintSuccess("✓ RDS instance %s is now available\n", dbIdentifier)
	} else {
		fmt.Printf("Note: Reboot may take several minutes. Use --wait to wait for completion.\n")
	}

	return nil
}

// waitForRDSStatus waits for an RDS instance to reach the specified status
func waitForRDSStatus(ctx context.Context, rdsClient *rds.Client, dbIdentifier, targetStatus string) error {
	maxWait := 30 * time.Minute
	pollInterval := 30 * time.Second
	deadline := time.Now().Add(maxWait)
	consecutiveErrors := 0
	maxConsecutiveErrors := 3

	for time.Now().Before(deadline) {
		resp, err := rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(dbIdentifier),
		})
		if err != nil {
			consecutiveErrors++
			if consecutiveErrors > maxConsecutiveErrors {
				return fmt.Errorf("failed to check instance status after %d retries: %w", maxConsecutiveErrors, err)
			}
			// Exponential backoff for transient errors
			backoff := pollInterval * time.Duration(consecutiveErrors)
			fmt.Printf("  Error checking status, retrying in %v... (%d/%d)\n", backoff, consecutiveErrors, maxConsecutiveErrors)
			time.Sleep(backoff)
			continue
		}
		consecutiveErrors = 0

		if len(resp.DBInstances) == 0 {
			return fmt.Errorf("instance %s not found", dbIdentifier)
		}

		status := aws.ToString(resp.DBInstances[0].DBInstanceStatus)
		fmt.Printf("  Current status: %s\n", status)

		if status == targetStatus {
			return nil
		}

		time.Sleep(pollInterval)
	}

	return fmt.Errorf("timeout waiting for instance to reach %s status", targetStatus)
}

// colorizeRDSStatus returns a colored status string for RDS status
func colorizeRDSStatus(status string) string {
	switch status {
	case "available":
		return colors.ColorSuccess("%s", status)
	case "stopped":
		return colors.ColorWarning("%s", status)
	case "starting", "stopping", "rebooting", "modifying", "backing-up":
		return colors.ColorData("%s", status)
	case "failed", "incompatible-restore", "incompatible-network":
		return colors.ColorError("%s", status)
	default:
		return status
	}
}

func init() {
	// List command flags
	rdsListCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")

	// Start command flags
	rdsStartCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	rdsStartCmd.Flags().BoolP("wait", "w", false, "Wait for the instance to become available")

	// Stop command flags
	rdsStopCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	rdsStopCmd.Flags().BoolP("wait", "w", false, "Wait for the instance to stop")

	// Reboot command flags
	rdsRebootCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	rdsRebootCmd.Flags().BoolP("force-failover", "f", false, "Force a failover for Multi-AZ instances")
	rdsRebootCmd.Flags().BoolP("wait", "w", false, "Wait for the instance to become available")
}
