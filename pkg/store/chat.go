package store

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
)

// MessageStore is a store for messages, used to store chat messages in messaging.Interface, its many be called multiple times,
// but only the last updates will be stored.
type MessageStore interface {

	// StoreMessage stores chat message to database
	StoreMessage(from gate.ID, message *messages.ChatMessage) error
}
