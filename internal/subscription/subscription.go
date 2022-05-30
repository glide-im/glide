package subscription

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/store"
	"github.com/glide-im/glide/pkg/subscription"
	"sync"
)

var _ subscription.Subscribe = (*subscriptionImpl)(nil)

type subscriptionImpl struct {
	gate  gate.Interface
	store store.SubscriptionStore

	unwrap *realSubscription
}

func NewSubscription(store store.SubscriptionStore) subscription.Interface {
	return &subscriptionImpl{
		store: store,
		unwrap: &realSubscription{
			store: store,
		},
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

func (s *subscriptionImpl) SetGateInterface(g gate.Interface) {
	s.unwrap.gate = g
}

func NewServer() (*subscriptionImpl, error) {
	return &subscriptionImpl{}, nil
}

var _ SubscribeWrap = (*realSubscription)(nil)

type realSubscription struct {
	mu       sync.RWMutex
	channels map[subscription.ChanID]subscription.Channel
	store    store.SubscriptionStore
	gate     gate.Interface
}

func (u *realSubscription) Subscribe(chID subscription.ChanID, sbID subscription.SubscriberID, extra interface{}) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	ch, ok := u.channels[chID]
	if !ok {
		return subscription.ErrChanNotExist
	}
	return ch.Subscribe(sbID, extra)
}

func (u *realSubscription) UnSubscribe(chID subscription.ChanID, id subscription.SubscriberID) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	ch, ok := u.channels[chID]
	if !ok {
		return subscription.ErrChanNotExist
	}

	return ch.Unsubscribe(id)
}

func (u *realSubscription) UpdateSubscriber(chID subscription.ChanID, id subscription.SubscriberID, update interface{}) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	ch, ok := u.channels[chID]
	if !ok {
		return subscription.ErrChanNotExist
	}
	return ch.UpdateSubscribe(id, update)
}

func (u *realSubscription) RemoveChannel(chID subscription.ChanID) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	_, ok := u.channels[chID]
	if !ok {
		return subscription.ErrChanNotExist
	}
	delete(u.channels, chID)
	return nil
}

func (u *realSubscription) CreateChannel(chID subscription.ChanID, update *subscription.ChanInfo) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.channels[chID]; ok {
		return subscription.ErrChanAlreadyExists
	}

	u.channels[chID] = newGroup(1, 0)

	return nil
}

func (u *realSubscription) UpdateChannel(chID subscription.ChanID, update *subscription.ChanInfo) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	ch, ok := u.channels[chID]
	if !ok {
		return subscription.ErrChanNotExist
	}

	return ch.Update(update)
}

func (u *realSubscription) Publish(chID subscription.ChanID, msg subscription.Message) error {
	u.mu.RUnlock()
	defer u.mu.RUnlock()

	ch, ok := u.channels[chID]
	if !ok {
		return subscription.ErrChanNotExist
	}
	return ch.Publish(msg)
}
