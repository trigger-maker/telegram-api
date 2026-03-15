package service

import (
	"context"
	"fmt"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// GetDialogs retrieves dialogs with cache
func (s *ChatService) GetDialogs(ctx context.Context, userID, sessionID uuid.UUID, req domain.GetChatsRequest) (*domain.ChatsResponse, error) {
	sess, err := s.getValidSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	cacheKey := fmt.Sprintf("tg:chats:%s:archived_%t", sessionID.String(), req.Archived)
	var allChats []domain.Chat
	fromCache := false

	if !req.Refresh {
		var cached domain.ChatsResponse
		if err := s.cacheRepo.GetJSON(ctx, cacheKey, &cached); err == nil && len(cached.Chats) > 0 {
			allChats = cached.Chats
			fromCache = true
			logger.Debug().Str("session_id", sessionID.String()).Int("cached_count", len(allChats)).Msg("chats from cache")
		}
	}

	if len(allChats) == 0 {
		client, err := s.createClient(ctx, sess)
		if err != nil {
			return nil, fmt.Errorf("create client: %w", err)
		}

		var result *domain.ChatsResponse
		err = client.Run(ctx, func(ctx context.Context) error {
			var runErr error
			tempReq := domain.GetChatsRequest{Limit: 100, Archived: req.Archived}
			result, runErr = s.tgManager.GetDialogs(ctx, client, tempReq)
			return runErr
		})
		if err != nil {
			return nil, fmt.Errorf("get dialogs: %w", err)
		}

		allChats = result.Chats

		if len(allChats) > 0 {
			cacheData := domain.ChatsResponse{Chats: allChats, TotalCount: len(allChats)}
			if err := s.cacheRepo.SetJSON(ctx, cacheKey, cacheData, s.cacheCfg.ChatsTTL); err != nil {
				logger.Warn().Err(err).Msg("error saving chats to cache")
			}
		}
	}

	total := len(allChats)
	start := req.Offset
	if start > total {
		start = total
	}
	end := start + req.Limit
	if end > total {
		end = total
	}

	return &domain.ChatsResponse{
		Chats:      allChats[start:end],
		TotalCount: total,
		HasMore:    end < total,
		FromCache:  fromCache,
	}, nil
}

// GetChatInfo retrieves chat info with cache
func (s *ChatService) GetChatInfo(ctx context.Context, userID, sessionID uuid.UUID, chatID int64) (*domain.Chat, error) {
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

// GetChatHistory retrieves chat history (no cache)
func (s *ChatService) GetChatHistory(ctx context.Context, userID, sessionID uuid.UUID, chatID int64, req domain.GetHistoryRequest) (*domain.HistoryResponse, error) {
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
