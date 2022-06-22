package message_handler

import (
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
	msg.From = m.From
	msg.To = m.To

	if m.GetAction() != messages.ActionChatMessageResend {
		err := d.store.StoreMessage(msg)
		if err != nil {
			logger.E("store chat message error %v", err)
			return err
		}
	}

	// sender resend message to receiver, server has already acked it
	// does the server should not ack it again ?
	if m.GetAction() != messages.ActionChatMessageResend {
		err := d.ackChatMessage(c, msg.Mid)
		if err != nil {
			logger.E("ack chat message error %v", err)
		}
	}

	pushMsg := messages.NewMessage(0, messages.ActionChatMessage, msg)

	if !d.dispatchAllDevice(msg.To, pushMsg) {
		// receiver offline, send offline message, and ack message
		err := d.ackNotifyMessage(c, msg.Mid)
		if err != nil {
			logger.E("ack notify message error %v", err)
		}
		return d.dispatchOffline(c, m)
	}
	return nil
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
	if d.offlineHandleFn != nil {
		d.offlineHandleFn(d, c, message)
	}
	return nil
}

// dispatchOnline 接收者在线, 直接投递消息
func (d *MessageHandler) dispatchOnline(c *gate.Info, msg *messages.ChatMessage) error {
	receiverMsg := msg
	msg.From = c.ID.UID()
	dispatchMsg := messages.NewMessage(-1, messages.ActionChatMessage, receiverMsg)
	return d.def.GetClientInterface().EnqueueMessage(c.ID, dispatchMsg)
}

// TODO optimize 2022-6-20 11:18:24
func (d *MessageHandler) dispatchAllDevice(uid string, m *messages.GlideMessage) bool {
	devices := []string{"", "1", "2", "3"}
	ok := false
	for _, device := range devices {
		id := gate.NewID("", uid, device)
		if d.def.GetClientInterface().IsOnline(id) {
			err := d.def.GetClientInterface().EnqueueMessage(id, m)
			if err != nil {
				logger.E("dispatch message error %v", err)
			} else {
				ok = true
			}
		}
	}
	return ok
}
