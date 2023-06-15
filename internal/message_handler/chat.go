package message_handler

import (
	messages2 "github.com/glide-im/glide/im_service/messages"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
)

// handleChatMessage 分发用户单聊消息
func (d *MessageHandler) handleChatMessage(c *gate.Info, m *messages.GlideMessage) error {
	msg := new(messages.ChatMessage)
	if !d.unmarshalData(c, m, msg) {
		return nil
	}
	msg.From = c.ID.UID()
	msg.To = m.To

	if msg.Mid == 0 && m.Action != messages.ActionChatMessageResend {
		// 当客户端发送一条 mid 为 0 的消息时表示这条消息未被服务端收到过, 或客户端未收到服务端的确认回执
		err := d.store.StoreMessage(msg)
		if err != nil {
			logger.E("store chat message error %v", err)
			return err
		}
	}
	// sender resend message to receiver, server has already acked it
	// does the server should not ack it again ?
	err := d.ackChatMessage(c, msg)
	if err != nil {
		logger.E("ack chat message error %v", err)
	}

	pushMsg := messages.NewMessage(0, messages2.ActionChatMessage, msg)

	if !d.dispatchAllDevice(msg.To, pushMsg) {
		// receiver offline, send offline message, and ack message
		err := d.ackNotifyMessage(c, msg)
		if err != nil {
			logger.E("ack notify message error %v", err)
		}
		return d.dispatchOffline(c, msg)
	}
	return nil
}

func (d *MessageHandler) handleChatRecallMessage(c *gate.Info, msg *messages.GlideMessage) error {
	return d.handleChatMessage(c, msg)
}

func (d *MessageHandler) ackNotifyMessage(c *gate.Info, m *messages.ChatMessage) error {
	ackNotify := messages.AckNotify{Mid: m.Mid}
	msg := messages.NewMessage(0, messages2.ActionAckNotify, &ackNotify)
	return d.def.GetClientInterface().EnqueueMessage(c.ID, msg)
}

func (d *MessageHandler) ackChatMessage(c *gate.Info, msg *messages.ChatMessage) error {
	ackMsg := messages.AckMessage{
		CliMid: msg.CliMid,
		Mid:    msg.Mid,
		Seq:    0,
	}
	ack := messages.NewMessage(0, messages2.ActionAckMessage, &ackMsg)
	return d.def.GetClientInterface().EnqueueMessage(c.ID, ack)
}

// dispatchOffline 接收者不在线, 离线推送
func (d *MessageHandler) dispatchOffline(c *gate.Info, message *messages.ChatMessage) error {
	logger.D("dispatch offline message %v %v", c.ID, message)
	err := d.store.StoreOffline(message)
	if err != nil {
		logger.E("store chat message error %v", err)
		return err
	}
	return nil
}

// dispatchOnline 接收者在线, 直接投递消息
func (d *MessageHandler) dispatchOnline(c *gate.Info, msg *messages.ChatMessage) error {
	receiverMsg := msg
	msg.From = c.ID.UID()
	dispatchMsg := messages.NewMessage(-1, messages2.ActionChatMessage, receiverMsg)
	return d.def.GetClientInterface().EnqueueMessage(c.ID, dispatchMsg)
}

// TODO optimize 2022-6-20 11:18:24
func (d *MessageHandler) dispatchAllDevice(uid string, m *messages.GlideMessage) bool {
	devices := []string{"", "1", "2", "3"}

	var ok = false
	for _, device := range devices {
		id := gate.NewID("", uid, device)
		err := d.def.GetClientInterface().EnqueueMessage(id, m)
		if err != nil {
			if !gate.IsClientNotExist(err) {
				logger.E("dispatch message error %v", err)
			}
		} else {
			ok = true
		}
	}
	return ok
}
