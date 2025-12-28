package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"ztictl/internal/ssm"
	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
)

// ssmSSHCmd represents the ssm ssh command
var ssmSSHCmd = &cobra.Command{
	Use:   "ssh [instance-identifier]",
	Short: "SSH to an instance through SSM tunnel",
	Long: `Connect to an EC2 instance via SSH using SSM Session Manager as the transport.
This allows SSH access without opening port 22 or managing bastion hosts.

If no instance identifier is provided, an interactive fuzzy finder will be launched.
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Requirements:
  - AWS CLI v2 with Session Manager plugin installed
  - SSH client installed
  - EC2 instance must have SSM agent running

Examples:
  ztictl ssm ssh i-1234567890abcdef0 --region ca-central-1
  ztictl ssm ssh my-server --region cac1 --user ubuntu
  ztictl ssm ssh i-abc123 -r cac1 -i ~/.ssh/my-key.pem`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		user, _ := cmd.Flags().GetString("user")
		identityFile, _ := cmd.Flags().GetString("identity")
		sshArgs, _ := cmd.Flags().GetStringArray("ssh-arg")

		var instanceIdentifier string
		if len(args) > 0 {
			instanceIdentifier = args[0]
		}

		if err := performSSHConnection(regionCode, instanceIdentifier, user, identityFile, sshArgs); err != nil {
			logging.LogError("SSH connection failed: %v", err)
			os.Exit(1)
		}
	},
}

// ssmSSHConfigCmd represents the ssm ssh-config command
var ssmSSHConfigCmd = &cobra.Command{
	Use:   "ssh-config [instance-identifier]",
	Short: "Generate SSH config entry for SSM-based SSH access",
	Long: `Generate an SSH config entry that allows native SSH access through SSM.
After adding the config, you can use standard SSH commands like:
  ssh <name>
  scp file.txt <name>:/path/
  rsync -avz ./dir <name>:/path/

If no instance identifier is provided, an interactive fuzzy finder will be launched.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Examples:
  ztictl ssm ssh-config i-1234567890abcdef0 --region ca-central-1 --name prod-web
  ztictl ssm ssh-config my-server --region cac1 --name dev-api --user ubuntu
  ztictl ssm ssh-config i-abc123 -r cac1 -n jump-box --append`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		name, _ := cmd.Flags().GetString("name")
		user, _ := cmd.Flags().GetString("user")
		identityFile, _ := cmd.Flags().GetString("identity")
		appendToConfig, _ := cmd.Flags().GetBool("append")

		var instanceIdentifier string
		if len(args) > 0 {
			instanceIdentifier = args[0]
		}

		if err := generateSSHConfig(regionCode, instanceIdentifier, name, user, identityFile, appendToConfig); err != nil {
			logging.LogError("SSH config generation failed: %v", err)
			os.Exit(1)
		}
	},
}

// performSSHConnection handles SSH over SSM connection
func performSSHConnection(regionCode, instanceIdentifier, user, identityFile string, extraArgs []string) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()
	ssmManager := ssm.NewManager(logger)

	// Use the shared instance selection logic
	instanceID, err := ssmManager.GetInstanceService().SelectInstanceWithFallback(
		ctx,
		instanceIdentifier,
		region,
		nil, // No filters
	)
	if err != nil {
		return fmt.Errorf("instance selection failed: %w", err)
	}

	// Validate instance state before attempting connection
	if err := ValidateInstanceState(ctx, ssmManager, instanceID, region, InstanceValidationRequirements{
		AllowedStates:    []string{"running"},
		RequireSSMOnline: true,
		Operation:        "ssh",
	}); err != nil {
		return err
	}

	// Default user based on common AMIs
	if user == "" {
		user = "ec2-user"
	}

	logging.LogInfo("Starting SSH connection to %s@%s via SSM in region: %s", user, instanceID, region)

	// Build SSH command with ProxyCommand
	sshCmd := getSSHCommand()
	proxyCommand := buildProxyCommand(instanceID, region)

	sshArgs := []string{
		"-o", fmt.Sprintf("ProxyCommand=%s", proxyCommand),
		"-o", "StrictHostKeyChecking=accept-new",
	}

	if identityFile != "" {
		sshArgs = append(sshArgs, "-i", identityFile)
	}

	// Add any extra SSH arguments
	sshArgs = append(sshArgs, extraArgs...)

	// Add target
	sshArgs = append(sshArgs, fmt.Sprintf("%s@%s", user, instanceID))

	cmd := exec.CommandContext(ctx, sshCmd, sshArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	return nil
}

