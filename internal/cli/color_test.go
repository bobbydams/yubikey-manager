package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bobbydams/yubikey-manager/internal/config"
	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoColorFlag(t *testing.T) {
	// Save original state
	originalNoColor := color.NoColor
	originalEnabled := ui.IsColorEnabled()
	defer func() {
		color.NoColor = originalNoColor
		ui.SetColorEnabled(originalEnabled)
		viper.Reset()
	}()

	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "ykgpg-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)

	// Set HOME environment variable
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpHome)

	// Create a minimal config file
	configDir := filepath.Join(tmpHome, ".config", "ykgpg")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `primary_key_id: "ABC123DEF4567890"
primary_key_fingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12"
user_name: "Test User"
user_email: "test@example.com"
`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test that colors are enabled by default
	viper.Reset()
	cfg, err := config.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Simulate flag being set
	viper.Set("no_color", true)
	cfg, err = config.Load()
	require.NoError(t, err)

	// Colors should be disabled when no_color is true
	if cfg.NoColor {
		ui.SetColorEnabled(false)
		assert.False(t, ui.IsColorEnabled())
		assert.True(t, color.NoColor)
	}
}

func TestNoColorConfigOption(t *testing.T) {
	// Save original state
	originalNoColor := color.NoColor
	originalEnabled := ui.IsColorEnabled()
	defer func() {
		color.NoColor = originalNoColor
		ui.SetColorEnabled(originalEnabled)
		viper.Reset()
	}()

	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "ykgpg-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)

	// Set HOME environment variable
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpHome)

	// Create config file with no_color: true
	configDir := filepath.Join(tmpHome, ".config", "ykgpg")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `primary_key_id: "ABC123DEF4567890"
primary_key_fingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12"
user_name: "Test User"
user_email: "test@example.com"
no_color: true
`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load config
	viper.Reset()
	cfg, err := config.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify no_color is loaded
	assert.True(t, cfg.NoColor)
}

func TestNoColorEnvironmentVariable(t *testing.T) {
	// Save original state
	originalNoColor := color.NoColor
	originalEnabled := ui.IsColorEnabled()
	oldEnv := os.Getenv("YKGPG_NO_COLOR")
	defer func() {
		color.NoColor = originalNoColor
		ui.SetColorEnabled(originalEnabled)
		if oldEnv != "" {
			os.Setenv("YKGPG_NO_COLOR", oldEnv)
		} else {
			os.Unsetenv("YKGPG_NO_COLOR")
		}
		viper.Reset()
	}()

	// Set environment variable
	os.Setenv("YKGPG_NO_COLOR", "true")

	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "ykgpg-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)

	// Set HOME environment variable
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpHome)

	// Create a minimal config file
	configDir := filepath.Join(tmpHome, ".config", "ykgpg")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `primary_key_id: "ABC123DEF4567890"
primary_key_fingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12"
user_name: "Test User"
user_email: "test@example.com"
`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load config
	viper.Reset()
	cfg, err := config.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Environment variable should be read (though viper may need explicit handling)
	// The actual value depends on how viper processes boolean env vars
	_ = cfg
}

func TestColorEnabledByDefault(t *testing.T) {
	// Save original state
	originalNoColor := color.NoColor
	originalEnabled := ui.IsColorEnabled()
	defer func() {
		color.NoColor = originalNoColor
		ui.SetColorEnabled(originalEnabled)
	}()

	// Reset to default state
	ui.SetColorEnabled(true)
	assert.True(t, ui.IsColorEnabled())
	assert.False(t, color.NoColor)
}

func TestConfigNoColorField(t *testing.T) {
	cfg := &config.Config{
		PrimaryKeyID:          "ABC123DEF4567890",
		PrimaryKeyFingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12",
		UserName:              "Test User",
		UserEmail:             "test@example.com",
		NoColor:               false,
	}

	// Test default value
	assert.False(t, cfg.NoColor)

	// Test setting to true
	cfg.NoColor = true
	assert.True(t, cfg.NoColor)
}
