package handler

import (
	"telegram-api/internal/domain"
	"telegram-api/internal/middleware"
	"telegram-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// List lists all sessions for a user
// @Summary List sessions
// @Description Returns all Telegram sessions for the user
// @Tags Sessions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=[]domain.TelegramSession}
// @Router /sessions [get]
func (h *SessionHandler) List(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(NewErrorResponse(401, "Not authenticated"))
	}

	sessions, err := h.service.ListSessions(c.Context(), userID)
	if err != nil {
		logger.Error().Err(err).Str("user_id", userID.String()).Msg("Error listing sessions")
		return c.Status(500).JSON(NewErrorResponse(500, "Error listing sessions"))
	}

	return c.JSON(NewSuccessResponse(sessions))
}

// Get retrieves a session by ID
// @Summary Get session
// @Description Returns session details. Use to check if QR was scanned (is_active=true).
// @Tags Sessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} handler.Response{data=domain.TelegramSession}
// @Failure 404 {object} handler.Response
// @Router /sessions/{id} [get]
func (h *SessionHandler) Get(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	session, err := h.service.GetSession(c.Context(), sessionID)
	if err != nil {
		return handleSessionError(c, err)
	}

	response := fiber.Map{
		"session": session,
	}

	if !session.IsActive {
		switch session.AuthState {
		case domain.SessionPending, domain.SessionCodeSent:
			response["status"] = "waiting"
			response["message"] = "Waiting for authentication..."
		case domain.SessionFailed:
			response["status"] = "failed"
			response["message"] = "Authentication failed. Create new session."
		}
	} else {
		response["status"] = "authenticated"
	}

	return c.JSON(NewSuccessResponse(response))
}

// Delete deletes a session
// @Summary Delete session
// @Description Deletes a Telegram session
// @Tags Sessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} handler.Response
// @Failure 404 {object} handler.Response
// @Router /sessions/{id} [delete]
func (h *SessionHandler) Delete(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	if err := h.service.DeleteSession(c.Context(), sessionID); err != nil {
		return handleSessionError(c, err)
	}

	return c.JSON(NewSuccessResponse(fiber.Map{"deleted": true}))
}
