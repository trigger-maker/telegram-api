// Package middleware provides HTTP middleware functions.
package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig returns the CORS configuration for the API.
func CORSConfig() cors.Config {
	allowedOrigins := os.Getenv("CORS_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}

	return cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		ExposeHeaders:    "Content-Length,Content-Type",
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// CORS returns a CORS middleware handler.
func CORS() fiber.Handler {
	return cors.New(CORSConfig())
}

// HandlePreflight handles preflight OPTIONS requests.
func HandlePreflight() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "OPTIONS" {
			c.Set("Access-Control-Allow-Origin", "*")
			c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Set("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,X-Requested-With")
			c.Set("Access-Control-Max-Age", "86400")
			return c.SendStatus(204)
		}
		return c.Next()
	}
}
