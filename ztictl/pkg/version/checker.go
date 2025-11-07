package version

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	githubAPIURL    = "https://api.github.com/repos/zsoftly/ztiaws/releases/latest"
	cacheExpiration = 24 * time.Hour
	installDocsURL  = "https://github.com/zsoftly/ztiaws/blob/main/INSTALLATION.md"
	maxRetries      = 3
	retryDelay      = 1 * time.Second
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

type VersionCache struct {
	LatestVersion string    `json:"latest_version"`
	CheckedAt     time.Time `json:"checked_at"`
}

// CheckLatestVersion checks if there's a newer version available
func CheckLatestVersion(currentVersion string) (isOutdated bool, latestVersion string, err error) {
	// Check if version check is explicitly disabled
	if os.Getenv("ZTICTL_SKIP_VERSION_CHECK") == "true" {
		return false, "", fmt.Errorf("version check disabled by ZTICTL_SKIP_VERSION_CHECK")
	}

	// Try to get from cache first
	cached, err := getFromCache()
	if err == nil && time.Since(cached.CheckedAt) < cacheExpiration {
		return compareVersions(currentVersion, cached.LatestVersion), cached.LatestVersion, nil
	}

	// Fetch from GitHub API with retry logic
	client := createHTTPClient()

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay * time.Duration(attempt))
		}

		req, err := http.NewRequest("GET", githubAPIURL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		// Add GitHub token authentication if available (increases rate limit from 60 to 5000/hour)
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			req.Header.Set("Authorization", "token "+token)
		}

		// Set User-Agent to identify the client
		req.Header.Set("User-Agent", "ztictl-version-checker")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("network request failed (attempt %d/%d): %w", attempt+1, maxRetries, err)
			continue
		}

		// Handle response
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
			// Don't retry on client errors (4xx)
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				break
			}
			continue
		}

		var release GitHubRelease
		if err := json.Unmarshal(body, &release); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		latestVersion = strings.TrimPrefix(release.TagName, "v")

		// Cache the result
		_ = saveToCache(latestVersion)

		return compareVersions(currentVersion, latestVersion), latestVersion, nil
	}

	return false, "", lastErr
}

// createHTTPClient creates an HTTP client configured for production environments
func createHTTPClient() *http.Client {
	// Create transport with proxy support
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment, // Respects HTTP_PROXY, HTTPS_PROXY, NO_PROXY
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12, // Use TLS 1.2 or higher
		},
	}

	// SECURITY WARNING: This option disables TLS certificate verification and makes
	// connections vulnerable to man-in-the-middle (MITM) attacks. Only use this in
	// trusted corporate environments with custom certificate authorities where you
	// cannot install the CA certificates. This should NEVER be used in production
	// without understanding the security implications.
	if os.Getenv("ZTICTL_INSECURE_SKIP_VERIFY") == "true" {
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	// Support custom proxy configuration
	if proxyURL := os.Getenv("ZTICTL_HTTPS_PROXY"); proxyURL != "" {
		if proxy, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}

	return &http.Client{
		Timeout:   10 * time.Second, // Increased timeout for production networks
		Transport: transport,
	}
}

// compareVersions returns true if current is older than latest
func compareVersions(current, latest string) bool {
	// Clean both versions (trim v/V prefix and strip pre-release/build suffixes like -rc1 or +build)
	current = cleanVersion(current)
	latest = cleanVersion(latest)

	// Parse versions as major.minor.patch
	currentParts := parseVersion(current)
	latestParts := parseVersion(latest)

	// Compare major.minor.patch numerically
	for i := 0; i < 3; i++ {
		if currentParts[i] < latestParts[i] {
			return true
		}
		if currentParts[i] > latestParts[i] {
			return false
		}
	}

	return false // versions are equal
}

// cleanVersion normalizes a version string by removing leading 'v'/'V' and
// stripping any pre-release (e.g., -rc1) or build metadata (e.g., +build.1)
// to leave only the numeric major.minor.patch portion for comparison.
func cleanVersion(v string) string {
	// Trim optional leading 'v' or 'V' character (if present)
	if len(v) > 0 {
		if v[0] == 'v' || v[0] == 'V' {
			v = v[1:]
		}
	}

	// Strip at first '-' or '+' (pre-release or build metadata)
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}

	return v
}

// parseVersion parses a semantic version string into [major, minor, patch]
func parseVersion(v string) [3]int {
	parts := strings.Split(v, ".")
	result := [3]int{0, 0, 0}

	for i := 0; i < len(parts) && i < 3; i++ {
		// Parse as int; on error, leave as 0 (invalid or empty segments default to 0)
		if num, err := strconv.Atoi(parts[i]); err == nil {
			result[i] = num
		}
	}

	return result
}

// getCacheFilePath returns the path to the version cache file
func getCacheFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ztictl_version_cache.json"), nil
}

// getFromCache reads the cached version info
func getFromCache() (*VersionCache, error) {
	cachePath, err := getCacheFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cache VersionCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// saveToCache writes version info to cache
func saveToCache(latestVersion string) error {
	cachePath, err := getCacheFilePath()
	if err != nil {
		return err
	}

	cache := VersionCache{
		LatestVersion: latestVersion,
		CheckedAt:     time.Now(),
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0600)
}

// PrintVersionWithCheck prints version info and checks for updates
func PrintVersionWithCheck(currentVersion string) {
	fmt.Printf("ztictl version %s\n", currentVersion)

	isOutdated, latestVersion, err := CheckLatestVersion(currentVersion)
	if err != nil {
		// Show error in debug mode or if explicitly requested
		if os.Getenv("ZTICTL_DEBUG") == "true" || os.Getenv("ZTICTL_VERBOSE_VERSION") == "true" {
			fmt.Fprintf(os.Stderr, "\n[WARN] Version check failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "       This is normal in restricted network environments.\n")
			fmt.Fprintf(os.Stderr, "       To disable this check, set: ZTICTL_SKIP_VERSION_CHECK=true\n")

			// Provide helpful hints for common issues
			if strings.Contains(err.Error(), "network request failed") {
				fmt.Fprintf(os.Stderr, "\n[INFO] Troubleshooting tips:\n")
				fmt.Fprintf(os.Stderr, "       - Check network connectivity to api.github.com\n")
				fmt.Fprintf(os.Stderr, "       - Configure proxy: export HTTPS_PROXY=http://proxy:port\n")
				fmt.Fprintf(os.Stderr, "       - Or use: export ZTICTL_HTTPS_PROXY=http://proxy:port\n")
				fmt.Fprintf(os.Stderr, "       - For corporate proxies with custom certs: export ZTICTL_INSECURE_SKIP_VERIFY=true\n")
				fmt.Fprintf(os.Stderr, "       - Authenticate to GitHub: export GITHUB_TOKEN=your_token\n")
			}
		}
		return
	}

	if isOutdated {
		fmt.Printf("\n[WARN] Update Available: %s -> %s\n", currentVersion, latestVersion)
		fmt.Printf("[INFO] Download: %s\n", installDocsURL)
	} else {
		// Only show "up to date" message in verbose mode
		if os.Getenv("ZTICTL_VERBOSE_VERSION") == "true" {
			fmt.Printf("\n[OK] You are using the latest version!\n")
		}
	}
}
