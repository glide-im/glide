package message_store_kafka

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

type MessageStore struct {
}

func (D *MessageStore) StoreChatMessage(from gate.ID, message *messages.ChatMessage) error {

	return nil
}

func (D *MessageStore) StoreChatMessageRecalled(mid int64, recallBy int64) error {

	return nil
}
