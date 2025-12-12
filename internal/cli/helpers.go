package cli

import (
	"context"

	"github.com/bobbydams/yubikey-manager/internal/gpg"
)

// removeMasterKey removes the master key from the local keyring.
func removeMasterKey(ctx context.Context, gpgSvc *gpg.Service, fingerprint string) error {
	// Export subkeys first
	subkeys, err := gpgSvc.ExportSecretSubkeys(ctx, fingerprint[:16])
	if err != nil {
		return err
	}

	// Export public key
	publicKey, err := gpgSvc.ExportPublicKey(ctx, fingerprint[:16])
	if err != nil {
		return err
	}

	// Delete secret key
	if err := gpgSvc.DeleteSecretKey(ctx, fingerprint); err != nil {
		return err
	}

	// Re-import public key and subkeys
	if err := gpgSvc.ImportKey(ctx, publicKey); err != nil {
		return err
	}
	if err := gpgSvc.ImportKey(ctx, subkeys); err != nil {
		return err
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
