package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/bobbydams/yubikey-manager/internal/gpg"
	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
)

func newVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "verify",
		Aliases:      []string{"check"},
		Short:        "Verify GPG and YubiKey setup",
		SilenceUsage: true, // Don't print usage on errors
		RunE:         runVerify,
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

	// Check YubiKey and find the signing subkey on it
	var signingSubkeyID string
	var cardInfo *gpg.CardInfo
	fmt.Print("Checking YubiKey presence... ")

	// Create a context with timeout for YubiKey detection to prevent hanging
	yubikeyCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	present, err := yubikeySvc.IsPresent(yubikeyCtx)
	if err == nil && present {
		// Use the same timeout context for getting card info
		cardInfo, err = yubikeySvc.GetCardInfo(yubikeyCtx)
		if err == nil {
			fmt.Printf("OK (serial: %s)\n", cardInfo.Serial)
			// Try to get signature key ID from card info first
			if sigKey, ok := cardInfo.Keys["Signature"]; ok && sigKey != "" && sigKey != "[none]" {
				fmt.Printf("  └─ Signature key on YubiKey: %s\n", sigKey)
				signingSubkeyID = sigKey
			} else {
				// If card info doesn't have the signature key, find it by matching card serial
				// The card serial format in GPG key listing is "0006 XXXXXXXX" where XXXXXXXX is the serial
				cardSerialFormatted := fmt.Sprintf("0006 %s", cardInfo.Serial)
				// Also try without space (some formats might differ)
				cardSerialFormattedAlt := fmt.Sprintf("0006%s", cardInfo.Serial)

				// First, try to find by card-no matching
				for _, key := range keys {
					// Look for signing subkeys (ssb) with S capability that are on this card
					if key.Type == "ssb" && contains(key.Capabilities, "S") {
						// Check if this key is on the current card by matching card-no
						if key.CardNo == cardSerialFormatted || key.CardNo == cardSerialFormattedAlt {
							// This is the signing subkey on the current YubiKey
							signingSubkeyID = key.KeyID
							fmt.Printf("  └─ Found signing subkey on YubiKey: %s\n", signingSubkeyID)
							break
						}
					}
				}

				// If still not found, try to use the most recent signing subkey that's on a card
				// This is a fallback when card-no doesn't match (e.g., after moving a key)
				if signingSubkeyID == "" {
					// Look for the most recent signing subkey that's on a card (has CardNo set)
					var latestSigningKey *gpg.Key
					for i := range keys {
						key := &keys[i]
						if key.Type == "ssb" && contains(key.Capabilities, "S") && key.CardNo != "" {
							// This is a signing subkey on a card
							if latestSigningKey == nil {
								latestSigningKey = key
							}
						}
					}
					if latestSigningKey != nil {
						signingSubkeyID = latestSigningKey.KeyID
						fmt.Printf("  └─ Using signing subkey on card: %s\n", signingSubkeyID)
						ui.LogInfo("  └─ Note: Using most recent signing subkey on a card. If this is wrong, specify the key ID manually.")
					}
				}
			}
		} else {
			// Check if it was a timeout
			if yubikeyCtx.Err() == context.DeadlineExceeded {
				fmt.Print("TIMEOUT (GPG may be waiting for input)\n")
				ui.LogWarning("  └─ gpg --card-status timed out. This may indicate:")
				ui.LogWarning("  └─ 1. GPG is waiting for PIN entry")
				ui.LogWarning("  └─ 2. Multiple YubiKeys detected and GPG is waiting for card selection")
				ui.LogWarning("  └─ 3. YubiKey needs to be touched/activated")
				ui.LogInfo("  └─ Try running 'gpg --card-status' manually to see what's happening")
			} else {
				fmt.Print("OK (unable to get card info)\n")
			}
		}
	} else {
		// Check if it was a timeout
		if yubikeyCtx.Err() == context.DeadlineExceeded {
			fmt.Print("TIMEOUT (GPG may be waiting for input)\n")
			ui.LogWarning("  └─ YubiKey detection timed out. GPG may be waiting for user interaction.")
		} else {
			fmt.Print("NOT PRESENT\n")
		}
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

	// Test signing with the specific subkey ID from the current YubiKey
	fmt.Print("Testing GPG signing... ")
	if signingSubkeyID == "" {
		// If we couldn't get the subkey ID from the card, try to find it by card serial
		// This handles the case where the card status shows "[none]" but the key is actually on the card
		if present && cardInfo != nil {
			cardSerialFormatted := fmt.Sprintf("0006 %s", cardInfo.Serial)
			cardSerialFormattedAlt := fmt.Sprintf("0006%s", cardInfo.Serial)
			for _, key := range keys {
				if key.Type == "ssb" && contains(key.Capabilities, "S") {
					if key.CardNo == cardSerialFormatted || key.CardNo == cardSerialFormattedAlt {
						signingSubkeyID = key.KeyID
						break
					}
				}
			}
		}
		// If still not found, we can't test signing without knowing which subkey to use
		// Don't fall back to primary key ID as it will prompt for card selection
		if signingSubkeyID == "" {
			fmt.Print("SKIPPED (unable to identify signing subkey on YubiKey)\n")
			ui.LogInfo("  └─ Could not find the signing subkey on the current YubiKey.")
			ui.LogInfo("  └─ This may happen if the subkey was recently moved to the YubiKey.")
			ui.LogInfo("  └─ Try running 'gpg --card-status' to verify the key is on the card.")
		}
	}

	// Only test signing if we found a signing subkey ID
	if signingSubkeyID != "" {
		// Use echo to provide input to gpg with explicit key ID from the current card
		// Try to use the full fingerprint if available, as it's more specific
		keyIDForSigning := signingSubkeyID
		for _, key := range keys {
			if key.KeyID == signingSubkeyID && key.Fingerprint != "" {
				// Use full fingerprint for more specificity
				keyIDForSigning = key.Fingerprint
				break
			}
		}

		// Create a context with timeout for the signing test
		// This prevents hanging if GPG prompts for PIN or card selection
		signingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		// First try non-interactive mode (works if PIN is cached or using GUI pinentry)
		testCmd := exec.CommandContext(signingCtx, "sh", "-c", fmt.Sprintf("echo 'test' | gpg --batch --pinentry-mode=loopback --default-key %s --sign --armor > /dev/null 2>&1", keyIDForSigning))
		if err := testCmd.Run(); err == nil {
			fmt.Print("OK\n")
		} else {
			// Non-interactive failed - offer interactive test
			fmt.Print("INTERACTIVE\n")
			ui.LogInfo("  └─ Automated test requires PIN entry.")

			if ui.Confirm("  └─ Run interactive signing test? (You'll need to enter your PIN)") {
				// Create a temporary file with test data to sign
				// This allows pinentry to use stdin/TTY for PIN entry
				tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("ykgpg-test-%d.txt", time.Now().Unix()))
				if err := os.WriteFile(tmpFile, []byte("test\n"), 0644); err != nil {
					fmt.Print("  └─ Testing signing... FAILED\n")
					ui.LogInfo("  └─ Error creating temp file: %v", err)
					errors++
				} else {
					defer os.Remove(tmpFile) // Clean up temp file

					fmt.Print("  └─ Testing signing (enter PIN when prompted)... ")
					// Flush stdout to ensure the prompt is visible before GPG runs
					os.Stdout.Sync()

					// Sign the file - this allows pinentry to use the TTY
					// Use --quiet to suppress most informational messages
					interactiveCmd := exec.Command("gpg", "--quiet", "--default-key", keyIDForSigning, "--sign", "--armor", "--output", "/dev/null", tmpFile)
					// Connect stdin for pinentry
					interactiveCmd.Stdin = os.Stdin
					// Capture stderr to filter out informational messages, but pinentry uses TTY directly
					var stderrBuf bytes.Buffer
					interactiveCmd.Stderr = &stderrBuf
					// Redirect stdout to /dev/null to avoid GPG output mixing with our formatting
					devNull, _ := os.OpenFile("/dev/null", os.O_WRONLY, 0)
					defer devNull.Close()
					interactiveCmd.Stdout = devNull

					// Ensure GPG_TTY is set for pinentry
					if tty := os.Getenv("GPG_TTY"); tty == "" {
						// Try to get TTY from /dev/tty
						if ttyFile, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
							ttyFile.Close()
							interactiveCmd.Env = append(os.Environ(), "GPG_TTY=/dev/tty")
						}
					}

					if err := interactiveCmd.Run(); err == nil {
						fmt.Print("OK\n")
					} else {
						fmt.Print("FAILED\n")
						// Only show stderr if it contains actual errors (not just informational messages)
						stderrStr := stderrBuf.String()
						if stderrStr != "" && !containsString(stderrStr, "using") {
							ui.LogInfo("  └─ GPG error: %s", stderrStr)
						}
						ui.LogInfo("  └─ Error: %v", err)
						ui.LogInfo("  └─ This might be due to PIN entry issues. Try manually:")
						ui.LogInfo("  └─   echo 'test' | gpg --default-key %s --sign --armor", keyIDForSigning)
						errors++
					}
				}
			} else {
				ui.LogInfo("  └─ To test manually: echo 'test' | gpg --default-key %s --sign --armor", keyIDForSigning)
			}
		}
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
