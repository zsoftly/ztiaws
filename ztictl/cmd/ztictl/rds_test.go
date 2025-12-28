package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRdsCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "RDS help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Manage AWS RDS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command to avoid state pollution
			cmd := &cobra.Command{
				Use:   rdsCmd.Use,
				Short: rdsCmd.Short,
				Long:  rdsCmd.Long,
			}

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

func TestRdsCmdStructure(t *testing.T) {
	if rdsCmd.Use != "rds" {
		t.Errorf("Expected Use to be 'rds', got %s", rdsCmd.Use)
	}

	if rdsCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if rdsCmd.Long == "" {
		t.Error("Command should have a long description")
	}

	// Verify subcommands are registered
	subcommands := rdsCmd.Commands()
	expectedSubcommands := []string{"list", "start", "stop", "reboot"}

	for _, expected := range expectedSubcommands {
		found := false
		for _, cmd := range subcommands {
			if cmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %s not found", expected)
		}
	}
}

func TestRdsListCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "List help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "List all RDS database instances",
		},
		{
			name:    "List with region flag",
			args:    []string{"--region", "cac1"},
			wantErr: false,
		},
		{
			name:    "List with short region flag",
			args:    []string{"-r", "ca-central-1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   rdsListCmd.Use,
				Short: rdsListCmd.Short,
				Long:  rdsListCmd.Long,
				Run: func(cmd *cobra.Command, args []string) {
					// Mock list functionality
					regionCode, _ := cmd.Flags().GetString("region")
					if regionCode == "cac1" {
						// Valid shortcode
					}
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region")

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

func TestRdsListCmdFlags(t *testing.T) {
	tests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{"region", "r", ""},
	}

	for _, tt := range tests {
		t.Run(tt.flagName+" flag", func(t *testing.T) {
			flag := rdsListCmd.Flags().Lookup(tt.flagName)
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

func TestRdsListCmdStructure(t *testing.T) {
	if rdsListCmd.Use != "list" {
		t.Errorf("Expected Use to be 'list', got %s", rdsListCmd.Use)
	}

	if rdsListCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if rdsListCmd.Long == "" {
		t.Error("Command should have a long description")
	}

	if rdsListCmd.Run == nil {
		t.Error("Command should have a Run function")
	}
}

func TestRdsStartCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Start help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Start a stopped RDS database instance",
		},
		{
			name:    "Start with db identifier",
			args:    []string{"my-database"},
			wantErr: false,
		},
		{
			name:    "Start with region flag",
			args:    []string{"my-database", "--region", "cac1"},
			wantErr: false,
		},
		{
			name:    "Start with wait flag",
			args:    []string{"my-database", "--wait"},
			wantErr: false,
		},
		{
			name:    "Start without db identifier",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Start with too many args",
			args:    []string{"db1", "db2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   rdsStartCmd.Use,
				Short: rdsStartCmd.Short,
				Long:  rdsStartCmd.Long,
				Args:  cobra.ExactArgs(1),
				Run: func(cmd *cobra.Command, args []string) {
					// Mock start functionality
					wait, _ := cmd.Flags().GetBool("wait")
					if wait {
						// Wait mode
					}
					if len(args) > 0 {
						dbIdentifier := args[0]
						if dbIdentifier == "" {
							t.Error("DB identifier should not be empty")
						}
					}
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region")
			cmd.Flags().BoolP("wait", "w", false, "Wait for completion")

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

func TestRdsStartCmdFlags(t *testing.T) {
	tests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{"region", "r", ""},
		{"wait", "w", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName+" flag", func(t *testing.T) {
			flag := rdsStartCmd.Flags().Lookup(tt.flagName)
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

func TestRdsStartCmdStructure(t *testing.T) {
	if rdsStartCmd.Use != "start <db-identifier>" {
		t.Errorf("Expected Use to be 'start <db-identifier>', got %s", rdsStartCmd.Use)
	}

	if rdsStartCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if rdsStartCmd.Long == "" {
		t.Error("Command should have a long description")
	}

	if rdsStartCmd.Run == nil {
		t.Error("Command should have a Run function")
	}

	// Test argument validation - requires exactly 1 argument
	err := rdsStartCmd.Args(rdsStartCmd, []string{})
	if err == nil {
		t.Error("Command should require exactly 1 argument")
	}

	err = rdsStartCmd.Args(rdsStartCmd, []string{"my-database"})
	if err != nil {
		t.Errorf("Command should allow 1 argument, got error: %v", err)
	}

	err = rdsStartCmd.Args(rdsStartCmd, []string{"db1", "db2"})
	if err == nil {
		t.Error("Command should not allow more than 1 argument")
	}
}

func TestRdsStopCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Stop help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Stop a running RDS database instance",
		},
		{
			name:    "Stop with db identifier",
			args:    []string{"my-database"},
			wantErr: false,
		},
		{
			name:    "Stop with region flag",
			args:    []string{"my-database", "--region", "cac1"},
			wantErr: false,
		},
		{
			name:    "Stop with wait flag",
			args:    []string{"my-database", "--wait"},
			wantErr: false,
		},
		{
			name:    "Stop without db identifier",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   rdsStopCmd.Use,
				Short: rdsStopCmd.Short,
				Long:  rdsStopCmd.Long,
				Args:  cobra.ExactArgs(1),
				Run: func(cmd *cobra.Command, args []string) {
					// Mock stop functionality
					wait, _ := cmd.Flags().GetBool("wait")
					if wait {
						// Wait mode
					}
					if len(args) > 0 {
						dbIdentifier := args[0]
						if dbIdentifier == "" {
							t.Error("DB identifier should not be empty")
						}
					}
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region")
			cmd.Flags().BoolP("wait", "w", false, "Wait for completion")

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

func TestRdsStopCmdFlags(t *testing.T) {
	tests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{"region", "r", ""},
		{"wait", "w", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName+" flag", func(t *testing.T) {
			flag := rdsStopCmd.Flags().Lookup(tt.flagName)
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

func TestRdsStopCmdStructure(t *testing.T) {
	if rdsStopCmd.Use != "stop <db-identifier>" {
		t.Errorf("Expected Use to be 'stop <db-identifier>', got %s", rdsStopCmd.Use)
	}

	if rdsStopCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if rdsStopCmd.Long == "" {
		t.Error("Command should have a long description")
	}

	if rdsStopCmd.Run == nil {
		t.Error("Command should have a Run function")
	}
}

func TestRdsRebootCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains string
	}{
		{
			name:     "Reboot help",
			args:     []string{"--help"},
			wantErr:  false,
			contains: "Reboot an RDS database instance",
		},
		{
			name:    "Reboot with db identifier",
			args:    []string{"my-database"},
			wantErr: false,
		},
		{
			name:    "Reboot with region flag",
			args:    []string{"my-database", "--region", "cac1"},
			wantErr: false,
		},
		{
			name:    "Reboot with wait flag",
			args:    []string{"my-database", "--wait"},
			wantErr: false,
		},
		{
			name:    "Reboot with force-failover flag",
			args:    []string{"my-database", "--force-failover"},
			wantErr: false,
		},
		{
			name:    "Reboot without db identifier",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   rdsRebootCmd.Use,
				Short: rdsRebootCmd.Short,
				Long:  rdsRebootCmd.Long,
				Args:  cobra.ExactArgs(1),
				Run: func(cmd *cobra.Command, args []string) {
					// Mock reboot functionality
					wait, _ := cmd.Flags().GetBool("wait")
					forceFailover, _ := cmd.Flags().GetBool("force-failover")
					if wait {
						// Wait mode
					}
					if forceFailover {
						// Force failover mode
					}
					if len(args) > 0 {
						dbIdentifier := args[0]
						if dbIdentifier == "" {
							t.Error("DB identifier should not be empty")
						}
					}
				},
			}

			cmd.Flags().StringP("region", "r", "", "AWS region")
			cmd.Flags().BoolP("wait", "w", false, "Wait for completion")
			cmd.Flags().BoolP("force-failover", "f", false, "Force failover")

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

func TestRdsRebootCmdFlags(t *testing.T) {
	tests := []struct {
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{"region", "r", ""},
		{"wait", "w", "false"},
		{"force-failover", "f", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName+" flag", func(t *testing.T) {
			flag := rdsRebootCmd.Flags().Lookup(tt.flagName)
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

func TestRdsRebootCmdStructure(t *testing.T) {
	if rdsRebootCmd.Use != "reboot <db-identifier>" {
		t.Errorf("Expected Use to be 'reboot <db-identifier>', got %s", rdsRebootCmd.Use)
	}

	if rdsRebootCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if rdsRebootCmd.Long == "" {
		t.Error("Command should have a long description")
	}

	if rdsRebootCmd.Run == nil {
		t.Error("Command should have a Run function")
	}
}

func TestColorizeRDSStatus(t *testing.T) {
	tests := []struct {
		status   string
		notEmpty bool
	}{
		{"available", true},
		{"stopped", true},
		{"starting", true},
		{"stopping", true},
		{"rebooting", true},
		{"modifying", true},
		{"backing-up", true},
		{"failed", true},
		{"incompatible-restore", true},
		{"incompatible-network", true},
		{"unknown-status", true},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := colorizeRDSStatus(tt.status)
			if tt.notEmpty && result == "" {
				t.Errorf("colorizeRDSStatus(%s) should not be empty", tt.status)
			}
			// The result should contain the original status text
			if !strings.Contains(result, tt.status) {
				t.Errorf("colorizeRDSStatus(%s) should contain the status text, got %s", tt.status, result)
			}
		})
	}
}

func TestRdsExamplesUseCorrectRegion(t *testing.T) {
	// Verify that examples use cac1/ca-central-1 instead of use1/us-east-1
	commands := []*cobra.Command{rdsListCmd, rdsStartCmd, rdsStopCmd, rdsRebootCmd}

	for _, cmd := range commands {
		t.Run(cmd.Name(), func(t *testing.T) {
			longDesc := cmd.Long

			if strings.Contains(longDesc, "us-east-1") && !strings.Contains(longDesc, "ca-central-1") {
				t.Errorf("RDS %s command examples should use ca-central-1, not us-east-1", cmd.Name())
			}

			if strings.Contains(longDesc, "-r use1") {
				t.Errorf("RDS %s command examples should use -r cac1, not -r use1", cmd.Name())
			}
		})
	}
}

func TestRdsArgumentValidation(t *testing.T) {
	// Test that start, stop, reboot require exactly 1 argument
	commands := []*cobra.Command{rdsStartCmd, rdsStopCmd, rdsRebootCmd}

	for _, cmd := range commands {
		t.Run(cmd.Name()+" requires 1 arg", func(t *testing.T) {
			err := cmd.Args(cmd, []string{})
			if err == nil {
				t.Errorf("RDS %s should require exactly 1 argument", cmd.Name())
			}

			err = cmd.Args(cmd, []string{"my-database"})
			if err != nil {
				t.Errorf("RDS %s should accept 1 argument, got error: %v", cmd.Name(), err)
			}

			err = cmd.Args(cmd, []string{"db1", "db2"})
			if err == nil {
				t.Errorf("RDS %s should not accept more than 1 argument", cmd.Name())
			}
		})
	}
}
