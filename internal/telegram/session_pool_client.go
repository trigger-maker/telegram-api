package telegram

import (
	"context"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/gotd/td/telegram"
)

// runClient runs the Telegram client and manages connection state
func (p *SessionPool) runClient(ctx context.Context, active *ActiveSession, client *telegram.Client) {
	err := client.Run(ctx, func(ctx context.Context) error {
		active.mu.Lock()
		active.API = client.API()
		active.IsConnected = true
		active.mu.Unlock()

		logger.Info().
			Str("session_id", active.SessionID.String()).
			Msg("Telegram client connected, listening for events...")

		<-ctx.Done()
		return ctx.Err()
	})

	active.mu.Lock()
	active.IsConnected = false
	active.mu.Unlock()

	if err != nil && ctx.Err() == nil {
		action, _, wrappedErr := HandleMTProtoError(ctx, active.SessionID, err, p.repo)
		logger.Error().Err(err).
			Str("session_id", active.SessionID.String()).
			Msg("Telegram client disconnected with error")

		p.dispatcher.Dispatch(active.SessionID, domain.EventSessionError, domain.SessionEventData{
			SessionID: active.SessionID,
			Error:     err.Error(),
		})

		if action == ActionStop {
			return
		}
	}
}
