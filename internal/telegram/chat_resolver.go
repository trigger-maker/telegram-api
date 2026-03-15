package telegram

import (
	"context"
	"fmt"

	"telegram-api/internal/domain"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// GetDialogs retrieves the list of chats/dialogs
func (m *ClientManager) GetDialogs(ctx context.Context, client *telegram.Client, req domain.GetChatsRequest) (*domain.ChatsResponse, error) {
	api := client.API()

	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	result, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      req.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get dialogs: %w", err)
	}

	var chats []domain.Chat
	var dialogs []tg.DialogClass
	var users map[int64]*tg.User
	var chatsMap map[int64]*tg.Chat
	var channelsMap map[int64]*tg.Channel
	var messagesMap map[int]*tg.Message

	switch d := result.(type) {
	case *tg.MessagesDialogs:
		dialogs = d.Dialogs
		users = buildUserMap(d.Users)
		chatsMap, channelsMap = buildChatMaps(d.Chats)
		messagesMap = buildMessageMap(d.Messages)
	case *tg.MessagesDialogsSlice:
		dialogs = d.Dialogs
		users = buildUserMap(d.Users)
		chatsMap, channelsMap = buildChatMaps(d.Chats)
		messagesMap = buildMessageMap(d.Messages)
	default:
		return nil, fmt.Errorf("unexpected dialogs type: %T", result)
	}

	for _, dlg := range dialogs {
		dialog, ok := dlg.(*tg.Dialog)
		if !ok {
			continue
		}

		chat := m.parseDialog(dialog, users, chatsMap, channelsMap, messagesMap)
		if chat != nil {
			if !req.Archived && chat.IsArchived {
				continue
			}
			chats = append(chats, *chat)
		}
	}

	return &domain.ChatsResponse{
		Chats:      chats,
		TotalCount: len(chats),
		HasMore:    len(dialogs) == req.Limit,
	}, nil
}

// GetChatInfo retrieves information about a specific chat
func (m *ClientManager) GetChatInfo(ctx context.Context, client *telegram.Client, chatID int64) (*domain.Chat, error) {
	api := client.API()

	if chatID > 0 {
		users, err := api.UsersGetUsers(ctx, []tg.InputUserClass{
			&tg.InputUser{UserID: chatID},
		})
		if err == nil && len(users) > 0 {
			if user, ok := users[0].(*tg.User); ok {
				return &domain.Chat{
					ID:        user.ID,
					Type:      domain.ChatTypePrivate,
					FirstName: user.FirstName,
					LastName:  user.LastName,
					Username:  user.Username,
				}, nil
			}
		}
	}

	if chatID < 0 {
		channelID := -chatID
		if channelID > 1000000000000 {
			channelID = channelID - 1000000000000
		}

		result, err := api.ChannelsGetChannels(ctx, []tg.InputChannelClass{
			&tg.InputChannel{ChannelID: channelID},
		})
		if err == nil {
			if chats, ok := result.(*tg.MessagesChats); ok && len(chats.Chats) > 0 {
				if ch, ok := chats.Chats[0].(*tg.Channel); ok {
					chatType := domain.ChatTypeSupergroup
					if ch.Broadcast {
						chatType = domain.ChatTypeChannel
					}
					return &domain.Chat{
						ID:       ch.ID,
						Type:     chatType,
						Title:    ch.Title,
						Username: ch.Username,
					}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("chat not found: %d", chatID)
}

// GetChatHistory retrieves the message history for a chat
func (m *ClientManager) GetChatHistory(ctx context.Context, client *telegram.Client, chatID int64, req domain.GetHistoryRequest) (*domain.HistoryResponse, error) {
	api := client.API()

	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	peer, err := m.resolvePeerByID(ctx, api, chatID)
	if err != nil {
		return nil, fmt.Errorf("resolve peer: %w", err)
	}

	result, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:       peer,
		Limit:      req.Limit,
		OffsetID:   req.OffsetID,
		OffsetDate: req.OffsetDate,
	})
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}

	var messages []domain.ChatMessage
	var msgList []tg.MessageClass
	var users map[int64]*tg.User

	switch h := result.(type) {
	case *tg.MessagesMessages:
		msgList = h.Messages
		users = buildUserMap(h.Users)
	case *tg.MessagesMessagesSlice:
		msgList = h.Messages
		users = buildUserMap(h.Users)
	case *tg.MessagesChannelMessages:
		msgList = h.Messages
		users = buildUserMap(h.Users)
	default:
		return nil, fmt.Errorf("unexpected history type: %T", result)
	}

	for _, m := range msgList {
		if msg, ok := m.(*tg.Message); ok {
			messages = append(messages, parseMessage(msg, users, chatID))
		}
	}

	return &domain.HistoryResponse{
		Messages:   messages,
		TotalCount: len(messages),
		HasMore:    len(messages) == req.Limit,
	}, nil
}
