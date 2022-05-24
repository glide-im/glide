package messaging

import (
	"errors"
	"github.com/glide-im/glide/pkg/client"
	"github.com/glide-im/glide/pkg/group"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/message_store"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/panjf2000/ants/v2"
)

// Handler default implementation of the messaging interface.
type Handler struct {

	// execPool 100 capacity goroutine pool, 假设每个消息处理需要10ms, 一个协程则每秒能处理100条消息
	execPool *ants.Pool

	// store express message store interface
	store message_store.MessageStore

	// handlers message handler function map for message action
	handlers map[messages.Action]client.MessageHandler

	group  group.Interface
	client client.Interface
}

func NewDefaultImpl(store message_store.MessageStore) (Interface, error) {

	if store == nil {
		return nil, errors.New("store is nil")
	}

	ret := Handler{
		store:    store,
		handlers: map[messages.Action]client.MessageHandler{},
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

func (d *Handler) Handle(cInfo *client.Info, msg *messages.GlideMessage) error {

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
			_ = d.client.EnqueueMessage(cInfo.ID, messages.NewMessage(-1, messages.ActionNotifyError, "unknown action"))
		}
	})
	if err != nil {
		if err == ants.ErrPoolOverload {
			return errors.New("message handle goroutine pool is overload")
		}
		if err == ants.ErrPoolClosed {
			return errors.New("message handle goroutine pool is closed")
		}
		_ = d.client.EnqueueMessage(cInfo.ID, messages.NewMessage(-1, messages.ActionNotifyError, "internal server error"))
		return errors.New("message handle goroutine pool submit error")
	}
	return nil
}

func (d *Handler) PutMessageHandler(action messages.Action, i client.MessageHandler) {
	d.handlers[action] = i
}

func (d *Handler) GetClientInterface() client.Interface {
	return d.client
}

func (d *Handler) GetGroupInterface() group.Interface {
	return d.group
}
