package auth

import (
	"context"
	"time"
)

type Role string

const (
	RoleAdmin     Role = "admin"
	RolePelanggan Role = "pelanggan"
)

type PlanID string

const (
	PlanTrial PlanID = "trial"
	PlanOcean PlanID = "ocean"
	PlanSea   PlanID = "sea"
)

type Plan struct {
	ID             PlanID    `json:"id"`
	Name           string    `json:"name"`
	DeviceLimit    int       `json:"device_limit"`    // e.g. 1, 2, 10
	MessageLimit   int       `json:"message_limit"`   // single sending message limit (100 for Trial/Ocean, -1 for Sea)
	BroadcastLimit int       `json:"broadcast_limit"` // broadcast bulk campaign limit (3 for Trial, 5000 for Ocean, -1 for Sea)
	ApiKeyLimit    int       `json:"api_key_limit"`   // e.g. 1, 2, 5
	WebhookLimit   int       `json:"webhook_limit"`   // e.g. 1, 4, 10
	DurationDays   int       `json:"duration_days"`   // 7 for Trial, 30 for Ocean/Sea
	PriceMonthly   int       `json:"price_monthly"`   // IDR: 0 for Trial, 35000 for Ocean, 75000 for Sea
	Description    string    `json:"description"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

var AvailablePlans = map[PlanID]Plan{
	PlanTrial: {
		ID:             PlanTrial,
		Name:           "Trial",
		DeviceLimit:    1,
		MessageLimit:   100,
		BroadcastLimit: 3,
		ApiKeyLimit:    1,
		WebhookLimit:   1,
		DurationDays:   7,
		PriceMonthly:   0,
		Description:    "No description set for this tier.",
		IsActive:       true,
	},
	PlanOcean: {
		ID:             PlanOcean,
		Name:           "Ocean",
		DeviceLimit:    2,
		MessageLimit:   100,  // 100 msgs/day
		BroadcastLimit: 5000, // 5,000 /month
		ApiKeyLimit:    1,
		WebhookLimit:   1,
		DurationDays:   30,
		PriceMonthly:   35000,
		Description:    "Ideal for individuals starting WhatsApp automation (100 messages sending limit & 5,000 broadcasts).",
		IsActive:       true,
	},
	PlanSea: {
		ID:             PlanSea,
		Name:           "Sea",
		DeviceLimit:    10,
		MessageLimit:   -1, // Unlimited sending
		BroadcastLimit: -1, // Unlimited broadcasts
		ApiKeyLimit:    2,
		WebhookLimit:   4,
		DurationDays:   30,
		PriceMonthly:   75000,
		Description:    "Designed for growing businesses with unlimited sending messages, unlimited broadcasts & multi-device power.",
		IsActive:       true,
	},
}

type User struct {
	ID             int64      `json:"id"`
	Email          string     `json:"email"`
	FullName       string     `json:"full_name"`
	PasswordHash   string     `json:"-"`
	Role           Role       `json:"role"`
	TierID         PlanID     `json:"tier_id"`
	TierName       string     `json:"tier_name"`
	TierExpiresAt  *time.Time `json:"tier_expires_at"`
	MessageCount   int        `json:"message_count"`
	BroadcastCount int        `json:"broadcast_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ApiKey struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	KeyValue  string    `json:"key_value"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UserWebhook struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	URL       string    `json:"url"`
	Secret    string    `json:"secret"`
	Events    string    `json:"events"`
	CreatedAt time.Time `json:"created_at"`
}

type AutoResponseTriggerType string

const (
	TriggerExact      AutoResponseTriggerType = "exact"
	TriggerContains   AutoResponseTriggerType = "contains"
	TriggerStartsWith AutoResponseTriggerType = "starts_with"
	TriggerRegex      AutoResponseTriggerType = "regex"
)

