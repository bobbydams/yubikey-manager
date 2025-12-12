package cli

import (
	"fmt"

	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
)

func newMetadataCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "set-metadata",
		Aliases: []string{"metadata"},
		Short:   "Set cardholder name and URL on YubiKey",
		RunE:    runMetadata,
	}
}

func runMetadata(cmd *cobra.Command, args []string) error {
	_, yubikeySvc, _ := getServices()
	ctx := cmd.Context()

	ui.PrintHeader("Set YubiKey Card Metadata")

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

	ui.LogInfo("Configuring YubiKey with serial: %s", cardInfo.Serial)

	fmt.Println()
	fmt.Println("This will set the cardholder name and other metadata on your YubiKey.")
	fmt.Println("This helps identify which YubiKey is which.")
	fmt.Println()
	fmt.Println("In the gpg prompt:")
	fmt.Println("1. Type: admin")
	fmt.Println("2. Type: name (then enter surname, then given name)")
	fmt.Println("3. Type: lang (then enter 'en')")
	fmt.Printf("4. Type: url (then enter: https://keys.openpgp.org/vks/v1/by-fingerprint/%s)\n", cfg.PrimaryKeyFingerprint)
	fmt.Println("5. Type: quit")
	fmt.Println()

	_, err = ui.Prompt("Press Enter to continue: ")
	if err != nil {
		return err
	}

	if err := yubikeySvc.EditCard(ctx); err != nil {
		return fmt.Errorf("failed to edit card: %w", err)
	}

	ui.LogSuccess("YubiKey metadata updated")

	return nil
}
