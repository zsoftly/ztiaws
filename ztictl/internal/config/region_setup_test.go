package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseAndNormalizeRegions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Shortcodes only",
			input:    "cac1,use1,euw1",
			expected: []string{"cac1", "use1", "euw1"},
		},
		{
			name:     "Full region names",
			input:    "ca-central-1,us-east-1,eu-west-1",
			expected: []string{"cac1", "use1", "euw1"},
		},
		{
			name:     "Mixed shortcodes and full names",
			input:    "cac1,us-east-1,euw1",
			expected: []string{"cac1", "use1", "euw1"},
		},
		{
			name:     "With spaces",
			input:    "cac1, use1 , euw1",
			expected: []string{"cac1", "use1", "euw1"},
		},
		{
			name:     "Duplicates removed",
			input:    "cac1,ca-central-1,cac1",
			expected: []string{"cac1"},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: nil, // parseAndNormalizeRegions returns nil for empty input
		},
		{
			name:     "Invalid regions skipped",
			input:    "cac1,invalid-region,use1",
			expected: []string{"cac1", "use1"},
		},
		{
			name:     "New AWS region format accepted",
			input:    "ap-south-2,me-central-1",
			expected: []string{"aps2", "mec1"}, // both are known regions
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAndNormalizeRegions(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeRegion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Shortcode unchanged",
			input:    "cac1",
			expected: "cac1",
		},
		{
			name:     "Full region to shortcode",
			input:    "ca-central-1",
			expected: "cac1",
		},
		{
			name:     "Unknown full region unchanged",
			input:    "il-central-1",
			expected: "il-central-1",
		},
		{
			name:     "Invalid format unchanged",
			input:    "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeRegion(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveRegionInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Shortcode to full region",
			input:    "cac1",
			expected: "ca-central-1",
		},
		{
			name:     "Full region unchanged",
			input:    "ca-central-1",
			expected: "ca-central-1",
		},
		{
			name:     "Unknown but valid format",
			input:    "il-central-1",
			expected: "il-central-1",
		},
		{
			name:     "Invalid format unchanged",
			input:    "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveRegionInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSaveRegionConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ztictl-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Override HOME for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	configPath := filepath.Join(tmpDir, ".ztictl.yaml")

	t.Run("Create new config", func(t *testing.T) {
		enabledRegions := []string{"cac1", "use1"}
		regionGroups := map[string][]string{
			"all":  {"cac1", "use1"},
			"prod": {"use1"},
		}

		// Save config (can't test fully without mocking Load())
		// We'll test the file writing part
		saveConfigForTest(t, configPath, enabledRegions, regionGroups, nil)

		// Read and verify
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var config map[string]interface{}
		err = yaml.Unmarshal(data, &config)
		require.NoError(t, err)

		// Check regions section
		regions, ok := config["regions"].(map[string]interface{})
		require.True(t, ok)

		enabled, ok := regions["enabled"].([]interface{})
		require.True(t, ok)
		assert.Len(t, enabled, 2)
	})

	t.Run("Preserve existing config", func(t *testing.T) {
		// Create initial config with SSO settings
		existingConfig := map[string]interface{}{
			"sso": map[string]interface{}{
				"region":    "ca-central-1",
				"start_url": "https://test.com/start",
			},
			"default_region": "us-east-1",
			"logging": map[string]interface{}{
				"directory": "/home/user/logs",
				"level":     "info",
			},
		}

		// Write initial config
		data, err := yaml.Marshal(existingConfig)
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0600)
		require.NoError(t, err)

		// Update with regions
		enabledRegions := []string{"cac1", "use1"}
		regionGroups := map[string][]string{
			"all": {"cac1", "use1"},
		}

		saveConfigForTest(t, configPath, enabledRegions, regionGroups, existingConfig)

		// Read and verify preservation
		data, err = os.ReadFile(configPath)
		require.NoError(t, err)

		var config map[string]interface{}
		err = yaml.Unmarshal(data, &config)
		require.NoError(t, err)

		// Check SSO section is preserved
		sso, ok := config["sso"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "ca-central-1", sso["region"])
		assert.Equal(t, "https://test.com/start", sso["start_url"])

		// Check other settings preserved
		assert.Equal(t, "us-east-1", config["default_region"])

		// Check regions were added
		regions, ok := config["regions"].(map[string]interface{})
		require.True(t, ok)
		assert.NotNil(t, regions["enabled"])
	})
}

// Helper function to simulate saveRegionConfig without Load()
func saveConfigForTest(t *testing.T, configPath string, enabledRegions []string, regionGroups map[string][]string, existingData map[string]interface{}) {
	var existingConfig map[string]interface{}

	if existingData != nil {
		existingConfig = existingData
	} else if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, &existingConfig); err != nil {
			existingConfig = make(map[string]interface{})
		}
	} else {
		existingConfig = make(map[string]interface{})
	}

	if existingConfig == nil {
		existingConfig = make(map[string]interface{})
	}

	existingConfig["regions"] = map[string]interface{}{
		"enabled": enabledRegions,
		"groups":  regionGroups,
	}

	data, err := yaml.Marshal(existingConfig)
	require.NoError(t, err)

	err = os.WriteFile(configPath, data, 0600)
	require.NoError(t, err)
}
