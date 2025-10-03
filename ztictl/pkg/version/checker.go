package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	// Strip git hash from current version (e.g., "2.8.0-abaf1976" -> "2.8.0")
	if idx := strings.Index(current, "-"); idx != -1 {
		current = current[:idx]
	}

	// Simple string comparison for semantic versions
	return current < latest
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
