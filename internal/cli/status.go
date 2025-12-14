package cli

import (
	"fmt"

	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current key and YubiKey status",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	gpgSvc, yubikeySvc, _ := getServices()
	ctx := cmd.Context()

	ui.PrintHeader("YubiKey GPG Manager Status")

	// Primary key info
	ui.PrintSection("PRIMARY KEY")
	ui.PrintKeyValueKey("Key ID", cfg.PrimaryKeyID)
	ui.PrintKeyValue("User", fmt.Sprintf("%s <%s>", cfg.UserName, cfg.UserEmail))
	fmt.Println()

	// Check if primary key exists
	keys, err := gpgSvc.ListSecretKeys(ctx, cfg.PrimaryKeyID)
	if err != nil {
		ui.LogError("Primary key not found in keyring: %v", err)
		return err
	}

	if len(keys) == 0 {
		ui.LogError("Primary key not found in keyring!")
		return fmt.Errorf("primary key not found")
	}

	// Show key details
	ui.PrintSection("KEY DETAILS")
	for _, key := range keys {
		ui.PrintKey(key.Type + " ")
		ui.PrintKey(key.KeyID)
		// Format capabilities as [S C E A] instead of [S C E A]
		if len(key.Capabilities) > 0 {
			capStr := ""
			for i, cap := range key.Capabilities {
				if i > 0 {
					capStr += " "
				}
				capStr += cap
			}
			fmt.Printf(" [%s]", capStr)
		}
		if key.Expires != "" {
			fmt.Printf(" expires: ")
			ui.PrintValue(key.Expires)
		}
		if key.CardNo != "" {
			fmt.Printf(" card-no: ")
			ui.PrintValue(key.CardNo)
		}
		fmt.Println()
	}
	fmt.Println()

	// YubiKey status
	ui.PrintSection("YUBIKEY STATUS")
	present, err := yubikeySvc.IsPresent(ctx)
	if err != nil {
		ui.LogWarning("Failed to check YubiKey: %v", err)
	} else if present {
		cardInfo, err := yubikeySvc.GetCardInfo(ctx)
		if err != nil {
			ui.LogWarning("Failed to get card info: %v", err)
		} else {
			ui.LogSuccess("YubiKey detected!")
			ui.PrintKeyValue("Serial", cardInfo.Serial)
			ui.PrintKeyValue("Cardholder", cardInfo.Cardholder)
			fmt.Println()
			ui.PrintLabel("Keys on this YubiKey:\n")
			for keyType, keyID := range cardInfo.Keys {
				ui.PrintLabel("  " + keyType + ": ")
				ui.PrintKey(keyID)
				fmt.Println()
			}
		}
	} else {
		ui.LogWarning("No YubiKey detected")
	}
	fmt.Println()

	return nil
}
