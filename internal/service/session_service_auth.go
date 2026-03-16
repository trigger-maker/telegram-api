package service

import (
	"context"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/internal/telegram"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

const (
	maxQRAttempts = 3               // Maximum automatic QR attempts
	qrTimeout     = 2 * time.Minute // Timeout per QR
)

// CreateSession creates a new Telegram session with the specified auth method.
func (s *SessionService) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	req *domain.CreateSessionRequest,
) (*domain.TelegramSession, string, error) {
	logger.Debug().
		Str("user_id", userID.String()).
		Str("auth_method", string(req.AuthMethod)).
		Str("session_name", req.SessionName).
		Msg("CreateSession started")

	if req.AuthMethod == domain.AuthMethodQR {
		return s.createSessionQR(ctx, userID, req)
	}
	return s.createSessionSMS(ctx, userID, req)
}

// validateSMSRequest validates SMS authentication request.
func (s *SessionService) validateSMSRequest(
	ctx context.Context,
	userID uuid.UUID,
	req *domain.CreateSessionRequest,
) error {
	if req.Phone == "" {
		logger.Warn().Msg("Empty phone in SMS auth")
		return domain.ErrInvalidPhoneNumber
	}

	existing, _ := s.sessionRepo.GetByUserAndPhone(ctx, userID, req.Phone)
	if existing != nil && existing.IsActive {
		logger.Warn().
			Str("phone", req.Phone).
			Str("existing_id", existing.ID.String()).
			Msg("Active session already exists for this number")
		return domain.ErrSessionAlreadyExists
	}

	return nil
}

// encryptAPIHash encrypts the API hash.
func (s *SessionService) encryptAPIHash(apiHash string) ([]byte, error) {
	logger.Debug().Msg("Encrypting api_hash...")
	apiHashEncrypted, err := s.tgManager.Encrypt([]byte(apiHash))
	if err != nil {
		logger.Error().Err(err).Msg("Error encrypting api_hash in SMS")
		return nil, domain.ErrInternal
	}
	logger.Debug().Msg("api_hash encrypted OK")
	return apiHashEncrypted, nil
}

// sendSMSCode sends SMS code to the phone number.
func (s *SessionService) sendSMSCode(ctx context.Context, apiID int, apiHash, phone string) (string, error) {
	logger.Debug().Str("phone", phone).Msg("Sending SMS code...")
	phoneCodeHash, err := s.tgManager.SendCode(ctx, apiID, apiHash, phone)
	if err != nil {
		logger.Error().Err(err).Str("phone", phone).Msg("Error sending SMS code")
		return "", domain.NewAppError(err, "Error sending code", 502)
	}
	logger.Debug().Msg("SMS code sent OK")
	return phoneCodeHash, nil
}

// createSMSSession creates and saves the SMS session.
func (s *SessionService) createSMSSession(
	ctx context.Context,
	userID uuid.UUID,
	req *domain.CreateSessionRequest,
	apiHashEncrypted []byte,
) (*domain.TelegramSession, error) {
	session := &domain.TelegramSession{
		ID:               uuid.New(),
		UserID:           userID,
		PhoneNumber:      req.Phone,
		APIID:            req.APIID,
		APIHashEncrypted: apiHashEncrypted,
		SessionName:      defaultSessionName(req.SessionName, req.Phone),
		AuthState:        domain.SessionCodeSent,
		IsActive:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	logger.Debug().
		Str("session_id", session.ID.String()).
		Str("session_name", session.SessionName).
		Msg("Saving session to DB...")

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		logger.Error().
			Err(err).
			Str("session_id", session.ID.String()).
			Str("session_name", session.SessionName).
			Msg("Error creating SMS session in DB")
		return nil, domain.ErrDatabase
	}
	logger.Debug().Msg("Session saved to DB OK")

	return session, nil
}