// generateSSHConfig generates an SSH config entry for SSM-based SSH access
func generateSSHConfig(regionCode, instanceIdentifier, name, user, identityFile string, appendToConfig bool) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()
	ssmManager := ssm.NewManager(logger)

	// Use the shared instance selection logic
	instanceID, err := ssmManager.GetInstanceService().SelectInstanceWithFallback(
		ctx,
		instanceIdentifier,
		region,
		nil, // No filters
	)
	if err != nil {
		return fmt.Errorf("instance selection failed: %w", err)
	}

	// Use instance ID as name if not provided
	if name == "" {
		name = instanceID
	}

	// Default user
	if user == "" {
		user = "ec2-user"
	}

	proxyCommand := buildProxyCommand(instanceID, region)

	// Build SSH config entry
	var configBuilder strings.Builder
	configBuilder.WriteString(fmt.Sprintf("\n# Added by ztictl for SSM SSH access\n"))
	configBuilder.WriteString(fmt.Sprintf("Host %s\n", name))
	configBuilder.WriteString(fmt.Sprintf("    HostName %s\n", instanceID))
	configBuilder.WriteString(fmt.Sprintf("    User %s\n", user))
	configBuilder.WriteString(fmt.Sprintf("    ProxyCommand %s\n", proxyCommand))
	configBuilder.WriteString("    StrictHostKeyChecking accept-new\n")

	if identityFile != "" {
		configBuilder.WriteString(fmt.Sprintf("    IdentityFile %s\n", identityFile))
	}

	configEntry := configBuilder.String()

	if appendToConfig {
		// Append to ~/.ssh/config
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		sshDir := filepath.Join(homeDir, ".ssh")
		configPath := filepath.Join(sshDir, "config")

		// Ensure .ssh directory exists
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			return fmt.Errorf("failed to create .ssh directory: %w", err)
		}

		// Check if entry already exists
		existingConfig, _ := os.ReadFile(configPath)
		if strings.Contains(string(existingConfig), fmt.Sprintf("Host %s\n", name)) {
			return fmt.Errorf("SSH config entry for '%s' already exists in %s", name, configPath)
		}

		// Append to config file
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open SSH config: %w", err)
		}
		defer f.Close()

		if _, err := f.WriteString(configEntry); err != nil {
			return fmt.Errorf("failed to write SSH config: %w", err)
		}

		logging.LogInfo("SSH config entry added to %s", configPath)
		fmt.Printf("\nYou can now connect using:\n")
		fmt.Printf("  ssh %s\n", name)
		fmt.Printf("  scp file.txt %s:/path/\n", name)
		fmt.Printf("  rsync -avz ./dir %s:/path/\n", name)
	} else {
		// Print to stdout
		fmt.Printf("# Add the following to your ~/.ssh/config:\n")
		fmt.Print(configEntry)
		fmt.Printf("\n# Or run with --append to automatically add it:\n")
		fmt.Printf("#   ztictl ssm ssh-config %s --region %s --name %s --append\n", instanceID, regionCode, name)
	}

	return nil
}

// ssmRDPCmd represents the ssm rdp command
var ssmRDPCmd = &cobra.Command{
	Use:   "rdp [instance-identifier]",
	Short: "RDP to a Windows instance through SSM tunnel",
	Long: `Connect to a Windows EC2 instance via RDP using SSM Session Manager as the transport.
This allows RDP access without opening port 3389 or managing bastion hosts.

The command sets up port forwarding from a local port to the instance's RDP port (3389),
then optionally launches your RDP client.

If no instance identifier is provided, an interactive fuzzy finder will be launched.
Instance identifier can be an instance ID (i-1234567890abcdef0) or instance name.
Region supports shortcuts: cac1 (ca-central-1), use1 (us-east-1), euw1 (eu-west-1), etc.

Requirements:
  - AWS CLI v2 with Session Manager plugin installed
  - Windows instance with RDP enabled
  - EC2 instance must have SSM agent running

Examples:
  ztictl ssm rdp i-1234567890abcdef0 --region ca-central-1
  ztictl ssm rdp my-windows-server --region cac1
  ztictl ssm rdp i-abc123 -r cac1 --local-port 13389
  ztictl ssm rdp i-abc123 -r cac1 --launch`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		regionCode, _ := cmd.Flags().GetString("region")
		localPort, _ := cmd.Flags().GetInt("local-port")
		launch, _ := cmd.Flags().GetBool("launch")

		var instanceIdentifier string
		if len(args) > 0 {
			instanceIdentifier = args[0]
		}

		if err := performRDPConnection(regionCode, instanceIdentifier, localPort, launch); err != nil {
			logging.LogError("RDP connection failed: %v", err)
			os.Exit(1)
		}
	},
}

