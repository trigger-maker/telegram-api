package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SessionStatus represents the status of a Telegram session.
type SessionStatus string

const (
	// SessionPending represents a pending session.
	SessionPending SessionStatus = "pending"
	// SessionCodeSent represents a session where code was sent.
	SessionCodeSent SessionStatus = "code_sent"
	// SessionPasswordRequired represents a session requiring 2FA password.
	SessionPasswordRequired SessionStatus = "password_required"
	// SessionAuthenticated represents an authenticated session.
	SessionAuthenticated SessionStatus = "authenticated"
	// SessionFailed represents a failed session.
	SessionFailed SessionStatus = "failed"
	// SessionBanned represents a banned session.
	SessionBanned SessionStatus = "banned"
	// SessionFrozen represents a frozen session.
	SessionFrozen SessionStatus = "frozen"
)

// AuthMethod represents the authentication method.
type AuthMethod string

const (
	// AuthMethodSMS represents SMS authentication.
	AuthMethodSMS AuthMethod = "sms"
	// AuthMethodQR represents QR code authentication.
	AuthMethodQR AuthMethod = "qr"
	// AuthMethodTData represents TData import authentication.
	AuthMethodTData AuthMethod = "tdata"
)

// TelegramSession represents a Telegram session.
type TelegramSession struct {
	ID               uuid.UUID     `json:"id"`
	UserID           uuid.UUID     `json:"user_id"`
	PhoneNumber      string        `json:"phone_number"`
	APIID            int           `json:"api_id"`
	APIHash          string        `json:"-"`
	APIHashEncrypted []byte        `json:"-"`
	SessionName      string        `json:"session_name"`
	SessionData      []byte        `json:"-"`
	AuthState        SessionStatus `json:"auth_state"`
	AuthMethod       AuthMethod    `json:"auth_method,omitempty"`
	TelegramUserID   int64         `json:"telegram_user_id,omitempty"`
	TelegramUsername string        `json:"telegram_username,omitempty"`
	IsActive         bool          `json:"is_active"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// CreateSessionRequest represents a request to create a session.
type CreateSessionRequest struct {
	Phone       string     `json:"phone,omitempty"`
	APIID       int        `json:"api_id" validate:"required,gt=0"`
	APIHash     string     `json:"api_hash" validate:"required,len=32"`
	SessionName string     `json:"session_name,omitempty"`
	AuthMethod  AuthMethod `json:"auth_method,omitempty"`
}

// VerifyCodeRequest represents a request to verify a verification code.
type VerifyCodeRequest struct {
	Code string `json:"code" validate:"required,min=5,max=6"`
}

// Verify2FARequest represents a request to verify 2FA password.
type Verify2FARequest struct {
	Password string `json:"password" validate:"required"`
}

// QRCodeResponse represents a QR code response.
type QRCodeResponse struct {
	Token      string `json:"token"`
	URL        string `json:"url"`
	QRImageB64 string `json:"qr_image_base64"`
	ExpiresIn  int    `json:"expires_in"`
}

// SessionRepository defines operations for session persistence.
type SessionRepository interface {
	Create(ctx context.Context, session *TelegramSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*TelegramSession, error)
	GetByPhone(ctx context.Context, phone string) (*TelegramSession, error)
	GetByUserAndPhone(ctx context.Context, userID uuid.UUID, phone string) (*TelegramSession, error)
	Update(ctx context.Context, session *TelegramSession) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]TelegramSession, error)
	ListAllActive(ctx context.Context) ([]TelegramSession, error)
	UpdateSessionData(sessionID string, data []byte) error
}
