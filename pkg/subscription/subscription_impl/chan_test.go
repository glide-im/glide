package subscription_impl

import (
	"github.com/glide-im/glide/pkg/gate"
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var normalOpts = &SubscriberOptions{
	Perm: PermWrite | PermRead,
}

func mockNewChannel(id subscription.ChanID) *Channel {
	channel, _ := NewChannel(id, &mockGate{}, &mockStore{}, &mockSeqStore{})
	return channel
}

type mockGate struct {
}

func (m mockGate) SetClientID(old gate.ID, new_ gate.ID) error {
	//TODO implement me
	panic("implement me")
}

func (m mockGate) UpdateClient(id gate.ID, info *gate.ClientSecrets) error {
	//TODO implement me
	panic("implement me")
}

func (m mockGate) ExitClient(id gate.ID) error {
	//TODO implement me
	panic("implement me")
}

func (m mockGate) GetClient(id gate.ID) gate.Client {
	//TODO implement me
	panic("implement me")
}

func (m mockGate) GetAll() map[gate.ID]gate.Info {
	//TODO implement me
	panic("implement me")
}

func (m mockGate) SetMessageHandler(h gate.MessageHandler) {
	//TODO implement me
	panic("implement me")
}

func (m mockGate) AddClient(cs gate.Client) {
	//TODO implement me
	panic("implement me")
}

func (m mockGate) EnqueueMessage(gate.ID, *messages.GlideMessage) error {
	return nil
}

type message struct{}

func (*message) GetFrom() subscription.SubscriberID {
	return ""
}

type mockSeqStore struct {
	segmentLen int64
	nextSeq    int64
}

func (m *mockSeqStore) NextSegmentSequence(subscription.ChanID, subscription.ChanInfo) (int64, int64, error) {
	seq := m.nextSeq
	m.nextSeq = seq + m.segmentLen
	return seq, m.segmentLen, nil
}

func TestGroup_Publish(t *testing.T) {
	channel := mockNewChannel("test")
	err2 := channel.Subscribe("test", normalOpts)
	assert.NoError(t, err2)
	msg := &PublishMessage{
		From:    "test",
		Type:    TypeNotify,
		Message: &messages.GlideMessage{},
	}
	err := channel.Publish(msg)
	assert.NoError(t, err)
	err = channel.Publish(msg)
	assert.NoError(t, err)
	err = channel.Publish(msg)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 50)
}

func TestChannel_Sleep(t *testing.T) {
	channel := mockNewChannel("test")
	err2 := channel.Subscribe("test", normalOpts)
	assert.NoError(t, err2)
	msg := &PublishMessage{
		From:    "test",
		Type:    TypeNotify,
		Message: &messages.GlideMessage{},
	}
	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(time.Millisecond * 50)
			_ = channel.Publish(msg)
		}
	}()

	time.Sleep(time.Millisecond * 50)
}

func TestChannel_PublishErr(t *testing.T) {
	channel := mockNewChannel("test")

	// invalid type
	err := channel.Publish(&PublishMessage{})
	assert.Error(t, err)

	// permission denied
	err = channel.Subscribe("t", &SubscriberOptions{Perm: PermRead})
	assert.NoError(t, err)
	err = channel.Publish(&PublishMessage{From: "t"})
	assert.Error(t, err)

	// muted
	channel.info.Muted = true
	err = channel.Publish(&PublishMessage{From: "t"})
	assert.Error(t, err)
}

func TestGroup_PublishUnknownType(t *testing.T) {
	group := mockNewChannel("test")
	err := group.Publish(&PublishMessage{})
	assert.EqualError(t, err, errUnknownMessageType)
}

func TestGroup_PublishUnexpectedMessageType(t *testing.T) {
	group := mockNewChannel("test")
	err := group.Publish(&PublishMessage{})
	assert.Error(t, err)
}

func TestChannel_nextSeq(t *testing.T) {
	m := &mockSeqStore{
		segmentLen: 4,
		nextSeq:    0,
	}
	channel, err := NewChannel("test", nil, nil, m)
	assert.NoError(t, err)

	for i := 1; i < 20; i++ {
		seq, err := channel.nextSeq()
		assert.NoError(t, err)
		assert.Equal(t, int64(i), seq)
	}
}

func TestChannel_Subscribe(t *testing.T) {
	channel := mockNewChannel("test")
	err := channel.Subscribe("sb_test", normalOpts)
	assert.NoError(t, err)
}

func TestChannel_SubscribeUpdate(t *testing.T) {
	channel := mockNewChannel("test")
	err := channel.Subscribe("sb_test", normalOpts)
	assert.NoError(t, err)
	err = channel.Subscribe("sb_test", &SubscriberOptions{Perm: PermNone})
	assert.NoError(t, err)
	assert.Equal(t, channel.subscribers["sb_test"].Perm, PermNone)
}

func TestChannel_Unsubscribe(t *testing.T) {
	channel := mockNewChannel("test")
	err := channel.Subscribe("sb_test", normalOpts)
	assert.NoError(t, err)
	err = channel.Unsubscribe("sb_test")
	assert.NoError(t, err)
	err = channel.Unsubscribe("sb_test")
	assert.EqualError(t, err, subscription.ErrNotSubscribed)
}

func TestChannel_Update(t *testing.T) {
	channel := mockNewChannel("test")
	err := channel.Update(&subscription.ChanInfo{Blocked: false, Muted: true})
	assert.NoError(t, err)
	assert.Equal(t, channel.info.Blocked, false)
	assert.Equal(t, channel.info.Muted, true)
}

func TestChannel_Close(t *testing.T) {
	channel := mockNewChannel("test")
	err := channel.Subscribe("t", normalOpts)
	assert.NoError(t, err)
	err = channel.Close()
	assert.NoError(t, err)
	err = channel.Publish(&PublishMessage{From: "t", Type: TypeMessage})
	assert.Error(t, err)
}
