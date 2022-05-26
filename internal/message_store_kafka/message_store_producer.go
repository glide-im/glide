package message_store_kafka

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

type MessageStore struct {
}

func (D *MessageStore) StoreMessage(from gate.ID, message *messages.ChatMessage) error {

	return nil
}
