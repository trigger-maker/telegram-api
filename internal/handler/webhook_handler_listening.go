package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// StartListening starts listening for Telegram events
// @Summary Start listening
// @Description Starts listening for Telegram events for this session
// @Tags Webhooks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} Response
// @Router /sessions/{id}/webhook/start [post]
func (h *WebhookHandler) StartListening(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	sess, err := h.sessionRepo.GetByID(c.Context(), sessionID)
	if err != nil {
		return c.Status(404).JSON(NewErrorResponse(404, "Session not found"))
	}
	if !sess.IsActive {
		return c.Status(400).JSON(NewErrorResponse(400, "Session not authenticated"))
	}

	webhook, _ := h.webhookRepo.GetBySessionID(c.Context(), sessionID)
	if webhook == nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Configure a webhook first"))
	}

	if err := h.pool.StartSession(c.Context(), sess); err != nil {
		return c.Status(500).JSON(NewErrorResponse(500, err.Error()))
	}

	return c.JSON(NewSuccessResponse(fiber.Map{
		"status":     "listening",
		"session_id": sessionID,
		"webhook":    webhook.URL,
	}))
}

// StopListening stops listening for Telegram events
// @Summary Stop listening
// @Description Stops listening for Telegram events for this session
// @Tags Webhooks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} Response
// @Router /sessions/{id}/webhook/stop [post]
func (h *WebhookHandler) StopListening(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	h.pool.StopSession(sessionID)

	return c.JSON(NewSuccessResponse(fiber.Map{
		"status":     "stopped",
		"session_id": sessionID,
	}))
}
