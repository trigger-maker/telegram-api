package service

import (
	"telegram-api/internal/config"
	"telegram-api/internal/domain"
	"telegram-api/internal/telegram"
)

// SessionService manages Telegram session operations
type SessionService struct {
	sessionRepo domain.SessionRepository
	userRepo    domain.UserRepository
	tgManager   *telegram.ClientManager
	cache       domain.CacheRepository
	config      *config.Config
}

// NewSessionService creates a new SessionService instance
func NewSessionService(
	sRepo domain.SessionRepository,
	uRepo domain.UserRepository,
	tgMgr *telegram.ClientManager,
	cache domain.CacheRepository,
	cfg *config.Config,
) *SessionService {
	return &SessionService{
		sessionRepo: sRepo,
		userRepo:    uRepo,
		tgManager:   tgMgr,
		cache:       cache,
		config:      cfg,
	}
}
