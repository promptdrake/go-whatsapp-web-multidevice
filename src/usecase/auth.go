package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	domainAuth "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/auth"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecretKey = []byte("gowa-saas-secure-jwt-signing-secret-2026")
)

type jwtClaims struct {
	UserID int64           `json:"uid"`
	Email  string          `json:"email"`
	Role   domainAuth.Role `json:"role"`
	TierID domainAuth.PlanID `json:"tier"`
	Exp    int64           `json:"exp"`
}

type authService struct {
	repo domainChatStorage.IChatStorageRepository
}

func NewAuthService(repo domainChatStorage.IChatStorageRepository) domainAuth.IAuthUsecase {
	return &authService{repo: repo}
}

func (s *authService) generateToken(user *domainAuth.User) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := jwtClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		TierID: user.TierID,
		Exp:    time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(claimsBytes)

	data := header + "." + payload
	h := hmac.New(sha256.New, jwtSecretKey)
	h.Write([]byte(data))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return data + "." + signature, nil
}

func (s *authService) VerifyToken(tokenString string) (*domainAuth.User, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, errors.New("missing token")
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	data := parts[0] + "." + parts[1]
	h := hmac.New(sha256.New, jwtSecretKey)
	h.Write([]byte(data))
	expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if hmac.Equal([]byte(parts[2]), []byte(expectedSig)) == false {
		return nil, errors.New("invalid token signature")
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid token payload")
	}

	var claims jwtClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, errors.New("invalid token claims")
	}

	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}

	user, err := s.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *authService) Register(ctx context.Context, req domainAuth.RegisterRequest) (domainAuth.AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	fullName := strings.TrimSpace(req.FullName)
	password := req.Password

	if email == "" || fullName == "" || password == "" {
		return domainAuth.AuthResponse{}, pkgError.ValidationError("Email, Full Name, and Password are required")
	}
	if len(password) < 6 {
		return domainAuth.AuthResponse{}, pkgError.ValidationError("Password must be at least 6 characters")
	}

	existing, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return domainAuth.AuthResponse{}, err
	}
	if existing != nil {
		return domainAuth.AuthResponse{}, pkgError.ValidationError("Email is already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domainAuth.AuthResponse{}, err
	}

	user := &domainAuth.User{
		Email:        email,
		FullName:     fullName,
		PasswordHash: string(hash),
		Role:         domainAuth.RolePelanggan,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return domainAuth.AuthResponse{}, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return domainAuth.AuthResponse{}, err
	}

	userResp, err := s.GetMe(ctx, user.ID)
	if err != nil {
		return domainAuth.AuthResponse{}, err
	}

	return domainAuth.AuthResponse{
		Token: token,
		User:  userResp,
	}, nil
}

