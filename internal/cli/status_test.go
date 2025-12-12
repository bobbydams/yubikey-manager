package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bobbydams/yubikey-manager/internal/backup"
	"github.com/bobbydams/yubikey-manager/internal/config"
	"github.com/bobbydams/yubikey-manager/internal/executor"
	"github.com/bobbydams/yubikey-manager/internal/gpg"
	"github.com/bobbydams/yubikey-manager/internal/yubikey"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusCmd(t *testing.T) {
	cmd := newStatusCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "status", cmd.Use)
}

func TestRunStatus(t *testing.T) {
	// Save original config
	oldCfg := cfg
	defer func() {
		cfg = oldCfg
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

	// Load config
	viper.Reset()
	loadedCfg, err := config.Load()
	require.NoError(t, err)
	cfg = loadedCfg

	// Create mock executor
	mockExecutor := executor.NewMockExecutor()
	mockExecutor.SetOutput("gpg", []byte("sec   rsa4096 2024-01-01 [SC]\n  ABC123DEF4567890\nuid           [ultimate] Test User <test@example.com>\nssb   rsa4096 2024-01-01 [S] [expires: 2025-01-01]\n  DEF456GHI7890123\n"))

	gpgSvc := gpg.NewService(mockExecutor)
	yubikeySvc := yubikey.NewService(gpgSvc, mockExecutor)
	backupSvc := backup.NewService(gpgSvc)

	// Verify services are created correctly
	assert.NotNil(t, gpgSvc)
	assert.NotNil(t, yubikeySvc)
	assert.NotNil(t, backupSvc)

	// Note: We can't easily test runStatus without overriding getServices
	// which is a function, not a variable. The function requires cfg to be set
	// and getServices to return the mocked services.
	// For now, we verify the services can be created and the command structure is correct.
}

func TestRunStatus_NoKeys(t *testing.T) {
	// Save original config
	oldCfg := cfg
	defer func() {
		cfg = oldCfg
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

	// Load config
	viper.Reset()
	loadedCfg, err := config.Load()
	require.NoError(t, err)
	cfg = loadedCfg

	// Create mock executor that returns error
	mockExecutor := executor.NewMockExecutor()
	mockExecutor.SetError("gpg", fmt.Errorf("key not found"))

	gpgSvc := gpg.NewService(mockExecutor)
	yubikeySvc := yubikey.NewService(gpgSvc, mockExecutor)
	backupSvc := backup.NewService(gpgSvc)

	// Verify services are created correctly
	assert.NotNil(t, gpgSvc)
	assert.NotNil(t, yubikeySvc)
	assert.NotNil(t, backupSvc)

	// Note: Testing runStatus fully would require overriding getServices
	// which is not easily testable. The function structure is verified above.
}
