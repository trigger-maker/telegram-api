package domain

import (
	"errors"
	"fmt"
)

var (
	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user deactivated")

	// Authentication errors
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
	ErrTokenRevoked = errors.New("token revoked")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("access denied")

	// Telegram session errors
	ErrSessionNotFound         = errors.New("session not found")
	ErrSessionAlreadyExists    = errors.New("session already exists for this number")
	ErrSessionNotActive        = errors.New("session not active")
	ErrSessionNotAuthenticated = errors.New("session not authenticated")
	ErrSessionInactive         = errors.New("session inactive")
	ErrInvalidPhoneNumber      = errors.New("invalid phone number")
	ErrInvalidCode             = errors.New("invalid verification code")
	ErrCodeExpired             = errors.New("verification code expired")
	ErrPasswordRequired        = errors.New("2FA password required")
	ErrInvalidPassword         = errors.New("invalid 2FA password")
	ErrAlreadyAuthenticated     = errors.New("session already authenticated")
	ErrTelegramError           = errors.New("Telegram error")
	ErrTelegramFloodWait       = errors.New("too many attempts, please wait")
	ErrTDataInvalid            = errors.New("invalid tdata files")

	// Message errors
	ErrMessageNotFound   = errors.New("message not found")
	ErrChatNotFound      = errors.New("chat not found")
	ErrPeerNotFound      = errors.New("recipient not found")
	ErrMediaNotSupported = errors.New("media type not supported")

	// Validation errors
	ErrValidation   = errors.New("validation error")
	ErrInvalidInput = errors.New("invalid input")

	// System errors
	ErrInternal          = errors.New("internal server error")
	ErrDatabase          = errors.New("database error")
	ErrCache             = errors.New("cache error")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

type AppError struct {
	Err     error
	Message string
	Code    string
	Status  int
	Details map[string]interface{}
}

type QRExpiredError struct {
	NewQR       string
	Attempt     int
	MaxAttempts int
	SessionID   string
	SessionName string
}

func (e *QRExpiredError) Error() string {
	return fmt.Sprintf("QR expired. Attempt %d/%d", e.Attempt, e.MaxAttempts)
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(err error, message string, status int) *AppError {
	return &AppError{Err: err, Message: message, Status: status}
}

func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}