func (s *authService) Login(ctx context.Context, req domainAuth.LoginRequest) (domainAuth.AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	password := req.Password

	if email == "" || password == "" {
		return domainAuth.AuthResponse{}, pkgError.ValidationError("Email and Password are required")
	}

	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return domainAuth.AuthResponse{}, err
	}
	if user == nil {
		return domainAuth.AuthResponse{}, pkgError.ValidationError("Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domainAuth.AuthResponse{}, pkgError.ValidationError("Invalid email or password")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return domainAuth.AuthResponse{}, err
	}

	userResp, err := s.GetMe(ctx, user.ID)
	if err != nil {
		return domainAuth.AuthResponse{}, err
	}

	return domainAuth.AuthResponse{
		Token: token,
		User:  userResp,
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID int64) (domainAuth.UserResponse, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil || user == nil {
		return domainAuth.UserResponse{}, errors.New("user not found")
	}

	deviceCount, _ := s.repo.CountDevicesByUserID(userID)

	hasActiveTier := false
	daysRemaining := 0
	if user.Role == domainAuth.RoleAdmin {
		hasActiveTier = true
		daysRemaining = 3650 // 10 years
		if user.TierID == "" {
			user.TierID = domainAuth.PlanSea
			user.TierName = "Sea"
		}
	} else if user.TierID != "" && user.TierExpiresAt != nil {
		if user.TierExpiresAt.After(time.Now()) {
			hasActiveTier = true
			daysRemaining = int(time.Until(*user.TierExpiresAt).Hours() / 24)
		}
	}

	deviceLimit := 0
	messageLimit := 0
	broadcastLimit := 0
	apiKeyLimit := 0
	webhookLimit := 0

	if user.Role == domainAuth.RoleAdmin {
		deviceLimit = 999
		messageLimit = -1
		broadcastLimit = -1
		apiKeyLimit = 999
		webhookLimit = 999
	} else if user.TierID != "" && hasActiveTier {
		plan, _ := s.repo.GetPlanByID(user.TierID)
		if plan != nil {
			deviceLimit = plan.DeviceLimit
			messageLimit = plan.MessageLimit
			broadcastLimit = plan.BroadcastLimit
			apiKeyLimit = plan.ApiKeyLimit
			webhookLimit = plan.WebhookLimit
		} else if defPlan, ok := domainAuth.AvailablePlans[user.TierID]; ok {
			deviceLimit = defPlan.DeviceLimit
			messageLimit = defPlan.MessageLimit
			broadcastLimit = defPlan.BroadcastLimit
			apiKeyLimit = defPlan.ApiKeyLimit
			webhookLimit = defPlan.WebhookLimit
		}
	}

	return domainAuth.UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		FullName:       user.FullName,
		Role:           user.Role,
		TierID:         user.TierID,
		TierName:       user.TierName,
		TierExpiresAt:  user.TierExpiresAt,
		HasActiveTier:  hasActiveTier,
		DaysRemaining:  daysRemaining,
		MessageCount:   user.MessageCount,
		BroadcastCount: user.BroadcastCount,
		DeviceLimit:    deviceLimit,
		MessageLimit:   messageLimit,
		BroadcastLimit: broadcastLimit,
		ApiKeyLimit:    apiKeyLimit,
		WebhookLimit:   webhookLimit,
		DeviceCount:    deviceCount,
		CreatedAt:      user.CreatedAt,
	}, nil
}

func (s *authService) SelectPlan(ctx context.Context, userID int64, req domainAuth.SelectPlanRequest) (domainAuth.UserResponse, error) {
	plan, err := s.repo.GetPlanByID(req.PlanID)
	if err != nil || plan == nil {
		if defPlan, ok := domainAuth.AvailablePlans[req.PlanID]; ok {
			plan = &defPlan
		} else {
			return domainAuth.UserResponse{}, pkgError.ValidationError(fmt.Sprintf("Invalid plan selected: %s", req.PlanID))
		}
	}

	duration := plan.DurationDays
	if duration <= 0 {
		duration = 30
	}
	expiresAt := time.Now().AddDate(0, 0, duration)
	if err := s.repo.UpdateUserTier(userID, plan.ID, plan.Name, &expiresAt); err != nil {
		return domainAuth.UserResponse{}, err
	}

	return s.GetMe(ctx, userID)
}

func (s *authService) GetApiKeys(ctx context.Context, userID int64) ([]domainAuth.ApiKey, error) {
	keys, err := s.repo.GetApiKeysByUserID(userID)
	if err != nil {
		return nil, err
	}
	var res []domainAuth.ApiKey
	for _, k := range keys {
		if k != nil {
			res = append(res, *k)
		}
	}
	return res, nil
}

func (s *authService) CreateApiKey(ctx context.Context, userID int64, req domainAuth.CreateApiKeyRequest) (domainAuth.ApiKey, error) {
	user, err := s.GetMe(ctx, userID)
	if err != nil {
		return domainAuth.ApiKey{}, err
	}
	if !user.HasActiveTier {
		return domainAuth.ApiKey{}, pkgError.ValidationError("An active subscription plan is required to generate API keys.")
	}

	keys, err := s.repo.GetApiKeysByUserID(userID)
	if err != nil {
		return domainAuth.ApiKey{}, err
	}
	if user.Role != domainAuth.RoleAdmin && len(keys) >= user.ApiKeyLimit {
		return domainAuth.ApiKey{}, fmt.Errorf("API key limit reached for your plan (%d max). Upgrade your plan to add more.", user.ApiKeyLimit)
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = fmt.Sprintf("API Key %d", len(keys)+1)
	}

	b := make([]byte, 24)
	_, _ = rand.Read(b)
	keyValue := "gowa_" + hex.EncodeToString(b)

	apiKey := &domainAuth.ApiKey{
		UserID:   userID,
		KeyValue: keyValue,
		Name:     name,
	}

	if err := s.repo.CreateApiKey(apiKey); err != nil {
		return domainAuth.ApiKey{}, err
	}

	return *apiKey, nil
}

