package message_store

import (
	"github.com/glide-im/glide/pkg/client"
	"github.com/glide-im/glide/pkg/messages"
)

// MessageStore is a store for messages, used to store chat messages in messaging.Interface, its many be called multiple times,
// but only the last updates will be stored.
type MessageStore interface {

	// StoreChatMessage stores chat message to database
	StoreChatMessage(from client.ID, message *messages.ChatMessage) error

	// StoreChatMessageRecalled update existing chat message to recalled, if exists.
	StoreChatMessageRecalled(mid int64, recallBy int64) error
}
