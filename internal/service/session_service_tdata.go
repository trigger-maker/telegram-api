package service

import (
	"context"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/internal/telegram"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// validateTDataRequest validates tdata import request.
func validateTDataRequest(apiID int, apiHash string, tdataFiles map[string][]byte) error {
	if apiID <= 0 {
		logger.Warn().Msg("Invalid api_id in tdata import")
		return domain.NewAppError(nil, "api_id must be positive", 400)
	}
	if apiHash == "" {
		logger.Warn().Msg("Empty api_hash in tdata import")
		return domain.NewAppError(nil, "api_hash required", 400)
	}
	if len(tdataFiles) == 0 {
		logger.Warn().Msg("No tdata files provided")
		return domain.NewAppError(nil, "tdata files required", 400)
	}
	return nil
}

// createTDataSession creates and saves the tdata session.
func (s *SessionService) createTDataSession(
	ctx context.Context,
	userID uuid.UUID,
	apiID int,
	apiHashEncrypted []byte,
	sessionName string,
) (*domain.TelegramSession, error) {
	if sessionName == "" {
		sessionName = "TData Import"
	}

	session := &domain.TelegramSession{
		ID:               uuid.New(),
		UserID:           userID,
		PhoneNumber:      "TData-pending",
		APIID:            apiID,
		APIHashEncrypted: apiHashEncrypted,
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

	return session, nil
}

// importTDataFiles imports tdata files through telegram manager.
func (s *SessionService) importTDataFiles(
	ctx context.Context,
	session *domain.TelegramSession,
	apiID int,
	apiHash string,
	tdataFiles map[string][]byte,
) (*telegram.TGUser, error) {
	logger.Debug().
		Str("session_id", session.ID.String()).
		Msg("Importing tdata through telegram manager...")

	user, err := s.tgManager.ImportTData(
		ctx,
		apiID,
		apiHash,
		session.SessionName,
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

	return user, nil
}

// updateSessionWithUserInfo updates session with telegram user info.
func (s *SessionService) updateSessionWithUserInfo(
	ctx context.Context,
	session *domain.TelegramSession,
	user *telegram.TGUser,
) error {
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
		return domain.ErrDatabase
	}

	return nil
}

// ImportTData imports Telegram Desktop session from tdata files.
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

	if err := validateTDataRequest(apiID, apiHash, tdataFiles); err != nil {
		return nil, err
	}

	logger.Debug().Msg("Encrypting api_hash...")
	apiHashEncrypted, err := s.tgManager.Encrypt([]byte(apiHash))
	if err != nil {
		logger.Error().Err(err).Msg("Error encrypting api_hash in tdata import")
		return nil, domain.ErrInternal
	}

	session, err := s.createTDataSession(ctx, userID, apiID, apiHashEncrypted, sessionName)
	if err != nil {
		return nil, err
	}

	user, err := s.importTDataFiles(ctx, session, apiID, apiHash, tdataFiles)
	if err != nil {
		return nil, err
	}

	if err := s.updateSessionWithUserInfo(ctx, session, user); err != nil {
		return nil, err
	}

	logger.Info().
		Str("session_id", session.ID.String()).
		Int64("telegram_user_id", user.ID).
		Str("telegram_username", user.Username).
		Msg("Tdata import successful")

	return session, nil
}
