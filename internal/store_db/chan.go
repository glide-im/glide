package store_db

import (
	"github.com/glide-im/glide/pkg/subscription"
	"math"
)

type SubscriptionStore struct {
}

func (c *SubscriptionStore) NextSegmentSequence(id subscription.ChanID, info subscription.ChanInfo) (int64, int64, error) {
	return 1, math.MaxInt64, nil
}

func (c *SubscriptionStore) StoreMessage(ch subscription.ChanID, msg subscription.Message) error {
	// TODO: implement 2022-6-15 16:37:10
	return nil
}
