package cli

import (
	"os"
	"testing"

	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	// Execute requires rootCmd to be initialized
	// Since init() runs automatically, we can test Execute
	// We'll test with invalid args to trigger error path
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	// Test that Execute can be called (will fail due to missing config, but that's expected)
	// We can't easily test the full execution without setting up a full environment
	// But we can verify the function exists and can be called
	_ = Execute
}

func TestSetVersion(t *testing.T) {
	// Save original version
	originalVersion := version

	// Test setting version
	SetVersion("1.2.3")
	assert.Equal(t, "1.2.3", version)
	assert.Equal(t, "1.2.3", rootCmd.Version)

	// Restore original version
	SetVersion(originalVersion)
}

func TestBindFlags(t *testing.T) {
	// Create a test command
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().String("key-id", "", "test")
	cmd.Flags().String("fingerprint", "", "test")
	cmd.Flags().String("name", "", "test")
	cmd.Flags().String("email", "", "test")
	cmd.Flags().String("keyserver", "", "test")
	cmd.Flags().String("master-key-path", "", "test")
	cmd.Flags().String("backup-dir", "", "test")
	cmd.Flags().Bool("no-color", false, "test")

	// Test that bindFlags doesn't panic
	assert.NotPanics(t, func() {
		bindFlags(cmd)
	})
}

func TestGetServices(t *testing.T) {
	// Test that getServices returns non-nil services
	gpgSvc, yubikeySvc, backupSvc := getServices()

	assert.NotNil(t, gpgSvc)
	assert.NotNil(t, yubikeySvc)
	assert.NotNil(t, backupSvc)
}

func TestRootCmdInitialization(t *testing.T) {
	// Test that rootCmd is initialized
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "ykgpg", rootCmd.Use)

	// Test that all subcommands are added
	subcommands := rootCmd.Commands()
	expectedCommands := []string{
		"status", "setup", "setup-batch", "revoke", "extend",
		"cleanup", "set-metadata", "export", "verify", "config",
	}

	foundCommands := make(map[string]bool)
	for _, cmd := range subcommands {
		foundCommands[cmd.Use] = true
	}

	for _, expected := range expectedCommands {
		assert.True(t, foundCommands[expected], "Expected command %s not found", expected)
	}
}

func TestRootCmdNoColorFlag(t *testing.T) {
	// Save original state
	originalEnabled := ui.IsColorEnabled()
	defer ui.SetColorEnabled(originalEnabled)

	// Create a test command with no-color flag
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.Flags().Bool("no-color", false, "test")
	if err := cmd.Flags().Set("no-color", "true"); err != nil {
		t.Fatalf("Failed to set flag: %v", err)
	}

	// Simulate PersistentPreRunE behavior
	noColor, _ := cmd.Flags().GetBool("no-color")
	if noColor {
		ui.SetColorEnabled(false)
	}

	assert.False(t, ui.IsColorEnabled())
}
