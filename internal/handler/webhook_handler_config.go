package handler

import (
	"time"

	"telegram-api/internal/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Configure configures a webhook for a session
// @Summary Configure webhook
// @Description Configures webhook URL to receive session events
// @Tags Webhooks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param body body domain.WebhookCreateRequest true "Configuration"
// @Success 200 {object} Response{data=domain.WebhookResponse}
// @Router /sessions/{id}/webhook [post].
func (h *WebhookHandler) Configure(c *fiber.Ctx) error {
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

	var req domain.WebhookCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid JSON"))
	}

	if req.URL == "" {
		return c.Status(400).JSON(NewErrorResponse(400, "URL required"))
	}

	if req.MaxRetries == 0 {
		req.MaxRetries = 3
	}
	if req.TimeoutMs == 0 {
		req.TimeoutMs = 5000
	}

	now := time.Now()
	webhook := &domain.WebhookConfig{
		ID:         uuid.New(),
		SessionID:  sessionID,
		URL:        req.URL,
		Secret:     req.Secret,
		Events:     req.Events,
		IsActive:   true,
		MaxRetries: req.MaxRetries,
		TimeoutMs:  req.TimeoutMs,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.webhookRepo.Create(c.Context(), webhook); err != nil {
		return c.Status(500).JSON(NewErrorResponse(500, "Error saving webhook"))
	}

	return c.JSON(NewSuccessResponse(domain.WebhookResponse{
		ID:        webhook.ID,
		SessionID: sessionID,
		URL:       webhook.URL,
		Events:    webhook.Events,
		IsActive:  webhook.IsActive,
	}))
}

// Get retrieves webhook configuration
// @Summary Get webhook
// @Description Returns current webhook configuration
// @Tags Webhooks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} Response{data=domain.WebhookConfig}
// @Router /sessions/{id}/webhook [get].
func (h *WebhookHandler) Get(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	webhook, err := h.webhookRepo.GetBySessionID(c.Context(), sessionID)
	if err != nil {
		return c.Status(500).JSON(NewErrorResponse(500, "Error getting webhook"))
	}
	if webhook == nil {
		return c.Status(404).JSON(NewErrorResponse(404, "Webhook not configured"))
	}

	return c.JSON(NewSuccessResponse(webhook))
}

// Delete removes webhook configuration
// @Summary Delete webhook
// @Description Removes webhook configuration
// @Tags Webhooks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} Response
// @Router /sessions/{id}/webhook [delete].
func (h *WebhookHandler) Delete(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	h.pool.StopSession(sessionID)

	if err := h.webhookRepo.Delete(c.Context(), sessionID); err != nil {
		return c.Status(500).JSON(NewErrorResponse(500, "Error deleting webhook"))
	}

	return c.JSON(NewSuccessResponse(fiber.Map{"deleted": true}))
}
