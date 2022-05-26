package message_handler

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"github.com/glide-im/glide/pkg/store"
)

type MessageHandler struct {
	def   *messaging.Handler
	store store.MessageStore
}

func NewHandler(store store.MessageStore) (*MessageHandler, error) {

	impl, err := messaging.NewDefaultImpl(store)
	if err != nil {
		return nil, err
	}

	ret := &MessageHandler{
		def:   impl,
		store: store,
	}
	ret.PutMessageHandler(messages.ActionChatMessage, ret.handleChatMessage)
	ret.PutMessageHandler(messages.ActionGroupMessage, ret.handleGroupMsg)
	ret.PutMessageHandler(messages.ActionAckRequest, ret.handleAckRequest)
	ret.PutMessageHandler(messages.ActionHeartbeat, ret.handleHeartbeat)
	ret.PutMessageHandler(messages.ActionClientCustom, ret.handleClientCustom)
	ret.PutMessageHandler(messages.ActionGroupMessageRecall, ret.handleGroupRecallMsg)
	ret.PutMessageHandler(messages.ActionAckGroupMsg, ret.handleAckGroupMsgRequest)
	ret.PutMessageHandler(messages.ActionApiAuth, ret.handleAuth)
	return ret, nil
}

func (d *MessageHandler) Handle(cInfo *gate.Info, msg *messages.GlideMessage) error {
	return d.def.Handle(cInfo, msg)
}

func (d *MessageHandler) PutMessageHandler(action messages.Action, i messaging.HandlerFunc) {
	d.def.PutMessageHandler(action, i)
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
func (d *MessageHandler) unwrap(c *gate.Info, msg *messages.GlideMessage, to interface{}) bool {
	err := msg.DeserializeData(to)
	if err != nil {
		logger.E("sender chat senderMsg %v", err)
		return false
	}
	return true
}
