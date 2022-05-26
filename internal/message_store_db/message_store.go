package message_store_db

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

type MessageStore struct {
}

func New() *MessageStore {
	return &MessageStore{}
}

func (D *MessageStore) StoreMessage(from gate.ID, message *messages.ChatMessage) error {

	return nil
}
