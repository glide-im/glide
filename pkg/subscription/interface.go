package subscription

import (
	"errors"
	"github.com/glide-im/glide/pkg/gate"
)

const (
	SubscriberSubscribe   int64 = 1
	SubscriberUnsubscribe       = 2
	SubscriberUpdate            = 5
)

const (
	ChanCreate int64 = 1
	ChanDelete       = 2
	ChanUpdate       = 3
)

var (
	ErrUnknownFlag = errors.New("unknown flag")
)

// ChanID is a unique identifier for a channel.
type ChanID string

type SubscriberID string

type Update struct {
	Flag int64
	ID   SubscriberID

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
	Interface

	SetGateInterface(gate gate.DefaultGateway)

	UpdateSubscriber(id ChanID, updates []Update) error

	UpdateChannel(id ChanID, update ChannelUpdate) error
}

type Server interface {
	Subscribe

	Run() error
}
