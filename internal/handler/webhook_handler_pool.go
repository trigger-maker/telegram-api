package handler

import (
	"github.com/gofiber/fiber/v2"
)

// PoolStatus returns information about active listening sessions
// @Summary Pool status
// @Description Returns information about active listening sessions
// @Tags Webhooks
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response
// @Router /pool/status [get]
func (h *WebhookHandler) PoolStatus(c *fiber.Ctx) error {
	activeIDs := h.pool.ListActive()

	sessions := make([]fiber.Map, 0, len(activeIDs))
	for _, id := range activeIDs {
		if active, ok := h.pool.GetActiveSession(id); ok {
			sessions = append(sessions, fiber.Map{
				"session_id":   id,
				"session_name": active.SessionName,
				"telegram_id":  active.TelegramID,
				"started_at":   active.StartedAt,
				"is_connected": active.IsConnected,
			})
		}
	}

	return c.JSON(NewSuccessResponse(fiber.Map{
		"active_count": len(sessions),
		"sessions":     sessions,
	}))
}
