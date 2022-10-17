package subscription_impl

import (
	"github.com/glide-im/glide/pkg/messages"
	"github.com/glide-im/glide/pkg/subscription"
)

const (
	errUnknownMessageType = "unknown message type"
)

const (
	typeUnknown = iota

	// TypeNotify is the notification message type.
	TypeNotify

	// TypeMessage is the chat message type.
	TypeMessage

	// TypeSystem is the system message type.
	TypeSystem
)

// PublishMessage is the message published to the channel.
type PublishMessage struct {
	// From the message sender.
	From subscription.SubscriberID
	// To specified receiver, empty express all subscribers will be received.
	To  []subscription.SubscriberID
	Seq int64
	// Type the message type.
	Type    int
	Message *messages.GlideMessage
}

func (p *PublishMessage) GetFrom() subscription.SubscriberID {
	return p.From
}

func (p *PublishMessage) GetChatMessage() (*messages.ChatMessage, error) {
	cm := &messages.ChatMessage{}
	err := p.Message.Data.Deserialize(cm)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func IsUnknownMessageType(err error) bool {
	return err.Error() == errUnknownMessageType
}

func isValidMessageType(t int) bool {
	return t > typeUnknown && t <= TypeSystem
}
