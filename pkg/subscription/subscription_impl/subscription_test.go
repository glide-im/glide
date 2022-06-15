package subscription_impl

import (
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockStore struct {
}

func (m *mockStore) NextSegmentSequence(id subscription.ChanID, info subscription.ChanInfo) (int64, int64, error) {
	return 1, 10000, nil
}

func (m *mockStore) StoreMessage(ch subscription.ChanID, msg subscription.Message) error {
	return nil
}

func TestRealSubscription_Publish(t *testing.T) {
	var sbp = NewSubscribeWrap(NewSubscription(&mockStore{}, &mockStore{}))
	m := PublishMessage{Type: TypeNotify}
	id := subscription.ChanID("test")
	err := sbp.CreateChannel(id, nil)
	assert.Nil(t, err)
	err = sbp.Publish(id, &m)
	assert.Nil(t, err)
}

func TestRealSubscription_PublishNotExist(t *testing.T) {
	var sbp = NewSubscribeWrap(NewSubscription(&mockStore{}, &mockStore{}))
	m := PublishMessage{}
	id := subscription.ChanID("test")
	err := sbp.Publish(id, &m)
	assert.ErrorContains(t, err, subscription.ErrChanNotExist)
}

func TestRealSubscription_CreateChannelExist(t *testing.T) {
	var sbp = NewSubscribeWrap(NewSubscription(&mockStore{}, &mockStore{}))
	id := subscription.ChanID("test")
	err := sbp.CreateChannel(id, nil)
	assert.Nil(t, err)
	err = sbp.CreateChannel(id, nil)
	assert.ErrorContains(t, err, subscription.ErrChanAlreadyExists)
}

func TestRealSubscription_CreateChannel(t *testing.T) {
	var sbp = NewSubscribeWrap(NewSubscription(&mockStore{}, &mockStore{}))
	id := subscription.ChanID("test")
	err := sbp.CreateChannel(id, nil)
	assert.Nil(t, err)
}

func TestRealSubscription_RemoveChannel(t *testing.T) {
	var sbp = NewSubscribeWrap(NewSubscription(&mockStore{}, &mockStore{}))
	id := subscription.ChanID("test")
	err := sbp.CreateChannel(id, nil)
	assert.Nil(t, err)
	err = sbp.RemoveChannel(id)
	assert.Nil(t, err)
}

func TestRealSubscription_RemoveChannelNotExit(t *testing.T) {
	var sbp = NewSubscribeWrap(NewSubscription(&mockStore{}, &mockStore{}))
	id := subscription.ChanID("test")
	err := sbp.RemoveChannel(id)
	assert.ErrorContains(t, err, subscription.ErrChanNotExist)
}

func TestRealSubscription_Subscribe(t *testing.T) {
	var sbp = NewSubscribeWrap(NewSubscription(&mockStore{}, &mockStore{}))
	id := subscription.ChanID("test")
	err := sbp.CreateChannel(id, nil)
	assert.Nil(t, err)
	err = sbp.Subscribe(id, "test", &SubscriberOptions{})
	assert.Nil(t, err)
}

func TestRealSubscription_UnSubscribe(t *testing.T) {
	var sbp = NewSubscribeWrap(NewSubscription(&mockStore{}, &mockStore{}))
	id := subscription.ChanID("test")
	err := sbp.CreateChannel(id, nil)
	assert.Nil(t, err)
	err = sbp.Subscribe(id, "test", &SubscriberOptions{})
	assert.Nil(t, err)
	err = sbp.UnSubscribe(id, "test")
	assert.Nil(t, err)
}
