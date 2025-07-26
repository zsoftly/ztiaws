package auth

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/browser"
	appconfig "ztictl/internal/config"
	"ztictl/internal/logging"
	"ztictl/pkg/errors"
)

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

// NewManager creates a new authentication manager
func NewManager(logger *logging.Logger) *Manager {
	return &Manager{
		logger: logger,
	}
}

// Login performs AWS SSO login with interactive account and role selection
func (m *Manager) Login(ctx context.Context, profileName string) error {
	cfg := appconfig.Get()

	if cfg.SSO.StartURL == "" {
		return errors.NewValidationError("SSO start URL not configured. Please run 'ztictl config init' first")
	}

	m.logger.Info("Starting AWS SSO authentication", "profile", profileName)

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.SSO.Region))
	if err != nil {
		return errors.NewAWSError("failed to load AWS config", err)
	}

	// Configure the profile with SSO settings
	if err := m.configureProfile(profileName, cfg); err != nil {
		return fmt.Errorf("failed to configure profile: %w", err)
	}

	// Check for valid cached token
	token, err := m.getCachedToken(cfg.SSO.StartURL)
	if err != nil || !m.isTokenValid(token) {
		m.logger.Info("No valid cached token found, initiating SSO login...")

		// Perform SSO login
		if err := m.performSSOLogin(ctx, awsCfg, profileName, cfg); err != nil {
			return err
		}

		// Get the new token
		token, err = m.getCachedToken(cfg.SSO.StartURL)
		if err != nil {
			return fmt.Errorf("failed to get token after login: %w", err)
		}
	} else {
		m.logger.Info("Using valid cached SSO token")
	}

	// Get available accounts
	accounts, err := m.listAccounts(ctx, awsCfg, token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	// Interactive account selection
	selectedAccount, err := m.selectAccount(accounts)
	if err != nil {
		return fmt.Errorf("account selection failed: %w", err)
	}

	m.logger.Info("Selected account", "id", selectedAccount.AccountID, "name", selectedAccount.AccountName)

	// Get available roles for the selected account
	roles, err := m.listAccountRoles(ctx, awsCfg, token.AccessToken, selectedAccount.AccountID)
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	// Interactive role selection
	selectedRole, err := m.selectRole(roles, selectedAccount)
	if err != nil {
		return fmt.Errorf("role selection failed: %w", err)
	}

	m.logger.Info("Selected role", "role", selectedRole.RoleName)

	// Update profile with selected account and role
	if err := m.updateProfileWithSelection(profileName, selectedAccount, selectedRole, cfg); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	m.logger.Info("AWS SSO authentication completed successfully",
		"profile", profileName,
		"account", selectedAccount.AccountName,
		"role", selectedRole.RoleName)

	return nil
}

// Logout performs AWS SSO logout
func (m *Manager) Logout(ctx context.Context, profileName string) error {
	if profileName != "" {
		m.logger.Info("Logging out from specific profile", "profile", profileName)
		// For now, we'll clear the specific profile's cached credentials
		// In a full implementation, this would clear SSO cache for the specific profile
	} else {
		m.logger.Info("Logging out from all SSO sessions")
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
	configPath := filepath.Join(os.Getenv("HOME"), ".aws", "config")

	content, err := os.ReadFile(configPath)
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
			m.logger.Warn("Failed to check authentication status", "profile", profiles[i].Name, "error", err)
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
	m.logger.Info("Retrieved credentials",
		"account", *callerIdentity.Account,
		"arn", *callerIdentity.Arn,
		"profile", profileName)

	return &Credentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Region:          awsCfg.Region,
	}, nil
}

// configureProfile sets up the AWS profile with SSO settings
func (m *Manager) configureProfile(profileName string, cfg *appconfig.Config) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".aws")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create AWS config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config")

	// Read existing config
	var content string
	if existing, err := os.ReadFile(configPath); err == nil {
		content = string(existing)
	}

	// Update or add profile configuration
	content = m.updateProfileInConfig(content, profileName, cfg)

	// Write back to file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write AWS config: %w", err)
	}

	return nil
}

// getCachedToken retrieves a cached SSO token
func (m *Manager) getCachedToken(startURL string) (*SSOToken, error) {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".aws", "sso", "cache")

	// First, try the expected filename based on SHA1 hash
	hasher := sha1.New()
	hasher.Write([]byte(startURL))
	hash := hex.EncodeToString(hasher.Sum(nil))
	expectedFile := filepath.Join(cacheDir, fmt.Sprintf("%s.json", hash))

	if content, err := os.ReadFile(expectedFile); err == nil {
		var token SSOToken
		if json.Unmarshal(content, &token) == nil && token.StartURL == startURL {
			return &token, nil
		}
	}

	// Fallback: search through all cache files
	var tokenFile string
	err := filepath.WalkDir(cacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			// Check if this file contains our start URL
			content, readErr := os.ReadFile(path)
			if readErr != nil {
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
	content, err := os.ReadFile(tokenFile)
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
func (m *Manager) performSSOLogin(ctx context.Context, awsCfg aws.Config, profileName string, cfg *appconfig.Config) error {
	m.logger.Info("Starting SSO device authorization flow...")

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
		m.logger.Warn("Failed to open browser automatically", "error", err)
		fmt.Printf("⚠️  Please manually open the URL above in your browser\n")
	} else {
		fmt.Printf("✅ Browser opened automatically\n")
	}

	fmt.Printf("⏳ Waiting for authentication completion (do not close this terminal)...\n\n")

	// Poll for token
	m.logger.Info("Polling for authentication completion...")

	ticker := time.NewTicker(time.Duration(authResp.Interval) * time.Second)
	defer ticker.Stop()

	timeout := time.After(time.Duration(authResp.ExpiresIn) * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("SSO login timed out after %d seconds", authResp.ExpiresIn)
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

			m.logger.Info("SSO login completed successfully")
			return nil
		}
	}
}

