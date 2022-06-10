package messages

type Action string

const (
	ActionChatMessage       Action = "message.chat"
	ActionChatMessageResend        = "message.chat.resend"
	ActionGroupMessage             = "message.group"
	ActionMessageFailed            = "message.failed.send"
	ActionClientCustom             = "message.cli"

	ActionNotifyNeedAuth      = "notify.auth"
	ActionNotifyKickOut       = "notify.kickout"
	ActionNotifyNewContact    = "notify.contact"
	ActionNotifyGroup         = "notify.group"
	ActionNotifyAccountLogin  = "notify.login"
	ActionNotifyAccountLogout = "notify.logout"
	ActionNotifyError         = "notify.error"
	ActionNotifyUnknownAction = "notify.unknown.action"

	ActionAckRequest  = "ack.request"
	ActionAckGroupMsg = "ack.group.msg"
	ActionAckMessage  = "ack.message"
	ActionAckNotify   = "ack.notify"

	ActionHeartbeat  = "heartbeat"
	ActionApiAuth    = "api.auth"
	ActionApiFailed  = "api.failed"
	ActionApiSuccess = "api.success"
)
