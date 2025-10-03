package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	// Try to get from cache first
	cached, err := getFromCache()
	if err == nil && time.Since(cached.CheckedAt) < cacheExpiration {
		return compareVersions(currentVersion, cached.LatestVersion), cached.LatestVersion, nil
	}

	// Fetch from GitHub API
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(githubAPIURL)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return false, "", err
	}

	latestVersion = strings.TrimPrefix(release.TagName, "v")

	// Cache the result
	_ = saveToCache(latestVersion)

	return compareVersions(currentVersion, latestVersion), latestVersion, nil
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
		// Silently ignore errors - don't interrupt version output
		return
	}

	if isOutdated {
		fmt.Printf("\nYour version of ztictl is out of date! The latest version\n")
		fmt.Printf("is %s. You can update by downloading from %s\n", latestVersion, installDocsURL)
	}
}
