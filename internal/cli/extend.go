package cli

import (
	"fmt"
	"os"

	"github.com/bobbydams/yubikey-manager/internal/executor"
	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
)

func newExtendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "extend",
		Short: "Extend expiration dates on keys",
		RunE:  runExtend,
	}
}

func runExtend(cmd *cobra.Command, args []string) error {
	gpgSvc, _, backupSvc := getServices()
	ctx := cmd.Context()

	ui.PrintHeader("Extend Key Expiration")

	// Show current expiration
	fmt.Println("Current key expiration status:")
	keys, err := gpgSvc.ListSecretKeys(ctx, cfg.PrimaryKeyID)
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	for _, key := range keys {
		fmt.Printf("  %s %s", key.Type, key.KeyID)
		if key.Expires != "" {
			fmt.Printf(" expires: %s", key.Expires)
		}
		fmt.Println()
	}
	fmt.Println()

	newExpiry, err := ui.Prompt("Enter new expiration (e.g., '5y' for 5 years, '2035-01-01' for specific date): ")
	if err != nil {
		return err
	}
	if newExpiry == "" {
		return fmt.Errorf("no expiration provided")
	}

	// Create backup
	backupPath, err := backupSvc.CreateBackup(ctx, cfg.PrimaryKeyID, cfg.BackupDir)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	ui.LogSuccess("Backup created at %s", backupPath)

	// Get master key
	masterKeyPath := cfg.MasterKeyPath
	if masterKeyPath == "" {
		masterKeyPath, err = ui.PromptRequired("Master key path: ")
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(masterKeyPath); err != nil {
		return fmt.Errorf("master key file not found: %w", err)
	}

	// Import master key
	ui.LogInfo("Importing master key...")
	exec := executor.NewRealExecutor()
	_, err = exec.Run(ctx, "gpg", "--import", masterKeyPath)
	if err != nil {
		return fmt.Errorf("failed to import master key: %w", err)
	}
	ui.LogSuccess("Master key imported")

	// Interactive expiration extension
	fmt.Println()
	fmt.Println("To extend expiration:")
	fmt.Println()
	fmt.Println("1. First, extend the PRIMARY key:")
	fmt.Println("   - Type: expire")
	fmt.Printf("   - Enter: %s\n", newExpiry)
	fmt.Println()
	fmt.Println("2. Then extend EACH subkey:")
	fmt.Println("   - Type: key 1")
	fmt.Println("   - Type: expire")
	fmt.Printf("   - Enter: %s\n", newExpiry)
	fmt.Println("   - Type: key 1 (to deselect)")
	fmt.Println("   - Repeat for key 2, key 3, etc.")
	fmt.Println()
	fmt.Println("3. Type: save")
	fmt.Println()

	_, err = ui.Prompt("Press Enter to continue: ")
	if err != nil {
		return err
	}

	if err := gpgSvc.EditKey(ctx, cfg.PrimaryKeyID); err != nil {
		return fmt.Errorf("failed to edit key: %w", err)
	}

	// Clean up
	if err := removeMasterKey(ctx, gpgSvc, cfg.PrimaryKeyFingerprint); err != nil {
		ui.LogWarning("Failed to remove master key: %v", err)
	}

	// Upload
	if ui.Confirm(fmt.Sprintf("Upload updated public key to %s?", cfg.Keyserver)) {
		ui.LogInfo("Uploading to keyserver...")
		_, err := exec.Run(ctx, "gpg", "--keyserver", cfg.Keyserver, "--send-keys", cfg.PrimaryKeyID)
		if err != nil {
			ui.LogWarning("Failed to upload to keyserver: %v", err)
		} else {
			ui.LogSuccess("Public key uploaded to %s", cfg.Keyserver)
		}
	}

	fmt.Println()
	ui.LogSuccess("Key expiration extended")

	// Show new status
	fmt.Println()
	fmt.Println("Updated expiration status:")
	keys, err = gpgSvc.ListSecretKeys(ctx, cfg.PrimaryKeyID)
	if err == nil {
		for _, key := range keys {
			fmt.Printf("  %s %s", key.Type, key.KeyID)
			if key.Expires != "" {
				fmt.Printf(" expires: %s", key.Expires)
			}
			fmt.Println()
		}
	}

	return nil
}
