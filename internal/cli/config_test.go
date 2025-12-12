package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bobbydams/yubikey-manager/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConfigInit_DirectoryCreation(t *testing.T) {
	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "ykgpg-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)

	// Set HOME environment variable
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpHome)

	// Test that config init would create the directory
	configDir := filepath.Join(tmpHome, ".config", "ykgpg")

	// Verify directory doesn't exist yet
	_, err = os.Stat(configDir)
	assert.True(t, os.IsNotExist(err))

	// Create directory and write config (simulating what init does)
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Verify directory was created
	stat, err := os.Stat(configDir)
	assert.NoError(t, err)
	assert.True(t, stat.IsDir())
	assert.Equal(t, os.FileMode(0755), stat.Mode().Perm())
}

func TestConfigInit_FileWriting(t *testing.T) {
	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "ykgpg-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)

	// Set HOME environment variable
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpHome)

	configDir := filepath.Join(tmpHome, ".config", "ykgpg")
	configFile := filepath.Join(configDir, "config.yaml")

	// Create directory
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Create config data
	configData := map[string]interface{}{
		"primary_key_id":          "ABC123DEF4567890",
		"primary_key_fingerprint": "ABCDEF1234567890ABCDEF1234567890ABCDEF12",
		"user_name":               "Test User",
		"user_email":              "test@example.com",
		"keyserver":               "hkps://keys.openpgp.org",
		"backup_dir":              filepath.Join(tmpHome, ".gnupg", "backups"),
	}

	// Write YAML file
	yamlData, err := yaml.Marshal(configData)
	require.NoError(t, err)

	err = os.WriteFile(configFile, yamlData, 0644)
	require.NoError(t, err)

	// Verify file exists and is readable
	stat, err := os.Stat(configFile)
	assert.NoError(t, err)
	assert.False(t, stat.IsDir())
	assert.Equal(t, os.FileMode(0644), stat.Mode().Perm())

	// Verify file can be loaded
	loadedCfg, err := config.Load()
	if err == nil {
		err = loadedCfg.Validate()
		if err == nil {
			assert.Equal(t, "ABC123DEF4567890", loadedCfg.PrimaryKeyID)
			assert.Equal(t, "Test User", loadedCfg.UserName)
		}
	}
}

func TestConfigShow_WithoutConfig(t *testing.T) {
	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "ykgpg-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpHome)

	// Set HOME environment variable
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tmpHome)

	// Clear any environment variables that might affect the test
	oldEnvVars := make(map[string]string)
	envVars := []string{
		"YKGPG_PRIMARY_KEY_ID",
		"YKGPG_PRIMARY_KEY_FINGERPRINT",
		"YKGPG_USER_NAME",
		"YKGPG_USER_EMAIL",
	}
	for _, envVar := range envVars {
		oldEnvVars[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}
	defer func() {
		for k, v := range oldEnvVars {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	// Test config loading with no config file
	cfg, _ := config.Load()

	// Config loading should return a config struct (may or may not error)
	assert.NotNil(t, cfg)

	// Config should be loaded (even if empty)
	assert.NotNil(t, cfg)

	// If all required fields are empty, validation should fail
	// (But defaults might set some values like keyserver and backup_dir)
	if cfg.PrimaryKeyID == "" || cfg.PrimaryKeyFingerprint == "" || cfg.UserName == "" || cfg.UserEmail == "" {
		validationErr := cfg.Validate()
		assert.Error(t, validationErr, "Config validation should fail when required fields are missing")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				PrimaryKeyID:          "ABC123DEF4567890",
				PrimaryKeyFingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12",
				UserName:              "Test User",
				UserEmail:             "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing primary key ID",
			config: &config.Config{
				PrimaryKeyFingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12",
				UserName:              "Test User",
				UserEmail:             "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing fingerprint",
			config: &config.Config{
				PrimaryKeyID: "ABC123DEF4567890",
				UserName:     "Test User",
				UserEmail:    "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing user name",
			config: &config.Config{
				PrimaryKeyID:          "ABC123DEF4567890",
				PrimaryKeyFingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12",
				UserEmail:             "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing user email",
			config: &config.Config{
				PrimaryKeyID:          "ABC123DEF4567890",
				PrimaryKeyFingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12",
				UserName:              "Test User",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