type AutoResponse struct {
	ID              int64                   `json:"id"`
	UserID          int64                   `json:"user_id"`
	DeviceID        string                  `json:"device_id"`
	TriggerType     AutoResponseTriggerType `json:"trigger_type"`
	TriggerKeyword  string                  `json:"trigger_keyword"`
	ResponseMessage string                  `json:"response_message"`
	IsActive        bool                    `json:"is_active"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

type CreateAutoResponseRequest struct {
	DeviceID        string                  `json:"device_id"`
	TriggerType     AutoResponseTriggerType `json:"trigger_type"`
	TriggerKeyword  string                  `json:"trigger_keyword"`
	ResponseMessage string                  `json:"response_message"`
	IsActive        *bool                   `json:"is_active,omitempty"`
}

type UpdateAutoResponseRequest struct {
	DeviceID        string                  `json:"device_id"`
	TriggerType     AutoResponseTriggerType `json:"trigger_type"`
	TriggerKeyword  string                  `json:"trigger_keyword"`
	ResponseMessage string                  `json:"response_message"`
	IsActive        *bool                   `json:"is_active,omitempty"`
}

// DTOs
type RegisterRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID             int64      `json:"id"`
	Email          string     `json:"email"`
	FullName       string     `json:"full_name"`
	Role           Role       `json:"role"`
	TierID         PlanID     `json:"tier_id"`
	TierName       string     `json:"tier_name"`
	TierExpiresAt  *time.Time `json:"tier_expires_at"`
	HasActiveTier  bool       `json:"has_active_tier"`
	DaysRemaining  int        `json:"days_remaining"`
	MessageCount   int        `json:"message_count"`
	BroadcastCount int        `json:"broadcast_count"`
	DeviceLimit    int        `json:"device_limit"`
	MessageLimit   int        `json:"message_limit"`
	BroadcastLimit int        `json:"broadcast_limit"`
	ApiKeyLimit    int        `json:"api_key_limit"`
	WebhookLimit   int        `json:"webhook_limit"`
	DeviceCount    int        `json:"device_count"`
	CreatedAt      time.Time  `json:"created_at"`
}

type SelectPlanRequest struct {
	PlanID PlanID `json:"plan_id"`
}

type CreateUserRequest struct {
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	Password     string `json:"password"`
	Role         Role   `json:"role"`
	TierID       PlanID `json:"tier_id"`
	DurationDays int    `json:"duration_days"`
}

type UpdatePlanRequest struct {
	TierID       PlanID `json:"tier_id"`
	DurationDays int    `json:"duration_days"`
}

type RenewPlanRequest struct {
	DurationDays int `json:"duration_days"`
}

type CreateApiKeyRequest struct {
	Name string `json:"name"`
}

type CreateWebhookRequest struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
	Events string `json:"events"`
}

type CreatePlanRequest struct {
	ID             PlanID `json:"id"`
	Name           string `json:"name"`
	DeviceLimit    int    `json:"device_limit"`
	MessageLimit   int    `json:"message_limit"`
	BroadcastLimit int    `json:"broadcast_limit"`
	ApiKeyLimit    int    `json:"api_key_limit"`
	WebhookLimit   int    `json:"webhook_limit"`
	DurationDays   int    `json:"duration_days"`
	PriceMonthly   int    `json:"price_monthly"`
	Description    string `json:"description"`
}

type UpdatePlanConfigRequest struct {
	Name           string `json:"name"`
	DeviceLimit    int    `json:"device_limit"`
	MessageLimit   int    `json:"message_limit"`
	BroadcastLimit int    `json:"broadcast_limit"`
	ApiKeyLimit    int    `json:"api_key_limit"`
	WebhookLimit   int    `json:"webhook_limit"`
	DurationDays   int    `json:"duration_days"`
	PriceMonthly   int    `json:"price_monthly"`
	Description    string `json:"description"`
	IsActive       *bool  `json:"is_active,omitempty"`
}

type IAuthUsecase interface {
	Register(ctx context.Context, req RegisterRequest) (AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (AuthResponse, error)
	GetMe(ctx context.Context, userID int64) (UserResponse, error)
	SelectPlan(ctx context.Context, userID int64, req SelectPlanRequest) (UserResponse, error)
	GetPlans(ctx context.Context) ([]Plan, error)

	// API Keys & Webhooks
	GetApiKeys(ctx context.Context, userID int64) ([]ApiKey, error)
	CreateApiKey(ctx context.Context, userID int64, req CreateApiKeyRequest) (ApiKey, error)
	DeleteApiKey(ctx context.Context, userID int64, keyID int64) error
	GetWebhooks(ctx context.Context, userID int64) ([]UserWebhook, error)
	CreateWebhook(ctx context.Context, userID int64, req CreateWebhookRequest) (UserWebhook, error)
	DeleteWebhook(ctx context.Context, userID int64, webhookID int64) error

	// Auto Response operations
	GetUserAutoResponses(ctx context.Context, userID int64) ([]AutoResponse, error)
	CreateAutoResponse(ctx context.Context, userID int64, req CreateAutoResponseRequest) (AutoResponse, error)
	UpdateAutoResponse(ctx context.Context, userID int64, id int64, req UpdateAutoResponseRequest) (AutoResponse, error)
	DeleteAutoResponse(ctx context.Context, userID int64, id int64) error

	// Admin User operations
	AdminListUsers(ctx context.Context) ([]UserResponse, error)
	AdminCreateUser(ctx context.Context, req CreateUserRequest) (UserResponse, error)
	AdminUpdateUserPlan(ctx context.Context, targetUserID int64, req UpdatePlanRequest) (UserResponse, error)
	AdminRenewUserPlan(ctx context.Context, targetUserID int64, req RenewPlanRequest) (UserResponse, error)
	AdminDeleteUser(ctx context.Context, targetUserID int64) error

	// Admin Plan operations
	AdminListPlans(ctx context.Context) ([]Plan, error)
	AdminCreatePlan(ctx context.Context, req CreatePlanRequest) (Plan, error)
	AdminUpdatePlanConfig(ctx context.Context, planID PlanID, req UpdatePlanConfigRequest) (Plan, error)
	AdminDeletePlan(ctx context.Context, planID PlanID) error

	// Quota & Authorization helpers
	ValidateUserDeviceQuota(ctx context.Context, userID int64) error
	ValidateUserSendQuota(ctx context.Context, userID int64, messageCount int) error
	ValidateUserBroadcastQuota(ctx context.Context, userID int64, messageCount int) error
	IncrementUserMessageCount(ctx context.Context, userID int64, count int) error
	IncrementUserBroadcastCount(ctx context.Context, userID int64, count int) error
	CheckAndPerformDailyReset(ctx context.Context) error
	CheckAndPerformMonthlyReset(ctx context.Context) error
	AdminManualResetQuotas(ctx context.Context) (int64, error)
	GetUserByApiKey(ctx context.Context, apiKey string) (*User, error)
	VerifyToken(tokenString string) (*User, error)
}
