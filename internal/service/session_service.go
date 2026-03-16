package service

import (
	"context"

	"telegram-api/internal/config"
	"telegram-api/internal/domain"
	"telegram-api/internal/telegram"

	"github.com/google/uuid"
)

// SessionServiceInterface defines the contract for session service operations.
type SessionServiceInterface interface {
	CreateSession(
		ctx context.Context,
		userID uuid.UUID,
		req *domain.CreateSessionRequest,
	) (*domain.TelegramSession, string, error)
	VerifyCode(ctx context.Context, sessionID uuid.UUID, code string) (*domain.TelegramSession, string, error)
	SubmitPassword(ctx context.Context, sessionID uuid.UUID, password string) (*domain.TelegramSession, error)
	RegenerateQR(ctx context.Context, sessionID uuid.UUID) (string, error)
	ImportTData(
		ctx context.Context,
		userID uuid.UUID,
		apiID int,
		apiHash string,
		sessionName string,
		tdataFiles map[string][]byte,
	) (*domain.TelegramSession, error)
	ListSessions(ctx context.Context, userID uuid.UUID) ([]domain.TelegramSession, error)
	GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.TelegramSession, error)
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
}

// SessionService manages Telegram session operations.
type SessionService struct {
	sessionRepo domain.SessionRepository
	userRepo    domain.UserRepository
	tgManager   *telegram.ClientManager
	cache       domain.CacheRepository
	config      *config.Config
}

// NewSessionService creates a new SessionService instance.
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
