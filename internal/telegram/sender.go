package telegram

import (
	"context"
	"fmt"

	"telegram-api/internal/domain"

	"github.com/google/uuid"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

// SendMessage sends a message using the session.
func (m *ClientManager) SendMessage(
	ctx context.Context,
	sess *domain.TelegramSession,
	req *domain.SendMessageRequest,
) error {
	if m.pool == nil {
		return domain.ErrSessionNotActive
	}

	active, ok := m.pool.GetActiveSession(sess.ID)
	if !ok {
		return domain.ErrSessionNotActive
	}

	return m.SendMessageWithAPIClient(ctx, active.API, req)
}

// SendMessageWithAPIClient sends a message using the Telegram API client.
func (m *ClientManager) SendMessageWithAPIClient(
	ctx context.Context,
	api *tg.Client,
	req *domain.SendMessageRequest,
) error {
	if api == nil {
		return domain.ErrSessionNotActive
	}

	sender := message.NewSender(api)

	peer, err := m.resolvePeer(ctx, api, req.To)
	if err != nil {
		return fmt.Errorf("resolve peer: %w", err)
	}

	builder := sender.To(peer)

	switch req.Type {
	case domain.MessageTypeText, "":
		_, err = builder.Text(ctx, req.Text)

	case domain.MessageTypePhoto:
		err = m.sendPhoto(ctx, api, builder, req)

	case domain.MessageTypeVideo:
		err = m.sendVideo(ctx, api, builder, req)

	case domain.MessageTypeAudio:
		err = m.sendAudio(ctx, api, builder, req)

	case domain.MessageTypeFile:
		err = m.sendFile(ctx, api, builder, req)

	default:
		_, err = builder.Text(ctx, req.Text)
	}

	if err != nil {
		action, _, wrappedErr := HandleMTProtoError(ctx, uuid.Nil, err, m.repo)
		if action == ActionStop {
			return wrappedErr
		}
	}

	return err
}
