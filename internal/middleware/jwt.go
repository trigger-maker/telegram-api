package middleware

import (
	"strings"

	"telegram-api/internal/domain"
	"telegram-api/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	// ContextKeyUserID is the context key for user ID.
	ContextKeyUserID = "userID"
	// ContextKeyUsername is the context key for username.
	ContextKeyUsername = "username"
	// ContextKeyRole is the context key for user role.
	ContextKeyRole = "userRole"
)

// JWTMiddleware returns a JWT authentication middleware.
func JWTMiddleware(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "MISSING_TOKEN",
					"message": "Authorization token required",
				},
			})
		}

		// Extract token: accepts "Bearer <token>" or just "<token>"
		token := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = authHeader[7:]
		}

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INVALID_TOKEN",
					"message": "Invalid or expired token",
				},
			})
		}

		// Parse UUID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INVALID_TOKEN",
					"message": "Malformed token",
				},
			})
		}

		// Verify user exists in database
		user, err := authService.GetUserByID(c.Context(), userID)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "USER_NOT_FOUND",
					"message": "User not found or deleted",
				},
			})
		}
		if !user.IsActive {
			return c.Status(403).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "USER_INACTIVE",
					"message": "User account is inactive",
				},
			})
		}

		c.Locals(ContextKeyUserID, userID)
		c.Locals(ContextKeyUsername, claims.Username)
		c.Locals(ContextKeyRole, claims.Role)

		return c.Next()
	}
}

// RequireRole returns a middleware that requires specific user roles.
func RequireRole(roles ...domain.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals(ContextKeyRole).(domain.Role)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "Access denied",
				},
			})
		}

		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "FORBIDDEN",
				"message": "No permission for this action",
			},
		})
	}
}

// GetUserID retrieves the user ID from the request context.
func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userID, ok := c.Locals(ContextKeyUserID).(uuid.UUID)
	if !ok {
		return uuid.Nil, domain.ErrUnauthorized
	}
	return userID, nil
}

// GetUsername retrieves the username from the request context.
func GetUsername(c *fiber.Ctx) string {
	username, _ := c.Locals(ContextKeyUsername).(string)
	return username
}

// GetUserRole retrieves the user role from the request context.
func GetUserRole(c *fiber.Ctx) domain.Role {
	role, _ := c.Locals(ContextKeyRole).(domain.Role)
	return role
}

// OptionalJWT returns an optional JWT authentication middleware.
func OptionalJWT(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		token := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = authHeader[7:]
		}

		claims, err := authService.ValidateToken(token)
		if err != nil {
			return c.Next()
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.Next()
		}

		c.Locals(ContextKeyUserID, userID)
		c.Locals(ContextKeyUsername, claims.Username)
		c.Locals(ContextKeyRole, claims.Role)

		return c.Next()
	}
}
