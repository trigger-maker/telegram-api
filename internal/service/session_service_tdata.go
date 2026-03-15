package service

import (
	"context"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// ImportTData imports Telegram Desktop session from tdata files
func (s *SessionService) ImportTData(
	ctx context.Context,
	userID uuid.UUID,
	apiID int,
	apiHash string,
	sessionName string,
	tdataFiles map[string][]byte,
) (*domain.TelegramSession, error) {
	logger.Debug().
		Str("user_id", userID.String()).
		Int("api_id", apiID).
		Str("session_name", sessionName).
		Msg("Starting tdata import...")

	if apiID <= 0 {
		logger.Warn().Msg("Invalid api_id in tdata import")
		return nil, domain.NewAppError(nil, "api_id must be positive", 400)
	}
	if apiHash == "" {
		logger.Warn().Msg("Empty api_hash in tdata import")
		return nil, domain.NewAppError(nil, "api_hash required", 400)
	}
	if len(tdataFiles) == 0 {
		logger.Warn().Msg("No tdata files provided")
		return nil, domain.NewAppError(nil, "tdata files required", 400)
	}

	if sessionName == "" {
		sessionName = "TData Import"
	}

	logger.Debug().Msg("Encrypting api_hash...")
	apiHashEncrypted, err := s.tgManager.Encrypt([]byte(apiHash))
	if err != nil {
		logger.Error().Err(err).Msg("Error encrypting api_hash in tdata import")
		return nil, domain.ErrInternal
	}

	session := &domain.TelegramSession{
		ID:               uuid.New(),
		UserID:           userID,
		PhoneNumber:      "TData-pending",
		ApiID:            apiID,
		ApiHashEncrypted: apiHashEncrypted,
		SessionName:      sessionName,
		AuthState:        domain.SessionPending,
		AuthMethod:       domain.AuthMethodTData,
		IsActive:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	logger.Debug().
		Str("session_id", session.ID.String()).
		Str("session_name", sessionName).
		Msg("Saving tdata session to DB...")

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		logger.Error().
			Err(err).
			Str("session_id", session.ID.String()).
			Msg("Error creating tdata session in DB")
		return nil, domain.ErrDatabase
	}

	logger.Debug().
		Str("session_id", session.ID.String()).
		Msg("Importing tdata through telegram manager...")

	user, err := s.tgManager.ImportTData(
		ctx,
		apiID,
		apiHash,
		sessionName,
		session.ID.String(),
		tdataFiles,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Str("session_id", session.ID.String()).
			Msg("Error importing tdata")

		session.AuthState = domain.SessionFailed
		_ = s.sessionRepo.Update(ctx, session)

		return nil, err
	}

	session.TelegramUserID = user.ID
	session.TelegramUsername = user.Username
	session.AuthState = domain.SessionAuthenticated
	session.IsActive = true
	session.UpdatedAt = time.Now()

	logger.Debug().
		Str("session_id", session.ID.String()).
		Int64("telegram_user_id", user.ID).
		Str("telegram_username", user.Username).
		Msg("Updating session with user info...")

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		logger.Error().
			Err(err).
			Str("session_id", session.ID.String()).
			Msg("Error updating session after tdata import")
		return nil, domain.ErrDatabase
	}

	logger.Info().
		Str("session_id", session.ID.String()).
		Int64("telegram_user_id", user.ID).
		Str("telegram_username", user.Username).
		Msg("Tdata import successful")

	return session, nil
}
