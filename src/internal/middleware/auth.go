package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"maj-core-srvc/src/internal/infra/logger"
	"maj-core-srvc/src/internal/pkg/response"
)

const AuthClaimsLocalsKey = "auth_claims"

// AuthClaims merepresentasikan payload standar di dalam JWT token.
type AuthClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email,omitempty"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// JWTAuth memvalidasi Bearer JWT token pada header Authorization.
func JWTAuth(secret string) fiber.Handler {
	// Precompute byte slice dan valid signing methods sekali saat inisialisasi middleware
	secretBytes := []byte(secret)
	validMethods := jwt.WithValidMethods([]string{
		jwt.SigningMethodHS256.Name,
		jwt.SigningMethodHS384.Name,
		jwt.SigningMethodHS512.Name,
	})
	keyFunc := func(t *jwt.Token) (any, error) {
		return secretBytes, nil
	}

	return func(c *fiber.Ctx) error {
		authHeader := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
		if authHeader == "" {
			return response.NewError(response.CodeUnauthorized, "Missing Authorization header")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.NewError(response.CodeUnauthorized, "Invalid Authorization header format (expected Bearer <token>)")
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			return response.NewError(response.CodeUnauthorized, "Invalid Authorization header format (expected Bearer <token>)")
		}

		claims := &AuthClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, keyFunc, validMethods)
		if err != nil || !token.Valid {
			return response.NewError(response.CodeUnauthorized, "Invalid or expired token")
		}

		c.Locals(AuthClaimsLocalsKey, claims)
		if claims.UserID != "" {
			c.SetUserContext(logger.WithUserID(c.UserContext(), claims.UserID))
		}
		return c.Next()
	}
}

// RequireRoles memastikan user yang sudah terautentikasi memiliki salah satu role yang diizinkan.
func RequireRoles(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := GetAuthClaims(c)
		if !ok {
			return response.NewError(response.CodeUnauthorized, "Unauthorized access")
		}

		for _, role := range allowedRoles {
			if strings.EqualFold(claims.Role, role) {
				return c.Next()
			}
		}

		return response.NewError(response.CodeForbidden, "Insufficient role permissions")
	}
}

// GetAuthClaims mengambil AuthClaims dari Fiber context yang sudah di-set oleh middleware JWTAuth.
func GetAuthClaims(c *fiber.Ctx) (*AuthClaims, bool) {
	val := c.Locals(AuthClaimsLocalsKey)
	if val == nil {
		return nil, false
	}
	claims, ok := val.(*AuthClaims)
	return claims, ok
}