// performRDPConnection handles RDP over SSM connection
func performRDPConnection(regionCode, instanceIdentifier string, localPort int, launch bool) error {
	region := resolveRegion(regionCode)
	ctx := context.Background()
	ssmManager := ssm.NewManager(logger)

	// Use the shared instance selection logic
	instanceID, err := ssmManager.GetInstanceService().SelectInstanceWithFallback(
		ctx,
		instanceIdentifier,
		region,
		nil, // No filters
	)
	if err != nil {
		return fmt.Errorf("instance selection failed: %w", err)
	}

	// Validate instance state before attempting connection
	if err := ValidateInstanceState(ctx, ssmManager, instanceID, region, InstanceValidationRequirements{
		AllowedStates:    []string{"running"},
		RequireSSMOnline: true,
		Operation:        "rdp",
	}); err != nil {
		return err
	}

	// Default local port
	if localPort == 0 {
		localPort = 33389 // Use non-standard port to avoid conflicts
	}

	remotePort := 3389 // Standard RDP port

	logging.LogInfo("Starting RDP tunnel to %s via SSM in region: %s", instanceID, region)
	fmt.Printf("\nRDP tunnel: localhost:%d -> %s:%d\n", localPort, instanceID, remotePort)
	fmt.Printf("\nConnect your RDP client to: localhost:%d\n", localPort)

	if launch {
		// Launch RDP client in background before starting the tunnel
		go func() {
			// Give the tunnel a moment to start
			launchRDPClient(localPort)
		}()
	} else {
		fmt.Printf("\nTo auto-launch RDP client, use: --launch\n")
		if runtime.GOOS == "windows" {
			fmt.Printf("Or manually run: mstsc /v:localhost:%d\n", localPort)
		} else if runtime.GOOS == "darwin" {
			fmt.Printf("Or manually run: open rdp://localhost:%d\n", localPort)
		} else {
			fmt.Printf("Or manually run: xfreerdp /v:localhost:%d /u:Administrator\n", localPort)
		}
	}

	fmt.Printf("\nPress Ctrl+C to stop the tunnel\n\n")

	// Start port forwarding (this blocks until interrupted)
	if err := ssmManager.ForwardPort(ctx, instanceID, region, localPort, remotePort); err != nil {
		return fmt.Errorf("port forwarding failed: %w", err)
	}

	return nil
}

// launchRDPClient attempts to launch the system's RDP client
func launchRDPClient(port int) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows: use mstsc (Microsoft Remote Desktop)
		cmd = exec.Command("mstsc", fmt.Sprintf("/v:localhost:%d", port))
	case "darwin":
		// macOS: use open with rdp:// URL (requires Microsoft Remote Desktop app)
		cmd = exec.Command("open", fmt.Sprintf("rdp://localhost:%d", port))
	default:
		// Linux: try xfreerdp, rdesktop, or remmina
		if _, err := exec.LookPath("xfreerdp"); err == nil {
			cmd = exec.Command("xfreerdp", fmt.Sprintf("/v:localhost:%d", port), "/u:Administrator")
		} else if _, err := exec.LookPath("rdesktop"); err == nil {
			cmd = exec.Command("rdesktop", fmt.Sprintf("localhost:%d", port))
		} else if _, err := exec.LookPath("remmina"); err == nil {
			cmd = exec.Command("remmina", "-c", fmt.Sprintf("rdp://localhost:%d", port))
		} else {
			logging.LogWarn("No RDP client found. Install xfreerdp, rdesktop, or remmina")
			return
		}
	}

	if cmd != nil {
		if err := cmd.Start(); err != nil {
			logging.LogWarn("Failed to launch RDP client: %v", err)
		}
	}
}

// buildProxyCommand builds the AWS SSM ProxyCommand for SSH
func buildProxyCommand(instanceID, region string) string {
	awsCmd := "aws"
	if runtime.GOOS == "windows" {
		awsCmd = "aws.exe"
	}

	// Using AWS-StartSSHSession document for SSH over SSM
	return fmt.Sprintf("%s ssm start-session --target %s --document-name AWS-StartSSHSession --parameters portNumber=%%p --region %s",
		awsCmd, instanceID, region)
}

// getSSHCommand returns the platform-appropriate SSH command
func getSSHCommand() string {
	if runtime.GOOS == "windows" {
		return "ssh.exe"
	}
	return "ssh"
}

func init() {
	// SSH command flags
	ssmSSHCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmSSHCmd.Flags().StringP("user", "u", "", "SSH username (default: ec2-user)")
	ssmSSHCmd.Flags().StringP("identity", "i", "", "Path to SSH private key file")
	ssmSSHCmd.Flags().StringArrayP("ssh-arg", "o", []string{}, "Additional SSH arguments (can be specified multiple times)")

	// SSH config command flags
	ssmSSHConfigCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmSSHConfigCmd.Flags().StringP("name", "n", "", "Friendly name for SSH config entry (default: instance ID)")
	ssmSSHConfigCmd.Flags().StringP("user", "u", "", "SSH username (default: ec2-user)")
	ssmSSHConfigCmd.Flags().StringP("identity", "i", "", "Path to SSH private key file")
	ssmSSHConfigCmd.Flags().BoolP("append", "a", false, "Append config entry to ~/.ssh/config")

	// RDP command flags
	ssmRDPCmd.Flags().StringP("region", "r", "", "AWS region or shortcode (cac1, use1, euw1, etc.) - default from config")
	ssmRDPCmd.Flags().IntP("local-port", "p", 33389, "Local port to forward (default: 33389)")
	ssmRDPCmd.Flags().BoolP("launch", "l", false, "Automatically launch RDP client")
}
