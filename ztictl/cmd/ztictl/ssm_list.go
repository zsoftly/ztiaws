package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"ztictl/internal/interactive"
	"ztictl/internal/ssm"
	awsservice "ztictl/pkg/aws"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
)

// ssmListCmd represents the ssm list command
var ssmListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all EC2 instances with their SSM status",
	Long: `List all EC2 instances in a region with their SSM agent status.
Shows all instances regardless of their state or SSM connectivity.
Optionally filter by tags, status, or name patterns.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		tagFilter, _ := cmd.Flags().GetString("tag")
		statusFilter, _ := cmd.Flags().GetString("status")
		nameFilter, _ := cmd.Flags().GetString("name")
		tableFormat, _ := cmd.Flags().GetBool("table")

		filters := &ssm.ListFilters{
			Tag:    tagFilter,
			Status: statusFilter,
			Name:   nameFilter,
		}

		if err := performInstanceListing(regionCode, filters, tableFormat); err != nil {
			logging.LogError("Instance listing failed: %v", err)
			os.Exit(1)
		}
	},
}

// performInstanceListing handles instance listing logic and returns errors instead of calling os.Exit
func performInstanceListing(regionCode string, filters *ssm.ListFilters, tableFormat bool) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()
	ssmManager := ssm.NewManager(logger)

	colors.PrintData("🔍 Fetching instances from region %s...\n", region)

	// Convert SSM filters to AWS filters
	awsFilters := &awsservice.ListFilters{
		Tag:    filters.Tag,
		Tags:   filters.Tags,
		Status: filters.Status,
		Name:   filters.Name,
	}

	instances, err := ssmManager.GetInstanceService().ListInstances(ctx, region, awsFilters)
	if err != nil {
		colors.PrintError("✗ Failed to list instances in region %s\n", region)
		return fmt.Errorf("failed to list instances: %w", err)
	}

	if len(instances) == 0 {
		colors.PrintWarning("⚠ No EC2 instances found in region: %s\n", region)
		return nil
	}

	colors.PrintSuccess("✓ Found %d instance(s) in region %s\n", len(instances), region)

	// Use table format if requested, otherwise use interactive fuzzy finder
	if tableFormat {
		printInstanceTable(instances, region)
		return nil
	}

	logging.LogInfo("Launching interactive instance browser...")

	// Use shared fuzzy finder (user can select or just browse)
	selected, err := interactive.SelectInstance(instances, "Browse EC2 instances")
	if err != nil {
		// User cancelled - that's OK for list command
		if err.Error() == "instance selection cancelled" {
			return nil
		}
		return fmt.Errorf("instance selection failed: %w", err)
	}

	// Display detailed instance information
	printInstanceDetails(selected, region)

	return nil
}

// printInstanceTable prints instances in a traditional table format
func printInstanceTable(instances []interactive.Instance, region string) {
	formatter := NewTableFormatter(2) // 2 spaces between columns

	// Prepare column data
	names := make([]string, len(instances))
	instanceIDs := make([]string, len(instances))
	privateIPs := make([]string, len(instances))
	publicIPs := make([]string, len(instances))
	states := make([]string, len(instances))
	ssmStatuses := make([]string, len(instances))
	platforms := make([]string, len(instances))

	for i, instance := range instances {
		// Name
		name := instance.Name
		if name == "" {
			name = "N/A"
		}
		names[i] = name

		// Instance ID
		instanceIDs[i] = instance.InstanceID

		// Private IP Address
		privateIPs[i] = instance.PrivateIPAddress

		// Public IP Address
		publicIP := instance.PublicIPAddress
		if publicIP == "" {
			publicIP = "N/A"
		}
		publicIPs[i] = publicIP

		// State
		states[i] = instance.State

		// Format SSM status with color indicators
		var ssmStatus string
		switch instance.SSMStatus {
		case "Online":
			ssmStatus = colors.ColorSuccess("✓ Online")
		case "ConnectionLost":
			ssmStatus = colors.ColorWarning("⚠ Lost")
		case "No Agent":
			ssmStatus = colors.ColorError("✗ No Agent")
		default:
			if instance.SSMStatus == "" {
				ssmStatus = colors.ColorError("✗ No Agent")
			} else {
				ssmStatus = colors.ColorWarning("? %s", instance.SSMStatus)
			}
		}
		ssmStatuses[i] = ssmStatus

		// Platform
		platforms[i] = instance.Platform
	}

	// Add columns to formatter
	formatter.AddColumn("Name", names, 8)
	formatter.AddColumn("Instance ID", instanceIDs, 12)
	formatter.AddColumn("Private IP", privateIPs, 10)
	formatter.AddColumn("Public IP", publicIPs, 10)
	formatter.AddColumn("State", states, 8)
	formatter.AddColumn("SSM Status", ssmStatuses, 10)
	formatter.AddColumn("Platform", platforms, 8)

	fmt.Printf("\n")
	colors.PrintHeader("All EC2 Instances in %s:\n", region)
	colors.PrintHeader("=====================================\n")

	// Print formatted header
	headerStr := formatter.FormatHeader()
	colors.PrintHeader("%s\n", headerStr)

	// Print formatted rows
	for i := 0; i < formatter.GetRowCount(); i++ {
		rowStr := formatter.FormatRow(i)
		fmt.Printf("%s\n", rowStr)
	}

	fmt.Printf("\n")
	colors.PrintData("Total: %d instances\n", len(instances))
	fmt.Printf("Note: Only instances with %s SSM status can be connected to via SSM\n", colors.ColorSuccess("'✓ Online'"))
	colors.PrintData("Usage: ztictl ssm connect <instance-id-or-name>\n")
}

// printInstanceDetails displays detailed information about the selected instance
func printInstanceDetails(instance *interactive.Instance, region string) {
	fmt.Printf("\n")
	colors.PrintHeader("═══════════════════════════════════════════════════════════════════\n")
	colors.PrintHeader("                    Instance Details\n")
	colors.PrintHeader("═══════════════════════════════════════════════════════════════════\n")
	fmt.Printf("\n")

	// Basic Information
	colors.PrintHeader("📋 Basic Information:\n")
	fmt.Printf("  Name:           %s\n", colors.ColorSuccess("%s", instance.Name))
	fmt.Printf("  Instance ID:    %s\n", colors.ColorData("%s", instance.InstanceID))
	fmt.Printf("  State:          %s\n", getColoredState(instance.State))
	fmt.Printf("  Platform:       %s\n", colors.ColorData("%s", instance.Platform))
	fmt.Printf("\n")

	// Network Information
	colors.PrintHeader("🌐 Network:\n")
	fmt.Printf("  Private IP:     %s\n", colors.ColorData("%s", instance.PrivateIPAddress))
	if instance.PublicIPAddress != "" {
		fmt.Printf("  Public IP:      %s\n", colors.ColorSuccess("%s", instance.PublicIPAddress))
	} else {
		fmt.Printf("  Public IP:      %s\n", colors.ColorWarning("%s", "N/A (private instance)"))
	}
	fmt.Printf("\n")

	// SSM Status
	colors.PrintHeader("🔌 SSM Agent:\n")
	var ssmStatus string
	switch instance.SSMStatus {
	case "Online":
		ssmStatus = colors.ColorSuccess("%s", "✓ Online - Ready for connections")
	case "ConnectionLost":
		ssmStatus = colors.ColorWarning("%s", "⚠ Connection Lost - May need troubleshooting")
	case "No Agent":
		ssmStatus = colors.ColorError("%s", "✗ No Agent - SSM not available")
	default:
		if instance.SSMStatus == "" {
			ssmStatus = colors.ColorError("%s", "✗ No Agent - SSM not available")
		} else {
			ssmStatus = colors.ColorWarning("? %s", instance.SSMStatus)
		}
	}
	fmt.Printf("  Status:         %s\n", ssmStatus)
	if instance.SSMAgentVersion != "" {
		fmt.Printf("  Agent Version:  %s\n", colors.ColorData("%s", instance.SSMAgentVersion))
	}
	if instance.LastPingDateTime != "" {
		fmt.Printf("  Last Ping:      %s\n", colors.ColorData("%s", instance.LastPingDateTime))
	}
	fmt.Printf("\n")

	// Tags
	if len(instance.Tags) > 0 {
		colors.PrintHeader("🏷️  Tags:\n")
		for key, value := range instance.Tags {
			fmt.Printf("  %s = %s\n", colors.ColorData("%s", key), colors.ColorSuccess("%s", value))
		}
		fmt.Printf("\n")
	}

	// Command Examples
	if instance.SSMStatus == "Online" {
		colors.PrintHeader("💻 Quick Commands:\n")
		fmt.Printf("\n")

		// Connect command
		colors.PrintData("  Connect to instance:\n")
		connectCmd := fmt.Sprintf("ztictl ssm connect %s --region %s", instance.InstanceID, region)
		fmt.Printf("    %s\n", colors.ColorSuccess("%s", connectCmd))
		fmt.Printf("\n")

		// Execute command
		colors.PrintData("  Execute remote command:\n")
		if instance.Platform == "Windows" || containsIgnoreCase(instance.Platform, "windows") {
			execCmd := fmt.Sprintf("ztictl ssm exec %s %s \"Get-Process\"", region, instance.InstanceID)
			fmt.Printf("    %s\n", colors.ColorSuccess("%s", execCmd))
		} else {
			execCmd := fmt.Sprintf("ztictl ssm exec %s %s \"uptime\"", region, instance.InstanceID)
			fmt.Printf("    %s\n", colors.ColorSuccess("%s", execCmd))
		}
		fmt.Printf("\n")

		// File transfer commands
		colors.PrintData("  Transfer files:\n")
		var uploadCmd, downloadCmd string
		if instance.Platform == "Windows" || containsIgnoreCase(instance.Platform, "windows") {
			uploadCmd = fmt.Sprintf("ztictl ssm transfer upload %s /local/file.txt C:\\\\remote\\\\file.txt --region %s", instance.InstanceID, region)
			downloadCmd = fmt.Sprintf("ztictl ssm transfer download %s C:\\\\remote\\\\file.txt /local/file.txt --region %s", instance.InstanceID, region)
		} else {
			uploadCmd = fmt.Sprintf("ztictl ssm transfer upload %s /local/file.txt /remote/file.txt --region %s", instance.InstanceID, region)
			downloadCmd = fmt.Sprintf("ztictl ssm transfer download %s /remote/file.txt /local/file.txt --region %s", instance.InstanceID, region)
		}
		fmt.Printf("    Upload:   %s\n", colors.ColorSuccess("%s", uploadCmd))
		fmt.Printf("    Download: %s\n", colors.ColorSuccess("%s", downloadCmd))
		fmt.Printf("\n")
	} else {
		colors.PrintWarning("⚠ SSM commands unavailable - Instance must have SSM Agent Online\n")
		fmt.Printf("\n")
		colors.PrintData("  To enable SSM:\n")
		fmt.Printf("    1. Ensure the instance has the SSM agent installed\n")
		fmt.Printf("    2. Verify the instance has the required IAM role (AmazonSSMManagedInstanceCore)\n")
		fmt.Printf("    3. Check network connectivity to SSM endpoints\n")
		fmt.Printf("\n")
	}

	colors.PrintHeader("═══════════════════════════════════════════════════════════════════\n")
	fmt.Printf("\n")
}

// getColoredState returns a colored state string
func getColoredState(state string) string {
	switch state {
	case "running":
		return colors.ColorSuccess("%s", "● running")
	case "stopped":
		return colors.ColorError("%s", "○ stopped")
	case "pending":
		return colors.ColorWarning("%s", "◐ pending")
	case "stopping":
		return colors.ColorWarning("%s", "◑ stopping")
	case "terminated":
		return colors.ColorError("%s", "✗ terminated")
	default:
		return colors.ColorData("%s", state)
	}
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func init() {
	ssmListCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmListCmd.Flags().StringP("tag", "t", "", "Filter by tag (format: key=value)")
	ssmListCmd.Flags().StringP("status", "s", "", "Filter by status (running, stopped, etc.)")
	ssmListCmd.Flags().StringP("name", "n", "", "Filter by name pattern")
	ssmListCmd.Flags().Bool("table", false, "Display instances in table format instead of interactive fuzzy finder")
}
