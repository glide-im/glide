package store

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscribe"
)

type SubscriptionStore interface {
	StoreMessage(ch subscribe.ChanID, msg *messages.ChatMessage) error
	StoreSeq(ch subscribe.ChanID, seq int64) error
}
