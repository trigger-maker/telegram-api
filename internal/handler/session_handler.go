package handler

import (
	"telegram-api/internal/domain"
	"telegram-api/internal/middleware"
	"telegram-api/internal/service"
	"telegram-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SessionHandler handles session-related HTTP requests.
type SessionHandler struct {
	service service.SessionServiceInterface
}

// NewSessionHandler creates a new SessionHandler instance.
func NewSessionHandler(s service.SessionServiceInterface) *SessionHandler {
	return &SessionHandler{service: s}
}

// RegisterRoutes registers session routes.
func (h *SessionHandler) RegisterRoutes(r fiber.Router) {
	sessions := r.Group("/sessions")
	sessions.Post("/", h.Create)
	sessions.Post("/:id/verify", h.VerifyCode)
	sessions.Post("/:id/qr/regenerate", h.RegenerateQR)
	sessions.Post("/:id/submit-password", h.SubmitPassword)
	sessions.Post("/import-tdata", h.ImportTData)
	sessions.Get("/", h.List)
	sessions.Get("/:id", h.Get)
	sessions.Delete("/:id", h.Delete)
}

// Create godoc
// @Summary Create Telegram session
// @Description Initiates authentication with Telegram (SMS or QR). For QR, the system listens automatically in background.
// @Tags Sessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateSessionRequest true "Credenciales Telegram"
// @Success 201 {object} handler.Response
// @Failure 400 {object} handler.Response
// @Failure 409 {object} handler.Response
// @Router /sessions [post].
func (h *SessionHandler) Create(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(NewErrorResponse(401, "No autenticado"))
	}

	var req domain.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error().Err(err).Msg("❌ Error parseando body en Create session")
		return c.Status(400).JSON(NewErrorResponse(400, "JSON inválido"))
	}

	if errs := ValidateStruct(&req); errs != nil {
		logger.Warn().Interface("errors", errs).Msg("⚠️ Validación fallida en Create session")
		return c.Status(400).JSON(Response{Success: false, Error: &ErrorResponse{Code: 400, Details: errs}})
	}

	logger.Debug().
		Str("user_id", userID.String()).
		Str("session_name", req.SessionName).
		Str("auth_method", string(req.AuthMethod)).
		Int("api_id", req.APIID).
		Msg("📝 Intentando crear sesión...")

	session, data, err := h.service.CreateSession(c.Context(), userID, &req)
	if err != nil {
		return handleSessionError(c, err)
	}

	response := fiber.Map{
		"session": session,
	}

	if req.AuthMethod == domain.AuthMethodQR {
		response["qr_image_base64"] = data
		response["message"] = "QR generated. The system listens automatically " +
			"(3 attempts, 2 min each). Use GET /sessions/:id to verify status."
	} else {
		response["phone_code_hash"] = data
		response["next_step"] = "POST /sessions/" + session.ID.String() + "/verify with {code}"
	}

	return c.Status(201).JSON(NewSuccessResponse(response))
}

// VerifyCode godoc
// @Summary Verify SMS code
// @Description Completes authentication with the code received by SMS
// @Tags Sessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Param body body domain.VerifyCodeRequest true "Código SMS"
// @Success 200 {object} handler.Response{data=domain.TelegramSession}
// @Failure 400 {object} handler.Response
// @Failure 410 {object} handler.Response
// @Router /sessions/{id}/verify [post].
func (h *SessionHandler) VerifyCode(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "ID inválido"))
	}

	var req domain.VerifyCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "JSON inválido"))
	}

	session, _, err := h.service.VerifyCode(c.Context(), sessionID, req.Code)
	if err != nil {
		return handleSessionError(c, err)
	}

	return c.JSON(NewSuccessResponse(session))
}

// List godoc
// @Summary List sessions
// @Description Returns all Telegram sessions of the user
// @Tags Sessions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=[]domain.TelegramSession}
// @Router /sessions [get].
func (h *SessionHandler) List(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(401).JSON(NewErrorResponse(401, "No autenticado"))
	}

	sessions, err := h.service.ListSessions(c.Context(), userID)
	if err != nil {
		logger.Error().Err(err).Str("user_id", userID.String()).Msg("Error listando sesiones")
		return c.Status(500).JSON(NewErrorResponse(500, "Error listando sesiones"))
	}

	return c.JSON(NewSuccessResponse(sessions))
}

// Get godoc
// @Summary Get session
// @Description Returns details of a session. Use to verify if QR was scanned (is_active=true).
// @Tags Sessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} handler.Response{data=domain.TelegramSession}
// @Failure 404 {object} handler.Response
// @Router /sessions/{id} [get].
func (h *SessionHandler) Get(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "ID inválido"))
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
			response["message"] = "Esperando autenticación..."
		case domain.SessionFailed:
			response["status"] = "failed"
			response["message"] = "Autenticación fallida. Cree nueva sesión."
		}
	} else {
		response["status"] = "authenticated"
	}

	return c.JSON(NewSuccessResponse(response))
}

// Delete godoc
// @Summary Delete session
// @Description Deletes a Telegram session
// @Tags Sessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} handler.Response
// @Failure 404 {object} handler.Response
// @Router /sessions/{id} [delete].
func (h *SessionHandler) Delete(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "ID inválido"))
	}

	if err := h.service.DeleteSession(c.Context(), sessionID); err != nil {
		return handleSessionError(c, err)
	}

	return c.JSON(NewSuccessResponse(fiber.Map{"deleted": true}))
}
