package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/bobbydams/yubikey-manager/internal/executor"
	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
)

func newSetupBatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup-batch",
		Short: "Add a signing subkey to a new YubiKey (semi-automated)",
		Long: `Setup a new YubiKey with a signing subkey using semi-automated mode.
This command creates the subkey automatically but still requires interaction
to move it to the YubiKey.`,
		RunE: runSetupBatch,
	}
}

func runSetupBatch(cmd *cobra.Command, args []string) error {
	gpgSvc, yubikeySvc, backupSvc := getServices()
	ctx := cmd.Context()

	ui.PrintHeader("Setup New YubiKey (Automated Mode)")

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

	// Create backup
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

	// Generate new signing subkey
	ui.LogInfo("Generating new ed25519 signing subkey...")

	expiryDate := time.Now().AddDate(5, 0, 0).Format("2006-01-02")
	_, err = exec.Run(ctx, "gpg", "--batch", "--passphrase-fd", "0", "--quick-add-key",
		cfg.PrimaryKeyFingerprint, "ed25519", "sign", expiryDate)
	if err != nil {
		return fmt.Errorf("failed to create subkey: %w", err)
	}

	ui.LogSuccess("New signing subkey created")

	// Move subkey to YubiKey (interactive)
	fmt.Println()
	ui.LogInfo("Moving new subkey to YubiKey...")
	fmt.Println()
	fmt.Println("The new subkey has been created. Now we need to move it to the YubiKey.")
	fmt.Println("GPG requires interaction for this step.")
	fmt.Println()
	fmt.Println("1. In the gpg prompt, type: list")
	fmt.Println("2. Find the newest [S] subkey (without a card-no line after it)")
	fmt.Println("3. Type: key N (where N is that subkey's number, probably 4 or 5)")
	fmt.Println("4. Type: keytocard")
	fmt.Println("5. Select: (1) Signature key")
	fmt.Println("6. Enter your PIN when prompted")
	fmt.Println("7. Type: save")
	fmt.Println()

	_, err = ui.Prompt("Press Enter to continue: ")
	if err != nil {
		return err
	}

	if err := gpgSvc.EditKey(ctx, cfg.PrimaryKeyID); err != nil {
		return fmt.Errorf("failed to edit key: %w", err)
	}

	// Clean up
	if ui.Confirm("Remove master key from local machine?") {
		if err := removeMasterKey(ctx, gpgSvc, cfg.PrimaryKeyFingerprint); err != nil {
			ui.LogWarning("Failed to remove master key: %v", err)
		}
	}

	// Upload to keyserver
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
	ui.LogSuccess("Setup complete for YubiKey %s", cardInfo.Serial)

	return nil
}
