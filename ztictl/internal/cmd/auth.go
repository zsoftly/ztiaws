package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ztictl/internal/auth"
	"ztictl/internal/config"
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
If no profile is specified, uses the default profile from configuration.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()

		var profileName string
		if len(args) > 0 {
			profileName = args[0]
		} else {
			profileName = cfg.SSO.DefaultProfile
			if profileName == "" {
				logger.Error("No profile specified and no default profile configured")
				logger.Info("Usage: ztictl auth login <profile-name>")
				os.Exit(1)
			}
		}

		logger.Info("Starting AWS SSO authentication", "profile", profileName)

		authManager := auth.NewManager(logger)
		ctx := context.Background()

		if err := authManager.Login(ctx, profileName); err != nil {
			logger.Error("Authentication failed", "error", err)
			os.Exit(1)
		}

		logger.Info("Authentication successful", "profile", profileName)
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
			logger.Error("Logout failed", "error", err)
			os.Exit(1)
		}

		if profileName != "" {
			logger.Info("Logout successful", "profile", profileName)
		} else {
			logger.Info("Logout successful for all profiles")
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

		fmt.Println("\nAWS Profiles:")
		fmt.Println("=============")
		for _, profile := range profiles {
			status := "❌ Not authenticated"
			if profile.IsAuthenticated {
				status = "✅ Authenticated"
			}
			fmt.Printf("%-20s %s\n", profile.Name, status)
			if profile.AccountID != "" {
				fmt.Printf("  Account: %s (%s)\n", profile.AccountID, profile.AccountName)
			}
			if profile.RoleName != "" {
				fmt.Printf("  Role: %s\n", profile.RoleName)
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
			logger.Error("Failed to get credentials", "profile", profileName, "error", err)
			logger.Info("Try authenticating with: ztictl auth login %s", profileName)
			os.Exit(1)
		}

		fmt.Printf("\n🔑 AWS Credentials for profile: %s\n", profileName)
		fmt.Println("----------------------------------------")
		fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", creds.AccessKeyID)
		fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", creds.SecretAccessKey)
		if creds.SessionToken != "" {
			fmt.Printf("export AWS_SESSION_TOKEN=%s\n", creds.SessionToken)
		}
		fmt.Printf("export AWS_REGION=%s\n", creds.Region)
		fmt.Println("----------------------------------------")
		fmt.Printf("To use these credentials in your current shell, run:\n")
		fmt.Printf("eval $(ztictl auth creds %s)\n", profileName)
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authProfilesCmd)
	authCmd.AddCommand(authCredsCmd)
}
