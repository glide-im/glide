package message_store_db

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
	"math"
)

type SubscriptionMessageStore struct {
}

func (c *SubscriptionMessageStore) NextSegmentSequence(id subscription.ChanID, info subscription.ChanInfo) (int64, int64, error) {
	return 1, math.MaxInt64, nil
}

func (c *SubscriptionMessageStore) StoreChannelMessage(ch subscription.ChanID, msg *messages.ChatMessage) error {
	return nil
}

type IdleSubscriptionStore struct {
}

func (i *IdleSubscriptionStore) NextSegmentSequence(id subscription.ChanID, info subscription.ChanInfo) (int64, int64, error) {
	return 1, math.MaxInt64, nil
}

func (i *IdleSubscriptionStore) StoreChannelMessage(ch subscription.ChanID, msg *messages.ChatMessage) error {
	return nil
}
