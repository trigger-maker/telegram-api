package handler

import (
	"telegram-api/internal/domain"
	"telegram-api/internal/middleware"
	"telegram-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetContacts retrieves contacts for a session
// @Summary List contacts
// @Description Retrieves the list of Telegram contacts (with Redis cache and pagination)
// @Tags Contacts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param limit query int false "Result limit (default 50, max 200)"
// @Param offset query int false "Pagination offset"
// @Param refresh query bool false "Force cache refresh"
// @Success 200 {object} Response{data=domain.ContactsResponse}
// @Router /sessions/{id}/contacts [get].
func (h *ChatHandler) GetContacts(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid session ID"))
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(NewErrorResponse(401, "Unauthorized"))
	}

	req := domain.GetContactsRequest{
		Limit:   c.QueryInt("limit", 50),
		Offset:  c.QueryInt("offset", 0),
		Refresh: c.QueryBool("refresh", false),
	}

	logger.Info().
		Str("session_id", sessionID.String()).
		Int("limit", req.Limit).
		Int("offset", req.Offset).
		Bool("refresh", req.Refresh).
		Msg("GET contacts")

	result, err := h.chatService.GetContacts(c.Context(), userID, sessionID, req)
	if err != nil {
		logger.Error().Err(err).Msg("error getting contacts")
		return h.handleError(c, err)
	}

	logger.Info().
		Int("returned", len(result.Contacts)).
		Int("total", result.TotalCount).
		Bool("has_more", result.HasMore).
		Bool("from_cache", result.FromCache).
		Msg("contacts retrieved")

	return c.JSON(NewSuccessResponse(result))
}
