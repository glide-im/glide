package store

import (
	"github.com/glide-im/glide/pkg/subscription"
)

type SubscriptionStore interface {

	// StoreMessage stores a published message.
	StoreMessage(ch subscription.ChanID, msg subscription.Message) error
}
