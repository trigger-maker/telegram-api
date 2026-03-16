package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrUserNotFound is returned when a user is not found in the database.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when attempting to create a user that already exists.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrEmailAlreadyExists is returned when attempting to register with an existing email.
	ErrEmailAlreadyExists = errors.New("email already registered")
	// ErrInvalidCredentials is returned when authentication credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserInactive is returned when a user account is deactivated.
	ErrUserInactive = errors.New("user deactivated")

	// ErrInvalidToken is returned when an authentication token is invalid.
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is returned when an authentication token has expired.
	ErrTokenExpired = errors.New("token expired")
	// ErrTokenRevoked is returned when an authentication token has been revoked.
	ErrTokenRevoked = errors.New("token revoked")
	// ErrUnauthorized is returned when authentication is required but not provided.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden is returned when access is denied due to insufficient permissions.
	ErrForbidden = errors.New("access denied")

	// ErrSessionNotFound is returned when a session is not found.
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionAlreadyExists is returned when a session already exists for this phone number.
	ErrSessionAlreadyExists = errors.New("session already exists for this number")
	// ErrSessionNotActive is returned when a session is not active.
	ErrSessionNotActive = errors.New("session not active")
	// ErrSessionNotAuthenticated is returned when a session is not authenticated.
	ErrSessionNotAuthenticated = errors.New("session not authenticated")
	// ErrSessionInactive is returned when a session is inactive.
	ErrSessionInactive = errors.New("session inactive")
	// ErrInvalidPhoneNumber is returned when a phone number is invalid.
	ErrInvalidPhoneNumber = errors.New("invalid phone number")
	// ErrInvalidCode is returned when a verification code is invalid.
	ErrInvalidCode = errors.New("invalid verification code")
	// ErrCodeExpired is returned when a verification code has expired.
	ErrCodeExpired = errors.New("verification code expired")
	// ErrPasswordRequired is returned when 2FA password is required.
	ErrPasswordRequired = errors.New("2FA password required")
	// ErrInvalidPassword is returned when 2FA password is invalid.
	ErrInvalidPassword = errors.New("invalid 2FA password")
	// ErrAlreadyAuthenticated is returned when a session is already authenticated.
	ErrAlreadyAuthenticated = errors.New("session already authenticated")
	// ErrTelegramError is returned when a Telegram error occurs.
	ErrTelegramError = errors.New("telegram error")
	// ErrTelegramFloodWait is returned when too many requests have been made.
	ErrTelegramFloodWait = errors.New("too many attempts, please wait")
	// ErrTDataInvalid is returned when tdata files are invalid.
	ErrTDataInvalid = errors.New("invalid tdata files")

	// ErrMessageNotFound is returned when a message is not found.
	ErrMessageNotFound = errors.New("message not found")
	// ErrChatNotFound is returned when a chat is not found.
	ErrChatNotFound = errors.New("chat not found")
	// ErrPeerNotFound is returned when a recipient is not found.
	ErrPeerNotFound = errors.New("recipient not found")
	// ErrMediaNotSupported is returned when a media type is not supported.
	ErrMediaNotSupported = errors.New("media type not supported")

	// ErrValidation is returned when input validation fails.
	ErrValidation = errors.New("validation error")
	// ErrInvalidInput is returned when input is invalid.
	ErrInvalidInput = errors.New("invalid input")

	// ErrInternal is returned when an internal server error occurs.
	ErrInternal = errors.New("internal server error")
	// ErrDatabase is returned when a database error occurs.
	ErrDatabase = errors.New("database error")
	// ErrCache is returned when a cache error occurs.
	ErrCache = errors.New("cache error")
	// ErrRateLimitExceeded is returned when rate limit is exceeded.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// AppError represents an application error with additional metadata.
type AppError struct {
	Err     error
	Message string
	Code    string
	Status  int
	Details map[string]interface{}
}

// QRExpiredError represents an error when a QR code expires.
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

// NewAppError creates a new AppError instance.
func NewAppError(err error, message string, status int) *AppError {
	return &AppError{Err: err, Message: message, Status: status}
}

// WithCode sets the error code.
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// WithDetails sets the error details.
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}
