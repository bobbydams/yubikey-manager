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

	t.Run("error on export subkeys", func(t *testing.T) {
		mockExecutor := executor.NewMockExecutor()
		// ExportSecretSubkeys calls: gpg --export-secret-subkeys KEYID (no --armor)
		keyID := "ABC123DEF4567890"[:16] // First 16 chars
		mockExecutor.SetError("gpg --export-secret-subkeys "+keyID, fmt.Errorf("export failed"))
		gpgSvc := gpg.NewService(mockExecutor)

		err := removeMasterKey(ctx, gpgSvc, "ABC123DEF4567890")
		assert.Error(t, err)
	})

	t.Run("error on export public key", func(t *testing.T) {
		mockExecutor := executor.NewMockExecutor()
		keyID := "ABC123DEF4567890"[:16]
		// First call (ExportSecretSubkeys) succeeds
		mockExecutor.SetOutput("gpg --export-secret-subkeys "+keyID, []byte("subkey data"))
		// Second call (ExportPublicKey) fails - uses --export --armor KEYID
		mockExecutor.SetError("gpg --export --armor "+keyID, fmt.Errorf("export public key failed"))
		gpgSvc := gpg.NewService(mockExecutor)

		err := removeMasterKey(ctx, gpgSvc, "ABC123DEF4567890")
		assert.Error(t, err)
	})

	t.Run("error on delete secret key", func(t *testing.T) {
		mockExecutor := executor.NewMockExecutor()
		keyID := "ABC123DEF4567890"[:16]
		fullFingerprint := "ABC123DEF4567890"
		// Export calls succeed
		mockExecutor.SetOutput("gpg --armor --export-secret-subkeys "+keyID, []byte("subkey data"))
		mockExecutor.SetOutput("gpg --armor --export "+keyID, []byte("public key data"))
		// Delete fails
		mockExecutor.SetError("gpg --batch --yes --delete-secret-keys "+fullFingerprint, fmt.Errorf("delete failed"))
		gpgSvc := gpg.NewService(mockExecutor)

		err := removeMasterKey(ctx, gpgSvc, "ABC123DEF4567890")
		// The error should be returned
		assert.Error(t, err)
	})
}
