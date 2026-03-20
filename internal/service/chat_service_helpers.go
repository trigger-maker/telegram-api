package service

import (
	"context"
	"fmt"

	"telegram-api/internal/domain"

	"github.com/google/uuid"
	"github.com/gotd/td/session"
	tgClient "github.com/gotd/td/telegram"
)

// getValidSession retrieves and validates a session.
func (s *ChatService) getValidSession(
	ctx context.Context,
	userID, sessionID uuid.UUID,
) (*domain.TelegramSession, error) {
	sess, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}
	if sess.UserID != userID {
		return nil, domain.ErrUnauthorized
	}
	if !sess.IsActive {
		return nil, domain.ErrSessionInactive
	}
	return sess, nil
}

// createClient creates a Telegram client from session data.
func (s *ChatService) createClient(ctx context.Context, sess *domain.TelegramSession) (*tgClient.Client, error) {
	apiHashBytes, err := s.tgManager.Decrypt(sess.APIHashEncrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt api_hash: %w", err)
	}

	sessionData, err := s.tgManager.Decrypt(sess.SessionData)
	if err != nil {
		return nil, fmt.Errorf("decrypt session: %w", err)
	}

	storage := &session.StorageMemory{}
	if err := storage.StoreSession(ctx, sessionData); err != nil {
		return nil, fmt.Errorf("store session: %w", err)
	}

	return tgClient.NewClient(sess.APIID, string(apiHashBytes), tgClient.Options{
		SessionStorage: storage,
		Device: tgClient.DeviceConfig{
			DeviceModel:    sess.SessionName,
			SystemVersion:  "1.0",
			AppVersion:     "1.0.0",
			SystemLangCode: "en",
			LangCode:       "en",
		},
	}), nil
}