func (s *authService) DeleteApiKey(ctx context.Context, userID int64, keyID int64) error {
	return s.repo.DeleteApiKey(keyID, userID)
}

func (s *authService) GetWebhooks(ctx context.Context, userID int64) ([]domainAuth.UserWebhook, error) {
	hooks, err := s.repo.GetUserWebhooksByUserID(userID)
	if err != nil {
		return nil, err
	}
	var res []domainAuth.UserWebhook
	for _, h := range hooks {
		if h != nil {
			res = append(res, *h)
		}
	}
	return res, nil
}

func (s *authService) CreateWebhook(ctx context.Context, userID int64, req domainAuth.CreateWebhookRequest) (domainAuth.UserWebhook, error) {
	user, err := s.GetMe(ctx, userID)
	if err != nil {
		return domainAuth.UserWebhook{}, err
	}
	if !user.HasActiveTier {
		return domainAuth.UserWebhook{}, pkgError.ValidationError("An active subscription plan is required to configure Webhooks.")
	}

	hooks, err := s.repo.GetUserWebhooksByUserID(userID)
	if err != nil {
		return domainAuth.UserWebhook{}, err
	}
	if user.Role != domainAuth.RoleAdmin && len(hooks) >= user.WebhookLimit {
		return domainAuth.UserWebhook{}, fmt.Errorf("Webhook limit reached for your plan (%d max). Upgrade to Sea plan for up to 4 webhooks.", user.WebhookLimit)
	}

	url := strings.TrimSpace(req.URL)
	if url == "" {
		return domainAuth.UserWebhook{}, pkgError.ValidationError("Webhook URL is required")
	}

	webhook := &domainAuth.UserWebhook{
		UserID: userID,
		URL:    url,
		Secret: req.Secret,
		Events: req.Events,
	}

	if err := s.repo.CreateUserWebhook(webhook); err != nil {
		return domainAuth.UserWebhook{}, err
	}

	return *webhook, nil
}

func (s *authService) DeleteWebhook(ctx context.Context, userID int64, webhookID int64) error {
	return s.repo.DeleteUserWebhook(webhookID, userID)
}

func (s *authService) AdminListUsers(ctx context.Context) ([]domainAuth.UserResponse, error) {
	users, err := s.repo.ListUsers()
	if err != nil {
		return nil, err
	}

	var res []domainAuth.UserResponse
	for _, u := range users {
		if u == nil {
			continue
		}
		resp, _ := s.GetMe(ctx, u.ID)
		res = append(res, resp)
	}
	return res, nil
}

