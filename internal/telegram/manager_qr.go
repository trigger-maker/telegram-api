package telegram

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"
	"telegram-api/pkg/utils"

	"github.com/google/uuid"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// handleTokenSuccess handles successful token authentication.
func handleTokenSuccess(auth *tg.AuthAuthorization) *TGUser {
	if u, ok := auth.User.(*tg.User); ok {
		return &TGUser{ID: u.ID, Username: u.Username}
	}
	return nil
}

// handleTokenMigrate handles token migration to different DC.
func handleTokenMigrate(
	ctx context.Context,
	client *telegram.Client,
	token *tg.AuthLoginTokenMigrateTo,
) (*TGUser, error) {
	logger.Info().Int("dc", token.DCID).Msg("QR scanned, migrating to DC...")

	if err := client.MigrateTo(ctx, token.DCID); err != nil {
		logger.Error().Err(err).Int("dc", token.DCID).Msg("Error migrating to DC")
		return nil, err
	}

	res, err := client.API().AuthImportLoginToken(ctx, token.Token)
	if err != nil {
		logger.Error().Err(err).Msg("Error importing token")
		return nil, err
	}

	success, ok := res.(*tg.AuthLoginTokenSuccess)
	if !ok {
		logger.Warn().Msgf("Unexpected type: %T", res)
		return nil, fmt.Errorf("unexpected token type")
	}

	auth, ok := success.Authorization.(*tg.AuthAuthorization)
	if !ok {
		return nil, fmt.Errorf("unexpected auth type")
	}

	u, ok := auth.User.(*tg.User)
	if !ok {
		return nil, fmt.Errorf("unexpected user type")
	}

	logger.Info().
		Int64("user_id", u.ID).
		Str("username", u.Username).
		Msg("✅ DC migration successful, user authenticated")

	return &TGUser{ID: u.ID, Username: u.Username}, nil
}

// generateQR generates QR code image from token.
func generateQR(token *tg.AuthLoginToken) string {
	tokenB64 := base64.URLEncoding.EncodeToString(token.Token)
	url := "tg://login?token=" + tokenB64
	qrImg, _ := utils.GenerateQRBase64(url)
	return qrImg
}

// sendFirstQR sends first QR code to channel if available.
func sendFirstQR(firstQR chan<- string, qrImg string) {
	select {
	case firstQR <- qrImg:
	default:
	}
}

// handleTokenSuccessInLoop handles successful token in auth loop.
func handleTokenSuccessInLoop(token *tg.AuthLoginTokenSuccess, result chan<- QRAuthResult) bool {
	auth, ok := token.Authorization.(*tg.AuthAuthorization)
	if !ok {
		return false
	}
	if u, ok := auth.User.(*tg.User); ok {
		result <- QRAuthResult{
			User: &TGUser{ID: u.ID, Username: u.Username},
		}
		return true
	}
	return false
}

// handleLoginTokenInLoop handles login token in auth loop.
func handleLoginTokenInLoop(
	ctx context.Context,
	m *ClientManager,
	client *telegram.Client,
	apiID int,
	apiHash string,
	storage *PersistentSessionStorage,
	token *tg.AuthLoginToken,
	sessionName string,
	attempt int,
	maxAttempts int,
	qrTimeout time.Duration,
	firstQR chan<- string,
	result chan<- QRAuthResult,
) bool {
	qrImg := generateQR(token)

	logger.Info().
		Str("session_name", sessionName).
		Int("attempt", attempt).
		Int("max", maxAttempts).
		Msg("QR generated, waiting for scan...")

	if attempt == 1 {
		sendFirstQR(firstQR, qrImg)
	}

	if user, sessionData, ok := m.waitForScan(ctx, client, apiID, apiHash, storage, qrTimeout); ok {
		result <- QRAuthResult{User: user, SessionData: sessionData}
		return true
	}

	logger.Info().
		Str("session_name", sessionName).
		Int("attempt", attempt).
		Msg("QR expired, generating new...")

	return false
}

// processAuthToken processes authentication token.
func processAuthToken(
	ctx context.Context,
	m *ClientManager,
	client *telegram.Client,
	apiID int,
	apiHash string,
	storage *PersistentSessionStorage,
	token interface{},
	sessionName string,
	attempt int,
	maxAttempts int,
	qrTimeout time.Duration,
	firstQR chan<- string,
	result chan<- QRAuthResult,
) bool {
	switch t := token.(type) {
	case *tg.AuthLoginTokenSuccess:
		return handleTokenSuccessInLoop(t, result)
	case *tg.AuthLoginToken:
		return handleLoginTokenInLoop(
			ctx, m, client, apiID, apiHash, storage, t,
			sessionName, attempt, maxAttempts, qrTimeout,
			firstQR, result,
		)
	}
	return false
}

