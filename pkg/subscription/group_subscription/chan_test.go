package group_subscription

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroup_Publish(t *testing.T) {
	group := NewGroup("test", 1)
	err := group.Publish(&PublishMessage{
		From:    "test",
		Seq:     1,
		Type:    TypeNotify,
		Message: &messages.GlideMessage{},
	})
	assert.NoError(t, err)
}

func TestGroup_PublishUnknownType(t *testing.T) {
	group := NewGroup("test", 1)
	err := group.Publish(&PublishMessage{})
	assert.EqualError(t, err, errUnknownMessageType)
}

func TestGroup_PublishUnexpectedMessageType(t *testing.T) {
	group := NewGroup("test", 1)
	err := group.Publish(&message{})
	assert.Error(t, err)
}

type message struct{}

func (*message) GetFrom() subscription.SubscriberID {
	return ""
}
