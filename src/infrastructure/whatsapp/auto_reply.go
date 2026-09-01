package whatsapp

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainAuth "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/auth"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func handleAutoReply(ctx context.Context, evt *events.Message, chatStorageRepo domainChatStorage.IChatStorageRepository, client *whatsmeow.Client) {
	if client == nil {
		return
	}

	// Skip messages sent by ourselves or channel broadcast broadcasts
	if evt.Info.IsFromMe || evt.Info.IsIncomingBroadcast() {
		return
	}

	// Extra safety: skip WhatsApp status updates and newsletters
	source := evt.Info.SourceString()
	chatStr := evt.Info.Chat.String()
	if strings.Contains(source, "broadcast") ||
		strings.HasSuffix(chatStr, "@broadcast") ||
		strings.HasPrefix(chatStr, "status@") ||
		strings.HasSuffix(chatStr, "@newsletter") {
		return
	}

	// Extract incoming message text
	incomingText := extractMessageText(evt)
	if strings.TrimSpace(incomingText) == "" {
		return
	}

	log.Infof("[AUTO_RESPONSE] Processing incoming message from %s: %q", evt.Info.SourceString(), incomingText)

	// Resolve Device ID from context, instance or client store
	deviceID := ""
	if inst, ok := DeviceFromContext(ctx); ok && inst != nil {
		deviceID = inst.ID()
	}
	if deviceID == "" && client.Store != nil && client.Store.ID != nil {
		deviceID = client.Store.ID.User
	}

	// Look up dynamic Auto Response rules from database
	var matchedReply string
	if chatStorageRepo != nil {
		rules, err := chatStorageRepo.GetActiveAutoResponses(ctx, deviceID)
		if err != nil {
			log.Errorf("[AUTO_RESPONSE] Error querying active rules: %v", err)
		} else {
			log.Infof("[AUTO_RESPONSE] Found %d active rules for device %q", len(rules), deviceID)
			matchedReply = matchAutoResponseRule(incomingText, rules, evt)
		}
	}

	// Fallback to global static .env auto reply message if no rule matched
	if matchedReply == "" && config.WhatsappAutoReplyMessage != "" {
		if !utils.IsGroupJID(evt.Info.Chat.String()) {
			matchedReply = config.WhatsappAutoReplyMessage
		}
	}

	if strings.TrimSpace(matchedReply) == "" {
		log.Debugf("[AUTO_RESPONSE] No rule matched for text: %q", incomingText)
		return
	}

	// Resolve recipient JID (1:1 chat or sender)
	recipientJID := evt.Info.Chat
	if recipientJID.IsEmpty() {
		recipientJID = evt.Info.Sender
	}
	if recipientJID.IsEmpty() {
		log.Errorf("[AUTO_RESPONSE] Unable to resolve recipient JID for event %s", evt.Info.ID)
		return
	}

	log.Infof("[AUTO_RESPONSE] Sending matched auto-reply to %s: %q", recipientJID.String(), matchedReply)

	// Use detached context with timeout to avoid event loop cancellation
	sendCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Send the automated response
	response, err := client.SendMessage(
		sendCtx,
		recipientJID,
		&waE2E.Message{Conversation: proto.String(matchedReply)},
	)
	if err != nil {
		log.Errorf("[AUTO_RESPONSE] Failed to send auto-reply to %s: %v", recipientJID.String(), err)
		return
	}

	// Store sent auto-reply message into chat storage
	if chatStorageRepo != nil {
		senderJID := ""
		if client.Store != nil && client.Store.ID != nil {
			senderJID = client.Store.ID.String()
		}
		_ = chatStorageRepo.StoreSentMessageWithContext(
			sendCtx,
			response.ID,
			senderJID,
			recipientJID.String(),
			matchedReply,
			response.Timestamp,
			nil,
		)
	}

	log.Infof("[AUTO_RESPONSE] Successfully sent auto-reply to %s (Response: %q)", recipientJID.String(), matchedReply)
}

