//go:build integration

package gpg

import (
	"context"
	"testing"

	"github.com/bobbydams/yubikey-manager/internal/executor"
	"github.com/stretchr/testify/require"
)

// TestService_ListSecretKeys_Integration tests listing secret keys with a real GPG instance.
// This test requires GPG to be installed and configured.
// Run with: go test -tags=integration ./...
func TestService_ListSecretKeys_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	exec := executor.NewRealExecutor()
	svc := NewService(exec)

	// This will fail if GPG is not available, which is expected
	_, err := svc.ListSecretKeys(context.Background(), "TEST_KEY_ID")

	// We don't assert on the result, just that it doesn't panic
	// In a real scenario, you'd set up test keys first
	_ = err
}

// TestService_CardStatus_Integration tests card status with a real YubiKey.
// This test requires a YubiKey to be connected.
func TestService_CardStatus_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	exec := executor.NewRealExecutor()
	svc := NewService(exec)

	// This will fail if no YubiKey is present, which is expected
	_, err := svc.CardStatus(context.Background())

	// We don't assert on the result, just that it doesn't panic
	_ = err
}
