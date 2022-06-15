package subscription_impl

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func mockNewChannel(id subscription.ChanID, seq int64) *Channel {
	return &Channel{
		id:        id,
		seq:       seq,
		seqRemain: math.MaxInt64,
	}
}

func TestGroup_Publish(t *testing.T) {
	group := mockNewChannel("test", 1)
	err := group.Publish(&PublishMessage{
		From:    "test",
		Seq:     1,
		Type:    TypeNotify,
		Message: &messages.GlideMessage{},
	})
	assert.NoError(t, err)
}

func TestGroup_PublishUnknownType(t *testing.T) {
	group := mockNewChannel("test", 1)
	err := group.Publish(&PublishMessage{})
	assert.EqualError(t, err, errUnknownMessageType)
}

func TestGroup_PublishUnexpectedMessageType(t *testing.T) {
	group := mockNewChannel("test", 1)
	err := group.Publish(&message{})
	assert.Error(t, err)
}

type message struct{}

func (*message) GetFrom() subscription.SubscriberID {
	return ""
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

type mockSeqStore struct {
	segmentLen int64
	nextSeq    int64
}

func (m *mockSeqStore) NextSegmentSequence(subscription.ChanID, subscription.ChanInfo) (int64, int64, error) {
	seq := m.nextSeq
	m.nextSeq = seq + m.segmentLen
	return seq, m.segmentLen, nil
}
