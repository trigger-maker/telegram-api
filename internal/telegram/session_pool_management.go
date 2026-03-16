package telegram

import (
	"context"
	"fmt"
	"time"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// StartSession starts a session and begins listening for events.
func (p *SessionPool) StartSession(_ context.Context, sess *domain.TelegramSession) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.sessions[sess.ID]; exists {
		return nil
	}

	apiHashBytes, err := p.manager.Decrypt(sess.APIHashEncrypted)
	if err != nil {
		return fmt.Errorf("decrypt api_hash: %w", err)
	}

	storage := NewPersistentSessionStorage(p.manager.crypter, p.repo, sess.ID.String())

	dispatcher := tg.NewUpdateDispatcher()
	client := telegram.NewClient(sess.APIID, string(apiHashBytes), telegram.Options{
		SessionStorage: storage,
		UpdateHandler:  dispatcher,
		Device: telegram.DeviceConfig{
			DeviceModel:    sess.SessionName,
			SystemVersion:  "1.0",
			AppVersion:     "1.0.0",
			SystemLangCode: "es",
			LangCode:       "es",
		},
	})

	// #nosec G118 -- Cancel function is stored in active.Cancel and called later
	sessionCtx, cancel := context.WithCancel(context.Background())

	active := &ActiveSession{
		SessionID:   sess.ID,
		SessionName: sess.SessionName,
		TelegramID:  sess.TelegramUserID,
		Client:      client,
		Cancel:      cancel,
		StartedAt:   time.Now(),
		IsConnected: false,
	}

	p.registerHandlers(dispatcher, active)
	go p.runClient(sessionCtx, active, client)

	p.sessions[sess.ID] = active

	p.dispatcher.Dispatch(sess.ID, domain.EventSessionStarted, domain.SessionEventData{
		SessionID:   sess.ID,
		SessionName: sess.SessionName,
		TelegramID:  sess.TelegramUserID,
	})

	logger.Info().
		Str("session_id", sess.ID.String()).
		Str("session_name", sess.SessionName).
		Msg("Session started in pool")

	return nil
}

// StopSession stops a session.
func (p *SessionPool) StopSession(sessionID uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if active, exists := p.sessions[sessionID]; exists {
		active.Cancel()
		delete(p.sessions, sessionID)

		p.dispatcher.Dispatch(sessionID, domain.EventSessionStopped, domain.SessionEventData{
			SessionID:   sessionID,
			SessionName: active.SessionName,
		})

		logger.Info().
			Str("session_id", sessionID.String()).
			Msg("Session stopped")
	}
}

// GetActiveSession retrieves an active session.
func (p *SessionPool) GetActiveSession(sessionID uuid.UUID) (*ActiveSession, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	active, exists := p.sessions[sessionID]
	return active, exists
}

// ActiveCount returns the number of active sessions.
func (p *SessionPool) ActiveCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.sessions)
}

// ListActive returns IDs of active sessions.
func (p *SessionPool) ListActive() []uuid.UUID {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ids := make([]uuid.UUID, 0, len(p.sessions))
	for id := range p.sessions {
		ids = append(ids, id)
	}
	return ids
}

// StartAllActive starts all active sessions from the database.
func (p *SessionPool) StartAllActive(ctx context.Context) error {
	sessions, err := p.repo.ListAllActive(ctx)
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		if err := p.StartSession(ctx, &sess); err != nil {
			logger.Error().Err(err).
				Str("session_id", sess.ID.String()).
				Msg("Error starting session")
		}
	}

	return nil
}
