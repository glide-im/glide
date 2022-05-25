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
	from := c.ID.UID()
	msg.From = from

	if m.GetAction() != messages.ActionChatMessageResend {
		if m.GetAction() == messages.ActionChatMessageRecall {
			r := &messages.Recall{}
			err := messages.DefaultCodec.Decode([]byte(msg.Content), r)
			if err != nil || r.RecallBy != from {
				return err
			}
			err = d.store.StoreChatMessageRecalled(r.Mid, r.RecallBy)

			//err = msgdao.ChatMsgDaoImpl.UpdateChatMessageStatus(r.Mid, r.RecallBy, msg.To, msgdao.ChatMessageStatusRecalled)
			if err != nil {
				logger.E("update message status error %v", err)
				return err
			}
		} else {
			lg := from
			sm := msg.To
			if lg < sm {
				lg, sm = sm, lg
			}
			sessionId := strconv.FormatInt(lg, 10) + "_" + strconv.FormatInt(sm, 10)
			logger.D("sessionId:%s", sessionId)
			//dbMsg := msgdao.ChatMessage{
			//	MID:       msg.Mid,
			//	From:      from,
			//	To:        msg.To,
			//	Type:      msg.Type,
			//	SendAt:    msg.SendAt,
			//	Content:   msg.Content,
			//	CliSeq:    msg.Seq,
			//	SessionID: sessionId,
			//}
			//_, err := msgdao.AddChatMessage(&dbMsg)

			// 保存消息
			err := d.store.StoreChatMessage(c.ID, msg)
			if err != nil {
				logger.E("save chat message error %v", err)
				return err
			}
		}
	}

	// 告诉客户端服务端已收到
	_ = d.ackChatMessage(c, msg.Mid)

	// 对方不在线, 下发确认包
	if !d.client.IsOnline(gate.NewID2(msg.To)) {
		_ = d.ackNotifyMessage(c, msg.Mid)
		//err := msgdao.AddOfflineMessage(msg.To, msg.Mid)
		//if err != nil {
		//	logger.E("save offline message error %v", err)
		//}
		return d.dispatchOffline(c, m)
	} else {
		return d.dispatchOnline(c, msg)
	}
}

func (d *MessageHandler) handleChatRecallMessage(c *gate.Info, msg *messages.GlideMessage) error {
	return d.handleChatMessage(c, msg)
}

func (d *MessageHandler) ackNotifyMessage(c *gate.Info, mid int64) error {
	ackNotify := messages.NewAckNotify(mid)
	msg := messages.NewMessage(0, messages.ActionAckNotify, &ackNotify)
	return d.client.EnqueueMessage(c.ID, msg)
}

func (d *MessageHandler) ackChatMessage(c *gate.Info, mid int64) error {
	ackMsg := messages.NewAckMessage(mid, 0)
	ack := messages.NewMessage(0, messages.ActionAckMessage, &ackMsg)
	return d.client.EnqueueMessage(c.ID, ack)
}

// dispatchOffline 接收者不在线, 离线推送
func (d *MessageHandler) dispatchOffline(c *gate.Info, message *messages.GlideMessage) error {
	logger.D("dispatch offline message %v %v", c.ID, message)
	return nil
}

// dispatchOnline 接收者在线, 直接投递消息
func (d *MessageHandler) dispatchOnline(c *gate.Info, msg *messages.ChatMessage) error {
	receiverMsg := msg
	msg.From = c.ID.UID()
	dispatchMsg := messages.NewMessage(-1, messages.ActionChatMessage, receiverMsg)
	return d.client.EnqueueMessage(c.ID, dispatchMsg)
}
