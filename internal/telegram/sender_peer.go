package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
)

// resolveByUsername resolves a peer by username.
func resolveByUsername(ctx context.Context, api *tg.Client, to string) (tg.InputPeerClass, error) {
	username := strings.TrimPrefix(to, "@")
	resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	if len(resolved.Users) > 0 {
		user, ok := resolved.Users[0].(*tg.User)
		if ok {
			return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, nil
		}
	}
	if len(resolved.Chats) > 0 {
		switch chat := resolved.Chats[0].(type) {
		case *tg.Channel:
			return &tg.InputPeerChannel{ChannelID: chat.ID, AccessHash: chat.AccessHash}, nil
		case *tg.Chat:
			return &tg.InputPeerChat{ChatID: chat.ID}, nil
		}
	}
	return nil, fmt.Errorf("peer not found: %s", to)
}

// resolveByPhone resolves a peer by phone number.
func resolveByPhone(ctx context.Context, api *tg.Client, to string) (tg.InputPeerClass, error) {
	contacts, err := api.ContactsImportContacts(ctx, []tg.InputPhoneContact{
		{Phone: to, FirstName: "Contact", LastName: ""},
	})
	if err != nil {
		return nil, err
	}
	if len(contacts.Users) > 0 {
		user, ok := contacts.Users[0].(*tg.User)
		if ok {
			return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, nil
		}
	}
	return nil, fmt.Errorf("phone not found: %s", to)
}

// resolveByID resolves a peer by numeric ID.
func resolveByID(ctx context.Context, api *tg.Client, to string) (tg.InputPeerClass, error) {
	userID, err := parseUserID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %s", to)
	}

	inputUsers := []tg.InputUserClass{
		&tg.InputUser{UserID: userID, AccessHash: 0},
	}
	users, err := api.UsersGetUsers(ctx, inputUsers)
	if err == nil && len(users) > 0 {
		if user, ok := users[0].(*tg.User); ok && user.ID != 0 {
			return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, nil
		}
	}

	if userID < 0 {
		chatID := -userID
		if chatID > 1000000000000 {
			channelID := chatID - 1000000000000
			return &tg.InputPeerChannel{ChannelID: channelID, AccessHash: 0}, nil
		}
		return &tg.InputPeerChat{ChatID: chatID}, nil
	}

	return &tg.InputPeerUser{UserID: userID, AccessHash: 0}, nil
}

// resolvePeer resolves a recipient to an InputPeer.
func (m *ClientManager) resolvePeer(ctx context.Context, api *tg.Client, to string) (tg.InputPeerClass, error) {
	if strings.HasPrefix(to, "@") {
		return resolveByUsername(ctx, api, to)
	}

	if strings.HasPrefix(to, "+") {
		return resolveByPhone(ctx, api, to)
	}

	if isNumeric(to) {
		return resolveByID(ctx, api, to)
	}

	return nil, fmt.Errorf("invalid recipient: use @username, +phone, or numeric ID")
}

// isNumeric checks if a string contains only digits (with optional leading minus).
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	start := 0
	if s[0] == '-' {
		start = 1
	}
	if start >= len(s) {
		return false
	}
	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// parseUserID converts a string to int64.
func parseUserID(s string) (int64, error) {
	var result int64
	negative := false
	start := 0

	if len(s) > 0 && s[0] == '-' {
		negative = true
		start = 1
	}

	for i := start; i < len(s); i++ {
		result = result*10 + int64(s[i]-'0')
	}

	if negative {
		result = -result
	}
	return result, nil
}
