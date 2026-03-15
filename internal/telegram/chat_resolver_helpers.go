package telegram

import (
	"context"
	"time"

	"telegram-api/internal/domain"

	"github.com/gotd/td/tg"
)

// parseDialog parses a Telegram dialog into domain Chat
func (m *ClientManager) parseDialog(dialog *tg.Dialog, users map[int64]*tg.User, chats map[int64]*tg.Chat, channels map[int64]*tg.Channel, messages map[int]*tg.Message) *domain.Chat {
	chat := &domain.Chat{
		UnreadCount: dialog.UnreadCount,
		IsPinned:    dialog.Pinned,
		IsArchived:  dialog.FolderID == 1,
	}

	if msg, ok := messages[dialog.TopMessage]; ok {
		chat.LastMessageID = msg.ID
		chat.LastMessage = truncateString(msg.Message, 100)
		chat.LastMessageAt = time.Unix(int64(msg.Date), 0)
	}

	switch p := dialog.Peer.(type) {
	case *tg.PeerUser:
		if user, ok := users[p.UserID]; ok {
			chat.ID = user.ID
			chat.Type = domain.ChatTypePrivate
			chat.FirstName = user.FirstName
			chat.LastName = user.LastName
			chat.Username = user.Username
		}
	case *tg.PeerChat:
		if c, ok := chats[p.ChatID]; ok {
			chat.ID = c.ID
			chat.Type = domain.ChatTypeGroup
			chat.Title = c.Title
		}
	case *tg.PeerChannel:
		if ch, ok := channels[p.ChannelID]; ok {
			chat.ID = ch.ID
			if ch.Broadcast {
				chat.Type = domain.ChatTypeChannel
			} else {
				chat.Type = domain.ChatTypeSupergroup
			}
			chat.Title = ch.Title
			chat.Username = ch.Username
		}
	default:
		return nil
	}

	return chat
}

// resolvePeerByID resolves a peer by ID
func (m *ClientManager) resolvePeerByID(ctx context.Context, api *tg.Client, chatID int64) (tg.InputPeerClass, error) {
	if chatID > 0 {
		return &tg.InputPeerUser{UserID: chatID}, nil
	}

	channelID := -chatID
	if channelID > 1000000000000 {
		channelID = channelID - 1000000000000
	}

	return &tg.InputPeerChannel{ChannelID: channelID}, nil
}

// buildUserMap builds a map of user ID to User
func buildUserMap(users []tg.UserClass) map[int64]*tg.User {
	m := make(map[int64]*tg.User)
	for _, u := range users {
		if user, ok := u.(*tg.User); ok {
			m[user.ID] = user
		}
	}
	return m
}

// buildChatMaps builds maps of chat ID to Chat and channel ID to Channel
func buildChatMaps(chats []tg.ChatClass) (map[int64]*tg.Chat, map[int64]*tg.Channel) {
	chatMap := make(map[int64]*tg.Chat)
	channelMap := make(map[int64]*tg.Channel)
	for _, c := range chats {
		switch ch := c.(type) {
		case *tg.Chat:
			chatMap[ch.ID] = ch
		case *tg.Channel:
			channelMap[ch.ID] = ch
		}
	}
	return chatMap, channelMap
}

// buildMessageMap builds a map of message ID to Message
func buildMessageMap(messages []tg.MessageClass) map[int]*tg.Message {
	m := make(map[int]*tg.Message)
	for _, msg := range messages {
		if message, ok := msg.(*tg.Message); ok {
			m[message.ID] = message
		}
	}
	return m
}

// parseMessage parses a Telegram message into domain ChatMessage
func parseMessage(msg *tg.Message, users map[int64]*tg.User, chatID int64) domain.ChatMessage {
	cm := domain.ChatMessage{
		ID:         msg.ID,
		ChatID:     chatID,
		Text:       msg.Message,
		Date:       time.Unix(int64(msg.Date), 0),
		IsOutgoing: msg.Out,
	}

	if msg.FromID != nil {
		if from, ok := msg.FromID.(*tg.PeerUser); ok {
			cm.FromID = from.UserID
			if user, ok := users[from.UserID]; ok {
				cm.FromName = user.FirstName
				if user.LastName != "" {
					cm.FromName += " " + user.LastName
				}
			}
		}
	}

	if reply, ok := msg.GetReplyTo(); ok {
		if header, ok := reply.(*tg.MessageReplyHeader); ok {
			cm.ReplyToID = header.ReplyToMsgID
		}
	}

	if msg.Media != nil {
		switch msg.Media.(type) {
		case *tg.MessageMediaPhoto:
			cm.MediaType = "photo"
		case *tg.MessageMediaDocument:
			cm.MediaType = "document"
		case *tg.MessageMediaGeo:
			cm.MediaType = "location"
		case *tg.MessageMediaContact:
			cm.MediaType = "contact"
		}
	}

	if fwd, ok := msg.GetFwdFrom(); ok && fwd.FromID != nil {
		cm.ForwardFrom = "forwarded"
	}

	return cm
}

// parseUserStatus parses a user status into status string and last seen time
func parseUserStatus(status tg.UserStatusClass) (string, *time.Time) {
	switch s := status.(type) {
	case *tg.UserStatusOnline:
		return "online", nil
	case *tg.UserStatusOffline:
		t := time.Unix(int64(s.WasOnline), 0)
		return "offline", &t
	case *tg.UserStatusRecently:
		return "recently", nil
	case *tg.UserStatusLastWeek:
		return "last_week", nil
	case *tg.UserStatusLastMonth:
		return "last_month", nil
	default:
		return "unknown", nil
	}
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
