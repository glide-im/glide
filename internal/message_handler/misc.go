package message_handler

import (
	"github.com/glide-im/glide/pkg/client"
	"github.com/glide-im/glide/pkg/messages"
)

func (d *MessageHandler) handleHeartbeat(cInfo *client.Info, msg *messages.GlideMessage) error {
	return nil
}

// handleAckRequest 处理接收者收到消息发回来的确认消息
func (d *MessageHandler) handleAckRequest(c *client.Info, msg *messages.GlideMessage) error {
	ackMsg := new(messages.AckRequest)
	if !d.unwrap(c, msg, ackMsg) {
		return nil
	}
	ackNotify := messages.NewMessage(0, messages.ActionAckNotify, ackMsg)
	// 通知发送者, 对方已收到消息
	d.enqueueMessage("", ackNotify)
	return nil
}

func (d *MessageHandler) handleClientCustom(c *client.Info, msg *messages.GlideMessage) error {
	m := new(messages.ClientCustom)
	if !d.unwrap(c, msg, m) {
		return nil
	}
	m2 := messages.NewMessage(0, messages.ActionClientCustom, m)
	d.enqueueMessage("", m2)
	return nil
}
