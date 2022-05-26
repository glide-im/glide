package store

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
)

type SubscriptionStore interface {
	StoreMessage(ch subscription.ChanID, msg *messages.ChatMessage) error
	StoreSeq(ch subscription.ChanID, seq int64) error
}