// createSessionSMS creates a session using SMS authentication.
func (s *SessionService) createSessionSMS(
	ctx context.Context,
	userID uuid.UUID,
	req *domain.CreateSessionRequest,
) (*domain.TelegramSession, string, error) {
	logger.Debug().Str("phone", req.Phone).Msg("Starting SMS auth...")

	if err := s.validateSMSRequest(ctx, userID, req); err != nil {
		return nil, "", err
	}

	apiHashEncrypted, err := s.encryptAPIHash(req.APIHash)
	if err != nil {
		return nil, "", err
	}

	phoneCodeHash, err := s.sendSMSCode(ctx, req.APIID, req.APIHash, req.Phone)
	if err != nil {
		return nil, "", err
	}

	session, err := s.createSMSSession(ctx, userID, req, apiHashEncrypted)
	if err != nil {
		return nil, "", err
	}

	_ = s.cache.Set(ctx, "tg:code:"+session.ID.String(), phoneCodeHash, 300)

	logger.Info().
		Str("session_id", session.ID.String()).
		Str("phone", req.Phone).
		Msg("SMS session created, code sent")

	return session, phoneCodeHash, nil
}

// createQRSession creates and saves the QR session.
func (s *SessionService) createQRSession(
	ctx context.Context,
	userID uuid.UUID,
	req *domain.CreateSessionRequest,
	apiHashEncrypted []byte,
) (*domain.TelegramSession, error) {
	sessionName := defaultSessionName(req.SessionName, "QR")

	session := &domain.TelegramSession{
		ID:               uuid.New(),
		UserID:           userID,
		PhoneNumber:      "QR-pending",
		APIID:            req.APIID,
		APIHashEncrypted: apiHashEncrypted,
		SessionName:      sessionName,
		AuthState:        domain.SessionPending,
		IsActive:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	logger.Debug().
		Str("session_id", session.ID.String()).
		Str("session_name", sessionName).
		Str("user_id", userID.String()).
		Msg("Saving QR session to DB...")

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		logger.Error().
			Err(err).
			Str("session_id", session.ID.String()).
			Str("session_name", sessionName).
			Str("user_id", userID.String()).
			Msg("Error creating QR session in DB")
		return nil, domain.ErrDatabase
	}
	logger.Debug().Str("session_id", session.ID.String()).Msg("QR session saved to DB OK")

	return session, nil
}

// startQRAuthProcess starts the QR authentication process.
func (s *SessionService) startQRAuthProcess(
	_ context.Context,
	sessionID uuid.UUID,
	apiID int,
	apiHash, sessionName string,
) (string, <-chan telegram.QRAuthResult, error) {
	logger.Debug().
		Int("api_id", apiID).
		Str("session_name", sessionName).
		Msg("Starting StartQRAuth...")

	qrImageB64, resultChan, err := s.tgManager.StartQRAuth(
		context.Background(),
		apiID,
		apiHash,
		sessionName,
		maxQRAttempts,
		qrTimeout,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Str("session_id", sessionID.String()).
			Str("session_name", sessionName).
			Int("api_id", apiID).
			Msg("Error starting QR auth")
		return "", nil, domain.NewAppError(err, "Error generating QR", 502)
	}
	logger.Debug().Int("qr_len", len(qrImageB64)).Msg("QR generated OK")

	return qrImageB64, resultChan, nil
}

// createSessionQR creates a session using QR authentication.
func (s *SessionService) createSessionQR(
	ctx context.Context,
	userID uuid.UUID,
	req *domain.CreateSessionRequest,
) (*domain.TelegramSession, string, error) {
	logger.Debug().
		Str("user_id", userID.String()).
		Str("session_name", req.SessionName).
		Int("api_id", req.APIID).
		Msg("Starting QR auth...")

	logger.Debug().Msg("Encrypting api_hash...")
	apiHashEncrypted, err := s.tgManager.Encrypt([]byte(req.APIHash))
	if err != nil {
		logger.Error().
			Err(err).
			Str("session_name", req.SessionName).
			Msg("Error encrypting api_hash in QR")
		return nil, "", domain.ErrInternal
	}
	logger.Debug().Int("encrypted_len", len(apiHashEncrypted)).Msg("api_hash encrypted OK")

	session, err := s.createQRSession(ctx, userID, req, apiHashEncrypted)
	if err != nil {
		return nil, "", err
	}

	qrImageB64, resultChan, err := s.startQRAuthProcess(ctx, session.ID, req.APIID, req.APIHash, session.SessionName)
	if err != nil {
		_ = s.sessionRepo.Delete(ctx, session.ID)
		return nil, "", err
	}

	go s.handleQRResult(ctx, session.ID, resultChan)

	logger.Info().
		Str("session_id", session.ID.String()).
		Str("session_name", session.SessionName).
		Msg("QR session started, waiting for scan in background...")

	return session, qrImageB64, nil
}

