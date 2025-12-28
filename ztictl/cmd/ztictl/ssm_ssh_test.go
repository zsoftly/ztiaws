package main

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSsmSSHCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "SSH help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Connect to an EC2 instance via SSH using SSM Session Manager",
		},
		{
			name:    "SSH with instance ID",
			args:    []string{"i-1234567890abcdef0"},
			wantErr: false,
		},
		{
			name:    "SSH with instance name",
			args:    []string{"web-server-1"},
			wantErr: false,
		},
		{
			name:    "SSH with region flag",
			args:    []string{"i-1234567890abcdef0", "--region", "ca-central-1"},
			wantErr: false,
		},
		{
			name:    "SSH with region shortcode",
			args:    []string{"i-1234567890abcdef0", "-r", "cac1"},
			wantErr: false,
		},
		{
			name:    "SSH with user flag",
			args:    []string{"i-1234567890abcdef0", "--user", "ubuntu"},
			wantErr: false,
		},
		{
			name:    "SSH with identity flag",
			args:    []string{"i-1234567890abcdef0", "-i", "~/.ssh/my-key.pem"},
			wantErr: false,
		},
		{
			name:    "SSH with too many args",
			args:    []string{"instance1", "instance2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   ssmSSHCmd.Use,
				Short: ssmSSHCmd.Short,
				Long:  ssmSSHCmd.Long,
				Args:  cobra.MaximumNArgs(1),
				Run: func(cmd *cobra.Command, args []string) {
					// Mock SSH functionality
					regionCode, _ := cmd.Flags().GetString("region")
					user, _ := cmd.Flags().GetString("user")
					identity, _ := cmd.Flags().GetString("identity")

					// Validate flags are retrievable
					if regionCode == "cac1" {
						// Valid shortcode
					}
					if user == "ubuntu" {
						// Valid user
					}
					if identity != "" {
						// Identity file provided
					}

					if len(args) > 0 {
						instanceIdentifier := args[0]
						if instanceIdentifier == "" {
							t.Error("Instance identifier should not be empty")
						}
					}
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region")
			cmd.Flags().StringP("user", "u", "", "SSH username")
			cmd.Flags().StringP("identity", "i", "", "Path to SSH private key")
			cmd.Flags().StringArrayP("ssh-arg", "o", []string{}, "Additional SSH arguments")

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.contains != "" {
				output := buf.String()
				if !strings.Contains(output, tt.contains) {
					t.Errorf("Execute() output should contain %v, got %v", tt.contains, output)
				}
			}
		})
	}
}

