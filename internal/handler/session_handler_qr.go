package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RegenerateQR regenerates QR code for a session
// @Summary Regenerate session QR
// @Description Generates a new QR code for an existing session that failed or expired
// @Tags Sessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} handler.Response
// @Failure 400 {object} handler.Response
// @Failure 404 {object} handler.Response
// @Router /sessions/{id}/qr/regenerate [post].
func (h *SessionHandler) RegenerateQR(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	qrImageB64, err := h.service.RegenerateQR(c.Context(), sessionID)
	if err != nil {
		return handleSessionError(c, err)
	}

	return c.JSON(NewSuccessResponse(fiber.Map{
		"session_id":      sessionID,
		"qr_image_base64": qrImageB64,
		"message":         "QR regenerated. System listens automatically (3 attempts, 2 min each).",
	}))
}
