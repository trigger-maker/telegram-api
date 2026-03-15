package telegram

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"telegram-api/pkg/logger"
	"telegram-api/pkg/utils"

	"github.com/google/uuid"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// StartQRAuth starts QR code authentication process
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
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				token, err := client.API().AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{
					APIID:     apiID,
					APIHash:   apiHash,
					ExceptIDs: []int64{},
				})
				if err != nil {
					if attempt == 1 {
						action, _, wrappedErr := HandleMTProtoError(ctx, sessionID, err, m.repo)
						if action == ActionStop {
							errChan <- wrappedErr
							return wrappedErr
						}
						errChan <- fmt.Errorf("export token: %w", err)
						return err
					}
					continue
				}

				switch t := token.(type) {
				case *tg.AuthLoginTokenSuccess:
					auth, ok := t.Authorization.(*tg.AuthAuthorization)
					if ok {
						if u, ok := auth.User.(*tg.User); ok {
							result <- QRAuthResult{
								User: &TGUser{ID: u.ID, Username: u.Username},
							}
							return nil
						}
					}

				case *tg.AuthLoginToken:
					tokenB64 := base64.URLEncoding.EncodeToString(t.Token)
					url := "tg://login?token=" + tokenB64

					qrImg, _ := utils.GenerateQRBase64(url)

					logger.Info().
						Str("session_name", sessionName).
						Int("attempt", attempt).
						Int("max", maxAttempts).
						Msg("QR generated, waiting for scan...")

					if attempt == 1 {
						select {
						case firstQR <- qrImg:
						default:
						}
					}

					if user, sessionData, ok := m.waitForScan(ctx, client, apiID, apiHash, storage, qrTimeout); ok {
						result <- QRAuthResult{User: user, SessionData: sessionData}
						return nil
					}

					logger.Info().
						Str("session_name", sessionName).
						Int("attempt", attempt).
						Msg("QR expired, generating new...")
				}
			}

			result <- QRAuthResult{Error: fmt.Errorf("max QR attempts reached")}
			return nil
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

// StartQRAuthWithSession starts QR code authentication with persistent session storage
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

		runErr := client.Run(ctx, func(ctx context.Context) error {
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				token, err := client.API().AuthExportLoginToken(ctx, &tg.AuthExportLoginTokenRequest{
					APIID:     apiID,
					APIHash:   apiHash,
					ExceptIDs: []int64{},
				})
				if err != nil {
					if attempt == 1 {
						sessionUUID, _ := uuid.Parse(sessionID)
						action, _, wrappedErr := HandleMTProtoError(ctx, sessionUUID, err, m.repo)
						if action == ActionStop {
							errChan <- wrappedErr
							return wrappedErr
						}
						errChan <- fmt.Errorf("export token: %w", err)
						return err
					}
					continue
				}

				switch t := token.(type) {
				case *tg.AuthLoginTokenSuccess:
					auth, ok := t.Authorization.(*tg.AuthAuthorization)
					if ok {
						if u, ok := auth.User.(*tg.User); ok {
							result <- QRAuthResult{
								User: &TGUser{ID: u.ID, Username: u.Username},
							}
							return nil
						}
					}

				case *tg.AuthLoginToken:
					tokenB64 := base64.URLEncoding.EncodeToString(t.Token)
					url := "tg://login?token=" + tokenB64

					qrImg, _ := utils.GenerateQRBase64(url)

					logger.Info().
						Str("session_name", sessionName).
						Int("attempt", attempt).
						Int("max", maxAttempts).
						Msg("QR generated, waiting for scan...")

					if attempt == 1 {
						select {
						case firstQR <- qrImg:
						default:
						}
					}

					if user, ok := m.waitForScanWithSession(ctx, client, apiID, apiHash, storage, qrTimeout); ok {
						result <- QRAuthResult{User: user}
						return nil
					}

					logger.Info().
						Str("session_name", sessionName).
						Int("attempt", attempt).
						Msg("QR expired, generating new...")
				}
			}

			result <- QRAuthResult{Error: fmt.Errorf("max QR attempts reached")}
			return nil
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

// waitForScan waits for the QR code to be scanned
func (m *ClientManager) waitForScan(
	ctx context.Context,
	client *telegram.Client,
	apiID int,
	apiHash string,
	storage *PersistentSessionStorage,
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
				u, ok := auth.User.(*tg.User)
				if !ok {
					continue
				}
				logger.Info().
					Int64("user_id", u.ID).
					Str("username", u.Username).
					Msg("✅ QR scanned successfully")
				return &TGUser{ID: u.ID, Username: u.Username}, nil, true

			case *tg.AuthLoginTokenMigrateTo:
				logger.Info().Int("dc", t.DCID).Msg("QR scanned, migrating to DC...")

				if err := client.MigrateTo(ctx, t.DCID); err != nil {
					logger.Error().Err(err).Int("dc", t.DCID).Msg("Error migrating to DC")
					continue
				}

				res, err := client.API().AuthImportLoginToken(ctx, t.Token)
				if err != nil {
					logger.Error().Err(err).Msg("Error importing token")
					continue
				}

				success, ok := res.(*tg.AuthLoginTokenSuccess)
				if !ok {
					logger.Warn().Msgf("Unexpected type: %T", res)
					continue
				}

				auth, ok := success.Authorization.(*tg.AuthAuthorization)
				if !ok {
					continue
				}

				u, ok := auth.User.(*tg.User)
				if !ok {
					continue
				}
				logger.Info().
					Int64("user_id", u.ID).
					Str("username", u.Username).
					Msg("✅ DC migration successful, user authenticated")
				return &TGUser{ID: u.ID, Username: u.Username}, nil, true

			case *tg.AuthLoginToken:
				continue
			}
		}
	}

	return nil, nil, false
}

// waitForScanWithSession waits for the QR code to be scanned with persistent session storage
func (m *ClientManager) waitForScanWithSession(
	ctx context.Context,
	client *telegram.Client,
	apiID int,
	apiHash string,
	storage *PersistentSessionStorage,
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
				u, ok := auth.User.(*tg.User)
				if !ok {
					continue
				}
				logger.Info().
					Int64("user_id", u.ID).
					Str("username", u.Username).
					Msg("✅ QR scanned successfully")
				return &TGUser{ID: u.ID, Username: u.Username}, true

			case *tg.AuthLoginTokenMigrateTo:
				logger.Info().Int("dc", t.DCID).Msg("QR scanned, migrating to DC...")

				if err := client.MigrateTo(ctx, t.DCID); err != nil {
					logger.Error().Err(err).Int("dc", t.DCID).Msg("Error migrating to DC")
					continue
				}

				res, err := client.API().AuthImportLoginToken(ctx, t.Token)
				if err != nil {
					logger.Error().Err(err).Msg("Error importing token")
					continue
				}

				success, ok := res.(*tg.AuthLoginTokenSuccess)
				if !ok {
					logger.Warn().Msgf("Unexpected type: %T", res)
					continue
				}

				auth, ok := success.Authorization.(*tg.AuthAuthorization)
				if !ok {
					continue
				}

				u, ok := auth.User.(*tg.User)
				if !ok {
					continue
				}

				logger.Info().
					Int64("user_id", u.ID).
					Str("username", u.Username).
					Msg("✅ DC migration successful, user authenticated")
				return &TGUser{ID: u.ID, Username: u.Username}, true

			case *tg.AuthLoginToken:
				continue
			}
		}
	}

	return nil, false
}
