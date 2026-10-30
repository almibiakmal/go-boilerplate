package middleware

import (
	"crypto/subtle"
	"strings"

	"github.com/gofiber/fiber/v2"

	"maj-core-srvc/src/internal/pkg/response"
)

const HeaderAPIKey = "X-API-Key"

// APIKeyAuth memvalidasi header X-API-Key untuk komunikasi internal antar-microservice (Service-to-Service).
func APIKeyAuth(expectedKey string) fiber.Handler {
	expectedBytes := []byte(strings.TrimSpace(expectedKey))

	return func(c *fiber.Ctx) error {
		if len(expectedBytes) == 0 {
			return response.NewError(response.CodeInternalError, "Internal API key is not configured")
		}

		providedKey := strings.TrimSpace(c.Get(HeaderAPIKey))
		if providedKey == "" {
			return response.NewError(response.CodeUnauthorized, "Missing X-API-Key header")
		}

		if subtle.ConstantTimeCompare([]byte(providedKey), expectedBytes) != 1 {
			return response.NewError(response.CodeUnauthorized, "Invalid X-API-Key")
		}

		return c.Next()
	}
}
