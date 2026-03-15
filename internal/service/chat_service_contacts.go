package service

import (
	"context"
	"fmt"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// GetContacts retrieves contacts with cache and pagination
func (s *ChatService) GetContacts(ctx context.Context, userID, sessionID uuid.UUID, req domain.GetContactsRequest) (*domain.ContactsResponse, error) {
	sess, err := s.getValidSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}

	cacheKey := fmt.Sprintf("tg:contacts:%s", sessionID.String())
	var allContacts []domain.Contact
	fromCache := false

	if !req.Refresh {
		var cached domain.ContactsResponse
		if err := s.cacheRepo.GetJSON(ctx, cacheKey, &cached); err == nil && len(cached.Contacts) > 0 {
			allContacts = cached.Contacts
			fromCache = true
			logger.Debug().Str("session_id", sessionID.String()).Int("cached_count", len(allContacts)).Msg("contacts from cache")
		}
	}

	if len(allContacts) == 0 {
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

		allContacts = result.Contacts

		if len(allContacts) > 0 {
			cacheData := domain.ContactsResponse{Contacts: allContacts, TotalCount: len(allContacts)}
			if err := s.cacheRepo.SetJSON(ctx, cacheKey, cacheData, s.cacheCfg.ContactsTTL); err != nil {
				logger.Warn().Err(err).Msg("error saving contacts to cache")
			}
		}
	}

	total := len(allContacts)
	start := req.Offset
	if start > total {
		start = total
	}
	end := start + req.Limit
	if end > total {
		end = total
	}

	return &domain.ContactsResponse{
		Contacts:   allContacts[start:end],
		TotalCount: total,
		HasMore:    end < total,
		FromCache:  fromCache,
	}, nil
}
