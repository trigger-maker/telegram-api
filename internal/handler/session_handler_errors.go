package handler

import (
	"fmt"

	"telegram-api/internal/domain"
	"telegram-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// handleSessionError handles session-related errors
func handleSessionError(c *fiber.Ctx, err error) error {
	switch err {
	case domain.ErrSessionNotFound:
		return c.Status(404).JSON(NewErrorResponse(404, "Session not found"))
	case domain.ErrSessionAlreadyExists:
		return c.Status(409).JSON(NewErrorResponse(409, "Session already exists for this number"))
	case domain.ErrCodeExpired:
		return c.Status(410).JSON(NewErrorResponse(410, "Code expired, request new"))
	case domain.ErrInvalidCode:
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid code"))
	case domain.ErrInvalidPassword:
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid 2FA password"))
	case domain.ErrAlreadyAuthenticated:
		return c.Status(409).JSON(NewErrorResponse(409, "Session already authenticated"))
	case domain.ErrInvalidPhoneNumber:
		return c.Status(400).JSON(NewErrorResponse(400, "Phone number required for SMS"))
	case domain.ErrTDataInvalid:
		return c.Status(422).JSON(NewErrorResponse(422, "Invalid tdata files"))
	case domain.ErrDatabase:
		logger.Error().Err(err).Msg("Database error in session")
		return c.Status(500).JSON(NewErrorResponse(500, "Database error"))
	case domain.ErrInternal:
		logger.Error().Err(err).Msg("Internal error in session")
		return c.Status(500).JSON(NewErrorResponse(500, "Internal error"))
	}

	if appErr, ok := err.(*domain.AppError); ok {
		logger.Error().
			Err(appErr.Err).
			Str("code", appErr.Code).
			Int("status", appErr.Status).
			Msg("AppError in session")
		return c.Status(appErr.Status).JSON(NewErrorResponse(appErr.Status, appErr.Message))
	}

	logger.Error().
		Err(err).
		Str("error_type", fmt.Sprintf("%T", err)).
		Msg("Unhandled error in session")

	return c.Status(500).JSON(NewErrorResponse(500, "Internal error"))
}
