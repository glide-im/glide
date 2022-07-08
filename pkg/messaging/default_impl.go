package messaging

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/panjf2000/ants/v2"
)

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
