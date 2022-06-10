package message_handler

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"strconv"
)

// handleChatMessage 分发用户单聊消息
func (d *MessageHandler) handleChatMessage(c *gate.Info, m *messages.GlideMessage) error {
	msg := new(messages.ChatMessage)
	if !d.unwrap(c, m, msg) {
		return nil
	}
	from, err := strconv.ParseInt(c.ID.UID(), 10, 64)
	if err != nil {
		return err
	}
	msg.From = from

	// 保存消息
	if m.GetAction() != messages.ActionChatMessageResend {
		err := d.store.StoreMessage(msg)
		if err != nil {
			logger.E("save chat message error %v", err)
			return err
		}
	}

	// 告诉客户端服务端已收到
	_ = d.ackChatMessage(c, msg.Mid)

	// 对方不在线, 下发确认包
	if !d.def.GetClientInterface().IsOnline(gate.NewID2(msg.To)) {
		_ = d.ackNotifyMessage(c, msg.Mid)
		return d.dispatchOffline(c, m)
	} else {
		return d.dispatchOnline(c, msg)
	}
}

func (d *MessageHandler) handleChatRecallMessage(c *gate.Info, msg *messages.GlideMessage) error {
	return d.handleChatMessage(c, msg)
}

func (d *MessageHandler) ackNotifyMessage(c *gate.Info, mid int64) error {
	ackNotify := messages.AckNotify{Mid: mid}
	msg := messages.NewMessage(0, messages.ActionAckNotify, &ackNotify)
	return d.def.GetClientInterface().EnqueueMessage(c.ID, msg)
}

func (d *MessageHandler) ackChatMessage(c *gate.Info, mid int64) error {
	ackMsg := messages.AckMessage{
		Mid: mid,
		Seq: 0,
	}
	ack := messages.NewMessage(0, messages.ActionAckMessage, &ackMsg)
	return d.def.GetClientInterface().EnqueueMessage(c.ID, ack)
}

// dispatchOffline 接收者不在线, 离线推送
func (d *MessageHandler) dispatchOffline(c *gate.Info, message *messages.GlideMessage) error {
	logger.D("dispatch offline message %v %v", c.ID, message)
	return nil
}

// dispatchOnline 接收者在线, 直接投递消息
func (d *MessageHandler) dispatchOnline(c *gate.Info, msg *messages.ChatMessage) error {
	receiverMsg := msg
	from, err := strconv.ParseInt(c.ID.UID(), 10, 64)
	if err != nil {
		return err
	}
	msg.From = from
	dispatchMsg := messages.NewMessage(-1, messages.ActionChatMessage, receiverMsg)
	return d.def.GetClientInterface().EnqueueMessage(c.ID, dispatchMsg)
}
