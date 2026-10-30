package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maj-core-srvc/src/internal/config"
	"maj-core-srvc/src/internal/middleware"
	"maj-core-srvc/src/internal/pkg/response"
)

const testJWTSecret = "super-secret-test-key"

func generateTestToken(t *testing.T, secret, userID, role string, exp time.Time) string {
	t.Helper()
	claims := middleware.AuthClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}

func TestJWTAuth_And_RequireRoles(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Get("/protected", middleware.JWTAuth(testJWTSecret), func(c *fiber.Ctx) error {
		claims, ok := middleware.GetAuthClaims(c)
		require.True(t, ok)
		return response.OK(c, "ok", fiber.Map{"user_id": claims.UserID})
	})

	app.Get("/admin", middleware.JWTAuth(testJWTSecret), middleware.RequireRoles("admin"), func(c *fiber.Ctx) error {
		return response.OK(c, "admin area", nil)
	})

	t.Run("missing authorization header returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("valid token returns 200", func(t *testing.T) {
		token := generateTestToken(t, testJWTSecret, "usr-123", "cashier", time.Now().Add(time.Hour))
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("insufficient role returns 403", func(t *testing.T) {
		token := generateTestToken(t, testJWTSecret, "usr-123", "cashier", time.Now().Add(time.Hour))
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("matching role returns 200", func(t *testing.T) {
		token := generateTestToken(t, testJWTSecret, "usr-999", "admin", time.Now().Add(time.Hour))
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestAPIKeyAuth(t *testing.T) {
	const expectedKey = "internal-secret-key"

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})
	app.Get("/internal/ping", middleware.APIKeyAuth(expectedKey), func(c *fiber.Ctx) error {
		return response.OK(c, "pong", nil)
	})

	t.Run("missing API key returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/internal/ping", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("wrong API key returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/internal/ping", nil)
		req.Header.Set(middleware.HeaderAPIKey, "wrong-key")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("valid API key returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/internal/ping", nil)
		req.Header.Set(middleware.HeaderAPIKey, expectedKey)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestRateLimiter(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})
	app.Use(middleware.RateLimiter(config.RateLimitConfig{
		Enabled:    true,
		Max:        2,
		Expiration: 10 * time.Second,
	}))
	app.Get("/test", func(c *fiber.Ctx) error {
		return response.OK(c, "ok", nil)
	})

	// Request 1 & 2 harus lolos (200)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Request 3 harus terkena rate limit (429 TOO_MANY_REQUESTS)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	var env response.Envelope
	err = json.NewDecoder(resp.Body).Decode(&env)
	require.NoError(t, err)
	assert.False(t, env.Success)
	assert.Equal(t, response.CodeTooManyRequests, env.Code)
}

