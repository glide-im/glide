package group_subscription

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

// PublishMessage is the message published to the group.
type PublishMessage struct {
	From    subscription.SubscriberID
	Seq     int64
	Type    int
	Message *messages.GlideMessage
}

func (p *PublishMessage) GetFrom() subscription.SubscriberID {
	return p.From
}

func IsUnknownMessageType(err error) bool {
	return err.Error() == errUnknownMessageType
}

func isValidMessageType(t int) bool {
	return t > typeUnknown && t <= TypeSystem
}