// saveTokenToCache saves an SSO token to the AWS cache
func (m *Manager) saveTokenToCache(tokenResp *ssooidc.CreateTokenOutput, startURL, region string) error {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".aws", "sso", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
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

	// Generate cache filename (AWS CLI compatible)
	hasher := sha1.New()
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
func (m *Manager) listAccounts(ctx context.Context, awsCfg aws.Config, accessToken string) ([]Account, error) {
	cfg := appconfig.Get()

	// Create a new config specifically for SSO operations using the configured SSO region
	ssoConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.SSO.Region), // Use the configured SSO region
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load SSO config: %w", err)
	}

	// Create SSO client with explicit configuration
	ssoClient := sso.NewFromConfig(ssoConfig)

	resp, err := ssoClient.ListAccounts(ctx, &sso.ListAccountsInput{
		AccessToken: aws.String(accessToken),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	accounts := make([]Account, len(resp.AccountList))
	for i, acc := range resp.AccountList {
		accounts[i] = Account{
			AccountID:    aws.ToString(acc.AccountId),
			AccountName:  aws.ToString(acc.AccountName),
			EmailAddress: aws.ToString(acc.EmailAddress),
		}
	}

	return accounts, nil
}

// selectAccount provides interactive account selection using fzf
func (m *Manager) selectAccount(accounts []Account) (*Account, error) {
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts available")
	}

	if len(accounts) == 1 {
		return &accounts[0], nil
	}

	// Prepare items for fzf
	idx, err := fuzzyfinder.Find(accounts, func(i int) string {
		return fmt.Sprintf("%s | %s", accounts[i].AccountID, accounts[i].AccountName)
	}, fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
		return fmt.Sprintf("Account ID: %s\nName: %s\nEmail: %s",
			accounts[i].AccountID, accounts[i].AccountName, accounts[i].EmailAddress)
	}))

	if err != nil {
		return nil, fmt.Errorf("account selection cancelled: %w", err)
	}

	return &accounts[idx], nil
}

// listAccountRoles retrieves available roles for an account
func (m *Manager) listAccountRoles(ctx context.Context, awsCfg aws.Config, accessToken, accountID string) ([]Role, error) {
	cfg := appconfig.Get()

	// Create a new config specifically for SSO operations using the configured SSO region
	ssoConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.SSO.Region), // Use the configured SSO region
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load SSO config: %w", err)
	}

	// Create SSO client with explicit configuration
	ssoClient := sso.NewFromConfig(ssoConfig)

	resp, err := ssoClient.ListAccountRoles(ctx, &sso.ListAccountRolesInput{
		AccessToken: aws.String(accessToken),
		AccountId:   aws.String(accountID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list account roles: %w", err)
	}

	roles := make([]Role, len(resp.RoleList))
	for i, role := range resp.RoleList {
		roles[i] = Role{
			RoleName:  aws.ToString(role.RoleName),
			AccountID: accountID,
		}
	}

	return roles, nil
}

// selectRole provides interactive role selection
func (m *Manager) selectRole(roles []Role, account *Account) (*Role, error) {
	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles available for account %s", account.AccountID)
	}

	if len(roles) == 1 {
		return &roles[0], nil
	}

	// Prepare items for fzf
	idx, err := fuzzyfinder.Find(roles, func(i int) string {
		return roles[i].RoleName
	}, fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
		return fmt.Sprintf("Role: %s\nAccount: %s\nAccount ID: %s",
			roles[i].RoleName, account.AccountName, account.AccountID)
	}))

	if err != nil {
		return nil, fmt.Errorf("role selection cancelled: %w", err)
	}

	return &roles[idx], nil
}

// updateProfileWithSelection updates the AWS profile with selected account and role
func (m *Manager) updateProfileWithSelection(profileName string, account *Account, role *Role, cfg *appconfig.Config) error {
	configPath := filepath.Join(os.Getenv("HOME"), ".aws", "config")

	// Read existing config
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read AWS config: %w", err)
	}

	// Update profile with account and role
	updatedContent := m.updateProfileWithAccountRole(string(content), profileName, account, role, cfg)

	// Write back
	if err := os.WriteFile(configPath, []byte(updatedContent), 0644); err != nil {
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
	for i, line := range lines {
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
	for i, line := range lines {
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
				for j < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[j]), "[") {
					nextLine := strings.TrimSpace(lines[j])
					if nextLine != "" && !strings.HasPrefix(nextLine, "sso_account_id") && !strings.HasPrefix(nextLine, "sso_role_name") {
						profileSettings = append(profileSettings, lines[j])
					}
					j++
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
	configDir := filepath.Join(os.Getenv("HOME"), ".aws", "sso", "cache")
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
			content, err := os.ReadFile(tokenFile)
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

func (m *Manager) clearSSOCache() error {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".aws", "sso", "cache")
	return os.RemoveAll(cacheDir)
}
