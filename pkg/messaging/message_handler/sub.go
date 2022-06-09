package message_handler

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
	"strconv"
)

// handleGroupMsg 分发群消息
func (d *MessageHandler) handleGroupMsg(c *gate.Info, msg *messages.GlideMessage) error {

	groupMsg := new(messages.ChatMessage)
	if !d.unwrap(c, msg, groupMsg) {
		return nil
	}
	groupMsg.From = c.ID.UID()

	var err error

	id := subscription.ChanID(strconv.FormatInt(groupMsg.To, 10))
	err = d.def.GetGroupInterface().PublishMessage(id, groupMsg)

	if err != nil {
		logger.E("dispatch group message error: %v", err)
		notify := messages.NewMessage(0, messages.ActionMessageFailed, messages.AckNotify{Mid: groupMsg.Mid})
		d.enqueueMessage(c.ID, notify)
	}

	return nil
}

func (d *MessageHandler) handleAckGroupMsgRequest(c *gate.Info, msg *messages.GlideMessage) error {
	ack := new(messages.AckGroupMessage)
	if !d.unwrap(c, msg, ack) {
		return nil
	}
	//err := msgdao.UpdateGroupMemberMsgState(ack.Gid, 0, ack.Mid, ack.Seq)
	//if err != nil {
	//
	//}
	return nil
}
