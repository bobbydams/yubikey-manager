package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				PrimaryKeyID:          "ABC123DEF4567890",
				PrimaryKeyFingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12",
				UserName:              "Test User",
				UserEmail:             "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing primary key ID",
			config: &Config{
				PrimaryKeyFingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12",
				UserName:              "Test User",
				UserEmail:             "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing fingerprint",
			config: &Config{
				PrimaryKeyID: "ABC123DEF4567890",
				UserName:     "Test User",
				UserEmail:    "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing user name",
			config: &Config{
				PrimaryKeyID:          "ABC123DEF4567890",
				PrimaryKeyFingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12",
				UserEmail:             "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "missing user email",
			config: &Config{
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

func TestLoad_WithConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `primary_key_id: "ABC123DEF4567890"
primary_key_fingerprint: "ABCDEF1234567890ABCDEF1234567890ABCDEF12"
user_name: "Test User"
user_email: "test@example.com"
keyserver: "hkps://keys.openpgp.org"
`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set up viper to use the temp directory
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	os.Setenv("HOME", tmpDir)
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(tmpDir)

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "ABC123DEF4567890", cfg.PrimaryKeyID)
	assert.Equal(t, "ABCDEF1234567890ABCDEF1234567890ABCDEF12", cfg.PrimaryKeyFingerprint)
	assert.Equal(t, "Test User", cfg.UserName)
	assert.Equal(t, "test@example.com", cfg.UserEmail)
}
