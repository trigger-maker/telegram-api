// Package postgres provides PostgreSQL repository implementations.
package postgres

import (
	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/jackc/pgx/v5/pgconn"
)

func wrapDBError(err error, op string) error {
	if err == nil {
		return nil
	}

	// Always log the original error
	if pgErr, ok := err.(*pgconn.PgError); ok {
		logger.Error().
			Err(err).
			Str("operation", op).
			Str("pg_code", pgErr.Code).
			Str("pg_message", pgErr.Message).
			Str("pg_detail", pgErr.Detail).
			Str("pg_constraint", pgErr.ConstraintName).
			Str("pg_table", pgErr.TableName).
			Str("pg_column", pgErr.ColumnName).
			Msg("PostgreSQL error")

		if pgErr.Code == "23505" { // unique_violation
			if pgErr.ConstraintName == "users_username_key" {
				return domain.ErrUserAlreadyExists
			}
			if pgErr.ConstraintName == "users_email_key" {
				return domain.ErrEmailAlreadyExists
			}
			// Any other unique violation
			logger.Warn().
				Str("constraint", pgErr.ConstraintName).
				Msg("Unhandled unique constraint violation")
		}
	} else {
		// Non-PostgreSQL error
		logger.Error().
			Err(err).
			Str("operation", op).
			Msg("Database error (non-PG)")
	}

	return domain.ErrDatabase
}
