package service

import (
	"context"
	"fmt"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/internal/telegram"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// handleQRResult processes QR auth result in background
func (s *SessionService) handleQRResult(sessionID uuid.UUID, resultChan <-chan telegram.QRAuthResult) {
	result, ok := <-resultChan
	if !ok {
		logger.Warn().Str("session_id", sessionID.String()).Msg("Channel closed without result")
		return
	}
	ctx := context.Background()
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		logger.Error().Err(err).Str("session_id", sessionID.String()).Msg("Session not found")
		return
	}
	if session.IsActive {
		return
	}
	if result.Error != nil {
		session.AuthState = domain.SessionFailed
		_ = s.sessionRepo.Update(ctx, session)
		logger.Warn().Err(result.Error).Str("session_id", sessionID.String()).Msg("QR auth failed")
		return
	}
	var encryptedSessionData []byte
	if len(result.SessionData) > 0 {
		encryptedSessionData, _ = s.tgManager.Encrypt(result.SessionData)
	}
	session.SessionData = encryptedSessionData
	session.AuthState = domain.SessionAuthenticated
	session.IsActive = true
	session.TelegramUserID = result.User.ID
	session.TelegramUsername = result.User.Username
	session.UpdatedAt = time.Now()
	session.PhoneNumber = fmt.Sprintf("TG-%d", result.User.ID)
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		logger.Error().Err(err).Msg("Error updating authenticated session")
		return
	}
	logger.Info().
		Str("session_id", sessionID.String()).
		Int64("telegram_user_id", result.User.ID).
		Str("telegram_username", result.User.Username).
		Msg("QR session authenticated successfully")
}

// completeAuth completes the authentication process
func (s *SessionService) completeAuth(ctx context.Context, session *domain.TelegramSession, user *telegram.TGUser, sessionData []byte, cacheKey string) (*domain.TelegramSession, error) {
	var encryptedSessionData []byte
	if len(sessionData) > 0 {
		encryptedSessionData, _ = s.tgManager.Encrypt(sessionData)
	}

	session.SessionData = encryptedSessionData
	session.AuthState = domain.SessionAuthenticated
	session.IsActive = true
	session.TelegramUserID = user.ID
	session.TelegramUsername = user.Username
	session.UpdatedAt = time.Now()

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, domain.ErrDatabase
	}

	_ = s.cache.Delete(ctx, cacheKey)

	logger.Info().
		Str("session_id", session.ID.String()).
		Int64("tg_user_id", user.ID).
		Str("tg_username", user.Username).
		Msg("Session authenticated")

	return session, nil
}

// defaultSessionName returns a default session name if not provided
func defaultSessionName(name, fallback string) string {
	if name != "" {
		return name
	}
	return "Session " + fallback
}
