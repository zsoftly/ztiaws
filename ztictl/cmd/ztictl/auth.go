package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"ztictl/internal/auth"
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
  ztictl auth login <profile>           # SSO login (profile required)
  ztictl auth logout [profile]          # SSO logout  
  ztictl auth profiles                  # List/manage profiles
  ztictl auth creds [profile]           # Show credentials`,
}

// authLoginCmd represents the auth login command
var authLoginCmd = &cobra.Command{
	Use:   "login [profile]",
	Short: "Login to AWS SSO",
	Long: `Login to AWS SSO with interactive account and role selection.
A profile name must be specified to ensure intentional credential management.

Note: AWS SSO authentication requires browser interaction and cannot be used in CI/CD pipelines.
For automated environments, use IAM-based authentication (OIDC, EC2 instance profiles, or IAM access keys).
See docs/CI_CD_AUTHENTICATION.md for details.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]

		// Check if running in non-interactive mode (CI/CD environment)
		execCtx := GetExecutionContext(cmd)
		if execCtx.NonInteractive || execCtx.IsCI {
			// Check if AWS credentials are already configured
			hasCredentials, credType := auth.DetectEnvironmentCredentials()

			if hasCredentials {
				logging.LogError("\nDetected CI/CD environment with %s\n", credType)
				fmt.Fprintf(os.Stderr, `
AWS SSO authentication requires browser interaction and cannot be used in CI/CD pipelines.

Your environment already has AWS credentials configured.
Use ztictl commands directly without 'auth login'.

For more information, see: docs/CI_CD_AUTHENTICATION.md
`)
				os.Exit(1)
			} else {
				logging.LogError("\nDetected CI/CD environment without AWS credentials\n")
				fmt.Fprintf(os.Stderr, `
AWS SSO authentication requires browser interaction and cannot be used in CI/CD pipelines.

To authenticate in CI/CD, you must configure IAM-based authentication:
  - GitHub Actions: Use OIDC federation
  - GitLab CI: Use OIDC federation
  - Jenkins: Use IAM access keys or EC2 instance profiles
  - Other platforms: See docs/CI_CD_AUTHENTICATION.md

Example GitHub Actions workflow:
  - name: Configure AWS Credentials
    uses: aws-actions/configure-aws-credentials@v2
    with:
      role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
      aws-region: ca-central-1

For complete examples and setup instructions:
  https://github.com/zsoftly/ztiaws/blob/main/docs/CI_CD_AUTHENTICATION.md
`)
				os.Exit(1)
			}
		}

		if err := performLogin(profileName); err != nil {
			logging.LogError("Login failed: %v", err)
			os.Exit(1)
		}
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

		if err := performLogout(profileName); err != nil {
			logging.LogError("Logout failed: %v", err)
			os.Exit(1)
		}
	},
}

// authProfilesCmd represents the auth profiles command
var authProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List and manage AWS profiles",
	Long:  `List all configured AWS profiles and their status.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := listAuthProfiles(); err != nil {
			logging.LogError("Failed to list profiles: %v", err)
			os.Exit(1)
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
		if err := showCredentials(args); err != nil {
			logging.LogError("Failed to show credentials: %v", err)
			logging.LogInfo("Usage: ztictl auth creds [profile-name]")
			os.Exit(1)
		}
	},
}

// performLogin handles the authentication login logic and returns errors instead of calling os.Exit
func performLogin(profileName string) error {
	authManager := auth.NewManager()
	ctx := context.Background()

	if err := authManager.Login(ctx, profileName); err != nil {
		return fmt.Errorf("authentication failed for profile %s: %w", profileName, err)
	}

	logging.LogSuccess("Authentication successful for profile: %s", profileName)
	return nil
}

// performLogout handles the authentication logout logic and returns errors instead of calling os.Exit
func performLogout(profileName string) error {
	authManager := auth.NewManager()
	ctx := context.Background()

	if err := authManager.Logout(ctx, profileName); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	if profileName != "" {
		logging.LogSuccess("Logout successful for profile: %s", profileName)
	} else {
		logging.LogSuccess("Logout successful for all profiles")
	}
	return nil
}

// listAuthProfiles handles the profile listing logic and returns errors instead of calling os.Exit
func listAuthProfiles() error {
	authManager := auth.NewManager()
	ctx := context.Background()

	profiles, err := authManager.ListProfiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		logging.LogInfo("No AWS profiles found")
		return nil
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
		_, _ = colors.Data.Printf("%-20s ", profile.Name) // #nosec G104
		fmt.Printf("%s\n", status)
		if profile.AccountID != "" {
			fmt.Printf("  Account: ")
			_, _ = colors.Data.Printf("%s ", profile.AccountID) // #nosec G104
			fmt.Printf("(")
			_, _ = colors.Data.Printf("%s", profile.AccountName) // #nosec G104
			fmt.Printf(")\n")
		}
		if profile.RoleName != "" {
			fmt.Printf("  Role: ")
			_, _ = colors.Data.Printf("%s\n", profile.RoleName) // #nosec G104
		}
		fmt.Println()
	}
	return nil
}

// showCredentials handles the credential display logic and returns errors instead of calling os.Exit
func showCredentials(args []string) error {
	var profileName string
	if len(args) > 0 {
		profileName = args[0]
	} else {
		// Try AWS_PROFILE environment variable
		profileName = os.Getenv("AWS_PROFILE")
		if profileName == "" {
			return fmt.Errorf("no profile specified. Usage: ztictl auth creds [profile-name]")
		}
	}

	authManager := auth.NewManager()
	ctx := context.Background()

	creds, err := authManager.GetCredentials(ctx, profileName)
	if err != nil {
		colors.PrintError("✗ Failed to get credentials for profile: %s\n", profileName)
		colors.PrintWarning("💡 Try authenticating with: ztictl auth login %s\n", profileName)
		return fmt.Errorf("failed to get credentials for profile %s: %w", profileName, err)
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
	return nil
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authProfilesCmd)
	authCmd.AddCommand(authCredsCmd)
}