func (s *authService) AdminCreateUser(ctx context.Context, req domainAuth.CreateUserRequest) (domainAuth.UserResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	fullName := strings.TrimSpace(req.FullName)
	if email == "" || fullName == "" || req.Password == "" {
		return domainAuth.UserResponse{}, pkgError.ValidationError("Email, Full Name, and Password are required")
	}

	existing, _ := s.repo.GetUserByEmail(email)
	if existing != nil {
		return domainAuth.UserResponse{}, pkgError.ValidationError("Email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return domainAuth.UserResponse{}, err
	}

	role := req.Role
	if role == "" {
		role = domainAuth.RolePelanggan
	}

	var expiresAt *time.Time
	tierName := ""
	tierID := req.TierID

	if tierID != "" {
		if plan, ok := domainAuth.AvailablePlans[tierID]; ok {
			tierName = plan.Name
			days := req.DurationDays
			if days <= 0 {
				days = plan.DurationDays
			}
			t := time.Now().AddDate(0, 0, days)
			expiresAt = &t
		}
	}

	user := &domainAuth.User{
		Email:         email,
		FullName:      fullName,
		PasswordHash:  string(hash),
		Role:          role,
		TierID:        tierID,
		TierName:      tierName,
		TierExpiresAt: expiresAt,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return domainAuth.UserResponse{}, err
	}

	return s.GetMe(ctx, user.ID)
}

func (s *authService) AdminUpdateUserPlan(ctx context.Context, targetUserID int64, req domainAuth.UpdatePlanRequest) (domainAuth.UserResponse, error) {
	plan, ok := domainAuth.AvailablePlans[req.TierID]
	if !ok {
		return domainAuth.UserResponse{}, pkgError.ValidationError("Invalid plan selected")
	}

	days := req.DurationDays
	if days <= 0 {
		days = plan.DurationDays
	}

	expiresAt := time.Now().AddDate(0, 0, days)
	if err := s.repo.UpdateUserTier(targetUserID, plan.ID, plan.Name, &expiresAt); err != nil {
		return domainAuth.UserResponse{}, err
	}

	return s.GetMe(ctx, targetUserID)
}

func (s *authService) AdminRenewUserPlan(ctx context.Context, targetUserID int64, req domainAuth.RenewPlanRequest) (domainAuth.UserResponse, error) {
	user, err := s.repo.GetUserByID(targetUserID)
	if err != nil || user == nil {
		return domainAuth.UserResponse{}, errors.New("user not found")
	}

	days := req.DurationDays
	if days <= 0 {
		days = 30
	}

	baseTime := time.Now()
	if user.TierExpiresAt != nil && user.TierExpiresAt.After(time.Now()) {
		baseTime = *user.TierExpiresAt
	}

	newExpiry := baseTime.AddDate(0, 0, days)
	planID := user.TierID
	tierName := user.TierName
	if planID == "" {
		planID = domainAuth.PlanOcean
		tierName = "Ocean"
	}

	if err := s.repo.UpdateUserTier(targetUserID, planID, tierName, &newExpiry); err != nil {
		return domainAuth.UserResponse{}, err
	}

	// Reset broadcast count on renewal
	user.BroadcastCount = 0
	_ = s.repo.UpdateUser(user)

	return s.GetMe(ctx, targetUserID)
}

func (s *authService) AdminDeleteUser(ctx context.Context, targetUserID int64) error {
	return s.repo.DeleteUser(targetUserID)
}

func (s *authService) ValidateUserDeviceQuota(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return nil // unauthenticated fallback if auth is disabled
	}

	user, err := s.GetMe(ctx, userID)
	if err != nil {
		return err
	}
	if user.Role == domainAuth.RoleAdmin {
		return nil
	}

	if !user.HasActiveTier {
		return pkgError.ValidationError("No active subscription tier found. Please select a plan (Ocean or Sea) to add WhatsApp devices.")
	}

	if user.DeviceCount >= user.DeviceLimit {
		return fmt.Errorf("Device limit reached for %s tier (%d max). Please upgrade to Sea plan for up to 3 devices.", user.TierName, user.DeviceLimit)
	}

	return nil
}

func (s *authService) ValidateUserSendQuota(ctx context.Context, userID int64, messageCount int) error {
	if userID <= 0 {
		return nil
	}

	user, err := s.GetMe(ctx, userID)
	if err != nil {
		return err
	}
	if user.Role == domainAuth.RoleAdmin {
		return nil
	}

	if !user.HasActiveTier {
		return pkgError.ValidationError("No active subscription tier found. Please activate your subscription to send messages.")
	}

	// -1 means unlimited sending
	if user.MessageLimit > 0 {
		if user.MessageCount+messageCount > user.MessageLimit {
			return fmt.Errorf("Message sending limit reached (%d / %d messages sent). Please upgrade to Sea plan for unlimited sending.", user.MessageCount, user.MessageLimit)
		}
	}

	return nil
}

func (s *authService) IncrementUserMessageCount(ctx context.Context, userID int64, count int) error {
	return s.repo.IncrementMessageCount(userID, count)
}

func (s *authService) ValidateUserBroadcastQuota(ctx context.Context, userID int64, messageCount int) error {
	if userID <= 0 {
		return nil
	}

	user, err := s.GetMe(ctx, userID)
	if err != nil {
		return err
	}
	if user.Role == domainAuth.RoleAdmin {
		return nil
	}

	if !user.HasActiveTier {
		return pkgError.ValidationError("No active subscription tier found. Please activate your subscription to broadcast messages.")
	}

	// -1 means unlimited broadcast
	if user.BroadcastLimit > 0 {
		if user.BroadcastCount+messageCount > user.BroadcastLimit {
			return fmt.Errorf("Broadcast limit exceeded (%d / %d broadcasts sent). Please upgrade to Sea plan for unlimited broadcasts.", user.BroadcastCount, user.BroadcastLimit)
		}
	}

	return nil
}

