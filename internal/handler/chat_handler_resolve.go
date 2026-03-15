package handler

import (
	"telegram-api/internal/domain"
	"telegram-api/internal/middleware"
	"telegram-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ResolvePeer resolves a username or phone number to a peer
// @Summary Resolve username or phone
// @Description Resolves a @username or phone number to a Telegram peer (with cache)
// @Tags Contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param body body domain.ResolveRequest true "Username or phone"
// @Success 200 {object} Response{data=domain.ResolvedPeer}
// @Router /sessions/{id}/resolve [post]
func (h *ChatHandler) ResolvePeer(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid session ID"))
	}

	var req domain.ResolveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid JSON"))
	}

	if req.Username == "" && req.Phone == "" {
		return c.Status(400).JSON(NewErrorResponse(400, "Username or phone required"))
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(NewErrorResponse(401, "Unauthorized"))
	}

	logger.Info().
		Str("session_id", sessionID.String()).
		Str("username", req.Username).
		Str("phone", req.Phone).
		Msg("POST resolve peer")

	result, err := h.chatService.ResolvePeer(c.Context(), userID, sessionID, req)
	if err != nil {
		logger.Error().Err(err).Msg("error resolving peer")
		return h.handleError(c, err)
	}

	logger.Info().Int64("peer_id", result.ID).Str("type", string(result.Type)).Msg("peer resolved")
	return c.JSON(NewSuccessResponse(result))
}
