package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"maj-core-srvc/src/internal/config"
	"maj-core-srvc/src/internal/pkg/response"
)

// RateLimiter membatasi jumlah request per IP dalam rentang waktu tertentu.
// Endpoint probe (/healthz, /readyz) dan dokumentasi (/swagger/*) di-bypass.
func RateLimiter(cfg config.RateLimitConfig) fiber.Handler {
	if !cfg.Enabled {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Expiration,
		Next: func(c *fiber.Ctx) bool {
			path := c.Path()
			return path == "/healthz" || path == "/readyz" || strings.HasPrefix(path, "/swagger")
		},
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return response.Error(c, response.NewError(
				response.CodeTooManyRequests,
				response.CodeTooManyRequests.DefaultMessage(),
			))
		},
	})
}
