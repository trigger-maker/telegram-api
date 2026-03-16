package service

import (
	"telegram-api/internal/config"
	"telegram-api/internal/domain"
	"telegram-api/internal/telegram"
)

// ChatService manages chat and contact operations.
type ChatService struct {
	sessionRepo domain.SessionRepository
	cacheRepo   domain.CacheRepository
	tgManager   *telegram.ClientManager
	cacheCfg    config.CacheConfig
}

// NewChatService creates a new ChatService instance.
func NewChatService(
	sessionRepo domain.SessionRepository,
	cacheRepo domain.CacheRepository,
	tgManager *telegram.ClientManager,
	cfg *config.Config,
) *ChatService {
	return &ChatService{
		sessionRepo: sessionRepo,
		cacheRepo:   cacheRepo,
		tgManager:   tgManager,
		cacheCfg:    cfg.Cache,
	}
}
