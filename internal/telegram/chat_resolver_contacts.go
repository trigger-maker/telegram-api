package telegram

import (
	"context"
	"fmt"
	"strings"

	"telegram-api/internal/domain"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// GetContacts retrieves the user's contacts list
func (m *ClientManager) GetContacts(ctx context.Context, client *telegram.Client) (*domain.ContactsResponse, error) {
	api := client.API()

	result, err := api.ContactsGetContacts(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("get contacts: %w", err)
	}

	contacts, ok := result.(*tg.ContactsContacts)
	if !ok {
		return &domain.ContactsResponse{Contacts: []domain.Contact{}, TotalCount: 0}, nil
	}

	users := buildUserMap(contacts.Users)
	var contactList []domain.Contact

	for _, c := range contacts.Contacts {
		if user, ok := users[c.UserID]; ok {
			contact := domain.Contact{
				ID:         user.ID,
				Phone:      user.Phone,
				FirstName:  user.FirstName,
				LastName:   user.LastName,
				Username:   user.Username,
				IsMutual:   c.Mutual,
				AccessHash: user.AccessHash,
			}

			if user.Status != nil {
				contact.Status, contact.LastSeenAt = parseUserStatus(user.Status)
			}

			contactList = append(contactList, contact)
		}
	}

	return &domain.ContactsResponse{
		Contacts:   contactList,
		TotalCount: len(contactList),
	}, nil
}

// ResolveUsername resolves a username or phone number to a peer
func (m *ClientManager) ResolveUsername(ctx context.Context, client *telegram.Client, req domain.ResolveRequest) (*domain.ResolvedPeer, error) {
	api := client.API()

	if req.Username != "" {
		username := strings.TrimPrefix(req.Username, "@")
		result, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
			Username: username,
		})
		if err != nil {
			return nil, fmt.Errorf("resolve username: %w", err)
		}

		users := buildUserMap(result.Users)
		_, channels := buildChatMaps(result.Chats)

		switch p := result.Peer.(type) {
		case *tg.PeerUser:
			if user, ok := users[p.UserID]; ok {
				return &domain.ResolvedPeer{
					ID:         user.ID,
					Type:       domain.ChatTypePrivate,
					Username:   user.Username,
					FirstName:  user.FirstName,
					LastName:   user.LastName,
					Phone:      user.Phone,
					AccessHash: user.AccessHash,
					IsBot:      user.Bot,
					IsVerified: user.Verified,
				}, nil
			}
		case *tg.PeerChannel:
			if ch, ok := channels[p.ChannelID]; ok {
				chatType := domain.ChatTypeSupergroup
				if ch.Broadcast {
					chatType = domain.ChatTypeChannel
				}
				return &domain.ResolvedPeer{
					ID:         ch.ID,
					Type:       chatType,
					Username:   ch.Username,
					Title:      ch.Title,
					AccessHash: ch.AccessHash,
					IsVerified: ch.Verified,
				}, nil
			}
		}

		return nil, fmt.Errorf("peer not found for username: %s", username)
	}

	if req.Phone != "" {
		phone := strings.TrimPrefix(req.Phone, "+")
		result, err := api.ContactsResolvePhone(ctx, phone)
		if err != nil {
			return nil, fmt.Errorf("resolve phone: %w", err)
		}

		users := buildUserMap(result.Users)
		if p, ok := result.Peer.(*tg.PeerUser); ok {
			if user, ok := users[p.UserID]; ok {
				return &domain.ResolvedPeer{
					ID:         user.ID,
					Type:       domain.ChatTypePrivate,
					Username:   user.Username,
					FirstName:  user.FirstName,
					LastName:   user.LastName,
					Phone:      user.Phone,
					AccessHash: user.AccessHash,
					IsBot:      user.Bot,
				}, nil
			}
		}

		return nil, fmt.Errorf("peer not found for phone: %s", phone)
	}

	return nil, fmt.Errorf("username or phone required")
}
