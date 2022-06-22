package message_handler

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

func (d *MessageHandler) handleHeartbeat(cInfo *gate.Info, msg *messages.GlideMessage) error {
	return nil
}

// handleAckRequest 处理接收者收到消息发回来的确认消息
func (d *MessageHandler) handleAckRequest(c *gate.Info, msg *messages.GlideMessage) error {
	ackMsg := new(messages.AckRequest)
	if !d.unmarshalData(c, msg, ackMsg) {
		return nil
	}
	ackNotify := messages.NewMessage(0, messages.ActionAckNotify, ackMsg)

	// 通知发送者, 对方已收到消息
	d.dispatchAllDevice(ackMsg.From, ackNotify)
	return nil
}

func (d *MessageHandler) handleClientCustom(c *gate.Info, msg *messages.GlideMessage) error {
	m := new(messages.ClientCustom)
	if !d.unmarshalData(c, msg, m) {
		return nil
	}
	m2 := messages.NewMessage(0, messages.ActionClientCustom, m)
	d.dispatchAllDevice(msg.To, m2)
	return nil
}
