package store_kafka

import (
	"github.com/glide-im/glide/pkg/messages"
)

type MessageStore struct {
}

func (D *MessageStore) StoreMessage(message *messages.ChatMessage) error {

	return nil
}
