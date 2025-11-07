package auth

import (
	"context"
	"crypto/sha1" // #nosec G505 -- SHA1 required for AWS CLI compatibility, not used for security
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	appconfig "ztictl/internal/config"
	"ztictl/pkg/errors"
	"ztictl/pkg/logging"
	"ztictl/pkg/security"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/browser"
	"golang.org/x/term"
)

const (
	// SSO authentication timeout constraints
	MinTimeoutSeconds = 60  // 1 minute minimum for user interaction
	MaxTimeoutSeconds = 180 // 3 minute maximum for security

	// Column formatting constants
	MinColumnWidth = 12
	MaxColumnWidth = 40
	ColumnPadding  = 2
)

// getTerminalWidth returns the width of the terminal or a default if it fails
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Fallback to a wider default for test environments
		return 120
	}
	return width
}

// Manager handles AWS SSO authentication operations
type Manager struct {
	logger *logging.Logger
}

// Profile represents an AWS profile with SSO information
type Profile struct {
	Name            string     `json:"name"`
	IsAuthenticated bool       `json:"is_authenticated"`
	AccountID       string     `json:"account_id,omitempty"`
	AccountName     string     `json:"account_name,omitempty"`
	RoleName        string     `json:"role_name,omitempty"`
	Region          string     `json:"region,omitempty"`
	SSOStartURL     string     `json:"sso_start_url,omitempty"`
	SSORegion       string     `json:"sso_region,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}

// Credentials represents AWS credentials
type Credentials struct {
	AccessKeyID     string     `json:"access_key_id"`
	SecretAccessKey string     `json:"secret_access_key"`
	SessionToken    string     `json:"session_token"`
	Region          string     `json:"region"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}

