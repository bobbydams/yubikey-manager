package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bobbydams/yubikey-manager/internal/gpg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockGPGService is a simple mock implementation of GPGService for testing.
type MockGPGService struct {
	ExportPublicKeyFunc  func(ctx context.Context, keyID string) ([]byte, error)
	ExportOwnerTrustFunc func(ctx context.Context) ([]byte, error)
	ListSecretKeysFunc   func(ctx context.Context, keyID string) ([]gpg.Key, error)
}

func (m *MockGPGService) ListSecretKeys(ctx context.Context, keyID string) ([]gpg.Key, error) {
	if m.ListSecretKeysFunc != nil {
		return m.ListSecretKeysFunc(ctx, keyID)
	}
	return nil, nil
}

func (m *MockGPGService) CardStatus(ctx context.Context) (*gpg.CardInfo, error) {
	return nil, nil
}

func (m *MockGPGService) ExportPublicKey(ctx context.Context, keyID string) ([]byte, error) {
	if m.ExportPublicKeyFunc != nil {
		return m.ExportPublicKeyFunc(ctx, keyID)
	}
	return nil, nil
}

func (m *MockGPGService) ExportSecretSubkeys(ctx context.Context, keyID string) ([]byte, error) {
	return nil, nil
}

func (m *MockGPGService) DeleteSecretKey(ctx context.Context, fingerprint string) error {
	return nil
}

func (m *MockGPGService) ImportKey(ctx context.Context, keyData []byte) error {
	return nil
}

func (m *MockGPGService) ExportOwnerTrust(ctx context.Context) ([]byte, error) {
	if m.ExportOwnerTrustFunc != nil {
		return m.ExportOwnerTrustFunc(ctx)
	}
	return nil, nil
}

func (m *MockGPGService) CheckTrustDB(ctx context.Context) error {
	return nil
}

func (m *MockGPGService) EditKey(ctx context.Context, keyID string) error {
	return nil
}

func TestService_CreateBackup(t *testing.T) {
	keyID := "ABC123DEF4567890"
	publicKeyData := []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n...\n-----END PGP PUBLIC KEY BLOCK-----")
	trustData := []byte("trust data")
	keys := []gpg.Key{
		{Type: "sec", KeyID: keyID, Capabilities: []string{"S", "C"}},
		{Type: "ssb", KeyID: "ABC123", Capabilities: []string{"S"}},
	}

	mockGPG := &MockGPGService{
		ExportPublicKeyFunc: func(ctx context.Context, kID string) ([]byte, error) {
			return publicKeyData, nil
		},
		ExportOwnerTrustFunc: func(ctx context.Context) ([]byte, error) {
			return trustData, nil
		},
		ListSecretKeysFunc: func(ctx context.Context, kID string) ([]gpg.Key, error) {
			return keys, nil
		},
	}
	svc := NewService(mockGPG)

	// Create a temporary directory for backups
	tmpDir, err := os.MkdirTemp("", "backup-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	backupPath, err := svc.CreateBackup(context.Background(), keyID, tmpDir)

	require.NoError(t, err)
	assert.NotEmpty(t, backupPath)

	// Verify backup files exist
	publicKeyPath := filepath.Join(backupPath, "public-key.asc")
	trustPath := filepath.Join(backupPath, "trustdb.txt")
	keyListPath := filepath.Join(backupPath, "key-list.txt")

	assert.FileExists(t, publicKeyPath)
	assert.FileExists(t, trustPath)
	assert.FileExists(t, keyListPath)

	// Verify file contents
	publicKeyContent, err := os.ReadFile(publicKeyPath)
	require.NoError(t, err)
	assert.Equal(t, publicKeyData, publicKeyContent)

	trustContent, err := os.ReadFile(trustPath)
	require.NoError(t, err)
	assert.Equal(t, trustData, trustContent)
}
