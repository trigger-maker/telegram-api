package service

import (
	"context"
	"fmt"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
)

// ResolvePeer resolves a username or phone to a peer with cache.
func (s *ChatService) ResolvePeer(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	req domain.ResolveRequest,
) (*domain.ResolvedPeer, error) {
	sess, err := s.getValidSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	identifier := req.Username
	if identifier == "" {
		identifier = req.Phone
	}
	cacheKey := fmt.Sprintf("tg:resolve:%s:%s", sessionID.String(), identifier)

	var cached domain.ResolvedPeer
	if err := s.cacheRepo.GetJSON(ctx, cacheKey, &cached); err == nil && cached.ID != 0 {
		logger.Debug().Str("identifier", identifier).Msg("peer from cache")
		return &cached, nil
	}

	client, err := s.createClient(ctx, sess)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	var result *domain.ResolvedPeer
	err = client.Run(ctx, func(ctx context.Context) error {
		var runErr error
		result, runErr = s.tgManager.ResolveUsername(ctx, client, req)
		return runErr
	})
	if err != nil {
		return nil, fmt.Errorf("resolve peer: %w", err)
	}

	if result != nil {
		_ = s.cacheRepo.SetJSON(ctx, cacheKey, result, s.cacheCfg.ResolveTTL)
	}

	return result, nil
}
