package service

import (
	"context"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// ListSessions returns all sessions for a user.
func (s *SessionService) ListSessions(ctx context.Context, userID uuid.UUID) ([]domain.TelegramSession, error) {
	return s.sessionRepo.ListByUserID(ctx, userID)
}

// GetSession returns a session by ID.
func (s *SessionService) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.TelegramSession, error) {
	return s.sessionRepo.GetByID(ctx, sessionID)
}

// DeleteSession deletes a session by ID.
func (s *SessionService) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.IsActive && len(session.SessionData) > 0 {
		logoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := s.tgManager.LogOut(
			logoutCtx,
			session.APIID,
			session.APIHashEncrypted,
			session.SessionData,
			session.SessionName,
		)
		if err != nil {
			logger.Warn().Err(err).Str("session_id", sessionID.String()).Msg("Error during Telegram logout")
		}
	}

	return s.sessionRepo.Delete(ctx, sessionID)
}
