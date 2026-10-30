package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"maj-core-srvc/src/internal/infra/logger"
	"maj-core-srvc/src/internal/pkg/response"
)

// ErrorHandler adalah centralized error handler untuk Fiber.
// Semua error yang di-return dari handler/middleware akan diformat menggunakan response.Error
// dan dicatat lewat context-aware logger jika merupakan error internal/server (5xx).
func ErrorHandler(c *fiber.Ctx, err error) error {
	ctx := c.UserContext()

	var appErr *response.AppError
	if errors.As(err, &appErr) {
		if appErr.Code.HTTPStatus() >= 500 {
			logger.Error(ctx, "application error",
				"code", appErr.Code,
				"path", c.Path(),
				"method", c.Method(),
				"err", appErr,
			)
		}
		return response.Error(c, appErr)
	}

	var fiberErr *fiber.Error
	if !errors.As(err, &fiberErr) {
		logger.Error(ctx, "unhandled error",
			"path", c.Path(),
			"method", c.Method(),
			"err", err,
		)
	}

	return response.Error(c, err)
}
