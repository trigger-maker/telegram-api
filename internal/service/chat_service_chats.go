package service

import (
	"context"
	"fmt"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// getChatsFromCache retrieves chats from cache if available.
func (s *ChatService) getChatsFromCache(
	ctx context.Context,
	cacheKey, sessionID string,
	refresh bool,
) ([]domain.Chat, bool) {
	if refresh {
		return nil, false
	}

	var cached domain.ChatsResponse
	if err := s.cacheRepo.GetJSON(ctx, cacheKey, &cached); err == nil && len(cached.Chats) > 0 {
		logger.Debug().Str("session_id", sessionID).Int("cached_count", len(cached.Chats)).Msg("chats from cache")
		return cached.Chats, true
	}

	return nil, false
}

// getChatsFromTelegram retrieves chats from Telegram API.
func (s *ChatService) getChatsFromTelegram(
	ctx context.Context,
	sess *domain.TelegramSession,
	cacheKey string,
	archived bool,
) ([]domain.Chat, error) {
	client, err := s.createClient(ctx, sess)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	var result *domain.ChatsResponse
	err = client.Run(ctx, func(ctx context.Context) error {
		var runErr error
		tempReq := domain.GetChatsRequest{Limit: 100, Archived: archived}
		result, runErr = s.tgManager.GetDialogs(ctx, client, tempReq)
		return runErr
	})
	if err != nil {
		return nil, fmt.Errorf("get dialogs: %w", err)
	}

	allChats := result.Chats
	if len(allChats) > 0 {
		cacheData := domain.ChatsResponse{Chats: allChats, TotalCount: len(allChats)}
		if err := s.cacheRepo.SetJSON(ctx, cacheKey, cacheData, s.cacheCfg.ChatsTTL); err != nil {
			logger.Warn().Err(err).Msg("error saving chats to cache")
		}
	}

	return allChats, nil
}

// paginateChats paginates the chat list.
func paginateChats(chats []domain.Chat, offset, limit int) ([]domain.Chat, int, bool) {
	total := len(chats)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	return chats[start:end], total, end < total
}

// GetDialogs retrieves dialogs with cache.
func (s *ChatService) GetDialogs(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	req domain.GetChatsRequest,
) (*domain.ChatsResponse, error) {
	sess, err := s.getValidSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	cacheKey := fmt.Sprintf("tg:chats:%s:archived_%t", sessionID.String(), req.Archived)
	allChats, fromCache := s.getChatsFromCache(ctx, cacheKey, sessionID.String(), req.Refresh)

	if len(allChats) == 0 {
		allChats, err = s.getChatsFromTelegram(ctx, sess, cacheKey, req.Archived)
		if err != nil {
			return nil, err
		}
	}

	chats, total, hasMore := paginateChats(allChats, req.Offset, req.Limit)

	return &domain.ChatsResponse{
		Chats:      chats,
		TotalCount: total,
		HasMore:    hasMore,
		FromCache:  fromCache,
	}, nil
}

// GetChatInfo retrieves chat info with cache.
func (s *ChatService) GetChatInfo(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	chatID int64,
) (*domain.Chat, error) {
	sess, err := s.getValidSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("tg:chat:%s:%d", sessionID.String(), chatID)

	var cached domain.Chat
	if err := s.cacheRepo.GetJSON(ctx, cacheKey, &cached); err == nil && cached.ID != 0 {
		logger.Debug().Int64("chat_id", chatID).Msg("chat info from cache")
		return &cached, nil
	}

	client, err := s.createClient(ctx, sess)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	var result *domain.Chat
	err = client.Run(ctx, func(ctx context.Context) error {
		var runErr error
		result, runErr = s.tgManager.GetChatInfo(ctx, client, chatID)
		return runErr
	})
	if err != nil {
		return nil, fmt.Errorf("get chat info: %w", err)
	}

	if result != nil {
		_ = s.cacheRepo.SetJSON(ctx, cacheKey, result, s.cacheCfg.ChatInfoTTL)
	}

	return result, nil
}

// GetChatHistory retrieves chat history (no cache).
func (s *ChatService) GetChatHistory(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	chatID int64,
	req domain.GetHistoryRequest,
) (*domain.HistoryResponse, error) {
	sess, err := s.getValidSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	client, err := s.createClient(ctx, sess)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	var result *domain.HistoryResponse
	err = client.Run(ctx, func(ctx context.Context) error {
		var runErr error
		result, runErr = s.tgManager.GetChatHistory(ctx, client, chatID, req)
		return runErr
	})
	if err != nil {
		return nil, fmt.Errorf("get chat history: %w", err)
	}

	return result, nil
}
