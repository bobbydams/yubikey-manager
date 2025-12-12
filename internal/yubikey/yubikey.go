package yubikey

import (
	"context"
	"fmt"

	"github.com/bobbydams/yubikey-manager/internal/executor"
	"github.com/bobbydams/yubikey-manager/internal/gpg"
)

// YubiKeyService provides operations for interacting with YubiKeys.
type YubiKeyService interface {
	// IsPresent checks if a YubiKey is currently connected.
	IsPresent(ctx context.Context) (bool, error)

	// GetCardInfo returns information about the connected YubiKey.
	GetCardInfo(ctx context.Context) (*gpg.CardInfo, error)

	// EditCard starts an interactive GPG card edit session.
	EditCard(ctx context.Context) error
}

// Service implements YubiKeyService.
type Service struct {
	gpgService gpg.GPGService
	exec       executor.Executor
}

// NewService creates a new YubiKey service.
func NewService(gpgService gpg.GPGService, exec executor.Executor) *Service {
	return &Service{
		gpgService: gpgService,
		exec:       exec,
	}
}

// IsPresent checks if a YubiKey is currently connected.
func (s *Service) IsPresent(ctx context.Context) (bool, error) {
	_, err := s.gpgService.CardStatus(ctx)
	if err != nil {
		// If card status fails, assume no YubiKey is present
		return false, nil
	}
	return true, nil
}

// GetCardInfo returns information about the connected YubiKey.
func (s *Service) GetCardInfo(ctx context.Context) (*gpg.CardInfo, error) {
	info, err := s.gpgService.CardStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get card info: %w", err)
	}
	return info, nil
}

// EditCard starts an interactive GPG card edit session.
func (s *Service) EditCard(ctx context.Context) error {
	args := []string{"--card-edit"}
	return s.exec.RunInteractive(ctx, "gpg", args...)
}
