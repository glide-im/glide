package message_handler

import (
	"github.com/glide-im/glide/pkg/auth"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscription"
)

var _ messaging.Messaging = (*MessageHandler)(nil)

type Options struct {
	// MessageStore chat message store
	MessageStore store.MessageStore

	// OfflineHandleFn client offline, handle message
	OfflineHandleFn func(h *MessageHandler, ci *gate.Info, pushMessage *messages.GlideMessage)

	// Auth used for client auth action handler messages.ActionApiAuth
	Auth auth.Interface

	// DontInitDefaultHandler true will not init default action handler, MessageHandler.InitDefaultHandler
	DontInitDefaultHandler bool

	// NotifyOnErr true express notify client on server error.
	NotifyOnErr bool
}

// MessageHandler .
type MessageHandler struct {
	def   *messaging.MessageInterfaceImpl
	store store.MessageStore

	auth            auth.Interface
	offlineHandleFn func(h *MessageHandler, ci *gate.Info, m *messages.GlideMessage)
}

func NewHandlerWithOptions(opts *Options) (*MessageHandler, error) {
	impl, err := messaging.NewDefaultImpl(opts.MessageStore)
	if err != nil {
		return nil, err
	}
	impl.SetNotifyErrorOnServer(opts.NotifyOnErr)
	ret := &MessageHandler{
		def:             impl,
		store:           opts.MessageStore,
		auth:            opts.Auth,
		offlineHandleFn: opts.OfflineHandleFn,
	}
	if !opts.DontInitDefaultHandler {
		ret.InitDefaultHandler(nil)
	}
	return ret, nil
}

func NewHandler(store store.MessageStore, auth auth.Interface) (*MessageHandler, error) {
	return NewHandlerWithOptions(&Options{
		MessageStore:           store,
		OfflineHandleFn:        nil,
		Auth:                   auth,
		DontInitDefaultHandler: false,
	})
}

// InitDefaultHandler add all default action handler.
// The action and HandlerFunc will pass to callback, the return value of callback will set as action handler, callback
// can be nil.
func (d MessageHandler) InitDefaultHandler(callback func(action messages.Action, fn messaging.HandlerFunc) messaging.HandlerFunc) {
	m := map[messages.Action]messaging.HandlerFunc{
		messages.ActionChatMessage:  d.handleChatMessage,
		messages.ActionGroupMessage: d.handleGroupMsg,
		messages.ActionAckRequest:   d.handleAckRequest,
		messages.ActionHeartbeat:    d.handleHeartbeat,
		messages.ActionAckGroupMsg:  d.handleAckGroupMsgRequest,
		messages.ActionApiAuth:      d.handleAuth,
	}
	for action, handlerFunc := range m {
		if callback != nil {
			handlerFunc = callback(action, handlerFunc)
		}
		d.AddHandler(messaging.NewActionHandler(action, handlerFunc))
	}
}

func (d *MessageHandler) AddHandler(i messaging.MessageHandler) {
	d.def.AddHandler(i)
}

func (d *MessageHandler) SetAuthorize(a auth.Interface) {
	d.auth = a
}

func (d *MessageHandler) Handle(cInfo *gate.Info, msg *messages.GlideMessage) error {
	return d.def.Handle(cInfo, msg)
}

func (d *MessageHandler) SetGate(g gate.Gateway) {
	d.def.SetGate(g)
}

func (d *MessageHandler) SetSubscription(s subscription.Interface) {
	d.def.SetSubscription(s)
}

// SetOfflineMessageHandler called while client is offline
func (d *MessageHandler) SetOfflineMessageHandler(fn func(h *MessageHandler, ci *gate.Info, m *messages.GlideMessage)) {
	d.offlineHandleFn = fn
}

func (d *MessageHandler) dispatchGroupMessage(gid int64, msg *messages.ChatMessage) error {
	return d.def.GetGroupInterface().PublishMessage("", nil)
}

func (d *MessageHandler) enqueueMessage(id gate.ID, message *messages.GlideMessage) {
	err := d.def.GetClientInterface().EnqueueMessage(id, message)
	if err != nil {
		logger.E("%v", err)
	}
}
func (d *MessageHandler) unmarshalData(c *gate.Info, msg *messages.GlideMessage, to interface{}) bool {
	err := msg.Data.Deserialize(to)
	if err != nil {
		logger.E("sender chat senderMsg %v", err)
		return false
	}
	return true
}
