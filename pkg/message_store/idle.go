package message_store

import (
	"github.com/glide-im/glide/pkg/client"
	"github.com/glide-im/glide/pkg/messages"
)

type IdleMessageStore struct {
}

func (i *IdleMessageStore) StoreChatMessage(client.ID, *messages.ChatMessage) error {
	return nil
}

func (i *IdleMessageStore) StoreChatMessageRecalled(int64, int64) error {
	return nil
}
