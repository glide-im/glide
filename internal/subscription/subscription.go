package subscription

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
)

type impl struct {
}

func NewSubscription() subscription.Interface {
	return &impl{}
}

func (s impl) UpdateSubscriber(id subscription.ChanID, updates []subscription.SubscriberUpdate) error {
	return nil
}

func (s impl) UpdateChannel(id subscription.ChanID, update subscription.ChannelUpdate) error {
	return nil
}

func (s impl) DispatchNotifyMessage(id subscription.ChanID, message *messages.GroupNotify) error {
	return nil
}

func (s impl) PublishMessage(id subscription.ChanID, action messages.Action, message *messages.ChatMessage) error {
	return nil
}

func NewServer() (*impl, error) {
	return &impl{}, nil
}
