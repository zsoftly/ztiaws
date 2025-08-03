package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"ztictl/internal/auth"
	"ztictl/internal/config"
	"ztictl/pkg/colors"
	"ztictl/pkg/logging"

	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "AWS SSO authentication commands",
	Long: `Manage AWS SSO authentication including login, logout, profile management, and credential display.

Examples:
  ztictl auth login [profile]           # Interactive SSO login
  ztictl auth logout [profile]          # SSO logout  
  ztictl auth profiles                  # List/manage profiles
  ztictl auth creds [profile]           # Show credentials`,
}

// authLoginCmd represents the auth login command
var authLoginCmd = &cobra.Command{
	Use:   "login [profile]",
	Short: "Login to AWS SSO",
	Long: `Login to AWS SSO with interactive account and role selection.
It is recommended to always specify a profile name to avoid confusion.
If no profile is specified, you will be prompted to confirm using the default profile.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		var profileName string
		if len(args) > 0 {
			profileName = args[0]
		} else {
			profileName = cfg.SSO.DefaultProfile
			if profileName == "" {
				logging.LogError("No profile specified and no default profile configured")
				logging.LogWarn("Usage: ztictl auth login <profile-name>")
				os.Exit(1)
			}

			// Prompt user to confirm using default profile (like bash version)
			logging.LogWarn("No profile specified. Using default: %s", profileName)
			colors.PrintData("Proceed with default profile? (y/n): ")

			var response string
			fmt.Scanln(&response)

			if strings.ToLower(strings.TrimSpace(response)) != "y" && strings.ToLower(strings.TrimSpace(response)) != "yes" {
				logging.LogWarn("Please run: ztictl auth login <profile-name>")
				os.Exit(0)
			}
		}

		logging.LogInfo("Starting AWS SSO authentication for profile: %s", profileName)

		authManager := auth.NewManager(logger)
		ctx := context.Background()

		if err := authManager.Login(ctx, profileName); err != nil {
			logging.LogError("Authentication failed for profile %s: %v", profileName, err)
			os.Exit(1)
		}

		logging.LogSuccess("Authentication successful for profile: %s", profileName)
	},
}

// authLogoutCmd represents the auth logout command
var authLogoutCmd = &cobra.Command{
	Use:   "logout [profile]",
	Short: "Logout from AWS SSO",
	Long:  `Logout from AWS SSO and clear cached credentials.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var profileName string
		if len(args) > 0 {
			profileName = args[0]
		}

		authManager := auth.NewManager(logger)
		ctx := context.Background()

		if err := authManager.Logout(ctx, profileName); err != nil {
			logging.LogError("Logout failed: %v", err)
			os.Exit(1)
		}

		if profileName != "" {
			logging.LogSuccess("Logout successful for profile: %s", profileName)
		} else {
			logging.LogSuccess("Logout successful for all profiles")
		}
	},
}

// authProfilesCmd represents the auth profiles command
var authProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List and manage AWS profiles",
	Long:  `List all configured AWS profiles and their status.`,
	Run: func(cmd *cobra.Command, args []string) {
		authManager := auth.NewManager(logger)
		ctx := context.Background()

		profiles, err := authManager.ListProfiles(ctx)
		if err != nil {
			logger.Error("Failed to list profiles", "error", err)
			os.Exit(1)
		}

		if len(profiles) == 0 {
			logger.Info("No AWS profiles found")
			return
		}

		fmt.Printf("\n")
		colors.PrintHeader("AWS Profiles:\n")
		colors.PrintHeader("=============\n")
		for _, profile := range profiles {
			var status string
			if profile.IsAuthenticated {
				status = colors.ColorSuccess("✅ Authenticated")
			} else {
				status = colors.ColorError("❌ Not authenticated")
			}
			colors.Data.Printf("%-20s ", profile.Name)
			fmt.Printf("%s\n", status)
			if profile.AccountID != "" {
				fmt.Printf("  Account: ")
				colors.Data.Printf("%s ", profile.AccountID)
				fmt.Printf("(")
				colors.Data.Printf("%s", profile.AccountName)
				fmt.Printf(")\n")
			}
			if profile.RoleName != "" {
				fmt.Printf("  Role: ")
				colors.Data.Printf("%s\n", profile.RoleName)
			}
			fmt.Println()
		}
	},
}

// authCredsCmd represents the auth creds command
var authCredsCmd = &cobra.Command{
	Use:   "creds [profile]",
	Short: "Show AWS credentials",
	Long: `Display AWS credentials for the specified profile in environment variable format.
If no profile is specified, uses the current AWS_PROFILE or default profile.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		var profileName string
		if len(args) > 0 {
			profileName = args[0]
		} else {
			// Try AWS_PROFILE environment variable first
			profileName = os.Getenv("AWS_PROFILE")
			if profileName == "" {
				profileName = cfg.SSO.DefaultProfile
			}
			if profileName == "" {
				logger.Error("No profile specified and no default profile found")
				logger.Info("Usage: ztictl auth creds [profile-name]")
				os.Exit(1)
			}
		}

		authManager := auth.NewManager(logger)
		ctx := context.Background()

		creds, err := authManager.GetCredentials(ctx, profileName)
		if err != nil {
			colors.PrintError("✗ Failed to get credentials for profile: %s\n", profileName)
			logger.Error("Failed to get credentials", "profile", profileName, "error", err)
			colors.PrintWarning("💡 Try authenticating with: ztictl auth login %s\n", profileName)
			os.Exit(1)
		}

		fmt.Printf("\n")
		colors.PrintHeader("🔑 AWS Credentials for profile: %s\n", profileName)
		colors.PrintHeader("----------------------------------------\n")

		// Platform-specific credential output
		switch runtime.GOOS {
		case "windows":
			// Windows Command Prompt instructions
			colors.PrintHeader("\nFor Command Prompt (cmd):\n")
			colors.PrintData("set AWS_ACCESS_KEY_ID=%s\n", creds.AccessKeyID)
			colors.PrintData("set AWS_SECRET_ACCESS_KEY=%s\n", creds.SecretAccessKey)
			if creds.SessionToken != "" {
				colors.PrintData("set AWS_SESSION_TOKEN=%s\n", creds.SessionToken)
			}
			colors.PrintData("set AWS_REGION=%s\n", creds.Region)

			colors.PrintHeader("\nFor PowerShell:\n")
			colors.PrintData("$env:AWS_ACCESS_KEY_ID=\"%s\"\n", creds.AccessKeyID)
			colors.PrintData("$env:AWS_SECRET_ACCESS_KEY=\"%s\"\n", creds.SecretAccessKey)
			if creds.SessionToken != "" {
				colors.PrintData("$env:AWS_SESSION_TOKEN=\"%s\"\n", creds.SessionToken)
			}
			colors.PrintData("$env:AWS_REGION=\"%s\"\n", creds.Region)

		default:
			// Unix/Linux/macOS instructions
			colors.PrintData("export AWS_ACCESS_KEY_ID=%s\n", creds.AccessKeyID)
			colors.PrintData("export AWS_SECRET_ACCESS_KEY=%s\n", creds.SecretAccessKey)
			if creds.SessionToken != "" {
				colors.PrintData("export AWS_SESSION_TOKEN=%s\n", creds.SessionToken)
			}
			colors.PrintData("export AWS_REGION=%s\n", creds.Region)
			colors.PrintHeader("----------------------------------------\n")
			fmt.Printf("To use these credentials in your current shell, run:\n")
			colors.PrintSuccess("eval $(ztictl auth creds %s)\n", profileName)
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authProfilesCmd)
	authCmd.AddCommand(authCredsCmd)
}
