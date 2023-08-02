package subscription_impl

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscription"
	"sync"
)

var _ subscription.Subscribe = (*subscriptionImpl)(nil)

type subscriptionImpl struct {
	unwrap *realSubscription
}

func NewSubscription(store store.SubscriptionStore, seqStore ChannelSequenceStore) subscription.Subscribe {
	return &subscriptionImpl{
		unwrap: newRealSubscription(store, seqStore),
	}
}

func (s *subscriptionImpl) UpdateSubscriber(id subscription.ChanID, updates []subscription.Update) error {

	var errs []error

	for _, update := range updates {
		var err error
		switch update.Flag {
		case subscription.SubscriberSubscribe:
			err = s.unwrap.Subscribe(id, update.ID, update.Extra)
		case subscription.SubscriberUnsubscribe:
			err = s.unwrap.UnSubscribe(id, update.ID)
		case subscription.SubscriberUpdate:
			err = s.unwrap.UpdateSubscriber(id, update.ID, update.Extra)
		default:
			return subscription.ErrUnknownFlag
		}
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}

	errMsg := ""
	for _, err := range errs {
		errMsg += err.Error() + "\n"
	}

	return errors.New(errMsg)
}

func (s *subscriptionImpl) UpdateChannel(id subscription.ChanID, update subscription.ChannelUpdate) error {

	switch update.Flag {
	case subscription.ChanCreate:
		info, ok := update.Extra.(*subscription.ChanInfo)
		if !ok {
			return errors.New("invalid channel info")
		}
		return s.unwrap.CreateChannel(id, info)
	case subscription.ChanUpdate:
		info, ok := update.Extra.(*subscription.ChanInfo)
		if !ok {
			return errors.New("invalid channel info")
		}
		return s.unwrap.UpdateChannel(id, info)
	case subscription.ChanDelete:
		return s.unwrap.RemoveChannel(id)
	default:
		return subscription.ErrUnknownFlag
	}
}

func (s *subscriptionImpl) PublishMessage(id subscription.ChanID, message subscription.Message) error {
	return s.unwrap.Publish(id, message)
}

func (s *subscriptionImpl) SetGateInterface(g gate.DefaultGateway) {
	s.unwrap.gate = g
}

var _ SubscribeWrap = (*realSubscription)(nil)

type realSubscription struct {
	mu       sync.RWMutex
	channels map[subscription.ChanID]subscription.Channel
	store    store.SubscriptionStore
	seqStore ChannelSequenceStore
	gate     gate.DefaultGateway
}

func newRealSubscription(msgStore store.SubscriptionStore, seqStore ChannelSequenceStore) *realSubscription {
	return &realSubscription{
		mu:       sync.RWMutex{},
		channels: make(map[subscription.ChanID]subscription.Channel),
		store:    msgStore,
		seqStore: seqStore,
	}
}

func (u *realSubscription) Subscribe(chID subscription.ChanID, sbID subscription.SubscriberID, extra interface{}) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	ch, ok := u.channels[chID]
	if !ok {
		return errors.New(subscription.ErrChanNotExist)
	}
	return ch.Subscribe(sbID, extra)
}

func (u *realSubscription) UnSubscribe(chID subscription.ChanID, id subscription.SubscriberID) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	ch, ok := u.channels[chID]
	if !ok {
		return errors.New(subscription.ErrChanNotExist)
	}

	return ch.Unsubscribe(id)
}

func (u *realSubscription) UpdateSubscriber(chID subscription.ChanID, id subscription.SubscriberID, update interface{}) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	ch, ok := u.channels[chID]
	if !ok {
		return errors.New(subscription.ErrChanNotExist)
	}
	return ch.Subscribe(id, update)
}

func (u *realSubscription) RemoveChannel(chID subscription.ChanID) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	_, ok := u.channels[chID]
	if !ok {
		return errors.New(subscription.ErrChanNotExist)
	}
	delete(u.channels, chID)
	return nil
}

func (u *realSubscription) CreateChannel(chID subscription.ChanID, update *subscription.ChanInfo) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.channels[chID]; ok {
		return errors.New(subscription.ErrChanAlreadyExists)
	}

	channel, err := NewChannel(chID, u.gate, u.store, u.seqStore)
	if err != nil {
		return err
	}
	err = channel.Update(update)
	if err != nil {
		return err
	}
	u.channels[chID] = channel
	return nil
}

func (u *realSubscription) UpdateChannel(chID subscription.ChanID, update *subscription.ChanInfo) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	ch, ok := u.channels[chID]
	if !ok {
		return errors.New(subscription.ErrChanNotExist)
	}

	return ch.Update(update)
}

func (u *realSubscription) Publish(chID subscription.ChanID, msg subscription.Message) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	ch, ok := u.channels[chID]
	if !ok {
		return errors.New(subscription.ErrChanNotExist)
	}
	return ch.Publish(msg)
}
