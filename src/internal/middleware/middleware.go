package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"maj-core-srvc/src/internal/config"
)

// Setup mendaftarkan global middleware standar pada instance Fiber.
func Setup(app *fiber.App, cfg *config.Config) {
	app.Use(requestid.New())
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	app.Use(cors.New())
	app.Use(RequestLogger())

	if cfg != nil {
		app.Use(RateLimiter(cfg.RateLimit))
	}
}
