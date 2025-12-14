package cli

import (
	"context"
	"fmt"
	"testing"

	"github.com/bobbydams/yubikey-manager/internal/executor"
	"github.com/bobbydams/yubikey-manager/internal/gpg"
	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{
			name:     "value exists in slice",
			slice:    []string{"a", "b", "c"},
			value:    "b",
			expected: true,
		},
		{
			name:     "value does not exist in slice",
			slice:    []string{"a", "b", "c"},
			value:    "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			value:    "a",
			expected: false,
		},
		{
			name:     "empty value",
			slice:    []string{"a", "b", "c"},
			value:    "",
			expected: false,
		},
		{
			name:     "value at start",
			slice:    []string{"a", "b", "c"},
			value:    "a",
			expected: true,
		},
		{
			name:     "value at end",
			slice:    []string{"a", "b", "c"},
			value:    "c",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "substring exists",
			s:        "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "substring does not exist",
			s:        "hello world",
			substr:   "foo",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "hello world",
			substr:   "",
			expected: true, // Empty substring always matches
		},
		{
			name:     "substring at start",
			s:        "hello world",
			substr:   "hello",
			expected: true,
		},
		{
			name:     "substring at end",
			s:        "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "substring longer than string",
			s:        "hello",
			substr:   "hello world",
			expected: false,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "a",
			expected: false,
		},
		{
			name:     "exact match",
			s:        "hello",
			substr:   "hello",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsString(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveMasterKey(t *testing.T) {
	// This is a complex function that requires real GPG operations
	// We'll test the error handling paths with mocked services
	ctx := context.Background()
	keyID := "ABC123DEF4567890"
	shortKeyID := keyID[:16]

	// Mock output for ListSecretKeys showing master key IS on machine (type "sec", not "sec#")
	// Format: gpg --list-secret-keys --keyid-format=long KEYID
	masterKeyOnMachineOutput := `sec   ed25519/ABC123DEF4567890 2025-09-05 [SC] [expires: 2030-09-04]
      Key fingerprint = FA57 C851 31F1 1B28 EE23  6A4F ABC1 23DE F456 7890
uid                 [ultimate] Test User <test@example.com>
ssb   cv25519/1234567890ABCDEF 2025-09-05 [E] [expires: 2030-09-04]
`
	// Mock output for ListSecretKeys showing master key is OFFLINE (type "sec#")
	masterKeyOfflineOutput := `sec#  ed25519/ABC123DEF4567890 2025-09-05 [SC] [expires: 2030-09-04]
      Key fingerprint = FA57 C851 31F1 1B28 EE23  6A4F ABC1 23DE F456 7890
uid                 [ultimate] Test User <test@example.com>
ssb>  cv25519/1234567890ABCDEF 2025-09-05 [E] [expires: 2030-09-04]
`

	t.Run("master key already offline - returns success", func(t *testing.T) {
		mockExecutor := executor.NewMockExecutor()
		mockExecutor.SetOutput("gpg --list-secret-keys --keyid-format=long "+shortKeyID, []byte(masterKeyOfflineOutput))
		gpgSvc := gpg.NewService(mockExecutor)

		err := removeMasterKey(ctx, gpgSvc, keyID)
		assert.NoError(t, err) // Should succeed without doing anything
	})

	t.Run("error on list secret keys", func(t *testing.T) {
		mockExecutor := executor.NewMockExecutor()
		mockExecutor.SetError("gpg --list-secret-keys --keyid-format=long "+shortKeyID, fmt.Errorf("list failed"))
		gpgSvc := gpg.NewService(mockExecutor)

		err := removeMasterKey(ctx, gpgSvc, keyID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list keys")
	})

	t.Run("error on export public key", func(t *testing.T) {
		mockExecutor := executor.NewMockExecutor()
		// ListSecretKeys succeeds - master key is on machine
		mockExecutor.SetOutput("gpg --list-secret-keys --keyid-format=long "+shortKeyID, []byte(masterKeyOnMachineOutput))
		// ExportSecretSubkeys can fail (we handle this gracefully)
		mockExecutor.SetError("gpg --export-secret-subkeys "+shortKeyID, fmt.Errorf("export subkeys failed"))
		// ExportPublicKey fails
		mockExecutor.SetError("gpg --export --armor "+shortKeyID, fmt.Errorf("export public key failed"))
		gpgSvc := gpg.NewService(mockExecutor)

		err := removeMasterKey(ctx, gpgSvc, keyID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to export public key")
	})

	t.Run("error on delete secret key", func(t *testing.T) {
		mockExecutor := executor.NewMockExecutor()
		// ListSecretKeys succeeds - master key is on machine
		mockExecutor.SetOutput("gpg --list-secret-keys --keyid-format=long "+shortKeyID, []byte(masterKeyOnMachineOutput))
		// ExportSecretSubkeys succeeds
		mockExecutor.SetOutput("gpg --export-secret-subkeys "+shortKeyID, []byte("subkey data"))
		// ExportPublicKey succeeds
		mockExecutor.SetOutput("gpg --export --armor "+shortKeyID, []byte("public key data"))
		// Delete fails
		mockExecutor.SetError("gpg --batch --yes --delete-secret-keys "+keyID, fmt.Errorf("delete failed"))
		gpgSvc := gpg.NewService(mockExecutor)

		err := removeMasterKey(ctx, gpgSvc, keyID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete secret key")
	})

	t.Run("success - full removal flow", func(t *testing.T) {
		mockExecutor := executor.NewMockExecutor()
		// ListSecretKeys succeeds - master key is on machine
		mockExecutor.SetOutput("gpg --list-secret-keys --keyid-format=long "+shortKeyID, []byte(masterKeyOnMachineOutput))
		// ExportSecretSubkeys succeeds
		mockExecutor.SetOutput("gpg --export-secret-subkeys "+shortKeyID, []byte("subkey data"))
		// ExportPublicKey succeeds
		mockExecutor.SetOutput("gpg --export --armor "+shortKeyID, []byte("public key data"))
		// Delete succeeds
		mockExecutor.SetOutput("gpg --batch --yes --delete-secret-keys "+keyID, []byte(""))
		// Import public key succeeds
		mockExecutor.SetOutput("gpg --import", []byte(""))
		gpgSvc := gpg.NewService(mockExecutor)

		err := removeMasterKey(ctx, gpgSvc, keyID)
		assert.NoError(t, err)
	})
}