// VerifyCode verifies the SMS code sent to the user.
func (s *SessionService) VerifyCode(
	ctx context.Context,
	sessionID uuid.UUID,
	code string,
) (*domain.TelegramSession, string, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, "", domain.ErrSessionNotFound
	}

	cacheKey := "tg:code:" + sessionID.String()
	phoneCodeHash, err := s.cache.Get(ctx, cacheKey)
	if err != nil || phoneCodeHash == "" {
		return nil, "", domain.ErrCodeExpired
	}

	apiHashBytes, err := s.tgManager.Decrypt(session.APIHashEncrypted)
	if err != nil {
		return nil, "", domain.ErrInternal
	}

	user, sessionData, passwordHint, err := s.tgManager.SignIn(
		ctx,
		session.APIID,
		string(apiHashBytes),
		session.PhoneNumber,
		code,
		phoneCodeHash,
	)
	if err != nil {
		if err == domain.ErrPasswordRequired {
			session.AuthState = domain.SessionPasswordRequired
			session.UpdatedAt = time.Now()
			if err := s.sessionRepo.Update(ctx, session); err != nil {
				logger.Error().Err(err).Str("session_id", sessionID.String()).Msg("Error updating session state")
				return nil, "", domain.ErrDatabase
			}
			logger.Info().Str("session_id", sessionID.String()).Str("hint", passwordHint).Msg("2FA password required")
			return session, passwordHint, nil
		}
		logger.Error().Err(err).Str("session_id", sessionID.String()).Msg("Error verifying code")
		return nil, "", domain.ErrInvalidCode
	}

	updatedSession, err := s.completeAuth(ctx, session, user, sessionData, cacheKey)
	return updatedSession, passwordHint, err
}

// SubmitPassword submits 2FA password for accounts with 2FA enabled.
func (s *SessionService) SubmitPassword(
	ctx context.Context,
	sessionID uuid.UUID,
	password string,
) (*domain.TelegramSession, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	if session.AuthState != domain.SessionPasswordRequired {
		return nil, domain.ErrAlreadyAuthenticated
	}

	apiHashBytes, err := s.tgManager.Decrypt(session.APIHashEncrypted)
	if err != nil {
		logger.Error().Err(err).Str("session_id", sessionID.String()).Msg("Error decrypting api_hash")
		return nil, domain.ErrInternal
	}

	user, sessionData, err := s.tgManager.SubmitPassword(
		ctx,
		session.ID.String(),
		session.APIID,
		string(apiHashBytes),
		password,
	)
	if err != nil {
		logger.Error().Err(err).Str("session_id", sessionID.String()).Msg("Error submitting password")
		return nil, domain.ErrInvalidPassword
	}

	cacheKey := "tg:code:" + sessionID.String()
	return s.completeAuth(ctx, session, user, sessionData, cacheKey)
}

// RegenerateQR generates a new QR for an existing session.
func (s *SessionService) RegenerateQR(ctx context.Context, sessionID uuid.UUID) (string, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return "", domain.ErrSessionNotFound
	}

	if session.IsActive && session.AuthState == domain.SessionAuthenticated {
		return "", domain.NewAppError(nil, "Session already authenticated", 400)
	}

	apiHashBytes, err := s.tgManager.Decrypt(session.APIHashEncrypted)
	if err != nil {
		logger.Error().Err(err).Str("session_id", sessionID.String()).Msg("Error decrypting api_hash")
		return "", domain.ErrInternal
	}

	session.AuthState = domain.SessionPending
	session.IsActive = false
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return "", domain.ErrDatabase
	}

	qrImageB64, resultChan, err := s.tgManager.StartQRAuth(
		context.Background(),
		session.APIID,
		string(apiHashBytes),
		session.SessionName,
		maxQRAttempts,
		qrTimeout,
	)
	if err != nil {
		logger.Error().Err(err).Str("session_id", sessionID.String()).Msg("Error regenerating QR")
		session.AuthState = domain.SessionFailed
		_ = s.sessionRepo.Update(ctx, session)
		return "", domain.NewAppError(err, "Error generating QR", 502)
	}

	go s.handleQRResult(ctx, session.ID, resultChan)

	logger.Info().
		Str("session_id", sessionID.String()).
		Str("session_name", session.SessionName).
		Msg("QR regenerated, waiting for scan...")

	return qrImageB64, nil
}
