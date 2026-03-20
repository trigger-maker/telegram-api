package postgres

import (
	"context"
	"errors"
	"time"

	"telegram-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL queries for refresh tokens.
const (
	// #nosec G101 -- SQL queries without hardcoded credentials
	queryCreateToken = `
		INSERT INTO refresh_tokens (id, user_id, token_hash, device_info, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5::inet, $6, $7)`

	// #nosec G101 -- SQL queries without hardcoded credentials
	queryGetTokenByHash = `
		SELECT id, user_id, token_hash, device_info, ip_address, expires_at, revoked_at, created_at
		FROM refresh_tokens 
		WHERE token_hash = $1`

	// #nosec G101 -- SQL queries without hardcoded credentials
	queryRevokeToken = `
		UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1 AND revoked_at IS NULL`

	// #nosec G101 -- SQL queries without hardcoded credentials
	queryRevokeAllUserTokens = `
		UPDATE refresh_tokens SET revoked_at = $2 WHERE user_id = $1 AND revoked_at IS NULL`

	// #nosec G101 -- SQL queries without hardcoded credentials
	queryDeleteExpiredTokens = `
		DELETE FROM refresh_tokens WHERE expires_at < $1 OR revoked_at IS NOT NULL`

	// #nosec G101 -- SQL queries without hardcoded credentials
	queryGetActiveTokensByUser = `
		SELECT id, user_id, token_hash, device_info, ip_address, expires_at, revoked_at, created_at
		FROM refresh_tokens 
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > $2
		ORDER BY created_at DESC`

	// #nosec G101 -- SQL queries without hardcoded credentials
	queryCountActiveTokensByUser = `
		SELECT COUNT(*) FROM refresh_tokens 
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > $2`
)

// RefreshTokenRepository implements domain.RefreshTokenRepository.
// Single Responsibility: Only handles refresh token operations.
type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

// NewRefreshTokenRepository creates a new instance of the repository.
func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

// Create creates a new refresh token.
func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	_, err := r.pool.Exec(ctx, queryCreateToken,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.DeviceInfo,
		nullableString(token.IPAddress),
		token.ExpiresAt,
		token.CreatedAt,
	)
	if err != nil {
		return wrapDBError(err, "create refresh token")
	}
	return nil
}

// GetByTokenHash gets a token by its hash.
func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	token := &domain.RefreshToken{}
	var ipAddr *string

	err := r.pool.QueryRow(ctx, queryGetTokenByHash, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.DeviceInfo,
		&ipAddr,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}
		return nil, wrapDBError(err, "get token by hash")
	}

	if ipAddr != nil {
		token.IPAddress = *ipAddr
	}

	return token, nil
}

// Revoke revokes a specific token.
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, queryRevokeToken, id, time.Now())
	if err != nil {
		return wrapDBError(err, "revoke token")
	}
	if result.RowsAffected() == 0 {
		return domain.ErrInvalidToken
	}
	return nil
}

// RevokeAllForUser revokes all tokens of a user.
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, queryRevokeAllUserTokens, userID, time.Now())
	if err != nil {
		return wrapDBError(err, "revoke all user tokens")
	}
	return nil
}

// DeleteExpired deletes expired or revoked tokens.
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.pool.Exec(ctx, queryDeleteExpiredTokens, time.Now())
	if err != nil {
		return 0, wrapDBError(err, "delete expired tokens")
	}
	return result.RowsAffected(), nil
}

// GetActiveByUserID gets all active tokens of a user.
func (r *RefreshTokenRepository) GetActiveByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]*domain.RefreshToken, error) {
	rows, err := r.pool.Query(ctx, queryGetActiveTokensByUser, userID, time.Now())
	if err != nil {
		return nil, wrapDBError(err, "get active tokens")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't override existing error.
			_ = err
		}
	}()

	var tokens []*domain.RefreshToken
	for rows.Next() {
		token := &domain.RefreshToken{}
		var ipAddr *string

		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.TokenHash,
			&token.DeviceInfo,
			&ipAddr,
			&token.ExpiresAt,
			&token.RevokedAt,
			&token.CreatedAt,
		)
		if err != nil {
			return nil, wrapDBError(err, "scan token")
		}

		if ipAddr != nil {
			token.IPAddress = *ipAddr
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, wrapDBError(err, "iterate tokens")
	}

	return tokens, nil
}

// CountActiveByUserID counts active tokens of a user.
func (r *RefreshTokenRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, queryCountActiveTokensByUser, userID, time.Now()).Scan(&count)
	if err != nil {
		return 0, wrapDBError(err, "count active tokens")
	}
	return count, nil
}

// nullableString converts empty string to nil for nullable fields.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Verificación en tiempo de compilación.
var _ domain.RefreshTokenRepository = (*RefreshTokenRepository)(nil)
