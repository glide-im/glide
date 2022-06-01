package store

import (
	"github.com/glide-im/glide/pkg/messages"
)

type IdleMessageStore struct {
}

func (i *IdleMessageStore) StoreMessage(*messages.ChatMessage) error {
	return nil
}