func extractMessageText(evt *events.Message) string {
	if evt == nil || evt.Message == nil {
		return ""
	}
	inner := utils.UnwrapMessage(evt.Message)
	if conv := inner.GetConversation(); conv != "" {
		return conv
	}
	if ext := inner.GetExtendedTextMessage(); ext != nil && ext.GetText() != "" {
		return ext.GetText()
	}
	if img := inner.GetImageMessage(); img != nil && img.GetCaption() != "" {
		return img.GetCaption()
	}
	if vid := inner.GetVideoMessage(); vid != nil && vid.GetCaption() != "" {
		return vid.GetCaption()
	}
	if doc := inner.GetDocumentMessage(); doc != nil && doc.GetCaption() != "" {
		return doc.GetCaption()
	}
	if protoMsg := inner.GetProtocolMessage(); protoMsg != nil {
		if edited := protoMsg.GetEditedMessage(); edited != nil {
			if ext := edited.GetExtendedTextMessage(); ext != nil && ext.GetText() != "" {
				return ext.GetText()
			}
			if conv := edited.GetConversation(); conv != "" {
				return conv
			}
		}
	}
	return ""
}

func matchAutoResponseRule(incomingText string, rules []domainAuth.AutoResponse, evt *events.Message) string {
	cleanText := strings.TrimSpace(incomingText)
	lowerText := strings.ToLower(cleanText)

	for _, rule := range rules {
		if !rule.IsActive {
			continue
		}

		keyword := strings.TrimSpace(rule.TriggerKeyword)
		if keyword == "" {
			continue
		}
		lowerKeyword := strings.ToLower(keyword)

		matched := false
		trigType := strings.ToLower(strings.TrimSpace(string(rule.TriggerType)))
		switch trigType {
		case "exact":
			matched = (lowerText == lowerKeyword)
		case "starts_with":
			matched = strings.HasPrefix(lowerText, lowerKeyword)
		case "contains":
			matched = strings.Contains(lowerText, lowerKeyword)
		case "regex":
			if re, err := regexp.Compile("(?i)" + keyword); err == nil {
				matched = re.MatchString(cleanText)
			}
		default:
			matched = strings.Contains(lowerText, lowerKeyword) || (lowerText == lowerKeyword)
		}

		if matched {
			log.Infof("[AUTO_RESPONSE] Rule %d matched! (Type: %s, Keyword: %q)", rule.ID, rule.TriggerType, rule.TriggerKeyword)
			return interpolateVariables(rule.ResponseMessage, evt)
		}
	}
	return ""
}

func interpolateVariables(template string, evt *events.Message) string {
	now := time.Now()
	pushName := ""
	senderUser := ""
	if evt != nil {
		pushName = strings.TrimSpace(evt.Info.PushName)
		senderUser = strings.TrimSpace(evt.Info.Sender.User)
	}
	if pushName == "" {
		pushName = senderUser
	}

	res := template
	// Replace {name}
	res = strings.ReplaceAll(res, "{name}", pushName)
	res = strings.ReplaceAll(res, "{NAME}", pushName)
	res = strings.ReplaceAll(res, "{Name}", pushName)

	// Replace {sender}
	res = strings.ReplaceAll(res, "{sender}", senderUser)
	res = strings.ReplaceAll(res, "{SENDER}", senderUser)
	res = strings.ReplaceAll(res, "{Sender}", senderUser)

	// Replace {time}
	timeStr := now.Format("15:04")
	res = strings.ReplaceAll(res, "{time}", timeStr)
	res = strings.ReplaceAll(res, "{TIME}", timeStr)
	res = strings.ReplaceAll(res, "{Time}", timeStr)

	// Replace {date}
	dateStr := now.Format("02 Jan 2006")
	res = strings.ReplaceAll(res, "{date}", dateStr)
	res = strings.ReplaceAll(res, "{DATE}", dateStr)
	res = strings.ReplaceAll(res, "{Date}", dateStr)

	return res
}
