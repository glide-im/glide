package subscribe

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscribe"
)

type impl struct {
	gate  gate.Interface
	store store.SubscriptionStore
}

func NewSubscription() subscribe.Interface {
	return &impl{}
}

func (s *impl) UpdateSubscriber(id subscribe.ChanID, updates []subscribe.SubscriberUpdate) error {
	return nil
}

func (s *impl) UpdateChannel(id subscribe.ChanID, update subscribe.ChannelUpdate) error {
	return nil
}

func (s *impl) PublishMessage(id subscribe.ChanID, message subscribe.Message) error {
	return nil
}

func (s *impl) SetGateInterface(q gate.Interface) {
	s.gate = q
}

func NewServer() (*impl, error) {
	return &impl{}, nil
}
