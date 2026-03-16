package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SubmitPassword submits 2FA password for accounts with 2FA enabled
// @Summary Submit 2FA password
// @Description Submits 2FA password for accounts with 2FA enabled
// @Tags Sessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param body body domain.SubmitPasswordRequest true "Password"
// @Success 200 {object} handler.Response{data=domain.TelegramSession}
// @Failure 400 {object} handler.Response
// @Failure 404 {object} handler.Response
// @Failure 409 {object} handler.Response
// @Router /sessions/{id}/submit-password [post].
func (h *SessionHandler) SubmitPassword(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	var req struct {
		Password string `json:"password" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid JSON body"))
	}

	if req.Password == "" {
		return c.Status(400).JSON(NewErrorResponse(400, "password is required"))
	}

	session, err := h.service.SubmitPassword(c.Context(), sessionID, req.Password)
	if err != nil {
		return handleSessionError(c, err)
	}

	return c.JSON(NewSuccessResponse(session))
}
