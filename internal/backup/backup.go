package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bobbydams/yubikey-manager/internal/gpg"
)

// BackupService provides operations for backing up GPG keys and trust database.
type BackupService interface {
	// CreateBackup creates a backup of the GPG keyring and trust database.
	CreateBackup(ctx context.Context, keyID string, backupDir string) (string, error)
}

// Service implements BackupService.
type Service struct {
	gpgService gpg.GPGService
}

// NewService creates a new backup service.
func NewService(gpgService gpg.GPGService) *Service {
	return &Service{gpgService: gpgService}
}

// BackupResult contains information about a created backup.
type BackupResult struct {
	Path      string
	Timestamp time.Time
}

// CreateBackup creates a backup of the GPG keyring and trust database.
// Returns the path to the created backup directory.
func (s *Service) CreateBackup(ctx context.Context, keyID string, backupDir string) (string, error) {
	// Create backup directory with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("gpg-backup-%s", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup public key
	publicKeyData, err := s.gpgService.ExportPublicKey(ctx, keyID)
	if err != nil {
		return "", fmt.Errorf("failed to export public key: %w", err)
	}

	publicKeyPath := filepath.Join(backupPath, "public-key.asc")
	if err := os.WriteFile(publicKeyPath, publicKeyData, 0644); err != nil {
		return "", fmt.Errorf("failed to write public key backup: %w", err)
	}

	// Backup trust database
	trustData, err := s.gpgService.ExportOwnerTrust(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to export ownertrust: %w", err)
	}

	trustPath := filepath.Join(backupPath, "trustdb.txt")
	if err := os.WriteFile(trustPath, trustData, 0644); err != nil {
		return "", fmt.Errorf("failed to write trustdb backup: %w", err)
	}

	// Save key list
	keys, err := s.gpgService.ListSecretKeys(ctx, keyID)
	if err != nil {
		return "", fmt.Errorf("failed to list secret keys: %w", err)
	}

	keyListPath := filepath.Join(backupPath, "key-list.txt")
	keyListContent := formatKeyList(keys)
	if err := os.WriteFile(keyListPath, []byte(keyListContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write key list backup: %w", err)
	}

	return backupPath, nil
}

// formatKeyList formats a list of keys into a readable string.
func formatKeyList(keys []gpg.Key) string {
	var result string
	for _, key := range keys {
		result += fmt.Sprintf("%s %s [%s]", key.Type, key.KeyID, formatCapabilities(key.Capabilities))
		if key.Expires != "" {
			result += fmt.Sprintf(" expires: %s", key.Expires)
		}
		if key.CardNo != "" {
			result += fmt.Sprintf(" card-no: %s", key.CardNo)
		}
		result += "\n"
	}
	return result
}

// formatCapabilities formats capability flags as a string.
func formatCapabilities(caps []string) string {
	return fmt.Sprintf("%v", caps)
}
