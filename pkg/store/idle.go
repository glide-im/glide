package store

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

type IdleMessageStore struct {
}

func (i *IdleMessageStore) StoreMessage(gate.ID, *messages.ChatMessage) error {
	return nil
}
