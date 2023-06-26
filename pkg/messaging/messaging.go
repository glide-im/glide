package messaging

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/panjf2000/ants/v2"
)

// MessageHandler is the interface for message handler
type MessageHandler interface {
	// Handle handles the message, returns true if the message is handled,
	// otherwise the message is delegated to next handler.
	Handle(h *MessageInterfaceImpl, cliInfo *gate.Info, message *messages.GlideMessage) bool
}

// Interface for messaging.
type Interface interface {

	// Handle handles message from gate, the entry point for the messaging.
	Handle(clientInfo *gate.Info, msg *messages.GlideMessage) error

	AddHandler(i MessageHandler)
}

type Messaging interface {
	Interface

	SetSubscription(g subscription.Interface)

	SetGate(g gate.Gateway)
}

// Server is the messaging server.
type Server interface {
	Messaging

	Run() error
}

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

type Options struct {
	NotifyServerError     bool
	MaxMessageConcurrency int
}

func onMessageHandlerPanic(i interface{}) {
	logger.E("MessageInterfaceImpl panic: %v", i)
}

// MessageInterfaceImpl default implementation of the messaging interface.
type MessageInterfaceImpl struct {

	// execPool 100 capacity goroutine pool, 假设每个消息处理需要10ms, 一个协程则每秒能处理100条消息
	execPool *ants.Pool

	// hc message handler chain
	hc *handlerChain

	subscription subscription.Interface
	gate         gate.Gateway

	// notifyOnSrvErr notify client on server error
	notifyOnSrvErr bool
}

func NewDefaultImpl(options *Options) (*MessageInterfaceImpl, error) {

	ret := MessageInterfaceImpl{
		notifyOnSrvErr: options.NotifyServerError,
		hc:             &handlerChain{},
	}

	var err error
	ret.execPool, err = ants.NewPool(
		options.MaxMessageConcurrency,
		ants.WithNonblocking(true),
		ants.WithPanicHandler(onMessageHandlerPanic),
		ants.WithPreAlloc(false),
	)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (d *MessageInterfaceImpl) Handle(cInfo *gate.Info, msg *messages.GlideMessage) error {

	if !msg.GetAction().IsInternal() {
		msg.From = cInfo.ID.UID()
	}
	logger.D("handle message: %s", msg)
	err := d.execPool.Submit(func() {
		handled := d.hc.handle(d, cInfo, msg)
		if !handled {
			if !msg.GetAction().IsInternal() {
				r := messages.NewMessage(msg.GetSeq(), messages.ActionNotifyUnknownAction, msg.GetAction())
				_ = d.gate.EnqueueMessage(cInfo.ID, r)
			}
			logger.W("action is not handled: %s", msg.GetAction())
		}
	})
	if err != nil {
		d.OnHandleMessageError(cInfo, msg, err)
		return err
	}
	return nil
}

func (d *MessageInterfaceImpl) AddHandler(i MessageHandler) {
	d.hc.add(i)
}

func (d *MessageInterfaceImpl) SetGate(g gate.Gateway) {
	d.gate = g
}

func (d *MessageInterfaceImpl) SetSubscription(g subscription.Interface) {
	d.subscription = g
}

func (d *MessageInterfaceImpl) SetNotifyErrorOnServer(enable bool) {
	d.notifyOnSrvErr = enable
}

func (d *MessageInterfaceImpl) GetClientInterface() gate.Gateway {
	return d.gate
}

func (d *MessageInterfaceImpl) GetGroupInterface() subscription.Interface {
	return d.subscription
}

func (d *MessageInterfaceImpl) OnHandleMessageError(cInfo *gate.Info, msg *messages.GlideMessage, err error) {
	if d.notifyOnSrvErr {
		_ = d.gate.EnqueueMessage(cInfo.ID, messages.NewMessage(-1, messages.ActionNotifyError, err.Error()))
	}
}

// handlerChain is a chain of MessageHandlers.
type handlerChain struct {
	h    MessageHandler
	next *handlerChain
}

func (hc *handlerChain) add(i MessageHandler) {
	if hc.next == nil {
		hc.next = &handlerChain{
			h: i,
		}
	} else {
		hc.next.add(i)
	}
}

func (hc handlerChain) handle(h2 *MessageInterfaceImpl, cliInfo *gate.Info, message *messages.GlideMessage) bool {
	if hc.h != nil && hc.h.Handle(h2, cliInfo, message) {
		return true
	}
	if hc.next != nil {
		return hc.next.handle(h2, cliInfo, message)
	}
	return false
}