// exportLoginToken exports login token with error handling.
func exportLoginToken(
	ctx context.Context,
	client *telegram.Client,
	apiID int,
	apiHash string,
	sessionID uuid.UUID,
	repo domain.SessionRepository,
	attempt int,
	errChan chan<- error,
) (interface{}, error) {
	token, err := client.API().AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{
		APIID:     apiID,
		APIHash:   apiHash,
		ExceptIDs: []int64{},
	})
	if err != nil {
		if attempt == 1 {
			action, _, wrappedErr := HandleMTProtoError(ctx, sessionID, err, repo)
			if action == ActionStop {
				errChan <- wrappedErr
				return nil, wrappedErr
			}
			errChan <- fmt.Errorf("export token: %w", err)
			return nil, err
		}
		return nil, err
	}
	return token, nil
}

// runQRAuthLoop runs QR authentication loop.
func runQRAuthLoop(
	ctx context.Context,
	m *ClientManager,
	client *telegram.Client,
	apiID int,
	apiHash string,
	storage *PersistentSessionStorage,
	sessionID uuid.UUID,
	repo domain.SessionRepository,
	sessionName string,
	maxAttempts int,
	qrTimeout time.Duration,
	firstQR chan<- string,
	result chan<- QRAuthResult,
	errChan chan<- error,
) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		token, err := exportLoginToken(ctx, client, apiID, apiHash, sessionID, repo, attempt, errChan)
		if err != nil {
			continue
		}

		if processAuthToken(
			ctx, m, client, apiID, apiHash, storage, token,
			sessionName, attempt, maxAttempts, qrTimeout,
			firstQR, result,
		) {
			return nil
		}
	}

	result <- QRAuthResult{Error: fmt.Errorf("max QR attempts reached")}
	return nil
}

// handleLoginTokenInLoopWithSession handles login token in auth loop with session.
func handleLoginTokenInLoopWithSession(
	ctx context.Context,
	m *ClientManager,
	client *telegram.Client,
	apiID int,
	apiHash string,
	storage *PersistentSessionStorage,
	token *tg.AuthLoginToken,
	sessionName string,
	attempt int,
	maxAttempts int,
	qrTimeout time.Duration,
	firstQR chan<- string,
	result chan<- QRAuthResult,
) bool {
	qrImg := generateQR(token)

	logger.Info().
		Str("session_name", sessionName).
		Int("attempt", attempt).
		Int("max", maxAttempts).
		Msg("QR generated, waiting for scan...")

	if attempt == 1 {
		sendFirstQR(firstQR, qrImg)
	}

	if user, ok := m.waitForScanWithSession(ctx, client, apiID, apiHash, storage, qrTimeout); ok {
		result <- QRAuthResult{User: user}
		return true
	}

	logger.Info().
		Str("session_name", sessionName).
		Int("attempt", attempt).
		Msg("QR expired, generating new...")

	return false
}

// processAuthTokenWithSession processes authentication token with session.
func processAuthTokenWithSession(
	ctx context.Context,
	m *ClientManager,
	client *telegram.Client,
	apiID int,
	apiHash string,
	storage *PersistentSessionStorage,
	token interface{},
	sessionName string,
	attempt int,
	maxAttempts int,
	qrTimeout time.Duration,
	firstQR chan<- string,
	result chan<- QRAuthResult,
) bool {
	switch t := token.(type) {
	case *tg.AuthLoginTokenSuccess:
		return handleTokenSuccessInLoop(t, result)
	case *tg.AuthLoginToken:
		return handleLoginTokenInLoopWithSession(
			ctx, m, client, apiID, apiHash, storage, t,
			sessionName, attempt, maxAttempts, qrTimeout,
			firstQR, result,
		)
	}
	return false
}

// runQRAuthLoopWithSession runs QR authentication loop with session.
func runQRAuthLoopWithSession(
	ctx context.Context,
	m *ClientManager,
	client *telegram.Client,
	apiID int,
	apiHash string,
	storage *PersistentSessionStorage,
	sessionID uuid.UUID,
	repo domain.SessionRepository,
	sessionName string,
	maxAttempts int,
	qrTimeout time.Duration,
	firstQR chan<- string,
	result chan<- QRAuthResult,
	errChan chan<- error,
) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		token, err := exportLoginToken(ctx, client, apiID, apiHash, sessionID, repo, attempt, errChan)
		if err != nil {
			continue
		}

		if processAuthTokenWithSession(
			ctx, m, client, apiID, apiHash, storage, token,
			sessionName, attempt, maxAttempts, qrTimeout,
			firstQR, result,
		) {
			return nil
		}
	}

	result <- QRAuthResult{Error: fmt.Errorf("max QR attempts reached")}
	return nil
}

