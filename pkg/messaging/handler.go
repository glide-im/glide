package messaging

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/panjf2000/ants/v2"
)

type HandlerFunc func(cliInfo *gate.Info, message *messages.GlideMessage) error

// Handler default implementation of the messaging interface.
type Handler struct {

	// execPool 100 capacity goroutine pool, 假设每个消息处理需要10ms, 一个协程则每秒能处理100条消息
	execPool *ants.Pool

	// store express message store interface
	store store.MessageStore

	// handlers message handler function map for message action
	handlers map[messages.Action]HandlerFunc

	subscription subscription.Interface
	gate         gate.Manager
}

func NewDefaultImpl(store store.MessageStore) (*Handler, error) {

	if store == nil {
		return nil, errors.New("store is nil")
	}

	ret := Handler{
		store:    store,
		handlers: map[messages.Action]HandlerFunc{},
	}

	var err error
	ret.execPool, err = ants.NewPool(1_0000,
		ants.WithNonblocking(true),
		ants.WithPanicHandler(func(i interface{}) {
			logger.E("message impl panic: %v", i)
		}),
		ants.WithPreAlloc(false),
	)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (d *Handler) Handle(cInfo *gate.Info, msg *messages.GlideMessage) error {

	if cInfo.ID == "" {
		return errors.New("unauthorized")
	}

	logger.D("new message: id=%v", cInfo.ID)
	err := d.execPool.Submit(func() {
		h, ok := d.handlers[messages.Action(msg.GetAction())]
		if ok {
			err := h(cInfo, msg)
			if err != nil {
				logger.E("message impl error: %v", err)
			}
		} else {
			_ = d.gate.EnqueueMessage(cInfo.ID, messages.NewMessage(-1, messages.ActionNotifyError, "unknown action"))
		}
	})
	if err != nil {
		if err == ants.ErrPoolOverload {
			return errors.New("message handle goroutine pool is overload")
		}
		if err == ants.ErrPoolClosed {
			return errors.New("message handle goroutine pool is closed")
		}
		_ = d.gate.EnqueueMessage(cInfo.ID, messages.NewMessage(-1, messages.ActionNotifyError, "internal server error"))
		return errors.New("message handle goroutine pool submit error")
	}
	return nil
}

func (d *Handler) Run() error {
	return errors.New("not implemented")
}

func (d *Handler) SetGate(g gate.Manager) {
	d.gate = g
}

func (d *Handler) SetSubscription(g subscription.Interface) {
	d.subscription = g
}

func (d *Handler) PutMessageHandler(action messages.Action, i HandlerFunc) {
	d.handlers[action] = i
}

func (d *Handler) GetClientInterface() gate.Manager {
	return d.gate
}

func (d *Handler) GetGroupInterface() subscription.Interface {
	return d.subscription
}