func (s *authService) IncrementUserBroadcastCount(ctx context.Context, userID int64, count int) error {
	return s.repo.IncrementBroadcastCount(userID, count)
}

func (s *authService) CheckAndPerformDailyReset(ctx context.Context) error {
	now := time.Now()
	period := fmt.Sprintf("daily:%s", now.Format("2006-01-02")) // e.g. "daily:2026-09-01"

	alreadyReset, err := s.repo.IsPeriodQuotaReset(ctx, period)
	if err != nil {
		return err
	}
	if alreadyReset {
		return nil
	}

	affected, err := s.repo.ResetDailyMessageQuotas(ctx, period)
	if err != nil {
		return err
	}
	logrus.Infof("[DAILY_QUOTA_RESET] Successfully reset daily direct message limits for period %s (%d users reset)", period, affected)
	return nil
}

func (s *authService) CheckAndPerformMonthlyReset(ctx context.Context) error {
	now := time.Now()
	period := fmt.Sprintf("monthly:%s", now.Format("2006-01")) // e.g. "monthly:2026-09"

	alreadyReset, err := s.repo.IsPeriodQuotaReset(ctx, period)
	if err != nil {
		return err
	}
	if alreadyReset {
		return nil
	}

	affected, err := s.repo.ResetMonthlyBroadcastQuotas(ctx, period)
	if err != nil {
		return err
	}
	logrus.Infof("[MONTHLY_QUOTA_RESET] Successfully reset monthly broadcast quotas for period %s (%d users reset)", period, affected)
	return nil
}

func (s *authService) AdminManualResetQuotas(ctx context.Context) (int64, error) {
	affected, err := s.repo.ManualResetAllQuotas(ctx)
	if err != nil {
		return 0, err
	}
	logrus.Infof("[MANUAL_QUOTA_RESET] Admin manually reset all user limits (%d users affected)", affected)
	return affected, nil
}

func (s *authService) GetUserByApiKey(ctx context.Context, apiKey string) (*domainAuth.User, error) {
	keyRecord, err := s.repo.GetApiKeyByValue(apiKey)
	if err != nil || keyRecord == nil {
		return nil, errors.New("invalid API key")
	}

	return s.repo.GetUserByID(keyRecord.UserID)
}

func (s *authService) GetPlans(ctx context.Context) ([]domainAuth.Plan, error) {
	plans, err := s.repo.GetAllPlans()
	if err != nil || len(plans) == 0 {
		var list []domainAuth.Plan
		for _, p := range domainAuth.AvailablePlans {
			list = append(list, p)
		}
		return list, nil
	}
	return plans, nil
}

func (s *authService) AdminListPlans(ctx context.Context) ([]domainAuth.Plan, error) {
	return s.GetPlans(ctx)
}

