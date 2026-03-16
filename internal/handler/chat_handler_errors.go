package handler

import (
	"telegram-api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// handleError handles chat-related errors.
func (h *ChatHandler) handleError(c *fiber.Ctx, err error) error {
	switch err {
	case domain.ErrSessionNotFound:
		return c.Status(404).JSON(NewErrorResponse(404, "Session not found"))
	case domain.ErrUnauthorized:
		return c.Status(403).JSON(NewErrorResponse(403, "No access to this session"))
	case domain.ErrSessionInactive:
		return c.Status(400).JSON(NewErrorResponse(400, "Session not authenticated"))
	default:
		return c.Status(500).JSON(NewErrorResponse(500, err.Error()))
	}
}
