package messaging

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

// MessageValidator is used to validate message.
// if error is not nil, this message will be handled by  MessageValidationHandler
// the second return value is the reply message, if not nil, the message will be sent to
// the client, if nil, the MessageValidationHandler will return the error message
type MessageValidator = func(msg *messages.GlideMessage) (error, *messages.GlideMessage)

// MessageValidationHandler validates message before handling
type MessageValidationHandler struct {
	validators []MessageValidator
}

func NewMessageValidationHandler(validators ...MessageValidator) *MessageValidationHandler {
	return &MessageValidationHandler{
		validators: validators,
	}
}

func (m *MessageValidationHandler) Handle(h *MessageInterfaceImpl, cliInfo *gate.Info, message *messages.GlideMessage) bool {

	for _, v := range m.validators {
		err, reply := v(message)
		if err != nil {
			if reply == nil {
				reply = messages.NewMessage(message.GetSeq(), messages.ActionNotifyError, err.Error())
			}
			_ = h.GetClientInterface().EnqueueMessage(cliInfo.ID, reply)
			return true
		}
	}

	return false
}

func DefaultMessageValidator(msg *messages.GlideMessage) (error, *messages.GlideMessage) {
	if msg.To == "" {
		return errors.New("message.To is empty"), nil
	}
	if msg.Action == "" {
		return errors.New("message.Action is empty"), nil
	}
	return nil, nil
}

// HandlerFunc is used to handle message with specified action in ActionHandler
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
