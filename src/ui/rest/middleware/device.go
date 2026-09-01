package middleware

import (
	"net/url"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

const DeviceIDHeader = "X-Device-Id"

// DeviceMiddleware fetches a device instance by header (preferred), path param, or query param
// and injects it into the context. It falls back to the default/only device for single-device mode.
func DeviceMiddleware(dm *whatsapp.DeviceManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Allow non-device-scoped public and management endpoints (e.g., swagger, auth, ui) to pass through.
		path := strings.TrimSpace(c.Path())
		origURL := strings.TrimSpace(c.OriginalURL())
		if isNonDeviceScopedPath(path) || isNonDeviceScopedPath(origURL) {
			return c.Next()
		}

		if dm == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(utils.ResponseData{
				Status:  fiber.StatusServiceUnavailable,
				Code:    "DEVICE_MANAGER_UNAVAILABLE",
				Message: "Device manager is not initialized",
				Results: nil,
			})
		}

		deviceID := strings.TrimSpace(c.Get(DeviceIDHeader))
		// URL-decode the header value to support non-ASCII characters
		if decoded, err := url.QueryUnescape(deviceID); err == nil {
			deviceID = decoded
		}
		if deviceID == "" {
			deviceID = strings.TrimSpace(c.Query("device_id"))
		}

		instance, resolvedID, err := dm.ResolveDevice(deviceID)
		if err != nil {
			// ResolveDevice returns an ID when provided but missing; use it for payload clarity.
			if resolvedID != "" || strings.TrimSpace(deviceID) != "" {
				return c.Status(fiber.StatusNotFound).JSON(utils.ResponseData{
					Status:  fiber.StatusNotFound,
					Code:    "DEVICE_NOT_FOUND",
					Message: "device not found; create a device first from /api/devices or provide a valid X-Device-Id",
					Results: map[string]string{"device_id": resolvedID},
				})
			}

			return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
				Status:  fiber.StatusBadRequest,
				Code:    "DEVICE_ID_REQUIRED",
				Message: "device_id is required via X-Device-Id header or device_id query",
				Results: nil,
			})
		}

		c.Locals("device_id", resolvedID)
		c.Locals("device", instance)
		c.SetContext(whatsapp.ContextWithDevice(c.Context(), instance))
		return c.Next()
	}
}

func isNonDeviceScopedPath(path string) bool {
	p := strings.ToLower(strings.TrimSpace(path))
	if p == "" || p == "/" || p == "/favicon.ico" || strings.Contains(p, "favicon") || p == config.AppBasePath || p == config.AppBasePath+"/" {
		return true
	}
	// Bypass Swagger API Documentation
	if strings.Contains(p, "swag") {
		return true
	}
	// Bypass SaaS Auth, Admin, Plans, Quotas, User endpoints, API Keys & Webhooks
	if strings.Contains(p, "auth") ||
		strings.Contains(p, "admin") ||
		strings.Contains(p, "plan") ||
		strings.Contains(p, "api-key") ||
		strings.Contains(p, "webhook") ||
		strings.Contains(p, "auto-response") ||
		strings.Contains(p, "user/select-plan") ||
		strings.Contains(p, "user/api-keys") ||
		strings.Contains(p, "user/webhooks") {
		return true
	}
	// Bypass Device management endpoints (creating slots, QR pairing, code pairing, deletion)
	if strings.HasPrefix(p, "/devices") || strings.Contains(p, "/devices") {
		return true
	}
	// Bypass Frontend UI Pages
	if p == "/login" || p == "/register" || p == "/dashboard" || strings.HasPrefix(p, "/dashboard") || strings.HasPrefix(p, "/admin") {
		return true
	}
	// Bypass App info & telemetry & WebSocket
	if strings.Contains(p, "/app/info") || strings.Contains(p, "/health") || strings.Contains(p, "/metrics") || strings.HasPrefix(p, "/ws") {
		return true
	}
	return false
}

