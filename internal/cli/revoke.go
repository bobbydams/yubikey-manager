package cli

import (
	"fmt"
	"os"

	"github.com/bobbydams/yubikey-manager/internal/executor"
	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
)

func newRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a subkey (for lost/compromised YubiKeys)",
		Long: `Revoke a signing subkey, typically because a YubiKey was lost or compromised.
This action CANNOT be undone!`,
		RunE: runRevoke,
	}
}

func runRevoke(cmd *cobra.Command, args []string) error {
	gpgSvc, _, backupSvc := getServices()
	ctx := cmd.Context()

	ui.PrintHeader("Revoke Subkey (Lost/Compromised)")

	ui.LogWarning("This will revoke a signing subkey, typically because a YubiKey was lost or compromised.")
	ui.LogWarning("This action CANNOT be undone!")
	fmt.Println()

	// Show current subkeys
	fmt.Println("Current signing subkeys:")
	fmt.Println()
	keys, err := gpgSvc.ListSecretKeys(ctx, cfg.PrimaryKeyID)
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	for _, key := range keys {
		if contains(key.Capabilities, "S") {
			fmt.Printf("  %s %s", key.Type, key.KeyID)
			if key.CardNo != "" {
				fmt.Printf(" card-no: %s", key.CardNo)
			}
			fmt.Println()
		}
	}
	fmt.Println()

	fmt.Println("Identify the subkey to revoke by its key ID (the hex string after ed25519/).")
	fmt.Println("If the YubiKey is lost, you can identify it by the card serial number.")
	fmt.Println()

	keyToRevoke, err := ui.Prompt("Enter the KEY ID to revoke (or 'q' to quit): ")
	if err != nil {
		return err
	}
	if keyToRevoke == "q" {
		return nil
	}

	// Verify key exists
	found := false
	for _, key := range keys {
		if key.KeyID == keyToRevoke || key.Fingerprint == keyToRevoke {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("key ID not found: %s", keyToRevoke)
	}

	if !ui.Confirm(fmt.Sprintf("Are you SURE you want to revoke key %s? This cannot be undone!", keyToRevoke)) {
		return nil
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

	// Interactive revocation
	fmt.Println()
	fmt.Println("To revoke the subkey:")
	fmt.Println()
	fmt.Println("1. In the gpg prompt, type: list")
	fmt.Println("2. Find the subkey matching: " + keyToRevoke)
	fmt.Println("3. Type: key N (where N is that subkey's number)")
	fmt.Println("4. Type: revkey")
	fmt.Println("5. Select reason: (1) Key has been compromised -OR- (2) Key is superseded")
	fmt.Println("6. Enter a description if desired")
	fmt.Println("7. Confirm the revocation")
	fmt.Println("8. Type: save")
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

	// Upload revocation
	ui.LogWarning("IMPORTANT: You must upload the updated key to propagate the revocation!")
	if ui.Confirm(fmt.Sprintf("Upload updated public key to %s?", cfg.Keyserver)) {
		ui.LogInfo("Uploading to keyserver...")
		_, err := exec.Run(ctx, "gpg", "--keyserver", cfg.Keyserver, "--send-keys", cfg.PrimaryKeyID)
		if err != nil {
			ui.LogWarning("Failed to upload to keyserver: %v", err)
			ui.LogWarning("Visit https://keys.openpgp.org/upload to upload manually.")
		} else {
			ui.LogSuccess("Public key uploaded to %s", cfg.Keyserver)
		}
	}

	fmt.Println()
	ui.LogSuccess("Subkey revoked. The revocation has been published.")
	fmt.Println()
	fmt.Println("Additional steps:")
	fmt.Println("  1. Remove the revoked key from GitHub/GitLab if it was registered there")
	fmt.Println("  2. Update any systems that had the old key configured")
	fmt.Println()

	return nil
}
