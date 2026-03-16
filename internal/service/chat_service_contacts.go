package service

import (
	"context"
	"fmt"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// getContactsFromCache retrieves contacts from cache if available.
func (s *ChatService) getContactsFromCache(
	ctx context.Context,
	cacheKey, sessionID string,
	refresh bool,
) ([]domain.Contact, bool) {
	if refresh {
		return nil, false
	}

	var cached domain.ContactsResponse
	if err := s.cacheRepo.GetJSON(ctx, cacheKey, &cached); err == nil && len(cached.Contacts) > 0 {
		logger.Debug().Str("session_id", sessionID).Int("cached_count", len(cached.Contacts)).Msg("contacts from cache")
		return cached.Contacts, true
	}

	return nil, false
}

// getContactsFromTelegram retrieves contacts from Telegram API.
func (s *ChatService) getContactsFromTelegram(
	ctx context.Context,
	sess *domain.TelegramSession,
	cacheKey string,
) ([]domain.Contact, error) {
	client, err := s.createClient(ctx, sess)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	var result *domain.ContactsResponse
	err = client.Run(ctx, func(ctx context.Context) error {
		var runErr error
		result, runErr = s.tgManager.GetContacts(ctx, client)
		return runErr
	})
	if err != nil {
		return nil, fmt.Errorf("get contacts: %w", err)
	}

	allContacts := result.Contacts
	if len(allContacts) > 0 {
		cacheData := domain.ContactsResponse{Contacts: allContacts, TotalCount: len(allContacts)}
		if err := s.cacheRepo.SetJSON(ctx, cacheKey, cacheData, s.cacheCfg.ContactsTTL); err != nil {
			logger.Warn().Err(err).Msg("error saving contacts to cache")
		}
	}

	return allContacts, nil
}

// paginateContacts paginates the contact list.
func paginateContacts(contacts []domain.Contact, offset, limit int) ([]domain.Contact, int, bool) {
	total := len(contacts)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	return contacts[start:end], total, end < total
}

// GetContacts retrieves contacts with cache and pagination.
func (s *ChatService) GetContacts(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	req domain.GetContactsRequest,
) (*domain.ContactsResponse, error) {
	sess, err := s.getValidSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}

	cacheKey := fmt.Sprintf("tg:contacts:%s", sessionID.String())
	allContacts, fromCache := s.getContactsFromCache(ctx, cacheKey, sessionID.String(), req.Refresh)

	if len(allContacts) == 0 {
		allContacts, err = s.getContactsFromTelegram(ctx, sess, cacheKey)
		if err != nil {
			return nil, err
		}
	}

	contacts, total, hasMore := paginateContacts(allContacts, req.Offset, req.Limit)

	return &domain.ContactsResponse{
		Contacts:   contacts,
		TotalCount: total,
		HasMore:    hasMore,
		FromCache:  fromCache,
	}, nil
}