// SSOToken represents an SSO access token from the cache
type SSOToken struct {
	StartURL    string    `json:"startUrl"`
	Region      string    `json:"region"`
	AccessToken string    `json:"accessToken"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// Account represents an AWS account from SSO
type Account struct {
	AccountID    string `json:"account_id"`
	AccountName  string `json:"account_name"`
	EmailAddress string `json:"email_address,omitempty"`
}

// Role represents an AWS role in an account
type Role struct {
	RoleName  string `json:"role_name"`
	AccountID string `json:"account_id"`
}

// CredentialType represents the type of AWS credentials detected
type CredentialType string

const (
	// CredentialTypeNone indicates no credentials detected
	CredentialTypeNone CredentialType = "none"
	// CredentialTypeAccessKeys indicates IAM access keys in environment
	CredentialTypeAccessKeys CredentialType = "IAM access keys"
	// CredentialTypeEC2Instance indicates EC2 instance profile
	CredentialTypeEC2Instance CredentialType = "EC2 instance profile" // #nosec G101
	// CredentialTypeECSTask indicates ECS task role
	CredentialTypeECSTask CredentialType = "ECS task role" // #nosec G101
	// CredentialTypeOIDC indicates OIDC assumed role
	CredentialTypeOIDC CredentialType = "OIDC assumed role"
)

// DetectEnvironmentCredentials checks for AWS credentials in the environment
// Returns true if credentials are found, along with the credential type
func DetectEnvironmentCredentials() (bool, CredentialType) {
	// Check for IAM access keys in environment variables
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		return true, CredentialTypeAccessKeys
	}

	// Check for ECS task role
	// ECS provides credentials via a relative URI
	if os.Getenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI") != "" {
		return true, CredentialTypeECSTask
	}

	// Check for OIDC assumed role
	// When using OIDC (GitHub Actions, GitLab CI, etc.), AWS_ROLE_ARN is set
	if os.Getenv("AWS_ROLE_ARN") != "" {
		return true, CredentialTypeOIDC
	}

	// Check for EC2 instance metadata (IMDS)
	if isEC2Instance() {
		return true, CredentialTypeEC2Instance
	}

	return false, CredentialTypeNone
}

// isEC2Instance checks if running on an EC2 instance by querying IMDS
// Uses a short timeout to avoid blocking if not on EC2
func isEC2Instance() bool {
	// Create HTTP client with very short timeout
	client := &http.Client{
		Timeout: 100 * time.Millisecond,
	}

	// Try to reach EC2 instance metadata service (IMDSv2 token endpoint)
	// Using the token endpoint because it's available on both IMDSv1 and IMDSv2
	req, err := http.NewRequest("PUT", "http://169.254.169.254/latest/api/token", nil)
	if err != nil {
		return false
	}

	// IMDSv2 requires this header
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "1")

	resp, err := client.Do(req)
	if err != nil {
		// Also try IMDSv1 as fallback (simple GET request)
		req, err = http.NewRequest("GET", "http://169.254.169.254/latest/meta-data/", nil)
		if err != nil {
			return false
		}

		resp, err = client.Do(req)
		if err != nil {
			return false
		}
	}

	defer resp.Body.Close()

	// If we get a 200 or 404, we're on EC2
	// 404 can happen on IMDSv1 with certain endpoints
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound
}

// NewManager creates a new authentication manager with a no-op logger
func NewManager() *Manager {
	return &Manager{
		logger: logging.NewNoOpLogger(),
	}
}

// NewManagerWithLogger creates a new authentication manager with a logger
func NewManagerWithLogger(logger *logging.Logger) *Manager {
	if logger == nil {
		logger = logging.NewNoOpLogger()
	}
	return &Manager{
		logger: logger,
	}
}

// getDisplayItemCount returns the number of items to display in the fuzzy finder
// Reads from ZTICTL_SELECTOR_HEIGHT environment variable, defaults to 5 items
//
// Environment variable: ZTICTL_SELECTOR_HEIGHT
// Default: 5 (matching fzf --height=20% behavior)
// Range: 1-20 items
//
// Examples:
//
//	export ZTICTL_SELECTOR_HEIGHT=3   # Show 3 items
//	export ZTICTL_SELECTOR_HEIGHT=10  # Show 10 items
func getDisplayItemCount() int {
	heightStr := os.Getenv("ZTICTL_SELECTOR_HEIGHT")
	if heightStr == "" {
		return 5 // Default to 5 items (matching fzf --height=20%)
	}

	height, err := strconv.Atoi(heightStr)
	if err != nil || height < 1 {
		logging.LogWarn("Invalid ZTICTL_SELECTOR_HEIGHT value '%s', using default of 5", heightStr)
		return 5
	}

	// Limit to reasonable range (1-20 items)
	if height > 20 {
		logging.LogWarn("ZTICTL_SELECTOR_HEIGHT of %d is too large, limiting to 20", height)
		return 20
	}

	return height
}

// Helper functions for dynamic column formatting

// wrapText wraps text to fit within a specified width
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	var currentLine strings.Builder

	for _, word := range words {
		// If adding this word would exceed width, start a new line
		if currentLine.Len() > 0 && currentLine.Len()+1+len(word) > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// calculateOptimalWidths calculates optimal column widths based on content
func calculateOptimalWidths(accounts []Account) (int, int, int) {
	maxIDWidth := MinColumnWidth
	maxNameWidth := MinColumnWidth
	maxEmailWidth := MinColumnWidth

	// Find maximum widths
	for _, account := range accounts {
		if len(account.AccountID) > maxIDWidth {
			maxIDWidth = len(account.AccountID)
		}
		if len(account.AccountName) > maxNameWidth {
			maxNameWidth = len(account.AccountName)
		}
		if len(account.EmailAddress) > maxEmailWidth {
			maxEmailWidth = len(account.EmailAddress)
		}
	}

	// Apply constraints and padding
	if maxIDWidth > MaxColumnWidth {
		maxIDWidth = MaxColumnWidth
	}
	if maxNameWidth > MaxColumnWidth {
		maxNameWidth = MaxColumnWidth
	}
	if maxEmailWidth > MaxColumnWidth {
		maxEmailWidth = MaxColumnWidth
	}

	return maxIDWidth + ColumnPadding, maxNameWidth + ColumnPadding, maxEmailWidth + ColumnPadding
}

// formatAccountRow formats an account into a multi-line display with equal column spacing
func formatAccountRow(account Account, idWidth, nameWidth, emailWidth int) string {
	// Wrap text for each column
	idLines := wrapText(account.AccountID, idWidth-ColumnPadding)
	nameLines := wrapText(account.AccountName, nameWidth-ColumnPadding)
	emailLines := wrapText(account.EmailAddress, emailWidth-ColumnPadding)

	// Find the maximum number of lines needed
	maxLines := len(idLines)
	if len(nameLines) > maxLines {
		maxLines = len(nameLines)
	}
	if len(emailLines) > maxLines {
		maxLines = len(emailLines)
	}

	// Pad all slices to the same length
	for len(idLines) < maxLines {
		idLines = append(idLines, "")
	}
	for len(nameLines) < maxLines {
		nameLines = append(nameLines, "")
	}
	for len(emailLines) < maxLines {
		emailLines = append(emailLines, "")
	}

	// Build the formatted string
	var result strings.Builder
	for i := 0; i < maxLines; i++ {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(fmt.Sprintf("%-*s%-*s%s",
			idWidth, idLines[i],
			nameWidth, nameLines[i],
			emailLines[i]))
	}

	return result.String()
}

// calculateOptimalRoleWidths calculates optimal column widths for roles
func calculateOptimalRoleWidths(roles []Role, account *Account) (int, int) {
	maxRoleWidth := MinColumnWidth
	accountInfo := fmt.Sprintf("Account: %s (%s)", account.AccountName, account.AccountID)
	accountInfoWidth := len(accountInfo)

	// Find maximum role width
	for _, role := range roles {
		if len(role.RoleName) > maxRoleWidth {
			maxRoleWidth = len(role.RoleName)
		}
	}

	// Apply constraints and padding
	if maxRoleWidth > MaxColumnWidth {
		maxRoleWidth = MaxColumnWidth
	}

	return maxRoleWidth + ColumnPadding, accountInfoWidth + ColumnPadding
}

// formatRoleRow formats a role into a multi-line display with equal column spacing
func formatRoleRow(role Role, account *Account, roleWidth, accountWidth int) string {
	accountInfo := fmt.Sprintf("Account: %s (%s)", account.AccountName, account.AccountID)

	// Wrap text for each column
	roleLines := wrapText(role.RoleName, roleWidth-ColumnPadding)
	accountLines := wrapText(accountInfo, accountWidth-ColumnPadding)

	// Find the maximum number of lines needed
	maxLines := len(roleLines)
	if len(accountLines) > maxLines {
		maxLines = len(accountLines)
	}

	// Pad all slices to the same length
	for len(roleLines) < maxLines {
		roleLines = append(roleLines, "")
	}
	for len(accountLines) < maxLines {
		accountLines = append(accountLines, "")
	}

	// Build the formatted string
	var result strings.Builder
	for i := 0; i < maxLines; i++ {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(fmt.Sprintf("%-*s%s",
			roleWidth, roleLines[i],
			accountLines[i]))
	}

	return result.String()
}

// getAWSConfigDir returns the AWS configuration directory path
func getAWSConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".aws")

	// Validate path to prevent directory traversal
	if err := security.ValidateFilePath(configDir, homeDir); err != nil {
		return "", fmt.Errorf("invalid AWS config directory path: %w", err)
	}

	return configDir, nil
}

// getAWSCacheDir returns the AWS SSO cache directory path
func getAWSCacheDir() (string, error) {
	configDir, err := getAWSConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "sso", "cache"), nil
}

// Login performs AWS SSO login with interactive account and role selection
func (m *Manager) Login(ctx context.Context, profileName string) error {
	cfg := appconfig.Get()

	// Log the SSO configuration for debugging
	logging.LogDebug("SSO Configuration | start_url=%s region=%s profile=%s", cfg.SSO.StartURL, cfg.SSO.Region, profileName)

	if cfg.SSO.StartURL == "" {
		return errors.NewValidationError("SSO start URL not configured. Please run 'ztictl config init --interactive' to set up your AWS SSO settings, or edit ~/.ztictl.yaml manually")
	}

	logging.LogInfo("Starting AWS SSO authentication | profile=%s start_url=%s", profileName, cfg.SSO.StartURL)

	// Step 1: Configure the profile with basic SSO settings first (like bash version)
	if err := m.configureProfile(profileName, cfg); err != nil {
		return fmt.Errorf("failed to configure profile: %w", err)
	}

	// Step 2: Load AWS config without specifying the profile (to avoid SSO validation issues)
	// Create a completely isolated AWS config that bypasses all profile loading
	awsCfg := aws.Config{
		Region:      cfg.SSO.Region,
		Credentials: aws.AnonymousCredentials{},
	}
	ssoClient := sso.NewFromConfig(awsCfg)

	// Step 3: Check for valid cached token
	token, err := m.getCachedToken(cfg.SSO.StartURL)
	if err != nil || !m.isTokenValid(token) {
		logging.LogInfo("No valid cached token found, initiating SSO login...")

		// Perform SSO login
		if err := m.performSSOLogin(ctx, awsCfg, cfg); err != nil {
			return err
		}

		// Get the new token
		token, err = m.getCachedToken(cfg.SSO.StartURL)
		if err != nil {
			return fmt.Errorf("failed to get token after login: %w", err)
		}
	} else {
		logging.LogInfo("Using valid cached SSO token")
	}

	// Step 4: Get available accounts
	accounts, err := m.listAccounts(ctx, ssoClient, token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	// Step 5: Interactive account selection
	selectedAccount, err := m.selectAccount(accounts)
	if err != nil {
		return fmt.Errorf("account selection failed: %w", err)
	}

	logging.LogInfo("Selected account | id=%s name=%s", selectedAccount.AccountID, selectedAccount.AccountName)

	// Step 6: Get available roles for the selected account
	roles, err := m.listAccountRoles(ctx, ssoClient, token.AccessToken, selectedAccount.AccountID)
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	// Step 7: Interactive role selection
	selectedRole, err := m.selectRole(roles, selectedAccount)
	if err != nil {
		return fmt.Errorf("role selection failed: %w", err)
	}

	logging.LogInfo("Selected role | role=%s", selectedRole.RoleName)

	// Step 8: Update profile with selected account and role
	if err := m.updateProfileWithSelection(profileName, selectedAccount, selectedRole, cfg); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	logging.LogInfo("AWS SSO authentication completed successfully | account=%s role=%s profile=%s", selectedAccount.AccountName, selectedRole.RoleName, profileName)

	// Print platform-specific success message
	m.printSuccessMessage(profileName, selectedAccount, selectedRole, cfg)

	return nil
}

// Logout performs AWS SSO logout
func (m *Manager) Logout(ctx context.Context, profileName string) error {
	if profileName != "" {
		logging.LogInfo("Logging out from specific profile | profile=%s", profileName)
		// For now, we'll clear the specific profile's cached credentials
		// In a full implementation, this would clear SSO cache for the specific profile
	} else {
		logging.LogInfo("Logging out from all SSO sessions")
		// Clear all SSO cache
		if err := m.clearSSOCache(); err != nil {
			return fmt.Errorf("failed to clear SSO cache: %w", err)
		}
	}

	return nil
}

// ListProfiles returns all configured AWS profiles
func (m *Manager) ListProfiles(ctx context.Context) ([]Profile, error) {
	// Read AWS config file to get all profiles
	configDir, err := getAWSConfigDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(configDir, "config")

	// Validate config path to prevent directory traversal
	if err := security.ValidateFilePath(configPath, configDir); err != nil {
		return nil, fmt.Errorf("invalid config file path: %w", err)
	}

	content, err := os.ReadFile(configPath) // #nosec G304
	if err != nil {
		if os.IsNotExist(err) {
			return []Profile{}, nil
		}
		return nil, fmt.Errorf("failed to read AWS config: %w", err)
	}

	profiles := m.parseProfiles(string(content))

	// Check authentication status for each profile
	for i := range profiles {
		isAuth, err := m.isProfileAuthenticated(ctx, profiles[i].Name)
		if err != nil {
			logging.LogWarn("Failed to check authentication status | profile=%s error=%v", profiles[i].Name, err)
		}
		profiles[i].IsAuthenticated = isAuth
	}

	return profiles, nil
}

// GetCredentials returns AWS credentials for a profile
func (m *Manager) GetCredentials(ctx context.Context, profileName string) (*Credentials, error) {
	// First, try to get an STS token to force credential resolution
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profileName),
	)
	if err != nil {
		return nil, errors.NewAuthError("failed to load AWS config for profile", err)
	}

	// Try to make an STS call to force credential resolution and caching
	stsClient := sts.NewFromConfig(awsCfg)
	callerIdentity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, errors.NewAuthError("failed to retrieve credentials", err)
	}

	// Now get the resolved credentials
	creds, err := awsCfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, errors.NewAuthError("failed to retrieve credentials after STS call", err)
	}

	// Verify the credentials work by checking the caller identity
	logging.LogInfo("Retrieved credentials | account=%s profile=%s", *callerIdentity.Account, profileName)

	return &Credentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Region:          awsCfg.Region,
	}, nil
}

// configureProfile sets up the AWS profile with SSO settings
func (m *Manager) configureProfile(profileName string, cfg *appconfig.Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".aws")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to create AWS config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config")

	// Validate config path to prevent directory traversal
	if err := security.ValidateFilePath(configPath, configDir); err != nil {
		return fmt.Errorf("invalid config file path: %w", err)
	}

	// Read existing config
	var content string
	// #nosec G304
	if existing, err := os.ReadFile(configPath); err == nil {
		content = string(existing)
	}

	// Update or add profile configuration
	content = m.updateProfileInConfig(content, profileName, cfg)

	// Write back to file
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write AWS config: %w", err)
	}

	return nil
}

// getCachedToken retrieves a cached SSO token
func (m *Manager) getCachedToken(startURL string) (*SSOToken, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".aws", "sso", "cache")

	// First, try the expected filename based on SHA1 hash (AWS CLI compatible)
	// SECURITY NOTE: SHA1 is cryptographically weak but required for AWS CLI compatibility.
	// The hash is only used for cache filename generation, not for security purposes.
	// AWS CLI expects SHA1-based filenames in ~/.aws/sso/cache/
	// TODO: Monitor AWS CLI for migration to SHA256 or other secure alternatives
	hasher := sha1.New() // #nosec G401 -- SHA1 required for AWS CLI compatibility
	hasher.Write([]byte(startURL))
	hash := hex.EncodeToString(hasher.Sum(nil))
	expectedFile := filepath.Join(cacheDir, fmt.Sprintf("%s.json", hash))

	// Validate cache file path to prevent directory traversal
	if err := security.ValidateFilePath(expectedFile, cacheDir); err != nil {
		// Log but don't fail - just skip this file
		logging.LogWarn("Invalid cache file path, skipping: %v", err)
		return nil, fmt.Errorf("no valid SSO token found")
	}

	// #nosec G304
	if content, err := os.ReadFile(expectedFile); err == nil {
		var token SSOToken
		if json.Unmarshal(content, &token) == nil && token.StartURL == startURL {
			return &token, nil
		}
	}

	// Fallback: search through all cache files
	var tokenFile string
	err = filepath.WalkDir(cacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logging.LogWarn("Error walking cache directory at %s: %v", path, err)
			return nil // Continue walking
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			// Validate file path to prevent directory traversal attacks
			if validateErr := security.ValidateFilePath(path, cacheDir); validateErr != nil {
				logging.LogWarn("Invalid cache file path found and skipped: %s", path)
				// Skip invalid paths but continue walking
				return nil //nolint:nilerr
			}

			// Check if this file contains our start URL
			content, readErr := os.ReadFile(path) // #nosec G304
			if readErr != nil {
				logging.LogWarn("Could not read cache file %s: %v", path, readErr)
				return nil
			}

			var token SSOToken
			if json.Unmarshal(content, &token) == nil && token.StartURL == startURL {
				tokenFile = path
				return filepath.SkipAll // Found it, stop walking
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search for token files: %w", err)
	}

	if tokenFile == "" {
		return nil, fmt.Errorf("no cached token found for start URL: %s", startURL)
	}

	// Read and parse the token file
	content, err := os.ReadFile(tokenFile) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token SSOToken
	if err := json.Unmarshal(content, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &token, nil
}

// isTokenValid checks if an SSO token is still valid
func (m *Manager) isTokenValid(token *SSOToken) bool {
	if token == nil {
		return false
	}
	return time.Now().Before(token.ExpiresAt)
}

// performSSOLogin initiates the SSO login flow
func (m *Manager) performSSOLogin(ctx context.Context, awsCfg aws.Config, cfg *appconfig.Config) error {
	logging.LogInfo("Starting SSO device authorization flow...")

	// Create SSO OIDC client
	ssoOIDCClient := ssooidc.NewFromConfig(awsCfg)

	// Register client
	registerResp, err := ssoOIDCClient.RegisterClient(ctx, &ssooidc.RegisterClientInput{
		ClientName: aws.String("ztictl"),
		ClientType: aws.String("public"),
	})
	if err != nil {
		return fmt.Errorf("failed to register SSO client: %w", err)
	}

	// Start device authorization
	authResp, err := ssoOIDCClient.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     registerResp.ClientId,
		ClientSecret: registerResp.ClientSecret,
		StartUrl:     aws.String(cfg.SSO.StartURL),
	})
	if err != nil {
		return fmt.Errorf("failed to start device authorization: %w", err)
	}

	// Display authorization information and open browser automatically
	authURL := aws.ToString(authResp.VerificationUriComplete)
	userCode := aws.ToString(authResp.UserCode)

	fmt.Printf("\n🔐 AWS SSO Authentication Required\n")
	fmt.Printf("   Opening browser automatically to: %s\n", authURL)
	fmt.Printf("   If browser doesn't open, copy the URL above\n")
	fmt.Printf("   Your verification code: %s\n\n", userCode)

	// Attempt to open browser automatically
	if err := browser.OpenURL(authURL); err != nil {
		logging.LogWarn("Failed to open browser automatically | error=%v", err)
		fmt.Printf("⚠️  Please manually open the URL above in your browser\n")
	} else {
		fmt.Printf("✅ Browser opened automatically\n")
	}

	fmt.Printf("⏳ Waiting for authentication completion (do not close this terminal)...\n\n")

	// Poll for token
	logging.LogInfo("Polling for authentication completion...")

	ticker := time.NewTicker(time.Duration(authResp.Interval) * time.Second)
	defer ticker.Stop()

	// Use intelligent timeout: respect AWS timeout but ensure reasonable minimum
	// This balances security with usability
	timeoutSeconds := authResp.ExpiresIn

	if timeoutSeconds < MinTimeoutSeconds {
		timeoutSeconds = MinTimeoutSeconds
	} else if timeoutSeconds > MaxTimeoutSeconds {
		timeoutSeconds = MaxTimeoutSeconds
	}

	timeout := time.After(time.Duration(timeoutSeconds) * time.Second)

	// Inform user of timeout duration in user-friendly format
	timeoutMinutes := float64(timeoutSeconds) / 60.0
	if timeoutSeconds >= 60 {
		if timeoutSeconds%60 == 0 {
			// Exact minutes
			fmt.Printf("⏰ Authentication timeout: %.0f minutes\n\n", timeoutMinutes)
		} else {
			// Minutes with seconds
			minutes := timeoutSeconds / 60
			seconds := timeoutSeconds % 60
			fmt.Printf("⏰ Authentication timeout: %d minutes %d seconds\n\n", minutes, seconds)
		}
	} else {
		// Less than a minute
		fmt.Printf("⏰ Authentication timeout: %d seconds\n\n", timeoutSeconds)
	}

	for {
		select {
		case <-timeout:
			return fmt.Errorf("SSO login timed out after %d seconds", timeoutSeconds)
		case <-ticker.C:
			tokenResp, err := ssoOIDCClient.CreateToken(ctx, &ssooidc.CreateTokenInput{
				ClientId:     registerResp.ClientId,
				ClientSecret: registerResp.ClientSecret,
				DeviceCode:   authResp.DeviceCode,
				GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
			})
			if err != nil {
				// Check for authorization pending - this is expected while user authenticates
				if strings.Contains(err.Error(), "authorization_pending") ||
					strings.Contains(err.Error(), "AuthorizationPendingException") {
					continue // Keep polling
				}
				// Check for slow down request
				if strings.Contains(err.Error(), "slow_down") {
					time.Sleep(5 * time.Second) // Wait extra time
					continue
				}
				return fmt.Errorf("failed to create token: %w", err)
			}

			// Success! Save the token to cache
			if err := m.saveTokenToCache(tokenResp, cfg.SSO.StartURL, cfg.SSO.Region); err != nil {
				return fmt.Errorf("failed to save token to cache: %w", err)
			}

			logging.LogInfo("SSO login completed successfully")
			return nil
		}
	}
}

// saveTokenToCache saves an SSO token to the AWS cache
func (m *Manager) saveTokenToCache(tokenResp *ssooidc.CreateTokenOutput, startURL, region string) error {
	cacheDir, err := getAWSCacheDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cacheDir, 0750); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create token structure
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	token := SSOToken{
		StartURL:    startURL,
		Region:      region,
		AccessToken: aws.ToString(tokenResp.AccessToken),
		ExpiresAt:   expiresAt,
	}

	// Generate cache filename (AWS CLI compatible using SHA1)
	// SECURITY NOTE: SHA1 is cryptographically weak but required for AWS CLI compatibility.
	// The hash is only used for cache filename generation, not for security purposes.
	// AWS CLI expects SHA1-based filenames in ~/.aws/sso/cache/
	// TODO: Monitor AWS CLI for migration to SHA256 or other secure alternatives
	hasher := sha1.New() // #nosec G401 -- SHA1 required for AWS CLI compatibility
	hasher.Write([]byte(startURL))
	hash := hex.EncodeToString(hasher.Sum(nil))
	filename := fmt.Sprintf("%s.json", hash)
	cachePath := filepath.Join(cacheDir, filename)

	// Save to file
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// listAccounts retrieves available AWS accounts from SSO
func (m *Manager) listAccounts(ctx context.Context, client sso.ListAccountsAPIClient, accessToken string) ([]Account, error) {
	var accounts []Account
	paginator := sso.NewListAccountsPaginator(client, &sso.ListAccountsInput{
		AccessToken: aws.String(accessToken),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list accounts: %w", err)
		}

		for _, acc := range page.AccountList {
			accounts = append(accounts, Account{
				AccountID:    aws.ToString(acc.AccountId),
				AccountName:  aws.ToString(acc.AccountName),
				EmailAddress: aws.ToString(acc.EmailAddress),
			})
		}
	}

	return accounts, nil
}

// selectAccount provides interactive account selection with fuzzy finder for search capability
func (m *Manager) selectAccount(accounts []Account) (*Account, error) {
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts available")
	}

	if len(accounts) == 1 {
		return &accounts[0], nil
	}

	// Always use fuzzy finder for multiple accounts to enable search functionality
	// This provides consistent search experience regardless of account count
	return m.selectAccountFuzzy(accounts)
}

// selectAccountFuzzy uses fuzzy finder for account selection with full search capabilities
func (m *Manager) selectAccountFuzzy(accounts []Account) (*Account, error) {
	return safeSelectAccountFuzzy(m, accounts)
}

// safeSelectAccountFuzzy wraps fuzzy finder with panic recovery
func safeSelectAccountFuzzy(m *Manager, accounts []Account) (account *Account, err error) {
	defer func() {
		if r := recover(); r != nil {
			_, _ = color.New(color.FgRed, color.Bold).Printf("\n❌ An unexpected error occurred in the account selector.\n")    // #nosec G104
			_, _ = color.New(color.FgRed).Printf("   Please report this issue at: https://github.com/zsoftly/ztiaws/issues\n") // #nosec G104
			_, _ = color.New(color.FgYellow).Printf("   Error details: %v\n\n", r)                                             // #nosec G104
			logging.LogError("Panic in account selector | error=%v", r)
			err = fmt.Errorf("account selector encountered an unexpected error: %v", r)
			account = nil
		}
	}()
	return m.selectAccountFuzzyFallback(accounts)
}

// calculateAccountSelectorWidth determines optimal width for account selector
// based on content length, capped by the terminal width
func calculateAccountSelectorWidth(accounts []Account) int {
	const (
		minWidth       = 80 // Minimum comfortable width for readability
		promptLen      = 22 // "🔍 Type to search > " length
		borderPad      = 6  // Border chars + internal padding
		previewPad     = 4  // Preview window padding
		listPreviewGap = 2  // Gap between list and preview
	)

	terminalWidth := getTerminalWidth()

	maxContentWidth := promptLen

	// Find the longest account display string
	for _, account := range accounts {
		// Format: "AccountID - AccountName"
		displayLen := len(account.AccountID) + 3 + len(account.AccountName) // 3 for " - "
		if displayLen > maxContentWidth {
			maxContentWidth = displayLen
		}
	}

	// Find longest preview content
	maxPreviewWidth := 0
	for _, account := range accounts {
		previewLen := len(fmt.Sprintf("Account ID:   %s", account.AccountID))
		nameLen := len(fmt.Sprintf("Account Name: %s", account.AccountName))
		emailLen := len(fmt.Sprintf("Email:        %s", account.EmailAddress))

		maxLen := previewLen
		if nameLen > maxLen {
			maxLen = nameLen
		}
		if emailLen > maxLen {
			maxLen = emailLen
		}
		if maxLen > maxPreviewWidth {
			maxPreviewWidth = maxLen
		}
	}

	// Total width: list + preview + gaps + borders
	listWidth := maxContentWidth + 4 // content + padding
	previewWidthActual := maxPreviewWidth + previewPad
	totalWidth := listWidth + listPreviewGap + previewWidthActual + borderPad

	// Apply bounds
	if totalWidth < minWidth {
		return minWidth
	}
	// Ensure the width does not exceed the terminal width, with a small padding
	if totalWidth > terminalWidth {
		return terminalWidth - 2 // 2 for padding from terminal edge
	}

	return totalWidth
}

func (m *Manager) selectAccountFuzzyFallback(accounts []Account) (*Account, error) {
	// Get configurable display height from environment variable, default to 5 items
	maxDisplayItems := getDisplayItemCount()

	// Height calculation for bordered box:
	//   - N item lines (default 5)
	//   - 1 header line
	//   - 1 prompt line
	//   - 1 separator line
	//   - 2 border lines (top + bottom)
	// Total: N + 5 lines
	totalHeight := maxDisplayItems + 5 // items + header + prompt + separator + borders

	// Calculate dynamic width based on content
	dynamicWidth := calculateAccountSelectorWidth(accounts)

	// Create mutex for thread-safe account list access
	var accountsMutex sync.RWMutex

	// Note: Pass &accounts (pointer) because WithHotReloadLock requires it
	idx, err := fuzzyfinder.Find(&accounts,
		func(i int) string {
			return fmt.Sprintf("%s - %s", accounts[i].AccountID, accounts[i].AccountName)
		},
		fuzzyfinder.WithCursorPosition(fuzzyfinder.CursorPositionBottom),
		fuzzyfinder.WithPromptString("🔍 Type to search > "),
		fuzzyfinder.WithHeader(fmt.Sprintf("Select AWS Account (%d available)", len(accounts))),
		fuzzyfinder.WithMode(fuzzyfinder.ModeSmart),
		fuzzyfinder.WithHotReloadLock(&accountsMutex),
		fuzzyfinder.WithHeight(totalHeight),
		fuzzyfinder.WithWidth(dynamicWidth),
		fuzzyfinder.WithHorizontalAlignment(fuzzyfinder.AlignLeft),
		fuzzyfinder.WithBorder(),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 || i >= len(accounts) {
				return ""
			}
			account := accounts[i]
			return fmt.Sprintf("Account ID:   %s\nAccount Name: %s\nEmail:        %s",
				account.AccountID,
				account.AccountName,
				account.EmailAddress)
		}),
	)

	if err != nil {
		if err.Error() == "abort" {
			_, _ = color.New(color.FgRed).Printf("❌ Account selection cancelled\n") // #nosec G104
			return nil, fmt.Errorf("account selection cancelled")
		}
		return nil, fmt.Errorf("account selection failed: %w", err)
	}

	// Display selection confirmation
	_, _ = color.New(color.FgGreen, color.Bold).Printf("✅ Selected: %s (%s)\n", accounts[idx].AccountName, accounts[idx].AccountID) // #nosec G104

	return &accounts[idx], nil
}

// listAccountRoles retrieves available roles for an account
func (m *Manager) listAccountRoles(ctx context.Context, client sso.ListAccountRolesAPIClient, accessToken, accountID string) ([]Role, error) {
	var roles []Role
	paginator := sso.NewListAccountRolesPaginator(client, &sso.ListAccountRolesInput{
		AccessToken: aws.String(accessToken),
		AccountId:   aws.String(accountID),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list account roles: %w", err)
		}

		for _, role := range page.RoleList {
			roles = append(roles, Role{
				RoleName:  aws.ToString(role.RoleName),
				AccountID: accountID,
			})
		}
	}

	return roles, nil
}

// selectRole provides interactive role selection with fuzzy finder for search capability
func (m *Manager) selectRole(roles []Role, account *Account) (*Role, error) {
	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles available for account %s", account.AccountID)
	}

	if len(roles) == 1 {
		return &roles[0], nil
	}

	// Always use fuzzy finder for multiple roles to enable search functionality
	// This provides consistent search experience regardless of role count
	return m.selectRoleFuzzy(roles, account)
}

// selectRoleFuzzy uses fuzzy finder for role selection with full search capabilities
func (m *Manager) selectRoleFuzzy(roles []Role, account *Account) (*Role, error) {
	return safeSelectRoleFuzzy(m, roles, account)
}

// safeSelectRoleFuzzy wraps fuzzy finder with panic recovery
func safeSelectRoleFuzzy(m *Manager, roles []Role, account *Account) (role *Role, err error) {
	defer func() {
		if r := recover(); r != nil {
			_, _ = color.New(color.FgRed, color.Bold).Printf("\n❌ An unexpected error occurred in the role selector.\n")       // #nosec G104
			_, _ = color.New(color.FgRed).Printf("   Please report this issue at: https://github.com/zsoftly/ztiaws/issues\n") // #nosec G104
			_, _ = color.New(color.FgYellow).Printf("   Error details: %v\n\n", r)                                             // #nosec G104
			logging.LogError("Panic in role selector | error=%v", r)
			err = fmt.Errorf("role selector encountered an unexpected error: %v", r)
			role = nil
		}
	}()
	return m.selectRoleFuzzyFallback(roles, account)
}

// calculateRoleSelectorWidth determines optimal width for role selector
// based on content length, capped by the terminal width
func calculateRoleSelectorWidth(roles []Role, account *Account) int {
	const (
		minWidth       = 80 // Minimum comfortable width for readability
		promptLen      = 22 // "🎭 Type to search > " length
		borderPad      = 6  // Border chars + internal padding
		previewPad     = 4  // Preview window padding
		listPreviewGap = 2  // Gap between list and preview
	)

	terminalWidth := getTerminalWidth()

	maxContentWidth := promptLen

	// Find the longest role name
	for _, role := range roles {
		if len(role.RoleName) > maxContentWidth {
			maxContentWidth = len(role.RoleName)
		}
	}

	// Also consider the header length
	headerLen := len(fmt.Sprintf("Select Role for %s (%d available)", account.AccountName, len(roles)))
	if headerLen > maxContentWidth {
		maxContentWidth = headerLen
	}

	// Find longest preview content
	maxPreviewWidth := 0
	for _, role := range roles {
		roleLen := len(fmt.Sprintf("Role:         %s", role.RoleName))
		accountLen := len(fmt.Sprintf("Account:      %s", account.AccountName))
		idLen := len(fmt.Sprintf("Account ID:   %s", account.AccountID))

		maxLen := roleLen
		if accountLen > maxLen {
			maxLen = accountLen
		}
		if idLen > maxLen {
			maxLen = idLen
		}
		if maxLen > maxPreviewWidth {
			maxPreviewWidth = maxLen
		}
	}

	// Total width: list + preview + gaps + borders
	listWidth := maxContentWidth + 4 // content + padding
	previewWidthActual := maxPreviewWidth + previewPad
	totalWidth := listWidth + listPreviewGap + previewWidthActual + borderPad

	// Apply bounds
	if totalWidth < minWidth {
		return minWidth
	}
	// Ensure the width does not exceed the terminal width, with a small padding
	if totalWidth > terminalWidth {
		return terminalWidth - 2 // 2 for padding from terminal edge
	}

	return totalWidth
}

func (m *Manager) selectRoleFuzzyFallback(roles []Role, account *Account) (*Role, error) {
	// Get configurable display height from environment variable, default to 5 items
	maxDisplayItems := getDisplayItemCount()

	// Height calculation for bordered box:
	//   - N item lines (default 5)
	//   - 1 header line
	//   - 1 prompt line
	//   - 1 separator line
	//   - 2 border lines (top + bottom)
	// Total: N + 5 lines
	totalHeight := maxDisplayItems + 5 // items + header + prompt + separator + borders

	// Calculate dynamic width based on content
	dynamicWidth := calculateRoleSelectorWidth(roles, account)

	// Create mutex for thread-safe roles list access
	var rolesMutex sync.RWMutex

	// Note: Pass &roles (pointer) because WithHotReloadLock requires it
	idx, err := fuzzyfinder.Find(&roles,
		func(i int) string {
			return roles[i].RoleName
		},
		fuzzyfinder.WithCursorPosition(fuzzyfinder.CursorPositionBottom),
		fuzzyfinder.WithPromptString("🎭 Type to search > "),
		fuzzyfinder.WithHeader(fmt.Sprintf("Select Role for %s (%d available)",
			account.AccountName, len(roles))),
		fuzzyfinder.WithMode(fuzzyfinder.ModeSmart),
		fuzzyfinder.WithHotReloadLock(&rolesMutex),
		fuzzyfinder.WithHeight(totalHeight),
		fuzzyfinder.WithWidth(dynamicWidth),
		fuzzyfinder.WithHorizontalAlignment(fuzzyfinder.AlignLeft),
		fuzzyfinder.WithBorder(),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 || i >= len(roles) {
				return ""
			}
			role := roles[i]
			return fmt.Sprintf("Role:         %s\nAccount:      %s\nAccount ID:   %s",
				role.RoleName,
				account.AccountName,
				account.AccountID)
		}),
	)

	if err != nil {
		if err.Error() == "abort" {
			_, _ = color.New(color.FgRed).Printf("❌ Role selection cancelled\n") // #nosec G104
			return nil, fmt.Errorf("role selection cancelled")
		}
		return nil, fmt.Errorf("role selection failed: %w", err)
	}

	// Display selection confirmation
	_, _ = color.New(color.FgGreen, color.Bold).Printf("✅ Selected: %s\n", roles[idx].RoleName) // #nosec G104

	return &roles[idx], nil
}

// updateProfileWithSelection updates the AWS profile with selected account and role
func (m *Manager) updateProfileWithSelection(profileName string, account *Account, role *Role, cfg *appconfig.Config) error {
	configDir, err := getAWSConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config")

	// Validate config path to prevent directory traversal
	if err := security.ValidateFilePath(configPath, configDir); err != nil {
		return fmt.Errorf("invalid config file path: %w", err)
	}

	// Read existing config
	content, err := os.ReadFile(configPath) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to read AWS config: %w", err)
	}

	// Update profile with account and role
	updatedContent := m.updateProfileWithAccountRole(string(content), profileName, account, role, cfg)

	// Write back
	if err := os.WriteFile(configPath, []byte(updatedContent), 0600); err != nil {
		return fmt.Errorf("failed to write AWS config: %w", err)
	}

	return nil
}

// Helper methods for config file manipulation, profile parsing, etc.
// These would contain the detailed implementation for AWS config file handling

func (m *Manager) updateProfileInConfig(content, profileName string, cfg *appconfig.Config) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inTargetProfile bool
	var targetProfileSection string
	profileUpdated := false

	targetProfileSection = fmt.Sprintf("[profile %s]", profileName)
	if profileName == "default" {
		targetProfileSection = "[default]"
	}

	// Parse existing content and update the target profile
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Check if this is a profile header
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			// If we were in target profile, we've finished processing it
			if inTargetProfile {
				inTargetProfile = false
			}

			// Check if this is our target profile
			if trimmedLine == targetProfileSection {
				inTargetProfile = true
				profileUpdated = true
				result = append(result, line)

				// Add or update the SSO settings for this profile
				result = append(result, fmt.Sprintf("sso_start_url = %s", cfg.SSO.StartURL))
				result = append(result, fmt.Sprintf("sso_region = %s", cfg.SSO.Region))
				result = append(result, fmt.Sprintf("region = %s", cfg.DefaultRegion))
				result = append(result, "output = json")

				// Skip existing settings for this profile - we'll replace them
				j := i + 1
				for j < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[j]), "[") {
					// Skip lines that are SSO-related settings we're replacing
					nextLine := strings.TrimSpace(lines[j])
					if strings.HasPrefix(nextLine, "sso_start_url") ||
						strings.HasPrefix(nextLine, "sso_region") ||
						strings.HasPrefix(nextLine, "region") ||
						strings.HasPrefix(nextLine, "output") {
						j++
						continue
					}

					// Keep other settings (like sso_account_id, sso_role_name)
					if nextLine != "" {
						result = append(result, lines[j])
					}
					j++
				}

				// Fast forward the main loop
				i = j - 1
				continue
			} else {
				// Different profile section
				result = append(result, line)
			}
		} else if !inTargetProfile {
			// Not in target profile, keep the line
			result = append(result, line)
		}
		// If in target profile, we're skipping old lines (handled above)
	}

	// If profile wasn't found, add it at the end
	if !profileUpdated {
		if len(result) > 0 && result[len(result)-1] != "" {
			result = append(result, "")
		}
		result = append(result, targetProfileSection)
		result = append(result, fmt.Sprintf("sso_start_url = %s", cfg.SSO.StartURL))
		result = append(result, fmt.Sprintf("sso_region = %s", cfg.SSO.Region))
		result = append(result, fmt.Sprintf("region = %s", cfg.DefaultRegion))
		result = append(result, "output = json")
	}

	return strings.Join(result, "\n")
}

func (m *Manager) updateProfileWithAccountRole(content, profileName string, account *Account, role *Role, cfg *appconfig.Config) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inTargetProfile bool
	var targetProfileSection string
	profileUpdated := false

	targetProfileSection = fmt.Sprintf("[profile %s]", profileName)
	if profileName == "default" {
		targetProfileSection = "[default]"
	}

	// Parse existing content and update the target profile
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Check if this is a profile header
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			// If we were in target profile, we've finished processing it
			if inTargetProfile {
				inTargetProfile = false
			}

			// Check if this is our target profile
			if trimmedLine == targetProfileSection {
				inTargetProfile = true
				profileUpdated = true
				result = append(result, line)

				// Collect all existing settings, then add/update account and role
				var profileSettings []string
				j := i + 1
				for ; j < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[j]), "["); j++ {
					nextLine := strings.TrimSpace(lines[j])
					if nextLine != "" && !strings.HasPrefix(nextLine, "sso_account_id") && !strings.HasPrefix(nextLine, "sso_role_name") {
						profileSettings = append(profileSettings, lines[j])
					}
				}

				// Add all existing settings first
				result = append(result, profileSettings...)

				// Add the account and role information
				result = append(result, fmt.Sprintf("sso_account_id = %s", account.AccountID))
				result = append(result, fmt.Sprintf("sso_role_name = %s", role.RoleName))

				// Fast forward the main loop
				i = j - 1
				continue
			} else {
				// Different profile section
				result = append(result, line)
			}
		} else if !inTargetProfile {
			// Not in target profile, keep the line
			result = append(result, line)
		}
		// If in target profile, we're skipping old lines (handled above)
	}

	// If profile wasn't found, create it with all settings
	if !profileUpdated {
		if len(result) > 0 && result[len(result)-1] != "" {
			result = append(result, "")
		}
		result = append(result, targetProfileSection)
		result = append(result, fmt.Sprintf("sso_start_url = %s", cfg.SSO.StartURL))
		result = append(result, fmt.Sprintf("sso_region = %s", cfg.SSO.Region))
		result = append(result, fmt.Sprintf("region = %s", cfg.DefaultRegion))
		result = append(result, "output = json")
		result = append(result, fmt.Sprintf("sso_account_id = %s", account.AccountID))
		result = append(result, fmt.Sprintf("sso_role_name = %s", role.RoleName))
	}

	return strings.Join(result, "\n")
}

func (m *Manager) parseProfiles(content string) []Profile {
	profiles := []Profile{}
	profileMap := make(map[string]*Profile) // Use map to avoid duplicates
	lines := strings.Split(content, "\n")

	var currentProfile *Profile

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for profile header [profile name] or [default]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Parse profile name
			profileLine := strings.Trim(line, "[]")
			var profileName string

			if profileLine == "default" {
				profileName = "default"
			} else if strings.HasPrefix(profileLine, "profile ") {
				profileName = strings.TrimPrefix(profileLine, "profile ")
			} else {
				continue // Skip unrecognized sections
			}

			// Check if profile already exists (handle duplicates)
			if existingProfile, exists := profileMap[profileName]; exists {
				currentProfile = existingProfile
			} else {
				// Create new profile
				currentProfile = &Profile{
					Name: profileName,
				}
				profileMap[profileName] = currentProfile
			}
			continue
		}

		// Parse key-value pairs for current profile
		if currentProfile != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "sso_start_url":
					currentProfile.SSOStartURL = value
				case "sso_region":
					currentProfile.SSORegion = value
				case "region":
					currentProfile.Region = value
				case "sso_account_id":
					currentProfile.AccountID = value
				case "sso_role_name":
					currentProfile.RoleName = value
				}
			}
		}
	}

	// Convert map to slice
	for _, profile := range profileMap {
		profiles = append(profiles, *profile)
	}

	return profiles
}

// isProfileAuthenticated checks if a profile has valid cached tokens
func (m *Manager) IsProfileAuthenticated(profileName string) bool {
	// Check if SSO token cache exists and is valid
	configDir, err := getAWSCacheDir()
	if err != nil {
		return false
	}
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return false
	}

	// Look for cached tokens for this profile
	files, err := os.ReadDir(configDir)
	if err != nil {
		return false
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			// Read and check if token is still valid
			tokenFile := filepath.Join(configDir, file.Name())

			// Validate token file path to prevent directory traversal
			if err := security.ValidateFilePath(tokenFile, configDir); err != nil {
				// Skip invalid paths but continue checking other files
				continue
			}

			content, err := os.ReadFile(tokenFile) // #nosec G304
			if err != nil {
				continue
			}

			var token struct {
				AccessToken string `json:"accessToken"`
				ExpiresAt   string `json:"expiresAt"`
			}

			if err := json.Unmarshal(content, &token); err != nil {
				continue
			}

			// Check if token is expired
			if expiresAt, err := time.Parse(time.RFC3339, token.ExpiresAt); err == nil {
				if time.Now().Before(expiresAt) {
					return true
				}
			}
		}
	}

	return false
}

func (m *Manager) isProfileAuthenticated(ctx context.Context, profileName string) (bool, error) {
	// First try the newer method that checks SSO cache directly
	if authenticated := m.IsProfileAuthenticated(profileName); authenticated {
		return true, nil
	}

	// Fallback to trying AWS API call
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profileName))
	if err != nil {
		return false, nil
	}

	stsClient := sts.NewFromConfig(awsCfg)
	_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	return err == nil, nil
}

// printSuccessMessage displays platform-specific instructions after successful authentication
func (m *Manager) printSuccessMessage(profileName string, account *Account, role *Role, cfg *appconfig.Config) {
	// Color setup - using high-intensity colors for better visibility
	successColor := color.New(color.FgHiGreen, color.Bold)
	infoColor := color.New(color.FgHiCyan)
	commandColor := color.New(color.FgHiYellow)

	fmt.Println()
	_, _ = successColor.Println("🎉 Successfully configured AWS SSO profile.") // #nosec G104
	fmt.Println("----------------------------------------")
	_, _ = infoColor.Printf("Account: %s\n", account.AccountName) // #nosec G104
	_, _ = infoColor.Printf("Role: %s\n", role.RoleName)          // #nosec G104
	_, _ = infoColor.Printf("Profile: %s\n", profileName)         // #nosec G104
	fmt.Println()

	// Platform-specific instructions
	_, _ = infoColor.Println("To use this profile, run:") // #nosec G104

	switch runtime.GOOS {
	case "windows":
		// Windows Command Prompt instructions
		fmt.Println()
		_, _ = infoColor.Println("For Command Prompt (cmd):")                     // #nosec G104
		_, _ = commandColor.Printf("set AWS_PROFILE=%s\n", profileName)           // #nosec G104
		_, _ = commandColor.Printf("set AWS_DEFAULT_REGION=%s\n", cfg.SSO.Region) // #nosec G104

		fmt.Println()
		_, _ = infoColor.Println("For PowerShell:")                                    // #nosec G104
		_, _ = commandColor.Printf("$env:AWS_PROFILE=\"%s\"\n", profileName)           // #nosec G104
		_, _ = commandColor.Printf("$env:AWS_DEFAULT_REGION=\"%s\"\n", cfg.SSO.Region) // #nosec G104

	default:
		// Unix/Linux/macOS instructions
		_, _ = commandColor.Printf("export AWS_PROFILE=%s AWS_DEFAULT_REGION=%s\n", profileName, cfg.SSO.Region) // #nosec G104
	}

	fmt.Println()
	_, _ = infoColor.Println("To view your credentials, run:")        // #nosec G104
	_, _ = commandColor.Printf("ztictl auth creds %s\n", profileName) // #nosec G104

	fmt.Println()
	_, _ = infoColor.Println("To list EC2 instances, run:") // #nosec G104
	_, _ = commandColor.Printf("ztictl ssm list\n")         // #nosec G104
}

func (m *Manager) clearSSOCache() error {
	// For Windows compatibility, we'll disable automatic cache clearing
	// Users can manually clear cache using AWS CLI: aws sso logout
	logging.LogInfo("Cache clearing disabled for security compatibility")
	return nil
}
