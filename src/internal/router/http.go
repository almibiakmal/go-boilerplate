package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	"maj-core-srvc/src/internal/config"
	"maj-core-srvc/src/internal/handler"
	"maj-core-srvc/src/internal/middleware"
	"maj-core-srvc/src/internal/pkg/response"
)

// Dependencies menampung konfigurasi dan semua handler yang akan didaftarkan ke router HTTP.
type Dependencies struct {
	Config         *config.Config
	HealthHandler  *handler.HealthHandler
	AuthHandler    *handler.AuthHandler
	PostHandler    *handler.PostHandler
	CommentHandler *handler.CommentHandler
}

// RegisterHTTP mendaftarkan seluruh route REST API untuk Simple Blog System ke dalam instance Fiber.
func RegisterHTTP(app *fiber.App, deps Dependencies) {
	// Swagger UI documentation (/swagger/index.html)
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Probe endpoints (Kubernetes / Docker healthcheck)
	if deps.HealthHandler != nil {
		app.Get("/healthz", deps.HealthHandler.Liveness)
		app.Get("/readyz", deps.HealthHandler.Readiness)
	}

	jwtSecret := "change-me-in-production"
	internalAPIKey := "internal-dev-key"
	if deps.Config != nil {
		jwtSecret = deps.Config.Auth.JWTSecret
		internalAPIKey = deps.Config.Auth.InternalAPIKey
	}
	jwtMiddleware := middleware.JWTAuth(jwtSecret)

	// API v1 group
	v1 := app.Group("/api/v1")

	// 1. Auth Routes
	if deps.AuthHandler != nil {
		auth := v1.Group("/auth")
		auth.Post("/register", deps.AuthHandler.Register)
		auth.Post("/login", deps.AuthHandler.Login)
		auth.Get("/me", jwtMiddleware, deps.AuthHandler.Me)
	}

	// 2. Posts & Post Comments Routes
	if deps.PostHandler != nil {
		posts := v1.Group("/posts")
		posts.Get("/", deps.PostHandler.List)
		posts.Get("/:id", deps.PostHandler.GetByID)
		posts.Post("/", jwtMiddleware, deps.PostHandler.Create)
		posts.Put("/:id", jwtMiddleware, deps.PostHandler.Update)
		posts.Delete("/:id", jwtMiddleware, deps.PostHandler.Delete)

		if deps.CommentHandler != nil {
			posts.Get("/:postId/comments", deps.CommentHandler.ListByPostID)
			posts.Post("/:postId/comments", jwtMiddleware, deps.CommentHandler.Create)
		}
	}

	// 3. Direct Comment Routes (Get, Update & Delete by Comment ID)
	if deps.CommentHandler != nil {
		comments := v1.Group("/comments")
		comments.Get("/:id", deps.CommentHandler.GetByID)
		comments.Put("/:id", jwtMiddleware, deps.CommentHandler.Update)
		comments.Delete("/:id", jwtMiddleware, deps.CommentHandler.Delete)
	}

	// 4. Internal Service-to-Service Routes (X-API-Key)
	if deps.PostHandler != nil {
		internalV1 := app.Group("/internal/v1", middleware.APIKeyAuth(internalAPIKey))
		internalV1.Get("/posts/:id", deps.PostHandler.GetByID)
	}

	// Fallback 404 Route
	app.Use(func(c *fiber.Ctx) error {
		return response.NewError(response.CodeNotFound, "Route not found")
	})
}
