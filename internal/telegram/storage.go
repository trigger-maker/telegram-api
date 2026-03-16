package telegram

import (
	"context"
	"fmt"

	"telegram-api/internal/domain"
	"telegram-api/pkg/crypto"

	"github.com/google/uuid"
)

// PersistentSessionStorage implements telegram.SessionStorage with encryption.
type PersistentSessionStorage struct {
	crypter   *crypto.Crypter
	repo      domain.SessionRepository
	sessionID string
}

// NewPersistentSessionStorage creates a new persistent session storage.
func NewPersistentSessionStorage(
	crypter *crypto.Crypter,
	repo domain.SessionRepository,
	sessionID string,
) *PersistentSessionStorage {
	return &PersistentSessionStorage{
		crypter:   crypter,
		repo:      repo,
		sessionID: sessionID,
	}
}

// StoreSession encrypts and saves session data to database.
func (s *PersistentSessionStorage) StoreSession(_ context.Context, data []byte) error {
	if s.crypter == nil {
		return fmt.Errorf("crypter not initialized")
	}

	encrypted, err := s.crypter.Encrypt(data)
	if err != nil {
		return fmt.Errorf("encrypt session data: %w", err)
	}

	err = s.repo.UpdateSessionData(s.sessionID, encrypted)
	if err != nil {
		return fmt.Errorf("update session data: %w", err)
	}

	return nil
}

// LoadSession loads and decrypts session data from database.
func (s *PersistentSessionStorage) LoadSession(ctx context.Context) ([]byte, error) {
	if s.crypter == nil {
		return []byte{}, fmt.Errorf("crypter not initialized")
	}

	sessionUUID, err := uuid.Parse(s.sessionID)
	if err != nil {
		return []byte{}, fmt.Errorf("parse session id: %w", err)
	}

	session, err := s.repo.GetByID(ctx, sessionUUID)
	if err != nil {
		if err == domain.ErrSessionNotFound {
			return []byte{}, nil
		}
		return []byte{}, fmt.Errorf("get session: %w", err)
	}

	if len(session.SessionData) == 0 {
		return []byte{}, nil
	}

	decrypted, err := s.crypter.Decrypt(session.SessionData)
	if err != nil {
		return []byte{}, fmt.Errorf("decrypt session data: %w", err)
	}

	return decrypted, nil
}