// StartQRAuth starts QR code authentication process.
func (m *ClientManager) StartQRAuth(
	ctx context.Context,
	apiID int,
	apiHash string,
	sessionName string,
	maxAttempts int,
	qrTimeout time.Duration,
) (qrImageB64 string, resultChan <-chan QRAuthResult, err error) {

	result := make(chan QRAuthResult, 1)
	firstQR := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(result)

		sessionID := uuid.New()
		storage := NewPersistentSessionStorage(m.crypter, m.repo, sessionID.String())
		client := m.newClient(apiID, apiHash, sessionName, storage)

		runErr := client.Run(ctx, func(ctx context.Context) error {
			return runQRAuthLoop(
				ctx, m, client, apiID, apiHash, storage,
				sessionID, m.repo, sessionName, maxAttempts,
				qrTimeout, firstQR, result, errChan,
			)
		})

		if runErr != nil && ctx.Err() == nil {
			result <- QRAuthResult{Error: runErr}
		}
	}()

	select {
	case qr := <-firstQR:
		return qr, result, nil
	case err := <-errChan:
		return "", nil, err
	case <-ctx.Done():
		return "", nil, ctx.Err()
	case <-time.After(15 * time.Second):
		return "", nil, fmt.Errorf("timeout generating first QR")
	}
}

// StartQRAuthWithSession starts QR code authentication with persistent session storage.
func (m *ClientManager) StartQRAuthWithSession(
	ctx context.Context,
	sessionID string,
	apiID int,
	apiHash string,
	sessionName string,
	maxAttempts int,
	qrTimeout time.Duration,
) (qrImageB64 string, resultChan <-chan QRAuthResult, err error) {

	result := make(chan QRAuthResult, 1)
	firstQR := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(result)

		storage := NewPersistentSessionStorage(m.crypter, m.repo, sessionID)
		client := m.newClient(apiID, apiHash, sessionName, storage)

		sessionUUID, _ := uuid.Parse(sessionID)

		runErr := client.Run(ctx, func(ctx context.Context) error {
			return runQRAuthLoopWithSession(
				ctx, m, client, apiID, apiHash, storage,
				sessionUUID, m.repo, sessionName, maxAttempts,
				qrTimeout, firstQR, result, errChan,
			)
		})

		if runErr != nil && ctx.Err() == nil {
			result <- QRAuthResult{Error: runErr}
		}
	}()

	select {
	case qr := <-firstQR:
		return qr, result, nil
	case err := <-errChan:
		return "", nil, err
	case <-ctx.Done():
		return "", nil, ctx.Err()
	case <-time.After(15 * time.Second):
		return "", nil, fmt.Errorf("timeout generating first QR")
	}
}

// waitForScan waits for the QR code to be scanned.
func (m *ClientManager) waitForScan(
	ctx context.Context,
	client *telegram.Client,
	apiID int,
	apiHash string,
	_ *PersistentSessionStorage,
	timeout time.Duration,
) (*TGUser, []byte, bool) {

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, nil, false
		case <-ticker.C:
			token, err := client.API().AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{
				APIID:     apiID,
				APIHash:   apiHash,
				ExceptIDs: []int64{},
			})
			if err != nil {
				continue
			}

			switch t := token.(type) {
			case *tg.AuthLoginTokenSuccess:
				auth, ok := t.Authorization.(*tg.AuthAuthorization)
				if !ok {
					continue
				}
				if user := handleTokenSuccess(auth); user != nil {
					logger.Info().
						Int64("user_id", user.ID).
						Str("username", user.Username).
						Msg("✅ QR scanned successfully")
					return user, nil, true
				}

			case *tg.AuthLoginTokenMigrateTo:
				user, err := handleTokenMigrate(ctx, client, t)
				if err != nil {
					continue
				}
				return user, nil, true

			case *tg.AuthLoginToken:
				continue
			}
		}
	}

	return nil, nil, false
}

// waitForScanWithSession waits for the QR code to be scanned with persistent session storage.
func (m *ClientManager) waitForScanWithSession(
	ctx context.Context,
	client *telegram.Client,
	apiID int,
	apiHash string,
	_ *PersistentSessionStorage,
	timeout time.Duration,
) (*TGUser, bool) {

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, false
		case <-ticker.C:
			token, err := client.API().AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{
				APIID:     apiID,
				APIHash:   apiHash,
				ExceptIDs: []int64{},
			})
			if err != nil {
				continue
			}

			switch t := token.(type) {
			case *tg.AuthLoginTokenSuccess:
				auth, ok := t.Authorization.(*tg.AuthAuthorization)
				if !ok {
					continue
				}
				if user := handleTokenSuccess(auth); user != nil {
					logger.Info().
						Int64("user_id", user.ID).
						Str("username", user.Username).
						Msg("✅ QR scanned successfully")
					return user, true
				}

			case *tg.AuthLoginTokenMigrateTo:
				user, err := handleTokenMigrate(ctx, client, t)
				if err != nil {
					continue
				}
				return user, true

			case *tg.AuthLoginToken:
				continue
			}
		}
	}

	return nil, false
}
