package rest

import (
	"fmt"
	"strconv"
	"time"

	domainAuth "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/auth"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest/middleware"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	Service domainAuth.IAuthUsecase
}

func InitRestAuth(app fiber.Router, service domainAuth.IAuthUsecase) AuthHandler {
	handler := AuthHandler{Service: service}

	// Public Auth & Plan routes
	app.Post("/api/auth/register", handler.Register)
	app.Post("/api/auth/login", handler.Login)
	app.Post("/api/auth/logout", handler.Logout)
	app.Get("/api/plans", handler.GetPlans)

	// Authenticated User routes
	app.Get("/api/auth/me", middleware.RequireAuth(), handler.GetMe)

	userGroup := app.Group("/api/user", middleware.RequireAuth())
	userGroup.Post("/select-plan", handler.SelectPlan)
	userGroup.Get("/api-keys", handler.GetApiKeys)
	userGroup.Post("/api-keys", handler.CreateApiKey)
	userGroup.Delete("/api-keys/:id", handler.DeleteApiKey)
	userGroup.Get("/webhooks", handler.GetWebhooks)
	userGroup.Post("/webhooks", handler.CreateWebhook)
	userGroup.Delete("/webhooks/:id", handler.DeleteWebhook)
	userGroup.Get("/auto-responses", handler.GetAutoResponses)
	userGroup.Post("/auto-responses", handler.CreateAutoResponse)
	userGroup.Put("/auto-responses/:id", handler.UpdateAutoResponse)
	userGroup.Delete("/auto-responses/:id", handler.DeleteAutoResponse)

	// Admin User routes
	adminGroup := app.Group("/api/admin/users", middleware.RequireAdmin())
	adminGroup.Get("", handler.AdminListUsers)
	adminGroup.Post("", handler.AdminCreateUser)
	adminGroup.Put("/:id/plan", handler.AdminUpdateUserPlan)
	adminGroup.Post("/:id/renew", handler.AdminRenewUserPlan)
	adminGroup.Delete("/:id", handler.AdminDeleteUser)

	// Admin Plan routes
	adminPlanGroup := app.Group("/api/admin/plans", middleware.RequireAdmin())
	adminPlanGroup.Get("", handler.AdminListPlans)
	adminPlanGroup.Post("", handler.AdminCreatePlan)
	adminPlanGroup.Put("/:id", handler.AdminUpdatePlanConfig)
	adminPlanGroup.Delete("/:id", handler.AdminDeletePlan)

	// Admin Quota Reset route
	app.Post("/api/admin/reset-quotas", middleware.RequireAdmin(), handler.AdminResetQuotas)

	return handler
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req domainAuth.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid JSON payload: " + err.Error(),
		})
	}

	resp, err := h.Service.Register(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "REGISTER_FAILED",
			Message: err.Error(),
		})
	}

	// Set auth cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    resp.Token,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: false,
		SameSite: "Lax",
	})

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Registration successful",
		Results: resp,
	})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req domainAuth.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid JSON payload: " + err.Error(),
		})
	}

	resp, err := h.Service.Login(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ResponseData{
			Status:  fiber.StatusUnauthorized,
			Code:    "LOGIN_FAILED",
			Message: err.Error(),
		})
	}

	// Set auth cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    resp.Token,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: false,
		SameSite: "Lax",
	})

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Login successful",
		Results: resp,
	})
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	cookiesToClear := []string{"auth_token", "gowa_auth_token", "token", "session"}
	for _, name := range cookiesToClear {
		c.Cookie(&fiber.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			Expires:  time.Now().Add(-24 * time.Hour),
			MaxAge:   -1,
			HTTPOnly: false,
			SameSite: "Lax",
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Logged out successfully",
	})
}

func (h *AuthHandler) GetMe(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ResponseData{
			Status:  fiber.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "Unauthorized",
		})
	}

	resp, err := h.Service.GetMe(c.Context(), user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ResponseData{
			Status:  fiber.StatusInternalServerError,
			Code:    "SERVER_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "User profile",
		Results: resp,
	})
}

func (h *AuthHandler) GetPlans(c fiber.Ctx) error {
	plans, err := h.Service.GetPlans(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ResponseData{
			Status:  fiber.StatusInternalServerError,
			Code:    "GET_PLANS_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Available plans",
		Results: plans,
	})
}

func (h *AuthHandler) SelectPlan(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ResponseData{
			Status:  fiber.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "Unauthorized",
		})
	}

	var req domainAuth.SelectPlanRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid plan payload: " + err.Error(),
		})
	}

	resp, err := h.Service.SelectPlan(c.Context(), user.ID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "PLAN_SELECTION_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Plan activated successfully",
		Results: resp,
	})
}

func (h *AuthHandler) GetApiKeys(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	keys, err := h.Service.GetApiKeys(c.Context(), user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ResponseData{
			Status:  500,
			Code:    "SERVER_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "API Keys",
		Results: keys,
	})
}

func (h *AuthHandler) CreateApiKey(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	var req domainAuth.CreateApiKeyRequest
	_ = c.Bind().Body(&req)

	key, err := h.Service.CreateApiKey(c.Context(), user.ID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "CREATE_KEY_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "API Key created",
		Results: key,
	})
}

func (h *AuthHandler) DeleteApiKey(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := h.Service.DeleteApiKey(c.Context(), user.ID, id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "DELETE_KEY_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "API Key deleted",
	})
}

func (h *AuthHandler) GetWebhooks(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	hooks, err := h.Service.GetWebhooks(c.Context(), user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ResponseData{
			Status:  500,
			Code:    "SERVER_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Webhooks",
		Results: hooks,
	})
}

