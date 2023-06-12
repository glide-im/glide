package store

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
)

// MessageStore is a store for messages, used to store chat messages in messaging.Interface, its many be called multiple times,
// but only the last updates will be stored.
type MessageStore interface {

	// StoreMessage stores chat message to database
	StoreMessage(message *messages.ChatMessage) error

	StoreOffline(message *messages.ChatMessage) error
}

type SubscriptionStore interface {

	// NextSegmentSequence return the next segment of specified channel, and segment length.
	NextSegmentSequence(id subscription.ChanID, info subscription.ChanInfo) (int64, int64, error)

	// StoreChannelMessage stores a published message.
	StoreChannelMessage(ch subscription.ChanID, msg subscription.Message) error
}
