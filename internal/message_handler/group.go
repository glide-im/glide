package message_handler

import (
	"github.com/glide-im/glide/pkg/client"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
)

// handleGroupMsg 分发群消息
func (d *MessageHandler) handleGroupMsg(c *client.Info, msg *messages.GlideMessage) error {

	groupMsg := new(messages.ChatMessage)
	if !d.unwrap(c, msg, groupMsg) {
		return nil
	}
	groupMsg.From = c.ID.UID()

	var err error
	if msg.GetAction() == messages.ActionGroupMessageRecall {
		err = d.dispatchRecallMessage(groupMsg.To, groupMsg)
	} else {
		err = d.dispatchGroupMessage(groupMsg.To, groupMsg)
	}
	if err != nil {
		logger.E("dispatch group message error: %v", err)
		notify := messages.NewMessage(0, messages.ActionMessageFailed, messages.NewAckNotify(groupMsg.Mid))
		d.enqueueMessage(c.ID, notify)
	}

	return nil
}

func (d *MessageHandler) handleGroupRecallMsg(c *client.Info, msg *messages.GlideMessage) error {
	return d.handleGroupMsg(c, msg)
}

func (d *MessageHandler) handleAckGroupMsgRequest(c *client.Info, msg *messages.GlideMessage) error {
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
