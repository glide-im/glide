package subscription_impl

import "github.com/glide-im/glide/pkg/subscription"

// SubscribeWrap the wrapper for subscription.Subscribe implementation, for convenience.
type SubscribeWrap interface {
	Subscribe(ch subscription.ChanID, id subscription.SubscriberID, extra interface{}) error

	UnSubscribe(ch subscription.ChanID, id subscription.SubscriberID) error

	UpdateSubscriber(ch subscription.ChanID, id subscription.SubscriberID, extra interface{}) error

	RemoveChannel(ch subscription.ChanID) error

	CreateChannel(ch subscription.ChanID, update *subscription.ChanInfo) error

	UpdateChannel(ch subscription.ChanID, update *subscription.ChanInfo) error

	Publish(ch subscription.ChanID, msg subscription.Message) error
}

func NewSubscribeWrap(subscribe subscription.Subscribe) SubscribeWrap {
	return &wrap{
		fac: subscribe,
	}
}

var _ SubscribeWrap = (*wrap)(nil)

type wrap struct {
	fac subscription.Subscribe
}

func (w *wrap) Subscribe(ch subscription.ChanID, id subscription.SubscriberID, extra interface{}) error {
	return w.fac.UpdateSubscriber(ch, []subscription.Update{
		{
			Flag:  subscription.SubscriberSubscribe,
			ID:    id,
			Extra: extra,
		},
	})
}

func (w *wrap) UnSubscribe(ch subscription.ChanID, id subscription.SubscriberID) error {
	return w.fac.UpdateSubscriber(ch, []subscription.Update{
		{
			Flag:  subscription.SubscriberUnsubscribe,
			ID:    id,
			Extra: nil,
		},
	})
}

func (w *wrap) UpdateSubscriber(ch subscription.ChanID, id subscription.SubscriberID, extra interface{}) error {
	return w.fac.UpdateSubscriber(ch, []subscription.Update{
		{
			Flag:  subscription.SubscriberUpdate,
			ID:    id,
			Extra: extra,
		},
	})
}

func (w *wrap) RemoveChannel(ch subscription.ChanID) error {
	return w.fac.UpdateChannel(ch, subscription.ChannelUpdate{
		Flag:  subscription.ChanDelete,
		Extra: nil,
	})
}

func (w *wrap) CreateChannel(ch subscription.ChanID, update *subscription.ChanInfo) error {
	return w.fac.UpdateChannel(ch, subscription.ChannelUpdate{
		Flag:  subscription.ChanCreate,
		Extra: update,
	})
}

func (w *wrap) UpdateChannel(ch subscription.ChanID, update *subscription.ChanInfo) error {
	return w.fac.UpdateChannel(ch, subscription.ChannelUpdate{
		Flag:  subscription.ChanUpdate,
		Extra: update,
	})
}

func (w *wrap) Publish(ch subscription.ChanID, msg subscription.Message) error {
	return w.fac.PublishMessage(ch, msg)
}
