package message_handler

import (
	messages2 "github.com/glide-im/glide/im_service/messages"
	"github.com/glide-im/glide/internal/pkg/db"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/logger"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/messaging"
	"time"
)

const (
	KeyRedisOfflineMsgPrefix = "im:msg:offline:"
)

var Enable = false

func GetHandleFn() func(h *MessageHandler, ci *gate.Info, m *messages.GlideMessage) {
	return handler
}

func handler(_ *MessageHandler, _ *gate.Info, m *messages.GlideMessage) {
	if !Enable {
		return
	}
	if m.GetAction() == messages2.ActionChatMessage || m.GetAction() == messages2.ActionChatMessageResend {
		c := messages.ChatMessage{}
		err := m.Data.Deserialize(&c)
		if err != nil {
			logger.E("deserialize chat message error: %v", err)
			return
		}
		bytes, err := messages.JsonCodec.Encode(m)
		if err != nil {
			logger.E("deserialize chat message error: %v", err)
			return
		}
		storeOfflineMessage(m.To, string(bytes))
	}
}

func storeOfflineMessage(to string, msg string) {
	key := KeyRedisOfflineMsgPrefix + to
	db.Redis.SAdd(key, msg)
	// TODO 2022-6-22 16:56:57 do not reset expire on new offline message arrived
	// use fixed time segment save offline msg reset segment only.
	db.Redis.Expire(key, time.Hour*24*2)
}

func PushOfflineMessage(h *messaging.MessageInterfaceImpl, id string) {
	key := KeyRedisOfflineMsgPrefix + id
	members, err := db.Redis.SMembers(key).Result()
	if err != nil {
		logger.ErrE("push offline msg error", err)
		return
	}
	for _, member := range members {
		msg := messages.NewEmptyMessage()
		err := messages.JsonCodec.Decode([]byte(member), msg)
		if err != nil {
			logger.ErrE("deserialize redis offline msg error", err)
			continue
		}
		id2 := gate.NewID2(id)
		_ = h.GetClientInterface().EnqueueMessage(id2, msg)
	}

	// AckOfflineMessage(id)
}

func AckOfflineMessage(id string) {
	key := KeyRedisOfflineMsgPrefix + id
	result, err := db.Redis.Del(key).Result()
	if err != nil {
		logger.ErrE("remove offline message error", err)
	}
	logger.I("user %s ack %d offline messages", id, result)
}
