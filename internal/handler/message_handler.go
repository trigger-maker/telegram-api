package handler

import (
	"telegram-api/internal/service"

	"github.com/gofiber/fiber/v2"
)

// MessageHandler handles message-related HTTP requests.
type MessageHandler struct {
	service service.MessageServiceInterface
}

// NewMessageHandler creates a new MessageHandler instance.
func NewMessageHandler(s service.MessageServiceInterface) *MessageHandler {
	return &MessageHandler{service: s}
}

// RegisterRoutes registers message routes.
func (h *MessageHandler) RegisterRoutes(r fiber.Router) {
	msg := r.Group("/sessions/:id/messages")
	msg.Post("/text", h.SendText)
	msg.Post("/photo", h.SendPhoto)
	msg.Post("/video", h.SendVideo)
	msg.Post("/audio", h.SendAudio)
	msg.Post("/file", h.SendFile)
	msg.Post("/bulk", h.SendBulk)

	r.Get("/messages/:jobId/status", h.GetStatus)
}
