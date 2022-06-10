package messaging

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

type HandlerFunc func(cliInfo *gate.Info, message *messages.GlideMessage) error

// ActionHandler is a handler for a specific message action.
type ActionHandler struct {
	action messages.Action
	fn     HandlerFunc
}

func NewActionHandler(action messages.Action, fn HandlerFunc) *ActionHandler {
	return &ActionHandler{
		action: action,
		fn:     fn,
	}
}

func (a *ActionHandler) Handle(h *MessageInterfaceImpl, cliInfo *gate.Info, message *messages.GlideMessage) bool {
	if message.GetAction() == a.action {
		err := a.fn(cliInfo, message)
		if err != nil {
			h.OnHandleMessageError(cliInfo, message, err)
		}
		return true
	}
	return false
}

type ReplyHandlerFunc func(cliInfo *gate.Info, message *messages.GlideMessage) (*messages.GlideMessage, error)

// ActionWithReplyHandler is a handler for a specific message action, this handler will return a reply message.
type ActionWithReplyHandler struct {
	action messages.Action
	fn     ReplyHandlerFunc
}

func NewActionWithReplyHandler(action messages.Action, fn ReplyHandlerFunc) *ActionWithReplyHandler {
	return &ActionWithReplyHandler{
		action: action,
		fn:     fn,
	}
}

func (rh *ActionWithReplyHandler) Handle(h *MessageInterfaceImpl, cInfo *gate.Info, msg *messages.GlideMessage) bool {
	if msg.GetAction() == rh.action {
		r, err := rh.fn(cInfo, msg)
		if err != nil {
			h.OnHandleMessageError(cInfo, msg, err)
		}
		_ = h.GetClientInterface().EnqueueMessage(cInfo.ID, r)
		return true
	}
	return false
}
