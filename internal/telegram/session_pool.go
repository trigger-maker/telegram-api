package telegram

import (
	"context"
	"sync"
	"time"

	"telegram-api/internal/domain"

	"github.com/google/uuid"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// SessionPoolInterface defines interface for SessionPool
type SessionPoolInterface interface {
	StartSession(ctx context.Context, sess *domain.TelegramSession) error
	StopSession(sessionID uuid.UUID)
	GetActiveSession(sessionID uuid.UUID) (*ActiveSession, bool)
	ListActive() []uuid.UUID
}

// SessionPool manages active Telegram clients
type SessionPool struct {
	sessions    map[uuid.UUID]*ActiveSession
	mu          sync.RWMutex
	manager     *ClientManager
	repo        domain.SessionRepository
	webhookRepo domain.WebhookRepository
	dispatcher  *EventDispatcher
}

// ActiveSession represents an active session listening for events
type ActiveSession struct {
	SessionID    uuid.UUID
	SessionName  string
	TelegramID   int64
	Client       *telegram.Client
	API          *tg.Client
	Cancel       context.CancelFunc
	StartedAt    time.Time
	IsConnected  bool
	LastActivity time.Time
	mu           sync.RWMutex
}

// NewSessionPool creates a new session pool
func NewSessionPool(
	manager *ClientManager,
	repo domain.SessionRepository,
	webhookRepo domain.WebhookRepository,
) *SessionPool {
	pool := &SessionPool{
		sessions:    make(map[uuid.UUID]*ActiveSession),
		manager:     manager,
		repo:        repo,
		webhookRepo: webhookRepo,
	}
	pool.dispatcher = NewEventDispatcher(webhookRepo)
	return pool
}
