package handler

import (
	"telegram-api/internal/domain"
	"telegram-api/internal/middleware"
	"telegram-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// InvalidateCache invalidates cache for a session
// @Summary Invalidate cache
// @Description Invalidates session cache (contacts, chats, or all)
// @Tags Cache
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param type query string false "Cache type to invalidate (contacts, chats, all)" default(all)
// @Success 200 {object} Response
// @Router /sessions/{id}/cache [delete]
func (h *ChatHandler) InvalidateCache(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid session ID"))
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(NewErrorResponse(401, "Unauthorized"))
	}

	_, err = h.chatService.GetContacts(c.Context(), userID, sessionID, domain.GetContactsRequest{Limit: 1})
	if err != nil {
		return h.handleError(c, err)
	}

	cacheType := c.Query("type", "all")

	if err := h.chatService.InvalidateCache(c.Context(), sessionID, cacheType); err != nil {
		logger.Error().Err(err).Str("type", cacheType).Msg("error invalidating cache")
		return c.Status(500).JSON(NewErrorResponse(500, "Error invalidating cache"))
	}

	logger.Info().
		Str("session_id", sessionID.String()).
		Str("type", cacheType).
		Msg("cache invalidated")

	return c.JSON(NewSuccessResponse(fiber.Map{
		"message":    "Cache invalidated successfully",
		"cache_type": cacheType,
	}))
}
