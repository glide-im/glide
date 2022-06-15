package store_kafka

import (
	"github.com/glide-im/glide/pkg/subscription"
)

type SubscriptionMessageStore struct {
}

func (c *SubscriptionMessageStore) StoreMessage(ch subscription.ChanID, msg subscription.Message) error {
	return nil
}
