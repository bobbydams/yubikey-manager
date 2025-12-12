package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
)

func newVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "verify",
		Aliases: []string{"check"},
		Short:   "Verify GPG and YubiKey setup",
		RunE:    runVerify,
	}
}

func runVerify(cmd *cobra.Command, args []string) error {
	gpgSvc, yubikeySvc, _ := getServices()
	ctx := cmd.Context()

	ui.PrintHeader("Verify GPG/YubiKey Setup")

	errors := 0

	// Check GPG key exists
	fmt.Print("Checking primary key exists... ")
	keys, err := gpgSvc.ListSecretKeys(ctx, cfg.PrimaryKeyID)
	if err == nil && len(keys) > 0 {
		fmt.Print("OK\n")
	} else {
		fmt.Print("FAILED\n")
		errors++
	}

	// Check master key is NOT on machine
	fmt.Print("Checking master key is offline... ")
	hasMaster := false
	for _, key := range keys {
		if key.Type == "sec" {
			hasMaster = true
			break
		}
	}
	if !hasMaster {
		fmt.Print("OK (sec# = offline)\n")
	} else {
		fmt.Print("WARNING (master key may be on machine)\n")
	}

	// Check YubiKey
	fmt.Print("Checking YubiKey presence... ")
	present, err := yubikeySvc.IsPresent(ctx)
	if err == nil && present {
		cardInfo, _ := yubikeySvc.GetCardInfo(ctx)
		fmt.Printf("OK (serial: %s)\n", cardInfo.Serial)
	} else {
		fmt.Print("NOT PRESENT\n")
	}

	// Check Git config
	fmt.Print("Checking Git signing key config... ")
	gitKey := getGitConfig("user.signingkey")
	if gitKey != "" && (containsString(gitKey, cfg.PrimaryKeyID) || containsString(gitKey, cfg.PrimaryKeyFingerprint)) {
		fmt.Print("OK\n")
	} else {
		fmt.Printf("MISMATCH (configured: %s)\n", gitKey)
	}

	// Check commit signing enabled
	fmt.Print("Checking Git commit signing enabled... ")
	gitSign := getGitConfig("commit.gpgsign")
	if gitSign == "true" {
		fmt.Print("OK\n")
	} else {
		fmt.Print("NOT ENABLED\n")
	}

	// Test signing
	fmt.Print("Testing GPG signing... ")
	testCmd := exec.CommandContext(ctx, "gpg", "--sign", "--armor")
	testCmd.Stdin = os.Stdin
	testCmd.Stdout = os.Stdout
	testCmd.Stderr = os.Stderr
	if err := testCmd.Run(); err == nil {
		fmt.Print("OK\n")
	} else {
		fmt.Print("FAILED\n")
		errors++
	}

	fmt.Println()
	if errors == 0 {
		ui.LogSuccess("All checks passed!")
	} else {
		ui.LogError("%d check(s) failed", errors)
	}

	if errors > 0 {
		return fmt.Errorf("verification failed")
	}

	return nil
}

// getGitConfig retrieves a git config value.
func getGitConfig(key string) string {
	cmd := exec.Command("git", "config", "--global", key)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output[:len(output)-1]) // Remove trailing newline
}
