package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export",
		Aliases: []string{"export-public"},
		Short:   "Export public key to file",
		RunE:    runExport,
	}

	cmd.Flags().StringP("output", "o", "", "Output file path (default: ~/public-key-YYYYMMDD.asc)")

	return cmd
}

func runExport(cmd *cobra.Command, args []string) error {
	gpgSvc, _, _ := getServices()
	ctx := cmd.Context()

	ui.PrintHeader("Export Public Key")

	outputFile, _ := cmd.Flags().GetString("output")
	if outputFile == "" {
		timestamp := time.Now().Format("20060102")
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		outputFile = filepath.Join(homeDir, fmt.Sprintf("public-key-%s.asc", timestamp))
	}

	// Export public key
	publicKeyData, err := gpgSvc.ExportPublicKey(ctx, cfg.PrimaryKeyID)
	if err != nil {
		return fmt.Errorf("failed to export public key: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputFile, publicKeyData, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	ui.LogSuccess("Public key exported to: %s", outputFile)
	fmt.Println()
	fmt.Println("You can:")
	fmt.Println("  1. Upload to https://keys.openpgp.org/upload")
	fmt.Println("  2. Add to GitHub: Settings → SSH and GPG keys → New GPG key")
	fmt.Println("  3. Share with others for encrypted communication")

	return nil
}
