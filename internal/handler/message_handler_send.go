package handler

import (
	"telegram-api/internal/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SendText sends a text message
// @Summary Send text
// @Description Sends a simple text message
// @Tags Messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param body body domain.TextMessageRequest true "Text message"
// @Success 202 {object} Response{data=domain.MessageResponse}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /sessions/{id}/messages/text [post].
func (h *MessageHandler) SendText(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	var req domain.TextMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid JSON"))
	}

	if req.To == "" || req.Text == "" {
		return c.Status(400).JSON(NewErrorResponse(400, "Fields 'to' and 'text' required"))
	}

	internal := &domain.SendMessageRequest{
		To:   req.To,
		Text: req.Text,
		Type: domain.MessageTypeText,
	}

	resp, err := h.service.SendMessage(c.Context(), sessionID, internal)
	if err != nil {
		return handleMessageError(c, err)
	}

	return c.Status(202).JSON(NewSuccessResponse(resp))
}

// parseSessionID extracts and validates session ID from request.
func parseSessionID(c *fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(c.Params("id"))
}

// sendMediaMessage sends a media message with the given parameters.
func (h *MessageHandler) sendMediaMessage(
	c *fiber.Ctx,
	sessionID uuid.UUID,
	msgType domain.MessageType,
	to, mediaURL, caption string,
) error {
	internal := &domain.SendMessageRequest{
		To:       to,
		Type:     msgType,
		MediaURL: mediaURL,
		Caption:  caption,
	}

	resp, err := h.service.SendMessage(c.Context(), sessionID, internal)
	if err != nil {
		return handleMessageError(c, err)
	}

	return c.Status(202).JSON(NewSuccessResponse(resp))
}

// SendPhoto sends a photo message
// @Summary Send photo
// @Description Sends an image with optional caption
// @Tags Messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param body body domain.PhotoMessageRequest true "Photo"
// @Success 202 {object} Response{data=domain.MessageResponse}
// @Failure 400 {object} Response
// @Router /sessions/{id}/messages/photo [post].
func (h *MessageHandler) SendPhoto(c *fiber.Ctx) error {
	sessionID, err := parseSessionID(c)
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	var req domain.PhotoMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid JSON"))
	}

	if req.To == "" || req.PhotoURL == "" {
		return c.Status(400).JSON(NewErrorResponse(400, "Fields 'to' and 'photo_url' required"))
	}

	return h.sendMediaMessage(c, sessionID, domain.MessageTypePhoto, req.To, req.PhotoURL, req.Caption)
}

// SendVideo sends a video message
// @Summary Send video
// @Description Sends a video with optional caption
// @Tags Messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param body body domain.VideoMessageRequest true "Video"
// @Success 202 {object} Response{data=domain.MessageResponse}
// @Failure 400 {object} Response
// @Router /sessions/{id}/messages/video [post].
func (h *MessageHandler) SendVideo(c *fiber.Ctx) error {
	sessionID, err := parseSessionID(c)
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	var req domain.VideoMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid JSON"))
	}

	if req.To == "" || req.VideoURL == "" {
		return c.Status(400).JSON(NewErrorResponse(400, "Fields 'to' and 'video_url' required"))
	}

	return h.sendMediaMessage(c, sessionID, domain.MessageTypeVideo, req.To, req.VideoURL, req.Caption)
}

// SendAudio sends an audio message
// @Summary Send audio
// @Description Sends an audio file
// @Tags Messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param body body domain.AudioMessageRequest true "Audio"
// @Success 202 {object} Response{data=domain.MessageResponse}
// @Failure 400 {object} Response
// @Router /sessions/{id}/messages/audio [post].
func (h *MessageHandler) SendAudio(c *fiber.Ctx) error {
	sessionID, err := parseSessionID(c)
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	var req domain.AudioMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid JSON"))
	}

	if req.To == "" || req.AudioURL == "" {
		return c.Status(400).JSON(NewErrorResponse(400, "Fields 'to' and 'audio_url' required"))
	}

	return h.sendMediaMessage(c, sessionID, domain.MessageTypeAudio, req.To, req.AudioURL, req.Caption)
}

// SendFile sends a file message
// @Summary Send document
// @Description Sends a file/document
// @Tags Messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param body body domain.FileMessageRequest true "File"
// @Success 202 {object} Response{data=domain.MessageResponse}
// @Failure 400 {object} Response
// @Router /sessions/{id}/messages/file [post].
func (h *MessageHandler) SendFile(c *fiber.Ctx) error {
	sessionID, err := parseSessionID(c)
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	var req domain.FileMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid JSON"))
	}

	if req.To == "" || req.FileURL == "" {
		return c.Status(400).JSON(NewErrorResponse(400, "Fields 'to' and 'file_url' required"))
	}

	return h.sendMediaMessage(c, sessionID, domain.MessageTypeFile, req.To, req.FileURL, req.Caption)
}

// SendBulk sends bulk text messages with delay
// @Summary Bulk send
// @Description Sends text message to multiple recipients with delay
// @Tags Messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param body body domain.BulkTextRequest true "Bulk message"
// @Success 202 {object} Response{data=[]domain.MessageResponse}
// @Failure 400 {object} Response
// @Router /sessions/{id}/messages/bulk [post].
func (h *MessageHandler) SendBulk(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid ID"))
	}

	var req domain.BulkTextRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid JSON"))
	}

	if len(req.Recipients) == 0 || req.Text == "" {
		return c.Status(400).JSON(NewErrorResponse(400, "Fields 'recipients' and 'text' required"))
	}

	internal := &domain.BulkMessageRequest{
		Recipients: req.Recipients,
		Text:       req.Text,
		Type:       domain.MessageTypeText,
		DelayMs:    req.DelayMs,
	}

	resp, err := h.service.SendBulk(c.Context(), sessionID, internal)
	if err != nil {
		return handleMessageError(c, err)
	}

	return c.Status(202).JSON(NewSuccessResponse(resp))
}
