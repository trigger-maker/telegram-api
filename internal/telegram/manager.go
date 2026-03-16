package telegram

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"telegram-api/internal/config"
	"telegram-api/internal/domain"
	"telegram-api/pkg/crypto"

	"github.com/gotd/td/telegram"
)

// ClientManager manages Telegram client operations.
type ClientManager struct {
	cfg        *config.Config
	repo       domain.SessionRepository
	crypter    *crypto.Crypter
	pool       *SessionPool
	httpClient *http.Client
	mu         sync.RWMutex
}

// TGUser represents a Telegram user.
type TGUser struct {
	ID       int64
	Username string
}

// SignInResult represents the result of sign in operation.
type SignInResult struct {
	User          *TGUser
	SessionData   []byte
	NeedsPassword bool
	PasswordHint  string
}

// QRAuthResult represents the result of QR authentication.
type QRAuthResult struct {
	User        *TGUser
	SessionData []byte
	Error       error
}

// NewManager creates a new ClientManager instance.
func NewManager(cfg *config.Config, repo domain.SessionRepository) (*ClientManager, error) {
	crypter, err := crypto.NewCrypter(cfg.Encryption.Key)
	if err != nil {
		return nil, fmt.Errorf("crypto init: %w", err)
	}

	return &ClientManager{
		cfg:     cfg,
		repo:    repo,
		crypter: crypter,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// SetPool sets the session pool for the manager.
func (m *ClientManager) SetPool(pool *SessionPool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pool = pool
}

// newClient creates a new Telegram client with the given configuration.
func (m *ClientManager) newClient(
	apiID int,
	apiHash, sessionName string,
	storage telegram.SessionStorage,
) *telegram.Client {
	return telegram.NewClient(apiID, apiHash, telegram.Options{
		SessionStorage: storage,
		Device: telegram.DeviceConfig{
			DeviceModel:    sessionName,
			SystemVersion:  "1.0",
			AppVersion:     "1.0.0",
			SystemLangCode: "es",
			LangCode:       "es",
		},
	})
}

// Encrypt encrypts data using the configured crypter.
func (m *ClientManager) Encrypt(data []byte) ([]byte, error) {
	return m.crypter.Encrypt(data)
}

// Decrypt decrypts data using the configured crypter.
func (m *ClientManager) Decrypt(data []byte) ([]byte, error) {
	return m.crypter.Decrypt(data)
}
