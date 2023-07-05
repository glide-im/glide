package messaging

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscription"
)

var _ Messaging = (*MessageHandlerImpl)(nil)

type MessageHandlerOptions struct {
	// MessageStore chat message store
	MessageStore store.MessageStore

	// DontInitDefaultHandler true will not init default action offlineMessageHandler, MessageHandlerImpl.InitDefaultHandler
	DontInitDefaultHandler bool

	// NotifyOnErr true express notify client on server error.
	NotifyOnErr bool
}

// MessageHandlerImpl .
type MessageHandlerImpl struct {
	def   *MessageInterfaceImpl
	store store.MessageStore

	userState *UserState
}

func NewHandlerWithOptions(gateway gate.Gateway, opts *MessageHandlerOptions) (*MessageHandlerImpl, error) {
	impl, err := NewDefaultImpl(&Options{
		NotifyServerError:     true,
		MaxMessageConcurrency: 10_0000,
	})
	if err != nil {
		return nil, err
	}
	impl.SetNotifyErrorOnServer(opts.NotifyOnErr)

	ret := &MessageHandlerImpl{
		def:       impl,
		store:     opts.MessageStore,
		userState: NewUserState(gateway),
	}
	if !opts.DontInitDefaultHandler {
		ret.InitDefaultHandler(nil)
	}
	return ret, nil
}

// InitDefaultHandler
// 初始化 message.Action 对应的默认 Handler, 部分类型的 Action 才有默认 Handler, 若要修改特定 Action 的默认 Handler 则可以在
// callback 回调中返回你需要的即可, callback 参数 fn 既是该 action 对的默认 Handler.
func (d *MessageHandlerImpl) InitDefaultHandler(callback func(action messages.Action, fn HandlerFunc) HandlerFunc) {

	m := map[messages.Action]HandlerFunc{
		messages.ActionChatMessage:     d.handleChatMessage,
		messages.ActionGroupMessage:    d.handleGroupMsg,
		messages.ActionApiGroupMembers: d.handleApiGroupMembers,
		messages.ActionAckRequest:      d.handleAckRequest,
		messages.ActionAckGroupMsg:     d.handleAckGroupMsgRequest,
		messages.AckOffline:            d.handleAckOffline,
		messages.ActionHeartbeat:       d.handleHeartbeat,
		messages.ActionInternalOnline:  d.handleInternalOnline,
		messages.ActionInternalOffline: d.handleInternalOffline,
		messages.ActionApiSubUserState: d.userState.subUserStateApi,
	}
	for action, handlerFunc := range m {
		if callback != nil {
			handlerFunc = callback(action, handlerFunc)
		}
		d.def.AddHandler(NewActionHandler(action, handlerFunc))
	}

	d.def.AddHandler(&ClientCustomMessageHandler{})
	d.def.AddHandler(NewActionHandler(messages.ActionHeartbeat, handleHeartbeat))
}

func (d *MessageHandlerImpl) AddHandler(i MessageHandler) {
	d.def.AddHandler(i)
}

func (d *MessageHandlerImpl) Handle(cInfo *gate.Info, msg *messages.GlideMessage) error {
	return d.def.Handle(cInfo, msg)
}

func (d *MessageHandlerImpl) SetGate(g gate.Gateway) {
	d.def.SetGate(g)
}

func (d *MessageHandlerImpl) SetSubscription(s subscription.Interface) {
	d.def.SetSubscription(s)
}

func (d *MessageHandlerImpl) dispatchGroupMessage(gid int64, msg *messages.ChatMessage) error {
	return d.def.GetGroupInterface().PublishMessage("", nil)
}

func (d *MessageHandlerImpl) enqueueMessage(id gate.ID, message *messages.GlideMessage) {
	err := d.def.GetClientInterface().EnqueueMessage(id, message)
	if err != nil {
		logger.E("%v", err)
	}
}
func (d *MessageHandlerImpl) unmarshalData(c *gate.Info, msg *messages.GlideMessage, to interface{}) bool {
	err := msg.Data.Deserialize(to)
	if err != nil {
		logger.E("sender chat senderMsg %v", err)
		return false
	}
	return true
}

func dispatch2AllDevice(h *MessageInterfaceImpl, uid string, m *messages.GlideMessage) bool {
	devices := []string{"", "1", "2", "3"}
	for _, device := range devices {
		id := gate.NewID("", uid, device)
		err := h.GetClientInterface().EnqueueMessage(id, m)
		if err != nil && !gate.IsClientNotExist(err) {
			logger.E("dispatch message error %v", err)
		}
	}
	return true
}
