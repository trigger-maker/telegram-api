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

// SQL queries as private constants (Single Responsibility: only user SQL).
const (
	// #nosec G101 -- SQL queries without hardcoded credentials
	queryCreateUser = `
		INSERT INTO users (id, username, email, password_hash, is_active, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	queryGetUserByID = `
		SELECT id, username, email, password_hash, is_active, role, last_login_at, created_at, updated_at
		FROM users WHERE id = $1`

	queryGetUserByUsername = `
		SELECT id, username, email, password_hash, is_active, role, last_login_at, created_at, updated_at
		FROM users WHERE username = $1`

	queryGetUserByEmail = `
		SELECT id, username, email, password_hash, is_active, role, last_login_at, created_at, updated_at
		FROM users WHERE email = $1`

	queryUpdateUser = `
		UPDATE users 
		SET username = $2, email = $3, is_active = $4, role = $5, updated_at = $6
		WHERE id = $1`

	queryUpdateLastLogin = `
		UPDATE users SET last_login_at = $2, updated_at = $2 WHERE id = $1`

	queryDeleteUser = `
		UPDATE users SET is_active = false, updated_at = $2 WHERE id = $1`

	queryExistsByUsername = `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	queryExistsByEmail = `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	// #nosec G101 -- SQL queries without hardcoded credentials
	queryUpdatePassword = `
		UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`
)

// UserRepository implements domain.UserRepository using PostgreSQL.
// Principle: Single Responsibility - Only handles user operations.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new repository instance.
// Dependency Inversion: Receives the pool as a dependency.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create creates a new user in the database.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx, queryCreateUser,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.IsActive,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return wrapDBError(err, "create user")
	}
	return nil
}

// GetByID gets a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user := &domain.User{}
	err := r.pool.QueryRow(ctx, queryGetUserByID, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.Role,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, wrapDBError(err, "get user by ID")
	}
	return user, nil
}

// GetByUsername gets a user by username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user := &domain.User{}
	err := r.pool.QueryRow(ctx, queryGetUserByUsername, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.Role,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, wrapDBError(err, "get user by username")
	}
	return user, nil
}

// GetByEmail gets a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	err := r.pool.QueryRow(ctx, queryGetUserByEmail, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.Role,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, wrapDBError(err, "get user by email")
	}
	return user, nil
}

// Update updates user data.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	result, err := r.pool.Exec(ctx, queryUpdateUser,
		user.ID,
		user.Username,
		user.Email,
		user.IsActive,
		user.Role,
		time.Now(),
	)
	if err != nil {
		return wrapDBError(err, "update user")
	}
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// UpdateLastLogin updates the last login date.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, queryUpdateLastLogin, id, time.Now())
	if err != nil {
		return wrapDBError(err, "update last login")
	}
	return nil
}

// Delete performs a soft delete of the user.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, queryDeleteUser, id, time.Now())
	if err != nil {
		return wrapDBError(err, "delete user")
	}
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// ExistsByUsername checks if a user with that username exists.
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, queryExistsByUsername, username).Scan(&exists)
	if err != nil {
		return false, wrapDBError(err, "check existence by username")
	}
	return exists, nil
}

// ExistsByEmail checks if a user with that email exists.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, queryExistsByEmail, email).Scan(&exists)
	if err != nil {
		return false, wrapDBError(err, "check existence by email")
	}
	return exists, nil
}

// UpdatePassword updates a user's password.
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	result, err := r.pool.Exec(ctx, queryUpdatePassword, id, passwordHash, time.Now())
	if err != nil {
		return wrapDBError(err, "update password")
	}
	if result.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// Compile-time verification that it implements the interface.
var _ domain.UserRepository = (*UserRepository)(nil)
