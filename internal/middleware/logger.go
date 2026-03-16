package middleware

import (
	"time"

	"telegram-api/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// RequestLogger returns a middleware that logs HTTP requests.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		logger.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", time.Since(start)).
			Str("ip", c.IP()).
			Msg("request")

		return err
	}
}