func (h *AuthHandler) CreateWebhook(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	var req domainAuth.CreateWebhookRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: err.Error(),
		})
	}

	hook, err := h.Service.CreateWebhook(c.Context(), user.ID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "CREATE_WEBHOOK_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Webhook created",
		Results: hook,
	})
}

func (h *AuthHandler) DeleteWebhook(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := h.Service.DeleteWebhook(c.Context(), user.ID, id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "DELETE_WEBHOOK_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Webhook deleted",
	})
}

func (h *AuthHandler) AdminListUsers(c fiber.Ctx) error {
	users, err := h.Service.AdminListUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ResponseData{
			Status:  500,
			Code:    "SERVER_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "All users",
		Results: users,
	})
}

func (h *AuthHandler) AdminCreateUser(c fiber.Ctx) error {
	var req domainAuth.CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: err.Error(),
		})
	}

	resp, err := h.Service.AdminCreateUser(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "CREATE_USER_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "User created successfully",
		Results: resp,
	})
}

func (h *AuthHandler) AdminUpdateUserPlan(c fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	var req domainAuth.UpdatePlanRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: err.Error(),
		})
	}

	resp, err := h.Service.AdminUpdateUserPlan(c.Context(), id, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "UPDATE_PLAN_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Plan updated successfully",
		Results: resp,
	})
}

func (h *AuthHandler) AdminRenewUserPlan(c fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	var req domainAuth.RenewPlanRequest
	_ = c.Bind().Body(&req)

	resp, err := h.Service.AdminRenewUserPlan(c.Context(), id, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "RENEW_PLAN_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Plan renewed successfully",
		Results: resp,
	})
}

func (h *AuthHandler) AdminDeleteUser(c fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := h.Service.AdminDeleteUser(c.Context(), id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "DELETE_USER_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "User deleted",
	})
}

func (h *AuthHandler) AdminListPlans(c fiber.Ctx) error {
	plans, err := h.Service.AdminListPlans(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ResponseData{
			Status:  fiber.StatusInternalServerError,
			Code:    "GET_PLANS_FAILED",
			Message: err.Error(),
		})
	}
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "All plans",
		Results: plans,
	})
}

func (h *AuthHandler) AdminCreatePlan(c fiber.Ctx) error {
	var req domainAuth.CreatePlanRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid plan payload: " + err.Error(),
		})
	}

	plan, err := h.Service.AdminCreatePlan(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "CREATE_PLAN_FAILED",
			Message: err.Error(),
		})
	}
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Plan created successfully",
		Results: plan,
	})
}

func (h *AuthHandler) AdminUpdatePlanConfig(c fiber.Ctx) error {
	planID := domainAuth.PlanID(c.Params("id"))
	var req domainAuth.UpdatePlanConfigRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid plan update payload: " + err.Error(),
		})
	}

	plan, err := h.Service.AdminUpdatePlanConfig(c.Context(), planID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "UPDATE_PLAN_FAILED",
			Message: err.Error(),
		})
	}
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Plan updated successfully",
		Results: plan,
	})
}

func (h *AuthHandler) AdminDeletePlan(c fiber.Ctx) error {
	planID := domainAuth.PlanID(c.Params("id"))
	if err := h.Service.AdminDeletePlan(c.Context(), planID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "DELETE_PLAN_FAILED",
			Message: err.Error(),
		})
	}
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Plan deleted successfully",
	})
}

// _____________________________________________________________________________________________________________________
// Auto Response Handlers

func (h *AuthHandler) GetAutoResponses(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ResponseData{
			Status:  fiber.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}

	responses, err := h.Service.GetUserAutoResponses(c.Context(), user.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "FETCH_AUTO_RESPONSES_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Auto responses retrieved successfully",
		Results: responses,
	})
}

func (h *AuthHandler) CreateAutoResponse(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ResponseData{
			Status:  fiber.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}

	var req domainAuth.CreateAutoResponseRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid request payload: " + err.Error(),
		})
	}

	ar, err := h.Service.CreateAutoResponse(c.Context(), user.ID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "CREATE_AUTO_RESPONSE_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Auto response rule created successfully",
		Results: ar,
	})
}

func (h *AuthHandler) UpdateAutoResponse(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ResponseData{
			Status:  fiber.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}

	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "INVALID_ID",
			Message: "Valid rule id is required",
		})
	}

	var req domainAuth.UpdateAutoResponseRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "BAD_REQUEST",
			Message: "Invalid request payload: " + err.Error(),
		})
	}

	ar, err := h.Service.UpdateAutoResponse(c.Context(), user.ID, id, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "UPDATE_AUTO_RESPONSE_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Auto response rule updated successfully",
		Results: ar,
	})
}

func (h *AuthHandler) DeleteAutoResponse(c fiber.Ctx) error {
	user := middleware.GetUserFromCtx(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(utils.ResponseData{
			Status:  fiber.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}

	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "INVALID_ID",
			Message: "Valid rule id is required",
		})
	}

	if err := h.Service.DeleteAutoResponse(c.Context(), user.ID, id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
			Status:  fiber.StatusBadRequest,
			Code:    "DELETE_AUTO_RESPONSE_FAILED",
			Message: err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Auto response rule deleted successfully",
	})
}

func (h *AuthHandler) AdminResetQuotas(c fiber.Ctx) error {
	affected, err := h.Service.AdminManualResetQuotas(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(utils.ResponseData{
			Status:  fiber.StatusInternalServerError,
			Code:    "RESET_FAILED",
			Message: "Failed to reset quotas: " + err.Error(),
		})
	}

	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: fmt.Sprintf("Successfully reset sending and broadcast quotas for all users (%d accounts updated)", affected),
		Results: fiber.Map{
			"affected_users": affected,
		},
	})
}


