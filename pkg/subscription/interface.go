package subscription

import (
	"github.com/glide-im/glide/pkg/gate"
)

const (
	SubscriberSubscribe   int64 = 1
	SubscriberUnsubscribe       = 2
	SubscriberMute              = 3
	SubscriberUnmute            = 4
)

const (
	ChanCreate int64 = 1
	ChanDelete       = 2
	ChanMute         = 3
	ChanUnmute       = 4
)

// ChanID is a unique identifier for a channel.
type ChanID string

type SubscriberUpdate struct {
	Flag int64
	ID   gate.ID

	Extra interface{}
}

type ChannelUpdate struct {
	Flag int64

	Extra interface{}
}

type Interface interface {
	PublishMessage(id ChanID, message Message) error
}

type Subscribe interface {
	SetGateInterface(gate gate.Interface)

	UpdateSubscriber(id ChanID, updates []SubscriberUpdate) error

	UpdateChannel(id ChanID, update ChannelUpdate) error
}

type Server interface {
	Subscribe

	Run() error
}
