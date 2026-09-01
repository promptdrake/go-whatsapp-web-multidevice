package whatsapp

import (
	"context"
	"time"

	domainAuth "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/auth"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// deviceChatStorage wraps a base repository and injects device_id into operations
// to keep chat/message separation per device without importing chatstorage to avoid cycles.
type deviceChatStorage struct {
	deviceID string
	base     domainChatStorage.IChatStorageRepository
}

func newDeviceChatStorage(deviceID string, base domainChatStorage.IChatStorageRepository) domainChatStorage.IChatStorageRepository {
	if base == nil {
		return nil
	}
	return &deviceChatStorage{
		deviceID: deviceID,
		base:     base,
	}
}

func (r *deviceChatStorage) withDeviceChat(chat *domainChatStorage.Chat) *domainChatStorage.Chat {
	if chat == nil {
		return nil
	}
	if chat.DeviceID == "" {
		chat.DeviceID = r.deviceID
	}
	return chat
}

func (r *deviceChatStorage) CreateMessage(ctx context.Context, evt *events.Message) error {
	return r.base.CreateMessage(ctx, evt)
}

func (r *deviceChatStorage) CreateReaction(ctx context.Context, evt *events.Message) error {
	return r.base.CreateReaction(ctx, evt)
}

func (r *deviceChatStorage) CreateIncomingCallRecord(ctx context.Context, evt *events.CallOffer, autoRejected bool) error {
	return r.base.CreateIncomingCallRecord(ctx, evt, autoRejected)
}

func (r *deviceChatStorage) StoreChat(chat *domainChatStorage.Chat) error {
	return r.base.StoreChat(r.withDeviceChat(chat))
}

func (r *deviceChatStorage) GetChat(jid string) (*domainChatStorage.Chat, error) {
	return r.base.GetChatByDevice(r.deviceID, jid)
}

func (r *deviceChatStorage) GetChatByDevice(deviceID, jid string) (*domainChatStorage.Chat, error) {
	return r.base.GetChatByDevice(deviceID, jid)
}

func (r *deviceChatStorage) GetChats(filter *domainChatStorage.ChatFilter) ([]*domainChatStorage.Chat, error) {
	if filter != nil && filter.DeviceID == "" {
		filter.DeviceID = r.deviceID
	}
	return r.base.GetChats(filter)
}

func (r *deviceChatStorage) DeleteChat(jid string) error {
	return r.base.DeleteChatByDevice(r.deviceID, jid)
}

func (r *deviceChatStorage) DeleteChatByDevice(deviceID, jid string) error {
	return r.base.DeleteChatByDevice(deviceID, jid)
}

func (r *deviceChatStorage) StoreMessage(message *domainChatStorage.Message) error {
	return r.base.StoreMessage(message)
}

func (r *deviceChatStorage) StoreMessageEdit(edit *domainChatStorage.MessageEdit) error {
	if edit != nil && edit.DeviceID == "" {
		edit.DeviceID = r.deviceID
	}
	return r.base.StoreMessageEdit(edit)
}

func (r *deviceChatStorage) StoreMessagesBatch(messages []*domainChatStorage.Message) error {
	return r.base.StoreMessagesBatch(messages)
}

func (r *deviceChatStorage) GetMessageByID(id string) (*domainChatStorage.Message, error) {
	return r.base.GetMessageByID(id)
}

func (r *deviceChatStorage) GetMessageByIDAndDevice(deviceID, id string) (*domainChatStorage.Message, error) {
	targetDeviceID := deviceID
	if targetDeviceID == "" {
		targetDeviceID = r.deviceID
	}
	return r.base.GetMessageByIDAndDevice(targetDeviceID, id)
}

func (r *deviceChatStorage) GetMessageByIDChatAndDevice(deviceID, chatJID, id string) (*domainChatStorage.Message, error) {
	targetDeviceID := deviceID
	if targetDeviceID == "" {
		targetDeviceID = r.deviceID
	}
	return r.base.GetMessageByIDChatAndDevice(targetDeviceID, chatJID, id)
}

func (r *deviceChatStorage) GetMessageEdits(originalMessageID, deviceID string) ([]*domainChatStorage.MessageEdit, error) {
	targetDeviceID := deviceID
	if targetDeviceID == "" {
		targetDeviceID = r.deviceID
	}
	return r.base.GetMessageEdits(originalMessageID, targetDeviceID)
}

func (r *deviceChatStorage) GetMessages(filter *domainChatStorage.MessageFilter) ([]*domainChatStorage.Message, error) {
	if filter != nil && filter.DeviceID == "" {
		filter.DeviceID = r.deviceID
	}
	return r.base.GetMessages(filter)
}

func (r *deviceChatStorage) SearchMessages(deviceID, chatJID, searchText string, limit int) ([]*domainChatStorage.Message, error) {
	targetDeviceID := deviceID
	if targetDeviceID == "" {
		targetDeviceID = r.deviceID
	}
	return r.base.SearchMessages(targetDeviceID, chatJID, searchText, limit)
}

func (r *deviceChatStorage) DeleteMessage(id, chatJID string) error {
	return r.base.DeleteMessageByDevice(r.deviceID, id, chatJID)
}

func (r *deviceChatStorage) DeleteMessageByDevice(deviceID, id, chatJID string) error {
	return r.base.DeleteMessageByDevice(deviceID, id, chatJID)
}

func (r *deviceChatStorage) UpsertChatwootMessageLink(link *domainChatStorage.ChatwootMessageLink) error {
	if link != nil && link.DeviceID == "" {
		link.DeviceID = r.deviceID
	}
	return r.base.UpsertChatwootMessageLink(link)
}

func (r *deviceChatStorage) GetChatwootMessageLinkByWhatsAppID(deviceID, waMessageID string) (*domainChatStorage.ChatwootMessageLink, error) {
	targetDeviceID := deviceID
	if targetDeviceID == "" {
		targetDeviceID = r.deviceID
	}
	return r.base.GetChatwootMessageLinkByWhatsAppID(targetDeviceID, waMessageID)
}

func (r *deviceChatStorage) GetChatwootMessageLinkByChatwootID(deviceID string, chatwootMessageID int) (*domainChatStorage.ChatwootMessageLink, error) {
	targetDeviceID := deviceID
	if targetDeviceID == "" {
		targetDeviceID = r.deviceID
	}
	return r.base.GetChatwootMessageLinkByChatwootID(targetDeviceID, chatwootMessageID)
}

func (r *deviceChatStorage) GetLatestChatwootMessageLinkByConversation(conversationID, accountID int, allowLegacyZero bool, configID int64) (*domainChatStorage.ChatwootMessageLink, error) {
	return r.base.GetLatestChatwootMessageLinkByConversation(conversationID, accountID, allowLegacyZero, configID)
}

func (r *deviceChatStorage) BackfillChatwootMessageLinkAccount(accountID int) (int64, error) {
	return r.base.BackfillChatwootMessageLinkAccount(accountID)
}

func (r *deviceChatStorage) CountChatwootMessageLinksByConfig(configID int64) (int, error) {
	return r.base.CountChatwootMessageLinksByConfig(configID)
}

func (r *deviceChatStorage) DeleteChatwootMessageLinksByConfig(configID int64) error {
	return r.base.DeleteChatwootMessageLinksByConfig(configID)
}

func (r *deviceChatStorage) SaveChatwootDeviceConfig(cfg *domainChatStorage.ChatwootDeviceConfig) error {
	return r.base.SaveChatwootDeviceConfig(cfg)
}

func (r *deviceChatStorage) UpdateChatwootDeviceConfigJID(deviceID, deviceJID string) (bool, error) {
	return r.base.UpdateChatwootDeviceConfigJID(deviceID, deviceJID)
}

func (r *deviceChatStorage) GetChatwootDeviceConfig(deviceID string) (*domainChatStorage.ChatwootDeviceConfig, error) {
	return r.base.GetChatwootDeviceConfig(deviceID)
}

func (r *deviceChatStorage) GetChatwootDeviceConfigByIdentifier(identifier string) (*domainChatStorage.ChatwootDeviceConfig, error) {
	return r.base.GetChatwootDeviceConfigByIdentifier(identifier)
}

func (r *deviceChatStorage) GetChatwootDeviceConfigByInbox(accountID, inboxID int) (*domainChatStorage.ChatwootDeviceConfig, error) {
	return r.base.GetChatwootDeviceConfigByInbox(accountID, inboxID)
}

func (r *deviceChatStorage) ListChatwootDeviceConfigs() ([]*domainChatStorage.ChatwootDeviceConfig, error) {
	return r.base.ListChatwootDeviceConfigs()
}

func (r *deviceChatStorage) DeleteChatwootDeviceConfig(deviceID string) error {
	return r.base.DeleteChatwootDeviceConfig(deviceID)
}

func (r *deviceChatStorage) CountChatwootDeviceConfigs() (int, error) {
	return r.base.CountChatwootDeviceConfigs()
}

func (r *deviceChatStorage) GetLatestUnreadChatwootMessageLinkByChat(deviceID, waChatJID string) (*domainChatStorage.ChatwootMessageLink, error) {
	targetDeviceID := deviceID
	if targetDeviceID == "" {
		targetDeviceID = r.deviceID
	}
	return r.base.GetLatestUnreadChatwootMessageLinkByChat(targetDeviceID, waChatJID)
}

func (r *deviceChatStorage) EnqueueChatwootForwardEvent(event *domainChatStorage.ChatwootForwardEvent) error {
	if event != nil && event.DeviceID == "" {
		event.DeviceID = r.deviceID
	}
	return r.base.EnqueueChatwootForwardEvent(event)
}

func (r *deviceChatStorage) ListDueChatwootForwardEvents(now time.Time, limit int) ([]*domainChatStorage.ChatwootForwardEvent, error) {
	return r.base.ListDueChatwootForwardEvents(now, limit)
}

func (r *deviceChatStorage) MarkChatwootForwardEventFailed(id int64, lastError string, nextAttemptAt time.Time) error {
	return r.base.MarkChatwootForwardEventFailed(id, lastError, nextAttemptAt)
}

func (r *deviceChatStorage) MarkChatwootForwardEventDone(id int64) error {
	return r.base.MarkChatwootForwardEventDone(id)
}

func (r *deviceChatStorage) StoreSentMessageWithContext(ctx context.Context, messageID string, senderJID string, recipientJID string, content string, timestamp time.Time, msg *waE2E.Message) error {
	if _, ok := DeviceFromContext(ctx); !ok && r.deviceID != "" {
		ctx = ContextWithDevice(ctx, NewDeviceInstance(r.deviceID, nil, nil))
	}
	return r.base.StoreSentMessageWithContext(ctx, messageID, senderJID, recipientJID, content, timestamp, msg)
}

func (r *deviceChatStorage) UpsertPollDefinition(definition *domainChatStorage.PollDefinition) error {
	if definition != nil && definition.DeviceID == "" {
		definition.DeviceID = r.deviceID
	}
	return r.base.UpsertPollDefinition(definition)
}

func (r *deviceChatStorage) GetPollDefinition(deviceID, chatJID, pollMessageID string) (*domainChatStorage.PollDefinition, error) {
	targetDeviceID := deviceID
	if targetDeviceID == "" {
		targetDeviceID = r.deviceID
	}
	return r.base.GetPollDefinition(targetDeviceID, chatJID, pollMessageID)
}

func (r *deviceChatStorage) GetPollDefinitionByIDAndDevice(deviceID, pollMessageID string) (*domainChatStorage.PollDefinition, error) {
	targetDeviceID := deviceID
	if targetDeviceID == "" {
		targetDeviceID = r.deviceID
	}
	return r.base.GetPollDefinitionByIDAndDevice(targetDeviceID, pollMessageID)
}

func (r *deviceChatStorage) AppendPollOption(deviceID, chatJID, pollMessageID string, option domainChatStorage.PollOption) error {
	targetDeviceID := deviceID
	if targetDeviceID == "" {
		targetDeviceID = r.deviceID
	}
	return r.base.AppendPollOption(targetDeviceID, chatJID, pollMessageID, option)
}

func (r *deviceChatStorage) GetChatMessageCount(chatJID string) (int64, error) {
	return r.base.GetChatMessageCountByDevice(r.deviceID, chatJID)
}

func (r *deviceChatStorage) GetChatMessageCountByDevice(deviceID, chatJID string) (int64, error) {
	return r.base.GetChatMessageCountByDevice(deviceID, chatJID)
}

func (r *deviceChatStorage) GetTotalMessageCount() (int64, error) {
	return r.base.GetTotalMessageCount()
}

func (r *deviceChatStorage) GetTotalChatCount() (int64, error) {
	return r.base.GetTotalChatCount()
}

func (r *deviceChatStorage) GetFilteredChatCount(filter *domainChatStorage.ChatFilter) (int64, error) {
	if filter != nil && filter.DeviceID == "" {
		filter.DeviceID = r.deviceID
	}
	return r.base.GetFilteredChatCount(filter)
}

func (r *deviceChatStorage) GetChatNameWithPushName(jid types.JID, chatJID string, senderUser string, pushName string) string {
	return r.base.GetChatNameWithPushNameByDevice(r.deviceID, jid, chatJID, senderUser, pushName)
}

func (r *deviceChatStorage) GetChatNameWithPushNameByDevice(deviceID string, jid types.JID, chatJID string, senderUser string, pushName string) string {
	return r.base.GetChatNameWithPushNameByDevice(deviceID, jid, chatJID, senderUser, pushName)
}

func (r *deviceChatStorage) GetStorageStatistics() (chatCount int64, messageCount int64, err error) {
	return r.base.GetStorageStatistics()
}

func (r *deviceChatStorage) TruncateAllChats() error {
	return r.base.TruncateAllChats()
}

func (r *deviceChatStorage) TruncateAllDataWithLogging(logPrefix string) error {
	return r.base.TruncateAllDataWithLogging(logPrefix)
}

func (r *deviceChatStorage) InitializeSchema() error {
	return r.base.InitializeSchema()
}

func (r *deviceChatStorage) DeleteDeviceData(deviceID string) error {
	if r.base == nil {
		return nil
	}
	target := deviceID
	if target == "" {
		target = r.deviceID
	}
	return r.base.DeleteDeviceData(target)
}

func (r *deviceChatStorage) SaveDeviceRecord(record *domainChatStorage.DeviceRecord) error {
	return r.base.SaveDeviceRecord(record)
}

func (r *deviceChatStorage) ListDeviceRecords() ([]*domainChatStorage.DeviceRecord, error) {
	return r.base.ListDeviceRecords()
}

// GetDeviceRecord delegates to the base repository.
func (r *deviceChatStorage) GetDeviceRecord(deviceID string) (*domainChatStorage.DeviceRecord, error) {
	return r.base.GetDeviceRecord(deviceID)
}

// GetDeviceRecordByJID fetches a device record by its JID.
func (r *deviceChatStorage) GetDeviceRecordByJID(jid string) (*domainChatStorage.DeviceRecord, error) {
	return r.base.GetDeviceRecordByJID(jid)
}

// DeleteDeviceRecord delegates to the base repository.
func (r *deviceChatStorage) DeleteDeviceRecord(deviceID string) error {
	return r.base.DeleteDeviceRecord(deviceID)
}

// SetDeviceWebhookURL sets or clears the webhook URL for a device.
// Pass a nil webhookURL to clear the device-specific webhook (falls back to global).
// Pass a non-nil string pointer to set a device-specific webhook override.
func (r *deviceChatStorage) SetDeviceWebhookURL(deviceID string, webhookURL *string) error {
	return r.base.SetDeviceWebhookURL(deviceID, webhookURL)
}

// GetDeviceWebhookURL retrieves the configured webhook URL for a device.
// Returns nil if no device-specific webhook is set (caller should use global).
func (r *deviceChatStorage) GetDeviceWebhookURL(deviceID string) (*string, error) {
	return r.base.GetDeviceWebhookURL(deviceID)
}

// SetDeviceWebhookConfig sets the complete webhook configuration for a device.
func (r *deviceChatStorage) SetDeviceWebhookConfig(deviceID string, config *domainChatStorage.DeviceWebhookConfig) error {
	return r.base.SetDeviceWebhookConfig(deviceID, config)
}

// GetDeviceWebhookConfig retrieves the complete webhook configuration for a device.
func (r *deviceChatStorage) GetDeviceWebhookConfig(deviceID string) (*domainChatStorage.DeviceWebhookConfig, error) {
	return r.base.GetDeviceWebhookConfig(deviceID)
}

func (r *deviceChatStorage) CreateUser(user *domainAuth.User) error {
	return r.base.CreateUser(user)
}

func (r *deviceChatStorage) GetUserByEmail(email string) (*domainAuth.User, error) {
	return r.base.GetUserByEmail(email)
}

func (r *deviceChatStorage) GetUserByID(id int64) (*domainAuth.User, error) {
	return r.base.GetUserByID(id)
}

func (r *deviceChatStorage) UpdateUser(user *domainAuth.User) error {
	return r.base.UpdateUser(user)
}

func (r *deviceChatStorage) UpdateUserTier(userID int64, tierID domainAuth.PlanID, tierName string, expiresAt *time.Time) error {
	return r.base.UpdateUserTier(userID, tierID, tierName, expiresAt)
}

func (r *deviceChatStorage) IncrementMessageCount(userID int64, count int) error {
	return r.base.IncrementMessageCount(userID, count)
}

func (r *deviceChatStorage) IncrementBroadcastCount(userID int64, count int) error {
	return r.base.IncrementBroadcastCount(userID, count)
}

func (r *deviceChatStorage) ResetDailyMessageQuotas(ctx context.Context, period string) (int64, error) {
	return r.base.ResetDailyMessageQuotas(ctx, period)
}

func (r *deviceChatStorage) ResetMonthlyBroadcastQuotas(ctx context.Context, period string) (int64, error) {
	return r.base.ResetMonthlyBroadcastQuotas(ctx, period)
}

func (r *deviceChatStorage) ResetMonthlyQuotas(ctx context.Context, period string) (int64, error) {
	return r.base.ResetMonthlyQuotas(ctx, period)
}

func (r *deviceChatStorage) IsPeriodQuotaReset(ctx context.Context, period string) (bool, error) {
	return r.base.IsPeriodQuotaReset(ctx, period)
}

func (r *deviceChatStorage) ManualResetAllQuotas(ctx context.Context) (int64, error) {
	return r.base.ManualResetAllQuotas(ctx)
}

func (r *deviceChatStorage) ListUsers() ([]*domainAuth.User, error) {
	return r.base.ListUsers()
}

func (r *deviceChatStorage) DeleteUser(userID int64) error {
	return r.base.DeleteUser(userID)
}

func (r *deviceChatStorage) CountDevicesByUserID(userID int64) (int, error) {
	return r.base.CountDevicesByUserID(userID)
}

func (r *deviceChatStorage) ListDeviceRecordsByUserID(userID int64) ([]*domainChatStorage.DeviceRecord, error) {
	return r.base.ListDeviceRecordsByUserID(userID)
}

func (r *deviceChatStorage) CreateApiKey(apiKey *domainAuth.ApiKey) error {
	return r.base.CreateApiKey(apiKey)
}

func (r *deviceChatStorage) GetApiKeysByUserID(userID int64) ([]*domainAuth.ApiKey, error) {
	return r.base.GetApiKeysByUserID(userID)
}

func (r *deviceChatStorage) GetApiKeyByValue(keyValue string) (*domainAuth.ApiKey, error) {
	return r.base.GetApiKeyByValue(keyValue)
}

func (r *deviceChatStorage) DeleteApiKey(id, userID int64) error {
	return r.base.DeleteApiKey(id, userID)
}

func (r *deviceChatStorage) CreateUserWebhook(webhook *domainAuth.UserWebhook) error {
	return r.base.CreateUserWebhook(webhook)
}

func (r *deviceChatStorage) GetUserWebhooksByUserID(userID int64) ([]*domainAuth.UserWebhook, error) {
	return r.base.GetUserWebhooksByUserID(userID)
}

func (r *deviceChatStorage) DeleteUserWebhook(id, userID int64) error {
	return r.base.DeleteUserWebhook(id, userID)
}

func (r *deviceChatStorage) GetAllPlans() ([]domainAuth.Plan, error) {
	return r.base.GetAllPlans()
}

func (r *deviceChatStorage) GetPlanByID(id domainAuth.PlanID) (*domainAuth.Plan, error) {
	return r.base.GetPlanByID(id)
}

func (r *deviceChatStorage) CreatePlan(plan *domainAuth.Plan) error {
	return r.base.CreatePlan(plan)
}

func (r *deviceChatStorage) UpdatePlan(plan *domainAuth.Plan) error {
	return r.base.UpdatePlan(plan)
}

func (r *deviceChatStorage) DeletePlan(id domainAuth.PlanID) error {
	return r.base.DeletePlan(id)
}

func (r *deviceChatStorage) GetActiveAutoResponses(ctx context.Context, deviceID string) ([]domainAuth.AutoResponse, error) {
	targetDeviceID := deviceID
	if targetDeviceID == "" {
		targetDeviceID = r.deviceID
	}
	return r.base.GetActiveAutoResponses(ctx, targetDeviceID)
}

func (r *deviceChatStorage) GetUserAutoResponses(ctx context.Context, userID int64) ([]domainAuth.AutoResponse, error) {
	return r.base.GetUserAutoResponses(ctx, userID)
}

func (r *deviceChatStorage) CreateAutoResponse(ctx context.Context, ar *domainAuth.AutoResponse) error {
	return r.base.CreateAutoResponse(ctx, ar)
}

func (r *deviceChatStorage) UpdateAutoResponse(ctx context.Context, ar *domainAuth.AutoResponse) error {
	return r.base.UpdateAutoResponse(ctx, ar)
}

func (r *deviceChatStorage) DeleteAutoResponse(ctx context.Context, id int64, userID int64) error {
	return r.base.DeleteAutoResponse(ctx, id, userID)
}


