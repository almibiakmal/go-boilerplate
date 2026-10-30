package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"maj-core-srvc/src/internal/infra/logger"
)

// RequestLogger menyuntikkan request_id ke UserContext (agar log di service/repository otomatis memiliki request_id)
// dan mencatat setiap HTTP request yang masuk menggunakan structured logger dengan status code yang akurat.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Propagasi request_id dari Fiber Locals ke Go context.Context
		reqID, _ := c.Locals("requestid").(string)
		if reqID != "" {
			c.SetUserContext(logger.WithRequestID(c.UserContext(), reqID))
		}

		err := c.Next()
		if err != nil {
			// Eksekusi ErrorHandler terlebih dahulu agar status code pada c.Response() ter-update
			if ehErr := c.App().ErrorHandler(c, err); ehErr != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		latency := time.Since(start)
		ctx := c.UserContext()
		status := c.Response().StatusCode()

		attrs := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"ip", c.IP(),
			"user_agent", c.Get(fiber.HeaderUserAgent),
		}

		switch {
		case status >= 500:
			logger.Error(ctx, "http request", attrs...)
		case status >= 400:
			logger.Warn(ctx, "http request", attrs...)
		default:
			logger.Info(ctx, "http request", attrs...)
		}

		return nil
	}
}
