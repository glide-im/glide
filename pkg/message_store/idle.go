package message_store

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

type IdleMessageStore struct {
}

func (i *IdleMessageStore) StoreChatMessage(gate.ID, *messages.ChatMessage) error {
	return nil
}

func (i *IdleMessageStore) StoreChatMessageRecalled(int64, int64) error {
	return nil
}
