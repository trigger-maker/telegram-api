package telegram

import (
	"context"
	"fmt"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// SendCode sends a verification code to the given phone number
func (m *ClientManager) SendCode(ctx context.Context, apiID int, apiHash, phone string) (string, error) {
	sessionID := uuid.New()
	storage := NewPersistentSessionStorage(m.crypter, m.repo, sessionID.String())
	client := m.newClient(apiID, apiHash, "SMS Auth", storage)

	var phoneCodeHash string

	err := client.Run(ctx, func(ctx context.Context) error {
		sent, err := client.API().AuthSendCode(ctx, &tg.AuthSendCodeRequest{
			PhoneNumber: phone,
			APIID:       apiID,
			APIHash:     apiHash,
			Settings:    tg.CodeSettings{},
		})
		if err != nil {
			return fmt.Errorf("send code: %w", err)
		}

		sc, ok := sent.(*tg.AuthSentCode)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}

		phoneCodeHash = sc.PhoneCodeHash
		return nil
	})

	if err != nil {
		action, _, wrappedErr := HandleMTProtoError(ctx, sessionID, err, m.repo)
		if action == ActionStop {
			return "", wrappedErr
		}
	}

	return phoneCodeHash, err
}

// SignIn signs in the user with the given phone number and code
func (m *ClientManager) SignIn(ctx context.Context, apiID int, apiHash, phone, code, codeHash string) (*TGUser, []byte, string, error) {
	sessionID := uuid.New()
	storage := NewPersistentSessionStorage(m.crypter, m.repo, sessionID.String())
	client := m.newClient(apiID, apiHash, "SMS Session", storage)

	var user *TGUser
	var sessionData []byte
	var passwordHint string

	err := client.Run(ctx, func(ctx context.Context) error {
		auth, err := client.API().AuthSignIn(ctx, &tg.AuthSignInRequest{
			PhoneNumber:   phone,
			PhoneCodeHash: codeHash,
			PhoneCode:     code,
		})
		if err != nil {
			if err.Error() == "SESSION_PASSWORD_NEEDED" {
				passwordInfo, pErr := client.API().AccountGetPassword(ctx)
				if pErr != nil {
					passwordHint = ""
				} else {
					passwordHint = passwordInfo.Hint
				}
				return err
			}
			return err
		}

		a, ok := auth.(*tg.AuthAuthorization)
		if !ok {
			return fmt.Errorf("unexpected auth response")
		}

		u, ok := a.User.(*tg.User)
		if !ok {
			return fmt.Errorf("unexpected user type")
		}
		user = &TGUser{ID: u.ID, Username: u.Username}

		data, err := storage.Bytes(nil)
		if err == nil {
			sessionData = data
		}

		return nil
	})

	if err != nil {
		if err.Error() == "SESSION_PASSWORD_NEEDED" {
			return user, sessionData, passwordHint, domain.ErrPasswordRequired
		}
		action, _, wrappedErr := HandleMTProtoError(ctx, sessionID, err, m.repo)
		if action == ActionStop {
			return user, sessionData, "", wrappedErr
		}
		return user, sessionData, "", err
	}

	return user, sessionData, "", nil
}

// SubmitPassword submits the 2FA password for the session
func (m *ClientManager) SubmitPassword(ctx context.Context, sessionID string, apiID int, apiHash, password string) (*TGUser, []byte, error) {
	storage := NewPersistentSessionStorage(m.crypter, m.repo, sessionID)
	client := m.newClient(apiID, apiHash, "SMS Session", storage)

	var user *TGUser

	err := client.Run(ctx, func(ctx context.Context) error {
		accountPassword, err := client.API().AccountGetPassword(ctx)
		if err != nil {
			return err
		}

		srp, err := telegram.SolvePassword(password, accountPassword.SRPID, accountPassword.SRPB, accountPassword.CurrentAlgo)
		if err != nil {
			return fmt.Errorf("solve password: %w", err)
		}

		auth, err := client.API().AuthCheckPassword(ctx, &tg.InputCheckPasswordSRP{
			SRPID: accountPassword.SRPID,
			A:     srp.A,
			M1:    srp.M1,
		})
		if err != nil {
			return err
		}

		a, ok := auth.(*tg.AuthAuthorization)
		if !ok {
			return fmt.Errorf("unexpected auth response")
		}

		u, ok := a.User.(*tg.User)
		if !ok {
			return fmt.Errorf("unexpected user type")
		}
		user = &TGUser{ID: u.ID, Username: u.Username}

		return nil
	})

	if err != nil {
		return nil, nil, domain.ErrInvalidPassword
	}

	return user, nil, nil
}

// SignInWithSession signs in the user with the given phone number and code using persistent session storage
func (m *ClientManager) SignInWithSession(ctx context.Context, sessionID string, apiID int, apiHash, phone, code, codeHash string) (*TGUser, error) {
	storage := NewPersistentSessionStorage(m.crypter, m.repo, sessionID)
	client := m.newClient(apiID, apiHash, "SMS Session", storage)

	var user *TGUser

	err := client.Run(ctx, func(ctx context.Context) error {
		auth, err := client.API().AuthSignIn(ctx, &tg.AuthSignInRequest{
			PhoneNumber:   phone,
			PhoneCodeHash: codeHash,
			PhoneCode:     code,
		})
		if err != nil {
			return err
		}

		a, ok := auth.(*tg.AuthAuthorization)
		if !ok {
			return fmt.Errorf("unexpected auth response")
		}

		u, ok := a.User.(*tg.User)
		if !ok {
			return fmt.Errorf("unexpected user type")
		}
		user = &TGUser{ID: u.ID, Username: u.Username}

		return nil
	})

	return user, err
}

// LogOut logs out the user from Telegram
func (m *ClientManager) LogOut(ctx context.Context, apiID int, apiHashEncrypted, sessionData []byte, sessionName string) error {
	if len(sessionData) == 0 {
		return nil
	}

	apiHashBytes, err := m.Decrypt(apiHashEncrypted)
	if err != nil {
		return fmt.Errorf("decrypt api_hash: %w", err)
	}

	decryptedSession, err := m.Decrypt(sessionData)
	if err != nil {
		return fmt.Errorf("decrypt session: %w", err)
	}

	sessionID := uuid.New().String()
	storage := NewPersistentSessionStorage(m.crypter, m.repo, sessionID)
	if err := storage.StoreSession(ctx, decryptedSession); err != nil {
		return fmt.Errorf("store session: %w", err)
	}

	client := m.newClient(apiID, string(apiHashBytes), sessionName, storage)

	err = client.Run(ctx, func(ctx context.Context) error {
		_, err := client.API().AuthLogOut(ctx)
		return err
	})

	if err != nil {
		logger.Warn().Err(err).Msg("Error during logout")
	} else {
		logger.Info().Str("session_name", sessionName).Msg("Session closed in Telegram")
	}
	return nil
}
