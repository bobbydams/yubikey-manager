package cli

import (
	"fmt"
	"os"

	"github.com/bobbydams/yubikey-manager/internal/executor"
	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Add a signing subkey to a new YubiKey (interactive)",
		Long: `Setup a new YubiKey with a signing subkey. This command guides you through
the interactive process of generating a new subkey and moving it to your YubiKey.`,
		RunE: runSetup,
	}
}

func runSetup(cmd *cobra.Command, args []string) error {
	gpgSvc, yubikeySvc, backupSvc := getServices()
	ctx := cmd.Context()

	ui.PrintHeader("Setup New YubiKey for Signing")

	// Check YubiKey presence
	present, err := yubikeySvc.IsPresent(ctx)
	if err != nil {
		return fmt.Errorf("failed to check YubiKey: %w", err)
	}
	if !present {
		ui.LogError("No YubiKey detected. Please insert a YubiKey and try again.")
		return fmt.Errorf("no YubiKey detected")
	}

	cardInfo, err := yubikeySvc.GetCardInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get card info: %w", err)
	}

	ui.LogInfo("Detected YubiKey with serial: %s", cardInfo.Serial)

	// Check if YubiKey already has a signing key
	if sigKey, ok := cardInfo.Keys["Signature"]; ok && sigKey != "" && sigKey != "[none]" {
		ui.LogWarning("This YubiKey already has a signature key configured: %s", sigKey)
		if !ui.Confirm("Continue anyway? This will add another signing subkey.") {
			return nil
		}
	}

	// Create backup
	ui.LogInfo("Creating backup before making changes...")
	backupPath, err := backupSvc.CreateBackup(ctx, cfg.PrimaryKeyID, cfg.BackupDir)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	ui.LogSuccess("Backup created at %s", backupPath)

	// Get master key
	masterKeyPath := cfg.MasterKeyPath
	if masterKeyPath == "" {
		fmt.Println()
		fmt.Println("Please enter the path to your master secret key backup.")
		fmt.Println("This is typically on a USB drive, e.g.:")
		fmt.Println("  /Volumes/USB_DRIVE/Your Name - yourdomain.com (YOUR_KEY_ID) – Secret")
		fmt.Println()

		var err error
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
	// Import using gpg
	_, err = exec.Run(ctx, "gpg", "--import", masterKeyPath)
	if err != nil {
		return fmt.Errorf("failed to import master key: %w", err)
	}
	ui.LogSuccess("Master key imported")

	// Verify master key is available
	keys, err := gpgSvc.ListSecretKeys(ctx, cfg.PrimaryKeyID)
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	hasMaster := false
	for _, key := range keys {
		if key.Type == "sec" && key.KeyID == cfg.PrimaryKeyID {
			hasMaster = true
			break
		}
	}

	if !hasMaster {
		return fmt.Errorf("master key still shows as unavailable. Import may have failed")
	}

	// Interactive subkey generation
	fmt.Println()
	ui.LogInfo("Generating new signing subkey...")
	fmt.Println()
	fmt.Println("Now we need to generate a new signing subkey. Follow these steps:")
	fmt.Println()
	fmt.Println("1. Run: gpg --edit-key", cfg.PrimaryKeyID)
	fmt.Println("2. At the gpg> prompt, type: addkey")
	fmt.Println("3. Select: (10) ECC (sign only)")
	fmt.Println("4. Select: (1) Curve 25519")
	fmt.Println("5. For expiration, enter: 5y")
	fmt.Println("6. Confirm the creation")
	fmt.Println("7. Type: save")
	fmt.Println()

	response, err := ui.Prompt("Press Enter when ready to run gpg --edit-key, or 'q' to quit: ")
	if err != nil {
		return err
	}
	if response == "q" {
		if err := removeMasterKey(ctx, gpgSvc, cfg.PrimaryKeyFingerprint); err != nil {
			return fmt.Errorf("failed to remove master key: %w", err)
		}
		return nil
	}

	if err := gpgSvc.EditKey(ctx, cfg.PrimaryKeyID); err != nil {
		return fmt.Errorf("failed to edit key: %w", err)
	}

	// Move subkey to YubiKey
	fmt.Println()
	ui.LogInfo("Now we'll move the new subkey to your YubiKey.")
	fmt.Println()
	fmt.Println("Steps to move the subkey to YubiKey:")
	fmt.Println()
	fmt.Println("1. Run: gpg --edit-key", cfg.PrimaryKeyID)
	fmt.Println("2. Type: list (to see all subkeys with numbers)")
	fmt.Println("3. Identify the NEW signing subkey (the one without a card-no)")
	fmt.Println("4. Type: key N (where N is the number of the new subkey)")
	fmt.Println("5. Type: keytocard")
	fmt.Println("6. Select: (1) Signature key")
	fmt.Println("7. Type: save")
	fmt.Println()

	_, err = ui.Prompt("Press Enter when ready to continue: ")
	if err != nil {
		return err
	}

	if err := gpgSvc.EditKey(ctx, cfg.PrimaryKeyID); err != nil {
		return fmt.Errorf("failed to edit key: %w", err)
	}

	// Clean up master key
	fmt.Println()
	if ui.Confirm("Remove master key from local machine?") {
		if err := removeMasterKey(ctx, gpgSvc, cfg.PrimaryKeyFingerprint); err != nil {
			ui.LogWarning("Failed to remove master key: %v", err)
		} else {
			ui.LogSuccess("Master key removed from local keyring")
		}
	} else {
		ui.LogWarning("Master key left on machine. Remember to remove it manually!")
	}

	// Upload to keyserver
	if ui.Confirm(fmt.Sprintf("Upload updated public key to %s?", cfg.Keyserver)) {
		ui.LogInfo("Uploading to keyserver...")
		exec := executor.NewRealExecutor()
		_, err := exec.Run(ctx, "gpg", "--keyserver", cfg.Keyserver, "--send-keys", cfg.PrimaryKeyID)
		if err != nil {
			ui.LogWarning("Failed to upload to keyserver: %v", err)
			ui.LogWarning("Visit https://keys.openpgp.org/upload to upload manually.")
		} else {
			ui.LogSuccess("Public key uploaded to %s", cfg.Keyserver)
		}
	}

	fmt.Println()
	ui.LogSuccess("YubiKey setup complete!")
	ui.LogInfo("Serial: %s", cardInfo.Serial)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Label this YubiKey physically (e.g., 'Key B - " + cardInfo.Serial + "')")
	fmt.Println("  2. Test signing: echo 'test' | gpg --sign --armor")
	fmt.Println("  3. Register this YubiKey with GitHub/GitLab if not already done")
	fmt.Println()

	return nil
}
