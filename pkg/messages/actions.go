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

	ActionAuthenticate          = "authenticate"
	ActionNotifyError           = "notify.error"
	ActionNotifySuccess         = "notify.success"
	ActionNotifyKickOut         = "notify.kickout"
	ActionNotifyForbidden       = "notify.forbidden"
	ActionNotifyUnauthenticated = "notify.unauthenticated"

	ActionInternalOnline  = "internal.online"
	ActionInternalOffline = "internal.offline"
)

func (a Action) IsInternal() bool {
	return strings.HasPrefix(string(a), "internal.")
}
