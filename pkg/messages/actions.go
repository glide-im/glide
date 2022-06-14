package messages

import "strings"

// Action is the type of action that is being performed.
type Action string

const (
	ActionHello               Action = "hello"
	ActionHeartbeat                  = "heartbeat"
	ActionNotifyUnknownAction        = "notify.unknown.action"

	ActionChatMessage       = "message.chat"
	ActionChatMessageResend = "message.chat.resend"
	ActionGroupMessage      = "message.group"
	ActionMessageFailed     = "message.failed.send"
	ActionClientCustom      = "message.cli"

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

	ActionInternalOnline  = "internal.online"
	ActionInternalOffline = "internal.offline"
)

func (a Action) IsInternal() bool {
	return strings.HasPrefix(string(a), "internal.")
}
