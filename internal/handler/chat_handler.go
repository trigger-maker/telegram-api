package handler

import (
	"telegram-api/internal/service"

	"github.com/gofiber/fiber/v2"
)

// ChatHandler handles chat-related HTTP requests
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler creates a new ChatHandler instance
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// RegisterRoutes registers chat routes
func (h *ChatHandler) RegisterRoutes(r fiber.Router) {
	chats := r.Group("/sessions/:id/chats")
	chats.Get("/", h.GetChats)
	chats.Get("/:chatId", h.GetChatInfo)
	chats.Get("/:chatId/history", h.GetChatHistory)

	contacts := r.Group("/sessions/:id/contacts")
	contacts.Get("/", h.GetContacts)

	r.Post("/sessions/:id/resolve", h.ResolvePeer)
	r.Delete("/sessions/:id/cache", h.InvalidateCache)
}