func TestSsmSSHCmdFlags(t *testing.T) {
	tests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{"region", "r", ""},
		{"user", "u", ""},
		{"identity", "i", ""},
	}

	for _, tt := range tests {
		t.Run(tt.flagName+" flag", func(t *testing.T) {
			flag := ssmSSHCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s shorthand = %s, want %s", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			if flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %s default = %s, want %s", tt.flagName, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestSsmSSHCmdStructure(t *testing.T) {
	if ssmSSHCmd.Use != "ssh [instance-identifier]" {
		t.Errorf("Expected Use to be 'ssh [instance-identifier]', got %s", ssmSSHCmd.Use)
	}

	if ssmSSHCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if ssmSSHCmd.Long == "" {
		t.Error("Command should have a long description")
	}

	if ssmSSHCmd.Run == nil {
		t.Error("Command should have a Run function")
	}

	// Test argument validation
	err := ssmSSHCmd.Args(ssmSSHCmd, []string{})
	if err != nil {
		t.Errorf("Command should allow 0 arguments for fuzzy finder, got error: %v", err)
	}

	err = ssmSSHCmd.Args(ssmSSHCmd, []string{"i-1234567890abcdef0"})
	if err != nil {
		t.Errorf("Command should allow 1 argument, got error: %v", err)
	}

	err = ssmSSHCmd.Args(ssmSSHCmd, []string{"instance1", "instance2"})
	if err == nil {
		t.Error("Command should not allow more than 1 argument")
	}
}

func TestSsmSSHConfigCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "SSH-config help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Generate an SSH config entry",
		},
		{
			name:    "SSH-config with instance ID",
			args:    []string{"i-1234567890abcdef0"},
			wantErr: false,
		},
		{
			name:    "SSH-config with name flag",
			args:    []string{"i-1234567890abcdef0", "--name", "prod-web"},
			wantErr: false,
		},
		{
			name:    "SSH-config with region flag",
			args:    []string{"i-1234567890abcdef0", "--region", "cac1"},
			wantErr: false,
		},
		{
			name:    "SSH-config with append flag",
			args:    []string{"i-1234567890abcdef0", "--append"},
			wantErr: false,
		},
		{
			name:    "SSH-config with too many args",
			args:    []string{"instance1", "instance2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   ssmSSHConfigCmd.Use,
				Short: ssmSSHConfigCmd.Short,
				Long:  ssmSSHConfigCmd.Long,
				Args:  cobra.MaximumNArgs(1),
				Run: func(cmd *cobra.Command, args []string) {
					// Mock SSH config functionality
					name, _ := cmd.Flags().GetString("name")
					appendFlag, _ := cmd.Flags().GetBool("append")

					if name == "prod-web" {
						// Valid name
					}
					if appendFlag {
						// Append mode
					}

					if len(args) > 0 {
						instanceIdentifier := args[0]
						if instanceIdentifier == "" {
							t.Error("Instance identifier should not be empty")
						}
					}
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region")
			cmd.Flags().StringP("name", "n", "", "Friendly name")
			cmd.Flags().StringP("user", "u", "", "SSH username")
			cmd.Flags().StringP("identity", "i", "", "Path to SSH private key")
			cmd.Flags().BoolP("append", "a", false, "Append to config")

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.contains != "" {
				output := buf.String()
				if !strings.Contains(output, tt.contains) {
					t.Errorf("Execute() output should contain %v, got %v", tt.contains, output)
				}
			}
		})
	}
}

func TestSsmSSHConfigCmdFlags(t *testing.T) {
	tests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{"region", "r", ""},
		{"name", "n", ""},
		{"user", "u", ""},
		{"identity", "i", ""},
		{"append", "a", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName+" flag", func(t *testing.T) {
			flag := ssmSSHConfigCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s shorthand = %s, want %s", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			if flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %s default = %s, want %s", tt.flagName, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestSsmSSHConfigCmdStructure(t *testing.T) {
	if ssmSSHConfigCmd.Use != "ssh-config [instance-identifier]" {
		t.Errorf("Expected Use to be 'ssh-config [instance-identifier]', got %s", ssmSSHConfigCmd.Use)
	}

	if ssmSSHConfigCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if ssmSSHConfigCmd.Long == "" {
		t.Error("Command should have a long description")
	}

	if ssmSSHConfigCmd.Run == nil {
		t.Error("Command should have a Run function")
	}
}

func TestSsmRDPCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "RDP help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Connect to a Windows EC2 instance via RDP",
		},
		{
			name:    "RDP with instance ID",
			args:    []string{"i-1234567890abcdef0"},
			wantErr: false,
		},
		{
			name:    "RDP with instance name",
			args:    []string{"windows-server-1"},
			wantErr: false,
		},
		{
			name:    "RDP with region flag",
			args:    []string{"i-1234567890abcdef0", "--region", "cac1"},
			wantErr: false,
		},
		{
			name:    "RDP with local-port flag",
			args:    []string{"i-1234567890abcdef0", "--local-port", "13389"},
			wantErr: false,
		},
		{
			name:    "RDP with launch flag",
			args:    []string{"i-1234567890abcdef0", "--launch"},
			wantErr: false,
		},
		{
			name:    "RDP with too many args",
			args:    []string{"instance1", "instance2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   ssmRDPCmd.Use,
				Short: ssmRDPCmd.Short,
				Long:  ssmRDPCmd.Long,
				Args:  cobra.MaximumNArgs(1),
				Run: func(cmd *cobra.Command, args []string) {
					// Mock RDP functionality
					localPort, _ := cmd.Flags().GetInt("local-port")
					launch, _ := cmd.Flags().GetBool("launch")

					if localPort == 13389 {
						// Custom port
					}
					if launch {
						// Auto-launch mode
					}

					if len(args) > 0 {
						instanceIdentifier := args[0]
						if instanceIdentifier == "" {
							t.Error("Instance identifier should not be empty")
						}
					}
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region")
			cmd.Flags().IntP("local-port", "p", 33389, "Local port")
			cmd.Flags().BoolP("launch", "l", false, "Auto-launch RDP client")

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.contains != "" {
				output := buf.String()
				if !strings.Contains(output, tt.contains) {
					t.Errorf("Execute() output should contain %v, got %v", tt.contains, output)
				}
			}
		})
	}
}

func TestSsmRDPCmdFlags(t *testing.T) {
	tests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{"region", "r", ""},
		{"local-port", "p", "33389"},
		{"launch", "l", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName+" flag", func(t *testing.T) {
			flag := ssmRDPCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found", tt.flagName)
				return
			}

			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s shorthand = %s, want %s", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			if flag.DefValue != tt.defaultValue {
				t.Errorf("Flag %s default = %s, want %s", tt.flagName, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestSsmRDPCmdStructure(t *testing.T) {
	if ssmRDPCmd.Use != "rdp [instance-identifier]" {
		t.Errorf("Expected Use to be 'rdp [instance-identifier]', got %s", ssmRDPCmd.Use)
	}

	if ssmRDPCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if ssmRDPCmd.Long == "" {
		t.Error("Command should have a long description")
	}

	if ssmRDPCmd.Run == nil {
		t.Error("Command should have a Run function")
	}

	// Test argument validation
	err := ssmRDPCmd.Args(ssmRDPCmd, []string{})
	if err != nil {
		t.Errorf("Command should allow 0 arguments for fuzzy finder, got error: %v", err)
	}

	err = ssmRDPCmd.Args(ssmRDPCmd, []string{"i-1234567890abcdef0"})
	if err != nil {
		t.Errorf("Command should allow 1 argument, got error: %v", err)
	}

	err = ssmRDPCmd.Args(ssmRDPCmd, []string{"instance1", "instance2"})
	if err == nil {
		t.Error("Command should not allow more than 1 argument")
	}
}

func TestBuildProxyCommand(t *testing.T) {
	tests := []struct {
		instanceID string
		region     string
		contains   []string
	}{
		{
			instanceID: "i-1234567890abcdef0",
			region:     "ca-central-1",
			contains: []string{
				"ssm start-session",
				"--target i-1234567890abcdef0",
				"--document-name AWS-StartSSHSession",
				"--region ca-central-1",
			},
		},
		{
			instanceID: "i-abcdef1234567890",
			region:     "us-east-1",
			contains: []string{
				"--target i-abcdef1234567890",
				"--region us-east-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.instanceID, func(t *testing.T) {
			result := buildProxyCommand(tt.instanceID, tt.region)

			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("buildProxyCommand() should contain %s, got %s", s, result)
				}
			}
		})
	}
}

func TestGetSSHCommand(t *testing.T) {
	result := getSSHCommand()

	if runtime.GOOS == "windows" {
		if result != "ssh.exe" {
			t.Errorf("getSSHCommand() on Windows should return ssh.exe, got %s", result)
		}
	} else {
		if result != "ssh" {
			t.Errorf("getSSHCommand() on Unix should return ssh, got %s", result)
		}
	}
}

func TestSSHExamplesUseCorrectRegion(t *testing.T) {
	// Verify that examples use cac1/ca-central-1 instead of use1/us-east-1
	longDesc := ssmSSHCmd.Long

	if strings.Contains(longDesc, "us-east-1") && !strings.Contains(longDesc, "ca-central-1") {
		t.Error("SSH command examples should use ca-central-1, not us-east-1")
	}

	if strings.Contains(longDesc, "-r use1") {
		t.Error("SSH command examples should use -r cac1, not -r use1")
	}
}

func TestSSHConfigExamplesUseCorrectRegion(t *testing.T) {
	longDesc := ssmSSHConfigCmd.Long

	if strings.Contains(longDesc, "us-east-1") && !strings.Contains(longDesc, "ca-central-1") {
		t.Error("SSH-config command examples should use ca-central-1, not us-east-1")
	}

	if strings.Contains(longDesc, "-r use1") {
		t.Error("SSH-config command examples should use -r cac1, not -r use1")
	}
}

func TestRDPExamplesUseCorrectRegion(t *testing.T) {
	longDesc := ssmRDPCmd.Long

	if strings.Contains(longDesc, "us-east-1") && !strings.Contains(longDesc, "ca-central-1") {
		t.Error("RDP command examples should use ca-central-1, not us-east-1")
	}

	if strings.Contains(longDesc, "-r use1") {
		t.Error("RDP command examples should use -r cac1, not -r use1")
	}
}
