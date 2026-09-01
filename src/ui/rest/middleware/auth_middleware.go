package middleware

import (
	"strings"

	domainAuth "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/auth"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

const (
	ContextUserKey   = "saas_user"
	ContextUserIDKey = "saas_user_id"
	ApiKeyHeader     = "X-API-Key"
)

// AuthResolverMiddleware inspects Authorization Bearer token, auth_token cookie, or X-API-Key
// and injects the authenticated *domainAuth.User into context if valid.
func AuthResolverMiddleware(authUsecase domainAuth.IAuthUsecase) fiber.Handler {
	return func(c fiber.Ctx) error {
		var token string

		// 1. Check Authorization: Bearer <token>
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. Check Cookie if header is not present
		if token == "" {
			token = c.Cookies("auth_token")
			if token == "" {
				token = c.Cookies("token")
			}
		}

		// 3. If token present, verify it
		if token != "" {
			if user, err := authUsecase.VerifyToken(token); err == nil && user != nil {
				c.Locals(ContextUserKey, user)
				c.Locals(ContextUserIDKey, user.ID)
				return c.Next()
			}
		}

		// 4. Check X-API-Key header (case-insensitive) or query param
		apiKey := strings.TrimSpace(c.Get(ApiKeyHeader))
		if apiKey == "" {
			apiKey = strings.TrimSpace(c.Get("api-key"))
		}
		if apiKey == "" {
			apiKey = strings.TrimSpace(c.Query("apikey"))
		}
		if apiKey == "" {
			apiKey = strings.TrimSpace(c.Query("api_key"))
		}
		if apiKey == "" {
			apiKey = strings.TrimSpace(c.Query("key"))
		}

		if apiKey != "" {
			if user, err := authUsecase.GetUserByApiKey(c.Context(), apiKey); err == nil && user != nil {
				c.Locals(ContextUserKey, user)
				c.Locals(ContextUserIDKey, user.ID)
				return c.Next()
			}
		}

		return c.Next()
	}
}

// RequireAuth enforces that a valid user is present in context.
func RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := GetUserFromCtx(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ResponseData{
				Status:  fiber.StatusUnauthorized,
				Code:    "UNAUTHORIZED",
				Message: "Authentication required. Please log in.",
			})
		}
		return c.Next()
	}
}

// RequireAdmin enforces that the authenticated user has role 'admin'.
func RequireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := GetUserFromCtx(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(utils.ResponseData{
				Status:  fiber.StatusUnauthorized,
				Code:    "UNAUTHORIZED",
				Message: "Authentication required. Please log in as admin.",
			})
		}
		if user.Role != domainAuth.RoleAdmin {
			return c.Status(fiber.StatusForbidden).JSON(utils.ResponseData{
				Status:  fiber.StatusForbidden,
				Code:    "FORBIDDEN",
				Message: "Admin privileges required to access this resource.",
			})
		}
		return c.Next()
	}
}

// GetUserFromCtx helper retrieves the *domainAuth.User from fiber context.
func GetUserFromCtx(c fiber.Ctx) *domainAuth.User {
	if val := c.Locals(ContextUserKey); val != nil {
		if user, ok := val.(*domainAuth.User); ok {
			return user
		}
	}
	return nil
}

// GetUserIDFromCtx helper retrieves user ID (or 0 if unauthenticated).
func GetUserIDFromCtx(c fiber.Ctx) int64 {
	if user := GetUserFromCtx(c); user != nil {
		return user.ID
	}
	return 0
}
