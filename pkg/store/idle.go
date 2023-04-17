package store

import (
	"github.com/glide-im/glide/pkg/messages"
)

var _ MessageStore = &IdleMessageStore{}

type IdleMessageStore struct {
}

func (i *IdleMessageStore) StoreOffline(message *messages.ChatMessage) error {
	return nil
}

func (i *IdleMessageStore) StoreMessage(*messages.ChatMessage) error {
	return nil
}
