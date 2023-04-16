package message_store_db

import (
	"github.com/glide-im/glide/pkg/subscription"
	"math"
)

type SubscriptionMessageStore struct {
}

func (c *SubscriptionMessageStore) NextSegmentSequence(id subscription.ChanID, info subscription.ChanInfo) (int64, int64, error) {
	return 1, math.MaxInt64, nil
}

func (c *SubscriptionMessageStore) StoreMessage(ch subscription.ChanID, msg subscription.Message) error {
	return nil
}

type IdleSubscriptionStore struct {
}

func (i *IdleSubscriptionStore) NextSegmentSequence(id subscription.ChanID, info subscription.ChanInfo) (int64, int64, error) {
	return 1, math.MaxInt64, nil
}

func (i *IdleSubscriptionStore) StoreMessage(ch subscription.ChanID, msg subscription.Message) error {
	return nil
}
