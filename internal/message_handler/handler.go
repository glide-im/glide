package message_handler

import (
	"github.com/glide-im/glide/pkg/client"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
)

type MessageHandler struct {
	def *messaging.Handler
}

func NewHandler(defaultImpl *messaging.Handler) (*MessageHandler, error) {
	ret := &MessageHandler{
		def: defaultImpl,
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

func (d *MessageHandler) Handle(cInfo *client.Info, msg *messages.GlideMessage) error {
	return d.def.Handle(cInfo, msg)
}

func (d *MessageHandler) PutMessageHandler(action messages.Action, i client.MessageHandler) {
	d.def.PutMessageHandler(action, i)
}

func (d *MessageHandler) dispatchGroupMessage(gid int64, msg *messages.ChatMessage) error {
	return d.def.GetGroupInterface().DispatchMessage(gid, messages.ActionChatMessage, msg)
}

func (d *MessageHandler) dispatchRecallMessage(gid int64, msg *messages.ChatMessage) error {
	return d.def.GetGroupInterface().DispatchMessage(gid, messages.ActionGroupMessageRecall, msg)
}

func (d *MessageHandler) enqueueMessage(id client.ID, message *messages.GlideMessage) {
	err := d.def.GetClientInterface().EnqueueMessage(id, message)
	if err != nil {
		logger.E("%v", err)
	}
}
func (d *MessageHandler) unwrap(c *client.Info, msg *messages.GlideMessage, to interface{}) bool {
	err := msg.DeserializeData(to)
	if err != nil {
		logger.E("sender chat senderMsg %v", err)
		return false
	}
	return true
}
