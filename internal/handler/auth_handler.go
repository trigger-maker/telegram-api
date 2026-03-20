// Package handler provides HTTP request handlers.
package handler

import (
	"telegram-api/internal/domain"
	"telegram-api/internal/middleware"
	"telegram-api/internal/service"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// RegisterRoutes registers authentication routes.
func (h *AuthHandler) RegisterRoutes(r fiber.Router) {
	auth := r.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/logout", middleware.JWTMiddleware(h.auth), h.Logout)
	auth.Get("/me", middleware.JWTMiddleware(h.auth), h.Me)
}

// Register godoc
// @Summary Register user
// @Description Creates a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body domain.CreateUserRequest true "User data"
// @Success 201 {object} handler.Response{data=domain.UserInfo}
// @Failure 400 {object} handler.Response
// @Failure 409 {object} handler.Response
// @Router /auth/register [post].
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req domain.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid body"))
	}
	if errs := ValidateStruct(&req); errs != nil {
		return c.Status(422).JSON(Response{Success: false, Error: &ErrorResponse{Code: 422, Details: errs}})
	}
	user, err := h.auth.Register(c.Context(), &req)
	if err != nil {
		return handleErr(c, err)
	}
	return c.Status(201).JSON(NewSuccessResponse(user.ToUserInfo()))
}

// Login godoc
// @Summary Login
// @Description Authenticates user and returns tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body domain.LoginRequest true "Credenciales"
// @Success 200 {object} handler.Response{data=domain.LoginResponse}
// @Failure 401 {object} handler.Response
// @Router /auth/login [post].
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid body"))
	}
	resp, err := h.auth.Login(c.Context(), &req, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(NewSuccessResponse(resp))
}

// Refresh godoc
// @Summary Renovar tokens
// @Description Generates new tokens using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} handler.Response{data=domain.LoginResponse}
// @Failure 401 {object} handler.Response
// @Router /auth/refresh [post].
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid body"))
	}
	resp, err := h.auth.RefreshTokens(c.Context(), req.RefreshToken, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(NewSuccessResponse(resp))
}

// Logout godoc
// @Summary Logout
// @Description Revokes the refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} handler.Response
// @Router /auth/logout [post].
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(NewErrorResponse(400, "Invalid request body"))
	}
	if err := h.auth.Logout(c.Context(), req.RefreshToken); err != nil {
		return handleErr(c, err)
	}
	return c.JSON(NewSuccessResponse(fiber.Map{"message": "ok"}))
}

// Me godoc
// @Summary Current user info
// @Description Returns information of the authenticated user
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handler.Response{data=domain.UserInfo}
// @Failure 401 {object} handler.Response
// @Router /auth/me [get].
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	user, err := h.auth.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(NewErrorResponse(404, "User not found"))
	}
	return c.JSON(NewSuccessResponse(user.ToUserInfo()))
}

func handleErr(c *fiber.Ctx, err error) error {
	switch err {
	case domain.ErrUserAlreadyExists, domain.ErrEmailAlreadyExists:
		return c.Status(409).JSON(NewErrorResponse(409, err.Error()))
	case domain.ErrInvalidCredentials, domain.ErrInvalidToken, domain.ErrTokenExpired:
		return c.Status(401).JSON(NewErrorResponse(401, err.Error()))
	case domain.ErrUserInactive:
		return c.Status(403).JSON(NewErrorResponse(403, err.Error()))
	default:
		return c.Status(500).JSON(NewErrorResponse(500, "Internal error"))
	}
}
