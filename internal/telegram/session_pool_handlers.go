package telegram

import (
	"context"
	"time"

	"telegram-api/internal/domain"

	"github.com/gotd/td/tg"
)

// registerHandlers registers event handlers for the session.
func (p *SessionPool) registerHandlers(dispatcher tg.UpdateDispatcher, active *ActiveSession) {
	dispatcher.OnNewMessage(func(_ context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok || msg.Out {
			return nil
		}

		active.mu.Lock()
		active.LastActivity = time.Now()
		active.mu.Unlock()

		data := p.parseMessage(e, msg)
		p.dispatcher.Dispatch(active.SessionID, domain.EventNewMessage, data)

		return nil
	})

	dispatcher.OnEditMessage(func(_ context.Context, e tg.Entities, update *tg.UpdateEditMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			return nil
		}

		data := p.parseMessage(e, msg)
		p.dispatcher.Dispatch(active.SessionID, domain.EventEditMessage, data)

		return nil
	})

	dispatcher.OnUserTyping(func(_ context.Context, _ tg.Entities, update *tg.UpdateUserTyping) error {
		data := domain.TypingEventData{
			ChatID: update.UserID,
			UserID: update.UserID,
			Action: "typing",
		}
		p.dispatcher.Dispatch(active.SessionID, domain.EventUserTyping, data)
		return nil
	})

	dispatcher.OnUserStatus(func(_ context.Context, _ tg.Entities, update *tg.UpdateUserStatus) error {
		data := domain.UserStatusEventData{
			UserID: update.UserID,
		}

		switch s := update.Status.(type) {
		case *tg.UserStatusOnline:
			data.Status = "online"
			p.dispatcher.Dispatch(active.SessionID, domain.EventUserOnline, data)
		case *tg.UserStatusOffline:
			data.Status = "offline"
			data.LastSeen = time.Unix(int64(s.WasOnline), 0)
			p.dispatcher.Dispatch(active.SessionID, domain.EventUserOffline, data)
		case *tg.UserStatusRecently:
			data.Status = "recently"
		}

		return nil
	})
}

// parseMessage parses a Telegram message into domain event data.
func (p *SessionPool) parseMessage(e tg.Entities, msg *tg.Message) domain.MessageEventData {
	data := domain.MessageEventData{
		MessageID: int64(msg.ID),
		Text:      msg.Message,
		Date:      time.Unix(int64(msg.Date), 0),
	}

	switch peer := msg.PeerID.(type) {
	case *tg.PeerUser:
		data.ChatID = peer.UserID
		data.ChatType = "private"
		if user, ok := e.Users[peer.UserID]; ok {
			data.FromID = user.ID
			data.FromName = user.FirstName
			if user.LastName != "" {
				data.FromName += " " + user.LastName
			}
		}
	case *tg.PeerChat:
		data.ChatID = peer.ChatID
		data.ChatType = "group"
	case *tg.PeerChannel:
		data.ChatID = peer.ChannelID
		data.ChatType = "channel"
	}

	if msg.Media != nil {
		switch msg.Media.(type) {
		case *tg.MessageMediaPhoto:
			data.MediaType = "photo"
		case *tg.MessageMediaDocument:
			data.MediaType = "document"
		}
	}

	if msg.ReplyTo != nil {
		if reply, ok := msg.ReplyTo.(*tg.MessageReplyHeader); ok {
			data.ReplyToID = int64(reply.ReplyToMsgID)
		}
	}

	return data
}
