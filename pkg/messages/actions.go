package messages

import "strings"

// Action is the type of action that is being performed.
type Action string

const (
	ActionHello               Action = "hello"
	ActionHeartbeat                  = "heartbeat"
	ActionNotifyUnknownAction        = "notify.unknown.action"

	ActionNotifyError = "notify.error"

	ActionInternalOnline  = "internal.online"
	ActionInternalOffline = "internal.offline"
)

func (a Action) IsInternal() bool {
	return strings.HasPrefix(string(a), "internal.")
}
