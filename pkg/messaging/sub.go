package messaging

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/glide-im/glide/pkg/subscription/subscription_impl"
)

// handleGroupMsg 分发群消息
func (d *MessageHandlerImpl) handleGroupMsg(c *gate.Info, msg *messages.GlideMessage) error {

	id := subscription.ChanID(msg.To)

	cm := messages.ChatMessage{}
	e := msg.Data.Deserialize(&cm)
	if e != nil {
		return e
	}

	m := subscription_impl.PublishMessage{
		From:    subscription.SubscriberID(msg.From),
		Message: msg,
		Type:    subscription_impl.TypeMessage,
	}
	err := d.def.GetGroupInterface().PublishMessage(id, &m)

	if err != nil {
		logger.E("dispatch group message error: %v", err)
		notify := messages.NewMessage(msg.GetSeq(), messages.ActionNotifyError, err.Error())
		d.enqueueMessage(c.ID, notify)
	} else {
		_ = d.ackChatMessage(c, &cm)
	}

	return nil
}

func (d *MessageHandlerImpl) handleApiGroupMembers(c *gate.Info, msg *messages.GlideMessage) error {
	//id := subscription.ChanID(msg.To)
	//
	//cm := messages.ChatMessage{}
	//e := msg.Data.Deserialize(&cm)
	//if e != nil {
	//	return e
	//}

	return nil
}

func (d *MessageHandlerImpl) handleAckGroupMsgRequest(c *gate.Info, msg *messages.GlideMessage) error {
	ack := new(messages.AckGroupMessage)
	if !d.unmarshalData(c, msg, ack) {
		return nil
	}
	//err := msgdao.UpdateGroupMemberMsgState(ack.Gid, 0, ack.Mid, ack.Seq)
	//if err != nil {
	//
	//}
	return nil
}
