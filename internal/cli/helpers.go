package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/bobbydams/yubikey-manager/internal/gpg"
)

// removeMasterKey removes the master key from the local keyring.
func removeMasterKey(ctx context.Context, gpgSvc *gpg.Service, fingerprint string) error {
	keyID := fingerprint
	if len(fingerprint) > 16 {
		keyID = fingerprint[:16]
	}

	// Check if master key is actually on the machine
	keys, err := gpgSvc.ListSecretKeys(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	hasMasterOnMachine := false
	for _, key := range keys {
		// "sec" (not "sec#") means master key is on machine
		if key.Type == "sec" && !strings.HasSuffix(key.Type, "#") {
			hasMasterOnMachine = true
			break
		}
	}

	if !hasMasterOnMachine {
		// Master key is already offline (sec#), nothing to remove
		return nil
	}

	// Export subkeys first (these may be stubs for keys on cards, but that's OK)
	subkeys, err := gpgSvc.ExportSecretSubkeys(ctx, keyID)
	if err != nil {
		// If export fails, subkeys might already be on cards - continue anyway
		subkeys = nil
	}

	// Export public key
	publicKey, err := gpgSvc.ExportPublicKey(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to export public key: %w", err)
	}

	// Delete secret key
	if err := gpgSvc.DeleteSecretKey(ctx, fingerprint); err != nil {
		return fmt.Errorf("failed to delete secret key: %w", err)
	}

	// Re-import public key
	if err := gpgSvc.ImportKey(ctx, publicKey); err != nil {
		return fmt.Errorf("failed to import public key: %w", err)
	}

	// Re-import subkeys if we have them
	if subkeys != nil && len(subkeys) > 0 {
		if err := gpgSvc.ImportKey(ctx, subkeys); err != nil {
			// This can fail if subkeys are on cards - not a fatal error
			// The key stubs will be recreated when the card is used
		}
	}

	return nil
}

// contains checks if a string slice contains a value.
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
