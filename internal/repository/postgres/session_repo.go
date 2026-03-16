package postgres

import (
	"context"
	"errors"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionRepository handles database operations for Telegram sessions.
type SessionRepository struct {
	db *pgxpool.Pool
}

// NewSessionRepository creates a new SessionRepository instance.
func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create inserts a new Telegram session into the database.
func (r *SessionRepository) Create(ctx context.Context, s *domain.TelegramSession) error {
	query := `
		INSERT INTO telegram_sessions (
			id, user_id, phone_number, api_id, api_hash_encrypted,
			session_name, session_data, auth_state, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		s.ID, s.UserID, s.PhoneNumber, s.APIID, s.APIHashEncrypted,
		s.SessionName, s.SessionData, s.AuthState, s.IsActive, s.CreatedAt, s.UpdatedAt,
	)
	return wrapDBError(err, "create session")
}

// Update updates an existing Telegram session in the database.
func (r *SessionRepository) Update(ctx context.Context, s *domain.TelegramSession) error {
	query := `
		UPDATE telegram_sessions SET 
			phone_number = $1, session_data = $2, auth_state = $3, telegram_user_id = $4,
			telegram_username = $5, is_active = $6, updated_at = NOW()
		WHERE id = $7
	`
	_, err := r.db.Exec(ctx, query,
		s.PhoneNumber, s.SessionData, s.AuthState, s.TelegramUserID,
		s.TelegramUsername, s.IsActive, s.ID,
	)
	return wrapDBError(err, "update session")
}

// GetByID retrieves a Telegram session by its ID.
func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TelegramSession, error) {
	query := `
		SELECT id, user_id, phone_number, api_id, api_hash_encrypted, session_name, 
			session_data, auth_state, COALESCE(telegram_user_id, 0), COALESCE(telegram_username, ''), 
			is_active, created_at, updated_at
		FROM telegram_sessions WHERE id = $1
	`
	var s domain.TelegramSession
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.PhoneNumber, &s.APIID, &s.APIHashEncrypted, &s.SessionName,
		&s.SessionData, &s.AuthState, &s.TelegramUserID, &s.TelegramUsername, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		logger.Error().Err(err).Str("id", id.String()).Msg("Error GetByID")
	}
	return &s, wrapDBError(err, "get session")
}

// GetByPhone retrieves a Telegram session by phone number.
func (r *SessionRepository) GetByPhone(ctx context.Context, phone string) (*domain.TelegramSession, error) {
	query := `
		SELECT id, user_id, phone_number, api_id, api_hash_encrypted, session_name,
			session_data, auth_state, COALESCE(telegram_user_id, 0), COALESCE(telegram_username, ''),
			is_active, created_at, updated_at
		FROM telegram_sessions WHERE phone_number = $1
	`
	var s domain.TelegramSession
	err := r.db.QueryRow(ctx, query, phone).Scan(
		&s.ID, &s.UserID, &s.PhoneNumber, &s.APIID, &s.APIHashEncrypted, &s.SessionName,
		&s.SessionData, &s.AuthState, &s.TelegramUserID, &s.TelegramUsername, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	return &s, wrapDBError(err, "get session by phone")
}

// GetByUserAndPhone retrieves a Telegram session by user ID and phone number.
func (r *SessionRepository) GetByUserAndPhone(
	ctx context.Context,
	userID uuid.UUID,
	phone string,
) (*domain.TelegramSession, error) {
	query := `
		SELECT id, user_id, phone_number, api_id, api_hash_encrypted, session_name,
			session_data, auth_state, COALESCE(telegram_user_id, 0), COALESCE(telegram_username, ''),
			is_active, created_at, updated_at
		FROM telegram_sessions WHERE user_id = $1 AND phone_number = $2
	`
	var s domain.TelegramSession
	err := r.db.QueryRow(ctx, query, userID, phone).Scan(
		&s.ID, &s.UserID, &s.PhoneNumber, &s.APIID, &s.APIHashEncrypted, &s.SessionName,
		&s.SessionData, &s.AuthState, &s.TelegramUserID, &s.TelegramUsername, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	return &s, wrapDBError(err, "get session by user and phone")
}

// ListByUserID retrieves all Telegram sessions for a given user ID.
func (r *SessionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.TelegramSession, error) {
	query := `
		SELECT id, user_id, phone_number, api_id, api_hash_encrypted, session_name,
			session_data, auth_state, COALESCE(telegram_user_id, 0), COALESCE(telegram_username, ''),
			is_active, created_at, updated_at
		FROM telegram_sessions WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		logger.Error().Err(err).Str("user_id", userID.String()).Msg("Error query ListByUserID")
		return nil, wrapDBError(err, "list sessions")
	}
	defer rows.Close()

	var sessions []domain.TelegramSession
	for rows.Next() {
		var s domain.TelegramSession
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.PhoneNumber, &s.APIID, &s.APIHashEncrypted, &s.SessionName,
			&s.SessionData, &s.AuthState, &s.TelegramUserID, &s.TelegramUsername, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			logger.Error().Err(err).Msg("Error scan ListByUserID")
			return nil, wrapDBError(err, "scan session")
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, wrapDBError(err, "rows error")
	}

	return sessions, nil
}

// ListAllActive retrieves all active Telegram sessions.
func (r *SessionRepository) ListAllActive(ctx context.Context) ([]domain.TelegramSession, error) {
	query := `
		SELECT id, user_id, phone_number, api_id, api_hash_encrypted, session_name,
			session_data, auth_state, COALESCE(telegram_user_id, 0), COALESCE(telegram_username, ''),
			is_active, created_at, updated_at
		FROM telegram_sessions WHERE is_active = true ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.Error().Err(err).Msg("Error query ListAllActive")
		return nil, wrapDBError(err, "list active sessions")
	}
	defer rows.Close()

	var sessions []domain.TelegramSession
	for rows.Next() {
		var s domain.TelegramSession
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.PhoneNumber, &s.APIID, &s.APIHashEncrypted, &s.SessionName,
			&s.SessionData, &s.AuthState, &s.TelegramUserID, &s.TelegramUsername, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			logger.Error().Err(err).Msg("Error scan ListAllActive")
			return nil, wrapDBError(err, "scan session")
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, wrapDBError(err, "rows error")
	}

	return sessions, nil
}

// Delete removes a Telegram session from the database.
func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM telegram_sessions WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return wrapDBError(err, "delete session")
	}
	if result.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

// UpdateSessionData updates the session data for a given session ID.
func (r *SessionRepository) UpdateSessionData(sessionID string, data []byte) error {
	query := `UPDATE telegram_sessions SET session_data = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.Exec(context.Background(), query, data, sessionID)
	if err != nil {
		return wrapDBError(err, "update session data")
	}
	if result.RowsAffected() == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

var _ domain.SessionRepository = (*SessionRepository)(nil)
