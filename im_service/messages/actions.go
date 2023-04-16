package messages

import (
	"github.com/glide-im/glide/pkg/messages"
)

const (
	ActionHello               messages.Action = "hello"
	ActionHeartbeat                           = "heartbeat"
	ActionNotifyUnknownAction                 = "notify.unknown.action"

	ActionChatMessage       = "message.chat"
	ActionChatMessageResend = "message.chat.resend"
	ActionGroupMessage      = "message.group"
	ActionMessageFailed     = "message.failed.send"

	ActionNotifyNeedAuth      = "notify.auth"
	ActionNotifyKickOut       = "notify.kickout"
	ActionNotifyNewContact    = "notify.contact"
	ActionNotifyGroup         = "notify.group"
	ActionNotifyAccountLogin  = "notify.login"
	ActionNotifyAccountLogout = "notify.logout"
	ActionNotifyError         = "notify.error"

	ActionAckRequest  = "ack.request"
	ActionAckGroupMsg = "ack.group.msg"
	ActionAckMessage  = "ack.message"
	ActionAckNotify   = "ack.notify"

	ActionApiAuth    = "api.auth"
	ActionApiFailed  = "api.failed"
	ActionApiSuccess = "api.success"

	ActionClientCustom = "message.cli"

	NotifyKickOut messages.Action = "notify.kickout"
	AckOffline                    = "ack.offline"
)
