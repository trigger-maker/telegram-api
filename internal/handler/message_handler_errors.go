package handler

import (
	"telegram-api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// handleMessageError handles message-related errors.
func handleMessageError(c *fiber.Ctx, err error) error {
	switch err {
	case domain.ErrSessionNotFound:
		return c.Status(404).JSON(NewErrorResponse(404, "Session not found"))
	case domain.ErrSessionNotActive:
		return c.Status(400).JSON(NewErrorResponse(400, "Session not authenticated"))
	default:
		if appErr, ok := err.(*domain.AppError); ok {
			return c.Status(appErr.Status).JSON(NewErrorResponse(appErr.Status, appErr.Message))
		}
		return c.Status(500).JSON(NewErrorResponse(500, err.Error()))
	}
}
