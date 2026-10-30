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
	ProductHandler *handler.ProductHandler
}

// RegisterHTTP mendaftarkan seluruh route REST API ke dalam instance Fiber.
func RegisterHTTP(app *fiber.App, deps Dependencies) {
	// Swagger UI documentation (/swagger/index.html)
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Probe endpoints (Kubernetes / Docker healthcheck)
	app.Get("/healthz", deps.HealthHandler.Liveness)
	app.Get("/readyz", deps.HealthHandler.Readiness)

	// Public / Standard API v1 group
	v1 := app.Group("/api/v1")

	// Sample resource: Products
	products := v1.Group("/products")
	products.Get("/", deps.ProductHandler.List)
	products.Post("/", deps.ProductHandler.Create)
	products.Get("/:id", deps.ProductHandler.GetByID)
	products.Put("/:id", deps.ProductHandler.Update)
	products.Delete("/:id", deps.ProductHandler.Delete)

	// Contoh route group yang diproteksi JWT Bearer Token & Internal API Key
	if deps.Config != nil {
		// Group untuk endpoint yang wajib login (Bearer Token)
		protectedV1 := v1.Group("/me", middleware.JWTAuth(deps.Config.Auth.JWTSecret))
		protectedV1.Get("/", func(c *fiber.Ctx) error {
			claims, _ := middleware.GetAuthClaims(c)
			return response.OK(c, "Authenticated user profile", claims)
		})

		// Group untuk komunikasi internal antar-microservice (X-API-Key)
		internalV1 := app.Group("/internal/v1", middleware.APIKeyAuth(deps.Config.Auth.InternalAPIKey))
		internalV1.Get("/products/:id", deps.ProductHandler.GetByID)
	}

	// Fallback 404 Route
	app.Use(func(c *fiber.Ctx) error {
		return response.NewError(response.CodeNotFound, "Route not found")
	})
}