func (s *authService) AdminCreatePlan(ctx context.Context, req domainAuth.CreatePlanRequest) (domainAuth.Plan, error) {
	if req.ID == "" || req.Name == "" {
		return domainAuth.Plan{}, errors.New("id and name are required")
	}
	duration := req.DurationDays
	if duration <= 0 {
		duration = 30
	}
	deviceLimit := req.DeviceLimit
	if deviceLimit <= 0 {
		deviceLimit = 1
	}

	plan := domainAuth.Plan{
		ID:             req.ID,
		Name:           req.Name,
		DeviceLimit:    deviceLimit,
		MessageLimit:   req.MessageLimit,
		BroadcastLimit: req.BroadcastLimit,
		ApiKeyLimit:    req.ApiKeyLimit,
		WebhookLimit:   req.WebhookLimit,
		DurationDays:   duration,
		PriceMonthly:   req.PriceMonthly,
		Description:    req.Description,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreatePlan(&plan); err != nil {
		return domainAuth.Plan{}, err
	}
	return plan, nil
}

func (s *authService) AdminUpdatePlanConfig(ctx context.Context, planID domainAuth.PlanID, req domainAuth.UpdatePlanConfigRequest) (domainAuth.Plan, error) {
	existing, err := s.repo.GetPlanByID(planID)
	if err != nil || existing == nil {
		return domainAuth.Plan{}, errors.New("plan not found")
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.DeviceLimit > 0 {
		existing.DeviceLimit = req.DeviceLimit
	}
	if req.MessageLimit != 0 {
		existing.MessageLimit = req.MessageLimit
	}
	if req.BroadcastLimit != 0 {
		existing.BroadcastLimit = req.BroadcastLimit
	}
	if req.ApiKeyLimit >= 0 {
		existing.ApiKeyLimit = req.ApiKeyLimit
	}
	if req.WebhookLimit >= 0 {
		existing.WebhookLimit = req.WebhookLimit
	}
	if req.DurationDays > 0 {
		existing.DurationDays = req.DurationDays
	}
	if req.PriceMonthly >= 0 {
		existing.PriceMonthly = req.PriceMonthly
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.UpdatePlan(existing); err != nil {
		return domainAuth.Plan{}, err
	}
	return *existing, nil
}

func (s *authService) AdminDeletePlan(ctx context.Context, planID domainAuth.PlanID) error {
	return s.repo.DeletePlan(planID)
}

// _____________________________________________________________________________________________________________________
// Auto Responses Usecase Implementation

func (s *authService) GetUserAutoResponses(ctx context.Context, userID int64) ([]domainAuth.AutoResponse, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}
	return s.repo.GetUserAutoResponses(ctx, userID)
}

func (s *authService) CreateAutoResponse(ctx context.Context, userID int64, req domainAuth.CreateAutoResponseRequest) (domainAuth.AutoResponse, error) {
	if userID <= 0 {
		return domainAuth.AutoResponse{}, errors.New("invalid user id")
	}
	keyword := strings.TrimSpace(req.TriggerKeyword)
	if keyword == "" {
		return domainAuth.AutoResponse{}, errors.New("trigger_keyword is required")
	}
	responseMsg := strings.TrimSpace(req.ResponseMessage)
	if responseMsg == "" {
		return domainAuth.AutoResponse{}, errors.New("response_message is required")
	}

	trigType := req.TriggerType
	if trigType == "" {
		trigType = domainAuth.TriggerContains
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	ar := &domainAuth.AutoResponse{
		UserID:          userID,
		DeviceID:        strings.TrimSpace(req.DeviceID),
		TriggerType:     trigType,
		TriggerKeyword:  keyword,
		ResponseMessage: responseMsg,
		IsActive:        isActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.CreateAutoResponse(ctx, ar); err != nil {
		return domainAuth.AutoResponse{}, err
	}
	return *ar, nil
}

func (s *authService) UpdateAutoResponse(ctx context.Context, userID int64, id int64, req domainAuth.UpdateAutoResponseRequest) (domainAuth.AutoResponse, error) {
	if userID <= 0 || id <= 0 {
		return domainAuth.AutoResponse{}, errors.New("invalid id")
	}

	existingRules, err := s.repo.GetUserAutoResponses(ctx, userID)
	if err != nil {
		return domainAuth.AutoResponse{}, err
	}

	var target *domainAuth.AutoResponse
	for _, r := range existingRules {
		if r.ID == id {
			target = &r
			break
		}
	}
	if target == nil {
		return domainAuth.AutoResponse{}, errors.New("auto response rule not found")
	}

	if req.TriggerKeyword != "" {
		target.TriggerKeyword = strings.TrimSpace(req.TriggerKeyword)
	}
	if req.ResponseMessage != "" {
		target.ResponseMessage = strings.TrimSpace(req.ResponseMessage)
	}
	if req.TriggerType != "" {
		target.TriggerType = req.TriggerType
	}
	target.DeviceID = strings.TrimSpace(req.DeviceID)
	if req.IsActive != nil {
		target.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateAutoResponse(ctx, target); err != nil {
		return domainAuth.AutoResponse{}, err
	}
	return *target, nil
}

func (s *authService) DeleteAutoResponse(ctx context.Context, userID int64, id int64) error {
	if userID <= 0 || id <= 0 {
		return errors.New("invalid id")
	}
	return s.repo.DeleteAutoResponse(ctx, id, userID)
}


