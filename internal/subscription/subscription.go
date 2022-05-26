package subscription

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscription"
)

type impl struct {
	gate  gate.Interface
	store store.SubscriptionStore
}

func NewSubscription() subscription.Interface {
	return &impl{}
}

func (s *impl) UpdateSubscriber(id subscription.ChanID, updates []subscription.SubscriberUpdate) error {
	return nil
}

func (s *impl) UpdateChannel(id subscription.ChanID, update subscription.ChannelUpdate) error {
	return nil
}

func (s *impl) PublishMessage(id subscription.ChanID, message subscription.Message) error {
	return nil
}

func (s *impl) SetGateInterface(q gate.Interface) {
	s.gate = q
}

func NewServer() (*impl, error) {
	return &impl{}, nil
}
