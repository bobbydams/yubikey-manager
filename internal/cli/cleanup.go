package cli

import (
	"fmt"

	"github.com/bobbydams/yubikey-manager/internal/executor"
	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
)

func newCleanupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Remove old/expired keys from keyring",
		RunE:  runCleanup,
	}
}

func runCleanup(cmd *cobra.Command, args []string) error {
	gpgSvc, _, _ := getServices()
	ctx := cmd.Context()

	ui.PrintHeader("Cleanup Old Keys")

	fmt.Println("Current keys in keyring:")
	fmt.Println()

	// List all keys (we'll need to list without a specific key ID)
	exec := executor.NewRealExecutor()
	output, err := exec.Run(ctx, "gpg", "--list-secret-keys", "--keyid-format=long")
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}
	fmt.Println(string(output))
	fmt.Println()

	// Identify potentially removable keys
	fmt.Println("Keys that might be candidates for removal:")
	fmt.Println()

	// Check for expired keys
	keys, err := gpgSvc.ListSecretKeys(ctx, cfg.PrimaryKeyID)
	if err == nil {
		for _, key := range keys {
			_ = key // In a real implementation, we'd parse and check the expiration date
			// For now, just iterate through keys
		}
	}

	// Keys not matching primary
	fmt.Println("Keys other than your primary (" + cfg.PrimaryKeyID + "):")
	// This would require parsing all keys, which is more complex
	// For now, we'll just show instructions
	fmt.Println("  (use gpg --list-secret-keys to see all keys)")
	fmt.Println()

	fmt.Println("To delete a key:")
	fmt.Println("  gpg --delete-secret-keys <KEY_ID>")
	fmt.Println("  gpg --delete-keys <KEY_ID>")
	fmt.Println()

	if ui.Confirm("Would you like to interactively delete keys?") {
		for {
			keyToDelete, err := ui.Prompt("Enter KEY ID to delete (or 'q' to quit): ")
			if err != nil {
				return err
			}
			if keyToDelete == "q" {
				break
			}

			if keyToDelete == cfg.PrimaryKeyID {
				ui.LogError("Cannot delete primary key!")
				continue
			}

			if ui.Confirm(fmt.Sprintf("Delete %s?", keyToDelete)) {
				// Delete secret key
				_, err := exec.Run(ctx, "gpg", "--batch", "--yes", "--delete-secret-keys", keyToDelete)
				if err != nil {
					ui.LogWarning("Failed to delete secret key: %v", err)
				}

				// Delete public key
				_, err = exec.Run(ctx, "gpg", "--batch", "--yes", "--delete-keys", keyToDelete)
				if err != nil {
					ui.LogWarning("Failed to delete public key: %v", err)
				} else {
					ui.LogSuccess("Deleted %s", keyToDelete)
				}
			}
		}
	}

	// Clean up trust database
	if ui.Confirm("Clean up trust database?") {
		if err := gpgSvc.CheckTrustDB(ctx); err != nil {
			ui.LogWarning("Failed to check trustdb: %v", err)
		} else {
			ui.LogSuccess("Trust database cleaned")
		}
	}

	return nil
}
