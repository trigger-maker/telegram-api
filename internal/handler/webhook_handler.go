package handler

import (
	"telegram-api/internal/domain"
	"telegram-api/internal/telegram"

	"github.com/gofiber/fiber/v2"
)

// WebhookHandler handles webhook-related HTTP requests.
type WebhookHandler struct {
	webhookRepo domain.WebhookRepository
	sessionRepo domain.SessionRepository
	pool        telegram.SessionPoolInterface
}

// NewWebhookHandler creates a new WebhookHandler instance.
func NewWebhookHandler(
	webhookRepo domain.WebhookRepository,
	sessionRepo domain.SessionRepository,
	pool telegram.SessionPoolInterface,
) *WebhookHandler {
	return &WebhookHandler{
		webhookRepo: webhookRepo,
		sessionRepo: sessionRepo,
		pool:        pool,
	}
}

// RegisterRoutes registers webhook routes.
func (h *WebhookHandler) RegisterRoutes(r fiber.Router) {
	wh := r.Group("/sessions/:id/webhook")
	wh.Post("/", h.Configure)
	wh.Get("/", h.Get)
	wh.Delete("/", h.Delete)
	wh.Post("/start", h.StartListening)
	wh.Post("/stop", h.StopListening)

	r.Get("/pool/status", h.PoolStatus)
}
